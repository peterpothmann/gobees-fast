package sys

import (
	"math"
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// emergenceWorld is a small world for the tests of step 3: B_terrestris, two nest sites and the resources
// that EmergenceNewQueens needs. Scaling 0.1 patches per metre (10 m per patch) as in the example farm.
type emergenceWorld struct {
	w         *ecs.World
	sys       *EmergenceNewQueens
	ticks     *globals.Ticks
	file      *params.ParameterFile
	derived   *globals.Derived
	counters  *globals.Counters
	active    *globals.ActiveBees
	species   ecs.Entity
	sp        *comp.Species
	bigSite   ecs.Entity // Grassland, 10000 m2 at (100, 100)
	smallSite ecs.Entity // Hedgerow, 100 m2 at (50.3, 50.2), radius 0.564 patches: only the patch centre (50, 50)
}

func newEmergenceWorld(t *testing.T, seed uint64) *emergenceWorld {
	t.Helper()
	w := ecs.NewWorld()
	ew := &emergenceWorld{w: w}
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(seed, 1)})
	p := params.Default()
	ew.file = &p.File
	ew.file.MergeHedges = false
	ecs.AddResource(w, &p.World)
	ecs.AddResource(w, &p.Constants)
	ecs.AddResource(w, ew.file)
	ew.ticks = &globals.Ticks{Ticks: 10}
	ecs.AddResource(w, ew.ticks)
	ew.derived = &globals.Derived{MortalityForager_per_s: 0}
	ecs.AddResource(w, ew.derived)
	ecs.AddResource(w, &globals.Landscape{Scaling_NLpatches_per_m: 0.1})
	ecs.AddResource(w, &globals.WhoCounter{})
	ew.counters = &globals.Counters{}
	ecs.AddResource(w, ew.counters)
	ew.active = &globals.ActiveBees{}
	ecs.AddResource(w, ew.active)

	fsMap := ecs.NewMap5[comp.Who, comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)
	ew.bigSite = fsMap.NewEntity(&comp.Who{Who: 0}, &comp.Position{X: 100, Y: 100},
		&comp.Foodsource{PatchType: "Grassland", Area_sqm: 10000, Radius_m: math.Sqrt(10000 / math.Pi)},
		&comp.FoodsourceState{}, &comp.Layer{Masterpatch: true})
	ew.smallSite = fsMap.NewEntity(&comp.Who{Who: 1}, &comp.Position{X: 50.3, Y: 50.2},
		&comp.Foodsource{PatchType: "Hedgerow", Area_sqm: 100, Radius_m: math.Sqrt(100 / math.Pi)},
		&comp.FoodsourceState{}, &comp.Layer{Masterpatch: true})

	sp := *terrestris(t)
	sp.NestsiteFoodsourceList = []ecs.Entity{ew.bigSite, ew.smallSite}
	sp.NestSiteArea = 10100
	ew.species = ecs.NewMap2[comp.Who, comp.Species](w).NewEntity(&comp.Who{Who: 2}, &sp)
	ew.sp = ecs.NewMap1[comp.Species](w).Get(ew.species)

	ew.sys = &EmergenceNewQueens{}
	ew.sys.Initialize(w)
	return ew
}

