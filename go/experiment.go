package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"gobees/input"
	"gobees/obs"
	"gobees/params"
)

// experiment is one run file of a call (flag -c) with its variations: its combinations, their parameters and its
// daily output. Without variations it has one combination, the run file itself.
type experiment struct {
	config       string // the run file of -c, "" for the built-in example farm
	files        input.Files
	combinations []input.Combination
	params       []params.Params // the parameters of each combination, with the flags applied
	runs         int             // runs per combination
	threads      int             // run.threads of the run file
	columns      []string        // the columns of the daily output
	output       string          // the daily output file, "" without
	results      []runResult     // the results of its runs, in the order of its jobs
}

// newExperiment reads the run file config ("" for the example farm), its combinations and their parameters with the
// flags applied.
func (c *cli) newExperiment(config string) (*experiment, error) {
	e := &experiment{config: config}
	var err error
	if config == "" {
		e.files = input.ExampleFarmFiles()
	} else if e.files, err = input.Read(config); err != nil {
		return nil, err
	}
	if e.combinations, err = e.files.Combinations(); err != nil {
		return nil, err
	}
	for _, combination := range e.combinations {
		p, err := c.paramsOf(combination.Files)
		if err != nil {
			return nil, err
		}
		e.params = append(e.params, p)
	}
	// the section run is checked by input.Build in paramsOf, and the variations cannot change it
	run := e.files.Config.Run
	e.threads = run.Threads
	if e.runs, err = c.runCount(run.Runs); err != nil {
		return nil, err
	}
	if seed := e.params[0].Seed; seed > math.MaxInt64-int64(e.runs)+1 {
		return nil, fmt.Errorf("seed %d: the seed of the last of %d runs is larger than %d", seed, e.runs,
			int64(math.MaxInt64))
	}
	e.columns, _ = obs.SelectColumns(e.params[0].Output.Columns) // checked in paramsOf
	return e, nil
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

// runCount returns the number of runs per combination: the flag --runs, else runs of the run file.
func (c *cli) runCount(runs int) (int, error) {
	if c.Runs == nil {
		return runs, nil
	}
	if *c.Runs < 1 {
		return 0, fmt.Errorf("--runs is %d, want 1 or more", *c.Runs)
	}
	return *c.Runs, nil
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
				experiments[0].config, threads, e.config, e.threads)
		}
	}
	return threads, nil
}

// outputFiles sets the daily output file of each experiment: output for one experiment, else output with the name
// of the run file in front of the extension (out.csv, -c a.yaml -c b.yaml: out_a.csv, out_b.csv). Two run files with
// the same name are an error, their outputs would overwrite each other.
func outputFiles(output string, experiments []*experiment) error {
	if output == "" {
		return nil
	}
	if len(experiments) == 1 {
		experiments[0].output = output
		return nil
	}
	ext := filepath.Ext(output)
	stem := strings.TrimSuffix(output, ext)
	used := map[string]string{}
	for _, e := range experiments {
		name := strings.TrimSuffix(filepath.Base(e.config), filepath.Ext(e.config))
		e.output = stem + "_" + name + ext
		if other, ok := used[e.output]; ok {
			return fmt.Errorf("the run files %s and %s have the same name, their outputs would both be %s",
				other, e.config, e.output)
		}
		used[e.output] = e.config
	}
	return nil
}

// jobs returns the runs of the experiment number index: each combination run e.runs times, the run r (from 0) with
// the seed of the run file + r. All combinations get the same seeds.
func (e *experiment) jobs(index int, several bool) []job {
	var jobs []job
	for k, combination := range e.combinations {
		describe := ""
		if len(combination.Values) > 0 {
			describe = ", " + e.files.Describe(combination)
		}
		for r := range e.runs {
			p := e.params[k] // the input is shared by all runs, the model does not change it
			p.Seed += int64(r)
			j := job{p: p, exp: index}
			j.label = fmt.Sprintf("run %d (seed %d%s)", r+1, p.Seed, describe)
			if several {
				j.label = e.config + ": " + j.label
			}
			if e.output != "" {
				j.columns = e.columns
				j.prefix = fmt.Sprintf("%d\t%d\t", r+1, p.Seed)
				for _, v := range combination.Values {
					j.prefix += v + "\t"
				}
			}
			jobs = append(jobs, j)
		}
	}
	return jobs
}

// header returns the columns of the daily output of the experiment: run, seed, the keys of the variations, t and the
// columns.
func (e *experiment) header() []string {
	h := []string{"run", "seed"}
	for _, v := range e.files.Config.Variations {
		h = append(h, v.Key)
	}
	h = append(h, "t")
	return append(h, e.columns...)
}
