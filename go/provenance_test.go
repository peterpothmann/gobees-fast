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

// outputOf returns the daily output and the provenance of the run file with the key name in the output folder out.
func outputOf(out, name string) (daily, provenance string) {
	return filepath.Join(out, name, dailyOutputFile), filepath.Join(out, name, provenanceFile)
}

func TestProvenanceRepeatsRun(t *testing.T) {
	// A run with output writes the provenance. The provenance as run file repeats the run: the same daily output, byte
	// for byte. The example farm and a run file with crop rotation over the turn of the year (two landscape files,
	// the first one is also the start).
	for _, tc := range []struct {
		name       string // the key name of the run file
		args       []string
		config     string
		landscapes []string // the landscape files in the run file of the run
	}{
		{"example_farm", []string{"-s", "6", "--days", "5"}, exampleFarmConfig,
			[]string{"_SYSTEM_Example_Farm_Foodsources.yaml"}},
		{"step7_crop", []string{"-c", filepath.Join("compare", "scenarios", "step7_crop.yaml"), "-s", "3",
			"--days", "370"}, filepath.Join("compare", "scenarios", "step7_crop.yaml"),
			[]string{"_SYSTEM_Example_Farm_Foodsources.yaml", "Example_Farm_PlotTest_Foodsources.yaml"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			first := filepath.Join(dir, "first")
			c := parse(t, append(tc.args, "-o", first)...)
			if err := c.run(io.Discard); err != nil {
				t.Fatal(err)
			}

			daily, prov := outputOf(first, tc.name)
			info, err := input.Read(prov)
			if err != nil {
				t.Fatalf("the provenance is no valid run file: %v", err)
			}
			if info.Config.Name != tc.name {
				t.Errorf("name = %q, want %q", info.Config.Name, tc.name)
			}
			run := info.Config.Run
			if run == nil {
				t.Fatal("provenance without section run")
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
			if run.Config != config || run.Output != daily || run.Version == nil || run.Started == "" ||
				len(run.Command) == 0 || !slices.Equal(info.Config.Output.Columns, obs.AllColumns()) {
				t.Errorf("run = %+v, want the record of the run", run)
			}

			// one byte-for-byte copy per landscape file with its name, with its SHA-256
			if len(run.Landscapes) != len(tc.landscapes) {
				t.Fatalf("landscapes = %+v, want copies of %v", run.Landscapes, tc.landscapes)
			}
			for i, l := range run.Landscapes {
				if l.Source != tc.landscapes[i] || l.File != tc.landscapes[i] {
					t.Errorf("landscape %d = %+v, want the copy %s", i, l, tc.landscapes[i])
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
			second := filepath.Join(dir, "second")
			c = parse(t, "-c", prov, "-o", second)
			if err := c.run(io.Discard); err != nil {
				t.Fatal(err)
			}
			secondDaily, secondProv := outputOf(second, tc.name)
			if !bytes.Equal(readFile(t, daily), readFile(t, secondDaily)) {
				t.Error("the run repeated from the provenance has a different daily output")
			}
			repeated, err := input.Read(secondProv)
			if err != nil {
				t.Fatal(err)
			}
			if repeated.Config.Run.Seed != run.Seed {
				t.Errorf("seed of the repeated run = %d, want %d", repeated.Config.Run.Seed, run.Seed)
			}
		})
	}
}

func TestProvenanceLandscapeNames(t *testing.T) {
	// The copies keep the names of the landscape files. A second file of the same name (from another folder) gets
	// _2, and no copy takes the name of the daily output or the provenance.
	cfg, landscapes := input.ExampleFarmConfig()
	farm := landscapes[cfg.Landscape]
	cfg.Landscape = "a/farm.yaml"
	cfg.CropRotation = []string{"a/farm.yaml", "b/farm.yaml", "c/provenance.yaml", "c/daily.csv"}
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", cfg, map[string]input.Landscape{"a/farm.yaml": farm, "b/farm.yaml": farm,
		"c/provenance.yaml": farm, "c/daily.csv": farm})

	out := filepath.Join(dir, "out")
	c := parse(t, "-c", filepath.Join(dir, "run.yaml"), "--days", "2", "-o", out)
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	_, prov := outputOf(out, cfg.Name)
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"farm.yaml", "farm_2.yaml", "provenance_2.yaml", "daily_2.csv"}
	if info.Config.Landscape != want[0] || !slices.Equal(info.Config.CropRotation, want) {
		t.Errorf("landscape, cropRotation = %q, %q, want %q", info.Config.Landscape, info.Config.CropRotation, want)
	}
	for i, l := range info.Config.Run.Landscapes {
		if l.File != want[i] || l.Source != cfg.CropRotation[i] {
			t.Errorf("landscape %d = %+v, want the copy %s of %s", i, l, want[i], cfg.CropRotation[i])
		}
	}
}

func TestProvenanceColumns(t *testing.T) {
	// A selection of columns (flag): the output holds t, ticks and these columns, the provenance holds them in
	// output.columns, and the run repeated from the provenance writes the same file.
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	c := parse(t, "--days", "5", "-s", "4", "-o", first, "--columns", "colonies,eggs,TotalColonies")
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	firstDaily, prov := outputOf(first, "example_farm")
	want := []string{"ticks", "colonies", "eggs", "TotalColonies"}
	if got := header(t, firstDaily); !slices.Equal(got, append([]string{"run", "seed", "t"}, want...)) {
		t.Errorf("header = %v, want run, seed, t and %v", got, want)
	}
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(info.Config.Output.Columns, want) {
		t.Errorf("output.columns = %v, want %v", info.Config.Output.Columns, want)
	}

	second := filepath.Join(dir, "second")
	c = parse(t, "-c", prov, "-o", second)
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	secondDaily, _ := outputOf(second, "example_farm")
	if !bytes.Equal(readFile(t, firstDaily), readFile(t, secondDaily)) {
		t.Error("the run repeated from the provenance has a different daily output")
	}

	// the flag overrides output.columns of the run file
	third := filepath.Join(dir, "third")
	c = parse(t, "-c", prov, "-o", third, "--columns", "males")
	if err := c.run(io.Discard); err != nil {
		t.Fatal(err)
	}
	thirdDaily, _ := outputOf(third, "example_farm")
	if got := header(t, thirdDaily); !slices.Equal(got, []string{"run", "seed", "t", "ticks", "males"}) {
		t.Errorf("header = %v, want run, seed, t, ticks, males", got)
	}
}

// header returns the column names in the first row of the output file at path.
func header(t *testing.T, path string) []string {
	t.Helper()
	first, _, _ := strings.Cut(string(readFile(t, path)), "\n")
	return strings.Split(strings.TrimSuffix(first, "\r"), "\t")
}

func TestProvenanceOnlyWithOutput(t *testing.T) {
	// without -o, no file is written
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
