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
	// gobees-fast founds as many colonies as NetLogo at the start of step 3 (0 on days 0 to 2), so the band holds.
	fast := t.TempDir()
	recordExperiment(t, fast, func(run, day int) float64 { return 0 })

	results, err := Compare(fast)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].Metrics) != 1 {
		t.Fatalf("results: %+v", results)
	}
	if m := results[0].Metrics[0]; m.BandDays != 3 {
		t.Errorf("band days %d, want 3", m.BandDays)
	}

	check := NetLogoCheck(results, "https://example.org/report")
	if check.Status != "✅" || check.Result != "1 of 1 validation experiments inside the NetLogo 5–95 % band on all days" {
		t.Errorf("check: %q %q", check.Status, check.Result)
	}
	for _, want := range []string{"| `step3` | ✅ 1 of 1 |", "https://example.org/report"} {
		if !strings.Contains(check.Details, want) {
			t.Errorf("details lack %q:\n%s", want, check.Details)
		}
	}
	report := Report(results, map[string]bool{"step3": true})
	for _, want := range []string{"## Step 1", "| colonies_founded | ✅ 3 of 3 |", "![step3](plots/step3.png)"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}

func TestNetLogoCheckOutsideBand(t *testing.T) {
	// 1 + run + day colonies founded on days 0 to 2: NetLogo founds none before day 3.
	fast := t.TempDir()
	recordExperiment(t, fast, func(run, day int) float64 { return float64(1 + run + day) })
	results, err := Compare(fast)
	if err != nil {
		t.Fatal(err)
	}
	check := NetLogoCheck(results, "")
	if check.Status != "❌" || !strings.Contains(check.Result, "0 of 1 validation experiments") ||
		!strings.Contains(check.Result, "outside: `step3`") {
		t.Errorf("check: %q %q", check.Status, check.Result)
	}
}

func TestWriteReportWithoutDumpEnv(t *testing.T) {
	t.Setenv(DumpEnv, "")
	if err := WriteReport(Section{Experiment: "x", Values: testValues()}); err != nil {
		t.Error(err)
	}
}
