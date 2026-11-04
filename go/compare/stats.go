package compare

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// DaysPerYear is the length of a model year in ticks.
const DaysPerYear = 365

// FamilyAlpha is the significance level of all two-sample tests of one comparison together. Each single test uses
// FamilyAlpha divided by the number of tests (Bonferroni correction).
const FamilyAlpha = 0.05

// Quantile returns the q-quantile of the values, linearly interpolated between the sorted values (type 7 of
// Hyndman and Fan 1996, the default of R). The values must not be empty.
func Quantile(values []float64, q float64) float64 {
	v := slices.Clone(values)
	slices.Sort(v)
	h := q * float64(len(v)-1)
	lo := int(math.Floor(h))
	if lo+1 >= len(v) {
		return v[len(v)-1]
	}
	return v[lo] + (h-float64(lo))*(v[lo+1]-v[lo])
}

// BandViolation is a day on which the median of the current version lies outside the 5 to 95 % band of the
// reference.
type BandViolation struct {
	Day         int
	Column      string
	Median      float64
	Q05, Q95    float64
	Ref, Actual int // number of runs that reached this day
}

// Band checks for each column and day that the median over the current runs lies within the 5 to 95 % band
// of the reference runs. Runs that ended early (extinction with StopExtinct?) count only for the days they
// reached. It returns the number of checks and the violations.
func Band(reference, current []Run, columns []string) (checks int, violations []BandViolation) {
	days := 0
	for _, r := range slices.Concat(reference, current) {
		days = max(days, len(r.Table.Rows))
	}
	for _, c := range columns {
		for d := range days {
			ref, act := dayValues(reference, c, d), dayValues(current, c, d)
			if len(ref) == 0 || len(act) == 0 {
				continue
			}
			checks++
			median, q05, q95 := Quantile(act, 0.5), Quantile(ref, 0.05), Quantile(ref, 0.95)
			if median < q05 || median > q95 {
				violations = append(violations, BandViolation{d + 1, c, median, q05, q95, len(ref), len(act)})
			}
		}
	}
	return checks, violations
}

// dayValues returns the values of the column on the day with index d (0 = first day) of all runs that reached it.
func dayValues(runs []Run, column string, d int) []float64 {
	values := []float64{}
	for _, r := range runs {
		if d < len(r.Table.Rows) {
			values = append(values, r.Table.Rows[d][r.Table.Column(column)])
		}
	}
	return values
}

// Summary identifies a yearly summary of a column: its maximum or mean over the days of a year.
type Summary struct {
	Column string
	Year   int    // 1, 2, ...
	Kind   string // "max" or "mean"
}

func (s Summary) String() string {
	return fmt.Sprintf("%s, year %d, %s", s.Column, s.Year, s.Kind)
}

// Summaries returns the maximum and the mean of each column in each year of the run. A year that the run did not
// reach is missing. A year that the run reached only in part is summarised over the days it reached.
func Summaries(t Table, columns []string) map[Summary]float64 {
	result := map[Summary]float64{}
	for _, c := range columns {
		i := t.Column(c)
		for start := 0; start < len(t.Rows); start += DaysPerYear {
			year := start/DaysPerYear + 1
			days := t.Rows[start:min(start+DaysPerYear, len(t.Rows))]
			maxValue, sum := math.Inf(-1), 0.0
			for _, row := range days {
				maxValue = max(maxValue, row[i])
				sum += row[i]
			}
			result[Summary{c, year, "max"}] = maxValue
			result[Summary{c, year, "mean"}] = sum / float64(len(days))
		}
	}
	return result
}

// SummaryTest is the two-sample test of one yearly summary.
type SummaryTest struct {
	Summary
	D, P        float64
	Ref, Actual int // number of runs
}

// CompareSummaries compares the yearly summaries of the reference and the current runs with a two-sample
// Kolmogorov-Smirnov test each. A summary is tested if at least 10 runs of each version have it. It returns all
// tests sorted by p, smallest first.
func CompareSummaries(reference, current []Run, columns []string) []SummaryTest {
	collect := func(runs []Run) map[Summary][]float64 {
		values := map[Summary][]float64{}
		for _, r := range runs {
			for s, v := range Summaries(r.Table, columns) {
				values[s] = append(values[s], v)
			}
		}
		return values
	}
	ref, act := collect(reference), collect(current)
	tests := []SummaryTest{}
	for s, a := range ref {
		b := act[s]
		if len(a) < 10 || len(b) < 10 {
			continue
		}
		d, p := KolmogorovSmirnov(a, b)
		tests = append(tests, SummaryTest{s, d, p, len(a), len(b)})
	}
	sortTests(tests)
	return tests
}

// sortTests sorts tests by p, smallest first, then by D, largest first.
func sortTests(tests []SummaryTest) {
	slices.SortFunc(tests, func(x, y SummaryTest) int {
		if x.P != y.P {
			return compareFloat(x.P, y.P)
		}
		if x.D != y.D {
			return compareFloat(y.D, x.D)
		}
		return strings.Compare(x.String(), y.String())
	})
}

func compareFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// KolmogorovSmirnov returns the two-sample Kolmogorov-Smirnov statistic D (the largest distance between the two
// empirical distribution functions) and its p-value. The p-value uses the asymptotic Kolmogorov distribution with
// the correction of Stephens (1970), as in Numerical Recipes (Press et al. 2007, section 14.3.3). With ties, which
// are common in count data, the test is conservative: the p-value is too large rather than too small.
func KolmogorovSmirnov(a, b []float64) (d, p float64) {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	n, m := float64(len(x)), float64(len(y))
	i, j := 0, 0
	for i < len(x) && j < len(y) {
		v := min(x[i], y[j])
		for i < len(x) && x[i] == v {
			i++
		}
		for j < len(y) && y[j] == v {
			j++
		}
		d = max(d, math.Abs(float64(i)/n-float64(j)/m))
	}
	ne := math.Sqrt(n * m / (n + m))
	return d, kolmogorovQ((ne + 0.12 + 0.11/ne) * d)
}

// kolmogorovQ returns Q_KS(lambda) = 2 * sum_{k>=1} (-1)^(k-1) * exp(-2 k^2 lambda^2), the probability that the
// Kolmogorov statistic exceeds lambda.
func kolmogorovQ(lambda float64) float64 {
	if lambda < 0.3 { // Q_KS(0.3) = 0.99999, the series converges slowly below
		return 1
	}
	sum, sign := 0.0, 1.0
	for k := 1.0; k <= 100; k++ {
		term := sign * math.Exp(-2*k*k*lambda*lambda)
		sum += term
		if math.Abs(term) < 1e-12 {
			break
		}
		sign = -sign
	}
	return min(1, max(0, 2*sum))
}
