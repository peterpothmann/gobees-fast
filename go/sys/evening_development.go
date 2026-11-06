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

// development [GO-15].
//
// All bees age by one day [DEV]. Adults: adult mortality and queen thresholds [MORA]. Brood: development to the
// next stage [DEVP], [DEVL], [DEVE]. Must run after queenProductionDate (finding 42, see
// trans/netlogo-findings.md#finding-42).
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
	// [DEV-01] ALL bees. ORDER: irrelevant for the state, see the annotation. Shuffled (decision C).
	// A bee dies only in its own turn [MORA], so every bee of the list is alive when its turn comes.
	s.bees = askAllBees(s.bees, s.beeFilter, s.rng)
	for _, e := range s.bees {
		_, bee, _, _, _, _ := s.beeMap.Get(e)
		whoCol := bee.Colony
		if bee.Stage == stage.Adult {
			// [DEV-02]
			bee.AdultAge = bee.AdultAge + 1
			if bee.AdultAge > 700 {
				panic("Assertion violated: Bee with 2 hibernations!")
			}
			// assertion adult bee with brood? = true: brood? is the stage since step F9, it cannot fail
			s.ageAdult(w, e)
			continue
		}
		// [DEV-03] a switch on the stage at the start: each bee advances at most one stage per day
		bee.BroodAge = bee.BroodAge + 1
		// assertion brood with brood? = false: brood? is the stage since step F9, it cannot fail
		switch bee.Stage {
		case stage.Pupa:
			s.developPupa(e)
		case stage.Larva:
			s.developLarva(e, whoCol)
		case stage.Egg:
			s.developEgg(e)
		}
		// [DEV-04] GUI ONLY, not translated.
	}
}

func (s *development) Finalize(w *ecs.World) {}

// ageAdult applies the maximum life span, the social-phase thresholds of the colony
// queen and the background mortality to an adult bee [MORA].
func (s *development) ageAdult(w *ecs.World, e ecs.Entity) {
	_, bee, _, th, _, _ := s.beeMap.Get(e)
	// [MORA-01] DieProc ends the bee's code (finding 28, see trans/netlogo-findings.md#finding-28)
	if bee.Caste == caste.Worker {
		if bee.AdultAge > s.speciesMap.Get(bee.Species).MaxLifespanWorkers {
			s.death.die(w, e, "Worker: adultAge > maxLifespanWorkers")
			return
		}
	}
	if bee.Caste == caste.Male {
		if bee.AdultAge > s.constants.MaxLifespanMales {
			s.death.die(w, e, "Male: adultAge > MaxLifespanMales")
			return
		}
	}

	// [MORA-02] no species check: also a Psithyrus queen (finding 44, see trans/netlogo-findings.md#finding-44)
	if bee.Caste == caste.Queen && bee.Mated && !bee.Colony.IsZero() {
		if _, counts := s.colonyMap.Get(bee.Colony); counts.AdultWorkers > 0 {
			*th = thresholds(sitQueenSocialPhase)
		}
	}

	// [MORA-03] inverted condition (finding 22, see trans/netlogo-findings.md#finding-22). With the value 0, no
	// random number is drawn.
	if s.constants.AdultBackgroundMortality > 0 && s.rng.Float64() > s.constants.AdultBackgroundMortality {
		s.death.die(w, e, "Adult bee: mortality in colony")
	}
}