// addQueen adds a hibernating queen that emerges at the given tick.
func (ew *emergenceWorld) addQueen(emergingDate, number, weight float64) ecs.Entity {
	bee := comp.Bee{Number: number, Stage: stage.Adult, Caste: caste.Queen, Activity: activity.Hibernate, Mated: true,
		Ploidy: 2, SpeciesID: ew.species, SpeciesName: ew.sp.Name, EmergingDate: emergingDate, Weight_mg: weight}
	gen := comp.Genetics{AllelesList: []float64{1, 2}, SpermathecaList: []float64{3}}
	return newBeeMapper(ew.w).NewEntity(&comp.Who{Who: ecs.GetResource[globals.WhoCounter](ew.w).Next()}, &comp.Position{},
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
	ew.sp.ChanceFindNest = 0 // nobody finds a nest, and with mortality 0 nobody dies
	q := ew.addQueen(10, 3, 700)
	ew.addQueen(11, 1, 700) // emerges tomorrow
	ew.sys.Update(ew.w)

	// [EMER-01] the cohort of 3 is split into 3 queens with number 1, identical copies (finding 32)
	beeMap := newBeeMapper(ew.w)
	_, _, orig, origGen, _, _, _ := beeMap.Get(q)
	emerged := 0
	whos := map[int64]bool{}
	for e, b := range ew.bees() {
		if b.EmergingDate != 10 {
			if b.Activity != activity.Hibernate {
				t.Error("the queen of tomorrow must still hibernate")
			}
			continue
		}
		emerged++
		who, _, _, gen, th, _, _ := beeMap.Get(e)
		whos[who.Who] = true
		// [EMER-02] resting, no colony; thresholds of the initiation phase
		if b.Number != 1 || b.Weight_mg != orig.Weight_mg || b.Activity != activity.Resting || !b.ColonyID.IsZero() ||
			*th != thresholds(sitQueenInitiationPhase) || gen.AllelesList[1] != origGen.AllelesList[1] {
			t.Errorf("queen %+v %+v", b, *th)
		}
	}
	if emerged != 3 || len(whos) != 3 {
		t.Fatalf("%d emerged queens with %d who numbers, want 3", emerged, len(whos))
	}
	// the copies do not share their lists with the original
	origGen.AllelesList[0] = 99
	for e := range ew.bees() {
		if _, _, _, gen, _, _, _ := beeMap.Get(e); e != q && gen.AllelesList[0] == 99 {
			t.Error("a copy shares its allele list with the original")
		}
	}
}

func TestEmergenceWinterMortality(t *testing.T) {
	ew := newEmergenceWorld(t, 2)
	ew.file.WinterMortality = true
	ew.sp.ChanceFindNest = 0
	// relative weight 0.32: survival probability 0.32 (finding 33)
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
	if ew.counters.TotalBeesEverDied != float64(n-survivors) {
		t.Errorf("TotalBeesEverDied = %v, want %d", ew.counters.TotalBeesEverDied, n-survivors)
	}
}

func TestNestSearchFoundsColonies(t *testing.T) {
	ew := newEmergenceWorld(t, 3)
	ew.sp.ChanceFindNest = 1
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
		// [CCOL-03] every queen has a colony and rests
		if b.ColonyID.IsZero() || b.Activity != activity.Resting {
			t.Fatalf("queen without colony: %+v", b)
		}
		colonies[b.ColonyID] = true
		_, qPos, _, _, _, _, _ := beeMap.Get(e)
		cPos, col, stores := colonyMap.Get(b.ColonyID)
		// the colony is at the queen's position, a patch centre (finding 25)
		if *cPos != *qPos || cPos.X != math.Trunc(cPos.X) || cPos.Y != math.Trunc(cPos.Y) {
			t.Fatalf("colony at %v, queen at %v", *cPos, *qPos)
		}
		// [NEST-03] the small hedge has only one patch centre in its circle: (50, 50)
		if cPos.X < 75 {
			atSmall++
			if *cPos != (comp.Position{X: 50, Y: 50}) {
				t.Errorf("colony at the small hedge at %v, want (50, 50)", *cPos)
			}
		} else if math.Hypot(cPos.X-100, cPos.Y-100) >= math.Sqrt(10000/math.Pi)*0.1 {
			t.Errorf("colony at %v outside the circle of the grassland", *cPos)
		}
		// [CCOL-04]
		if col.QueenProduction || col.SwitchPointDate != params.NotSetHigh ||
			col.SpeciesIDcolony != ew.species || !almostEqual(stores.EnergyStore_kJ, 0.873) {
			t.Fatalf("colony %+v %+v", col, stores)
		}
	}
	if len(colonies) != n || ew.counters.TotalColoniesEverProduced != n {
		t.Errorf("%d colonies, TotalColoniesEverProduced %v, want %d", len(colonies), ew.counters.TotalColoniesEverProduced, n)
	}
	// [NSFS-01] area-weighted: 100 / 10100 = 1 % at the small hedge, about 10 of 1000 (sd 3.1)
	if atSmall > 25 {
		t.Errorf("%d colonies at the small hedge, want about 10", atSmall)
	}
}

func TestNestSearchMortality(t *testing.T) {
	ew := newEmergenceWorld(t, 4)
	ew.sp.ChanceFindNest = 0
	ew.derived.MortalityForager_per_s = 1e-5 // "high"
	const n = 2000
	for range n {
		ew.addQueen(10, 1, 700)
	}
	ew.sys.Update(ew.w)
	// [NEST-04] finding 26: 1 - (1 - 1e-5) ^ 21600 = 0.194, expected 388 deaths, sd 17.7
	deaths := n - len(ew.bees())
	want := (1 - math.Pow(1-1e-5, 6*3600)) * n
	if math.Abs(float64(deaths)-want) > 4*17.7 {
		t.Errorf("%d queens died while searching, want about %.0f", deaths, want)
	}
}

func TestNestSearchWithoutNestSites(t *testing.T) {
	ew := newEmergenceWorld(t, 5)
	ew.sp.NestsiteFoodsourceList = nil
	ew.addQueen(10, 1, 700)
	ew.addQueen(10, 1, 700)
	ew.sys.Update(ew.w)
	// [EMER-04] a species without nest sites: the queens die
	if len(ew.bees()) != 0 || ew.counters.TotalBeesEverDied != 2 {
		t.Errorf("%d queens left, TotalBeesEverDied %v; want 0, 2", len(ew.bees()), ew.counters.TotalBeesEverDied)
	}
}

