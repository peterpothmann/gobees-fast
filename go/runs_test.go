package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"
	"gobees/params"
)

// dropFields returns the lines of the table in file without the header and without the first n fields of each line.
func dropFields(t *testing.T, file []byte, n int) []string {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(string(file), "\n"), "\n")
	for i, l := range lines[1:] {
		lines[i+1] = strings.Join(strings.Split(l, "\t")[n:], "\t")
	}
	return lines[1:]
}

func TestRunsOutput(t *testing.T) {
	// Three runs in one file: the columns run (from 1) and seed in front, then the rows of the runs in the order of the
	// runs. Each run is the run of a single call with its seed, and the file does not depend on the number of threads.
	dir := t.TempDir()
	var files [][]byte
	for _, threads := range []string{"1", "3"} {
		out := filepath.Join(dir, "threads"+threads)
		c := parse(t, "-s", "5", "--days", "40", "--repetitions", "3", "--threads", threads, "-o", out,
			"--columns", "colonies,eggs")
		path, _ := outputOf(out, "example_farm")
		var b bytes.Buffer
		if err := c.run(&b); err != nil {
			t.Fatal(err)
		}
		files = append(files, readFile(t, path))
		if !strings.Contains(b.String(), "runs: 3, seeds: 5 to 7") ||
			!strings.Contains(b.String(), "run 3 (seed 7): colonies founded: ") {
			t.Errorf("summary = %q, want the runs and seeds", b.String())
		}
	}
	if !bytes.Equal(files[0], files[1]) {
		t.Error("the output with 1 thread differs from the output with 3 threads")
	}

	lines := strings.Split(strings.TrimSuffix(string(files[0]), "\n"), "\n")
	if lines[0] != "run\tseed\tt\tticks\tcolonies\teggs" {
		t.Errorf("header = %q", lines[0])
	}
	if len(lines) != 1+3*40 {
		t.Fatalf("%d lines, want a header and 3 runs of 40 days", len(lines))
	}
	for i, seed := range []string{"5", "6", "7"} {
		out := filepath.Join(dir, "single"+seed)
		c := parse(t, "-s", seed, "--days", "40", "-o", out, "--columns", "colonies,eggs")
		if err := c.run(&bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		single, _ := outputOf(out, "example_farm")
		want := dropFields(t, readFile(t, single), 2)
		got := dropFields(t, files[0], 2)[i*40 : (i+1)*40]
		if !slices.Equal(got, want) {
			t.Errorf("run %d: rows differ from the single run with seed %s", i+1, seed)
		}
		for j, l := range lines[1+i*40 : 1+(i+1)*40] {
			if !strings.HasPrefix(l, fmt.Sprintf("%d\t%s\t", i+1, seed)) {
				t.Errorf("run %d, row %d = %q, want run %d and seed %s in front", i+1, j, l, i+1, seed)
				break
			}
		}
	}
}

func TestRunsOneRun(t *testing.T) {
	// no --repetitions and run.repetitions 1: one run, also with the columns run and seed, and repetitions: 1 in the
	// provenance
	out := t.TempDir()
	c := parse(t, "-s", "2", "--days", "3", "-o", out, "--columns", "eggs")
	path, prov := outputOf(out, "example_farm")
	var b bytes.Buffer
	if err := c.run(&b); err != nil {
		t.Fatal(err)
	}
	if got := header(t, path); !slices.Equal(got, []string{"run", "seed", "t", "ticks", "eggs"}) {
		t.Errorf("header = %v, want run, seed, t, ticks, eggs", got)
	}
	lines := strings.Split(strings.TrimSpace(string(readFile(t, path))), "\n")
	if len(lines) != 4 || !strings.HasPrefix(lines[1], "1\t2\t0\t1\t") || !strings.HasPrefix(lines[3], "1\t2\t2\t3\t") {
		t.Errorf("rows = %q, want run 1, seed 2, t 0 to 2 and ticks 1 to 3", lines[1:])
	}
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	if r := info.Config.Run.Repetitions; r != 1 {
		t.Errorf("run.repetitions = %d, want 1", r)
	}
	// the summary of one run keeps its lines (the benchmark reads them)
	if !strings.Contains(b.String(), "seed: 2\n") || !strings.Contains(b.String(), "finished after 3 days\n") {
		t.Errorf("summary = %q", b.String())
	}
}

func TestRunsProvenanceRepeatsRuns(t *testing.T) {
	// the provenance records repetitions, threads and the seed of the first repetition. Repeated from it, the runs
	// write the same file. The flag --repetitions overrides run.repetitions.
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	c := parse(t, "-s", "11", "--days", "20", "--repetitions", "3", "--threads", "2", "-o", first)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	firstDaily, prov := outputOf(first, "example_farm")
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	run := info.Config.Run
	if run.Repetitions != 3 || run.Threads != 2 || run.Seed != 11 {
		t.Fatalf("run.repetitions, run.threads, run.seed = %d, %d, %d, want 3, 2, 11", run.Repetitions, run.Threads,
			run.Seed)
	}

	second := filepath.Join(dir, "second")
	c = parse(t, "-c", prov, "-o", second)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	secondDaily, _ := outputOf(second, "example_farm")
	if !bytes.Equal(readFile(t, firstDaily), readFile(t, secondDaily)) {
		t.Error("the runs repeated from the provenance have a different daily output")
	}

	third := filepath.Join(dir, "third")
	c = parse(t, "-c", prov, "-o", third, "--repetitions", "2")
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	thirdDaily, _ := outputOf(third, "example_farm")
	if got := len(strings.Split(strings.TrimSpace(string(readFile(t, thirdDaily))), "\n")); got != 1+2*20 {
		t.Errorf("%d lines with --repetitions 2, want a header and 2 runs of 20 days", got)
	}
}

func TestRunFlagsAtLeastOne(t *testing.T) {
	// seed, repetitions and threads have no special value 0: a flag below 1 is an error
	for _, args := range [][]string{{"-s", "0"}, {"--seed=-3"}, {"--repetitions=0"}, {"--repetitions=-1"},
		{"--threads=0"}, {"--threads=-2"}} {
		c := parse(t, append(args, "--days", "1")...)
		if err := c.run(&bytes.Buffer{}); err == nil {
			t.Errorf("%v: want an error", args)
		}
	}
	// the last seed must not overflow
	c := parse(t, "-s", "9223372036854775807", "--repetitions", "2", "--days", "1")
	if err := c.run(&bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "last of 2 repetitions") {
		t.Errorf("seed overflow: error = %v", err)
	}
}

func TestRepetitionsAndThreads(t *testing.T) {
	// repetitions and threads: the flag, else the section run of the run file
	cfg, landscapes := input.ExampleFarmConfig()
	days := 1
	cfg.Run = &input.Run{Seed: 1, Days: &days, Repetitions: 4, Threads: 3}
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	for _, tc := range []struct {
		args                 []string
		repetitions, threads int
	}{
		{nil, 4, 3},
		{[]string{"--threads", "2", "--repetitions", "5"}, 5, 2},
	} {
		c := parse(t, append([]string{"-c", filepath.Join(dir, "run.yaml")}, tc.args...)...)
		e, err := c.newExperiment(c.Config[0])
		if err != nil {
			t.Fatal(err)
		}
		if e.repetitions != tc.repetitions {
			t.Errorf("%v: repetitions = %d, want %d", tc.args, e.repetitions, tc.repetitions)
		}
		threads, err := c.threadCount([]*experiment{e})
		if err != nil || threads != tc.threads {
			t.Errorf("%v: threads = %d, %v, want %d", tc.args, threads, err, tc.threads)
		}
	}

	// the provenance records the threads used
	out := filepath.Join(dir, "out")
	c := parse(t, "-c", filepath.Join(dir, "run.yaml"), "-o", out)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	_, prov := outputOf(out, cfg.Name)
	info, err := input.Read(prov)
	if err != nil {
		t.Fatal(err)
	}
	if info.Config.Run.Threads != 3 {
		t.Errorf("run.threads of the provenance = %d, want 3", info.Config.Run.Threads)
	}
}

func TestRunSetFailure(t *testing.T) {
	// A run that fails (here: parameters without input, the model panics) ends the call with an error that names the
	// run. The other runs do not hang.
	p := params.Default()
	p.Seed = 1
	rs := runSet{threads: 2, jobs: seedJobs(p, 4)}
	err := rs.run(func(i int, r runResult) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "run 1 (seed 1)") {
		t.Errorf("error = %v, want the failed run 1", err)
	}
}

// seedJobs returns n jobs of p without output, the job i with the seed p.Seed + i.
func seedJobs(p params.Params, n int) []job {
	jobs := make([]job, n)
	for i := range jobs {
		jobs[i].p = p
		jobs[i].p.Seed += int64(i)
		jobs[i].label = fmt.Sprintf("run %d (seed %d)", i+1, jobs[i].p.Seed)
	}
	return jobs
}

func TestRunSetDoneError(t *testing.T) {
	// an error of done (e.g. a full disk) ends the call too
	c := parse(t)
	p, _, err := c.parameters()
	if err != nil {
		t.Fatal(err)
	}
	p.Termination.MaxTicks = 2
	p.Seed = 3
	rs := runSet{threads: 2, jobs: seedJobs(p, 5)}
	var seen []int
	err = rs.run(func(i int, r runResult) error {
		seen = append(seen, i)
		if i == 1 {
			return errSkipped
		}
		return nil
	})
	if err != errSkipped || !slices.Equal(seen, []int{0, 1}) {
		t.Errorf("err, seen = %v, %v, want the error of run 2 and the runs 0 and 1", err, seen)
	}
}
