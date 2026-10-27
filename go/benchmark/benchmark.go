// Package benchmark compares the run time of gobees with NetLogo on the example farm. NetLogo runs the experiment
// netlogo.xml (run_netlogo.sh), gobees runs the same seeds and days, and the report shows both. See README.md.
package benchmark

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"gobees/globals"
	"gobees/internal/examplefiles"
	"gobees/model"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// Run is the run time of one model run.
type Run struct {
	Number int     // run number, counted as BehaviorSpace does, from 1
	Seed   float64 // RAND_SEED
	Setup  time.Duration
	Days   time.Duration // all simulated days after the setup
	// ColoniesFounded is TotalColoniesEverProduced at the end of the run. It shows whether the runs of both models
	// did a comparable amount of work.
	ColoniesFounded float64
}

// parameterFile is the parameter file of the example farm, which NetLogo's Setup reads with the default interface
// settings [INIT-04].
const parameterFile = "_SYSTEM_Example_Farm_Parameters.csv"

// RunGobees runs the example farm once per seed for the given number of days, with threads runs at the same time.
// The runs get the run numbers 1, 2, ... in the order of the seeds. As in netlogo.xml, StopExtinct? is false, so
// that every run simulates all days.
//
// The input files are written to a temporary directory before the runs. The timed setup reads them from there,
// including the parameter file, as NetLogo's Setup does.
func RunGobees(seeds []float64, days, threads int) ([]Run, error) {
	dir, err := os.MkdirTemp("", "gobees-benchmark")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	files := map[string]string{
		parameterFile:                          examplefiles.ExampleFarmParameters,
		"_SYSTEM_Habitats.csv":                 examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt": examplefiles.ExampleFarmFoodsources,
		"_SYSTEM_Flowerspecies.csv":            examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":      examplefiles.BumbleSpecies,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			return nil, err
		}
	}

	runs := make([]Run, len(seeds))
	errs := make([]error, len(seeds))
	next := make(chan int)
	var wg sync.WaitGroup
	for range threads {
		wg.Go(func() {
			for i := range next {
				runs[i], errs[i] = runGobees(dir, seeds[i], days)
				runs[i].Number = i + 1
			}
		})
	}
	for i := range seeds {
		next <- i
	}
	close(next)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return runs, nil
}

// runGobees runs the example farm once, with the input files in dir.
func runGobees(dir string, seed float64, days int) (Run, error) {
	start := time.Now()
	file, err := params.ReadParameterFile(filepath.Join(dir, parameterFile))
	if err != nil {
		return Run{}, err
	}
	p := params.Default()
	p.WorkingDirectory.Path = dir
	p.File = file
	p.File.StopExtinct = false
	p.Interface.RAND_SEED = seed
	p.Termination.MaxTicks = days
	a := model.Default(p, nil)
	a.Initialize()
	setup := time.Since(start)

	start = time.Now()
	for range days {
		a.Update()
	}
	run := Run{Seed: seed, Setup: setup, Days: time.Since(start)}
	run.ColoniesFounded = ecs.GetResource[globals.Counters](a.World).TotalColoniesEverProduced
	return run, nil
}
