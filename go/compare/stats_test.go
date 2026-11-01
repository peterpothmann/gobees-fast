package compare

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestQuantile(t *testing.T) {
	v := []float64{4, 1, 3, 2, 5} // sorted 1 to 5
	for q, want := range map[float64]float64{0: 1, 0.05: 1.2, 0.5: 3, 0.95: 4.8, 1: 5} {
		if got := Quantile(v, q); math.Abs(got-want) > 1e-12 {
			t.Errorf("Quantile(%v): got %v, want %v", q, got, want)
		}
	}
	if got := Quantile([]float64{7}, 0.95); got != 7 {
		t.Errorf("one value: got %v", got)
	}
}

func TestKolmogorovQ(t *testing.T) {
	// Worked example: Q_KS(lambda) = 2 * (exp(-2 lambda^2) - exp(-8 lambda^2) + ...).
	// lambda = 1: 2 * (0.135335 - 0.000335 + 0.0000000152) = 0.270000.
	// lambda = 1.36 is the 5 % point of the Kolmogorov distribution.
	for lambda, want := range map[float64]float64{0.1: 1, 1: 0.270000, 1.36: 0.0494, 2: 0.000671} {
		if got := kolmogorovQ(lambda); math.Abs(got-want) > 1e-4 {
			t.Errorf("Q_KS(%v): got %v, want %v", lambda, got, want)
		}
	}
}

func TestKolmogorovSmirnov(t *testing.T) {
	a := []float64{1, 2, 3, 4, 5}
	if d, p := KolmogorovSmirnov(a, []float64{5, 4, 3, 2, 1}); d != 0 || p != 1 {
		t.Errorf("same values: got D %v, p %v", d, p)
	}
	if d, p := KolmogorovSmirnov([]float64{0, 0, 0}, []float64{0, 0}); d != 0 || p != 1 {
		t.Errorf("constant values: got D %v, p %v", d, p)
	}

	// Worked example: 100 values 1..100 against 100 values 21..120: D = 0.2.
	// lambda = (sqrt(50) + 0.12 + 0.11/sqrt(50)) * 0.2 = 1.44133, Q_KS = 2 * exp(-2 * 1.44133^2) = 0.03138.
	x, y := make([]float64, 100), make([]float64, 100)
	for i := range x {
		x[i], y[i] = float64(i+1), float64(i+21)
	}
	d, p := KolmogorovSmirnov(x, y)
	if math.Abs(d-0.2) > 1e-12 || math.Abs(p-0.03138) > 1e-4 {
		t.Errorf("shifted values: got D %v, p %v, want 0.2 and 0.03138", d, p)
	}

	// ties: 50 zeros and 50 ones against 100 ones, D = 0.5
	z := make([]float64, 100)
	ones := make([]float64, 100)
	for i := range z {
		z[i], ones[i] = float64(i%2), 1
	}
	if d, _ := KolmogorovSmirnov(z, ones); d != 0.5 {
		t.Errorf("ties: got D %v, want 0.5", d)
	}
}

// TestKolmogorovSmirnovFalseAlarms draws both samples from the same distribution. About 5 % of the tests may
// have p < 0.05, the asymptotic p-value with 100 values each is slightly conservative.
func TestKolmogorovSmirnovFalseAlarms(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	alarms := 0
	for range 1000 {
		a, b := make([]float64, 100), make([]float64, 100)
		for i := range a {
			a[i], b[i] = rng.NormFloat64(), rng.NormFloat64()
		}
		if _, p := KolmogorovSmirnov(a, b); p < 0.05 {
			alarms++
		}
	}
	if alarms < 20 || alarms > 70 {
		t.Errorf("%d of 1000 tests with p < 0.05, want about 50", alarms)
	}
}

// runs returns one run per value list, with a column ticks and a column A holding the values.
func runs(values ...[]float64) []Run {
	result := []Run{}
	for _, v := range values {
		tab := Table{Header: []string{"ticks", "A"}}
		for d, x := range v {
			tab.Rows = append(tab.Rows, []float64{float64(d + 1), x})
		}
		result = append(result, Run{Table: tab})
	}
	return result
}

func TestBand(t *testing.T) {
	// reference band on day 1: values 0..20 (5 % = 1, 95 % = 19); on day 2 only the run with value 20 is left
	ref := [][]float64{}
	for i := range 21 {
		ref = append(ref, []float64{float64(i)})
	}
	ref[20] = []float64{20, 5}

	checks, violations := Band(runs(ref...), runs([]float64{10, 5}, []float64{10, 5}), []string{"A"})
	if checks != 2 || len(violations) != 0 {
		t.Errorf("inside the band: %d checks, violations %v", checks, violations)
	}

	_, violations = Band(runs(ref...), runs([]float64{19.5}, []float64{19.5}), []string{"A"})
	if len(violations) != 1 || violations[0].Day != 1 || violations[0].Median != 19.5 || violations[0].Q95 != 19 {
		t.Errorf("median above the band: got %v", violations)
	}
}

func TestSummaries(t *testing.T) {
	// 1.5 years: day d has value d
	v := make([]float64, DaysPerYear+DaysPerYear/2)
	for i := range v {
		v[i] = float64(i + 1)
	}
	s := Summaries(runs(v)[0].Table, []string{"A"})
	want := map[Summary]float64{
		{"A", 1, "max"}: 365, {"A", 1, "mean"}: 183,
		{"A", 2, "max"}: 547, {"A", 2, "mean"}: 456.5, // days 366 to 547
	}
	if len(s) != len(want) {
		t.Errorf("got %d summaries, want %d: %v", len(s), len(want), s)
	}
	for k, w := range want {
		if s[k] != w {
			t.Errorf("%v: got %v, want %v", k, s[k], w)
		}
	}
}

func TestCompareSummaries(t *testing.T) {
	ref, same, shifted := [][]float64{}, [][]float64{}, [][]float64{}
	for i := range 50 {
		ref = append(ref, []float64{float64(i)})
		same = append(same, []float64{float64(49 - i)})
		shifted = append(shifted, []float64{float64(i + 25)})
	}
	tests := CompareSummaries(runs(ref...), runs(same...), []string{"A"})
	if len(tests) != 2 || tests[0].P != 1 || tests[1].P != 1 {
		t.Errorf("same values: got %v", tests)
	}
	tests = CompareSummaries(runs(ref...), runs(shifted...), []string{"A"})
	if len(tests) != 2 || tests[0].D != 0.5 || tests[0].P > 0.001 {
		t.Errorf("shifted by half the range: got %v", tests)
	}
	if tests := CompareSummaries(runs(ref[:9]...), runs(same[:9]...), []string{"A"}); len(tests) != 0 {
		t.Errorf("fewer than 10 runs are not tested: got %v", tests)
	}
}
