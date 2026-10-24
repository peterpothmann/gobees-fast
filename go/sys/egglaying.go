package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// eggLayer lays eggs [EGG], [EGGP].
// NetLogo calls EgglayingProc from ActivityProc (step 5).
// Until then only the tests call it.
// Must not be called while a query is open (it creates entities).
//
// NetLogo: EgglayingProc, EggsParameterSettingProc.
type eggLayer struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	file      *params.ParameterFile
	constants *params.Constants
	counters  *globals.Counters
	who       *globals.WhoCounter

	beeMap     *beeMapper
	colonyMap  *ecs.Map3[comp.Colony, comp.ColonyStores, comp.ColonyCounts]
	speciesMap *ecs.Map1[comp.Species]
}

func newEggLayer(w *ecs.World, rng *rand.Rand) eggLayer {
	return eggLayer{
		rng:        rng,
		ticks:      ecs.GetResource[globals.Ticks](w),
		file:       ecs.GetResource[params.ParameterFile](w),
		constants:  ecs.GetResource[params.Constants](w),
		counters:   ecs.GetResource[globals.Counters](w),
		who:        ecs.GetResource[globals.WhoCounter](w),
		beeMap:     newBeeMapper(w),
		colonyMap:  ecs.NewMap3[comp.Colony, comp.ColonyStores, comp.ColonyCounts](w),
		speciesMap: ecs.NewMap1[comp.Species](w),
	}
}

// EgglayingProc lets the bee e (the mother) lay one batch of eggs if the colony stores cover the cost, and takes
// 24 h of her time [EGG].
func (l *eggLayer) EgglayingProc(e ecs.Entity) {
	_, _, mother, _, _, _, _ := l.beeMap.Get(e)
	sp := l.speciesMap.Get(mother.SpeciesID)
	colonyID := mother.ColonyID

	// [EGG-01] individual-based colonies pay for one egg (finding 43, see trans/netlogo-findings.md#finding-43)
	motherCaste := mother.Caste
	newCohorts := 1.0
	beesInCohort := sp.Batchsize
	eggWeight := sp.DevWeightEgg_mg
	pollenToMass := sp.PollenToBodymassFactor
	colony, stores, _ := l.colonyMap.Get(colonyID)
	if !colony.CohortBased {
		newCohorts = sp.Batchsize
		beesInCohort = 1
	}
	pollenCost_g := beesInCohort * eggWeight * pollenToMass / 1000
	energyCost_kJ := pollenCost_g * l.constants.EnergyRequiredForPollenAssimilation_kJ_per_g

	// [EGG-02]
	if stores.PollenStore_g > pollenCost_g && stores.EnergyStore_kJ > energyCost_kJ {
		eggs := l.hatch(e, int(newCohorts))
		// ORDER: irrelevant, each egg only sets itself. Shuffled (decision C).
		util.Shuffle(l.rng, eggs)
		ticks := float64(l.ticks.Ticks)
		for _, egg := range eggs {
			colony, _, _ := l.colonyMap.Get(colonyID)
			_, _, b, _, _, _, _ := l.beeMap.Get(egg)
			if motherCaste == caste.Queen {
				if ticks > colony.SwitchPointDate {
					b.Ploidy = 1
				} else {
					b.Ploidy = 2
				}
			} else {
				b.Ploidy = 1 // never run: workers do not lay eggs, see the annotation
			}
			l.EggsParameterSettingProc(egg, beesInCohort)
		}
		_, stores, _ := l.colonyMap.Get(colonyID)
		stores.PollenStore_g = stores.PollenStore_g - pollenCost_g
		stores.EnergyStore_kJ = stores.EnergyStore_kJ - energyCost_kJ
	}

	// [EGG-03] also if no eggs were laid
	_, _, mother, _, _, _, _ = l.beeMap.Get(e)
	mother.PersonalTime_s = mother.PersonalTime_s + 24*3600
}

