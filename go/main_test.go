package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobees/input"
	"gobees/params"
	"gobees/schedule"

	"github.com/mlange-42/ark/ecs"
)

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
	if c.Config != "" {
		t.Errorf("Config = %q, want empty (example farm)", c.Config)
	}
	defaults := params.Default()
	if c.Seed != 0 || c.Days != defaults.Termination.MaxTicks {
		t.Errorf("Seed, Days = %d, %d, want 0, %d", c.Seed, c.Days, defaults.Termination.MaxTicks)
	}
}

func TestParametersExampleFarm(t *testing.T) {
	// Without a run file: the example farm built into package input.
	c := parse(t)
	p, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	want := input.ExampleFarm()
	if p.Parameters != want.Parameters {
		t.Errorf("File = %+v, want the example farm parameters %+v", p.Parameters, want.Parameters)
	}
	if len(p.Input.Species) != 1 || len(p.Input.Foodsources[p.Input.Derived.StartLandscape].Patches) != 73 {
		t.Errorf("Input: want B_terrestris and the 73 flower patches of the example farm")
	}
	if p.Seed != 0 || p.Termination.MaxTicks != 365 {
		t.Errorf("Seed, MaxTicks = %v, %d, want 0, 365", p.Seed, p.Termination.MaxTicks)
	}
}

func TestParametersRunFile(t *testing.T) {
	// A run file with all 7 species. Its landscape file lies in a subfolder, relative to the run file.
	c, landscapes := input.ExampleFarmConfig()
	c.InitialQueens = []params.SpeciesQueens{{Species: "B_terrestris", Queens: 150}, {Species: "B_pascuorum", Queens: 150},
		{Species: "B_lapidarius", Queens: 150}, {Species: "B_hortorum", Queens: 150}, {Species: "B_pratorum", Queens: 150},
		{Species: "B_hypnorum", Queens: 150}, {Species: "Psithyrus", Queens: 100}}
	landscape := landscapes[c.Landscape]
	c.Landscape = "maps/farm.yaml"

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "maps"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, w func(*os.File) error) {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := w(f); err != nil {
			t.Fatal(err)
		}
	}
	write("run.yaml", func(f *os.File) error { return input.WriteConfig(f, c) })
	write(filepath.Join("maps", "farm.yaml"), func(f *os.File) error { return input.WriteLandscape(f, landscape) })

	cli := parse(t, "-c", filepath.Join(dir, "run.yaml"), "-s", "7", "--days", "1")
	p, err := cli.parameters()
	if err != nil {
		t.Fatal(err)
	}
	if p.Seed != 7 || p.Termination.MaxTicks != 1 {
		t.Errorf("Seed, MaxTicks = %v, %d, want 7, 1", p.Seed, p.Termination.MaxTicks)
	}

	a := schedule.Default(p, nil)
	a.Run()
	derived := ecs.GetResource[params.Derived](a.World)
	wantSpecies := []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum", "Psithyrus"}
	if !slices.Equal(derived.SpeciesNames, wantSpecies) {
		t.Errorf("SpeciesNames = %v, want %v", derived.SpeciesNames, wantSpecies)
	}
}

func TestParametersMissingFile(t *testing.T) {
	// NetLogo loads _SYSTEM_Parameters.csv instead of a missing parameter file [RPAR-01]. gobees-fast stops, so that
	// a wrong path does not run silently with other parameters.
	c := parse(t, "-c", filepath.Join(t.TempDir(), "missing.yaml"))
	if _, err := c.parameters(); err == nil {
		t.Error("want an error for a missing run file")
	}
}

func TestCLIRemovedFlags(t *testing.T) {
	// -m (MyMap) is removed in step F6, -p and -d (NetLogo files) in step F7. The landscape is in the run file.
	for _, args := range [][]string{{"-m", "_SYSTEM_Example_Farm.png"}, {"-p", "x.csv"}, {"-d", "."}} {
		var c cli
		parser, err := newParser(&c)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.Parse(args); err == nil {
			t.Errorf("want an error for the removed flag %s", args[0])
		}
	}
}
