package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"gobees/input"
	"gobees/obs"
	"gobees/params"

	"github.com/alecthomas/kong"
)

// defaultOutputFile is the daily output file of -o, relative to the working directory (go/ with go run .).
const defaultOutputFile = "daily_output.csv"

// cli holds the command-line flags.
type cli struct {
	Config      string   `short:"c" type:"path" placeholder:"PATH" help:"Run file (YAML, see package input). Its landscape files are relative to its folder. Default: the example farm, which is built in. tools/netlogo2gobees converts NetLogo input files into run files."`
	Seed        *int64   `short:"s" placeholder:"SEED" help:"Random seed of the first run, 1 or more. Default: run.seed of the run file."`
	Days        *int     `placeholder:"DAYS" help:"Number of days (ticks) to run. 0 means no limit: the run only ends with stopExtinct when the population is extinct. Default: run.days of the run file, else ${default_days}."`
	Output      bool     `short:"o" help:"Write the daily output (the daily totals and further values, see obs.DailyOutputExtended) to ${default_output_file}, one row per tick and run, tab separated, with the columns run (from 1), seed and t in front, and the run info next to it (${default_run_info}, a run file that repeats the run)."`
	OutputFile  string   `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o). The run info is PATH without extension plus _run_info.yaml."`
	Columns     []string `placeholder:"A,B,C" help:"Columns of the daily output, comma separated, in this order (needs -o or -f). The column ticks is always first. Default: output.columns of the run file, else all columns. --list-columns shows the names."`
	ListColumns bool     `help:"Print the columns of the daily output with a description and exit."`
	Runs        *int     `placeholder:"N" help:"Number of runs in one process, 1 or more. The run i has the seed SEED + i - 1. The daily output of all runs is one file. Default: run.runs of the run file."`
	Threads     *int     `placeholder:"T" help:"Number of runs at the same time, 1 or more (on a cluster: the CPUs of the job). It does not change the output. Default: run.threads of the run file."`
}

// parameters returns the model parameters for the flags and the files read for them. Without a run file, these are
// the example farm parameters. The flags -s and --days override the section run of the run file.
func (c *cli) parameters() (params.Params, input.Files, error) {
	var f input.Files
	if c.Config == "" {
		f = input.ExampleFarmFiles()
	} else {
		var err error
		if f, err = input.Read(c.Config); err != nil {
			return params.Params{}, f, err
		}
	}
	p, err := input.Build(f.Config, f.Landscapes)
	if err != nil {
		return p, f, err
	}
	if c.Seed != nil {
		if *c.Seed < 1 {
			return p, f, fmt.Errorf("--seed is %d, want 1 or more", *c.Seed)
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
		return p, f, err
	}
	return p, f, nil
}

// outputFile returns the daily output file of the flags, "" without output.
func (c *cli) outputFile() string {
	if c.OutputFile == "" && c.Output {
		return defaultOutputFile
	}
	return c.OutputFile
}

// listColumns writes the columns of the daily output with their description to w.
func listColumns(w io.Writer) {
	columns := obs.AllColumns()
	width := 0
	for _, name := range columns {
		width = max(width, len(name))
	}
	for _, name := range columns {
		fmt.Fprintf(w, "%-*s  %s\n", width, name, obs.ColumnDoc(name))
	}
}

// newParser returns the parser of the command-line flags into c.
func newParser(c *cli) (*kong.Kong, error) {
	defaults := params.Default()
	return kong.New(c,
		kong.Name("gobees"),
		kong.Description("Go port of the NetLogo model BEE-STEWARD. Without flags, it runs the example farm."),
		kong.UsageOnError(),
		kong.Vars{
			"default_output_file": defaultOutputFile,
			"default_run_info":    runInfoPath(defaultOutputFile),
			"default_days":        strconv.Itoa(defaults.Termination.MaxTicks),
		},
	)
}

func main() {
	var c cli
	parser, err := newParser(&c)
	if err != nil {
		log.Fatal(err)
	}
	_, err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	if err := c.run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// runCount returns the number of runs: the flag --runs, else runs of the section run of the run file (checked by
// input.Build).
func (c *cli) runCount(f input.Files) (int, error) {
	if c.Runs == nil {
		return f.Config.Run.Runs, nil
	}
	if *c.Runs < 1 {
		return 0, fmt.Errorf("--runs is %d, want 1 or more", *c.Runs)
	}
	return *c.Runs, nil
}

// threadCount returns the number of runs at the same time: the flag --threads, else threads of the section run of the
// run file (checked by input.Build).
func (c *cli) threadCount(f input.Files) (int, error) {
	if c.Threads == nil {
		return f.Config.Run.Threads, nil
	}
	if *c.Threads < 1 {
		return 0, fmt.Errorf("--threads is %d, want 1 or more", *c.Threads)
	}
	return *c.Threads, nil
}

// run runs the model for the flags and writes a summary to w. With an output file, it writes the daily output and,
// before the run, the run info.
func (c *cli) run(w io.Writer) error {
	if c.ListColumns {
		listColumns(w)
		return nil
	}
	if len(c.Columns) > 0 && c.outputFile() == "" {
		return errors.New("--columns needs an output: add -o or -f")
	}
	started := time.Now()
	p, files, err := c.parameters()
	if err != nil {
		return err
	}
	runs, err := c.runCount(files)
	if err != nil {
		return err
	}
	threads, err := c.threadCount(files)
	if err != nil {
		return err
	}
	if p.Seed > math.MaxInt64-int64(runs)+1 {
		return fmt.Errorf("seed %d: the seed of the last of %d runs is larger than %d", p.Seed, runs, int64(math.MaxInt64))
	}

	rs := runSet{p: p, runs: runs, threads: threads}
	var output *outputWriter
	outputFile := c.outputFile()
	if outputFile != "" {
		// checked in parameters
		rs.columns, _ = obs.SelectColumns(p.Output.Columns)

		// before the runs, so that a run that fails can be repeated
		info := newRunInfo(files, c.Config, &p, runs, threads, outputFile, rs.columns, started)
		if err := info.write(runInfoPath(outputFile)); err != nil {
			return err
		}
		if output, err = newOutputWriter(outputFile, rs.columns); err != nil {
			return err
		}
	}

	var results []runResult
	err = rs.run(func(i int, r runResult) error {
		if output != nil {
			if err := output.write(r.rows); err != nil {
				return err
			}
		}
		r.rows = nil
		results = append(results, r)
		return nil
	})
	if output != nil {
		if cerr := output.close(); err == nil {
			err = cerr
		}
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "species: %v\n", results[0].species)
	if runs > 1 {
		fmt.Fprintf(w, "runs: %d, seeds: %d to %d, threads: %d\n", runs, p.Seed, p.Seed+int64(runs)-1,
			min(threads, runs))
		for i, r := range results {
			fmt.Fprintf(w, "run %d (seed %d): colonies founded: %v, bees died: %v, finished after %d days\n",
				i+1, p.Seed+int64(i), r.founded, r.died, r.days)
		}
	} else {
		r := results[0]
		fmt.Fprintf(w, "seed: %d\n", p.Seed)
		fmt.Fprintf(w, "colonies founded: %v, bees died: %v\n", r.founded, r.died)
		fmt.Fprintf(w, "finished after %d days\n", r.days)
	}
	if outputFile != "" {
		fmt.Fprintf(w, "daily output written to %s, run info to %s\n", outputFile, runInfoPath(outputFile))
	}
	return nil
}
