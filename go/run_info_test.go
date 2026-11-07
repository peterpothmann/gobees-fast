package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"
	"gobees/obs"
)

func TestRunInfoPath(t *testing.T) {
	for _, tc := range []struct{ output, want string }{
		{"daily_output.csv", "daily_output_run_info.yaml"},
		{filepath.Join("out", "a.b.tsv"), filepath.Join("out", "a.b_run_info.yaml")},
		{filepath.Join("out", "result"), filepath.Join("out", "result_run_info.yaml")},
	} {
		if got := runInfoPath(tc.output); got != tc.want {
			t.Errorf("runInfoPath(%q) = %q, want %q", tc.output, got, tc.want)
		}
	}
}

func TestRunInfoRepeatsRun(t *testing.T) {
	// A run with output writes the run info. The run info as run file repeats the run: the same daily output, byte
	// for byte. The example farm and a run file with crop rotation over the turn of
	// the year (two landscape files, the first one is also the start).
	for _, tc := range []struct {
		name       string
		args       []string
		config     string
		landscapes []string // the landscape files in the run file of the run
	}{
		{"example farm", []string{"-s", "6", "--days", "5"}, exampleFarmConfig,
			[]string{"_SYSTEM_Example_Farm_Foodsources.yaml"}},
		{"crop rotation", []string{"-c", filepath.Join("compare", "scenarios", "step7_crop.yaml"), "-s", "3",
			"--days", "370"}, filepath.Join("compare", "scenarios", "step7_crop.yaml"),
			[]string{"_SYSTEM_Example_Farm_Foodsources.yaml", "Example_Farm_PlotTest_Foodsources.yaml"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			first := filepath.Join(dir, "first", "daily_output.csv")
			c := parse(t, append(tc.args, "-f", first)...)
			if err := c.run(io.Discard); err != nil {
				t.Fatal(err)
			}

			info, err := input.Read(runInfoPath(first))
			if err != nil {
				t.Fatalf("the run info is no valid run file: %v", err)
			}
			run := info.Config.Run
			if run == nil {
				t.Fatal("run info without section run")
			}
			if run.Seed != *c.Seed || run.Days == nil || *run.Days != *c.Days {
				t.Errorf("run.seed, run.days = %d, %v, want %d and %d", run.Seed, run.Days, *c.Seed, *c.Days)
			}
			config := tc.config
			if config != exampleFarmConfig {
				config, err = filepath.Abs(config) // kong makes paths absolute
				if err != nil {
					t.Fatal(err)
				}
			}
			if run.Config != config || run.Output != first || run.Version == nil || run.Started == "" ||
				len(run.Command) == 0 || !slices.Equal(info.Config.Output.Columns, obs.AllColumns()) {
				t.Errorf("run = %+v, want the record of the run", run)
			}

			// one byte-for-byte copy per landscape file, with its SHA-256
			if len(run.Landscapes) != len(tc.landscapes) {
				t.Fatalf("landscapes = %+v, want copies of %v", run.Landscapes, tc.landscapes)
			}
			for i, l := range run.Landscapes {
				if l.Source != tc.landscapes[i] || l.File != "daily_output_"+tc.landscapes[i] {
					t.Errorf("landscape %d = %+v, want the copy daily_output_%s", i, l, tc.landscapes[i])
				}
				raw := info.Raw[l.File]
				sum := sha256.Sum256(raw)
				if hex.EncodeToString(sum[:]) != l.SHA256 {
					t.Errorf("%s: SHA-256 differs from the copy", l.File)
				}
			}
			if info.Config.Landscape != run.Landscapes[0].File {
				t.Errorf("landscape = %q, want the copy %q", info.Config.Landscape, run.Landscapes[0].File)
			}
			for _, n := range info.Config.CropRotation {
				if _, ok := info.Raw[n]; !ok {
					t.Errorf("cropRotation %q is not a copy", n)
				}
			}

			// repeated without -s and --days: seed and days come from the section run
			second := filepath.Join(dir, "second", "daily_output.csv")
			c = parse(t, "-c", runInfoPath(first), "-f", second)
			if err := c.run(io.Discard); err != nil {
				t.Fatal(err)
			}
			a, b := readFile(t, first), readFile(t, second)
			if !bytes.Equal(a, b) {
				t.Error("the run repeated from the run info has a different daily output")
			}
			repeated, err := input.Read(runInfoPath(second))
			if err != nil {
				t.Fatal(err)
			}
			if repeated.Config.Run.Seed != run.Seed {
				t.Errorf("seed of the repeated run = %d, want %d", repeated.Config.Run.Seed, run.Seed)
			}
		})
	}
}

func TestRunInfoColumns(t *testing.T) {
	// A selection of columns (flag): the output holds t, ticks and these columns, the run info holds them in
	// output.columns, and the run repeated from the run info writes the same file.
	dir := t.TempDir()
	first := filepath.Join(dir, "first.csv")
	c := parse(t, "--days", "5", "-s", "4", "-f", first, "--columns", "colonies,eggs,TotalColonies")
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	want := []string{"ticks", "colonies", "eggs", "TotalColonies"}
	if got := header(t, first); !slices.Equal(got, append([]string{"run", "seed", "t"}, want...)) {
		t.Errorf("header = %v, want run, seed, t and %v", got, want)
	}
	info, err := input.Read(runInfoPath(first))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(info.Config.Output.Columns, want) {
		t.Errorf("output.columns = %v, want %v", info.Config.Output.Columns, want)
	}

	second := filepath.Join(dir, "second.csv")
	c = parse(t, "-c", runInfoPath(first), "-f", second)
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readFile(t, first), readFile(t, second)) {
		t.Error("the run repeated from the run info has a different daily output")
	}

	// the flag overrides output.columns of the run file
	third := filepath.Join(dir, "third.csv")
	c = parse(t, "-c", runInfoPath(first), "-f", third, "--columns", "males")
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := header(t, third); !slices.Equal(got, []string{"run", "seed", "t", "ticks", "males"}) {
		t.Errorf("header = %v, want run, seed, t, ticks, males", got)
	}
}

// header returns the column names in the first row of the output file at path.
func header(t *testing.T, path string) []string {
	t.Helper()
	first, _, _ := strings.Cut(string(readFile(t, path)), "\n")
	return strings.Split(strings.TrimSuffix(first, "\r"), "\t")
}

func TestRunInfoOnlyWithOutput(t *testing.T) {
	// without -o and -f, no file is written
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	c := parse(t, "--days", "1", "-c", filepath.Join(wd, "compare", "scenarios", "step7.yaml"))
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("files written without output: %v", entries)
	}
}

// readFile returns the content of the file at path.
func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
