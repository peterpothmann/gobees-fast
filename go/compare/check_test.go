package compare

import (
	"strings"
	"testing"
	"time"

	"gobees/validation"
)

// netLogoRuns returns 20 runs of the scenario step7 whose column of each NetLogo metric holds value(netlogo median
// of the day) on every day.
func netLogoRuns(t *testing.T, value func(median float64) float64) []Run {
	t.Helper()
	text, _, err := validation.NetLogoReference("step7")
	if err != nil {
		t.Fatal(err)
	}
	quantiles, err := validation.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	header := []string{"ticks"}
	for k := range quantiles {
		if k.Step == 1 {
			header = append(header, k.Metric)
		}
	}
	tab := Table{Header: header}
	for day := 1; day <= 1095; day++ {
		row := []float64{float64(day)}
		for _, m := range header[1:] {
			row = append(row, value(quantiles[validation.Key{Step: day, Metric: m}].Q50))
		}
		tab.Rows = append(tab.Rows, row)
	}
	result := make([]Run, 20)
	for i := range result {
		result[i] = Run{Seed: int64(i + 1), Table: tab}
	}
	return result
}

func TestCompareNetLogo(t *testing.T) {
	r, err := CompareNetLogo(Scenarios[0], netLogoRuns(t, func(m float64) float64 { return m }))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Metrics) != 15 || !r.Passed() {
		t.Fatalf("%d metrics, passed %v", len(r.Metrics), r.Passed())
	}
	if c := NetLogoCell(r); !c.Passed || c.Text != "15 of 15 metrics in band" {
		t.Errorf("cell: %+v", c)
	}

	// far above the NetLogo runs on every day where they are not all 0
	r, err = CompareNetLogo(Scenarios[0], netLogoRuns(t, func(m float64) float64 { return 2*m + 1e6 }))
	if err != nil {
		t.Fatal(err)
	}
	c := NetLogoCell(r)
	if c.Passed || c.Text != "0 of 15 metrics in band" || !strings.Contains(c.Details, "- colonies: 0 of 1095 days") {
		t.Errorf("cell: %+v", c)
	}
}

func TestCells(t *testing.T) {
	ref := runs([]float64{1, 2, 3}, []float64{1, 2, 4})
	if c := ExactCell(ref, ref, []string{"ticks", "A"}); !c.Passed || c.Text != "identical" {
		t.Errorf("exact, equal: %+v", c)
	}
	c := ExactCell(ref, runs([]float64{1, 2, 3}, []float64{1, 2, 5}), []string{"ticks", "A"})
	if c.Passed || c.Text != "1 of 2 runs differ" || !strings.Contains(c.Details, "seed 0: day 3, column A") {
		t.Errorf("exact, one run differs: %+v", c)
	}
	if c := MissingColumnsCell([]string{"A"}); c.Passed || !c.Ran {
		t.Errorf("missing columns: %+v", c)
	}
}

func TestComment(t *testing.T) {
	passed := Cell{Ran: true, Passed: true, Text: "identical"}
	results := []ScenarioResult{{
		Scenario: Scenarios[0], Changed: passed,
		Drift:            Cell{Ran: true, Text: "differs (3 days outside the band, 0 Kolmogorov-Smirnov tests)", Details: "- day 7"},
		NetLogo:          Cell{Ran: true, Passed: true, Text: "15 of 15 metrics in band"},
		ReferenceColumns: []string{"A", "B"}, BaseColumns: []string{"A", "B"},
		NotInReference: []string{"C"}, NotInBase: []string{"C"},
	}}
	info := CommentInfo{Phase: 1, Reference: "ae372c6", Base: "40dc301", Seeds: 100, RunURL: "https://example.org/run"}
	comment := Comment(results, &Speed{Base: 200 * time.Millisecond, Current: 100 * time.Millisecond, Days: 365}, info)
	for _, want := range []string{
		CommentMarker,
		"| Scenario | its previous step (`ae372c6`), exact | gobees (`40dc301`), statistical | NetLogo |",
		"| Example farm, 3 years | ✅ identical | ❌ differs (3 days outside the band, 0 Kolmogorov-Smirnov tests) | ✅ 15 of 15 metrics in band |",
		"| Badgers | ❌ no result | ❌ no result | ❌ no result |",
		"**Speed:** 2.00× as fast as gobees `40dc301` (200ms → 100ms per run of 365 days",
		"| Example farm, 3 years | 1095 | 2 | 2 | 0 |",
		"| Badgers | 1095 | – | – | – |",
		"Not compared with the previous step and gobees: 1 columns it does not write (C).",
		"<summary>Example farm, 3 years: differences</summary>\n\n**Against gobees:** differs",
		"Logs: [run](https://example.org/run)",
	} {
		if !strings.Contains(comment, want) {
			t.Errorf("comment lacks %q:\n%s", want, comment)
		}
	}
	if strings.Count(comment, "<details>") != 2 {
		t.Errorf("want the runs and the differences of one scenario:\n%s", comment)
	}

	info.Phase = 2
	if c := Comment(nil, nil, info); !strings.Contains(c, "its previous step (`ae372c6`), statistical") ||
		!strings.Contains(c, "**Speed:** no result") {
		t.Errorf("phase 2 without results:\n%s", c)
	}
}

func TestReport(t *testing.T) {
	r, err := CompareNetLogo(Scenarios[0], netLogoRuns(t, func(m float64) float64 { return m }))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := WriteReport(dir, []ScenarioResult{{Scenario: Scenarios[0], NetLogoResult: r}}); err != nil {
		t.Fatal(err)
	}
	report := Report([]ScenarioResult{{Scenario: Scenarios[0], NetLogoResult: r}}, map[string]bool{"step7": true})
	for _, want := range []string{"| Example farm, 3 years (`step7`) | ✅ 15 of 15 |", "| colonies | ✅ 1095 of 1095 |",
		"![step7](plots/step7.png)"} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
}
