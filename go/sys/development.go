package sys

import (
	"math"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/event"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Development [GO-15].
//
// All bees age by one day [DEV]. Adults: adult mortality and queen thresholds [MORA]. Brood: development to the
// next stage [DEVP], [DEVL], [DEVE]. Must run after QueenProductionDate (finding 42, see
// trans/netlogo-findings.md#finding-42).
//
// NetLogo: DevelopmentProc, Development_Mortality_AdultsProc, Development_PupaeProc, Development_LarvaeProc,
// Development_EggsProc, DetermineCaste_REP.
type Development struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	constants *params.Constants
	death     beeDeath
	// queenEmerged is event.QueenEmerged.
	queenEmerged ecs.Event

	beeFilter  *ecs.Filter1[comp.Bee]
	beeMap     *beeMapper
	colonyMap  *ecs.Map2[comp.Colony, comp.ColonyCounts]
	speciesMap *ecs.Map1[comp.Species]
}

func (s *Development) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.queenEmerged = w.Event(event.QueenEmerged).For(ecs.C[comp.Bee]())
	s.death = newBeeDeath(w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = newBeeMapper(w)
	s.colonyMap = ecs.NewMap2[comp.Colony, comp.ColonyCounts](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
}

func (s *Development) Update(w *ecs.World) {
	// [DEV-01] ALL bees. ORDER: irrelevant for the state, see the annotation. Shuffled (decision C).
	// A bee dies only in its own turn [MORA], so every bee of the list is alive when its turn comes.
	for _, e := range askAllBees(s.beeFilter, s.rng) {
		_, _, bee, _, _, _, _ := s.beeMap.Get(e)
		whoCol := bee.ColonyID
		if bee.Stage == stage.Adult {
			// [DEV-02]
			bee.AdultAge = bee.AdultAge + 1
			if bee.AdultAge > 700 {
				panic("Assertion violated: Bee with 2 hibernations!")
			}
			if bee.Brood {
				panic("Assertion violated: adult bee with brood? = true (DevelopmentProc)")
			}
			s.Development_Mortality_AdultsProc(w, e)
			continue
		}
		// [DEV-03] a switch on the stage at the start: each bee advances at most one stage per day
		bee.BroodAge = bee.BroodAge + 1
		if !bee.Brood {
			panic("Assertion violated: brood with brood? = false (DevelopmentProc)")
		}
		switch bee.Stage {
		case stage.Pupa:
			s.Development_PupaeProc(e)
		case stage.Larva:
			s.Development_LarvaeProc(e, whoCol)
		case stage.Egg:
			s.Development_EggsProc(e)
		}
		// [DEV-04] GUI ONLY, not translated.
	}
}

func (s *Development) Finalize(w *ecs.World) {}

// Development_Mortality_AdultsProc applies the maximum life span, the social-phase thresholds of the colony
// queen and the background mortality to an adult bee [MORA].
func (s *Development) Development_Mortality_AdultsProc(w *ecs.World, e ecs.Entity) {
	_, _, bee, _, th, _, _ := s.beeMap.Get(e)
	// [MORA-01] DieProc ends the bee's code (finding 28, see trans/netlogo-findings.md#finding-28)
	if bee.Caste == caste.Worker {
		if bee.AdultAge > s.speciesMap.Get(bee.SpeciesID).MaxLifespanWorkers {
			s.death.DieProc(w, e, "Worker: adultAge > maxLifespanWorkers")
			return
		}
	}
	if bee.Caste == caste.Male {
		if bee.AdultAge > s.constants.MaxLifespanMales {
			s.death.DieProc(w, e, "Male: adultAge > MaxLifespanMales")
			return
		}
	}

	// [MORA-02] no species check: also a Psithyrus queen (finding 44, see trans/netlogo-findings.md#finding-44)
	if bee.Caste == caste.Queen && bee.Mated && !bee.ColonyID.IsZero() {
		if _, counts := s.colonyMap.Get(bee.ColonyID); counts.AllAdultWorkers > 0 {
			*th = thresholds(sitQueenSocialPhase)
		}
	}

	// [MORA-03] inverted condition (finding 22, see trans/netlogo-findings.md#finding-22). With the value 0, no
	// random number is drawn.
	if s.constants.MortalityAdultsBackground_daily > 0 && s.rng.Float64() > s.constants.MortalityAdultsBackground_daily {
		s.death.DieProc(w, e, "Adult bee: mortality in colony")
	}
}

// Development_PupaeProc lets a pupa become an adult once it is old enough and incubated enough [DEVP].
// totalAdultsProduced of the colony: read only by the egg accounting [CHK-04], removed in step F4.
func (s *Development) Development_PupaeProc(e ecs.Entity) {
	_, _, bee, _, th, forager, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.SpeciesID)
	// [DEVP-01] cumulTimePupa_d: never read, removed in step F3.

	// [DEVP-02]
	if bee.Caste == caste.Worker {
		if dev.CumulIncubationReceived_kJ >= sp.DevIncubationEmergingTH_kJ && bee.BroodAge >= sp.DevAgeEmergingMin_d {
			bee.Stage = stage.Adult
			bee.Brood = false
			forager.Cropvolume_myl = CropAndPelletSizeREP("nectar", bee.Weight_mg, sp)
			forager.PollenPellets_g = CropAndPelletSizeREP("pollen", bee.Weight_mg, sp)
			*th = thresholds(sitWorker)
			forager.GlossaLength_mm = ProboscisLengthREP(bee.Weight_mg, sp)
		}
	}

	// [DEVP-03] males get no crop, pollen load, thresholds or proboscis
	if bee.Caste == caste.Male {
		if dev.CumulIncubationReceived_kJ >= sp.DevIncubationEmergingTH_kJ && bee.BroodAge >= sp.DevAgeEmergingMin_d {
			bee.Brood = false
			bee.Stage = stage.Adult
		}
	}

	// [DEVP-04]
	if bee.Caste == caste.Queen {
		if dev.CumulIncubationReceived_kJ >= sp.DevIncubation_Q_EmergingTH_kJ && bee.BroodAge >= sp.DevAge_Q_EmergingMin_d {
			ticks := float64(s.ticks.Ticks)
			yearEndSeason := (365 * math.Ceil(ticks/365)) + sp.SeasonStop
			yearStartSeason := 365 * math.Ceil(ticks/365)
			for bee.EmergingDate <= yearStartSeason || bee.EmergingDate > yearEndSeason {
				bee.EmergingDate = (365 * math.Ceil(ticks/365)) + util.Round(util.RandomNormal(s.rng, sp.EmergingDay_mean, sp.EmergingDay_sd))
			}
			bee.Stage = stage.Adult
			bee.Brood = false
			s.queenEmerged.Emit(e) // totalAdultQueensEverProduced: output only, counted by the observers (step F4)
			forager.Cropvolume_myl = CropAndPelletSizeREP("nectar", bee.Weight_mg, sp)
			forager.PollenPellets_g = CropAndPelletSizeREP("pollen", bee.Weight_mg, sp)
			*th = thresholds(sitYoungQueen)
			forager.GlossaLength_mm = ProboscisLengthREP(bee.Weight_mg, sp)
		}
	}

	// [DEVP-05]
	if bee.Caste == caste.Undefined {
		panic("Assertion violated: undefined caste! (Development_PupaeProc)")
	}
}

