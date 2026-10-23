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

func TestPatchCentresWithin(t *testing.T) {
	// [NEST-03] strict "<": an assumed foodsource of 100 m2 (radius 0.564 patches with 10 m per patch) at the
	// coordinates of flower patch 24 of the example farm (133.548, 112.5). The nearest patch centres (134, 112)
	// and (134, 113) are at distance 0.674 > 0.564: none (finding 25, NetLogo: runtime error).
	if got := PatchCentresWithin(133.548, 112.5, 0.564, 0, 300, 0, 210); len(got) != 0 {
		t.Errorf("got %v, want no patch centre", got)
	}
	// (132, 170.5) with radius 0.564: (132, 170) and (132, 171) at distance 0.5
	if got := PatchCentresWithin(132, 170.5, 0.564, 0, 300, 0, 210); len(got) != 2 {
		t.Errorf("got %v, want 2 patch centres", got)
	}
	// distance exactly r is not inside; world borders cut the circle
	if got := PatchCentresWithin(0, 0, 1, 0, 300, 0, 210); len(got) != 1 || got[0] != [2]int{0, 0} {
		t.Errorf("got %v, want only (0, 0)", got)
	}
}
