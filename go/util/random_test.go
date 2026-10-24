package util

import (
	"math/rand/v2"
	"testing"
)

func TestRandomInt(t *testing.T) {
	// NetLogo: random 0 is 0, random n is truncated, random of a negative n is in (n, 0] ([PCH-05])
	rng := rand.New(rand.NewPCG(1, 2))
	for range 1000 {
		if got := RandomInt(rng, 0); got != 0 {
			t.Fatalf("random 0 = %v, want 0", got)
		}
		if got := RandomInt(rng, 0.7); got != 0 {
			t.Fatalf("random 0.7 = %v, want 0", got)
		}
		if got := RandomInt(rng, 3.9); got < 0 || got > 2 {
			t.Fatalf("random 3.9 = %v, want 0, 1 or 2", got)
		}
		if got := RandomInt(rng, -3); got > 0 || got < -2 {
			t.Fatalf("random -3 = %v, want 0, -1 or -2", got)
		}
	}
}
