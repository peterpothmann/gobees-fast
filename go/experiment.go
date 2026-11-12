package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gobees/input"
	"gobees/obs"
	"gobees/params"
)

// experiment is one run file of a call (flag -c) with its variations: its combinations and its output. Without
// variations it has one combination, the run file itself. It keeps only the values of each combination: the runs
// build the parameters of their combination when they start (combinationInput).
type experiment struct {
	config      string // the run file of -c, "" for the built-in example farm
	name        string // the key name of the run file, the folder of its output
	files       input.Files
	values      [][]string        // the values of the variations in each combination (input.Combination.Values)
	first       params.Params     // the parameters of the first combination, with the flags applied
	sums        map[string]string // the SHA-256 of each landscape file at the start, by its name in the run file
	repetitions int               // runs per combination
	threads     int               // run.threads of the run file
	columns     []string          // the columns of the daily output
	dir         string            // the output folder, "" without output
	output      string            // the daily output file in dir, "" without
	results     []runResult       // the results of its runs, in the order of its jobs

	part *runPart // the part of the runs that runs (--part, run.part), nil: all runs
	cli  *cli     // the flags, which paramsOf applies to each combination
}

// runPart is the part i of n of the runs of an experiment (--part i/n).
type runPart struct{ i, n int }

// String returns the part as in --part, e.g. 3/50.
func (p runPart) String() string { return fmt.Sprintf("%d/%d", p.i, p.n) }

// folder returns the output folder of the part in the folder of the run file, e.g. part_03 of 50: the number with
// as many digits as n, so that the folders sort in the order of the parts.
func (p runPart) folder() string { return fmt.Sprintf("part_%0*d", len(strconv.Itoa(p.n)), p.i) }

// newExperiment reads the run file config ("" for the example farm) and checks each of its combinations with the
// flags applied, so that an error in any combination stops the call before the first run. It keeps the parameters of
// the first combination only.
func (c *cli) newExperiment(config string) (*experiment, error) {
	e := &experiment{config: config, cli: c, sums: map[string]string{}}
	var err error
	if config == "" {
		e.files = input.ExampleFarmFiles()
	} else if e.files, err = input.Read(config); err != nil {
		return nil, err
	}
	for k := range e.files.CombinationCount() {
		combination, err := e.files.Combination(k)
		if err != nil {
			return nil, err
		}
		p, err := c.paramsOf(combination.Files)
		if err != nil {
			return nil, err
		}
		if k == 0 {
			e.first = p
		}
		e.values = append(e.values, combination.Values)
		for name, raw := range combination.Files.Raw {
			sum := sha256Hex(raw)
			if old, ok := e.sums[name]; ok && old != sum {
				return nil, fmt.Errorf("landscape file %s changed while gobees-fast read the combinations", name)
			}
			e.sums[name] = sum
		}
	}
	// the name and the section run are checked by input.Build in paramsOf, and the variations cannot change them
	e.name = e.files.Config.Name
	run := e.files.Config.Run
	e.threads = run.Threads
	if e.repetitions, err = c.repetitionCount(run.Repetitions); err != nil {
		return nil, err
	}
	if err := e.setPart(c.Part, run.Part); err != nil {
		return nil, err
	}
	if seed := e.first.Seed; seed > math.MaxInt64-int64(e.repetitions)+1 {
		return nil, fmt.Errorf("seed %d: the seed of the last of %d repetitions is larger than %d", seed,
			e.repetitions, int64(math.MaxInt64))
	}
	e.columns, _ = obs.SelectColumns(e.first.Output.Columns) // checked in paramsOf
	return e, nil
}

// setPart sets the part of the runs: the flag --part, else run.part of the run file (checked by input.Build). A part
// needs at least one run, so n can be at most the number of runs.
func (e *experiment) setPart(flag *string, fromFile string) error {
	text := fromFile
	if flag != nil {
		text = *flag // also "", which is an error of ParsePart
	} else if text == "" {
		return nil
	}
	i, n, err := input.ParsePart(text)
	if err != nil {
		return fmt.Errorf("--part %w", err)
	}
	if runs := e.combinations() * e.repetitions; n > runs {
		return fmt.Errorf("--part %d/%d: %s has %d runs (%d combinations x %d repetitions), so at most %d parts",
			i, n, e.source(), runs, e.combinations(), e.repetitions, runs)
	}
	e.part = &runPart{i: i, n: n}
	return nil
}

