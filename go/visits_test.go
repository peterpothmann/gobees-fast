package main

import (
	"bytes"
	"cmp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gobees/input"
)

func TestVisitsOutput(t *testing.T) {
	// output.visits writes visits.csv next to the daily output: run, seed and t in front as in the daily output, one
	// row per day and foodsource with visits, the landscape named as its copy in the output folder. The file does not
	// depend on the threads, the provenance repeats it, and the daily output is the same as without visits.
	cfg, landscapes := input.ExampleFarmConfig()
	farm := landscapes[cfg.Landscape]
	cfg.Landscape = "a/farm.yaml"
	cfg.Output.Visits = true
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", cfg, map[string]input.Landscape{cfg.Landscape: farm})
	cfg.Output.Visits = false
	writeRunFile(t, dir, "plain.yaml", cfg, nil)

	const days = 60
	var files [][]byte
	for _, threads := range []string{"1", "2"} {
		out := filepath.Join(dir, "threads"+threads)
		c := parse(t, "-c", filepath.Join(dir, "run.yaml"), "-s", "3", "--days", strconv.Itoa(days),
			"--repetitions", "2", "--threads", threads, "-o", out)
		var b bytes.Buffer
		if err := c.run(&b); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(b.String(), "visits written to ") {
			t.Errorf("summary = %q, want the visits file", b.String())
		}
		files = append(files, readFile(t, filepath.Join(out, cfg.Name, visitsFile)))
	}
	if !bytes.Equal(files[0], files[1]) {
		t.Error("the visits output with 1 thread differs from the one with 2 threads")
	}

	lines := strings.Split(strings.TrimSuffix(string(files[0]), "\n"), "\n")
	if want := "run\tseed\tt\tticks\tlandscape\tid\tflowerSpecies\tnectarVisits\tpollenVisits"; lines[0] != want {
		t.Fatalf("header = %q, want %q", lines[0], want)
	}
	if len(lines) < 10 {
		t.Fatalf("%d lines, want visits on many days", len(lines))
	}
	runs := map[string]bool{}
	var last []string
	for _, l := range lines[1:] {
		f := strings.Split(l, "\t")
		if len(f) != 9 {
			t.Fatalf("row %q: %d fields, want 9", l, len(f))
		}
		runs[f[0]+" "+f[1]] = true
		tick, _ := strconv.Atoi(f[2])
		ticks, _ := strconv.Atoi(f[3])
		nectar, _ := strconv.Atoi(f[7])
		pollen, _ := strconv.Atoi(f[8])
		if ticks != tick+1 || ticks > days || f[4] != "farm.yaml" || nectar+pollen < 1 || f[6] == "" {
			t.Errorf("row %q: want ticks = t + 1 <= %d, landscape farm.yaml, a flower species, a visit", l, days)
		}
		// sorted by id and flower species within a day of a run
		if last != nil && last[0] == f[0] && last[2] == f[2] {
			idLast, _ := strconv.ParseFloat(last[5], 64)
			id, _ := strconv.ParseFloat(f[5], 64)
			if cmp.Compare(idLast, id) > 0 || (idLast == id && last[6] == f[6]) {
				t.Errorf("rows %q and %q: not sorted by id, or a foodsource twice", last, f)
			}
		}
		last = f
	}
	if len(runs) != 2 || !runs["1 3"] || !runs["2 4"] {
		t.Errorf("runs and seeds %v, want 1 3 and 2 4", runs)
	}

	// the provenance records the key and repeats the visits
	out := filepath.Join(dir, "threads1")
	daily, prov := outputOf(out, cfg.Name)
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Config.Output.Visits {
		t.Error("provenance: output.visits is false")
	}
	again := filepath.Join(dir, "again")
	repeat := parse(t, "-c", prov, "-o", again)
	if err := repeat.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(readFile(t, filepath.Join(again, cfg.Name, visitsFile)), files[0]) {
		t.Error("the runs repeated from the provenance have a different visits output")
	}

	// without output.visits: no visits file, the same daily output
	plain := filepath.Join(dir, "plain")
	c := parse(t, "-c", filepath.Join(dir, "plain.yaml"), "-s", "3", "--days", strconv.Itoa(days),
		"--repetitions", "2", "--threads", "1", "-o", plain)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	plainDaily, _ := outputOf(plain, cfg.Name)
	if !bytes.Equal(readFile(t, plainDaily), readFile(t, daily)) {
		t.Error("the daily output differs with and without visits")
	}
	if _, err := os.Stat(filepath.Join(plain, cfg.Name, visitsFile)); !os.IsNotExist(err) {
		t.Errorf("without output.visits: %s exists (%v)", visitsFile, err)
	}
}
