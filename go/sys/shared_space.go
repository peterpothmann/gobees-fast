package sys

import (
	"math"
	"math/rand/v2"
)

// randomPointInCircle returns a uniform random point within the circle of radius r [m] around (x, y) [m], with two
// random numbers: the distance from the centre r * sqrt(u) (uniform in area) and the angle 2 pi v.
func randomPointInCircle(x, y, r float64, rng *rand.Rand) (float64, float64) {
	d := r * math.Sqrt(rng.Float64())
	a := 2 * math.Pi * rng.Float64()
	return x + d*math.Cos(a), y + d*math.Sin(a)
}

// distanceBetween returns the distance between two points [m], without wrapping.
//
// NetLogo: distance, distancexy.
func distanceBetween(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}
