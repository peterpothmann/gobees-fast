package validation

// Keep this file the same as go/validation/summary_test.go of gobees-fast.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// testValues returns 2 runs of 400 days of the metric "bees": run 1 counts the days, run 2 is 10 every day.
func testValues() Values {
	steps := make([][]float64, 401)
	for d := range steps {
		steps[d] = []float64{float64(d), 10}
	}
	return Values{"bees": steps}
}

func TestRunSummaries(t *testing.T) {
	// day 0, year 1 = days 1 to 365, year 2 = days 366 to 400 (a short last year)
	got := RunSummaries(testValues())
	want := map[Summary][]float64{
		{"bees", 0, "value"}: {0, 10},
		{"bees", 1, "max"}:   {365, 10},
		{"bees", 1, "mean"}:  {183, 10},
		{"bees", 2, "max"}:   {400, 10},
		{"bees", 2, "mean"}:  {383, 10},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestQuantile(t *testing.T) {
	// type 7, as Python: statistics.quantiles([1, 2, 3, 4, 10], n=20, method="inclusive")[0] = 1.2, [18] = 8.8
	v := []float64{10, 1, 3, 2, 4}
	for q, want := range map[float64]float64{0.05: 1.2, 0.5: 3, 0.95: 8.8, 1: 10, 0: 1} {
		if got := Quantile(v, q); got < want-1e-12 || got > want+1e-12 {
			t.Errorf("Quantile(%v) = %v, want %v", q, got, want)
		}
	}
}

func TestWriteDumpRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := WriteDump(dir, "exp", testValues()); err != nil {
		t.Fatal(err)
	}
	qb, err := os.ReadFile(filepath.Join(dir, "exp.csv"))
	if err != nil {
		t.Fatal(err)
	}
	q, err := Parse(string(qb))
	if err != nil {
		t.Fatal(err)
	}
	if got := q[Key{365, "bees"}]; got.N != 2 || got.Q50 != (365+10)/2.0 || got.Q05 != 10+0.05*355 {
		t.Errorf("quantiles at day 365: %+v", got)
	}
	rb, err := os.ReadFile(filepath.Join(dir, "exp_runs.csv"))
	if err != nil {
		t.Fatal(err)
	}
	runs, err := ParseRuns(string(rb))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(runs, RunSummaries(testValues())) {
		t.Errorf("run summaries changed by writing and reading: %v", runs)
	}
}
