package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"
	"gobees/obs"
	"gobees/params"
	"gobees/schedule"

	"github.com/mlange-42/ark/ecs"
)

// parameters returns the parameters of the first combination of the first run file of the flags (the example farm
// without -c), with the flags applied, and the files read.
func (c *cli) parameters() (params.Params, input.Files, error) {
	config := ""
	if len(c.Config) > 0 {
		config = c.Config[0]
	}
	e, err := c.newExperiment(config)
	if err != nil {
		return params.Params{}, input.Files{}, err
	}
	return e.params[0], e.files, nil
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
	if len(c.Config) != 0 {
		t.Errorf("Config = %q, want none (example farm)", c.Config)
	}
	// not set: the section run of the run file applies (TestParametersRunSection)
	if c.Seed != nil || c.Days != nil || c.Runs != nil || c.Threads != nil {
		t.Errorf("Seed, Days, Runs, Threads = %v, %v, %v, %v, want nil", c.Seed, c.Days, c.Runs, c.Threads)
	}
	// days set to 0: no limit, also against the run file
	c = parse(t, "-s", "2", "--days", "0")
	if c.Seed == nil || *c.Seed != 2 || c.Days == nil || *c.Days != 0 {
		t.Errorf("Seed, Days = %v, %v, want 2, 0", c.Seed, c.Days)
	}
}

func TestCLIColumns(t *testing.T) {
	// --columns is comma separated, in the order given. Not set: nil, the run file or all columns apply
	if c := parse(t); c.Columns != nil || c.ListColumns {
		t.Errorf("Columns, ListColumns = %v, %v, want nil, false", c.Columns, c.ListColumns)
	}
	c := parse(t, "--columns", "eggs,colonies")
	if !slices.Equal(c.Columns, []string{"eggs", "colonies"}) {
		t.Errorf("Columns = %v, want [eggs colonies]", c.Columns)
	}
	p, _, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.Output.Columns, []string{"eggs", "colonies"}) {
		t.Errorf("Output.Columns = %v, want the flag", p.Output.Columns)
	}
}

func TestColumnsUnknown(t *testing.T) {
	// an unknown name is an error with the list of columns, from the flag and from the run file, also without output
	c := parse(t, "--columns", "eggs,egg")
	if _, _, err := c.parameters(); err == nil || !strings.Contains(err.Error(), "larvae") {
		t.Errorf("flag: error = %v, want the name and the list of columns", err)
	}

	cfg, landscapes := input.ExampleFarmConfig()
	cfg.Output.Columns = []string{"nonsense"}
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	c = parse(t, "-c", filepath.Join(dir, "run.yaml"), "--days", "1")
	err := c.run(io.Discard)
	if err == nil || !strings.Contains(err.Error(), `"nonsense"`) {
		t.Errorf("run file: error = %v, want the unknown name", err)
	}
}

func TestColumnsNeedOutput(t *testing.T) {
	c := parse(t, "--columns", "eggs", "--days", "1")
	if err := c.run(io.Discard); err == nil {
		t.Error("want an error for --columns without -o or -f")
	}
}

func TestListColumns(t *testing.T) {
	var b bytes.Buffer
	c := parse(t, "--list-columns")
	if err := c.run(&b); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	all := obs.AllColumns()
	if len(lines) != len(all) {
		t.Fatalf("%d lines, want %d", len(lines), len(all))
	}
	for i, name := range all {
		if !strings.HasPrefix(lines[i], name+" ") || !strings.HasSuffix(lines[i], obs.ColumnDoc(name)) {
			t.Errorf("line %d = %q, want the name %s and its description", i, lines[i], name)
		}
	}
}

func TestParametersRunSection(t *testing.T) {
	// the section run of a run file sets seed and days, the flags override it
	c, landscapes := input.ExampleFarmConfig()
	days := 3
	c.Run = &input.Run{Seed: 5, Days: &days, Runs: 1, Threads: 1}
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

	// the section run and its keys seed, runs and threads are needed, each 1 or more
	for _, tc := range []struct {
		run       *input.Run
		situation string
	}{
		{nil, "no section run"},
		{&input.Run{Runs: 1, Threads: 1}, "no seed"},
		{&input.Run{Seed: 1, Threads: 1}, "no runs"},
		{&input.Run{Seed: 1, Runs: 1}, "no threads"},
		{&input.Run{Seed: -1, Runs: 1, Threads: 1}, "seed below 1"},
	} {
		c.Run = tc.run
		writeRunFile(t, dir, "run.yaml", c, map[string]input.Landscape{c.Landscape: landscapes[c.Landscape]})
		cli := parse(t, "-c", path)
		if _, _, err := cli.parameters(); err == nil || !strings.Contains(err.Error(), "run") {
			t.Errorf("%s: error = %v, want an error about the section run", tc.situation, err)
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
	if p.Seed != 1 || p.Termination.MaxTicks != 365 {
		t.Errorf("Seed, MaxTicks = %v, %d, want 1, 365 (run.seed of the example farm)", p.Seed, p.Termination.MaxTicks)
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
