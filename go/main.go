package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gobees/globals"
	"gobees/model"
	"gobees/obs"
	"gobees/params"

	"github.com/alecthomas/kong"
	"github.com/mlange-42/ark-tools/reporter"
	"github.com/mlange-42/ark/ecs"
)

// defaultOutputFile is the daily output file of -o, relative to the working directory (go/ with go run .).
const defaultOutputFile = "daily_output.csv"

// defaultDir is the folder of the .nlogo, relative to go/ (with go run .).
const defaultDir = "../netlogo"

// exampleFarmParameters is the parameter file NetLogo reads with the default interface settings (MyParametersFile,
// see [INIT-02]).
const exampleFarmParameters = "_SYSTEM_Example_Farm_Parameters.csv"

// cli holds the command-line flags.
type cli struct {
	Parameters string `short:"p" type:"path" placeholder:"PATH" help:"NetLogo parameter file (MyParametersFile). Default: ${example_farm_parameters} in --dir. A missing file is an error (NetLogo loads _SYSTEM_Parameters.csv instead [RPAR-01])."`
	Map        string `short:"m" default:"${default_map}" placeholder:"NAME" help:"Map (MyMap). Only the name is used: the foodsources are read from <NAME without .png>_Foodsources.txt in --dir [PAR-05]."`
	Dir        string `short:"d" type:"path" default:"${default_dir}" placeholder:"PATH" help:"Folder of the input files, the folder of the .nlogo in NetLogo. The map and the file names in the parameter file are relative to it."`
	Seed       int64  `short:"s" default:"0" help:"Random seed (RAND_SEED). 0 means a random seed [INIT-01]."`
	Days       int    `default:"${default_days}" help:"Number of days (ticks) to run. 0 means no limit: the run only ends with StopExtinct? when the population is extinct [GO-20]."`
	Output     bool   `short:"o" help:"Write the daily output (the totals of OutputDailyProc and further values, see obs.DailyOutputExtended) to ${default_output_file}, one row per tick, tab separated."`
	OutputFile string `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o)."`
}

// parameters returns the model parameters for the flags. Without flags, these are the example farm parameters.
func (c *cli) parameters() (params.Params, error) {
	p := params.Default()
	p.WorkingDirectory.Path = c.Dir
	p.Interface.MyMap = c.Map
	p.Interface.RAND_SEED = float64(c.Seed)
	p.Termination.MaxTicks = c.Days

	// [INIT-02] NetLogo chooses MyParametersFile from the interface settings. Here the file is given directly.
	path := c.Parameters
	if path == "" {
		path = filepath.Join(c.Dir, exampleFarmParameters)
	}
	// [INIT-04]
	file, err := params.ReadParameterFile(path)
	if err != nil {
		return p, err
	}
	p.File = file
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
			"default_output_file":     defaultOutputFile,
			"default_dir":             defaultDir,
			"default_map":             defaults.Interface.MyMap,
			"default_days":            strconv.Itoa(defaults.Termination.MaxTicks),
			"example_farm_parameters": exampleFarmParameters,
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

	a := model.Default(p, nil)

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
	derived := ecs.GetResource[globals.Derived](a.World)
	counters := ecs.GetResource[globals.Counters](a.World)
	fmt.Printf("species: %v\n", derived.BeeSpeciesInitialQueensList)
	fmt.Printf("colonies founded: %v, bees died: %v\n", counters.TotalColoniesEverProduced, counters.TotalBeesEverDied)
	fmt.Printf("finished after %d days\n", ticks.Ticks)
	if outputFile != "" {
		fmt.Printf("daily output written to %s\n", outputFile)
	}
}
