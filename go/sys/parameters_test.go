package sys

import (
	"reflect"
	"testing"

	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

func TestParametersDefaultFiles(t *testing.T) {
	p := params.Default()
	p.WorkingDirectory.Path = "../../netlogo"
	file, err := params.ReadParameterFile("../../netlogo/_SYSTEM_Example_Farm_Parameters.csv")
	if err != nil {
		t.Fatal(err)
	}
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
		CSS_OptionsList:             []string{"Legume", "Margin", "newCSO_Legume", "BirdBumble_Margin", "GameCover_Plot", "Plot"},
		CropRotationList:            []string{""}, // finding 3: length 1, no crop rotation
	}
	if !reflect.DeepEqual(*d, want) {
		t.Errorf("got  %+v\nwant %+v", *d, want)
	}
}

func TestMortalityForagerUnknownModel(t *testing.T) {
	// Finding 5: case-sensitive, any other value gives 0.
	for _, m := range []string{"High", "medium", ""} {
		if got := mortalityForager(m); got != 0 {
			t.Errorf("mortalityForager(%q) = %v, want 0", m, got)
		}
	}
}

func TestQueensInitial(t *testing.T) {
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
		q, s, err := queensInitial(c.input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(q, c.queens) || !reflect.DeepEqual(s, c.species) {
			t.Errorf("queensInitial(%q) = %v, %v; want %v, %v", c.input, q, s, c.queens, c.species)
		}
	}
	if _, _, err := queensInitial("B_terrestris many"); err == nil {
		t.Error("expected an error for a non-numeric number of queens")
	}
}
