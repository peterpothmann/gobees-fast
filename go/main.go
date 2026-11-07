package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gobees/globals"
	"gobees/input"
	"gobees/obs"
	"gobees/params"
	"gobees/schedule"

	"github.com/alecthomas/kong"
	"github.com/mlange-42/ark-tools/reporter"
	"github.com/mlange-42/ark/ecs"
)

// defaultOutputFile is the daily output file of -o, relative to the working directory (go/ with go run .).
const defaultOutputFile = "daily_output.csv"

// cli holds the command-line flags.
type cli struct {
	Config     string `short:"c" type:"path" placeholder:"PATH" help:"Run file (YAML, see package input). Its landscape files are relative to its folder. Default: the example farm, which is built in. tools/netlogo2gobees converts NetLogo input files into run files."`
	Seed       *int64 `short:"s" placeholder:"SEED" help:"Random seed. 0 means a random seed. Default: run.seed of the run file, else a random seed."`
	Days       *int   `placeholder:"DAYS" help:"Number of days (ticks) to run. 0 means no limit: the run only ends with stopExtinct when the population is extinct. Default: run.days of the run file, else ${default_days}."`
	Output     bool   `short:"o" help:"Write the daily output (the daily totals and further values, see obs.DailyOutputExtended) to ${default_output_file}, one row per tick, tab separated, and the run info next to it (${default_run_info}, a run file that repeats the run)."`
	OutputFile string `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o). The run info is PATH without extension plus _run_info.yaml."`
}

// parameters returns the model parameters for the flags and the files read for them. Without a run file, these are
// the example farm parameters. The flags -s and --days override the section run of the run file. The seed stays 0
// if neither gives one (a random seed, see randomSeed).
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
		p.Seed = *c.Seed
	}
	if c.Days != nil {
		p.Termination.MaxTicks = *c.Days
	}
	return p, f, nil
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

// run runs the model for the flags and writes a summary to w. With an output file, it writes the daily output and,
// before the run, the run info.
func (c *cli) run(w io.Writer) error {
	started := time.Now()
	p, files, err := c.parameters()
	if err != nil {
		return err
	}
	if p.Seed == 0 {
		// drawn here and not in params.Apply, so that the run info records it
		p.Seed = randomSeed()
	}

	a := schedule.Default(p, nil)
	// after all model systems: colonies founded and bees died for the summary below (the benchmark reads it)
	counters := obs.Counters{}
	a.AddSystem(&counters)

	outputFile := c.OutputFile
	if c.Output && outputFile == "" {
		outputFile = defaultOutputFile
	}
	if outputFile != "" {
		// after all model systems, so that each row holds the state at the end of the day.
		// reporter.CSV writes a row on every tick, creates missing folders and flushes the file in Finalize.
		// The comparisons with the reference commit (go/compare) read this output.
		observer := &obs.DailyOutputExtended{}
		a.AddSystem(&reporter.CSV{Observer: observer, File: outputFile, Sep: "\t"})

		// before the run, so that a run that fails can be repeated. reporter.CSV writes the column t in front.
		columns := append([]string{"t"}, observer.Header()...)
		info := newRunInfo(files, c.Config, &p, outputFile, columns, started)
		if err := info.write(runInfoPath(outputFile)); err != nil {
			return err
		}
	}

	a.Run()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	derived := ecs.GetResource[params.Derived](a.World)
	fmt.Fprintf(w, "species: %v\n", derived.InitialQueens)
	fmt.Fprintf(w, "seed: %d\n", p.Seed)
	fmt.Fprintf(w, "colonies founded: %v, bees died: %v\n", counters.TotalColoniesEverProduced, counters.TotalBeesEverDied)
	fmt.Fprintf(w, "finished after %d days\n", ticks.Ticks)
	if outputFile != "" {
		fmt.Fprintf(w, "daily output written to %s, run info to %s\n", outputFile, runInfoPath(outputFile))
	}
	return nil
}
