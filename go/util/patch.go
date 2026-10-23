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

// Distance returns the distance between two points in NetLogo patch units, without wrapping.
//
// NetLogo: distance, distancexy.
func Distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}
