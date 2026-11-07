package sys

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/globals"
	"gobees/input"
	"gobees/params"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

func TestMergeHedges(t *testing.T) {
	field := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 0, Y: 0, Area: 10000}}
	small := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 1, Y: 0, Area: 1000}} // below 6500 m2, never linked
	big := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 100, Y: 0, Area: 500}}
	h1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 2, Y: 0, Area: 300}}
	h2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 3, Y: 0, Area: 200}}
	mergeHedges([]*flowerPatch{field, small, h1, h2, big}, rand.New(rand.NewPCG(1, 2)))

	// the largest hedge is the master hedge, it keeps its position and gets the area of the others.
	if big.Area != 1000 || h1.Area != 0 || h2.Area != 0 {
		t.Errorf("areas: big %v, h1 %v, h2 %v; want 1000, 0, 0", big.Area, h1.Area, h2.Area)
	}
	// new radius
	if math.Abs(big.Radius-math.Sqrt(1000/math.Pi)) > 1e-12 || h1.Radius != 0 {
		t.Errorf("radius: big %v, h1 %v", big.Radius, h1.Radius)
	}
	if field.Area != 10000 || small.Area != 1000 {
		t.Error("non-hedge patches must not change")
	}
}

func TestMergeHedgesEqualAreaIsRandom(t *testing.T) {
	// ties are broken randomly, as by NetLogo's sort-on, so either of two equally large hedges can be the master.
	masters := map[string]int{}
	for seed := range uint64(50) {
		field := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: 0, Y: 0, Area: 7000}}
		a := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 1, Y: 0, Area: 100}}
		b := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 2, Y: 0, Area: 100}}
		mergeHedges([]*flowerPatch{field, a, b}, rand.New(rand.NewPCG(seed, 7)))
		switch {
		case a.Area == 200 && b.Area == 0:
			masters["a"]++
		case b.Area == 200 && a.Area == 0:
			masters["b"]++
		default:
			t.Fatalf("unexpected areas %v, %v", a.Area, b.Area)
		}
	}
	if masters["a"] == 0 || masters["b"] == 0 {
		t.Errorf("both hedges should become master in some runs, got %v", masters)
	}
}

func TestMergeHedgesEquallyNearFieldIsRandom(t *testing.T) {
	// a random one among equally near fields, as by NetLogo's min-one-of.
	linked := map[string]int{}
	for seed := range uint64(50) {
		left := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: -5, Y: 0, Area: 7000}}
		right := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: 5, Y: 0, Area: 7000}}
		h1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 0, Y: 0, Area: 300}}
		h2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: -4, Y: 0, Area: 100}} // always linked to left
		mergeHedges([]*flowerPatch{left, right, h1, h2}, rand.New(rand.NewPCG(seed, 3)))
		if h2.Area == 0 {
			linked["left"]++ // h1 and h2 share the field, h1 (larger) got the area
		} else {
			linked["right"]++
		}
	}
	if linked["left"] == 0 || linked["right"] == 0 {
		t.Errorf("the hedge should be linked to both fields in some runs, got %v", linked)
	}
}

// testFlower returns the flower species of a layer of TestCreateLayers, from its values in the test table:
// the foodsource does not store its flower species.
func testFlower(fs comp.Foodsource) string {
	switch {
	case fs.PatchType == "Scrub":
		return "Scrub" // not split, NetLogo: flowerSpecies_relativeAbundanceList (list patchType 1)
	case fs.StartDay == 100:
		return "A"
	case fs.NectarPerFlower == 0:
		return "C"
	}
	return "B"
}

