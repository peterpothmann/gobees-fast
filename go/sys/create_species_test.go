package sys

import (
	"math/rand/v2"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

func TestNestHabitatsListREP(t *testing.T) {
	masterpatches := []masterpatch{
		{patchType: "Grassland"}, {patchType: "Grass_improved"}, {patchType: "Improved_Grassland"},
		{patchType: "Hedgerow"}, {patchType: "Grassland"},
	}
	got, err := NestHabitatsListREP([]string{"Grassland*", "Garden", "Hedgerow"}, masterpatches, rand.New(rand.NewPCG(1, 1)))
	if err != nil {
		t.Fatal(err)
	}
	// [NHAB-02] "Grassland*" matches every habitat type that contains "Grassland" (finding 20, see trans/netlogo-findings.md#finding-20), each habitat once.
	// Entries without "*" are copied, also if they do not exist on the map ("Garden").
	slices.Sort(got)
	want := []string{"Garden", "Grassland", "Hedgerow", "Improved_Grassland"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q (in any order)", got, want)
	}
	// "*" not at the end gives an assertion
	if _, err := NestHabitatsListREP([]string{"Gr*ss"}, masterpatches, rand.New(rand.NewPCG(1, 1))); err == nil {
		t.Error("expected an error for a '*' that is not at the end")
	}
}

func TestNestSites(t *testing.T) {
	w := ecs.NewWorld()
	e1, e2, e3 := w.NewEntity(), w.NewEntity(), w.NewEntity()
	masterpatches := []masterpatch{
		{e1, 1, "Scrub", 100}, {e2, 2, "Grassland", 200}, {e3, 3, "Improved_Grassland", 400},
	}
	// [SPEC-06] exact match only: "Grassland" does not select "Improved_Grassland"
	sites, area := nestSites([]string{"Grassland", "Scrub"}, masterpatches)
	if !slices.Equal(sites, []ecs.Entity{e1, e2}) || area != 300 {
		t.Errorf("sites %v, area %v; want [e1 e2], 300", sites, area)
	}
}

func TestCreateSpeciesExampleFarm(t *testing.T) {
	a := newExampleFarmSetup(t, 1)
	w := a.World

	query := ecs.NewFilter2[comp.Who, comp.Species](w).Query()
	species := []comp.Species{}
	whos := []int64{}
	for query.Next() {
		who, sp := query.Get()
		species = append(species, *sp)
		whos = append(whos, who.Who)
	}
	// The example farm parameter file simulates "B_terrestris 500" only.
	if len(species) != 1 || species[0].Name != "B_terrestris" {
		t.Fatalf("species = %v, want only B_terrestris", species)
	}
	sp := species[0]

	// Species are created after all foodsources, so their who number is larger than every foodsource's.
	maxFoodsourceWho := int64(-1)
	fq := ecs.NewFilter2[comp.Who, comp.Foodsource](w).Query()
	for fq.Next() {
		who, _ := fq.Get()
		maxFoodsourceWho = max(maxFoodsourceWho, who.Who)
	}
	if whos[0] <= maxFoodsourceWho {
		t.Errorf("species who %d, largest foodsource who %d", whos[0], maxFoodsourceWho)
	}

	// [SPEC-04] the example farm has the habitat types Crop_Cereals, Grass_improved, Grassland, Hedgerow, Scrub,
	// Woodland. "Grassland*" matches only "Grassland" ("Grass_improved" does not contain "Grassland").
	gotHabitats := slices.Clone(sp.NestHabitatsList)
	slices.Sort(gotHabitats)
	wantHabitats := []string{"Garden", "Grassland", "Hedgerow", "Nestboxes", "Scrub", "Woodland"}
	if !slices.Equal(gotHabitats, wantHabitats) {
		t.Errorf("nestHabitatsList = %q, want %q", gotHabitats, wantHabitats)
	}

	// [SPEC-06] nest sites: all masterpatches with one of these habitat types, sorted by who, and their area
	wantSites := []ecs.Entity{}
	wantArea := 0.0
	for _, m := range collectMasterpatches(w) {
		if slices.Contains(wantHabitats, m.patchType) {
			wantSites = append(wantSites, m.entity)
			wantArea += m.area_sqm
		}
	}
	if len(wantSites) == 0 || !slices.Equal(sp.NestsiteFoodsourceList, wantSites) || sp.NestSiteArea != wantArea {
		t.Errorf("nest sites: %d with %v m2, want %d with %v m2", len(sp.NestsiteFoodsourceList), sp.NestSiteArea, len(wantSites), wantArea)
	}

	// the who counter has been advanced for the species
	who := ecs.GetResource[globals.WhoCounter](w)
	if next := who.Next(); next != whos[0]+1 {
		t.Errorf("next who %d, want %d", next, whos[0]+1)
	}
}
