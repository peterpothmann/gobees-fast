//go:build ui || windows || (darwin && cgo)

package ui

import (
	"slices"
	"testing"
)

func TestFlowOrderStable(t *testing.T) {
	base := []flow{{0, 0, 5, 1, 3, colAccent, -1}, {0, 0, 2, 1, 3, colAccent, -1}, {0, 0, 2, 1, 3, colSecond, 1}, {0, 0, 9, 9, 7, colAccent, -1}}
	order := func(fs []flow) []flow {
		fs = slices.Clone(fs)
		slices.SortFunc(fs, flowOrder)
		return fs
	}
	want := order(base)
	for i := 0; i < 50; i++ {
		perm := slices.Clone(base)
		perm[i%4], perm[(i+1)%4] = perm[(i+1)%4], perm[i%4]
		if got := order(perm); !slices.Equal(got, want) {
			t.Fatalf("order depends on the input order: %v vs %v", got, want)
		}
	}
}

// TestPlaceLabels: labels on the same spot, as for the nectar and pollen line to the same patch, do not cover
// each other.
func TestPlaceLabels(t *testing.T) {
	labels := []mapLabel{}
	for range 5 {
		labels = append(labels, mapLabel{x0: 100, y0: 100, x1: 130, y1: 100, text: "25 bees"})
	}
	rs := placeLabels(labels)
	for i := range rs {
		for j := i + 1; j < len(rs); j++ {
			if rs[i].Overlaps(rs[j]) {
				t.Errorf("labels %d and %d overlap: %v %v", i, j, rs[i], rs[j])
			}
		}
	}
}