// DetermineCaste_REP returns the caste of a female larva at larval day 3 [CAST]: queen if she is heavy enough
// and her colony produces queens, else worker. dev_Q_DeterminationWeight_mg is 0 for all species, so every
// female larva becomes a queen once the colony produces queens (finding 40, see
// trans/netlogo-findings.md#finding-40).
func DetermineCaste_REP(weight_mg float64, sp *comp.Species, col *comp.Colony) caste.Caste {
	// [CAST-01]
	mycaste := caste.Worker
	if weight_mg >= sp.Dev_Q_DeterminationWeight_mg && col.QueenProduction {
		mycaste = caste.Queen
	}
	return mycaste
}

// Development_LarvaeProc determines the caste of a female larva at larval day 3 and lets a larva pupate once it
// is old enough, incubated enough and heavy enough [DEVL]. whoCol is the larva's own colony.
func (s *Development) Development_LarvaeProc(e ecs.Entity, whoCol ecs.Entity) {
	_, _, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.SpeciesID)
	// [DEVL-01]
	dev.CumulTimeLarva_d = dev.CumulTimeLarva_d + 1

	// [DEVL-02]
	if bee.Caste == caste.Undefined && dev.CumulTimeLarva_d == sp.Dev_larvalAge_QueenDetermination_d {
		colony, _ := s.colonyMap.Get(whoCol)
		bee.Caste = DetermineCaste_REP(bee.Weight_mg, sp, colony)
		if bee.Caste == caste.Undefined {
			panic("Assertion violated: undefined caste! (Development_LarvaeProc)")
		}
	}

	// [DEVL-03] the brackets around (worker || male) are required (finding 41, see
	// trans/netlogo-findings.md#finding-41)
	if (bee.Caste == caste.Worker || bee.Caste == caste.Male) &&
		dev.CumulIncubationReceived_kJ >= sp.DevIncubationPupationTH_kJ &&
		bee.BroodAge >= sp.DevAgePupationMin_d &&
		bee.Weight_mg >= sp.DevWeightPupationMin_mg {
		bee.Stage = stage.Pupa // totalPupaeProduced: never read, removed in step F3
	}

	if bee.Caste == caste.Queen &&
		dev.CumulIncubationReceived_kJ >= sp.DevIncubation_Q_PupationTH_kJ &&
		bee.BroodAge >= sp.DevAge_Q_PupationMin_d &&
		bee.Weight_mg >= sp.DevWeight_Q_PupationMin_mg {
		bee.Stage = stage.Pupa
	}
}

// Development_EggsProc lets an egg hatch into a larva once it is old enough and incubated enough [DEVE].
func (s *Development) Development_EggsProc(e ecs.Entity) {
	_, _, bee, _, _, _, dev := s.beeMap.Get(e)
	sp := s.speciesMap.Get(bee.SpeciesID)
	// [DEVE-01] cumulTimeEgg_d and totalLarvaeProduced: never read, removed in step F3.
	if dev.CumulIncubationReceived_kJ >= sp.DevIncubationHatchingTH_kJ && bee.BroodAge >= sp.DevAgeHatchingMin_d {
		bee.Stage = stage.Larva
	}
}
