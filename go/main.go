package main

import (
	"fmt"
	"log"
	"path/filepath"

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

// cli holds the command-line flags.
type cli struct {
	Output     bool   `short:"o" help:"Write the daily output of OutputDailyProc to ${default_output_file}, one row per tick, tab separated."`
	OutputFile string `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o)."`
}

func main() {
	var c cli
	kong.Parse(&c,
		kong.Name("gobees"),
		kong.Description("Go port of the NetLogo model BEE-STEWARD. Runs the example farm."),
		kong.UsageOnError(),
		kong.Vars{"default_output_file": defaultOutputFile},
	)

	p := params.Default()
	p.WorkingDirectory.Path = filepath.Join("..", "netlogo")

	// [INIT-04] With the default interface settings, NetLogo reads MyParametersFile (see [INIT-02]).
	file, err := params.ReadParameterFile(filepath.Join(p.WorkingDirectory.Path, "_SYSTEM_Example_Farm_Parameters.csv"))
	if err != nil {
		log.Fatal(err)
	}
	p.File = file

	a := model.Default(p, nil)

	outputFile := c.OutputFile
	if c.Output && outputFile == "" {
		outputFile = defaultOutputFile
	}
	if outputFile != "" {
		// after all model systems, so that each row holds the state at the end of the day.
		// reporter.CSV writes a row on every tick, creates missing folders and flushes the file in Finalize.
		a.AddSystem(&reporter.CSV{Observer: &obs.DailyOutput{}, File: outputFile, Sep: "\t"})
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
