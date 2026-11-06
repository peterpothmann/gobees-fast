package sys

import (
	"math"
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// emergenceWorld is a small world for the tests of queenEmergence: B_terrestris, two nest sites and the resources
// that queenEmergence needs. Scaling 0.1 patches per metre (10 m per patch) as in the example farm.
type emergenceWorld struct {
	w          *ecs.World
	sys        *queenEmergence
	ticks      *globals.Ticks
	parameters *params.Parameters
	derived    *params.Derived
	tally      *tally
	active     *globals.ActiveBees
	species    int // index of B_terrestris
	sp         *params.Species
	ns         *globals.SpeciesNestSites // nest sites of B_terrestris
	bigSite    ecs.Entity                // Grassland, 10000 m2 at (100, 100)
	smallSite  ecs.Entity                // Hedgerow, 100 m2 at (50.3, 50.2), radius 0.564 patches: only the patch centre (50, 50)
}

func newEmergenceWorld(t *testing.T, seed uint64) *emergenceWorld {
	t.Helper()
	w := ecs.NewWorld()
	ew := &emergenceWorld{w: w}
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(seed, 1)})
	p := params.Default()
	ew.parameters = &p.Parameters
	ew.parameters.MergeHedges = false
	ecs.AddResource(w, &p.World)
	ecs.AddResource(w, &p.Constants)
	ecs.AddResource(w, ew.parameters)
	ew.ticks = &globals.Ticks{Ticks: 10}
	ecs.AddResource(w, ew.ticks)
	ew.derived = &params.Derived{ForagerMortality: 0}
	ecs.AddResource(w, ew.derived)
	ecs.AddResource(w, &globals.Landscape{Scale: 0.1})
	ew.tally = newTally(w)
	ew.active = &globals.ActiveBees{}
	ecs.AddResource(w, ew.active)

	fsMap := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)
	ew.bigSite = fsMap.NewEntity(&comp.Position{X: 100, Y: 100},
		&comp.Foodsource{PatchType: "Grassland", Area: 10000, Radius: math.Sqrt(10000 / math.Pi)},
		&comp.FoodsourceState{}, &comp.Layer{IsMasterpatch: true})
	ew.smallSite = fsMap.NewEntity(&comp.Position{X: 50.3, Y: 50.2},
		&comp.Foodsource{PatchType: "Hedgerow", Area: 100, Radius: math.Sqrt(100 / math.Pi)},
		&comp.FoodsourceState{}, &comp.Layer{IsMasterpatch: true})

	in := &params.Input{Species: []params.Species{*terrestris(t)}}
	ecs.AddResource(w, in)
	ew.species = 0
	ew.sp = &in.Species[ew.species]
	nestSites := &globals.NestSites{BySpecies: []globals.SpeciesNestSites{
		{Sites: []ecs.Entity{ew.bigSite, ew.smallSite}, Area: 10100}}}
	ecs.AddResource(w, nestSites)
	ew.ns = &nestSites.BySpecies[ew.species]

	ew.sys = &queenEmergence{}
	ew.sys.Initialize(w)
	return ew
}

// addQueen adds a hibernating queen that emerges at the given tick.
func (ew *emergenceWorld) addQueen(emergingDate day.Abs, number int, weight float64) ecs.Entity {
	bee := comp.Bee{Number: number, Stage: stage.Adult, Caste: caste.Queen, Activity: activity.Hibernate, Mated: true,
		Species: ew.species, EmergenceDate: day.At(emergingDate), Weight: weight}
	gen := comp.Genetics{Alleles: []float64{1, 2}, Spermatheca: []float64{3}}
	return newBeeMapper(ew.w).NewEntity(&comp.Position{},
		&bee, &gen, &comp.Thresholds{}, &comp.Forager{}, &comp.Development{})
}

// bees returns all bee entities with their Bee component.
func (ew *emergenceWorld) bees() map[ecs.Entity]comp.Bee {
	result := map[ecs.Entity]comp.Bee{}
	query := ecs.NewFilter1[comp.Bee](ew.w).Query()
	for query.Next() {
		result[query.Entity()] = *query.Get()
	}
	return result
}

