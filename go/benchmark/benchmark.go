package benchmark

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"gobees/examplefiles"
)

// Run is the run time of one model run.
type Run struct {
	Number int     // run number, counted as BehaviorSpace does, from 1
	Seed   float64 // RAND_SEED
	// Setup and Days are NetLogo's Setup and its simulated days. Zero for the Go programs.
	Setup time.Duration
	Days  time.Duration
	// Program is the wall time of a Go program (RunPrograms), from the start of the process to its end. Zero for
	// NetLogo.
	Program time.Duration
	// ColoniesFounded is TotalColoniesEverProduced at the end of the run. It shows whether the runs of both models
	// did a comparable amount of work.
	ColoniesFounded float64
}

// Whole returns the time of the whole run: the wall time of a Go program, NetLogo's setup and days.
func (r Run) Whole() time.Duration {
	if r.Program > 0 {
		return r.Program
	}
	return r.Setup + r.Days
}

// parameterFile is the parameter file of the example farm, which NetLogo's Setup reads with the default interface
// settings [INIT-04].
const parameterFile = "_SYSTEM_Example_Farm_Parameters.csv"

// writeInputFiles writes the input files of the example farm into a new temporary directory and returns it.
func writeInputFiles() (string, error) {
	dir, err := os.MkdirTemp("", "gobees-benchmark")
	if err != nil {
		return "", err
	}
	files := map[string]string{
		parameterFile:                          examplefiles.ExampleFarmParameters,
		"_SYSTEM_Habitats.csv":                 examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt": examplefiles.ExampleFarmFoodsources,
		"_SYSTEM_Flowerspecies.csv":            examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":      examplefiles.BumbleSpecies,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

// runAll runs one run per seed, with threads runs at the same time. The runs get the run numbers 1, 2, ... in the
// order of the seeds.
func runAll(seeds []float64, threads int, run func(seed float64) (Run, error)) ([]Run, error) {
	runs := make([]Run, len(seeds))
	errs := make([]error, len(seeds))
	next := make(chan int)
	var wg sync.WaitGroup
	for range threads {
		wg.Go(func() {
			for i := range next {
				runs[i], errs[i] = run(seeds[i])
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

// RunPrograms runs the example farm with the command line program bin (go/main.go of any commit), once per seed
// for the given number of days, with threads runs at the same time, and times each program as a whole. The program
// writes no daily output. It cannot turn StopExtinct? off: a run that ends before the last day is an error, so that
// all programs simulate the same days.
//
// The same command line in gobees and gobees-fast makes the two comparable: both are timed the same way, also
// though the code of the base commit cannot time its setup and days separately. netLogoInput: the program reads the
// NetLogo input files of the example farm (-d, gobees). Otherwise it runs its built-in example farm (gobees-fast
// from step F7 on, which reads run files). Both include reading the input in the time.
//
// Before the timed runs, the program runs once with the first seed, untimed. Without it, the first runs of a newly
// built program took up to 1.7 times as long as the others (local run, Windows).
func RunPrograms(bin string, netLogoInput bool, seeds []float64, days, threads int) ([]Run, error) {
	dir, err := writeInputFiles()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	input := []string{}
	if netLogoInput {
		input = []string{"-d", dir}
	}
	if len(seeds) > 0 {
		if _, err := runProgram(bin, input, seeds[0], days); err != nil {
			return nil, err
		}
	}
	return runAll(seeds, threads, func(seed float64) (Run, error) { return runProgram(bin, input, seed, days) })
}

// Lines of the summary that go/main.go prints at the end of a run.
var (
	coloniesLine = regexp.MustCompile(`(?m)^colonies founded: (\S+),`)
	daysLine     = regexp.MustCompile(`(?m)^finished after (\d+) days`)
)

// runProgram runs the program once, with the input arguments input.
func runProgram(bin string, input []string, seed float64, days int) (Run, error) {
	args := append(slices.Clone(input), "-s", strconv.FormatInt(int64(seed), 10), "--days", strconv.Itoa(days))
	cmd := exec.Command(bin, args...)
	start := time.Now()
	out, err := cmd.CombinedOutput()
	wall := time.Since(start)
	if err != nil {
		return Run{}, fmt.Errorf("%s, seed %g: %w\n%s", filepath.Base(bin), seed, err, out)
	}
	run := Run{Seed: seed, Program: wall}
	return run, parseSummary(&run, string(out), days)
}

// parseSummary reads the colonies founded from the summary of a program and checks that it ran all days.
func parseSummary(run *Run, out string, days int) error {
	d, c := daysLine.FindStringSubmatch(out), coloniesLine.FindStringSubmatch(out)
	if d == nil || c == nil {
		return fmt.Errorf("seed %g: no summary in the output of the program:\n%s", run.Seed, out)
	}
	if d[1] != strconv.Itoa(days) {
		return fmt.Errorf("seed %g: the run ended after %s of %d days (StopExtinct?)", run.Seed, d[1], days)
	}
	colonies, err := strconv.ParseFloat(c[1], 64)
	if err != nil {
		return fmt.Errorf("seed %g: %w", run.Seed, err)
	}
	run.ColoniesFounded = colonies
	return nil
}
