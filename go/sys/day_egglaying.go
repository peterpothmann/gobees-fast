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

// eggLayer lays eggs. activities.act calls it.
// Must not be called while a query is open (it creates entities).
//
// NetLogo: EgglayingProc, EggsParameterSettingProc.
type eggLayer struct {
	rng        *rand.Rand
	ticks      *globals.Ticks
	parameters *params.Parameters
	constants  *params.Constants

	beeMap     *beeMapper
	colonyMap  *ecs.Map2[comp.Colony, comp.ColonyStores]
	speciesMap *speciesTable
}

func newEggLayer(w *ecs.World, rng *rand.Rand) eggLayer {
	return eggLayer{
		rng:        rng,
		ticks:      ecs.GetResource[globals.Ticks](w),
		parameters: ecs.GetResource[params.Parameters](w),
		constants:  ecs.GetResource[params.Constants](w),
		beeMap:     newBeeMapper(w),
		colonyMap:  ecs.NewMap2[comp.Colony, comp.ColonyStores](w),
		speciesMap: newSpeciesTable(w),
	}
}

// layEggs lets the bee e (the mother) lay one batch of eggs if the colony stores cover the cost, and takes
// 24 h of her time. It returns the new egg cohorts (none if the stores were too small).
func (l *eggLayer) layEggs(e ecs.Entity) []ecs.Entity {
	_, mother, _, _, _, _ := l.beeMap.Get(e)
	sp := l.speciesMap.Get(mother.Species)
	colonyID := mother.Colony

	// one cohort of batchsize eggs. There are no individual-based colonies, which in NetLogo lay batchsize
	// cohorts of 1 egg.
	motherCaste := mother.Caste
	beesInCohort := sp.BatchSize
	pollenCost := float64(beesInCohort) * sp.EggWeight * sp.PollenToBodymassFactor / 1000 // [g]
	energyCost := pollenCost * l.constants.Colony.PollenAssimilationEnergy                // [kJ]

	var eggs []ecs.Entity
	_, stores := l.colonyMap.Get(colonyID)
	if stores.Pollen > pollenCost && stores.Energy > energyCost {
		// NetLogo shuffles the new cohorts. One cohort draws no random number.
		egg := l.hatch(e)
		eggs = []ecs.Entity{egg}
		// the ploidy is passed to initEggs, its only reader, instead of being stored in the egg: diploid = ploidy 2.
		colony, _ := l.colonyMap.Get(colonyID)
		var diploid bool
		if motherCaste == caste.Queen {
			if colony.SwitchPointDate.Before(l.ticks.Ticks) {
				diploid = false
			} else {
				diploid = true
			}
		} else {
			// never run: workers do not lay eggs, their egg-laying threshold is above every stimulus.
			// SIMPLIFY: remove the branch.
			diploid = false
		}
		l.initEggs(egg, beesInCohort, diploid)
		_, stores := l.colonyMap.Get(colonyID)
		stores.Pollen = stores.Pollen - pollenCost
		stores.Energy = stores.Energy - energyCost
	}

	// also if no eggs were laid
	_, mother, _, _, _, _ = l.beeMap.Get(e)
	mother.PersonalTime = mother.PersonalTime + l.constants.Activities.EggLayingTime
	return eggs
}

// hatch creates a copy of the bee e with all its values.
//
// NetLogo: hatch 1.
func (l *eggLayer) hatch(e ecs.Entity) ecs.Entity {
	pos, bee, gen, th, forager, dev := l.beeMap.Get(e)
	// copy the values first: creating entities can move the components in memory
	p, b, g, t, f, d := *pos, *bee, *gen, *th, *forager, *dev
	g.Alleles = slices.Clone(g.Alleles)
	g.Spermatheca = slices.Clone(g.Spermatheca)
	f.KnownNectarPatches = slices.Clone(f.KnownNectarPatches)
	f.KnownPollenPatches = slices.Clone(f.KnownPollenPatches)
	return l.beeMap.NewEntity(&p, &b, &g, &t, &f, &d)
}

// initEggs sets the alleles, the caste and the start values of a new egg (cohort). The egg
// still holds all values of its mother. diploid is the ploidy of the egg (false: 1, true: 2).
func (l *eggLayer) initEggs(e ecs.Entity, beesInCohort int, diploid bool) {
	_, b, g, th, f, dev := l.beeMap.Get(e)
	sp := l.speciesMap.Get(b.Species)

	// one allele of the mother. Diploid: one of the spermatheca in front (item 0 = father).
	myAllelesList := []float64{g.Alleles[l.rng.IntN(len(g.Alleles))]}
	g.Alleles = myAllelesList
	b.Caste = caste.Undefined
	if !diploid {
		b.Caste = caste.Male
	} else {
		g.Alleles = append([]float64{g.Spermatheca[l.rng.IntN(len(g.Spermatheca))]}, g.Alleles...)
		if l.parameters.SexLocus && g.Alleles[0] == g.Alleles[1] {
			b.Caste = caste.Male
		}
	}

	// kept from the mother: Colony, Species, MtDNA and the position.
	g.Spermatheca = []float64{}
	b.Number = beesInCohort
	b.Activity = activity.Resting
	b.AdultAge = 0
	b.BroodAge = 0
	dev.LarvaTime = 0
	f.CropVolume = 0
	f.PollenPellets = 0
	f.CurrentFoodsource = ecs.Entity{}
	f.NectarSource = ecs.Entity{}
	f.PollenSource = ecs.Entity{}
	b.Stage = stage.Egg
	b.Mated = false
	*th = thresholds(&l.constants.Thresholds, sitEgg)
	f.KnownNectarPatches = []ecs.Entity{}
	f.KnownPollenPatches = []ecs.Entity{}
	b.Weight = sp.EggWeight
	dev.IncubationReceived = 0
	b.EmergenceDate = day.Date{} // unset (NetLogo: NotSetLow), drawn when the queen emerges
	f.ExpectedNectarTrip = 0
	f.ExpectedPollenTrip = 0
	f.GlossaLength = 0
	f.CarriedNectar = 0
	b.PersonalTime = 0
	f.PollenForager = false
	f.CarriedPollen = 0
}