func TestEmergenceSplitsCohort(t *testing.T) {
	ew := newEmergenceWorld(t, 1)
	ew.sp.DailyNestSiteChance = 0 // nobody finds a nest, and with mortality 0 nobody dies
	q := ew.addQueen(10, 3, 700)
	ew.addQueen(11, 1, 700) // emerges tomorrow
	ew.sys.Update(ew.w)

	// the cohort of 3 is split into 3 queens with number 1, identical copies
	beeMap := newBeeMapper(ew.w)
	_, orig, origGen, _, _, _ := beeMap.Get(q)
	emerged := 0
	for e, b := range ew.bees() {
		if b.EmergenceDate != day.At(10) {
			if b.Activity != activity.Hibernate {
				t.Error("the queen of tomorrow must still hibernate")
			}
			continue
		}
		emerged++
		_, _, gen, th, _, _ := beeMap.Get(e)
		// resting, no colony, thresholds of the initiation phase
		if b.Number != 1 || b.Weight != orig.Weight || b.Activity != activity.Resting || !b.Colony.IsZero() ||
			*th != thresholds(sitQueenInitiationPhase) || gen.Alleles[1] != origGen.Alleles[1] {
			t.Errorf("queen %+v %+v", b, *th)
		}
	}
	if emerged != 3 {
		t.Fatalf("%d emerged queens, want 3", emerged)
	}
	// the copies do not share their lists with the original
	origGen.Alleles[0] = 99
	for e := range ew.bees() {
		if _, _, gen, _, _, _ := beeMap.Get(e); e != q && gen.Alleles[0] == 99 {
			t.Error("a copy shares its allele list with the original")
		}
	}
}

func TestEmergenceWinterMortality(t *testing.T) {
	ew := newEmergenceWorld(t, 2)
	ew.parameters.WinterMortality = true
	ew.sp.DailyNestSiteChance = 0
	// relative weight 0.32: survival probability 0.32
	const n = 2000
	for range n {
		ew.addQueen(10, 1, 590+0.32*390)
	}
	ew.sys.Update(ew.w)
	survivors := len(ew.bees())
	// expected 640, standard deviation sqrt(2000 * 0.32 * 0.68) = 20.9
	if math.Abs(float64(survivors)-640) > 4*20.9 {
		t.Errorf("%d of %d queens survived, want about 640", survivors, n)
	}
	if ew.tally.beesDied != n-survivors {
		t.Errorf("%v bees died, want %d", ew.tally.beesDied, n-survivors)
	}
}

func TestNestSearchFoundsColonies(t *testing.T) {
	ew := newEmergenceWorld(t, 3)
	ew.sp.DailyNestSiteChance = 1
	const n = 1000
	for range n {
		ew.addQueen(10, 1, 700)
	}
	ew.sys.Update(ew.w)

	colonyMap := ecs.NewMap3[comp.Position, comp.Colony, comp.ColonyStores](ew.w)
	beeMap := newBeeMapper(ew.w)
	atSmall := 0
	colonies := map[ecs.Entity]bool{}
	for e, b := range ew.bees() {
		// every queen has a colony and rests
		if b.Colony.IsZero() || b.Activity != activity.Resting {
			t.Fatalf("queen without colony: %+v", b)
		}
		colonies[b.Colony] = true
		qPos, _, _, _, _, _ := beeMap.Get(e)
		cPos, col, stores := colonyMap.Get(b.Colony)
		// the colony is at the queen's position, a patch centre
		if *cPos != *qPos || cPos.X != math.Trunc(cPos.X) || cPos.Y != math.Trunc(cPos.Y) {
			t.Fatalf("colony at %v, queen at %v", *cPos, *qPos)
		}
		// the small hedge has only one patch centre in its circle: (50, 50)
		if cPos.X < 75 {
			atSmall++
			if *cPos != (comp.Position{X: 50, Y: 50}) {
				t.Errorf("colony at the small hedge at %v, want (50, 50)", *cPos)
			}
		} else if math.Hypot(cPos.X-100, cPos.Y-100) >= math.Sqrt(10000/math.Pi)*0.1 {
			t.Errorf("colony at %v outside the circle of the grassland", *cPos)
		}
		if col.QueenProduction || col.SwitchPointDate.IsSet() ||
			col.Species != ew.species || !almostEqual(stores.Energy, 0.873) {
			t.Fatalf("colony %+v %+v", col, stores)
		}
	}
	if len(colonies) != n {
		t.Errorf("%d colonies, want %d", len(colonies), n)
	}
	// area-weighted: 100 / 10100 = 1 % at the small hedge, about 10 of 1000 (sd 3.1)
	if atSmall > 25 {
		t.Errorf("%d colonies at the small hedge, want about 10", atSmall)
	}
}

