package validation

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"gobees/compare"
)

// The comparison of the recorded results of gobees-fast (WriteReport) with NetLogo, written by cmd/report as the
// validation report and as the NetLogo row of the pull request comment (NetLogoCheck, compare.Comment).
//
// Two criteria per metric:
//   - Band: the gobees-fast median lies within the NetLogo 5-95 % band on every compared day. Enforced by the tests.
//   - KS: a two-sample Kolmogorov-Smirnov test per run summary (RunSummaries: the value at day 0, the yearly
//     maximum and mean) against NetLogo. Bonferroni-corrected per experiment, family alpha compare.FamilyAlpha.
//     Reported in the full report, not enforced.

// Test is one Kolmogorov-Smirnov test of a run summary.
type Test struct {
	Summary Summary
	D, P    float64
}

// MetricResult is the comparison of one metric of an experiment.
type MetricResult struct {
	Metric   string
	Days     int    // compared days 0 to Days-1
	BandDays int    // days on which the gobees-fast median lies within the NetLogo 5-95 % band
	NetLogo  []Test // gobees-fast against NetLogo, nil without NetLogo run summaries
	Row      ReportRow
}

// ExperimentResult is the comparison of one experiment.
type ExperimentResult struct {
	Meta    Meta
	Metrics []MetricResult
	// NetLogoAlpha is the Bonferroni-corrected level: FamilyAlpha / number of tests.
	NetLogoAlpha float64
}

// HasNetLogo reports whether NetLogo run summaries exist for the experiment.
func (e ExperimentResult) HasNetLogo() bool { return e.Metrics[0].NetLogo != nil }

// Compare compares the experiments recorded in fastDir (gobees-fast) with NetLogo. The experiments are sorted by
// Meta.Order.
func Compare(fastDir string) ([]ExperimentResult, error) {
	metas, err := filepath.Glob(filepath.Join(fastDir, "*_meta.json"))
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, fmt.Errorf("no recorded experiments in %s", fastDir)
	}
	results := []ExperimentResult{}
	for _, path := range metas {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var meta Meta
		if err := json.Unmarshal(b, &meta); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		r, err := compareExperiment(meta, fastDir)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", meta.Experiment, err)
		}
		results = append(results, r)
	}
	slices.SortFunc(results, func(a, b ExperimentResult) int { return strings.Compare(a.Meta.Order, b.Meta.Order) })
	return results, nil
}

// readRecord reads the recorded results of an experiment in dir. ok is false if dir has none.
func readRecord(dir, experiment string) (q map[Key]Quantiles, runs map[Summary][]float64, ok bool, err error) {
	qb, err := os.ReadFile(filepath.Join(dir, experiment+".csv"))
	if err != nil {
		return nil, nil, false, nil
	}
	rb, err := os.ReadFile(filepath.Join(dir, experiment+"_runs.csv"))
	if err != nil {
		return nil, nil, false, err
	}
	if q, err = Parse(string(qb)); err != nil {
		return nil, nil, false, err
	}
	if runs, err = ParseRuns(string(rb)); err != nil {
		return nil, nil, false, err
	}
	return q, runs, true, nil
}

func compareExperiment(meta Meta, fastDir string) (ExperimentResult, error) {
	fastQ, fastRuns, ok, err := readRecord(fastDir, meta.Experiment)
	if err != nil || !ok {
		return ExperimentResult{}, fmt.Errorf("no recorded results in %s: %v", fastDir, err)
	}
	nlText, nlRunsText, err := NetLogoReference(meta.Experiment)
	if err != nil {
		return ExperimentResult{}, err
	}
	nlQ, err := Parse(nlText)
	if err != nil {
		return ExperimentResult{}, err
	}
	var nlRuns map[Summary][]float64
	if nlRunsText != "" {
		if nlRuns, err = ParseRuns(nlRunsText); err != nil {
			return ExperimentResult{}, err
		}
	}

	result := ExperimentResult{Meta: meta}
	nNetLogo := 0
	for _, m := range meta.Metrics {
		r := MetricResult{Metric: m.Metric, Days: m.Days, Row: ReportRow{Metric: m.Metric, Steps: m.Days}}
		for d := range m.Days {
			nl, ok := nlQ[Key{d, m.Metric}]
			if !ok {
				return result, fmt.Errorf("no NetLogo reference for %s at step %d", m.Metric, d)
			}
			fast := fastQ[Key{d, m.Metric}].Q50
			tol := 1e-9 * math.Max(1, math.Abs(nl.Q50)) // float sums in a random order differ in the last digits
			if fast >= nl.Q05-tol && fast <= nl.Q95+tol {
				r.BandDays++
			}
			r.Row.Medians = append(r.Row.Medians, fast)
			r.Row.Bands = append(r.Row.Bands, nl)
		}
		r.Row.Within = r.BandDays
		if nlRuns != nil {
			for _, s := range summaries(m) {
				r.NetLogo = append(r.NetLogo, ksTest(s, fastRuns[s], nlRuns[s]))
				nNetLogo++
			}
		}
		result.Metrics = append(result.Metrics, r)
	}
	result.NetLogoAlpha = compare.FamilyAlpha / float64(max(1, nNetLogo))
	return result, nil
}

