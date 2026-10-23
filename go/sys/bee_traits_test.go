package sys

import (
	"math"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/internal/examplefiles"
	"gobees/params"
)

// terrestris returns B_terrestris as read from the example species file.
func terrestris(t *testing.T) *comp.Species {
	t.Helper()
	constants := params.DefaultConstants()
	species, err := parseSpeciesFile(strings.NewReader(examplefiles.BumbleSpecies), []string{"B_terrestris"}, &constants)
	if err != nil {
		t.Fatal(err)
	}
	return &species[0]
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

func TestThresholdLevelREP(t *testing.T) {
	// [THR] all 24 values of ThresholdLevelREP
	h := params.NotSetHigh
	want := map[string][6]float64{ // egg, worker, youngQueen, QueenInitiationPhase, QueenSocialPhase, Psith
		thEggLaying:      {h, h, h, 0.1, 0, 0.2},
		thPollenForaging: {h, 0.9, h, 0.7, h, h},
		thNectarForaging: {h, 0.9, h, 0.7, h, h},
		thNursing:        {h, 0.9, h, 0.5, 0.9, h},
	}
	situations := []string{sitEgg, sitWorker, sitYoungQueen, sitQueenInitiationPhase, sitQueenSocialPhase, sitPsith}
	for thType, values := range want {
		for i, sit := range situations {
			if got := ThresholdLevelREP(thType, sit); got != values[i] {
				t.Errorf("ThresholdLevelREP(%s, %s) = %v, want %v", thType, sit, got, values[i])
			}
		}
	}
	expectPanic(t, "unknown situation", func() { ThresholdLevelREP(thNursing, "drone") })
	expectPanic(t, "unknown type", func() { ThresholdLevelREP("guarding", sitWorker) })
}

func TestCropAndPelletSizeREP(t *testing.T) {
	sp := terrestris(t)
	// [LOAD] worked example: queen of 785 mg: min (173, 315.57) = 173, min (0.15, 315.57 * 0.0003) = 0.094671
	if got := CropAndPelletSizeREP("nectar", 785, sp); got != 173 {
		t.Errorf("crop volume 785 mg = %v, want 173 (species maximum)", got)
	}
	if got := CropAndPelletSizeREP("pollen", 785, sp); !almostEqual(got, 0.094671) {
		t.Errorf("pollen pellets 785 mg = %v, want 0.094671", got)
	}
	// lightest worker, 62.4 mg: 62.4 * 0.402 = 25.0848 microlitres, 25.0848 * 0.0003 = 0.00752544 g
	if got := CropAndPelletSizeREP("nectar", 62.4, sp); !almostEqual(got, 25.0848) {
		t.Errorf("crop volume 62.4 mg = %v, want 25.0848", got)
	}
	if got := CropAndPelletSizeREP("pollen", 62.4, sp); !almostEqual(got, 0.00752544) {
		t.Errorf("pollen pellets 62.4 mg = %v, want 0.00752544", got)
	}
	expectPanic(t, "weight 0", func() { CropAndPelletSizeREP("nectar", 0, sp) })
	expectPanic(t, "unknown forage (result 0)", func() { CropAndPelletSizeREP("water", 785, sp) })
}

func TestProboscisLengthREP(t *testing.T) {
	sp := terrestris(t)
	// [PROB] finding 34: 6.9 mm at 62.4 mg (lightest worker), 11.1 mm at 980 mg (heaviest queen)
	cases := map[float64]float64{62.4: 6.9, 980: 11.1, (62.4 + 980) / 2: 9.0}
	for weight, want := range cases {
		if got := ProboscisLengthREP(weight, sp); !almostEqual(got, want) {
			t.Errorf("ProboscisLengthREP(%v) = %v, want %v", weight, got, want)
		}
	}
	expectPanic(t, "weight below the lightest worker", func() { ProboscisLengthREP(50, sp) })
	expectPanic(t, "weight above the heaviest queen", func() { ProboscisLengthREP(1000, sp) })
}

func TestWintermortalityProbREP(t *testing.T) {
	sp := terrestris(t) // queen weights 590 to 980 mg
	// [WMOR] 0.64 / (1 + e^(-22 * (relative weight - 0.32)))
	cases := map[float64]float64{
		590 + 0.32*390: 0.32,                               // relative weight 0.32: half of 0.64
		980:            0.64 / (1 + math.Exp(-22*0.68)),    // heaviest: 0.6399998
		590:            0.64 / (1 + math.Exp(-22*(-0.32))), // lightest: 0.00056
	}
	for weight, want := range cases {
		if got := WintermortalityProbREP(weight, sp); !almostEqual(got, want) {
			t.Errorf("WintermortalityProbREP(%v) = %v, want %v", weight, got, want)
		}
	}
	if got := WintermortalityProbREP(980, sp); got > 0.64 {
		t.Errorf("survival must stay below 0.64 (finding 33), got %v", got)
	}
	expectPanic(t, "weight above the maximum", func() { WintermortalityProbREP(981, sp) })
}
