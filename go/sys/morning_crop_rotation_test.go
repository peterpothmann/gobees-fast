package sys

import (
	"maps"
	"math/rand/v2"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/globals"
	"gobees/input"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// smallLandscape is a landscape file made by hand for the crop rotation tests: two flower patches with large areas,
// so that their layers are not removed as too small [LAYR-06], [LAYR-07].
var smallLandscape = input.Landscape{MetresPerPatch: 10, Patches: []input.Patch{
	{Id: 0, PatchType: "Scrub", X: 50, Y: 50, Area_sqm: 100000, FlowerSpeciesList: []input.FlowerEntry{
		{Name: "Bramble", Value: 0.211, HasValue: true}, {Name: "Dandelion", Value: 0.113, HasValue: true}}},
	{Id: 1, PatchType: "Garden", X: 120, Y: 80, Area_sqm: 50000, FlowerSpeciesList: []input.FlowerEntry{
		{Name: "Dandelion", Value: 0.5, HasValue: true}}},
}}

// Landscape files of the crop rotation tests: the example farm (exampleLandscape, the landscape of the run file
// built into package input) and smallLandscape (smallLandscapeFile).
const (
	exampleLandscape   = "_SYSTEM_Example_Farm_Foodsources.yaml"
	smallLandscapeFile = "small_Foodsources.yaml"
)

// cropWorld runs the example farm setup with the crop rotation list and smallLandscape among the landscape files.
func cropWorld(t *testing.T, list []string) (*app.App, *cropRotation) {
	t.Helper()
	cr := &cropRotation{}
	a := newExampleFarmSetupFiles(t, 1, map[string]input.Landscape{smallLandscapeFile: smallLandscape},
		func(c *input.Config) { c.CropRotation = list },
		&tick{}, &calendarUpdate{}, &initialQueensSetup{}, cr)
	return a, cr
}

// foodsourceSummary returns the number of foodsources and the masterpatches by patch type.
func foodsourceSummary(w *ecs.World) (int, map[string]int) {
	n := 0
	masters := map[string]int{}
	q := ecs.NewFilter2[comp.Foodsource, comp.Layer](w).Query()
	for q.Next() {
		fs, layer := q.Get()
		n++
		if layer.Masterpatch {
			masters[fs.PatchType]++
		}
	}
	return n, masters
}

// setDay sets ticks and Day.
func setDay(w *ecs.World, ticks day.Abs) {
	ecs.GetResource[globals.Ticks](w).Ticks = ticks
	ecs.GetResource[globals.Calendar](w).Day = dayOfYear(ticks)
}

func TestCropRotation(t *testing.T) {
	a, cr := cropWorld(t, []string{exampleLandscape, smallLandscapeFile})
	w := a.World
	landscape := ecs.GetResource[globals.Landscape](w)
	nBefore, _ := foodsourceSummary(w)

	// a bee that knows a foodsource, and a colony with range lists
	var anyFs ecs.Entity
	fsq := ecs.NewFilter1[comp.Foodsource](w).Query()
	for fsq.Next() {
		anyFs = fsq.Entity()
	}
	var bee ecs.Entity
	bq := ecs.NewFilter1[comp.Forager](w).Query()
	for bq.Next() {
		bee = bq.Entity()
	}
	f := ecs.NewMap1[comp.Forager](w).Get(bee)
	f.KnownMasterpatchesNectarList = []ecs.Entity{anyFs}
	f.KnownMasterpatchesPollenList = []ecs.Entity{anyFs}
	f.NectarsourceToGoTo, f.PollensourceToGoTo, f.CurrentFoodsource = anyFs, anyFs, anyFs
	inRange := []ecs.Entity{anyFs}
	col := ecs.NewMap6[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w).NewEntity(&comp.Position{}, &comp.Colony{}, &comp.ColonyStores{}, &comp.ColonyStimuli{}, &comp.ColonyCounts{},
		&comp.ColonyRange{AllPatchesInRangeList: inRange,
			MasterpatchesWithPollenlayersInFlowerAndRangeList: inRange, MasterpatchesWithNectarlayersInFlowerAndRangeList: inRange})

	// [MORN-09] only on Day 365
	setDay(w, 364)
	cr.Update(w)
	if landscape.FoodsourcesFile != exampleLandscape {
		t.Fatalf("crop rotation on day 364")
	}

	// [CROP-01] ticks 365: year 1, item 1
	setDay(w, 365)
	cr.Update(w)
	if landscape.FoodsourcesFile != smallLandscapeFile {
		t.Errorf("FoodsourcesFile %q, want %s", landscape.FoodsourcesFile, smallLandscapeFile)
	}
	// [CROP-03] only the foodsources of the new file
	n, masters := foodsourceSummary(w)
	if !maps.Equal(masters, map[string]int{"Scrub": 1, "Garden": 1}) || n != 3 {
		t.Errorf("%d foodsources, masterpatches %v; want 3 layers, one Scrub and one Garden masterpatch", n, masters)
	}
	// step F12: the landscape holds the new masterpatches only
	newMasters := ecs.GetResource[globals.Landscape](w).Masterpatches
	if len(newMasters) != 2 || !w.Alive(newMasters[0]) || !w.Alive(newMasters[1]) {
		t.Errorf("landscape masterpatches %v, want the 2 new ones", newMasters)
	}
	foodsources := ecs.NewFilter1[comp.Foodsource](w).Query()
	if got := foodsources.Count(); got != 3 {
		t.Errorf("%d foodsources, want 3", got)
	}
	foodsources.Close()

	// [CROP-02] the bee forgets her foodsources
	f = ecs.NewMap1[comp.Forager](w).Get(bee)
	if len(f.KnownMasterpatchesNectarList)+len(f.KnownMasterpatchesPollenList) != 0 ||
		!f.NectarsourceToGoTo.IsZero() || !f.PollensourceToGoTo.IsZero() || !f.CurrentFoodsource.IsZero() {
		t.Errorf("bee still knows foodsources: %+v", *f)
	}
	// the colony empties its in-flower lists, but keeps allPatchesInRangeList
	r := ecs.NewMap1[comp.ColonyRange](w).Get(col)
	if len(r.MasterpatchesWithPollenlayersInFlowerAndRangeList)+len(r.MasterpatchesWithNectarlayersInFlowerAndRangeList) != 0 {
		t.Errorf("in-flower lists not empty: %+v", *r)
	}
	if !slices.Equal(r.AllPatchesInRangeList, inRange) || w.Alive(anyFs) {
		t.Errorf("allPatchesInRangeList must keep the dead foodsource")
	}
	// [FRNG] reads the dead foodsources: NetLogo runtime error, gobees panics
	founder := newColonyFounder(w, rand.New(ecs.GetResource[resource.Rand](w)))
	expectPanic(t, "FoodsourcesInFlowerAndRangeProc after a crop rotation", func() { founder.flowersInRange(col) })

	// [SPEC-06] for B_terrestris: Scrub and Garden are nest habitats. [SPEC-07] is not computed again (step F11).
	for i, ns := range ecs.GetResource[globals.NestSites](w).BySpecies {
		if len(ns.NestsiteFoodsourceList) != 2 || ns.NestSiteArea != 150000 {
			t.Errorf("species %d: %d nest sites, area %v; want 2 and 150000", i, len(ns.NestsiteFoodsourceList), ns.NestSiteArea)
		}
	}

	// ticks 730: year 2, item 0, the example farm again
	setDay(w, 730)
	cr.Update(w)
	if n, _ := foodsourceSummary(w); landscape.FoodsourcesFile != exampleLandscape || n != nBefore {
		t.Errorf("year 2: FoodsourcesFile %q with %d foodsources, want the example farm with %d", landscape.FoodsourcesFile, n, nBefore)
	}
}

func TestCropRotationDefaultList(t *testing.T) {
	// without crop rotation (NetLogo: "" gives [""] with length 1, finding 3, see trans/netlogo-findings.md#finding-3)
	a, cr := cropWorld(t, nil)
	w := a.World
	nBefore, _ := foodsourceSummary(w)
	mastersBefore := slices.Clone(ecs.GetResource[globals.Landscape](w).Masterpatches)
	setDay(w, 365)
	cr.Update(w)
	// the same foodsources, not new ones (before step F12: the same largest who number)
	if n, _ := foodsourceSummary(w); n != nBefore || !slices.Equal(ecs.GetResource[globals.Landscape](w).Masterpatches, mastersBefore) {
		t.Errorf("foodsources changed without crop rotation")
	}
}
