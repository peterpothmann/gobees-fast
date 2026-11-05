package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// larvaeFeeding [GO-13].
//
// Per colony, the larvae grow by a share of their largest possible gain [WGAIN]. The share is set by the store
// (pollen or energy) that is emptier relative to its ideal size. The colony pays the pollen and the energy to
// assimilate it [FEED].
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
	// ORDER of the colonies: irrelevant, each colony only feeds its own larvae from its own stores. Shuffled
	// (decision C). The larvae are grouped by colony once: nothing changes stage or colony here.
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		s.feedColony(col, s.larvae[col])
	}
}

func (s *larvaeFeeding) Finalize(w *ecs.World) {}

// feedColony runs [FEED-01] to [FEED-03] for one colony.
func (s *larvaeFeeding) feedColony(col ecs.Entity, myLarvae []ecs.Entity) {
	stores := s.colonyMap.Get(col)

	// [FEED-01] both 0 if an ideal store is 0: no feeding
	relativePollenToBeFed := 0.0
	relativeEnergyToBeFed := 0.0
	if stores.IdealPollenStore_g*stores.IdealEnergyStore_kJ > 0 {
		relativePollenToBeFed = stores.PollenStore_g / stores.IdealPollenStore_g
		relativeEnergyToBeFed = stores.EnergyStore_kJ / stores.IdealEnergyStore_kJ
		// NetLogo: median (list 0 1 x)
		relativePollenToBeFed = min(max(relativePollenToBeFed, 0), 1)
		relativeEnergyToBeFed = min(max(relativeEnergyToBeFed, 0), 1)
	}
	growthLimitingFactor := relativePollenToBeFed
	if relativeEnergyToBeFed < relativePollenToBeFed {
		growthLimitingFactor = relativeEnergyToBeFed
	}

	// [FEED-02] ORDER: irrelevant, the store does not change within the loop, each larva changes only itself.
	// The cap compares the pollen of ONE larva with the whole store (see the annotation).
	totalPollenFedToday_g := 0.0
	util.Shuffle(s.rng, myLarvae)
	for _, e := range myLarvae {
		b := s.beeMap.Get(e)
		sp := s.speciesMap.Get(b.SpeciesID)
		pollenReceivedToday_mg := growthLimitingFactor * (maxWeightGainToday(b, sp) / sp.PollenToBodymassFactor)
		if pollenReceivedToday_mg > stores.PollenStore_g*1000 {
			pollenReceivedToday_mg = stores.PollenStore_g * 1000
			if pollenReceivedToday_mg < 0 {
				pollenReceivedToday_mg = 0
			}
		}
		oldWeight_mg := b.Weight_mg
		// float64(...): no fused multiply-add, see CompetitionPointDateREP
		b.Weight_mg = b.Weight_mg + float64(pollenReceivedToday_mg*sp.PollenToBodymassFactor)
		if b.Weight_mg < 0 {
			panic("BUG in FeedPOLLENProc")
		}
		if b.Weight_mg < oldWeight_mg {
			panic("BUG in FeedPOLLENProc: WeightLoss")
		}
		totalPollenFedToday_g = totalPollenFedToday_g + (float64(b.Number) * pollenReceivedToday_mg / 1000)
	}

	// [FEED-03]
	totalEnergyFedToday_kJ := totalPollenFedToday_g * s.constants.EnergyRequiredForPollenAssimilation_kJ_per_g
	stores.PollenStore_g = stores.PollenStore_g - totalPollenFedToday_g
	if stores.PollenStore_g < 0 {
		// NetLogo: type "negative pollen store! Ticks: " show ticks (show adds the calling agent, by its who number).
		// gobees-fast names the Ark entity ID of the colony instead (step F12).
		fmt.Printf("negative pollen store! Ticks: (colony %d): %d\n", col.ID(), s.ticks.Ticks)
	}
	stores.EnergyStore_kJ = stores.EnergyStore_kJ - totalEnergyFedToday_kJ
}
