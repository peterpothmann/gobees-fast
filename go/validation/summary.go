package validation

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// DumpEnv is the environment variable with a folder. If it is set, every comparison test writes the results of its
// runs there (WriteDump): the CI collects them from gobees and gobees-fast and compares them (cmd/report).
// Keep this file the same as go/validation/summary.go of gobees: both write the same format.
const DumpEnv = "VALIDATION_DUMP_DIR"

// DaysPerYear is the length of a period of the run summaries.
const DaysPerYear = 365

// Values holds the metrics of all runs of an experiment: Values[metric][step][run], step 0 is the setup.
type Values map[string][][]float64

// Summary identifies a summary of one metric of a run: the value at step 0 (Period 0, Kind "value"), or the maximum
// or mean over the steps of a year (Period 1: steps 1 to 365, 2: 366 to 730, ...; Kind "max" or "mean"). A last
// year with fewer steps uses the steps it has.
type Summary struct {
	Metric string
	Period int
	Kind   string
}

func (s Summary) String() string {
	if s.Period == 0 {
		return s.Metric + ", setup"
	}
	return fmt.Sprintf("%s, year %d, %s", s.Metric, s.Period, s.Kind)
}

// RunSummaries returns the summaries of every run, Summaries[summary][run], computed in the same way as
// scripts/aggregate_netlogo.py does for NetLogo.
func RunSummaries(v Values) map[Summary][]float64 {
	result := map[Summary][]float64{}
	for metric, steps := range v {
		last := len(steps) - 1
		for run := range steps[0] {
			result[Summary{metric, 0, "value"}] = append(result[Summary{metric, 0, "value"}], steps[0][run])
			for year := 1; (year-1)*DaysPerYear+1 <= last; year++ {
				maxV, sum, n := 0.0, 0.0, 0
				for step := (year-1)*DaysPerYear + 1; step <= min(year*DaysPerYear, last); step++ {
					x := steps[step][run]
					if n == 0 || x > maxV {
						maxV = x
					}
					sum += x
					n++
				}
				result[Summary{metric, year, "max"}] = append(result[Summary{metric, year, "max"}], maxV)
				result[Summary{metric, year, "mean"}] = append(result[Summary{metric, year, "mean"}], sum/float64(n))
			}
		}
	}
	return result
}

// Quantile returns the q-quantile of the values with linear interpolation between the order statistics (type 7, as
// Python's statistics.quantiles with method "inclusive").
func Quantile(values []float64, q float64) float64 {
	v := slices.Clone(values)
	slices.Sort(v)
	h := q * float64(len(v)-1)
	lo := int(h)
	if lo+1 >= len(v) {
		return v[len(v)-1]
	}
	return v[lo] + (h-float64(lo))*(v[lo+1]-v[lo])
}

// WriteDump writes the results of an experiment into dir: <experiment>.csv with the quantiles per step and metric
// (the format of the NetLogo reference, Parse) and <experiment>_runs.csv with the run summaries (ParseRuns).
func WriteDump(dir, experiment string, v Values) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	metrics := make([]string, 0, len(v))
	for m := range v {
		metrics = append(metrics, m)
	}
	slices.Sort(metrics)

	var b strings.Builder
	b.WriteString("step,metric,n,q05,q25,q50,q75,q95\n")
	for step := range v[metrics[0]] {
		for _, m := range metrics {
			x := v[m][step]
			fmt.Fprintf(&b, "%d,%s,%d,%s,%s,%s,%s,%s\n", step, m, len(x), num(Quantile(x, 0.05)), num(Quantile(x, 0.25)),
				num(Quantile(x, 0.5)), num(Quantile(x, 0.75)), num(Quantile(x, 0.95)))
		}
	}
	if err := os.WriteFile(filepath.Join(dir, experiment+".csv"), []byte(b.String()), 0o644); err != nil {
		return err
	}

	summaries := RunSummaries(v)
	keys := make([]Summary, 0, len(summaries))
	for k := range summaries {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b Summary) int {
		if c := strings.Compare(a.Metric, b.Metric); c != 0 {
			return c
		}
		if a.Period != b.Period {
			return a.Period - b.Period
		}
		return strings.Compare(a.Kind, b.Kind)
	})
	b.Reset()
	b.WriteString("run,metric,period,kind,value\n")
	for _, k := range keys {
		for run, x := range summaries[k] {
			fmt.Fprintf(&b, "%d,%s,%d,%s,%s\n", run+1, k.Metric, k.Period, k.Kind, num(x))
		}
	}
	return os.WriteFile(filepath.Join(dir, experiment+"_runs.csv"), []byte(b.String()), 0o644)
}

// Dump writes the results of an experiment with WriteDump into the folder of DumpEnv. Does nothing if it is not set.
func Dump(experiment string, v Values) error {
	dir := os.Getenv(DumpEnv)
	if dir == "" {
		return nil
	}
	return WriteDump(dir, experiment, v)
}

// num formats a number so that parsing it gives the same float64.
func num(x float64) string {
	return strconv.FormatFloat(x, 'g', -1, 64)
}

// ParseRuns reads a table of run summaries (WriteDump, scripts/aggregate_netlogo.py): Summaries[summary][i], in the
// order of the rows.
func ParseRuns(table string) (map[Summary][]float64, error) {
	rows, err := csv.NewReader(strings.NewReader(table)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || len(rows[0]) != 5 || rows[0][0] != "run" {
		return nil, fmt.Errorf("expected the header run,metric,period,kind,value")
	}
	result := map[Summary][]float64{}
	for i, r := range rows[1:] {
		period, err := strconv.Atoi(r[2])
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		x, err := strconv.ParseFloat(r[4], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		k := Summary{r[1], period, r[3]}
		result[k] = append(result[k], x)
	}
	return result, nil
}
