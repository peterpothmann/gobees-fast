package sys

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestDistanceBetween(t *testing.T) {
	if got := distanceBetween(0, 0, 3, 4); got != 5 {
		t.Errorf("distanceBetween = %v, want 5", got)
	}
}

func TestRandomPointInCircle(t *testing.T) {
	// 100000 points in a circle of radius 10 m around (1000, 2000): all inside, uniform in area. The mean squared
	// distance from the centre is r^2 / 2 = 50 m2 (sd of the mean 0.09), each quadrant holds a quarter (sd 0.0014).
	rng := rand.New(rand.NewPCG(1, 2))
	const n, r = 100000, 10.0
	sumSq := 0.0
	quadrant := [4]int{}
	for range n {
		x, y := randomPointInCircle(1000, 2000, r, rng)
		d := distanceBetween(1000, 2000, x, y)
		if d > r {
			t.Fatalf("point (%v, %v) at %v m, outside the circle", x, y, d)
		}
		sumSq += d * d
		q := 0
		if x > 1000 {
			q++
		}
		if y > 2000 {
			q += 2
		}
		quadrant[q]++
	}
	if mean := sumSq / n; math.Abs(mean-50) > 0.5 {
		t.Errorf("mean squared distance %v m2, want 50", mean)
	}
	for q, k := range quadrant {
		if share := float64(k) / n; math.Abs(share-0.25) > 0.01 {
			t.Errorf("quadrant %d holds %v of the points, want 0.25", q, share)
		}
	}
	// radius 0: the centre
	if x, y := randomPointInCircle(5, 6, 0, rng); x != 5 || y != 6 {
		t.Errorf("radius 0: (%v, %v), want (5, 6)", x, y)
	}
}
