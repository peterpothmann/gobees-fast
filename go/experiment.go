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
// output. Without variations it has one combination, the run file itself.
type experiment struct {
	config       string // the run file of -c, "" for the built-in example farm
	name         string // the key name of the run file, the folder of its output
	files        input.Files
	combinations []input.Combination
	params       []params.Params // the parameters of each combination, with the flags applied
	repetitions  int             // runs per combination
	threads      int             // run.threads of the run file
	columns      []string        // the columns of the daily output
	dir          string          // the output folder, "" without output
	output       string          // the daily output file in dir, "" without
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
	// the name and the section run are checked by input.Build in paramsOf, and the variations cannot change them
	e.name = e.files.Config.Name
	run := e.files.Config.Run
	e.threads = run.Threads
	if e.repetitions, err = c.repetitionCount(run.Repetitions); err != nil {
		return nil, err
	}
	if seed := e.params[0].Seed; seed > math.MaxInt64-int64(e.repetitions)+1 {
		return nil, fmt.Errorf("seed %d: the seed of the last of %d repetitions is larger than %d", seed,
			e.repetitions, int64(math.MaxInt64))
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
// the same name, as for the file system of Windows and macOS.
func outputDirs(output string, experiments []*experiment) error {
	if output == "" {
		return nil
	}
	used := map[string]*experiment{}
	for _, e := range experiments {
		e.dir = filepath.Join(output, e.name)
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
// all combinations, so that the column run tells each run of the daily output apart.
func (e *experiment) jobs(index int, several bool) []job {
	var jobs []job
	for k, combination := range e.combinations {
		describe := ""
		if len(combination.Values) > 0 {
			describe = ", " + e.files.Describe(combination)
		}
		for r := range e.repetitions {
			p := e.params[k] // the input is shared by all runs, the model does not change it
			p.Seed += int64(r)
			number := k*e.repetitions + r + 1
			j := job{p: p, exp: index}
			j.label = fmt.Sprintf("run %d (seed %d%s)", number, p.Seed, describe)
			if several {
				j.label = e.name + ": " + j.label
			}
			if e.output != "" {
				j.columns = e.columns
				j.prefix = fmt.Sprintf("%d\t%d\t", number, p.Seed)
				for _, v := range combination.Values {
					j.prefix += v + "\t"
				}
			}
			jobs = append(jobs, j)
		}
	}
	return jobs
}

// header returns the columns of the daily output of the experiment: run (from 1 over all combinations), seed, the keys
// of the variations, t and the columns.
func (e *experiment) header() []string {
	h := []string{"run", "seed"}
	for _, v := range e.files.Config.Variations {
		h = append(h, v.Key)
	}
	h = append(h, "t")
	return append(h, e.columns...)
}
