package validation

import (
	"strings"
	"testing"
)

// recordExperiment records an experiment with the metric "bees", value(run, day) for 100 runs and 3 days, in dir.
func recordExperiment(t *testing.T, dir string, value func(run, day int) float64) {
	t.Helper()
	t.Setenv(DumpEnv, dir)
	v := Values{"bees": make([][]float64, 3)}
	for d := range 3 {
		for run := range 100 {
			v["bees"][d] = append(v["bees"][d], value(run, d))
		}
	}
	s := Section{Order: "01", Step: "Step 1", Experiment: "step3", Description: "test", Runs: 100,
		Rows: []ReportRow{{Metric: "colonies_founded", Steps: 3}}, Values: Values{"colonies_founded": v["bees"]}}
	if err := WriteReport(s); err != nil {
		t.Fatal(err)
	}
}

func TestCompare(t *testing.T) {
	// gobees-fast and gobees found as many colonies as NetLogo at the start of step 3 (0 on days 0 to 2), so the
	// band holds. gobees differs from gobees-fast: every summary differs, the KS tests against gobees fail.
	fast, gobees := t.TempDir(), t.TempDir()
	recordExperiment(t, fast, func(run, day int) float64 { return 0 })
	recordExperiment(t, gobees, func(run, day int) float64 { return float64(run + day + 1) })

	results, err := Compare(fast, gobees)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].Metrics) != 1 {
		t.Fatalf("results: %+v", results)
	}
	m := results[0].Metrics[0]
	if m.BandDays != 3 || len(m.Gobees) != 3 || len(m.Row.BaseMedians) != 3 {
		t.Errorf("band days %d, gobees tests %d, gobees medians %d; want 3, 3, 3", m.BandDays, len(m.Gobees),
			len(m.Row.BaseMedians))
	}
	if sig := significant(m.Gobees, results[0].GobeesAlpha); len(sig) != 3 {
		t.Errorf("significant tests against gobees: %v, want all 3", m.Gobees)
	}

	comment := Comment(results, "https://example.org/report")
	report := Report(results, map[string]bool{"step3": true})
	for _, want := range []string{CommentMarker, "| `step3` | ✅ 1 of 1 |", "❌ 3 of 3 |", "colonies_founded, setup",
		"https://example.org/report"} {
		if !strings.Contains(comment, want) {
			t.Errorf("comment lacks %q:\n%s", want, comment)
		}
	}
	for _, want := range []string{"## Step 1", "| colonies_founded | ✅ 3 of 3 |", "![step3](plots/step3.png)"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

func TestCompareWithoutGobees(t *testing.T) {
	fast := t.TempDir()
	recordExperiment(t, fast, func(run, day int) float64 { return 0 })
	results, err := Compare(fast, "")
	if err != nil {
		t.Fatal(err)
	}
	if results[0].HasGobees() || !strings.Contains(Comment(results, ""), "– no data") {
		t.Errorf("without gobees: %+v", results[0])
	}
}

func TestWriteReportWithoutDumpEnv(t *testing.T) {
	t.Setenv(DumpEnv, "")
	if err := WriteReport(Section{Experiment: "x", Values: testValues()}); err != nil {
		t.Error(err)
	}
}
