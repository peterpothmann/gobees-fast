package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// larvaeNeeds calculates per colony the pollen its larvae need today for the largest possible growth, and the
// energy to assimilate it. Read by the foraging stimuli, which set the ideal stores, and these limit the
// feeding in the evening.
//
// NetLogo: NeedNectarPollenLarvaeTodayProc.
type larvaeNeeds struct {
	rng       *rand.Rand
	constants *params.Constants

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	storesMap    *ecs.Map1[comp.ColonyStores]
	beeMap       *ecs.Map1[comp.Bee]
	speciesMap   *speciesTable

	larvae map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *larvaeNeeds) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.storesMap = ecs.NewMap1[comp.ColonyStores](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *larvaeNeeds) Update(w *ecs.World) {
	// ORDER: irrelevant, sums. Shuffled, as NetLogo asks the colonies and the larvae in random order.
	// The larvae are grouped by colony once: nothing changes stage or colony here.
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		pollenNeedMyColony := 0.0 // [g]
		myLarvae := s.larvae[col]
		shuffle(s.rng, myLarvae)
		for _, e := range myLarvae {
			b := s.beeMap.Get(e)
			sp := s.speciesMap.Get(b.Species)
			pollenNeedMyColony = pollenNeedMyColony + float64(b.Number)*(maxWeightGainToday(b, sp)/sp.PollenToBodymassFactor)/1000
		}
		stores := s.storesMap.Get(col)
		stores.LarvaePollenNeedToday = pollenNeedMyColony
		// NOTE: despite its name, EnergyNeedToday is only the energy for the larvae's pollen. Incubation and flights take
		// energy from the store elsewhere.
		stores.EnergyNeedToday = stores.LarvaePollenNeedToday * s.constants.Colony.PollenAssimilationEnergy
	}
}

func (s *larvaeNeeds) Finalize(w *ecs.World) {}

// maxWeightGainToday returns the largest weight gain of a larva in one day [mg]: weight * (growthFactor
// - 1), but not beyond the maximum pupation weight of its caste. NetLogo calls it only for the larva itself,
// so it takes the larva and its species.
//
// NetLogo: MaxWeightGainToday_mg_REP.
func maxWeightGainToday(b *comp.Bee, sp *params.Species) float64 {
	memoQPupationMax := sp.QueenPupationWeightMax
	memoWPupationMax := sp.PupationWeightMax
	// float64(...) forces rounding after the multiplication, as in Java (no fused multiply-add, see
	// competitionPointDate).
	maxWeightGain := float64(b.Weight*sp.GrowthFactor) - b.Weight // [mg]
	if b.Caste == caste.Queen && b.Weight+maxWeightGain > memoQPupationMax {
		maxWeightGain = memoQPupationMax - b.Weight
	}
	if (b.Caste == caste.Worker || b.Caste == caste.Undefined || b.Caste == caste.Male) && b.Weight+maxWeightGain > memoWPupationMax {
		maxWeightGain = memoWPupationMax - b.Weight
	}
	return maxWeightGain
}