// runRange returns the runs of the part, from lo to hi - 1 (numbered from 0 over all combinations): the block i of n
// blocks of consecutive runs, whose sizes differ by 1 at most. Without a part, all runs.
func (e *experiment) runRange() (lo, hi int) {
	runs := e.combinations() * e.repetitions
	if e.part == nil {
		return 0, runs
	}
	return (e.part.i - 1) * runs / e.part.n, e.part.i * runs / e.part.n
}

// partText returns the part for the provenance, "" without a part.
func (e *experiment) partText() string {
	if e.part == nil {
		return ""
	}
	return e.part.String()
}

// combinations returns the number of combinations of the experiment.
func (e *experiment) combinations() int { return len(e.values) }

// params returns the parameters of the combination k with the flags applied. newExperiment has checked them. A
// landscape file that changed since the start is an error: the run would not be the one of the provenance.
func (e *experiment) params(k int) (params.Params, error) {
	if k == 0 {
		return e.first, nil
	}
	combination, err := e.files.Combination(k)
	if err != nil {
		return params.Params{}, err
	}
	for name, raw := range combination.Files.Raw {
		if sha256Hex(raw) != e.sums[name] {
			return params.Params{}, fmt.Errorf("landscape file %s changed after the start: the provenance holds its "+
				"SHA-256 at the start", name)
		}
	}
	return e.cli.paramsOf(combination.Files)
}

// combinationInput holds the parameters of one combination for its runs (its repetitions). The first run that starts
// builds them, the last run that ends drops them. So an experiment holds the parameters of the combinations of the
// runs at the same time, not of all its combinations. Runs of several threads share it.
type combinationInput struct {
	build func() (params.Params, error)

	mu   sync.Mutex
	p    *params.Params // nil before the first run and after the last
	err  error          // the error of build, returned to every run of the combination
	left int            // runs that have not ended
}

// newCombinationInput returns the input of a combination of runs runs that build builds.
func newCombinationInput(runs int, build func() (params.Params, error)) *combinationInput {
	return &combinationInput{build: build, left: runs}
}

// get returns the parameters, built by the first call. The runs share them, the model does not change them.
func (ci *combinationInput) get() (params.Params, error) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	if ci.p == nil && ci.err == nil {
		p, err := ci.build()
		if err != nil {
			ci.err = err
		} else {
			ci.p = &p
		}
	}
	if ci.err != nil {
		return params.Params{}, ci.err
	}
	return *ci.p, nil
}

// done is called once by every run of the combination when it ends. The last one drops the parameters.
func (ci *combinationInput) done() {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.left--
	if ci.left == 0 {
		ci.p = nil
	}
}

// paramsOf returns the parameters of the run file f with the flags -s, --days and --columns applied.
func (c *cli) paramsOf(f input.Files) (params.Params, error) {
	p, err := input.Build(f.Config, f.Landscapes)
	if err != nil {
		return p, err
	}
	if c.Seed != nil {
		if *c.Seed < 1 {
			return p, fmt.Errorf("--seed is %d, want 1 or more", *c.Seed)
		}
		p.Seed = *c.Seed
	}
	if c.Days != nil {
		p.Termination.MaxTicks = *c.Days
	}
	if len(c.Columns) > 0 {
		p.Output.Columns = c.Columns
	}
	// also without output, so that a wrong name in a run file is found at the start
	if _, err := obs.SelectColumns(p.Output.Columns); err != nil {
		return p, err
	}
	return p, nil
}

// repetitionCount returns the number of runs per combination: the flag --repetitions, else repetitions of the run
// file.
func (c *cli) repetitionCount(repetitions int) (int, error) {
	if c.Repetitions == nil {
		return repetitions, nil
	}
	if *c.Repetitions < 1 {
		return 0, fmt.Errorf("--repetitions is %d, want 1 or more", *c.Repetitions)
	}
	return *c.Repetitions, nil
}

