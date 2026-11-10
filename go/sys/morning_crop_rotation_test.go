package sys

import (
	"maps"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/globals"
	"gobees/input"
	"gobees/params"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// smallLandscape is a landscape file made by hand for the crop rotation tests: two flower patches with large areas,
// so that their layers are not removed as too small.
var smallLandscape = input.Landscape{MetresPerPatch: 10, Patches: []input.Patch{
	{ID: 0, PatchType: "Scrub", X: 50, Y: 50, Area: 100000, Flowers: []params.FlowerEntry{
		{Name: "Bramble", Value: 0.211}, {Name: "Dandelion", Value: 0.113}}},
	{ID: 1, PatchType: "Garden", X: 120, Y: 80, Area: 50000, Flowers: []params.FlowerEntry{
		{Name: "Dandelion", Value: 0.5}}},
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

// foodsourceSummary returns the number of layers and the flower patches by patch type.
func foodsourceSummary(w *ecs.World) (int, map[string]int) {
	n := 0
	patches := map[string]int{}
	q := ecs.NewFilter1[comp.FlowerPatch](w).Query()
	for q.Next() {
		fp := q.Get()
		n += len(fp.Layers)
		patches[fp.PatchType]++
	}
	return n, patches
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

	// a bee that knows a flower patch, and a colony with range lists
	var anyFs ecs.Entity
	fsq := ecs.NewFilter1[comp.FlowerPatch](w).Query()
	for fsq.Next() {
		anyFs = fsq.Entity()
	}
	var bee ecs.Entity
	bq := ecs.NewFilter1[comp.Forager](w).Query()
	for bq.Next() {
		bee = bq.Entity()
	}
	f := ecs.NewMap1[comp.Forager](w).Get(bee)
	f.KnownNectarPatches = []ecs.Entity{anyFs}
	f.KnownPollenPatches = []ecs.Entity{anyFs}
	src := comp.Source{Patch: anyFs}
	f.NectarSource, f.PollenSource, f.CurrentFoodsource = src, src, src
	inRange := []ecs.Entity{anyFs}
	col := ecs.NewMap6[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w).NewEntity(&comp.Position{}, &comp.Colony{}, &comp.ColonyStores{}, &comp.ColonyStimuli{}, &comp.ColonyCounts{},
		&comp.ColonyRange{Patches: inRange,
			PollenPatches: inRange, NectarPatches: inRange})

	// only on Day 365
	setDay(w, 364)
	cr.Update(w)
	if landscape.File != exampleLandscape {
		t.Fatalf("crop rotation on day 364")
	}

	// ticks 365: year 1, item 1
	setDay(w, 365)
	cr.Update(w)
	if landscape.File != smallLandscapeFile {
		t.Errorf("File %q, want %s", landscape.File, smallLandscapeFile)
	}
	// only the flower patches of the new file
	n, masters := foodsourceSummary(w)
	if !maps.Equal(masters, map[string]int{"Scrub": 1, "Garden": 1}) || n != 3 {
		t.Errorf("%d layers, flower patches %v; want 3 layers, one Scrub and one Garden flower patch", n, masters)
	}
	// the landscape holds the new flower patches only
	newMasters := ecs.GetResource[globals.Landscape](w).Patches
	if len(newMasters) != 2 || !w.Alive(newMasters[0]) || !w.Alive(newMasters[1]) {
		t.Errorf("landscape flower patches %v, want the 2 new ones", newMasters)
	}
	patches := ecs.NewFilter1[comp.FlowerPatch](w).Query()
	if got := patches.Count(); got != 2 {
		t.Errorf("%d flower patches, want 2", got)
	}
	patches.Close()

	// the bee forgets her foodsources
	f = ecs.NewMap1[comp.Forager](w).Get(bee)
	if len(f.KnownNectarPatches)+len(f.KnownPollenPatches) != 0 ||
		!f.NectarSource.IsZero() || !f.PollenSource.IsZero() || !f.CurrentFoodsource.IsZero() {
		t.Errorf("bee still knows foodsources: %+v", *f)
	}
	// the colony empties its in-flower lists, but keeps allPatchesInRangeList
	r := ecs.NewMap1[comp.ColonyRange](w).Get(col)
	if len(r.PollenPatches)+len(r.NectarPatches) != 0 {
		t.Errorf("in-flower lists not empty: %+v", *r)
	}
	if !slices.Equal(r.Patches, inRange) || w.Alive(anyFs) {
		t.Errorf("allPatchesInRangeList must keep the dead flower patch")
	}
	// reads the dead flower patches: a panic, as NetLogo stops with a runtime error
	founder := newColonyFounder(w)
	expectPanic(t, "flowersInRange after a crop rotation", func() { founder.flowersInRange(col) })

	// nest sites for B_terrestris: Scrub and Garden are nest habitats
	for i, ns := range ecs.GetResource[globals.NestSites](w).BySpecies {
		if len(ns.Sites) != 2 || ns.Area != 150000 {
			t.Errorf("species %d: %d nest sites, area %v; want 2 and 150000", i, len(ns.Sites), ns.Area)
		}
	}

	// ticks 730: year 2, item 0, the example farm again
	setDay(w, 730)
	cr.Update(w)
	if n, _ := foodsourceSummary(w); landscape.File != exampleLandscape || n != nBefore {
		t.Errorf("year 2: File %q with %d layers, want the example farm with %d", landscape.File, n, nBefore)
	}
}

func TestCropRotationDefaultList(t *testing.T) {
	// without crop rotation (NetLogo turns "" into [""], a list of length 1)
	a, cr := cropWorld(t, nil)
	w := a.World
	nBefore, _ := foodsourceSummary(w)
	mastersBefore := slices.Clone(ecs.GetResource[globals.Landscape](w).Patches)
	setDay(w, 365)
	cr.Update(w)
	// the same foodsources, not new ones
	if n, _ := foodsourceSummary(w); n != nBefore || !slices.Equal(ecs.GetResource[globals.Landscape](w).Patches, mastersBefore) {
		t.Errorf("foodsources changed without crop rotation")
	}
}