func TestNestSearchMortality(t *testing.T) {
	ew := newEmergenceWorld(t, 4)
	ew.sp.DailyNestSiteChance = 0
	ew.derived.ForagerMortality = 1e-5 // "high"
	const n = 2000
	for range n {
		ew.addQueen(10, 1, 700)
	}
	ew.sys.Update(ew.w)
	// 1 - (1 - 1e-5) ^ 21600 = 0.194, expected 388 deaths, sd 17.7
	deaths := n - len(ew.bees())
	want := (1 - math.Pow(1-1e-5, 6*3600)) * n
	if math.Abs(float64(deaths)-want) > 4*17.7 {
		t.Errorf("%d queens died while searching, want about %.0f", deaths, want)
	}
}

func TestNestSearchWithoutNestSites(t *testing.T) {
	ew := newEmergenceWorld(t, 5)
	ew.ns.Sites = nil
	ew.addQueen(10, 1, 700)
	ew.addQueen(10, 1, 700)
	ew.sys.Update(ew.w)
	// a species without nest sites: the queens die
	if len(ew.bees()) != 0 || ew.tally.beesDied != 2 {
		t.Errorf("%d queens left, %v bees died; want 0, 2", len(ew.bees()), ew.tally.beesDied)
	}
}

func TestNestSearchNoPatchCentre(t *testing.T) {
	// the hedge of TestPatchCentresWithin, 100 m2 at (133.548, 112.5), has no patch centre in its circle (radius
	// 0.564 patches). NetLogo stops with a runtime error. Here the colony is placed at the centre of the patch cell of
	// the foodsource. y = 112.5 lies on the border between two cells, NetLogo rounds .5 up: (134, 113).
	ew := newEmergenceWorld(t, 6)
	hedge := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](ew.w).NewEntity(&comp.Position{X: 133.548, Y: 112.5},
		&comp.Foodsource{PatchType: "Hedgerow", Area: 100, Radius: math.Sqrt(100 / math.Pi)},
		&comp.FoodsourceState{}, &comp.Layer{IsMasterpatch: true})
	ew.ns.Sites = []ecs.Entity{hedge}
	ew.ns.Area = 100
	ew.sp.DailyNestSiteChance = 1
	ew.addQueen(10, 1, 700)
	ew.sys.Update(ew.w)

	colonyMap := ecs.NewMap1[comp.Position](ew.w)
	for _, b := range ew.bees() {
		if b.Colony.IsZero() {
			t.Fatalf("queen without colony: %+v", b)
		}
		if pos := colonyMap.Get(b.Colony); *pos != (comp.Position{X: 134, Y: 113}) {
			t.Errorf("colony at %v, want (134, 113)", *pos)
		}
	}
}

func TestNestSiteChoicePastEnd(t *testing.T) {
	// the shares sum to 0.5, so p lies above the sum in about half of the
	// draws. The loop runs past the end and the last site of the list is chosen (NetLogo: runtime error).
	ew := newEmergenceWorld(t, 7)
	ew.ns.Sites = []ecs.Entity{ew.bigSite}
	ew.ns.Area = 20000
	for range 100 {
		if got := ew.sys.chooseNestSite(ew.ns); got != ew.bigSite {
			t.Fatalf("chooseNestSite = %v, want the grassland %v", got, ew.bigSite)
		}
	}
}

func TestDieRemovesFromActiveBees(t *testing.T) {
	ew := newEmergenceWorld(t, 6)
	a := ew.addQueen(0, 5, 700)
	b := ew.addQueen(0, 1, 700)
	ew.active.List = []ecs.Entity{a, b}
	ew.sys.death.die(ew.w, a, "test")
	// TotalBeesEverDied + number, removed from the activity queue and from the world
	if ew.tally.beesDied != 5 || len(ew.active.List) != 1 || ew.active.List[0] != b || ew.w.Alive(a) {
		t.Errorf("%v bees died, queue %v, alive %v", ew.tally.beesDied, ew.active.List, ew.w.Alive(a))
	}
}

