package compare

import (
	"fmt"
	"slices"
	"strings"
)

// Cell is the result of one comparison of a scenario, a cell of the table in the comment (Comment).
type Cell struct {
	Ran     bool // false: the comparison gave no result (e.g. a program failed), shown as failed
	Passed  bool
	Text    string // short result, e.g. "identical"
	Details string // Markdown, shown below the table if the comparison failed, may be empty
}

// ScenarioResult holds the comparisons of the current code in a scenario.
type ScenarioResult struct {
	Scenario Scenario
	Changed  Cell // against the reference commit: exact (phase 1) or statistical (phases 2 and 3)
	Drift    Cell // against the base commit (gobees), statistical
	NetLogo  Cell
	// ReferenceColumns and BaseColumns are the compared columns without the time columns, NotInReference and
	// NotInBase the columns of the current code that the other version does not write.
	ReferenceColumns, BaseColumns []string
	NotInReference, NotInBase     []string
	// Removed are the columns the current code removes on purpose (REMOVED_COLUMNS of commits.env), not compared.
	Removed       []string
	NetLogoResult NetLogoResult
}

// ExactCell compares the runs value by value (Exact): passed if every run is equal on the columns.
func ExactCell(reference, current []Run, columns []string) Cell {
	failures := []string{}
	for i := range reference {
		if err := Exact(reference[i].Table, current[i].Table, columns); err != nil {
			failures = append(failures, fmt.Sprintf("seed %d: %v", reference[i].Seed, err))
		}
	}
	if len(failures) == 0 {
		return Cell{Ran: true, Passed: true, Text: "identical"}
	}
	var b strings.Builder
	b.WriteString("First difference of each run that differs (at most 5):\n\n")
	for _, f := range failures[:min(5, len(failures))] {
		fmt.Fprintf(&b, "- %s\n", f)
	}
	return Cell{Ran: true, Text: fmt.Sprintf("%d of %d runs differ", len(failures), len(reference)),
		Details: b.String()}
}

// StatisticalCell compares the distributions over the runs on the columns: passed if the daily median of the
// current runs lies within the 5 to 95 % band of the other runs on every day and column (Band, relative tolerance
// 1e-9), and no yearly summary differs in a two-sample Kolmogorov-Smirnov test at FamilyAlpha divided by the number
// of tests (CompareSummaries, values rounded to 9 significant digits). Tolerance and rounding as against NetLogo:
// float sums whose last digits depend on the order are no difference of the model.
func StatisticalCell(other, current []Run, columns []string) Cell {
	columns = dataColumns(columns)
	checks, violations := Band(other, current, columns)
	tests := CompareSummaries(other, current, columns)
	alpha := FamilyAlpha / float64(max(1, len(tests)))
	significant := slices.DeleteFunc(slices.Clone(tests), func(s SummaryTest) bool { return s.P >= alpha })
	if len(violations) == 0 && len(significant) == 0 {
		return Cell{Ran: true, Passed: true, Text: "equal"}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Daily band: %d of %d days and columns outside.\n\n", len(violations), checks)
	for _, v := range violations[:min(5, len(violations))] {
		fmt.Fprintf(&b, "- day %d, %s: median %v, band [%v, %v]\n", v.Day, v.Column, v.Median, v.Q05, v.Q95)
	}
	fmt.Fprintf(&b, "\nYearly maximum and mean: %d of %d Kolmogorov-Smirnov tests with p < %.2g.\n\n",
		len(significant), len(tests), alpha)
	for _, s := range significant[:min(5, len(significant))] {
		fmt.Fprintf(&b, "- %s: D = %.3f, p = %.3g\n", s.Summary, s.D, s.P)
	}
	return Cell{Ran: true, Text: fmt.Sprintf("differs (%d days outside the band, %d Kolmogorov-Smirnov tests)",
		len(violations), len(significant)), Details: b.String()}
}

// NetLogoCell is the cell of the comparison with NetLogo: passed if every metric lies within the NetLogo band on
// every day.
func NetLogoCell(r NetLogoResult) Cell {
	inside := 0
	var b strings.Builder
	for _, m := range r.Metrics {
		if m.Inside == r.Days {
			inside++
			continue
		}
		for d, median := range m.Medians {
			q := m.Bands[d]
			tol := tolerance(q.Q50)
			if median < q.Q05-tol || median > q.Q95+tol {
				fmt.Fprintf(&b, "- %s: %d of %d days inside the band, first outside: day %d, median %v, NetLogo "+
					"[%v, %v]\n", m.Metric, m.Inside, r.Days, d+1, median, q.Q05, q.Q95)
				break
			}
		}
	}
	return Cell{Ran: true, Passed: r.Passed(), Text: fmt.Sprintf("%d of %d metrics in band", inside, len(r.Metrics)),
		Details: b.String()}
}

// MissingColumnsCell is the cell of a comparison that cannot run because the current code lacks columns of the
// other version: the column names stay the same in all steps (obs.DailyOutputExtended).
func MissingColumnsCell(missing []string) Cell {
	return Cell{Ran: true, Text: "columns missing: " + strings.Join(missing, ", ")}
}

// dataColumns returns the columns without the time columns (t, ticks) and the columns of the run (run, seed).
func dataColumns(columns []string) []string {
	return slices.DeleteFunc(slices.Clone(columns), func(c string) bool {
		return c == "t" || c == "ticks" || c == "run" || c == "seed"
	})
}
