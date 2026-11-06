package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	Seed       int64  `short:"s" default:"0" help:"Random seed (RAND_SEED). 0 means a random seed [INIT-01]."`
	Days       int    `default:"${default_days}" help:"Number of days (ticks) to run. 0 means no limit: the run only ends with StopExtinct? when the population is extinct [GO-20]."`
	Output     bool   `short:"o" help:"Write the daily output (the totals of OutputDailyProc and further values, see obs.DailyOutputExtended) to ${default_output_file}, one row per tick, tab separated."`
	OutputFile string `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o)."`
}

// parameters returns the model parameters for the flags, with all input files read. Without a run file, these are
// the example farm parameters.
func (c *cli) parameters() (params.Params, error) {
	// [INIT-02] NetLogo chooses MyParametersFile from the interface settings. Here the run file is given directly.
	// [INIT-04], [INIT-05]
	p := input.ExampleFarm()
	if c.Config != "" {
		var err error
		if p, err = input.Load(c.Config); err != nil {
			return p, err
		}
	}
	p.Seed = c.Seed
	p.Termination.MaxTicks = c.Days
	return p, nil
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

	p, err := c.parameters()
	if err != nil {
		log.Fatal(err)
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
		a.AddSystem(&reporter.CSV{Observer: &obs.DailyOutputExtended{}, File: outputFile, Sep: "\t"})
	}

	a.Run()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	derived := ecs.GetResource[params.Derived](a.World)
	fmt.Printf("species: %v\n", derived.InitialQueens)
	fmt.Printf("colonies founded: %v, bees died: %v\n", counters.TotalColoniesEverProduced, counters.TotalBeesEverDied)
	fmt.Printf("finished after %d days\n", ticks.Ticks)
	if outputFile != "" {
		fmt.Printf("daily output written to %s\n", outputFile)
	}
}
