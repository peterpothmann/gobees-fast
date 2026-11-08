package compare

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"gobees/examplefiles"
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

// ReadTable reads a tab-separated table with a header line, as the daily output of gobees and gobees-fast. The values of
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

// Input is the input format of a program.
type Input string

const (
	// NetLogoInput: the NetLogo input files, with -d and -p (gobees, and gobees-fast before step F7). The program
	// writes the daily output into the file of -f.
	NetLogoInput Input = "netlogo"
	// YAMLInput: a run file of the current format, with --config. The program writes its output into a folder of the
	// key name of the run file in the folder of -o.
	YAMLInput Input = "yaml"
)

// ParseInput returns the input format of s, NetLogoInput if s is empty.
func ParseInput(s string) (Input, error) {
	switch Input(s) {
	case "", NetLogoInput:
		return NetLogoInput, nil
	case YAMLInput:
		return YAMLInput, nil
	}
	return "", fmt.Errorf("unknown input format %q, expected %q or %q", s, NetLogoInput, YAMLInput)
}

// Args returns the command line arguments of the input files of scenario s (nil: the example farm) in dir, which
// holds the files of WriteInputFiles.
func (in Input) Args(dir string, s *Scenario) []string {
	if in == YAMLInput {
		if s == nil {
			return nil // the example farm is built in
		}
		return []string{"--config", filepath.Join(dir, scenariosDir, s.ConfigFile())}
	}
	args := []string{"-d", dir}
	if s != nil {
		args = append(args, "-p", filepath.Join(dir, s.ParameterFileName()))
	}
	return args
}

// outputArgs returns the flags that make a program of the input format write its daily output into the folder dir.
func (in Input) outputArgs(dir string) []string {
	if in == YAMLInput {
		return []string{"-o", dir}
	}
	return []string{"-f", filepath.Join(dir, "daily_output.csv")}
}

// dailyOutput returns the daily output file that a program of the input format wrote with outputArgs(dir). A run file
// writes daily.csv into the folder of its key name, the only folder in dir.
func (in Input) dailyOutput(dir string) (string, error) {
	if in != YAMLInput {
		return filepath.Join(dir, "daily_output.csv"), nil
	}
	files, err := filepath.Glob(filepath.Join(dir, "*", "daily.csv"))
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("%d daily outputs in %s, want one", len(files), dir)
	}
	return files[0], nil
}

// RunProgram runs the gobees program bin once with the input arguments args of the input format in (Input.Args).
// With output, it returns the daily output in Run.Table, otherwise the program writes none (for timing).
func RunProgram(bin string, in Input, args []string, seed int64, days int, output bool) (Run, error) {
	args = append(slices.Clone(args), "-s", strconv.FormatInt(seed, 10), "--days", strconv.Itoa(days))
	dir := ""
	if output {
		// a folder of its own: gobees-fast writes the provenance and the copies of the landscape files next to the
		// daily output
		var err error
		if dir, err = os.MkdirTemp("", "gobees-compare-*"); err != nil {
			return Run{}, err
		}
		defer os.RemoveAll(dir)
		args = append(args, in.outputArgs(dir)...)
	}

	start := time.Now()
	if text, err := exec.Command(bin, args...).CombinedOutput(); err != nil {
		return Run{}, fmt.Errorf("%s, seed %d: %w\n%s", filepath.Base(bin), seed, err, text)
	}
	run := Run{Seed: seed, Duration: time.Since(start)}
	if !output {
		return run, nil
	}

	out, err := in.dailyOutput(dir)
	if err != nil {
		return run, fmt.Errorf("%s, seed %d: %w", filepath.Base(bin), seed, err)
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
func RunPrograms(bin string, in Input, args []string, seeds []int64, days, threads int, output bool) ([]Run, error) {
	runs := make([]Run, len(seeds))
	errs := make([]error, len(seeds))
	next := make(chan int)
	var wg sync.WaitGroup
	for range threads {
		wg.Go(func() {
			for i := range next {
				runs[i], errs[i] = RunProgram(bin, in, args, seeds[i], days, output)
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

// WriteInputFiles writes into dir the input files of the example farm, as the command line reads them with -d and
// the default flags, the second map of the crop rotation scenario and the parameter files of the scenarios
// (ParameterFileName). All versions with NetLogo input read the same files. The run files of the scenarios
// (ScenarioFiles) go into the subfolder scenarios.
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
	for _, s := range Scenarios {
		text, err := s.ParameterFile()
		if err != nil {
			return err
		}
		files[s.ParameterFileName()] = text
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, scenariosDir), 0o755); err != nil {
		return err
	}
	return fs.WalkDir(ScenarioFiles, scenariosDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := ScenarioFiles.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, filepath.FromSlash(path)), b, 0o644)
	})
}

// Seeds returns the seeds 1 to n.
func Seeds(n int) []int64 {
	seeds := make([]int64, n)
	for i := range seeds {
		seeds[i] = int64(i + 1)
	}
	return seeds
}
