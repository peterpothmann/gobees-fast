package util

import "testing"

func TestForaging_periodREPConstant(t *testing.T) {
	if got := Foraging_periodREP("Constant 8 hrs", 100); got != 8*3600 {
		t.Errorf("Constant 8 hrs: got %v, want %v", got, 8*3600)
	}
	if got := Foraging_periodREP("Constant 24 hrs", 1); got != 24*3600 {
		t.Errorf("Constant 24 hrs: got %v, want %v", got, 24*3600)
	}
}

func TestForaging_periodREPUnknownWeather(t *testing.T) {
	// [FPER-01] no option matches: -1
	for _, w := range []string{"", "Constant 13 hrs", "constant 8 hrs"} {
		if got := Foraging_periodREP(w, 1); got != -1 {
			t.Errorf("Foraging_periodREP(%q) = %v, want -1", w, got)
		}
	}
}

func TestForaging_periodREPExample(t *testing.T) {
	// Finding 1 (see trans/netlogo-findings.md#finding-1): the list has 366 items, ".5" is item 42 (0-based).
	if len(foragingHoursExample) != 366 {
		t.Fatalf("len = %d, want 366 as in NetLogo", len(foragingHoursExample))
	}
	cases := []struct {
		day   int
		hours float64
	}{
		{1, 0},
		{2, 3.1},
		{42, 0},   // probably intended: 0.5
		{43, 0.5}, // the ".5"
		{44, 9.2},
		{365, 0},
	}
	for _, c := range cases {
		if got := Foraging_periodREP("foragingHoursExample", c.day); got != c.hours*3600 {
			t.Errorf("day %d: got %v s, want %v s", c.day, got, c.hours*3600)
		}
	}
}
