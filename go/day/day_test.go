package day

import "testing"

// TestYearEnd checks YearEnd against NetLogo's 365 * ceiling (ticks / 365).
func TestYearEnd(t *testing.T) {
	for _, c := range []struct{ d, want Abs }{{0, 0}, {1, 365}, {364, 365}, {365, 365}, {366, 730}, {730, 730}, {731, 1095}} {
		if got := YearEnd(c.d); got != c.want {
			t.Errorf("YearEnd(%d) = %d, want %d", c.d, got, c.want)
		}
	}
}

// TestDate checks that an unset date behaves like NotSetHigh in NetLogo: it never lies before a day and is no day.
func TestDate(t *testing.T) {
	var unset Date
	if unset.IsSet() || unset.Before(1_000_000) || unset.Is(0) {
		t.Errorf("unset date: IsSet %v, Before %v, Is %v, want all false", unset.IsSet(), unset.Before(1_000_000), unset.Is(0))
	}
	d := At(10)
	if !d.IsSet() || d.Day() != 10 {
		t.Errorf("At(10): IsSet %v, Day %d", d.IsSet(), d.Day())
	}
	if d.Before(10) || !d.Before(11) {
		t.Errorf("At(10): Before(10) %v, Before(11) %v, want false, true", d.Before(10), d.Before(11))
	}
	if !d.Is(10) || d.Is(11) {
		t.Errorf("At(10): Is(10) %v, Is(11) %v, want true, false", d.Is(10), d.Is(11))
	}
	defer func() {
		if recover() == nil {
			t.Error("Day of an unset date: no panic")
		}
	}()
	unset.Day()
}

func TestDateString(t *testing.T) {
	if s := At(95).String(); s != "95" {
		t.Errorf("At(95).String() = %q", s)
	}
	if s := (Date{}).String(); s != "unset" {
		t.Errorf("unset String() = %q", s)
	}
}
