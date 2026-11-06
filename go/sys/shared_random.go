package sys

import (
	"math"
	"math/rand/v2"
)

// shuffle shuffles a slice with the model RNG. Used wherever NetLogo asks several agents in random order.
func shuffle[T any](rng *rand.Rand, list []T) {
	rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

// randomFloat returns a random number in [0, x). NetLogo: random-float x.
func randomFloat(rng *rand.Rand, x float64) float64 {
	return rng.Float64() * x
}

// randomInt returns a random integer in [0, n). NetLogo: random n (n is truncated to an integer).
// As in NetLogo, a truncated n of 0 gives 0 (rand.IntN would panic), and a negative n gives a number in (n, 0].
func randomInt(rng *rand.Rand, n float64) float64 {
	k := int(n)
	if k == 0 {
		return 0
	}
	if k < 0 {
		return -float64(rng.IntN(-k))
	}
	return float64(rng.IntN(k))
}

// randomNormal returns a normally distributed random number. NetLogo: random-normal mean sd.
func randomNormal(rng *rand.Rand, mean, sd float64) float64 {
	return mean + sd*rng.NormFloat64()
}

// round rounds like NetLogo's round: .5 is rounded up (towards positive infinity), e.g. 2.5 -> 3, -2.5 -> -2.
func round(x float64) float64 {
	return math.Floor(x + 0.5)
}
