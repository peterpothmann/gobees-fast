package sys

import (
	"math"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
)

func TestMaxWeightGainToday(t *testing.T) {
	// worked example: 1.5 * 1.88^k, workers capped at 249.5 mg on the ninth feeding, queens at 980 mg
	// on the eleventh.
	sp := terrestris(t)
	for _, tc := range []struct {
		caste  caste.Caste
		capped int // feeding on which the cap binds
		cap    float64
	}{{caste.Worker, 9, 249.5}, {caste.Male, 9, 249.5}, {caste.Undefined, 9, 249.5}, {caste.Queen, 11, 980}} {
		b := comp.Bee{Caste: tc.caste, Weight: sp.EggWeight}
		for k := 1; k <= 12; k++ {
			b.Weight += maxWeightGainToday(&b, sp)
			want := math.Min(1.5*math.Pow(1.88, float64(k)), tc.cap)
			if !almostEqual(b.Weight, want) {
				t.Errorf("%v after %d feedings: %v mg, want %v", tc.caste, k, b.Weight, want)
			}
			if k == tc.capped-1 && b.Weight >= tc.cap {
				t.Errorf("%v reached the cap before feeding %d", tc.caste, tc.capped)
			}
		}
	}
}

func TestLarvaeNeeds(t *testing.T) {
	// worked example: 12 larvae of 10 mg: 0.1056 g pollen, 0.65472 kJ
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	other := bw.addColony(nil)
	bw.addBrood(col, stage.Larva, caste.Undefined, 12, 6, 10, 0)
	bw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0)  // eggs need nothing
	bw.addBrood(col, stage.Pupa, caste.Worker, 5, 15, 200, 0)    // pupae need nothing
	bw.addBrood(other, stage.Larva, caste.Worker, 1, 10, 100, 0) // other colony: 88 mg
	bw.need.Update(bw.w)
	_, s, _ := bw.colony(col)
	if !almostEqual(s.LarvaePollenNeedToday, 0.1056) || !almostEqual(s.EnergyNeedToday, 0.1056*6.2) {
		t.Errorf("pollen need %v g, energy need %v kJ; want 0.1056, 0.65472", s.LarvaePollenNeedToday, s.EnergyNeedToday)
	}
	_, s, _ = bw.colony(other)
	if !almostEqual(s.LarvaePollenNeedToday, 0.088) {
		t.Errorf("other colony: pollen need %v g, want 0.088", s.LarvaePollenNeedToday)
	}
}