func TestPatchesInRangeAndFlowerLists(t *testing.T) {
	ew := newEmergenceWorld(t, 8)
	w := ew.w
	fsMap := ecs.NewMap4[comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)
	// colony at (0, 0), range 758 m * 0.1 = 75.8 patches
	newFs := func(x, pollen, nectar float64, master ecs.Entity) ecs.Entity {
		e := fsMap.NewEntity(&comp.Position{X: x, Y: 0}, &comp.Foodsource{},
			&comp.FoodsourceState{Pollen: pollen, Nectar: nectar}, &comp.Layer{Masterpatch: master})
		if master.IsZero() {
			_, _, _, l := fsMap.Get(e)
			l.Masterpatch = e
		}
		return e
	}
	a := newFs(70, 1, 0, ecs.Entity{}) // pollen only
	b := newFs(75.8, 0, 1, a)          // exactly at the range limit (<=), nectar only, layer of a
	c := newFs(80, 1, 1, ecs.Entity{}) // out of range
	d := newFs(10, 1, 1, ecs.Entity{}) // pollen and nectar
	e := newFs(20, 0, 0, ecs.Entity{}) // in range, not in flower
	_ = c

	colony := ew.sys.colonies.colonyMap.NewEntity(&comp.Position{X: 0, Y: 0}, &comp.Colony{},
		&comp.ColonyStores{}, &comp.ColonyStimuli{}, &comp.ColonyCounts{}, &comp.ColonyRange{})
	ew.sys.colonies.patchesInRange(colony)
	ew.sys.colonies.flowersInRange(colony)
	r := ecs.NewMap1[comp.ColonyRange](w).Get(colony)

	sameSet := func(got, want []ecs.Entity) bool {
		if len(got) != len(want) {
			return false
		}
		m := map[ecs.Entity]int{}
		for _, x := range got {
			m[x]++
		}
		for _, x := range want {
			if m[x] != 1 {
				return false
			}
		}
		return true
	}
	// also the small hedge of newEmergenceWorld at (50.3, 50.2), distance 71.1 (not in flower).
	// The grassland at (100, 100) is at distance 141.4.
	if !sameSet(r.Layers, []ecs.Entity{a, b, d, e, ew.smallSite}) {
		t.Errorf("in range: %v", r.Layers)
	}
	checks := map[string][2][]ecs.Entity{
		"pollen masters": {r.PollenPatches, {a, d}},
		"nectar masters": {r.NectarPatches, {a, d}}, // b is a layer of a
	}
	for name, c := range checks {
		if !sameSet(c[0], c[1]) {
			t.Errorf("%s: got %v, want %v", name, c[0], c[1])
		}
	}
}

func TestExampleFarmFirstYear(t *testing.T) {
	// One year with initialQueensSetup, tick and queenEmergence only: every initial queen has either founded a colony,
	// died (winter or nest search), or is still searching or hibernating at the end.
	a := newExampleFarmSetup(t, 11, &initialQueensSetup{}, &tick{}, &queenEmergence{})
	w := a.World
	tally := newTally(w) // Setup kills no bee
	for range 365 {
		a.Update()
	}
	// without colonyUpdate no colony dies: the living colonies are all colonies founded
	colonies := ecs.NewFilter1[comp.Colony](w).Query()
	founded := colonies.Count()
	colonies.Close()
	searching, hibernating := 0, 0
	query := ecs.NewFilter1[comp.Bee](w).Query()
	for query.Next() {
		b := query.Get()
		switch {
		case b.Colony.IsZero() && b.Activity == activity.Hibernate:
			hibernating++
		case b.Colony.IsZero():
			searching++
		}
	}
	died := tally.beesDied
	t.Logf("founded %v, died %v, still searching %d, hibernating %d", founded, died, searching, hibernating)
	if founded+died+searching+hibernating != 500 {
		t.Errorf("founded %v + died %v + searching %d + hibernating %d != 500", founded, died, searching, hibernating)
	}
	// about 60 % survive the winter, about 56 % of them found a colony
	if founded < 100 || founded > 250 {
		t.Errorf("%v colonies founded, expected roughly 170", founded)
	}
}
