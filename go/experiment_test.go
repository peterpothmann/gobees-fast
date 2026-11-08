package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"

	"go.yaml.in/yaml/v3"
)

// values returns the items of the YAML flow sequence text as nodes, e.g. values(t, "[0, 2]").
func values(t *testing.T, text string) []yaml.Node {
	t.Helper()
	var n yaml.Node
	if err := yaml.Unmarshal([]byte(text), &n); err != nil {
		t.Fatal(err)
	}
	var items []yaml.Node
	for _, c := range n.Content[0].Content {
		items = append(items, *c)
	}
	return items
}

// writeExampleFarm writes the example farm as file into dir, with the key name, the section run and the variations
// given.
func writeExampleFarm(t *testing.T, dir, file, name string, run input.Run, variations []input.Variation) string {
	t.Helper()
	cfg, landscapes := input.ExampleFarmConfig()
	cfg.Name = name
	cfg.Run = &run
	cfg.Variations = variations
	writeRunFile(t, dir, file, cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	return filepath.Join(dir, file)
}

func TestExperimentOutput(t *testing.T) {
	// 2 values of badgers x 2 repetitions: one file, the column of the varied key after run and seed, the runs
	// numbered over both combinations, each combination with the same seeds. The rows of a combination are those of
	// the run file with that value, and the file does not depend on the number of threads.
	dir := t.TempDir()
	days := 30
	experiment := writeExampleFarm(t, dir, "experiment.yaml", "badgers",
		input.Run{Seed: 4, Days: &days, Repetitions: 2, Threads: 1},
		[]input.Variation{{Key: "parameters.badgers", Values: values(t, "[0, 3]")}})

	var files [][]byte
	var summary string
	for _, threads := range []string{"1", "4"} {
		out := filepath.Join(dir, "threads"+threads)
		c := parse(t, "-c", experiment, "-o", out, "--threads", threads, "--columns", "colonies,badgers")
		var b bytes.Buffer
		if err := c.run(&b); err != nil {
			t.Fatal(err)
		}
		daily, _ := outputOf(out, "badgers")
		files = append(files, readFile(t, daily))
		summary = b.String()
	}
	if !bytes.Equal(files[0], files[1]) {
		t.Error("the output with 1 thread differs from the output with 4 threads")
	}
	if !strings.Contains(summary, "runs: 4, seeds: 4 to 5, threads: 4, combinations: 2, repetitions: 2") ||
		!strings.Contains(summary, "run 4 (seed 5, parameters.badgers=3): colonies founded: ") {
		t.Errorf("summary = %q", summary)
	}

	lines := strings.Split(strings.TrimSuffix(string(files[0]), "\n"), "\n")
	if lines[0] != "run\tseed\tparameters.badgers\tt\tticks\tcolonies\tbadgers" {
		t.Errorf("header = %q", lines[0])
	}
	if len(lines) != 1+4*days {
		t.Fatalf("%d lines, want a header and 4 runs of %d days", len(lines), days)
	}
	for i, prefix := range []string{"1\t4\t0\t", "2\t5\t0\t", "3\t4\t3\t", "4\t5\t3\t"} {
		row := lines[1+i*days]
		if !strings.HasPrefix(row, prefix) {
			t.Errorf("first row of run %d = %q, want %q in front", i+1, row, prefix)
		}
	}

	// the combination with 3 badgers is the run file with badgers: 3
	cfg, landscapes := input.ExampleFarmConfig()
	cfg.Run = &input.Run{Seed: 4, Days: &days, Repetitions: 2, Threads: 1}
	cfg.Parameters.Badgers = 3
	writeRunFile(t, dir, "three.yaml", cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	single := filepath.Join(dir, "three")
	c := parse(t, "-c", filepath.Join(dir, "three.yaml"), "-o", single, "--columns", "colonies,badgers")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	singleDaily, _ := outputOf(single, cfg.Name)
	if got, want := dropFields(t, files[0], 3)[2*days:], dropFields(t, readFile(t, singleDaily), 2); !slices.Equal(got, want) {
		t.Error("the combination badgers=3 differs from the run file with badgers: 3")
	}

	// the provenance is the experiment: repeated from it, the same file
	_, prov := outputOf(filepath.Join(dir, "threads1"), "badgers")
	again := filepath.Join(dir, "again")
	c = parse(t, "-c", prov, "-o", again)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	againDaily, _ := outputOf(again, "badgers")
	if !bytes.Equal(files[0], readFile(t, againDaily)) {
		t.Error("the experiment repeated from the provenance has a different daily output")
	}
}

func TestSeveralRunFiles(t *testing.T) {
	// two run files in one call: one pool of runs, an output folder per run file named by its key name, each output
	// equal to the run file run alone. The two run files have the same file name in different folders.
	dir := t.TempDir()
	days := 20
	for _, sub := range []string{"x", "y"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	a := writeExampleFarm(t, filepath.Join(dir, "x"), "run.yaml", "farm-a",
		input.Run{Seed: 1, Days: &days, Repetitions: 2, Threads: 3}, nil)
	b := writeExampleFarm(t, filepath.Join(dir, "y"), "run.yaml", "farm-b",
		input.Run{Seed: 7, Days: &days, Repetitions: 1, Threads: 3},
		[]input.Variation{{Key: "parameters.sexLocus", Values: values(t, "[false, true]")}})

	out := filepath.Join(dir, "out")
	c := parse(t, "-c", a, "-c", b, "-o", out)
	var summary bytes.Buffer
	if err := c.run(&summary); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ config, name string }{{a, "farm-a"}, {b, "farm-b"}} {
		daily, prov := outputOf(out, tc.name)
		alone := filepath.Join(dir, "alone")
		c := parse(t, "-c", tc.config, "-o", alone)
		if err := c.run(&bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		aloneDaily, _ := outputOf(alone, tc.name)
		if !bytes.Equal(readFile(t, daily), readFile(t, aloneDaily)) {
			t.Errorf("%s differs from the run file run alone", daily)
		}
		if _, err := input.Read(prov); err != nil {
			t.Errorf("provenance of %s: %v", tc.name, err)
		}
	}
	if !strings.Contains(summary.String(), "farm-a:\n") || !strings.Contains(summary.String(), "farm-b:\n") {
		t.Errorf("summary = %q, want a part per run file", summary.String())
	}

	// different threads in the run files: an error unless --threads is set
	d := writeExampleFarm(t, dir, "d.yaml", "d", input.Run{Seed: 1, Days: &days, Repetitions: 1, Threads: 2}, nil)
	c = parse(t, "-c", a, "-c", d)
	if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "--threads") {
		t.Errorf("different threads: error = %v", err)
	}
	c = parse(t, "-c", a, "-c", d, "--threads", "2")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Errorf("different threads with --threads: %v", err)
	}

	// two run files with the same key name would write into the same folder, also when the case differs
	for _, name := range []string{"farm-a", "Farm-A"} {
		same := writeExampleFarm(t, dir, "same.yaml", name,
			input.Run{Seed: 1, Days: &days, Repetitions: 1, Threads: 3}, nil)
		c = parse(t, "-c", a, "-c", same, "-o", filepath.Join(dir, "same"))
		if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "same name") {
			t.Errorf("name %s: error = %v, want same name", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "same")); err == nil {
		t.Error("output written although two run files have the same name")
	}
}
