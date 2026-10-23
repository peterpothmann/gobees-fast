package util

import "testing"

func TestPcor(t *testing.T) {
	// Finding 24 (see trans/netlogo-findings.md#finding-24): the 13 coordinates of the example farm on a patch border, and some others.
	cases := map[float64]int{
		0: 0, 0.49: 0, 0.5: 1, 132: 132,
		170.5: 171, 155.5: 156, 121.5: 122, 112.5: 113, 221.5: 222, 105.5: 106, 208.5: 209,
		88.5: 89, 81.5: 82, 142.5: 143, 153.5: 154, 146.5: 147, 48.5: 49, 210.5: 211,
		133.548: 134, 37.833: 38, 300: 300,
	}
	for v, want := range cases {
		if got := Pcor(v); got != want {
			t.Errorf("Pcor(%v) = %d, want %d", v, got, want)
		}
	}
}

func TestDistance(t *testing.T) {
	if got := Distance(0, 0, 3, 4); got != 5 {
		t.Errorf("Distance = %v, want 5", got)
	}
}