// summaries returns the run summaries compared for a metric: the value at day 0, and the maximum and the mean of
// every year that has compared days.
func summaries(m MetricMeta) []Summary {
	s := []Summary{{m.Metric, 0, "value"}}
	for year := 1; (year-1)*DaysPerYear+1 <= m.Days-1; year++ {
		s = append(s, Summary{m.Metric, year, "max"}, Summary{m.Metric, year, "mean"})
	}
	return s
}

// ksTest compares the summaries of two versions. Values are rounded to 9 significant digits first: float sums in a
// random order differ in the last digits, which is no difference of the model.
func ksTest(s Summary, a, b []float64) Test {
	d, p := compare.KolmogorovSmirnov(round9(a), round9(b))
	return Test{Summary: s, D: d, P: p}
}

func round9(values []float64) []float64 {
	r := make([]float64, len(values))
	for i, x := range values {
		r[i], _ = strconv.ParseFloat(strconv.FormatFloat(x, 'g', 9, 64), 64)
	}
	return r
}

// significant returns the tests with p < alpha, smallest p first.
func significant(tests []Test, alpha float64) []Test {
	s := slices.DeleteFunc(slices.Clone(tests), func(t Test) bool { return t.P >= alpha })
	slices.SortFunc(s, func(a, b Test) int {
		if a.P != b.P {
			if a.P < b.P {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Summary.String(), b.Summary.String())
	})
	return s
}

// counts sums the band days and the significant tests of an experiment.
func (e ExperimentResult) counts() (bandOK, metrics, nlSig, nlTests int) {
	for _, m := range e.Metrics {
		metrics++
		if m.BandDays == m.Days {
			bandOK++
		}
		nlSig += len(significant(m.NetLogo, e.NetLogoAlpha))
		nlTests += len(m.NetLogo)
	}
	return
}

// cell formats the result of one comparison for a table: a mark and the number of significant tests.
func cell(sig, tests int, available bool) string {
	switch {
	case !available:
		return "– no data"
	case sig == 0:
		return fmt.Sprintf("✅ 0 of %d", tests)
	}
	return fmt.Sprintf("❌ %d of %d", sig, tests)
}

func bandCell(ok, n int) string {
	if ok == n {
		return fmt.Sprintf("✅ %d of %d", ok, n)
	}
	return fmt.Sprintf("❌ %d of %d", ok, n)
}

func experimentName(m Meta) string {
	if m.Title != "" {
		return fmt.Sprintf("%s (`%s`)", m.Title, m.Experiment)
	}
	return "`" + m.Experiment + "`"
}

// summaryTable returns the table with one row per experiment.
func summaryTable(results []ExperimentResult) string {
	var b strings.Builder
	b.WriteString("| Experiment | Band: metrics inside on all days | KS vs NetLogo: significant tests |\n" +
		"| --- | --- | --- |\n")
	for _, e := range results {
		bandOK, metrics, nlSig, nlTests := e.counts()
		fmt.Fprintf(&b, "| %s | %s | %s |\n", experimentName(e.Meta), bandCell(bandOK, metrics),
			cell(nlSig, nlTests, e.HasNetLogo()))
	}
	return b.String()
}

const criteria = "Band: the median of the gobees-fast runs lies within the 5-95 % band of the NetLogo runs on every " +
	"compared day (enforced by the tests, weak). KS: two-sample Kolmogorov-Smirnov test of the run summaries (value " +
	"at day 0, maximum and mean of every year) of gobees-fast against NetLogo, Bonferroni-corrected per experiment " +
	"(family alpha 0.05). Reported, not enforced. 100 runs per model and experiment."

// NetLogoCheck returns the NetLogo row of the pull request comment (compare.Comment): passed if every metric of every
// experiment lies within the NetLogo band on all days. reportURL links to the full report, may be empty.
func NetLogoCheck(results []ExperimentResult, reportURL string) compare.Check {
	c := compare.Question("3-netlogo")
	var b strings.Builder
	b.WriteString("| Experiment | Metrics inside the band on all days |\n| --- | --- |\n")
	outside := []string{}
	for _, e := range results {
		bandOK, metrics, _, _ := e.counts()
		if bandOK < metrics {
			outside = append(outside, "`"+e.Meta.Experiment+"`")
		}
		fmt.Fprintf(&b, "| %s | %s |\n", experimentName(e.Meta), bandCell(bandOK, metrics))
	}
	if reportURL != "" {
		fmt.Fprintf(&b, "\nFull report with plots and the Kolmogorov-Smirnov tests: %s\n", reportURL)
	}
	c.Details = b.String()

	c.Status = "✅"
	c.Result = fmt.Sprintf("%d of %d validation experiments inside the NetLogo 5–95 %% band on all days",
		len(results)-len(outside), len(results))
	if len(outside) > 0 {
		c.Status = "❌"
		c.Result += ", outside: " + strings.Join(outside, ", ")
	}
	runs := 0
	if len(results) > 0 {
		runs = results[0].Meta.Runs
	}
	c.Scope = fmt.Sprintf("The %d validation experiments of go/validation, %d runs per experiment and model.",
		len(results), runs)
	return c
}

// Report returns the full validation report as Markdown. The plot of an experiment is plots/<experiment>.png,
// relative to the report, if plots holds the experiment.
func Report(results []ExperimentResult, plots map[string]bool) string {
	var b strings.Builder
	b.WriteString("# Validation of gobees-fast\n\n" + criteria + " Details: go/validation/README.md.\n\n")
	b.WriteString(summaryTable(results))
	lastStep := ""
	for _, e := range results {
		if e.Meta.Step != lastStep {
			fmt.Fprintf(&b, "\n## %s\n", e.Meta.Step)
			lastStep = e.Meta.Step
		}
		if e.Meta.Title != "" {
			fmt.Fprintf(&b, "\n### %s\n", e.Meta.Title)
		}
		fmt.Fprintf(&b, "\n%s\n\n", e.Meta.Description)
		b.WriteString("| Metric | Band: days inside | KS vs NetLogo |\n| --- | --- | --- |\n")
		for _, m := range e.Metrics {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", m.Metric, bandCell(m.BandDays, m.Days),
				testCell(m.NetLogo, e.NetLogoAlpha, e.HasNetLogo()))
		}
		if plots[e.Meta.Experiment] {
			fmt.Fprintf(&b, "\n![%s](plots/%s.png)\n", e.Meta.Experiment, e.Meta.Experiment)
		}
	}
	b.WriteString("\n" + plotCaption + "\n")
	return b.String()
}

