package compare

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// CommentMarker starts the pull request comment, so that the workflow finds and updates it.
const CommentMarker = "<!-- check-report -->"

// CommentInfo is what the comment says besides the results.
type CommentInfo struct {
	Phase             int    // 1: exact comparison with the reference, 2 and 3: statistical
	Reference, Base   string // the commits, e.g. "40dc301"
	Seeds             int    // runs per version and scenario, seeds 1 to Seeds
	ReportURL, RunURL string // links to the NetLogo report and to the workflow run, may be empty
}

// Speed is the result of the speed comparison: the median wall time per run of the base commit and of the current
// code.
type Speed struct {
	Base, Current time.Duration
	Days          int
}

// Comment returns the pull request comment: one row per scenario of Scenarios and one column per comparison, the
// speed, the runs, and the details of every comparison that failed. A scenario without a result is shown as failed.
// speed is nil if the speed comparison gave no result.
func Comment(results []ScenarioResult, speed *Speed, info CommentInfo) string {
	byName := map[string]ScenarioResult{}
	for _, r := range results {
		byName[r.Scenario.Name] = r
	}
	ordered := make([]ScenarioResult, len(Scenarios))
	for i, s := range Scenarios {
		r, ok := byName[s.Name]
		if !ok {
			r = ScenarioResult{Scenario: s}
		}
		ordered[i] = r
	}
	changedKind := "exact"
	if info.Phase != 1 {
		changedKind = "statistical"
	}

	var b strings.Builder
	b.WriteString(CommentMarker + "\n## Check of this step\n\ngobees-fast in this pull request, compared with:\n\n")
	fmt.Fprintf(&b, "| Scenario | its previous step (`%s`), %s | gobees (`%s`), statistical | NetLogo |\n"+
		"| --- | --- | --- | --- |\n", info.Reference, changedKind, info.Base)
	for _, r := range ordered {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", r.Scenario.Title, cellText(r.Changed), cellText(r.Drift),
			cellText(r.NetLogo))
	}
	if speed != nil {
		fmt.Fprintf(&b, "\n**Speed:** %.2f× as fast as gobees `%s` (%v → %v per run of %d days on the example "+
			"farm, without output)\n", float64(speed.Base)/float64(speed.Current), info.Base,
			speed.Base.Round(time.Millisecond), speed.Current.Round(time.Millisecond), speed.Days)
	} else {
		b.WriteString("\n**Speed:** no result\n")
	}

	b.WriteString("\n<details><summary>Runs</summary>\n\n")
	fmt.Fprintf(&b, "Every version runs each scenario with seeds 1 to %d and StopExtinct? = false and is compared "+
		"from day 1 on. NetLogo: 100 runs of the experiment of the same name in `go/validation/experiments`.\n\n",
		info.Seeds)
	b.WriteString("- **Exact:** every value of every compared column is equal, in every run.\n" +
		"- **Statistical:** on every day, the median of gobees-fast lies within the 5 to 95 % band of the other " +
		"version, and no yearly maximum or mean of a column differs in a Kolmogorov-Smirnov test at 0.05 divided " +
		"by the number of tests.\n" +
		"- **NetLogo:** on every day, the median of each NetLogo metric lies within the 5 to 95 % band of the " +
		"NetLogo runs.\n\n")
	b.WriteString("| Scenario | Days | Columns compared with the previous step | with gobees | NetLogo metrics |\n" +
		"| --- | ---: | ---: | ---: | ---: |\n")
	for _, r := range ordered {
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s |\n", r.Scenario.Title, r.Scenario.Days,
			count(r.ReferenceColumns, r.Changed.Ran), count(r.BaseColumns, r.Drift.Ran),
			count(r.NetLogoResult.Metrics, r.NetLogo.Ran))
	}
	notCompared(&b, ordered)
	b.WriteString("\n</details>\n")

	for _, r := range ordered {
		details := ""
		for _, c := range []struct {
			name string
			cell Cell
		}{{"its previous step", r.Changed}, {"gobees", r.Drift}, {"NetLogo", r.NetLogo}} {
			if c.cell.Ran && !c.cell.Passed && c.cell.Details != "" {
				details += fmt.Sprintf("**Against %s:** %s\n\n%s\n", c.name, c.cell.Text, c.cell.Details)
			}
		}
		if details != "" {
			fmt.Fprintf(&b, "\n<details><summary>%s: differences</summary>\n\n%s\n</details>\n", r.Scenario.Title,
				details)
		}
	}

	if info.ReportURL != "" {
		fmt.Fprintf(&b, "\nNetLogo report with plots and Kolmogorov-Smirnov tests: %s\n", info.ReportURL)
	}
	if info.RunURL != "" {
		fmt.Fprintf(&b, "\nLogs: [run](%s)\n", info.RunURL)
	}
	return b.String()
}

func cellText(c Cell) string {
	switch {
	case !c.Ran:
		return "❌ no result"
	case c.Passed:
		return "✅ " + c.Text
	}
	return "❌ " + c.Text
}

// count returns the number of items as text, or "–" if the comparison did not run.
func count[T any](items []T, ran bool) string {
	if !ran {
		return "–"
	}
	return fmt.Sprint(len(items))
}

