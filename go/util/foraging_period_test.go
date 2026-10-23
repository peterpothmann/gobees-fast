package util

import "testing"

func TestForagingPeriodConstant(t *testing.T) {
	if got := ForagingPeriod("Constant 8 hrs", 100); got != 8*3600 {
		t.Errorf("Constant 8 hrs: got %v, want %v", got, 8*3600)
	}
	if got := ForagingPeriod("Constant 24 hrs", 1); got != 24*3600 {
		t.Errorf("Constant 24 hrs: got %v, want %v", got, 24*3600)
	}
}

func TestForagingPeriodUnknownWeather(t *testing.T) {
	// [FPER-01] no option matches: -1
	for _, w := range []string{"", "Constant 13 hrs", "constant 8 hrs"} {
		if got := ForagingPeriod(w, 1); got != -1 {
			t.Errorf("ForagingPeriod(%q) = %v, want -1", w, got)
		}
	}
}

func TestForagingPeriodExample(t *testing.T) {
	// Finding 1: the list has 366 items, ".5" is item 42 (0-based).
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
		if got := ForagingPeriod("foragingHoursExample", c.day); got != c.hours*3600 {
			t.Errorf("day %d: got %v s, want %v s", c.day, got, c.hours*3600)
		}
	}
}
