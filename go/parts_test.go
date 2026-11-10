package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"
)

func TestParts(t *testing.T) {
	// 2 combinations x 3 repetitions = 6 runs in 4 parts of 1, 2, 1 and 2 runs. Each part writes into its own folder
	// with the run numbers and seeds of all runs. The parts appended in order give the daily output of all runs, byte
	// for byte. The provenance of a part repeats the part.
	dir := t.TempDir()
	days := 20
	experiment := writeExampleFarm(t, dir, "experiment.yaml", "badgers",
		input.Run{Seed: 4, Days: &days, Repetitions: 3, Threads: 2},
		[]input.Variation{{Key: "parameters.badgers", Values: values(t, "[0, 3]")}})
	whole := filepath.Join(dir, "whole")
	c := parse(t, "-c", experiment, "-o", whole, "--columns", "colonies")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	wholeDaily, _ := outputOf(whole, "badgers")
	all := readFile(t, wholeDaily)
	header, _, _ := strings.Cut(string(all), "\n")

	out := filepath.Join(dir, "out")
	var joined []byte
	wantRuns := [][]string{{"1"}, {"2", "3"}, {"4"}, {"5", "6"}}
	for i := 1; i <= 4; i++ {
		c := parse(t, "-c", experiment, "-o", out, "--columns", "colonies", "--part", fmt.Sprintf("%d/4", i))
		var b bytes.Buffer
		if err := c.run(&b); err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("part %d/4: runs", i); !strings.Contains(b.String(), want) {
			t.Errorf("summary %q, want %q", b.String(), want)
		}
		daily, prov := outputOf(filepath.Join(out, "badgers"), fmt.Sprintf("part_%d", i))
		part := readFile(t, daily)
		h, rows, _ := strings.Cut(string(part), "\n")
		if h != header {
			t.Errorf("part %d: header %q, want %q", i, h, header)
		}
		var runs []string
		for _, l := range strings.Split(strings.TrimSpace(rows), "\n") {
			if r, _, _ := strings.Cut(l, "\t"); !slices.Contains(runs, r) {
				runs = append(runs, r)
			}
		}
		if !slices.Equal(runs, wantRuns[i-1]) {
			t.Errorf("part %d: runs %q, want %q", i, runs, wantRuns[i-1])
		}
		joined = append(joined, rows...)

		info, err := input.Read(prov)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Config.Run.Part; got != fmt.Sprintf("%d/4", i) {
			t.Errorf("part %d: run.part of the provenance = %q", i, got)
		}
		if i == 3 {
			again := filepath.Join(dir, "again")
			c := parse(t, "-c", prov, "-o", again)
			if err := c.run(&bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			againDaily, _ := outputOf(filepath.Join(again, "badgers"), "part_3")
			if !bytes.Equal(readFile(t, againDaily), part) {
				t.Error("part 3 repeated from its provenance has a different daily output")
			}
		}
	}
	if want := all[len(header)+1:]; !bytes.Equal(joined, want) {
		t.Errorf("the parts appended differ from the daily output of all runs:\n%s\nwant\n%s", joined, want)
	}
}

func TestPartErrors(t *testing.T) {
	dir := t.TempDir()
	days := 2
	experiment := writeExampleFarm(t, dir, "experiment.yaml", "badgers",
		input.Run{Seed: 1, Days: &days, Repetitions: 3, Threads: 1},
		[]input.Variation{{Key: "parameters.badgers", Values: values(t, "[0, 3]")}})
	for _, tc := range []struct{ part, want string }{
		{"0/4", "want I/N"},
		{"5/4", "want I/N"},
		{"a/4", "want I/N"},
		{"3", "want I/N"},
		{"", "want I/N"},
		{"1/7", "has 6 runs (2 combinations x 3 repetitions), so at most 6 parts"},
	} {
		c := parse(t, "-c", experiment, "--part", tc.part)
		if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("--part %q: error = %v, want %q", tc.part, err, tc.want)
		}
	}

	// run.part of a run file is checked as the flag
	wrong := writeExampleFarm(t, dir, "wrong.yaml", "wrong",
		input.Run{Seed: 1, Days: &days, Repetitions: 1, Threads: 1, Part: "2/1"}, nil)
	c := parse(t, "-c", wrong)
	if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "run.part") {
		t.Errorf("run.part 2/1: error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "badgers")); err == nil {
		t.Error("output written")
	}
}

func TestPartFolder(t *testing.T) {
	// as many digits as n, so that the folders sort in the order of the parts
	for _, tc := range []struct {
		part runPart
		want string
	}{{runPart{3, 4}, "part_3"}, {runPart{3, 50}, "part_03"}, {runPart{50, 50}, "part_50"}, {runPart{7, 100}, "part_007"}} {
		if got := tc.part.folder(); got != tc.want {
			t.Errorf("%s: folder %q, want %q", tc.part, got, tc.want)
		}
	}
}
