package sys

import (
	"math"
	"testing"

	"gobees/comp"
	"gobees/input"
	"gobees/params"
)

// terrestris returns B_terrestris as read from the example species file.
func terrestris(t *testing.T) *params.Species {
	t.Helper()
	p := exampleFarmParams(t, nil, func(c *input.Config) {}) // the example farm has only B_terrestris
	return &p.Input.Species[0]
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Abs(b))
}

func expectPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: expected a panic (NetLogo: assertion or runtime error)", name)
		}
	}()
	f()
}

// defaults are the NetLogo values of the constants, which the test worlds use.
var defaults = params.DefaultConstants()

func TestThresholds(t *testing.T) {
	// all 24 values with the NetLogo values (NetLogo: ThresholdLevelREP)
	h := params.NotSetHigh
	want := map[string]comp.Thresholds{ // egg laying, nectar foraging, pollen foraging, nursing
		sitEgg:                  {Egglaying: h, NectarForaging: h, PollenForaging: h, Nursing: h},
		sitWorker:               {Egglaying: h, NectarForaging: 0.9, PollenForaging: 0.9, Nursing: 0.9},
		sitYoungQueen:           {Egglaying: h, NectarForaging: h, PollenForaging: h, Nursing: h},
		sitQueenInitiationPhase: {Egglaying: 0.1, NectarForaging: 0.7, PollenForaging: 0.7, Nursing: 0.5},
		sitQueenSocialPhase:     {Egglaying: 0, NectarForaging: h, PollenForaging: h, Nursing: 0.9},
		sitPsith:                {Egglaying: 0.2, NectarForaging: h, PollenForaging: h, Nursing: h},
	}
	for sit, th := range want {
		if got := thresholds(&defaults.Thresholds, sit); got != th {
			t.Errorf("thresholds(&defaults.Thresholds, %s) = %+v, want %+v", sit, got, th)
		}
	}
	expectPanic(t, "unknown situation", func() { thresholds(&defaults.Thresholds, "drone") })

	// each value of the run file lands in its place
	c := params.Thresholds{
		EggLaying:      params.EggLayingThresholds{QueenInitiationPhase: 1, QueenSocialPhase: 2, Psithyrus: 3},
		PollenForaging: params.ForagingThresholds{Worker: 4, QueenInitiationPhase: 5},
		NectarForaging: params.ForagingThresholds{Worker: 6, QueenInitiationPhase: 7},
		Nursing:        params.NursingThresholds{Worker: 8, QueenInitiationPhase: 9, QueenSocialPhase: 10},
	}
	set := map[string]comp.Thresholds{
		sitWorker:               {Egglaying: h, NectarForaging: 6, PollenForaging: 4, Nursing: 8},
		sitQueenInitiationPhase: {Egglaying: 1, NectarForaging: 7, PollenForaging: 5, Nursing: 9},
		sitQueenSocialPhase:     {Egglaying: 2, NectarForaging: h, PollenForaging: h, Nursing: 10},
		sitPsith:                {Egglaying: 3, NectarForaging: h, PollenForaging: h, Nursing: h},
	}
	for sit, th := range set {
		if got := thresholds(&c, sit); got != th {
			t.Errorf("thresholds(&defaults.Thresholds, %s) with set values = %+v, want %+v", sit, got, th)
		}
	}
}

func TestCropVolumeAndPollenLoad(t *testing.T) {
	sp := terrestris(t)
	// worked example: queen of 785 mg: min (173, 315.57) = 173, min (0.15, 315.57 * 0.0003) = 0.094671
	if got := cropVolume(785, sp, &defaults.Foraging); got != 173 {
		t.Errorf("crop volume 785 mg = %v, want 173 (species maximum)", got)
	}
	if got := pollenLoad(785, sp, &defaults.Foraging); !almostEqual(got, 0.094671) {
		t.Errorf("pollen pellets 785 mg = %v, want 0.094671", got)
	}
	// lightest worker, 62.4 mg: 62.4 * 0.402 = 25.0848 microlitres, 25.0848 * 0.0003 = 0.00752544 g
	if got := cropVolume(62.4, sp, &defaults.Foraging); !almostEqual(got, 25.0848) {
		t.Errorf("crop volume 62.4 mg = %v, want 25.0848", got)
	}
	if got := pollenLoad(62.4, sp, &defaults.Foraging); !almostEqual(got, 0.00752544) {
		t.Errorf("pollen pellets 62.4 mg = %v, want 0.00752544", got)
	}
	expectPanic(t, "weight 0", func() { cropVolume(0, sp, &defaults.Foraging) })
}

func TestProboscisLength(t *testing.T) {
	sp := terrestris(t)
	// 6.9 mm at 62.4 mg (lightest worker), 11.1 mm at 980 mg (heaviest queen)
	cases := map[float64]float64{62.4: 6.9, 980: 11.1, (62.4 + 980) / 2: 9.0}
	for weight, want := range cases {
		if got := proboscisLength(weight, sp); !almostEqual(got, want) {
			t.Errorf("proboscisLength(%v) = %v, want %v", weight, got, want)
		}
	}
	expectPanic(t, "weight below the lightest worker", func() { proboscisLength(50, sp) })
	expectPanic(t, "weight above the heaviest queen", func() { proboscisLength(1000, sp) })
}

func TestWinterMortalityProb(t *testing.T) {
	sp := terrestris(t) // queen weights 590 to 980 mg
	// 0.64 / (1 + e^(-22 * (relative weight - 0.32)))
	cases := map[float64]float64{
		590 + 0.32*390: 0.32,                               // relative weight 0.32: half of 0.64
		980:            0.64 / (1 + math.Exp(-22*0.68)),    // heaviest: 0.6399998
		590:            0.64 / (1 + math.Exp(-22*(-0.32))), // lightest: 0.00056
	}
	for weight, want := range cases {
		if got := winterMortalityProb(weight, sp); !almostEqual(got, want) {
			t.Errorf("winterMortalityProb(%v) = %v, want %v", weight, got, want)
		}
	}
	if got := winterMortalityProb(980, sp); got > 0.64 {
		t.Errorf("survival must stay below 0.64, got %v", got)
	}
	expectPanic(t, "weight above the maximum", func() { winterMortalityProb(981, sp) })
}
