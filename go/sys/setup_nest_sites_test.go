package sys

import (
	"math/rand/v2"
	"slices"
	"testing"

	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestExpandNestHabitats(t *testing.T) {
	masterpatches := []masterpatch{
		{patchType: "Grassland"}, {patchType: "Grass_improved"}, {patchType: "Improved_Grassland"},
		{patchType: "Hedgerow"}, {patchType: "Grassland"},
	}
	got, err := expandNestHabitats([]string{"Grassland*", "Garden", "Hedgerow"}, masterpatches, rand.New(rand.NewPCG(1, 1)))
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
	if _, err := expandNestHabitats([]string{"Gr*ss"}, masterpatches, rand.New(rand.NewPCG(1, 1))); err == nil {
		t.Error("expected an error for a '*' that is not at the end")
	}
}

func TestNestSites(t *testing.T) {
	w := ecs.NewWorld()
	e1, e2, e3 := w.NewEntity(), w.NewEntity(), w.NewEntity()
	masterpatches := []masterpatch{
		{e1, "Scrub", 100}, {e2, "Grassland", 200}, {e3, "Improved_Grassland", 400},
	}
	// [SPEC-06] exact match only: "Grassland" does not select "Improved_Grassland"
	sites, area := nestSites([]string{"Grassland", "Scrub"}, masterpatches)
	if !slices.Equal(sites, []ecs.Entity{e1, e2}) || area != 300 {
		t.Errorf("sites %v, area %v; want [e1 e2], 300", sites, area)
	}
}

func TestNestSitesSetupExampleFarm(t *testing.T) {
	a := newExampleFarmSetup(t, 1)
	w := a.World

	// The example farm parameter file simulates "B_terrestris 500" only.
	species := ecs.GetResource[params.Input](w).Species
	if len(species) != 1 || species[0].Name != "B_terrestris" {
		t.Fatalf("species = %v, want only B_terrestris", species)
	}
	nestSites := ecs.GetResource[globals.NestSites](w).BySpecies
	if len(nestSites) != 1 {
		t.Fatalf("nest sites of %d species, want 1", len(nestSites))
	}
	ns := nestSites[0]

	// [SPEC-04] the example farm has the habitat types Crop_Cereals, Grass_improved, Grassland, Hedgerow, Scrub,
	// Woodland. "Grassland*" matches only "Grassland" ("Grass_improved" does not contain "Grassland").
	gotHabitats := slices.Clone(ns.Habitats)
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
			wantArea += m.area
		}
	}
	if len(wantSites) == 0 || !slices.Equal(ns.Sites, wantSites) || ns.Area != wantArea {
		t.Errorf("nest sites: %d with %v m2, want %d with %v m2", len(ns.Sites), ns.Area, len(wantSites), wantArea)
	}
	// the species keep the list of the run file
	if slices.Equal(species[0].NestHabitats, ns.Habitats) {
		t.Errorf("params.Input.Species holds the expanded list %q", species[0].NestHabitats)
	}
}
