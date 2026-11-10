package sys

import (
	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// broodMortality kills brood that is older than the maximum age of its stage, and after the competition point all
// eggs.
//
// NetLogo: MortalityBroodProc.
type broodMortality struct {
	ticks *globals.Ticks
	death beeDeath

	beeFilter  *ecs.Filter1[comp.Bee]
	beeMap     *ecs.Map2[comp.Bee, comp.Development]
	colonyMap  *ecs.Map1[comp.Colony]
	speciesMap *speciesTable

	bees []ecs.Entity // buffer of allBees
}

func (s *broodMortality) Initialize(w *ecs.World) {
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.death = newBeeDeath(w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap2[comp.Bee, comp.Development](w)
	s.colonyMap = ecs.NewMap1[comp.Colony](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *broodMortality) Update(w *ecs.World) {
	ticks := s.ticks.Ticks
	// in this order for each bee. ORDER of the bees: irrelevant, each bee changes only itself and draws no random
	// number. Not shuffled (NetLogo: random order).
	// die ends the bee's turn: continue after it.
	s.bees = allBees(s.bees, s.beeFilter)
	for _, e := range s.bees {
		b, dev := s.beeMap.Get(e)
		sp := s.speciesMap.Get(b.Species)
		if b.Stage == stage.Egg && float64(b.BroodAge) > sp.HatchingAgeMax {
			s.death.die(w, e, "Egg: broodAge > HatchingAgeMax")
			continue
		}

		if b.Stage == stage.Larva {
			if ((b.Caste == caste.Worker || b.Caste == caste.Male || b.Caste == caste.Undefined) && float64(b.BroodAge) > sp.PupationAgeMax) ||
				(b.Caste == caste.Queen && float64(b.BroodAge) > sp.QueenPupationAgeMax) {
				var relativeIncub, relativeWeight float64
				if b.Caste == caste.Queen {
					relativeWeight = b.Weight / sp.QueenPupationWeightMin
					relativeIncub = dev.IncubationReceived / sp.QueenPupationIncubation
				} else {
					relativeWeight = b.Weight / sp.PupationWeightMin
					relativeIncub = dev.IncubationReceived / sp.PupationIncubation
				}
				if relativeWeight >= 1 && relativeIncub >= 1 {
					panic("Neither Weight or Incubation reason for death: broodMortality (1)")
				}
				s.death.die(w, e, "Larva: broodAge > max. pupation age")
				continue
			}
		}

		if b.Stage == stage.Pupa {
			if b.Caste == caste.Undefined {
				panic("Pupa with undefined caste (broodMortality)!")
			}
			if ((b.Caste == caste.Worker || b.Caste == caste.Male) && float64(b.BroodAge) > sp.EmergenceAgeMax) ||
				(b.Caste == caste.Queen && float64(b.BroodAge) > sp.QueenEmergenceAgeMax) {
				s.death.die(w, e, "Pupa: broodAge > max. emerging age")
				continue
			}
		}

		// adults without colony are skipped
		if !b.Colony.IsZero() {
			if colony := s.colonyMap.Get(b.Colony); colony.CompetitionPointDate.Before(ticks) {
				if b.Stage == stage.Egg {
					s.death.die(w, e, "Egg: CP!")
				}
			}
		}
	}
}

func (s *broodMortality) Finalize(w *ecs.World) {}