func TestCreateLayers(t *testing.T) {
	w := ecs.NewWorld()
	flowers := &params.FlowerSpeciesTable{
		Names: []string{"A", "B", "C"},
		Rows: []params.FlowerSpecies{
			{Name: "A", PollenPerFlower: 0.01, NectarPerFlower: 0.001, StartDay: 100, StopDay: 200},
			{Name: "B", PollenPerFlower: 0.0001, NectarPerFlower: 0.001},
			{Name: "C", PollenPerFlower: 0.01, NectarPerFlower: 0},
		},
	}
	file := &params.Parameters{AbundanceBoost: 1, FoodsourceLimit: 1, LimitFoodsources: true, RemoveEmptyFoodsources: true}
	// limits: 180 microlitres nectar, 0.05 g pollen
	p1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 10.2, Y: 20.4,
		Area: 100, Flowers: []params.FlowerEntry{{Name: "A", Value: 2, HasValue: true}, {Name: "Margin"},
			{Name: "B", Value: 1, HasValue: true}, {Name: "C", Value: 1, HasValue: true}}}}
	p2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Crop_Cereals", X: 10.4, Y: 19.6,
		Area: 100, Flowers: []params.FlowerEntry{{Name: "B", Value: 1, HasValue: true}}}}
	p3 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Scrub", X: 50, Y: 50, Area: 400}}

	masterpatches, err := createLayers(w, []*flowerPatch{p1, p2, p3}, flowers, file, []string{"Margin", "Plot"}, rand.New(rand.NewPCG(1, 1)))
	if err != nil {
		t.Fatal(err)
	}

	byKey := map[string]testLayer{}
	query := ecs.NewFilter2[comp.Foodsource, comp.Layer](w).Query()
	for query.Next() {
		fs, l := query.Get()
		byKey[fs.PatchType+"/"+testFlower(*fs)] = testLayer{query.Entity(), *fs, *l}
	}
	if len(byKey) != 4 {
		t.Fatalf("got %d foodsources, want 4: %v", len(byKey), byKey)
	}
	a, c, crop, scrub := byKey["Grassland/A"], byKey["Grassland/C"], byKey["Crop_Cereals/B"], byKey["Scrub/Scrub"]

	// A: pollen 100 * 2 * 0.01 = 2 g, nectar 100 * 2 * 1 microlitre = 200 microlitres, flowering period from the table
	if a.fs.PollenMax != 2 || a.fs.NectarMax != 200 || a.fs.NectarPerFlower != 1 || a.fs.StartDay != 100 || a.fs.StopDay != 200 {
		t.Errorf("layer A = %+v", a.fs)
	}
	// the stewardship option "Margin" gets no layer: 4 foodsources (checked above)
	// C: nectar 0 below the limit, pollen 1 g kept
	if c.fs.NectarMax != 0 || c.fs.PollenMax != 1 {
		t.Errorf("layer C nectar %v, pollen %v; want 0, 1", c.fs.NectarMax, c.fs.PollenMax)
	}
	// B on the grassland is removed (both below the limits), but kept on the crop field ("Crop_")
	if _, ok := byKey["Grassland/B"]; ok {
		t.Error("layer B on Grassland should be removed")
	}
	if crop.fs.NectarMax != 0 || crop.fs.PollenMax != 0 {
		t.Errorf("crop layer B = %+v, want no food", crop.fs)
	}
	// masterpatches
	if !a.l.IsMasterpatch || a.l.Masterpatch != a.e || c.l.IsMasterpatch || c.l.Masterpatch != a.e {
		t.Errorf("A master %v -> %v, C master %v -> %v", a.l.IsMasterpatch, a.l.Masterpatch, c.l.IsMasterpatch, c.l.Masterpatch)
	}
	if !crop.l.IsMasterpatch || crop.l.Masterpatch != crop.e {
		t.Error("the crop layer must be its own masterpatch")
	}
	// a flower patch with an empty list stays, without food
	if !scrub.l.IsMasterpatch || scrub.fs.NectarMax != 0 || scrub.fs.PollenMax != 0 ||
		scrub.fs.StartDay != 0 || scrub.fs.StopDay != 0 {
		t.Errorf("scrub = %+v %+v", scrub.fs, scrub.l)
	}
	// A, C and the crop layer lie in the same patch cell (10, 20). A and C share the list of the grassland, in the
	// order of creation (A before C). The crop layer has its own list (NetLogo: one list of all three).
	for _, l := range []testLayer{a, c} {
		if want := []ecs.Entity{a.e, c.e}; !slices.Equal(l.l.PatchLayers, want) {
			t.Errorf("%s/%s: layersInPatchList %v, want %v", l.fs.PatchType, testFlower(l.fs), l.l.PatchLayers, want)
		}
	}
	if !slices.Equal(crop.l.PatchLayers, []ecs.Entity{crop.e}) {
		t.Errorf("crop: layersInPatchList %v, want %v", crop.l.PatchLayers, []ecs.Entity{crop.e})
	}
	if !slices.Equal(scrub.l.PatchLayers, []ecs.Entity{scrub.e}) {
		t.Errorf("scrub: layersInPatchList %v", scrub.l.PatchLayers)
	}
	// the masterpatches in NetLogo's who order. The scrub is not split and gets its who number before all layers,
	// so it comes first. The grassland and the crop field follow in shuffled order.
	if len(masterpatches) != 3 || masterpatches[0] != scrub.e ||
		!slices.Contains(masterpatches[1:], a.e) || !slices.Contains(masterpatches[1:], crop.e) {
		t.Errorf("masterpatches %v, want scrub %v first, then A %v and crop %v", masterpatches, scrub.e, a.e, crop.e)
	}
}

