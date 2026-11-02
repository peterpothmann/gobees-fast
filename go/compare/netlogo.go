package compare

import (
	"fmt"
	"math"
	"slices"
	"strconv"

	"gobees/validation"
)

// NetLogoMetric is the comparison of one metric of a scenario with the NetLogo runs of its experiment.
type NetLogoMetric struct {
	Metric  string
	Inside  int                    // days on which the median of the current runs lies within the NetLogo band
	Medians []float64              // median of the current runs, days 1 to Days
	Bands   []validation.Quantiles // NetLogo quantiles, days 1 to Days
	Tests   []SummaryTest          // Kolmogorov-Smirnov tests of the yearly summaries, smallest p first
}

// NetLogoResult is the comparison of a scenario with the NetLogo runs of its experiment.
//
// Band (enforced): on every day, the median of the current runs lies within the 5 to 95 % band of the NetLogo runs,
// with a relative tolerance of 1e-9 for float sums whose last digits depend on the order. Kolmogorov-Smirnov
// (reported in the NetLogo report only): the yearly maximum and mean of each metric, Bonferroni-corrected per
// scenario (Alpha). Both from day 1 on: the programs write no row for the setup (day 0).
type NetLogoResult struct {
	Days    int
	Metrics []NetLogoMetric
	Alpha   float64
}

// Passed reports whether every metric lies within the NetLogo band on every day.
func (r NetLogoResult) Passed() bool {
	for _, m := range r.Metrics {
		if m.Inside < r.Days {
			return false
		}
	}
	return true
}

// CompareNetLogo compares the runs of the current code in a scenario with the NetLogo runs of its experiment, on all
// metrics of the NetLogo reference. Each must be a column of the runs.
func CompareNetLogo(s Scenario, runs []Run) (NetLogoResult, error) {
	quantilesText, runsText, err := validation.NetLogoReference(s.Name)
	if err != nil {
		return NetLogoResult{}, err
	}
	quantiles, err := validation.Parse(quantilesText)
	if err != nil {
		return NetLogoResult{}, err
	}
	var netlogoRuns map[validation.Summary][]float64
	if runsText != "" {
		if netlogoRuns, err = validation.ParseRuns(runsText); err != nil {
			return NetLogoResult{}, err
		}
	}

	// the metrics of the reference, in the order of the columns
	table := runs[0].Table
	metrics := []string{}
	for k := range quantiles {
		if k.Step == 1 {
			if table.Column(k.Metric) < 0 {
				return NetLogoResult{}, fmt.Errorf("%s: no column %s in the output", s.Name, k.Metric)
			}
			metrics = append(metrics, k.Metric)
		}
	}
	slices.SortFunc(metrics, func(a, b string) int { return table.Column(a) - table.Column(b) })

	result := NetLogoResult{Days: s.Days}
	tests := 0
	for _, metric := range metrics {
		m := NetLogoMetric{Metric: metric}
		for day := 1; day <= s.Days; day++ {
			q, ok := quantiles[validation.Key{Step: day, Metric: metric}]
			if !ok {
				return result, fmt.Errorf("%s: no NetLogo reference for %s on day %d", s.Name, metric, day)
			}
			values := dayValues(runs, metric, day-1)
			if len(values) == 0 {
				return result, fmt.Errorf("%s: the runs end before day %d", s.Name, day)
			}
			median := Quantile(values, 0.5)
			tol := 1e-9 * math.Max(1, math.Abs(q.Q50))
			if median >= q.Q05-tol && median <= q.Q95+tol {
				m.Inside++
			}
			m.Medians, m.Bands = append(m.Medians, median), append(m.Bands, q)
		}
		if netlogoRuns != nil {
			m.Tests = netLogoTests(runs, metric, netlogoRuns)
			tests += len(m.Tests)
		}
		result.Metrics = append(result.Metrics, m)
	}
	result.Alpha = FamilyAlpha / float64(max(1, tests))
	return result, nil
}

// netLogoTests compares the yearly summaries of a metric with those of the NetLogo runs. Values are rounded to 9
// significant digits first: float sums in a random order differ in the last digits, which is no difference of the
// model.
func netLogoTests(runs []Run, metric string, netlogo map[validation.Summary][]float64) []SummaryTest {
	current := map[Summary][]float64{}
	for _, r := range runs {
		for s, v := range Summaries(r.Table, []string{metric}) {
			current[s] = append(current[s], v)
		}
	}
	tests := []SummaryTest{}
	for s, a := range current {
		b := netlogo[validation.Summary{Metric: s.Column, Period: s.Year, Kind: s.Kind}]
		if len(a) < 10 || len(b) < 10 {
			continue
		}
		d, p := KolmogorovSmirnov(round9(b), round9(a))
		tests = append(tests, SummaryTest{s, d, p, len(b), len(a)})
	}
	sortTests(tests)
	return tests
}

func round9(values []float64) []float64 {
	r := make([]float64, len(values))
	for i, x := range values {
		r[i], _ = strconv.ParseFloat(strconv.FormatFloat(x, 'g', 9, 64), 64)
	}
	return r
}
