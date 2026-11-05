package util

import "math"

// Pcor returns the patch coordinate (pxcor or pycor) of a turtle coordinate.
//
// NetLogo rounds .5 up, e.g. 121.5 -> 122 (verified in NetLogo 5.3.1, finding 24, see trans/netlogo-findings.md#finding-24).
// math.Floor(v + 0.5) is only correct for v >= -0.5, which holds in this world (0-300 x 0-210, no wrapping).
//
// NetLogo: pxcor, pycor, patch-here, -here.
func Pcor(v float64) int {
	return int(math.Floor(v + 0.5))
}

// PatchCentresWithin returns the centres (pxcor, pycor) of all patches of the world whose distance to (x, y) is
// strictly less than r. NetLogo: patches with [ distance myself < r ] (used in [NEST-03]).
func PatchCentresWithin(x, y, r float64, minPxcor, maxPxcor, minPycor, maxPycor int) [][2]int {
	result := [][2]int{}
	// The patch centres are at integer coordinates, so we can limit the search to the bounding box of the circle.
	x0, x1 := max(minPxcor, int(math.Floor(x-r))), min(maxPxcor, int(math.Ceil(x+r)))
	y0, y1 := max(minPycor, int(math.Floor(y-r))), min(maxPycor, int(math.Ceil(y+r)))

	// Check all patch centres in the bounding box
	for px := x0; px <= x1; px++ {
		for py := y0; py <= y1; py++ {
			// The distance is strictly less than r, not less than or equal to r.
			if Distance(x, y, float64(px), float64(py)) < r {
				// Add those that are strictly inside the circle.
				result = append(result, [2]int{px, py})
			}
		}
	}
	return result
}

// Distance returns the distance between two points in NetLogo patch units, without wrapping.
//
// NetLogo: distance, distancexy.
func Distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}
