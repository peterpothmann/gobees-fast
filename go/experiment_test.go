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

// writeExampleFarm writes the example farm as name into dir, with the section run and the variations given.
func writeExampleFarm(t *testing.T, dir, name string, run input.Run, variations []input.Variation) string {
	t.Helper()
	cfg, landscapes := input.ExampleFarmConfig()
	cfg.Run = &run
	cfg.Variations = variations
	writeRunFile(t, dir, name, cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	return filepath.Join(dir, name)
}

func TestExperimentOutput(t *testing.T) {
	// 2 values of badgers x 2 runs: one file, the column of the varied key after run and seed, each combination
	// with the same seeds. The rows of a combination are those of the run file with that value, and the file does
	// not depend on the number of threads.
	dir := t.TempDir()
	days := 30
	experiment := writeExampleFarm(t, dir, "experiment.yaml", input.Run{Seed: 4, Days: &days, Runs: 2, Threads: 1},
		[]input.Variation{{Key: "parameters.badgers", Values: values(t, "[0, 3]")}})

	var files [][]byte
	var summary string
	for _, threads := range []string{"1", "4"} {
		out := filepath.Join(dir, "threads"+threads+".csv")
		c := parse(t, "-c", experiment, "-f", out, "--threads", threads, "--columns", "colonies,badgers")
		var b bytes.Buffer
		if err := c.run(&b); err != nil {
			t.Fatal(err)
		}
		files = append(files, readFile(t, out))
		summary = b.String()
	}
	if !bytes.Equal(files[0], files[1]) {
		t.Error("the output with 1 thread differs from the output with 4 threads")
	}
	if !strings.Contains(summary, "runs: 2, seeds: 4 to 5, threads: 4, combinations: 2") ||
		!strings.Contains(summary, "run 2 (seed 5, parameters.badgers=3): colonies founded: ") {
		t.Errorf("summary = %q", summary)
	}

	lines := strings.Split(strings.TrimSuffix(string(files[0]), "\n"), "\n")
	if lines[0] != "run\tseed\tparameters.badgers\tt\tticks\tcolonies\tbadgers" {
		t.Errorf("header = %q", lines[0])
	}
	if len(lines) != 1+4*days {
		t.Fatalf("%d lines, want a header and 4 runs of %d days", len(lines), days)
	}
	for i, prefix := range []string{"1\t4\t0\t", "2\t5\t0\t", "1\t4\t3\t", "2\t5\t3\t"} {
		row := lines[1+i*days]
		if !strings.HasPrefix(row, prefix) {
			t.Errorf("first row of run %d = %q, want %q in front", i+1, row, prefix)
		}
	}

	// the combination with 3 badgers is the run file with badgers: 3
	cfg, landscapes := input.ExampleFarmConfig()
	cfg.Run = &input.Run{Seed: 4, Days: &days, Runs: 2, Threads: 1}
	cfg.Parameters.Badgers = 3
	writeRunFile(t, dir, "three.yaml", cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	single := filepath.Join(dir, "three.csv")
	c := parse(t, "-c", filepath.Join(dir, "three.yaml"), "-f", single, "--columns", "colonies,badgers")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got, want := dropFields(t, files[0], 3)[2*days:], dropFields(t, readFile(t, single), 2); !slices.Equal(got, want) {
		t.Error("the combination badgers=3 differs from the run file with badgers: 3")
	}

	// the run info is the experiment: repeated from it, the same file
	again := filepath.Join(dir, "again.csv")
	c = parse(t, "-c", runInfoPath(filepath.Join(dir, "threads1.csv")), "-f", again)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(files[0], readFile(t, again)) {
		t.Error("the experiment repeated from the run info has a different daily output")
	}
}

func TestSeveralRunFiles(t *testing.T) {
	// two run files in one call: one pool of runs, an output and a run info per run file, each equal to the run file
	// run alone
	dir := t.TempDir()
	days := 20
	a := writeExampleFarm(t, dir, "a.yaml", input.Run{Seed: 1, Days: &days, Runs: 2, Threads: 3}, nil)
	b := writeExampleFarm(t, dir, "b.yaml", input.Run{Seed: 7, Days: &days, Runs: 1, Threads: 3},
		[]input.Variation{{Key: "parameters.sexLocus", Values: values(t, "[false, true]")}})

	out := filepath.Join(dir, "out", "daily.csv")
	c := parse(t, "-c", a, "-c", b, "-f", out)
	var summary bytes.Buffer
	if err := c.run(&summary); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ config, output string }{{a, "daily_a.csv"}, {b, "daily_b.csv"}} {
		got := readFile(t, filepath.Join(dir, "out", tc.output))
		alone := filepath.Join(dir, "alone_"+tc.output)
		c := parse(t, "-c", tc.config, "-f", alone)
		if err := c.run(&bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, readFile(t, alone)) {
			t.Errorf("%s differs from the run file run alone", tc.output)
		}
		if _, err := input.Read(runInfoPath(filepath.Join(dir, "out", tc.output))); err != nil {
			t.Errorf("run info of %s: %v", tc.output, err)
		}
	}
	if !strings.Contains(summary.String(), a+":\n") || !strings.Contains(summary.String(), b+":\n") {
		t.Errorf("summary = %q, want a part per run file", summary.String())
	}

	// different threads in the run files: an error unless --threads is set
	d := writeExampleFarm(t, dir, "d.yaml", input.Run{Seed: 1, Days: &days, Runs: 1, Threads: 2}, nil)
	c = parse(t, "-c", a, "-c", d)
	if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "--threads") {
		t.Errorf("different threads: error = %v", err)
	}
	c = parse(t, "-c", a, "-c", d, "--threads", "2")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Errorf("different threads with --threads: %v", err)
	}

	// two run files with the same name would write the same output
	other := filepath.Join(dir, "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	same := writeExampleFarm(t, other, "a.yaml", input.Run{Seed: 1, Days: &days, Runs: 1, Threads: 3}, nil)
	c = parse(t, "-c", a, "-c", same, "-f", out)
	if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "same name") {
		t.Errorf("same name: error = %v", err)
	}
}
