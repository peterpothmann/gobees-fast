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
	field := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 0, Y: 0, Area_sqm: 10000}}
	small := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 1, Y: 0, Area_sqm: 1000}} // below 6500 m2, never linked
	big := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 100, Y: 0, Area_sqm: 500}}
	h1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 2, Y: 0, Area_sqm: 300}}
	h2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 3, Y: 0, Area_sqm: 200}}
	mergeHedges([]*flowerPatch{field, small, h1, h2, big}, rand.New(rand.NewPCG(1, 2)))

	// [HEDG-04] the largest hedge is the master hedge, it keeps its position and gets the area of the others.
	if big.Area_sqm != 1000 || h1.Area_sqm != 0 || h2.Area_sqm != 0 {
		t.Errorf("areas: big %v, h1 %v, h2 %v; want 1000, 0, 0", big.Area_sqm, h1.Area_sqm, h2.Area_sqm)
	}
	// [HEDG-05] new radius
	if math.Abs(big.Radius_m-math.Sqrt(1000/math.Pi)) > 1e-12 || h1.Radius_m != 0 {
		t.Errorf("radius: big %v, h1 %v", big.Radius_m, h1.Radius_m)
	}
	if field.Area_sqm != 10000 || small.Area_sqm != 1000 {
		t.Error("non-hedge patches must not change")
	}
}

func TestMergeHedgesEqualAreaIsRandom(t *testing.T) {
	// Finding 14 (see trans/netlogo-findings.md#finding-14): sort-on breaks ties randomly, so either of two equally large hedges can be the master.
	masters := map[string]int{}
	for seed := range uint64(50) {
		field := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: 0, Y: 0, Area_sqm: 7000}}
		a := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 1, Y: 0, Area_sqm: 100}}
		b := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 2, Y: 0, Area_sqm: 100}}
		mergeHedges([]*flowerPatch{field, a, b}, rand.New(rand.NewPCG(seed, 7)))
		switch {
		case a.Area_sqm == 200 && b.Area_sqm == 0:
			masters["a"]++
		case b.Area_sqm == 200 && a.Area_sqm == 0:
			masters["b"]++
		default:
			t.Fatalf("unexpected areas %v, %v", a.Area_sqm, b.Area_sqm)
		}
	}
	if masters["a"] == 0 || masters["b"] == 0 {
		t.Errorf("both hedges should become master in some runs, got %v", masters)
	}
}

