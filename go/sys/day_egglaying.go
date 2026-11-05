package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

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

	beeMap     *beeMapper
	colonyMap  *ecs.Map2[comp.Colony, comp.ColonyStores]
	speciesMap *speciesTable
}

func newEggLayer(w *ecs.World, rng *rand.Rand) eggLayer {
	return eggLayer{
		rng:        rng,
		ticks:      ecs.GetResource[globals.Ticks](w),
		file:       ecs.GetResource[params.ParameterFile](w),
		constants:  ecs.GetResource[params.Constants](w),
		beeMap:     newBeeMapper(w),
		colonyMap:  ecs.NewMap2[comp.Colony, comp.ColonyStores](w),
		speciesMap: newSpeciesTable(w),
	}
}

// layEggs lets the bee e (the mother) lay one batch of eggs if the colony stores cover the cost, and takes
// 24 h of her time [EGG]. It returns the new egg cohorts (none if the stores were too small).
func (l *eggLayer) layEggs(e ecs.Entity) []ecs.Entity {
	_, mother, _, _, _, _ := l.beeMap.Get(e)
	sp := l.speciesMap.Get(mother.SpeciesID)
	colonyID := mother.ColonyID

	// [EGG-01] one cohort of batchsize eggs. gobees-fast has no individual-based colonies, which lay batchsize
	// cohorts of 1 egg in NetLogo (findings 30, 43, see trans/netlogo-findings.md#finding-43).
	motherCaste := mother.Caste
	beesInCohort := sp.Batchsize
	pollenCost_g := float64(beesInCohort) * sp.DevWeightEgg_mg * sp.PollenToBodymassFactor / 1000
	energyCost_kJ := pollenCost_g * l.constants.EnergyRequiredForPollenAssimilation_kJ_per_g

	// [EGG-02]
	var eggs []ecs.Entity
	_, stores := l.colonyMap.Get(colonyID)
	if stores.PollenStore_g > pollenCost_g && stores.EnergyStore_kJ > energyCost_kJ {
		// NetLogo shuffles the new cohorts (decision C). One cohort draws no random number.
		egg := l.hatch(e)
		eggs = []ecs.Entity{egg}
		// NetLogo sets the ploidy of the egg (bees-own), only EggsParameterSettingProc reads it. gobees-fast passes
		// it as a parameter (step F9): diploid = ploidy 2.
		colony, _ := l.colonyMap.Get(colonyID)
		var diploid bool
		if motherCaste == caste.Queen {
			if colony.SwitchPointDate.Before(l.ticks.Ticks) {
				diploid = false
			} else {
				diploid = true
			}
		} else {
			diploid = false // never run: workers do not lay eggs, see the annotation
		}
		l.initEggs(egg, beesInCohort, diploid)
		_, stores := l.colonyMap.Get(colonyID)
		stores.PollenStore_g = stores.PollenStore_g - pollenCost_g
		stores.EnergyStore_kJ = stores.EnergyStore_kJ - energyCost_kJ
	}

	// [EGG-03] also if no eggs were laid
	_, mother, _, _, _, _ = l.beeMap.Get(e)
	mother.PersonalTime_s = mother.PersonalTime_s + 24*3600
	return eggs
}

// hatch creates a copy of the bee e with all its values. NetLogo: hatch 1 (with a new who number, not kept since
// step F12).
func (l *eggLayer) hatch(e ecs.Entity) ecs.Entity {
	pos, bee, gen, th, forager, dev := l.beeMap.Get(e)
	// copy the values first: creating entities can move the components in memory
	p, b, g, t, f, d := *pos, *bee, *gen, *th, *forager, *dev
	g.AllelesList = slices.Clone(g.AllelesList)
	g.SpermathecaList = slices.Clone(g.SpermathecaList)
	f.KnownMasterpatchesNectarList = slices.Clone(f.KnownMasterpatchesNectarList)
	f.KnownMasterpatchesPollenList = slices.Clone(f.KnownMasterpatchesPollenList)
	return l.beeMap.NewEntity(&p, &b, &g, &t, &f, &d)
}

// initEggs sets the alleles, the caste and the start values of a new egg (cohort) [EGGP]. The egg
// still holds all values of its mother. diploid is the ploidy of [EGG-02] (false: 1, true: 2).
func (l *eggLayer) initEggs(e ecs.Entity, beesInCohort int, diploid bool) {
	_, b, g, th, f, dev := l.beeMap.Get(e)
	sp := l.speciesMap.Get(b.SpeciesID)

	// [EGGP-01] one allele of the mother; diploid: one of the spermatheca in front (item 0 = father)
	myAllelesList := []float64{g.AllelesList[l.rng.IntN(len(g.AllelesList))]}
	g.AllelesList = myAllelesList
	b.Caste = caste.Undefined
	// assertion Wrong ploidy! (neither 1 nor 2): a bool has no third value (step F9)
	if !diploid {
		b.Caste = caste.Male
	} else {
		g.AllelesList = append([]float64{g.SpermathecaList[l.rng.IntN(len(g.SpermathecaList))]}, g.AllelesList...)
		if l.file.SexLocus && g.AllelesList[0] == g.AllelesList[1] {
			b.Caste = caste.Male
		}
	}

	// [EGGP-02], [EGGP-03] kept from the mother: ColonyID, SpeciesID (speciesName: from the species), MtDNA (and
	// the position, GUI only). activityList: GUI log, not translated. brood? true: the stage egg (step F9).
	g.SpermathecaList = []float64{}
	b.Number = beesInCohort // totalBeesEverProduced: read only by [CHK-01], removed in step F4
	b.Activity = activity.Resting
	b.AdultAge = 0
	b.BroodAge = 0
	dev.CumulTimeLarva_d = 0
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
	b.EmergingDate = day.Date{} // unset (NetLogo: NotSetLow), drawn when the queen emerges [DEVP-04]
	f.Expectation_NectarTrip_s = 0
	f.Expectation_PollenTrip_s = 0
	f.GlossaLength_mm = 0
	f.NectarLoadSquadron_kJ = 0
	b.PersonalTime_s = 0
	f.PollenForager = false
	f.PollenLoadSquadron_g = 0
	// totalEggsProduced of the colony: read only by the egg accounting [CHK-04], removed in step F4
}
