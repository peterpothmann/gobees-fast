package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// MortalityBrood [GO-16].
//
// Brood that is older than the maximum age of its stage dies, and after the competition point all eggs die
// [MORB]. Each death is counted in one brood death counter of the colony (egg accounting [CHK-04]).
//
// NetLogo: MortalityBroodProc.
type MortalityBrood struct {
	rng   *rand.Rand
	ticks *globals.Ticks
	death beeDeath

	beeFilter  *ecs.Filter1[comp.Bee]
	beeMap     *ecs.Map2[comp.Bee, comp.Development]
	colonyMap  *ecs.Map1[comp.Colony]
	speciesMap *ecs.Map1[comp.Species]
}

func (s *MortalityBrood) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.death = newBeeDeath(w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap2[comp.Bee, comp.Development](w)
	s.colonyMap = ecs.NewMap1[comp.Colony](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
}

func (s *MortalityBrood) Update(w *ecs.World) {
	ticks := float64(s.ticks.Ticks)
	// [MORB-01] to [MORB-04] in this order for each bee. ORDER of the bees: irrelevant, each bee changes only
	// itself and the counters of its colony (sums). Shuffled (decision C).
	// DieProc ends the bee's code (finding 28, see trans/netlogo-findings.md#finding-28): continue after it.
	for _, e := range askAllBees(s.beeFilter, s.rng) {
		b, dev := s.beeMap.Get(e)
		sp := s.speciesMap.Get(b.SpeciesID)
		// memoNumber and the brood death counters of the colony (eggDeathsIncubation, larvaDeathsWeight,
		// larvaDeathsIncubation, pupaDeathsIncubation, broodDeathsCP): read only by the egg accounting [CHK-04],
		// removed in step F4. The assertion of [MORB-02] stays.

		// [MORB-01]
		if b.Stage == stage.Egg && b.BroodAge > sp.DevAgeHatchingMax_d {
			s.death.DieProc(w, e, "Egg: broodAge > devAgeHatchingMax_d")
			continue
		}

		// [MORB-02]
		if b.Stage == stage.Larva {
			if ((b.Caste == caste.Worker || b.Caste == caste.Male || b.Caste == caste.Undefined) && b.BroodAge > sp.DevAgePupationMax_d) ||
				(b.Caste == caste.Queen && b.BroodAge > sp.DevAge_Q_PupationMax_d) {
				var relativeIncub, relativeWeight float64
				if b.Caste == caste.Queen {
					relativeWeight = b.Weight_mg / sp.DevWeight_Q_PupationMin_mg
					relativeIncub = dev.CumulIncubationReceived_kJ / sp.DevIncubation_Q_PupationTH_kJ
				} else {
					relativeWeight = b.Weight_mg / sp.DevWeightPupationMin_mg
					relativeIncub = dev.CumulIncubationReceived_kJ / sp.DevIncubationPupationTH_kJ
				}
				if relativeWeight >= 1 && relativeIncub >= 1 {
					panic("Neither Weight or Incubation reason for death: MortalityBroodProc (1)")
				}
				s.death.DieProc(w, e, "Larva: broodAge > max. pupation age")
				continue
			}
		}

		// [MORB-03]
		if b.Stage == stage.Pupa {
			if b.Caste == caste.Undefined {
				panic("Pupa with undefined caste (MortalityProc)!")
			}
			if ((b.Caste == caste.Worker || b.Caste == caste.Male) && b.BroodAge > sp.DevAgeEmergingMax_d) ||
				(b.Caste == caste.Queen && b.BroodAge > sp.DevAge_Q_EmergingMax_d) {
				s.death.DieProc(w, e, "Pupa: broodAge > max. emerging age")
				continue
			}
		}

		// [MORB-04] adults without colony are skipped
		if !b.ColonyID.IsZero() {
			if colony := s.colonyMap.Get(b.ColonyID); ticks > colony.CompetitionPointDate {
				if b.Stage == stage.Egg {
					s.death.DieProc(w, e, "Egg: CP!")
				}
			}
		}
	}
}

func (s *MortalityBrood) Finalize(w *ecs.World) {}