// testLayer is a foodsource as read back in TestCreateLayers.
type testLayer struct {
	e  ecs.Entity
	fs comp.Foodsource
	l  comp.Layer
}

// newExampleFarmSetup runs landscapeSetup and nestSitesSetup on the example farm (the run file built into
// package input) with the given seed, followed by the extra parts.
func newExampleFarmSetup(t *testing.T, seed float64, extra ...app.System) *app.App {
	t.Helper()
	return newExampleFarmSetupWith(t, seed, func(c *input.Config) {}, extra...)
}

// newExampleFarmSetupWithModifier runs the setup incl. initialQueensSetup with the given
// InitialQueensFactor.
func newExampleFarmSetupWithModifier(t *testing.T, modifier float64) *app.App {
	t.Helper()
	return newExampleFarmSetupWith(t, 1,
		func(c *input.Config) { c.Parameters.InitialQueensFactor = modifier }, &initialQueensSetup{})
}

// newExampleFarmSetupWith is newExampleFarmSetup with a function that changes the run file before the setup.
func newExampleFarmSetupWith(t *testing.T, seed float64, change func(c *input.Config), extra ...app.System) *app.App {
	t.Helper()
	return newExampleFarmSetupFiles(t, seed, nil, change, extra...)
}

// newExampleFarmSetupFiles is newExampleFarmSetupWith with further landscape files (name: content), e.g. maps of a
// crop rotation.
func newExampleFarmSetupFiles(t *testing.T, seed float64, landscapes map[string]input.Landscape,
	change func(c *input.Config), extra ...app.System) *app.App {
	t.Helper()
	p := exampleFarmParams(t, landscapes, change)
	p.Seed = int64(seed)

	a := app.New()
	p.Apply(a.World)
	ecs.AddResource(a.World, &globals.ActiveBees{})
	ecs.AddResource(a.World, &globals.Daytime{})
	a.AddSystem(&landscapeSetup{})
	a.AddSystem(&nestSitesSetup{})
	for _, s := range extra {
		a.AddSystem(s)
	}
	a.Initialize()
	return a
}

// exampleFarmParams returns the parameters of the example farm (the run file built into package input), after
// change has changed the run file, with the further landscape files.
func exampleFarmParams(t *testing.T, landscapes map[string]input.Landscape, change func(c *input.Config)) params.Params {
	t.Helper()
	c, all := input.ExampleFarmConfig()
	for name, l := range landscapes {
		all[name] = l
	}
	change(&c)
	p, err := input.Build(c, all)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCreateFoodsourcesExampleFarm(t *testing.T) {
	counts := map[int]int{}
	for seed := 1; seed <= 20; seed++ {
		a := newExampleFarmSetup(t, float64(seed))
		w := a.World
		landscape := ecs.GetResource[globals.Landscape](w)
		if landscape.Scale != 0.1 {
			t.Fatalf("Scaling_NLpatches/m = %v, want 0.1 (1 / 10 m)", landscape.Scale)
		}

		masterpatches := collectMasterpatches(w)
		counts[len(masterpatches)]++

		// every layer points to a masterpatch of the same flower patch, i.e. same position and patch type
		posMap := ecs.NewMap2[comp.Position, comp.Foodsource](w)
		query := ecs.NewFilter3[comp.Position, comp.Foodsource, comp.Layer](w).Query()
		for query.Next() {
			pos, fs, l := query.Get()
			mpos, mfs := posMap.Get(l.Masterpatch)
			if *mpos != *pos || mfs.PatchType != fs.PatchType || mfs.Area != fs.Area {
				t.Errorf("layer %s at %v points to masterpatch %s at %v", fs.PatchType, *pos, mfs.PatchType, *mpos)
			}
		}
	}
	// NetLogo gives 36 masterpatches after the setup. 36 with every seed: the random
	// hedge merge order does not change the number for the example farm.
	if counts[36] != 20 {
		t.Errorf("masterpatches over 20 seeds: %v, want always 36 as in NetLogo", counts)
	}
}
