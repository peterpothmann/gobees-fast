package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/globals"
	"gobees/input"
	"gobees/internal/examplefiles"
	"gobees/model"
	"gobees/params"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// inputDir writes the input files of the example farm to a temporary directory, as the folder of the .nlogo.
func inputDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		exampleFarmParameters:                  examplefiles.ExampleFarmParameters,
		"_SYSTEM_Habitats.csv":                 examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt": examplefiles.ExampleFarmFoodsources,
		"_SYSTEM_Flowerspecies.csv":            examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":      examplefiles.BumbleSpecies,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// parse parses the command-line flags args.
func parse(t *testing.T, args ...string) cli {
	t.Helper()
	var c cli
	parser, err := newParser(&c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse(args); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCLIDefaults(t *testing.T) {
	c := parse(t)
	if c.Parameters != "" {
		t.Errorf("Parameters = %q, want empty (example farm)", c.Parameters)
	}
	if filepath.Base(c.Dir) != "netlogo" {
		t.Errorf("Dir = %q, want the folder netlogo", c.Dir)
	}
	defaults := params.Default()
	if c.Seed != 0 || c.Days != defaults.Termination.MaxTicks {
		t.Errorf("Seed, Days = %d, %d, want 0, %d", c.Seed, c.Days, defaults.Termination.MaxTicks)
	}
}

func TestParametersExampleFarm(t *testing.T) {
	dir := inputDir(t)
	c := parse(t, "-d", dir)
	p, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	want, err := input.ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	if p.File != want {
		t.Errorf("File = %+v, want the example farm parameters %+v", p.File, want)
	}
	// the input files of --dir are read (input.Load), the map is the example farm [PAR-05]
	if len(p.Input.Species) != 1 || len(p.Input.Foodsources[input.StartFoodsourcesFile].Patches) != 73 {
		t.Errorf("Input: want B_terrestris and the 73 flower patches of the example farm from %s", dir)
	}
	if p.Seed != 0 || p.Termination.MaxTicks != 365 {
		t.Errorf("Seed, MaxTicks = %v, %d, want 0, 365", p.Seed, p.Termination.MaxTicks)
	}
}

func TestParametersImport(t *testing.T) {
	// The parameter file lies outside --dir, as in the validation experiments. The files it names (species,
	// flowers, habitats) and the map are read from --dir, as NetLogo reads them from the folder of the .nlogo.
	dir := inputDir(t)
	text, err := validation.ParameterFiles.ReadFile("experiments/step8_species_Parameters.csv")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "species_Parameters.csv")
	if err := os.WriteFile(file, text, 0o644); err != nil {
		t.Fatal(err)
	}

	c := parse(t, "-p", file, "-d", dir, "-s", "7", "--days", "1")
	p, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	const species = "B_terrestris 150 B_pascuorum 150 B_lapidarius 150 B_hortorum 150 B_pratorum 150 B_hypnorum 150 Psithyrus 100"
	if p.File.BeeSpeciesInitialQueensListAsString != species {
		t.Errorf("BeeSpeciesInitialQueensListAsString = %q, want %q", p.File.BeeSpeciesInitialQueensListAsString, species)
	}
	if p.Seed != 7 || p.Termination.MaxTicks != 1 {
		t.Errorf("Seed, MaxTicks = %v, %d, want 7, 1", p.Seed, p.Termination.MaxTicks)
	}

	a := model.Default(p, nil)
	a.Run()
	derived := ecs.GetResource[globals.Derived](a.World)
	wantSpecies := []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum", "Psithyrus"}
	if !slices.Equal(derived.SpeciesList, wantSpecies) {
		t.Errorf("SpeciesList = %v, want %v", derived.SpeciesList, wantSpecies)
	}
}

func TestParametersMissingFile(t *testing.T) {
	// NetLogo loads _SYSTEM_Parameters.csv instead [RPAR-01]. gobees stops, so that a wrong path does not run
	// silently with other parameters.
	dir := inputDir(t)
	c := parse(t, "-p", filepath.Join(dir, "missing_Parameters.csv"), "-d", dir)
	if _, err := c.parameters(); err == nil {
		t.Error("want an error for a missing parameter file")
	}
}

func TestCLINoMapFlag(t *testing.T) {
	// Step F6: -m (MyMap) is removed, every run starts on the example farm [PAR-05].
	var c cli
	parser, err := newParser(&c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"-m", "_SYSTEM_Example_Farm.png"}); err == nil {
		t.Error("want an error for the removed flag -m")
	}
}
