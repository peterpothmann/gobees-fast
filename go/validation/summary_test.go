package validation

import (
	"testing"
)

func TestParseRuns(t *testing.T) {
	_, text, err := NetLogoReference("step7")
	if err != nil || text == "" {
		t.Fatalf("no run summaries of step7: %v", err)
	}
	runs, err := ParseRuns(text)
	if err != nil {
		t.Fatal(err)
	}
	// 3 years: the value at step 0 and the maximum and the mean of each year, 100 runs each
	for _, s := range []Summary{{"colonies", 0, "value"}, {"colonies", 1, "max"}, {"eggs", 3, "mean"}} {
		if n := len(runs[s]); n != 100 {
			t.Errorf("%s: %d runs, want 100", s, n)
		}
	}
	if got := runs[Summary{"hibernating_queens", 0, "value"}][0]; got != 500 {
		t.Errorf("hibernating queens at setup: got %v, want 500", got)
	}
}

func TestParseRunsHeader(t *testing.T) {
	if _, err := ParseRuns("step,metric\n"); err == nil {
		t.Error("a table with another header must be an error")
	}
}