func TestDieRemovesFromActiveBees(t *testing.T) {
	ew := newEmergenceWorld(t, 6)
	a := ew.addQueen(0, 5, 700)
	b := ew.addQueen(0, 1, 700)
	ew.active.List = []ecs.Entity{a, b}
	ew.sys.death.DieProc(ew.w, a, "test")
	// [DIE-01] TotalBeesEverDied + number, removed from the activity queue and from the world
	if ew.counters.TotalBeesEverDied != 5 || len(ew.active.List) != 1 || ew.active.List[0] != b || ew.w.Alive(a) {
		t.Errorf("TotalBeesEverDied %v, queue %v, alive %v", ew.counters.TotalBeesEverDied, ew.active.List, ew.w.Alive(a))
	}
}

func TestPatchesInRangeAndFlowerLists(t *testing.T) {
	ew := newEmergenceWorld(t, 8)
	w := ew.w
	fsMap := ecs.NewMap5[comp.Who, comp.Position, comp.Foodsource, comp.FoodsourceState, comp.Layer](w)
	// colony at (0, 0); range 758 m * 0.1 = 75.8 patches
	newFs := func(x, pollen, nectar float64, master ecs.Entity) ecs.Entity {
		e := fsMap.NewEntity(&comp.Who{}, &comp.Position{X: x, Y: 0}, &comp.Foodsource{},
			&comp.FoodsourceState{Pollen_g: pollen, Nectar_myl: nectar}, &comp.Layer{MasterpatchID: master})
		if master.IsZero() {
			_, _, _, _, l := fsMap.Get(e)
			l.MasterpatchID = e
		}
		return e
	}
	a := newFs(70, 1, 0, ecs.Entity{}) // pollen only
	b := newFs(75.8, 0, 1, a)          // exactly at the range limit (<=), nectar only, layer of a
	c := newFs(80, 1, 1, ecs.Entity{}) // out of range
	d := newFs(10, 1, 1, ecs.Entity{}) // pollen and nectar
	e := newFs(20, 0, 0, ecs.Entity{}) // in range, not in flower
	_ = c

	colony := ew.sys.colonies.colonyMap.NewEntity(&comp.Who{}, &comp.Position{X: 0, Y: 0}, &comp.Colony{},
		&comp.ColonyStores{}, &comp.ColonyStimuli{}, &comp.ColonyCounts{}, &comp.ColonyRange{})
	ew.sys.colonies.PatchesInRangeProc(colony)
	ew.sys.colonies.FoodsourcesInFlowerAndRangeProc(colony)
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
	// [PRNG-01] also the small hedge of newEmergenceWorld at (50.3, 50.2), distance 71.1 (not in flower);
	// the grassland at (100, 100) is at distance 141.4
	if !sameSet(r.AllPatchesInRangeList, []ecs.Entity{a, b, d, e, ew.smallSite}) {
		t.Errorf("in range: %v", r.AllPatchesInRangeList)
	}
	// [FRNG-01]
	checks := map[string][2][]ecs.Entity{
		"pollen masters": {r.MasterpatchesWithPollenlayersInFlowerAndRangeList, {a, d}},
		"nectar masters": {r.MasterpatchesWithNectarlayersInFlowerAndRangeList, {a, d}}, // b is a layer of a
	}
	for name, c := range checks {
		if !sameSet(c[0], c[1]) {
			t.Errorf("%s: got %v, want %v", name, c[0], c[1])
		}
	}
}

func TestExampleFarmFirstYear(t *testing.T) {
	// One year with the translated parts of steps 1 to 3: every initial queen has either founded a colony,
	// died (winter or nest search), or is still searching or hibernating at the end.
	a := newExampleFarmSetup(t, 11, &CreateInitialQueens{}, &Tick{}, &EmergenceNewQueens{})
	for range 365 {
		a.Update()
	}
	w := a.World
	counters := ecs.GetResource[globals.Counters](w)
	searching, hibernating := 0, 0
	query := ecs.NewFilter1[comp.Bee](w).Query()
	for query.Next() {
		b := query.Get()
		switch {
		case b.ColonyID.IsZero() && b.Activity == activity.Hibernate:
			hibernating++
		case b.ColonyID.IsZero():
			searching++
		}
	}
	founded := counters.TotalColoniesEverProduced
	died := counters.TotalBeesEverDied
	t.Logf("founded %v, died %v, still searching %d, hibernating %d", founded, died, searching, hibernating)
	if founded+died+float64(searching+hibernating) != 500 {
		t.Errorf("founded %v + died %v + searching %d + hibernating %d != 500", founded, died, searching, hibernating)
	}
	// findings 26 and 33: about 60 % survive the winter, about 56 % of them found a colony
	if founded < 100 || founded > 250 {
		t.Errorf("%v colonies founded, expected roughly 170", founded)
	}
}