// developPupa lets a pupa become an adult once it is old enough and incubated enough [DEVP].
// totalAdultsProduced of the colony: read only by the egg accounting [CHK-04], removed in step F4.
func (s *development) developPupa(e ecs.Entity) {
	_, bee, _, th, forager, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	// [DEVP-01] cumulTimePupa_d: never read, removed in step F3.

	// [DEVP-02]
	if bee.Caste == caste.Worker {
		if dev.IncubationReceived >= sp.EmergenceIncubation && float64(bee.BroodAge) >= sp.EmergenceAgeMin {
			bee.Stage = stage.Adult
			forager.CropVolume = cropVolume(bee.Weight, sp)
			forager.PollenPellets = pollenLoad(bee.Weight, sp)
			*th = thresholds(sitWorker)
			forager.GlossaLength = proboscisLength(bee.Weight, sp)
		}
	}

	// [DEVP-03] males get no crop, pollen load, thresholds or proboscis
	if bee.Caste == caste.Male {
		if dev.IncubationReceived >= sp.EmergenceIncubation && float64(bee.BroodAge) >= sp.EmergenceAgeMin {
			bee.Stage = stage.Adult
		}
	}

	// [DEVP-04]
	if bee.Caste == caste.Queen {
		if dev.IncubationReceived >= sp.QueenEmergenceIncubation && float64(bee.BroodAge) >= sp.QueenEmergenceAgeMin {
			// The emerging date of a new queen is unset (NetLogo: NotSetLow, set in EggsParameterSettingProc), so it is
			// drawn at least once.
			yearStartSeason := day.YearEnd(s.ticks.Ticks)
			yearEndSeason := yearStartSeason + day.Abs(sp.SeasonStop)
			for !bee.EmergenceDate.IsSet() || bee.EmergenceDate.Day() <= yearStartSeason || bee.EmergenceDate.Day() > yearEndSeason {
				bee.EmergenceDate = day.At(yearStartSeason + day.Abs(round(randomNormal(s.rng, sp.EmergenceDayMean, sp.EmergenceDaySD))))
			}
			bee.Stage = stage.Adult
			s.queenEmerged.Emit(e) // totalAdultQueensEverProduced: output only, counted by the observers (step F4)
			forager.CropVolume = cropVolume(bee.Weight, sp)
			forager.PollenPellets = pollenLoad(bee.Weight, sp)
			*th = thresholds(sitYoungQueen)
			forager.GlossaLength = proboscisLength(bee.Weight, sp)
		}
	}

	// [DEVP-05]
	if bee.Caste == caste.Undefined {
		panic("Assertion violated: undefined caste! (Development_PupaeProc)")
	}
}

// determineCaste returns the caste of a female larva at larval day 3 [CAST]: queen if she is heavy enough
// and her colony produces queens, else worker. dev_Q_DeterminationWeight_mg is 0 for all species, so every
// female larva becomes a queen once the colony produces queens (finding 40, see
// trans/netlogo-findings.md#finding-40).
// weight is in [mg].
func determineCaste(weight float64, sp *params.Species, col *comp.Colony) caste.Caste {
	// [CAST-01]
	mycaste := caste.Worker
	if weight >= sp.QueenDeterminationWeight && col.QueenProduction {
		mycaste = caste.Queen
	}
	return mycaste
}

// developLarva determines the caste of a female larva at larval day 3 and lets a larva pupate once it
// is old enough, incubated enough and heavy enough [DEVL]. whoCol is the larva's own colony.
func (s *development) developLarva(e ecs.Entity, whoCol ecs.Entity) {
	_, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	// [DEVL-01]
	dev.LarvaTime = dev.LarvaTime + 1

	// [DEVL-02]
	if bee.Caste == caste.Undefined && dev.LarvaTime == sp.QueenDeterminationAge {
		colony, _ := s.colonyMap.Get(whoCol)
		bee.Caste = determineCaste(bee.Weight, sp, colony)
		if bee.Caste == caste.Undefined {
			panic("Assertion violated: undefined caste! (Development_LarvaeProc)")
		}
	}

	// [DEVL-03] the brackets around (worker || male) are required (finding 41, see
	// trans/netlogo-findings.md#finding-41)
	if (bee.Caste == caste.Worker || bee.Caste == caste.Male) &&
		dev.IncubationReceived >= sp.PupationIncubation &&
		float64(bee.BroodAge) >= sp.PupationAgeMin &&
		bee.Weight >= sp.PupationWeightMin {
		bee.Stage = stage.Pupa // totalPupaeProduced: never read, removed in step F3
	}

	if bee.Caste == caste.Queen &&
		dev.IncubationReceived >= sp.QueenPupationIncubation &&
		float64(bee.BroodAge) >= sp.QueenPupationAgeMin &&
		bee.Weight >= sp.QueenPupationWeightMin {
		bee.Stage = stage.Pupa
	}
}

// developEgg lets an egg hatch into a larva once it is old enough and incubated enough [DEVE].
func (s *development) developEgg(e ecs.Entity) {
	_, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.Species)
	// [DEVE-01] cumulTimeEgg_d and totalLarvaeProduced: never read, removed in step F3.
	if dev.IncubationReceived >= sp.HatchingIncubation && float64(bee.BroodAge) >= sp.HatchingAgeMin {
		bee.Stage = stage.Larva
	}
}
