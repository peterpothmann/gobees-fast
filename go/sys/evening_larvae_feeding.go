package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// larvaeFeeding feeds the larvae. Per colony, the larvae grow by a share of their largest possible gain. The share
// is set by the store (pollen or energy) that is emptier relative to its ideal size. The colony pays the pollen and
// the energy to assimilate it.
//
// NetLogo: FeedLarvaeProc.
type larvaeFeeding struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	constants *params.Constants

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyMap    *ecs.Map1[comp.ColonyStores]
	beeMap       *ecs.Map1[comp.Bee]
	speciesMap   *speciesTable

	larvae map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *larvaeFeeding) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.constants = ecs.GetResource[params.Constants](w)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyMap = ecs.NewMap1[comp.ColonyStores](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *larvaeFeeding) Update(w *ecs.World) {
	// ORDER of the colonies: irrelevant, each colony only feeds its own larvae from its own stores. Shuffled,
	// as NetLogo asks the colonies in random order. The larvae are grouped by colony once: nothing changes stage or
	// colony here.
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		s.feedColony(col, s.larvae[col])
	}
}

func (s *larvaeFeeding) Finalize(w *ecs.World) {}

// feedColony runs for one colony.
func (s *larvaeFeeding) feedColony(col ecs.Entity, myLarvae []ecs.Entity) {
	stores := s.colonyMap.Get(col)

	// both 0 if an ideal store is 0: no feeding
	relativePollenToBeFed := 0.0
	relativeEnergyToBeFed := 0.0
	if stores.IdealPollen*stores.IdealEnergy > 0 {
		relativePollenToBeFed = stores.Pollen / stores.IdealPollen
		relativeEnergyToBeFed = stores.Energy / stores.IdealEnergy
		relativePollenToBeFed = min(max(relativePollenToBeFed, 0), 1)
		relativeEnergyToBeFed = min(max(relativeEnergyToBeFed, 0), 1)
	}
	growthLimitingFactor := relativePollenToBeFed
	if relativeEnergyToBeFed < relativePollenToBeFed {
		growthLimitingFactor = relativeEnergyToBeFed
	}

	// ORDER: irrelevant, the store does not change within the loop, each larva changes only itself.
	// The cap compares the pollen of ONE larva with the whole store, as in NetLogo, so the store can become negative.
	totalPollenFedToday := 0.0 // [g]
	shuffle(s.rng, myLarvae)
	for _, e := range myLarvae {
		b := s.beeMap.Get(e)
		sp := s.speciesMap.Get(b.Species)
		pollenReceivedToday := growthLimitingFactor * (maxWeightGainToday(b, sp) / sp.PollenToBodymassFactor) // [mg]
		if pollenReceivedToday > stores.Pollen*1000 {
			pollenReceivedToday = stores.Pollen * 1000
			if pollenReceivedToday < 0 {
				pollenReceivedToday = 0
			}
		}
		oldWeight := b.Weight // [mg]
		// float64(...): no fused multiply-add, so the result equals NetLogo, see competitionPointDate
		b.Weight = b.Weight + float64(pollenReceivedToday*sp.PollenToBodymassFactor)
		if b.Weight < 0 {
			panic("BUG in feedColony")
		}
		if b.Weight < oldWeight {
			panic("BUG in feedColony: WeightLoss")
		}
		totalPollenFedToday = totalPollenFedToday + (float64(b.Number) * pollenReceivedToday / 1000)
	}

	totalEnergyFedToday := totalPollenFedToday * s.constants.PollenAssimilationEnergy // [kJ]
	stores.Pollen = stores.Pollen - totalPollenFedToday
	if stores.Pollen < 0 {
		// message as in NetLogo, with the Ark entity ID of the colony instead of its who number
		fmt.Printf("negative pollen store! Ticks: (colony %d): %d\n", col.ID(), s.ticks.Ticks)
	}
	stores.Energy = stores.Energy - totalEnergyFedToday
}
