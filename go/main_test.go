package main

import (
	"bytes"
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
	// not set: the section run of the run file or the defaults apply (TestParametersRunSection)
	if c.Seed != nil || c.Days != nil {
		t.Errorf("Seed, Days = %v, %v, want nil, nil", c.Seed, c.Days)
	}
	// set to 0: a random seed and no limit, also against the run file
	c = parse(t, "-s", "0", "--days", "0")
	if c.Seed == nil || *c.Seed != 0 || c.Days == nil || *c.Days != 0 {
		t.Errorf("Seed, Days = %v, %v, want 0, 0", c.Seed, c.Days)
	}
}

func TestParametersRunSection(t *testing.T) {
	// the section run of a run file sets seed and days, the flags override it
	c, landscapes := input.ExampleFarmConfig()
	days := 3
	c.Run = &input.Run{Seed: 5, Days: &days}
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", c, map[string]input.Landscape{c.Landscape: landscapes[c.Landscape]})
	path := filepath.Join(dir, "run.yaml")

	for _, tc := range []struct {
		args      []string
		seed      int64
		days      int
		situation string
	}{
		{nil, 5, 3, "no flags: the section run"},
		{[]string{"-s", "9", "--days", "0"}, 9, 0, "flags override the section run"},
		{[]string{"-s", "0"}, 0, 3, "-s 0: a random seed, drawn in run"},
	} {
		cli := parse(t, append([]string{"-c", path}, tc.args...)...)
		p, _, err := cli.parameters()
		if err != nil {
			t.Fatal(err)
		}
		if p.Seed != tc.seed || p.Termination.MaxTicks != tc.days {
			t.Errorf("%s: Seed, MaxTicks = %d, %d, want %d, %d", tc.situation, p.Seed, p.Termination.MaxTicks,
				tc.seed, tc.days)
		}
	}
}

// writeRunFile writes the run file c as name and its landscape files into dir.
func writeRunFile(t *testing.T, dir, name string, c input.Config, landscapes map[string]input.Landscape) {
	t.Helper()
	var b bytes.Buffer
	if err := input.WriteConfig(&b, c); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	for n, l := range landscapes {
		b.Reset()
		if err := input.WriteLandscape(&b, l); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, n)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, n), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParametersExampleFarm(t *testing.T) {
	// Without a run file: the example farm built into package input.
	c := parse(t)
	p, _, err := c.parameters()
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
	writeRunFile(t, dir, "run.yaml", c, map[string]input.Landscape{c.Landscape: landscape})

	cli := parse(t, "-c", filepath.Join(dir, "run.yaml"), "-s", "7", "--days", "1")
	p, _, err := cli.parameters()
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
	// a missing run file is an error, so that a wrong path does not run silently with other parameters
	c := parse(t, "-c", filepath.Join(t.TempDir(), "missing.yaml"))
	if _, _, err := c.parameters(); err == nil {
		t.Error("want an error for a missing run file")
	}
}

func TestCLIRemovedFlags(t *testing.T) {
	// the flags of older versions (-m, -p, -d) are errors: the landscape and all parameters are in the run file
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
