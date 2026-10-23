package sys

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobees/globals"
	"gobees/internal/examplefiles"
	"gobees/params"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// wantCSOList are the stewardship option names in _SYSTEM_Habitats.csv, in file order [CSO].
var wantCSOList = []string{"Legume", "Margin", "newCSO_Legume", "BirdBumble_Margin", "GameCover_Plot", "Plot"}

func TestParametersExampleFarm(t *testing.T) {
	// The system reads the HabitatsFile from the working directory, so the copy is written to a temporary one.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "_SYSTEM_Habitats.csv"), []byte(examplefiles.Habitats), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := params.ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	p := params.Default()
	p.WorkingDirectory.Path = dir
	p.File = file

	a := app.New()
	p.Apply(a.World)
	a.AddSystem(&Parameters{})
	a.Initialize()
	d := ecs.GetResource[globals.Derived](a.World)

	want := globals.Derived{
		BeeSpeciesInitialQueensList: []globals.SpeciesQueens{{Species: "B_terrestris", Queens: 500}},
		SpeciesList:                 []string{"B_terrestris"},
		MortalityForager_per_s:      1.0e-05,
		FoodsourcesFile:             "_SYSTEM_Example_Farm_Foodsources.txt",
		CSS_OptionsList:             wantCSOList,
		CropRotationList:            []string{""}, // finding 3 (see trans/netlogo-findings.md#finding-3): length 1, no crop rotation
	}
	if !reflect.DeepEqual(*d, want) {
		t.Errorf("got  %+v\nwant %+v", *d, want)
	}
}

func TestParseCSOListExample(t *testing.T) {
	// The original file has CRLF line endings, the copy LF. Both must give the same list.
	for _, lineEnd := range []string{"\n", "\r\n"} {
		got, err := parseCSOList(strings.NewReader(strings.ReplaceAll(examplefiles.Habitats, "\n", lineEnd)))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, wantCSOList) {
			t.Errorf("line end %q: got %q, want %q", lineEnd, got, wantCSOList)
		}
	}
}

func TestAllCSOListMissingFile(t *testing.T) {
	// NetLogo shows a message and continues with an empty list. gobees stops with an error.
	if _, err := AllCSOlistREP(filepath.Join(t.TempDir(), "missing.csv")); err == nil {
		t.Error("expected an error for a missing HabitatsFile")
	}
}

func TestMortalityForagerUnknownModel(t *testing.T) {
	// Finding 5 (see trans/netlogo-findings.md#finding-5): case-sensitive, any other value gives 0.
	for _, m := range []string{"High", "medium", ""} {
		if got := mortalityForager(m); got != 0 {
			t.Errorf("mortalityForager(%q) = %v, want 0", m, got)
		}
	}
}

func TestQueensInitialProc(t *testing.T) {
	cases := []struct {
		input   string
		queens  []globals.SpeciesQueens
		species []string
	}{
		{"B_hortorum 100 B_terrestris 500",
			[]globals.SpeciesQueens{{Species: "B_hortorum", Queens: 100}, {Species: "B_terrestris", Queens: 500}},
			[]string{"B_hortorum", "B_terrestris"}},
		{"A 1 B", []globals.SpeciesQueens{{Species: "A", Queens: 1}}, []string{"A"}}, // trailing name is dropped
		{"", []globals.SpeciesQueens{}, []string{}},                                  // [""] gives no species
	}
	for _, c := range cases {
		q, s, err := QueensInitialProc(c.input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(q, c.queens) || !reflect.DeepEqual(s, c.species) {
			t.Errorf("QueensInitialProc(%q) = %v, %v; want %v, %v", c.input, q, s, c.queens, c.species)
		}
	}
	if _, _, err := QueensInitialProc("B_terrestris many"); err == nil {
		t.Error("expected an error for a non-numeric number of queens")
	}
}