func TestMergeHedgesEquallyNearFieldIsRandom(t *testing.T) {
	// [HEDG-03] min-one-of picks a random one among equally near fields.
	linked := map[string]int{}
	for seed := range uint64(50) {
		left := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: -5, Y: 0, Area_sqm: 7000}}
		right := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: 5, Y: 0, Area_sqm: 7000}}
		h1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: 0, Y: 0, Area_sqm: 300}}
		h2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Hedgerow", X: -4, Y: 0, Area_sqm: 100}} // always linked to left
		mergeHedges([]*flowerPatch{left, right, h1, h2}, rand.New(rand.NewPCG(seed, 3)))
		if h2.Area_sqm == 0 {
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
// the foodsource does not store its flower species (removed in step F3).
func testFlower(fs comp.Foodsource) string {
	switch {
	case fs.PatchType == "Scrub":
		return "Scrub" // not split, NetLogo: flowerSpecies_relativeAbundanceList (list patchType 1)
	case fs.StartDay == 100:
		return "A"
	case fs.NectarFlowerVolume_myl == 0:
		return "C"
	}
	return "B"
}

func TestCreateLayers(t *testing.T) {
	w := ecs.NewWorld()
	flowers := &params.FlowerSpeciesTable{
		Names: []string{"A", "B", "C"},
		Rows: []params.FlowerSpecies{
			{Name: "A", Pollen_g_per_flower: 0.01, Nectar_ml_per_flower: 0.001, StartDay: 100, StopDay: 200},
			{Name: "B", Pollen_g_per_flower: 0.0001, Nectar_ml_per_flower: 0.001},
			{Name: "C", Pollen_g_per_flower: 0.01, Nectar_ml_per_flower: 0},
		},
	}
	file := &params.ParameterFile{AbundanceBoost: 1, FoodSourceLimit: 1, MinSizeFoodSources: true, RemoveEmptyFoodSources: true}
	// limits [LAYR-01]: 180 microlitres nectar, 0.05 g pollen
	p1 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Grassland", X: 10.2, Y: 20.4,
		Area_sqm: 100, FlowerSpeciesList: []params.FlowerEntry{{Name: "A", Value: 2, HasValue: true}, {Name: "Margin"},
			{Name: "B", Value: 1, HasValue: true}, {Name: "C", Value: 1, HasValue: true}}}}
	p2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Crop_Cereals", X: 10.4, Y: 19.6,
		Area_sqm: 100, FlowerSpeciesList: []params.FlowerEntry{{Name: "B", Value: 1, HasValue: true}}}}
	p3 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Scrub", X: 50, Y: 50, Area_sqm: 400}}

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

	// [LAYR-05] A: pollen 100 * 2 * 0.01 = 2 g, nectar 100 * 2 * 1 microlitre = 200 microlitres, flowering period from the table
	if a.fs.PollenMax_g != 2 || a.fs.NectarMax_myl != 200 || a.fs.NectarFlowerVolume_myl != 1 || a.fs.StartDay != 100 || a.fs.StopDay != 200 {
		t.Errorf("layer A = %+v", a.fs)
	}
	// [LAYR-04] the stewardship option "Margin" gets no layer: 4 foodsources (checked above)
	// [LAYR-06] C: nectar 0 below the limit, pollen 1 g kept
	if c.fs.NectarMax_myl != 0 || c.fs.PollenMax_g != 1 {
		t.Errorf("layer C nectar %v, pollen %v; want 0, 1", c.fs.NectarMax_myl, c.fs.PollenMax_g)
	}
	// [LAYR-07] B on the grassland is removed (both below the limits), but kept on the crop field ("Crop_")
	if _, ok := byKey["Grassland/B"]; ok {
		t.Error("layer B on Grassland should be removed")
	}
	if crop.fs.NectarMax_myl != 0 || crop.fs.PollenMax_g != 0 {
		t.Errorf("crop layer B = %+v, want no food", crop.fs)
	}
	// [LAYR-08] masterpatches
	if !a.l.Masterpatch || a.l.MasterpatchID != a.e || c.l.Masterpatch || c.l.MasterpatchID != a.e {
		t.Errorf("A master %v -> %v, C master %v -> %v", a.l.Masterpatch, a.l.MasterpatchID, c.l.Masterpatch, c.l.MasterpatchID)
	}
	if !crop.l.Masterpatch || crop.l.MasterpatchID != crop.e {
		t.Error("the crop layer must be its own masterpatch")
	}
	// [LAYR-09] finding 19 (see trans/netlogo-findings.md#finding-19): a flower patch with an empty list stays, without food
	if !scrub.l.Masterpatch || scrub.fs.NectarMax_myl != 0 || scrub.fs.PollenMax_g != 0 ||
		scrub.fs.StartDay != 0 || scrub.fs.StopDay != 0 {
		t.Errorf("scrub = %+v %+v", scrub.fs, scrub.l)
	}
	// [LAYR-10] finding 15 (see trans/netlogo-findings.md#finding-15), fixed in step F5: A, C and the crop layer lie
	// in the same patch cell (10, 20). A and C share the list of the grassland, sorted by who (A before C: order of
	// creation, step F12). The crop layer has its own list (NetLogo: one list of all three).
	for _, l := range []testLayer{a, c} {
		if want := []ecs.Entity{a.e, c.e}; !slices.Equal(l.l.LayersInPatchList, want) {
			t.Errorf("%s/%s: layersInPatchList %v, want %v", l.fs.PatchType, testFlower(l.fs), l.l.LayersInPatchList, want)
		}
	}
	if !slices.Equal(crop.l.LayersInPatchList, []ecs.Entity{crop.e}) {
		t.Errorf("crop: layersInPatchList %v, want %v", crop.l.LayersInPatchList, []ecs.Entity{crop.e})
	}
	if !slices.Equal(scrub.l.LayersInPatchList, []ecs.Entity{scrub.e}) {
		t.Errorf("scrub: layersInPatchList %v", scrub.l.LayersInPatchList)
	}
	// step F12: the masterpatches in who order. The scrub is not split and took its who number in [FSRC-04], before
	// all layers, so it comes first. The grassland and the crop field follow in the shuffled order of [LAYR-03].
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

// newExampleFarmSetup runs the initialisation systems of steps 1 and 2 on the example farm (the run file built into
// package input) with the given seed, followed by the extra parts.
func newExampleFarmSetup(t *testing.T, seed float64, extra ...app.System) *app.App {
	t.Helper()
	return newExampleFarmSetupWith(t, seed, func(c *input.Config) {}, extra...)
}

// newExampleFarmSetupWithModifier runs the setup incl. initialQueensSetup with the given
// InitialQueensModifierBehaviorSpace.
func newExampleFarmSetupWithModifier(t *testing.T, modifier float64) *app.App {
	t.Helper()
	return newExampleFarmSetupWith(t, 1,
		func(c *input.Config) { c.Parameters.InitialQueensModifierBehaviorSpace = modifier }, &initialQueensSetup{})
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
		if landscape.Scaling_NLpatches_per_m != 0.1 {
			t.Fatalf("Scaling_NLpatches/m = %v, want 0.1 (1 / 10 m)", landscape.Scaling_NLpatches_per_m)
		}

		masterpatches := collectMasterpatches(w)
		counts[len(masterpatches)]++

		// every layer points to a masterpatch of the same flower patch, i.e. same position and patch type
		posMap := ecs.NewMap2[comp.Position, comp.Foodsource](w)
		query := ecs.NewFilter3[comp.Position, comp.Foodsource, comp.Layer](w).Query()
		for query.Next() {
			pos, fs, l := query.Get()
			mpos, mfs := posMap.Get(l.MasterpatchID)
			if *mpos != *pos || mfs.PatchType != fs.PatchType || mfs.Area_sqm != fs.Area_sqm {
				t.Errorf("layer %s at %v points to masterpatch %s at %v", fs.PatchType, *pos, mfs.PatchType, *mpos)
			}
		}
	}
	// Finding 21 (see trans/netlogo-findings.md#finding-21): NetLogo gave 36 masterpatches after Setup. gobees gives 36 with every seed: the random
	// hedge merge order does not change the number for the example farm.
	if counts[36] != 20 {
		t.Errorf("masterpatches over 20 seeds: %v, want always 36 as in NetLogo", counts)
	}
}
