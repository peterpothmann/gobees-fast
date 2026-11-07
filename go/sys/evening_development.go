package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// development lets all bees age by one day. Adults: adult mortality and queen thresholds. Brood: development to the
// next stage. Must run after queenProductionDate, so that queenProductionDate reads the larvae that got their caste
// on the previous evening.
//
// NetLogo: DevelopmentProc, Development_Mortality_AdultsProc, Development_PupaeProc, Development_LarvaeProc,
// Development_EggsProc, DetermineCaste_REP.
type development struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	constants *params.Constants
	death     beeDeath
	// queenEmerged is event.QueenEmerged.
	queenEmerged ecs.Event

	beeFilter  *ecs.Filter1[comp.Bee]
	beeMap     *beeMapper
	colonyMap  *ecs.Map2[comp.Colony, comp.ColonyCounts]
	speciesMap *speciesTable

	bees []ecs.Entity // buffer of askAllBees
}

func (s *development) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.queenEmerged = w.Event(event.QueenEmerged).For(ecs.C[comp.Bee]())
	s.death = newBeeDeath(w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = newBeeMapper(w)
	s.colonyMap = ecs.NewMap2[comp.Colony, comp.ColonyCounts](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *development) Update(w *ecs.World) {
	// ALL bees. ORDER: irrelevant for the state. Shuffled, as NetLogo asks the bees in random order.
	// A bee dies only in its own turn, so every bee of the list is alive when its turn comes.
	s.bees = askAllBees(s.bees, s.beeFilter, s.rng)
	for _, e := range s.bees {
		_, bee, _, _, _, _ := s.beeMap.Get(e)
		whoCol := bee.Colony
		if bee.Stage == stage.Adult {
			bee.AdultAge = bee.AdultAge + 1
			if bee.AdultAge > 700 {
				panic("Assertion violated: Bee with 2 hibernations!")
			}
			s.ageAdult(w, e)
			continue
		}
		// a switch on the stage at the start: each bee advances at most one stage per day
		bee.BroodAge = bee.BroodAge + 1
		switch bee.Stage {
		case stage.Pupa:
			s.developPupa(e)
		case stage.Larva:
			s.developLarva(e, whoCol)
		case stage.Egg:
			s.developEgg(e)
		}
	}
}

func (s *development) Finalize(w *ecs.World) {}

// ageAdult applies the maximum life span, the social-phase thresholds of the colony
// queen and the background mortality to an adult bee.
func (s *development) ageAdult(w *ecs.World, e ecs.Entity) {
	_, bee, _, th, _, _ := s.beeMap.Get(e)
	// die ends the bee's turn: return after it.
	if bee.Caste == caste.Worker {
		if bee.AdultAge > s.speciesMap.Get(bee.Species).MaxLifespanWorkers {
			s.death.die(w, e, "Worker: adultAge > maxLifespanWorkers")
			return
		}
	}
	if bee.Caste == caste.Male {
		if bee.AdultAge > s.constants.Bees.MaxLifespanMales {
			s.death.die(w, e, "Male: adultAge > MaxLifespanMales")
			return
		}
	}

	// no species check: a Psithyrus queen also gets the social-phase thresholds, as in NetLogo
	if bee.Caste == caste.Queen && bee.Mated && !bee.Colony.IsZero() {
		if _, counts := s.colonyMap.Get(bee.Colony); counts.AdultWorkers > 0 {
			*th = thresholds(&s.constants.Thresholds, sitQueenSocialPhase)
		}
	}

	// inverted condition as in NetLogo: a bee dies with probability 1 - AdultBackgroundMortality. With the value 0,
	// no random number is drawn.
	if s.constants.Bees.AdultBackgroundMortality > 0 && s.rng.Float64() > s.constants.Bees.AdultBackgroundMortality {
		s.death.die(w, e, "Adult bee: mortality in colony")
	}
}

// developPupa lets a pupa become an adult once it is old enough and incubated enough.
func (s *development) developPupa(e ecs.Entity) {
	_, bee, _, th, forager, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	if bee.Caste == caste.Worker {
		if dev.IncubationReceived >= sp.EmergenceIncubation && float64(bee.BroodAge) >= sp.EmergenceAgeMin {
			bee.Stage = stage.Adult
			forager.CropVolume = cropVolume(bee.Weight, sp, &s.constants.Foraging)
			forager.PollenPellets = pollenLoad(bee.Weight, sp, &s.constants.Foraging)
			*th = thresholds(&s.constants.Thresholds, sitWorker)
			forager.GlossaLength = proboscisLength(bee.Weight, sp)
		}
	}

	// males get no crop, pollen load, thresholds or proboscis
	if bee.Caste == caste.Male {
		if dev.IncubationReceived >= sp.EmergenceIncubation && float64(bee.BroodAge) >= sp.EmergenceAgeMin {
			bee.Stage = stage.Adult
		}
	}

	if bee.Caste == caste.Queen {
		if dev.IncubationReceived >= sp.QueenEmergenceIncubation && float64(bee.BroodAge) >= sp.QueenEmergenceAgeMin {
			// The emergence date of a new queen is unset, so it is drawn at least once.
			yearStartSeason := day.YearEnd(s.ticks.Ticks)
			yearEndSeason := yearStartSeason + day.Abs(sp.SeasonStop)
			for !bee.EmergenceDate.IsSet() || bee.EmergenceDate.Day() <= yearStartSeason || bee.EmergenceDate.Day() > yearEndSeason {
				bee.EmergenceDate = day.At(yearStartSeason + day.Abs(round(randomNormal(s.rng, sp.EmergenceDayMean, sp.EmergenceDaySD))))
			}
			bee.Stage = stage.Adult
			s.queenEmerged.Emit(e) // the observers count the emerged queens
			forager.CropVolume = cropVolume(bee.Weight, sp, &s.constants.Foraging)
			forager.PollenPellets = pollenLoad(bee.Weight, sp, &s.constants.Foraging)
			*th = thresholds(&s.constants.Thresholds, sitYoungQueen)
			forager.GlossaLength = proboscisLength(bee.Weight, sp)
		}
	}

	if bee.Caste == caste.Undefined {
		panic("Assertion violated: undefined caste! (developPupa)")
	}
}

// determineCaste returns the caste of a female larva at larval day 3: queen if she is heavy enough
// and her colony produces queens, else worker. QueenDeterminationWeight is 0 for all species, so every
// female larva becomes a queen once the colony produces queens.
// weight is in [mg].
func determineCaste(weight float64, sp *params.Species, col *comp.Colony) caste.Caste {
	mycaste := caste.Worker
	if weight >= sp.QueenDeterminationWeight && col.QueenProduction {
		mycaste = caste.Queen
	}
	return mycaste
}

// developLarva determines the caste of a female larva at larval day 3 and lets a larva pupate once it
// is old enough, incubated enough and heavy enough. whoCol is the larva's own colony.
func (s *development) developLarva(e ecs.Entity, whoCol ecs.Entity) {
	_, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	dev.LarvaTime = dev.LarvaTime + 1

	if bee.Caste == caste.Undefined && dev.LarvaTime == sp.QueenDeterminationAge {
		colony, _ := s.colonyMap.Get(whoCol)
		bee.Caste = determineCaste(bee.Weight, sp, colony)
		if bee.Caste == caste.Undefined {
			panic("Assertion violated: undefined caste! (developLarva)")
		}
	}

	// the brackets around (worker || male) are required: NetLogo evaluates and/or from left to right, so its
	// condition means (worker or male) and ...
	if (bee.Caste == caste.Worker || bee.Caste == caste.Male) &&
		dev.IncubationReceived >= sp.PupationIncubation &&
		float64(bee.BroodAge) >= sp.PupationAgeMin &&
		bee.Weight >= sp.PupationWeightMin {
		bee.Stage = stage.Pupa
	}

	if bee.Caste == caste.Queen &&
		dev.IncubationReceived >= sp.QueenPupationIncubation &&
		float64(bee.BroodAge) >= sp.QueenPupationAgeMin &&
		bee.Weight >= sp.QueenPupationWeightMin {
		bee.Stage = stage.Pupa
	}
}

// developEgg lets an egg hatch into a larva once it is old enough and incubated enough.
func (s *development) developEgg(e ecs.Entity) {
	_, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	if dev.IncubationReceived >= sp.HatchingIncubation && float64(bee.BroodAge) >= sp.HatchingAgeMin {
		bee.Stage = stage.Larva
	}
}