// threadCount returns the number of runs at the same time: the flag --threads, else threads of the run files, which
// must then be the same in all of them.
func (c *cli) threadCount(experiments []*experiment) (int, error) {
	if c.Threads != nil {
		if *c.Threads < 1 {
			return 0, fmt.Errorf("--threads is %d, want 1 or more", *c.Threads)
		}
		return *c.Threads, nil
	}
	threads := experiments[0].threads
	for _, e := range experiments[1:] {
		if e.threads != threads {
			return 0, fmt.Errorf("the run files ask for different threads (%s: %d, %s: %d): set --threads",
				experiments[0].source(), threads, e.source(), e.threads)
		}
	}
	return threads, nil
}

// outputDirs sets the output folder of each experiment: the folder of its name (the key name of its run file) in the
// folder output, one or several run files alike (-o out, -c farm.yaml -c sussex.yaml: out/farm, out/sussex). Two run
// files with the same name are an error, their outputs would overwrite each other. Names that differ only in case are
// the same name, as for the file system of Windows and macOS. With a part, the output goes into the folder of the
// part in it (out/farm/part_03), so that all jobs of an array can write into the same folder out.
func outputDirs(output string, experiments []*experiment) error {
	if output == "" {
		return nil
	}
	used := map[string]*experiment{}
	for _, e := range experiments {
		e.dir = filepath.Join(output, e.name)
		if e.part != nil {
			e.dir = filepath.Join(e.dir, e.part.folder())
		}
		e.output = filepath.Join(e.dir, dailyOutputFile)
		if other, ok := used[strings.ToLower(e.name)]; ok {
			return fmt.Errorf("the run files %s and %s have the same name %q, their outputs would both go to %s: "+
				"change the key name of one of them", other.source(), e.source(), e.name, e.dir)
		}
		used[strings.ToLower(e.name)] = e
	}
	return nil
}

// source returns the run file of the experiment for messages: its path, or the built-in example farm.
func (e *experiment) source() string {
	if e.config == "" {
		return exampleFarmConfig
	}
	return e.config
}

// jobs returns the runs of the experiment number index: each combination run e.repetitions times, the repetition r
// (from 0) with the seed of the run file + r. All combinations get the same seeds. The runs are numbered from 1 over
// all combinations, so that the column run tells each run of the daily output apart. With a part, only its runs, with
// the same numbers and seeds. The runs of a combination share its input (combinationInput).
func (e *experiment) jobs(index int, several bool) []job {
	var jobs []job
	lo, hi := e.runRange()
	for k, values := range e.values {
		// the repetitions of the combination k in the part
		first, last := max(lo, k*e.repetitions), min(hi, (k+1)*e.repetitions)
		if first >= last {
			continue
		}
		describe := ""
		if len(values) > 0 {
			describe = ", " + e.files.Describe(input.Combination{Values: values})
		}
		in := newCombinationInput(last-first, func() (params.Params, error) { return e.params(k) })
		for run := first; run < last; run++ {
			r := run - k*e.repetitions
			seed := e.first.Seed + int64(r) // the variations cannot change the seed
			number := run + 1
			j := job{input: in, seed: seed, exp: index}
			j.label = fmt.Sprintf("run %d (seed %d%s)", number, seed, describe)
			if several {
				j.label = e.name + ": " + j.label
			}
			if e.output != "" {
				j.columns = e.columns
				j.prefix = fmt.Sprintf("%d\t%d\t", number, seed)
				for _, v := range values {
					j.prefix += v + "\t"
				}
			}
			jobs = append(jobs, j)
		}
	}
	return jobs
}

// header returns the columns of the daily output of the experiment: run (from 1 over all combinations), seed, the keys
// of the variations (site for the sites), t and the columns.
func (e *experiment) header() []string {
	h := []string{"run", "seed"}
	for _, v := range e.files.Config.Variations {
		h = append(h, v.Column())
	}
	h = append(h, "t")
	return append(h, e.columns...)
}
