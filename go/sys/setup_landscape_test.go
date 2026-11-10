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
	mergeHedges([]*flowerPatch{field, small, h1, h2, big}, 6500, rand.New(rand.NewPCG(1, 2)))

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
		mergeHedges([]*flowerPatch{field, a, b}, 6500, rand.New(rand.NewPCG(seed, 7)))
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
		mergeHedges([]*flowerPatch{left, right, h1, h2}, 6500, rand.New(rand.NewPCG(seed, 3)))
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

// testFlowers returns the flower species of the layers of a flower patch of TestCreateLayers (Flower, an index in the
// flower species A, B, C of the test).
func testFlowers(fp *comp.FlowerPatch) []string {
	names := []string{}
	for _, l := range fp.Layers {
		names = append(names, []string{"A", "B", "C"}[l.Flower])
	}
	return names
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
		Area: 100, Flowers: []params.FlowerEntry{{Name: "A", Value: 2}, {Name: "Margin", Value: 1}, {Name: "B", Value: 1},
			{Name: "C", Value: 1}}}}
	p2 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Crop_Cereals", X: 10.4, Y: 19.6,
		Area: 100, Flowers: []params.FlowerEntry{{Name: "B", Value: 1}}}}
	p3 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Scrub", X: 50, Y: 50, Area: 400}}
	// every layer of the woodland is removed: no flower patch (NetLogo: no masterpatch)
	p4 := &flowerPatch{FlowerPatch: params.FlowerPatch{PatchType: "Woodland", X: 70, Y: 70,
		Area: 100, Flowers: []params.FlowerEntry{{Name: "B", Value: 1}}}}

	patches, err := createLayers(w, []*flowerPatch{p1, p2, p3, p4}, flowers, file, &defaults.Landscape,
		[]string{"Margin", "Plot"}, rand.New(rand.NewPCG(1, 1)))
	if err != nil {
		t.Fatal(err)
	}

	byType := map[string]ecs.Entity{}
	query := ecs.NewFilter1[comp.FlowerPatch](w).Query()
	for query.Next() {
		byType[query.Get().PatchType] = query.Entity()
	}
	if len(byType) != 3 {
		t.Fatalf("got %d flower patches, want 3 (no woodland): %v", len(byType), byType)
	}
	fpMap := ecs.NewMap1[comp.FlowerPatch](w)
	grass, crop, scrub := fpMap.Get(byType["Grassland"]), fpMap.Get(byType["Crop_Cereals"]), fpMap.Get(byType["Scrub"])

	// the stewardship option "Margin" gets no layer, also with a value, B on the grassland is removed (both below the limits). A and C
	// stay in the order of the flower list (NetLogo: who order).
	if got := testFlowers(grass); !slices.Equal(got, []string{"A", "C"}) {
		t.Fatalf("grassland layers %v, want [A C]", got)
	}
	a, c := grass.Layers[0], grass.Layers[1]
	// A: pollen 100 * 2 * 0.01 = 2 g, nectar 100 * 2 * 1 microlitre = 200 microlitres, flowering period from the table
	if a.PollenMax != 2 || a.NectarMax != 200 || a.NectarPerFlower != 1 || a.StartDay != 100 || a.StopDay != 200 {
		t.Errorf("layer A = %+v", a)
	}
	// C: nectar 0 below the limit, pollen 1 g kept
	if c.NectarMax != 0 || c.PollenMax != 1 {
		t.Errorf("layer C nectar %v, pollen %v; want 0, 1", c.NectarMax, c.PollenMax)
	}
	// B is kept on the crop field ("Crop_"), without food
	if got := testFlowers(crop); !slices.Equal(got, []string{"B"}) || crop.Layers[0].NectarMax != 0 ||
		crop.Layers[0].PollenMax != 0 {
		t.Errorf("crop layers %v = %+v, want B without food", got, crop.Layers)
	}
	// a flower patch with an empty list stays, without layers
	if len(scrub.Layers) != 0 || scrub.Area != 400 {
		t.Errorf("scrub = %+v", *scrub)
	}
	// A, C and the crop layer lie in the same patch cell (10, 20). The grassland and the crop field keep their own
	// layers (NetLogo: one list of all three, finding 15).
	// the flower patches in NetLogo's who order of their masterpatches. The scrub is not split and gets its who
	// number before all layers, so it comes first. The grassland and the crop field follow in shuffled order.
	if len(patches) != 3 || patches[0] != byType["Scrub"] ||
		!slices.Contains(patches[1:], byType["Grassland"]) || !slices.Contains(patches[1:], byType["Crop_Cereals"]) {
		t.Errorf("patches %v, want scrub first, then grassland and crop field: %v", patches, byType)
	}
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

		patches := collectPatches(w)
		counts[len(patches)]++

		// landscape.Patches holds every flower patch once
		all := []ecs.Entity{}
		query := ecs.NewFilter1[comp.FlowerPatch](w).Query()
		for query.Next() {
			all = append(all, query.Entity())
		}
		sorted := slices.Clone(landscape.Patches)
		slices.SortFunc(sorted, func(a, b ecs.Entity) int { return int(a.ID()) - int(b.ID()) })
		slices.SortFunc(all, func(a, b ecs.Entity) int { return int(a.ID()) - int(b.ID()) })
		if !slices.Equal(sorted, all) {
			t.Errorf("seed %d: landscape.Patches %v, flower patches %v", seed, landscape.Patches, all)
		}
	}
	// NetLogo gives 36 masterpatches after the setup. 36 with every seed: the random
	// hedge merge order does not change the number for the example farm.
	if counts[36] != 20 {
		t.Errorf("masterpatches over 20 seeds: %v, want always 36 as in NetLogo", counts)
	}
}
