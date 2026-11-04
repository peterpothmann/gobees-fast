// Package compare checks a step of gobees-fast: it runs the scenarios (scenario.go) with the current code, the
// reference commit and the base commit (gobees) and compares the current code with each of them and with NetLogo.
//
// All versions run as command line programs (go/main.go) with the same input files, parameter file, seeds and days,
// and write their daily output with -f. The tables are compared on the columns both versions have, matched by name:
//
//   - Exact: every value of every common column must be equal (phase 1).
//   - Statistical: distributions over many seeds (phases 2 and 3, and always against gobees), see stats.go.
//   - NetLogo: the daily median of each metric within the band of the NetLogo runs, see netlogo.go.
//   - Speed: run times of the current code and of the base commit, reported only.
//
// The programs are built by run_compare.sh, the tests read their paths from the environment (reference_test.go)
// and write the pull request comment and the NetLogo report (report.go). See README.md.
package compare

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"gobees/internal/examplefiles"
)

// Table is the daily output of one run: one row per day, the columns of Header.
type Table struct {
	Header []string
	Rows   [][]float64
}

// RoundedColumns are rounded to whole numbers when a table is read (ReadTable), in the output of every version, so
// every comparison sees the rounded values.
//
// foraging_period_s: NetLogo and the versions before step F2 compute the foraging period as hours * 3600, which
// misses whole seconds in the last bit on 15 of the 365 days of the example weather year (e.g. 2.2 * 3600 =
// 7920.000000000001). From step F2 on, gobees-fast reads whole seconds from the weather file (Weather =
// "foragingHoursExample", scenario step8_weather). The model gives the same results, only this column differs.
var RoundedColumns = []string{"foraging_period_s"}

// ReadTable reads a tab-separated table with a header line, as written by reporter.CSV. The values of
// RoundedColumns are rounded to whole numbers.
func ReadTable(r io.Reader) (Table, error) {
	t := Table{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if t.Header == nil {
			t.Header = fields
			continue
		}
		if len(fields) != len(t.Header) {
			return t, fmt.Errorf("row %d has %d values for %d columns", len(t.Rows)+1, len(fields), len(t.Header))
		}
		row := make([]float64, len(fields))
		for i, f := range fields {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return t, fmt.Errorf("row %d, column %s: %w", len(t.Rows)+1, t.Header[i], err)
			}
			if slices.Contains(RoundedColumns, t.Header[i]) {
				v = math.Round(v)
			}
			row[i] = v
		}
		t.Rows = append(t.Rows, row)
	}
	if t.Header == nil {
		return t, fmt.Errorf("empty table")
	}
	return t, scanner.Err()
}

// Column returns the index of the column name, or -1.
func (t Table) Column(name string) int {
	return slices.Index(t.Header, name)
}

// Columns returns the columns of the reference that the current version also has, the columns only the reference
// has (removed) and the columns only the current version has (added), each in the order of their table.
func Columns(reference, current Table) (common, removed, added []string) {
	for _, c := range reference.Header {
		if current.Column(c) >= 0 {
			common = append(common, c)
		} else {
			removed = append(removed, c)
		}
	}
	for _, c := range current.Header {
		if reference.Column(c) < 0 {
			added = append(added, c)
		}
	}
	return common, removed, added
}

// Exact compares two runs value by value on the given columns. It returns nil if they are equal, otherwise an
// error that names the first day and column that differ.
func Exact(reference, current Table, columns []string) error {
	if len(reference.Rows) != len(current.Rows) {
		return fmt.Errorf("%d days in the reference, %d in the current version", len(reference.Rows),
			len(current.Rows))
	}
	ticks := reference.Column("ticks")
	for i := range reference.Rows {
		for _, c := range columns {
			r, a := reference.Rows[i][reference.Column(c)], current.Rows[i][current.Column(c)]
			if r != a {
				day := float64(i + 1)
				if ticks >= 0 {
					day = reference.Rows[i][ticks]
				}
				return fmt.Errorf("day %v, column %s: reference %v, current %v", day, c, r, a)
			}
		}
	}
	return nil
}

// Run is one run of a program.
type Run struct {
	Seed     int64
	Table    Table
	Duration time.Duration // wall time of the program, incl. reading the input files
}

// RunProgram runs the gobees program bin once with the input files in dir and the parameter file parameters (a
// path, empty for the example farm file in dir). With output, it returns the daily output in Run.Table, otherwise
// the program writes none (for timing).
func RunProgram(bin, dir, parameters string, seed int64, days int, output bool) (Run, error) {
	args := []string{"-d", dir, "-s", strconv.FormatInt(seed, 10), "--days", strconv.Itoa(days)}
	if parameters != "" {
		args = append(args, "-p", parameters)
	}
	out := ""
	if output {
		f, err := os.CreateTemp("", "gobees-compare-*.csv")
		if err != nil {
			return Run{}, err
		}
		f.Close()
		defer os.Remove(f.Name())
		out = f.Name()
		args = append(args, "-f", out)
	}

	start := time.Now()
	if text, err := exec.Command(bin, args...).CombinedOutput(); err != nil {
		return Run{}, fmt.Errorf("%s, seed %d: %w\n%s", filepath.Base(bin), seed, err, text)
	}
	run := Run{Seed: seed, Duration: time.Since(start)}
	if !output {
		return run, nil
	}

	f, err := os.Open(out)
	if err != nil {
		return run, err
	}
	defer f.Close()
	if run.Table, err = ReadTable(f); err != nil {
		return run, fmt.Errorf("%s, seed %d: %w", filepath.Base(bin), seed, err)
	}
	return run, nil
}

// RunPrograms runs the program once per seed (RunProgram), with threads runs at the same time. The runs are in the
// order of the seeds.
func RunPrograms(bin, dir, parameters string, seeds []int64, days, threads int, output bool) ([]Run, error) {
	runs := make([]Run, len(seeds))
	errs := make([]error, len(seeds))
	next := make(chan int)
	var wg sync.WaitGroup
	for range threads {
		wg.Go(func() {
			for i := range next {
				runs[i], errs[i] = RunProgram(bin, dir, parameters, seeds[i], days, output)
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

// WriteInputFiles writes the input files of the example farm into dir, as the command line reads them with -d
// and the default flags, and the second map of the crop rotation scenario. All versions read the same files.
func WriteInputFiles(dir string) error {
	files := map[string]string{
		"_SYSTEM_Example_Farm_Parameters.csv":   examplefiles.ExampleFarmParameters,
		"_SYSTEM_Habitats.csv":                  examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt":  examplefiles.ExampleFarmFoodsources,
		"Example_Farm_PlotTest_Foodsources.txt": examplefiles.ExampleFarmPlotTestFoodsources,
		"_SYSTEM_Flowerspecies.csv":             examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":       examplefiles.BumbleSpecies,
		// Weather = "foragingHoursExample" of step8_weather (versions from step F2 on; the others ignore it)
		"_SYSTEM_Weather_foragingHoursExample.csv": examplefiles.WeatherForagingHoursExample,
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Seeds returns the seeds 1 to n.
func Seeds(n int) []int64 {
	seeds := make([]int64, n)
	for i := range seeds {
		seeds[i] = int64(i + 1)
	}
	return seeds
}
