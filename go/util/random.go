package util

import (
	"math"
	"math/rand/v2"
)

// Shuffle shuffles a slice with the model RNG. Used wherever NetLogo uses ask over several agents
// (decision C in trans/data-model.md).
func Shuffle[T any](rng *rand.Rand, list []T) {
	rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

// RandomFloat returns a random number bbetween [0, x). NetLogo: random-float x.
func RandomFloat(rng *rand.Rand, x float64) float64 {
	return rng.Float64() * x
}

// RandomInt returns a random integer between [0, n) for n > 0. NetLogo: random n (n is truncated to an integer).
func RandomInt(rng *rand.Rand, n float64) float64 {
	return float64(rng.IntN(int(n)))
}

// RandomNormal returns a normally distributed random number. NetLogo: random-normal mean sd.
func RandomNormal(rng *rand.Rand, mean, sd float64) float64 {
	return mean + sd*rng.NormFloat64()
}

// Round rounds like NetLogo's round: .5 is rounded up (towards positive infinity), e.g. 2.5 -> 3, -2.5 -> -2.
func Round(x float64) float64 {
	return math.Floor(x + 0.5)
}