// notCompared lists the columns of the current code that the other versions do not write, once for both if they
// are the same. The columns are the same in all scenarios.
func notCompared(b *strings.Builder, results []ScenarioResult) {
	var reference, base, removed []string
	for _, r := range results {
		if removed == nil {
			removed = r.Removed
		}
		if r.Changed.Ran && reference == nil {
			reference = r.NotInReference
		}
		if r.Drift.Ran && base == nil {
			base = r.NotInBase
		}
	}
	line := func(who string, columns []string) {
		if len(columns) > 0 {
			fmt.Fprintf(b, "\nNot compared with %s: %d columns it does not write (%s).\n", who, len(columns),
				strings.Join(columns, ", "))
		}
	}
	if len(removed) > 0 {
		fmt.Fprintf(b, "\nRemoved on purpose with their model branch (`REMOVED_COLUMNS`): %s.\n",
			strings.Join(removed, ", "))
	}
	if slices.Equal(reference, base) {
		line("the previous step and gobees", reference)
		return
	}
	line("the previous step", reference)
	line("gobees", base)
}

// WriteReport writes the NetLogo report validation-report.md and its plots (plots/<scenario>.png) into dir. It
// removes the plots of an earlier report first, so that dir holds only the results of these runs.
func WriteReport(dir string, results []ScenarioResult) error {
	plotDir := filepath.Join(dir, "plots")
	if err := os.RemoveAll(plotDir); err != nil {
		return err
	}
	if err := os.MkdirAll(plotDir, 0o755); err != nil {
		return err
	}
	plots := map[string]bool{}
	for _, r := range results {
		if len(r.NetLogoResult.Metrics) == 0 {
			continue
		}
		ok, err := writePlot(filepath.Join(plotDir, r.Scenario.Name+".png"), r.NetLogoResult.Metrics)
		if err != nil {
			return fmt.Errorf("%s: %w", r.Scenario.Name, err)
		}
		plots[r.Scenario.Name] = ok
	}
	return os.WriteFile(filepath.Join(dir, "validation-report.md"), []byte(Report(results, plots)), 0o644)
}

// Report returns the NetLogo report as Markdown. The plot of a scenario is plots/<scenario>.png, relative to the
// report, if plots holds the scenario.
func Report(results []ScenarioResult, plots map[string]bool) string {
	var b strings.Builder
	b.WriteString("# Validation of gobees-fast against NetLogo\n\n" +
		"Written by go/compare for every pull request (go/compare/README.md). Each scenario runs 100 times in " +
		"gobees-fast and in NetLogo (the experiment of the same name in go/validation/experiments) and is compared " +
		"from day 1 on:\n\n" +
		"- Band (enforced): on every day, the median of the gobees-fast runs lies within the 5-95 % band of the " +
		"NetLogo runs.\n" +
		"- KS (reported only): two-sample Kolmogorov-Smirnov test of the yearly maximum and mean of each metric, " +
		"Bonferroni-corrected per scenario (family alpha 0.05).\n\n" +
		"The experiments of parts of the model (step3, step4a, step6) are tests in go/sys and not part of this " +
		"report.\n\n")
	b.WriteString("| Scenario | Band: metrics inside on all days | KS: significant tests |\n| --- | --- | --- |\n")
	for _, r := range results {
		inside, sig, tests := 0, 0, 0
		for _, m := range r.NetLogoResult.Metrics {
			if m.Inside == r.NetLogoResult.Days {
				inside++
			}
			sig += len(significantTests(m.Tests, r.NetLogoResult.Alpha))
			tests += len(m.Tests)
		}
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s |\n", r.Scenario.Title, r.Scenario.Name,
			mark(inside == len(r.NetLogoResult.Metrics), fmt.Sprintf("%d of %d", inside, len(r.NetLogoResult.Metrics))),
			mark(sig == 0, fmt.Sprintf("%d of %d", sig, tests)))
	}
	for _, r := range results {
		fmt.Fprintf(&b, "\n## %s (`%s`)\n\n%s %d days.\n\n", r.Scenario.Title, r.Scenario.Name, r.Scenario.Description,
			r.Scenario.Days)
		b.WriteString("| Metric | Band: days inside | KS |\n| --- | --- | --- |\n")
		for _, m := range r.NetLogoResult.Metrics {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", m.Metric,
				mark(m.Inside == r.NetLogoResult.Days, fmt.Sprintf("%d of %d", m.Inside, r.NetLogoResult.Days)),
				testCell(m.Tests, r.NetLogoResult.Alpha))
		}
		if plots[r.Scenario.Name] {
			fmt.Fprintf(&b, "\n![%s](plots/%s.png)\n", r.Scenario.Name, r.Scenario.Name)
		}
	}
	b.WriteString("\n" + plotCaption + "\n")
	return b.String()
}

func mark(passed bool, text string) string {
	if passed {
		return "✅ " + text
	}
	return "❌ " + text
}

// significantTests returns the tests with p < alpha, smallest p first (tests are sorted so).
func significantTests(tests []SummaryTest, alpha float64) []SummaryTest {
	return slices.DeleteFunc(slices.Clone(tests), func(s SummaryTest) bool { return s.P >= alpha })
}

// testCell formats the tests of one metric: the number of significant tests and the smallest p.
func testCell(tests []SummaryTest, alpha float64) string {
	if len(tests) == 0 {
		return "– no data"
	}
	sig := significantTests(tests, alpha)
	if len(sig) == 0 {
		return fmt.Sprintf("✅ 0 of %d", len(tests))
	}
	return fmt.Sprintf("❌ %d of %d, e.g. year %d, %s (p = %.2g)", len(sig), len(tests), sig[0].Year, sig[0].Kind,
		sig[0].P)
}
