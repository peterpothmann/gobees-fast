package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// NeedNectarPollenLarvaeToday [GO-05].
//
// Calculates per colony the pollen its larvae need today for the largest possible growth, and the energy to
// assimilate it [NEED]. Read by the foraging stimuli (step 5), which set the ideal stores, and these limit the
// feeding in the evening [FEED-01].
//
// NetLogo: NeedNectarPollenLarvaeTodayProc.
type NeedNectarPollenLarvaeToday struct {
	rng       *rand.Rand
	constants *params.Constants

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
	storesMap    *ecs.Map1[comp.ColonyStores]
	beeMap       *ecs.Map1[comp.Bee]
	speciesMap   *ecs.Map1[comp.Species]

	larvae map[ecs.Entity][]ecs.Entity // buffer of beesByColony
}

func (s *NeedNectarPollenLarvaeToday) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)

	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.storesMap = ecs.NewMap1[comp.ColonyStores](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
}

func (s *NeedNectarPollenLarvaeToday) Update(w *ecs.World) {
	// [NEED-01] ORDER: irrelevant, sums. Shuffled (decision C).
	// The larvae are grouped by colony once: nothing changes stage or colony here.
	s.larvae = beesByColony(s.larvae, s.beeFilter, func(b *comp.Bee) bool { return b.Stage == stage.Larva })
	for _, col := range askColonies(s.colonyFilter, s.rng) {
		pollenNeedMyColony_g := 0.0
		myLarvae := s.larvae[col]
		util.Shuffle(s.rng, myLarvae)
		for _, e := range myLarvae {
			b := s.beeMap.Get(e)
			sp := s.speciesMap.Get(b.SpeciesID)
			pollenNeedMyColony_g = pollenNeedMyColony_g + float64(b.Number)*(MaxWeightGainToday_mg_REP(b, sp)/sp.PollenToBodymassFactor)/1000
		}
		stores := s.storesMap.Get(col)
		stores.PollenNeedLarvaeToday_g = pollenNeedMyColony_g
		// NOTE: only the energy for the larvae's pollen, see the annotation.
		stores.EnergyNeedToday_kJ = stores.PollenNeedLarvaeToday_g * s.constants.EnergyRequiredForPollenAssimilation_kJ_per_g
	}
}

func (s *NeedNectarPollenLarvaeToday) Finalize(w *ecs.World) {}

// MaxWeightGainToday_mg_REP returns the largest weight gain of a larva in one day [WGAIN]: weight * (growthFactor
// - 1), but not beyond the maximum pupation weight of its caste. NetLogo calls it only for the larva itself
// (myID = who), so it takes the larva and its species.
func MaxWeightGainToday_mg_REP(b *comp.Bee, sp *comp.Species) float64 {
	// [WGAIN-01]
	memoQPupationMax := sp.DevWeight_Q_PupationMax_mg
	memoWPupationMax := sp.DevWeightPupationMax_mg
	// float64(...) forces rounding after the multiplication, as in Java (no fused multiply-add, see
	// CompetitionPointDateREP).
	maxWeightGain_mg := float64(b.Weight_mg*sp.GrowthFactor) - b.Weight_mg
	if b.Caste == caste.Queen && b.Weight_mg+maxWeightGain_mg > memoQPupationMax {
		maxWeightGain_mg = memoQPupationMax - b.Weight_mg
	}
	if (b.Caste == caste.Worker || b.Caste == caste.Undefined || b.Caste == caste.Male) && b.Weight_mg+maxWeightGain_mg > memoWPupationMax {
		maxWeightGain_mg = memoWPupationMax - b.Weight_mg
	}
	return maxWeightGain_mg
}
