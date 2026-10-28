package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/globals"
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
	if c.Map != defaults.Interface.MyMap || c.Seed != 0 || c.Days != defaults.Termination.MaxTicks {
		t.Errorf("Map, Seed, Days = %q, %d, %d, want %q, 0, %d",
			c.Map, c.Seed, c.Days, defaults.Interface.MyMap, defaults.Termination.MaxTicks)
	}
}

func TestParametersExampleFarm(t *testing.T) {
	dir := inputDir(t)
	c := parse(t, "-d", dir)
	p, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	want, err := params.ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	if p.File != want {
		t.Errorf("File = %+v, want the example farm parameters %+v", p.File, want)
	}
	if p.WorkingDirectory.Path != dir {
		t.Errorf("WorkingDirectory = %q, want %q", p.WorkingDirectory.Path, dir)
	}
	if p.Interface.MyMap != "_SYSTEM_Example_Farm.png" || p.Interface.RAND_SEED != 0 || p.Termination.MaxTicks != 365 {
		t.Errorf("MyMap, RAND_SEED, MaxTicks = %q, %v, %d, want _SYSTEM_Example_Farm.png, 0, 365",
			p.Interface.MyMap, p.Interface.RAND_SEED, p.Termination.MaxTicks)
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

	c := parse(t, "-p", file, "-d", dir, "-m", "_SYSTEM_Example_Farm.png", "-s", "7", "--days", "1")
	p, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	const species = "B_terrestris 150 B_pascuorum 150 B_lapidarius 150 B_hortorum 150 B_pratorum 150 B_hypnorum 150 Psithyrus 100"
	if p.File.BeeSpeciesInitialQueensListAsString != species {
		t.Errorf("BeeSpeciesInitialQueensListAsString = %q, want %q", p.File.BeeSpeciesInitialQueensListAsString, species)
	}
	if p.Interface.RAND_SEED != 7 || p.Termination.MaxTicks != 1 {
		t.Errorf("RAND_SEED, MaxTicks = %v, %d, want 7, 1", p.Interface.RAND_SEED, p.Termination.MaxTicks)
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