// testCell formats the tests of one metric: the number of significant tests and the smallest p.
func testCell(tests []Test, alpha float64, available bool) string {
	if !available {
		return "– no data"
	}
	sig := significant(tests, alpha)
	if len(sig) == 0 {
		return fmt.Sprintf("✅ 0 of %d", len(tests))
	}
	return fmt.Sprintf("❌ %d of %d, e.g. %s (p = %.2g)", len(sig), len(tests),
		strings.TrimPrefix(sig[0].Summary.String(), sig[0].Summary.Metric+", "), sig[0].P)
}

// WritePlots writes plots/<experiment>.png into dir for every experiment and returns the experiments that got one.
func WritePlots(dir string, results []ExperimentResult) (map[string]bool, error) {
	plotDir := filepath.Join(dir, "plots")
	if err := os.MkdirAll(plotDir, 0o755); err != nil {
		return nil, err
	}
	written := map[string]bool{}
	for _, e := range results {
		rows := make([]ReportRow, len(e.Metrics))
		for i, m := range e.Metrics {
			rows[i] = m.Row
		}
		ok, err := writePlot(filepath.Join(plotDir, e.Meta.Experiment+".png"), rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Meta.Experiment, err)
		}
		written[e.Meta.Experiment] = ok
	}
	return written, nil
}