// hatch creates n copies of the bee e with all its values and a new who number each, and returns them in the
// order of creation. NetLogo: hatch n.
func (l *eggLayer) hatch(e ecs.Entity, n int) []ecs.Entity {
	_, pos, bee, gen, th, forager, dev := l.beeMap.Get(e)
	// copy the values first: creating entities can move the components in memory
	p, b, g, t, f, d := *pos, *bee, *gen, *th, *forager, *dev
	result := make([]ecs.Entity, 0, n)
	for range n {
		gc, fc := g, f
		gc.AllelesList = slices.Clone(g.AllelesList)
		gc.SpermathecaList = slices.Clone(g.SpermathecaList)
		fc.KnownMasterpatchesNectarList = slices.Clone(f.KnownMasterpatchesNectarList)
		fc.KnownMasterpatchesPollenList = slices.Clone(f.KnownMasterpatchesPollenList)
		pc, bc, tc, dc := p, b, t, d
		result = append(result, l.beeMap.NewEntity(&comp.Who{Who: l.who.Next()}, &pc, &bc, &gc, &tc, &fc, &dc))
	}
	return result
}

// EggsParameterSettingProc sets the alleles, the caste and the start values of a new egg (cohort) [EGGP]. The egg
// still holds all values of its mother, and its ploidy is set.
func (l *eggLayer) EggsParameterSettingProc(e ecs.Entity, beesInCohort float64) {
	_, _, b, g, th, f, dev := l.beeMap.Get(e)
	sp := l.speciesMap.Get(b.SpeciesID)

	// [EGGP-01] one allele of the mother; diploid: one of the spermatheca in front (item 0 = father)
	myAllelesList := []float64{g.AllelesList[l.rng.IntN(len(g.AllelesList))]}
	g.AllelesList = myAllelesList
	b.Caste = caste.Undefined
	if b.Ploidy == 1 {
		b.Caste = caste.Male
	} else {
		if b.Ploidy != 2 {
			panic("Wrong ploidy! (EggsParameterSettingProc)")
		}
		g.AllelesList = append([]float64{g.SpermathecaList[l.rng.IntN(len(g.SpermathecaList))]}, g.AllelesList...)
		if l.file.SexLocus && g.AllelesList[0] == g.AllelesList[1] {
			b.Caste = caste.Male
		}
	}

	// [EGGP-02], [EGGP-03] kept from the mother: ColonyID, SpeciesID, SpeciesName, MtDNA (and the position, GUI
	// only). activityList: GUI log, not translated.
	g.SpermathecaList = []float64{}
	b.Number = beesInCohort
	l.counters.TotalBeesEverProduced += b.Number
	b.Activity = activity.Resting
	b.AdultAge = 0
	b.Brood = true
	b.BroodAge = 0
	dev.CumulTimeEgg_d = 0
	dev.CumulTimeLarva_d = 0
	dev.CumulTimePupa_d = 0
	f.Cropvolume_myl = 0
	f.PollenPellets_g = 0
	f.CurrentFoodsource = ecs.Entity{}
	f.NectarsourceToGoTo = ecs.Entity{}
	f.PollensourceToGoTo = ecs.Entity{}
	b.Stage = stage.Egg
	b.Mated = false
	*th = thresholds(sitEgg)
	f.KnownMasterpatchesNectarList = []ecs.Entity{}
	f.KnownMasterpatchesPollenList = []ecs.Entity{}
	b.Weight_mg = sp.DevWeightEgg_mg
	dev.CumulIncubationReceived_kJ = 0
	b.EmergingDate = params.NotSetLow
	f.Expectation_NectarTrip_s = 0
	f.Expectation_PollenTrip_s = 0
	f.GlossaLength_mm = 0
	f.NectarLoadSquadron_kJ = 0
	b.PersonalTime_s = 0
	f.PollenForager = false
	f.PollenLoadSquadron_g = 0
	_, _, counts := l.colonyMap.Get(b.ColonyID)
	counts.TotalEggsProduced += beesInCohort
}
