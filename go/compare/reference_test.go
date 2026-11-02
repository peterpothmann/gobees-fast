package compare

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Environment of the comparisons, set by run_compare.sh.
const (
	referenceEnv       = "COMPARE_REFERENCE_BIN"    // program of the reference commit (exact and statistical comparison)
	currentEnv         = "COMPARE_CURRENT_BIN"      // program of the current code
	baseEnv            = "COMPARE_BASE_BIN"         // program of the base commit (drift and speed)
	referenceCommitEnv = "COMPARE_REFERENCE_COMMIT" // REFERENCE_COMMIT of commits.env, shown in the results
	baseCommitEnv      = "COMPARE_BASE_COMMIT"      // BASE_COMMIT of commits.env, shown in the results
	threadsEnv         = "COMPARE_THREADS"          // runs at the same time, default: number of CPUs
	checksEnv          = "COMPARE_CHECKS"           // folder the results are written to (WriteCheck), none if empty
)

const (
	seeds = 100             // runs per version of the exact, the statistical and the drift comparison, seeds 1 to 100
	days  = 3 * DaysPerYear // days per run (1095)

	speedSeeds = 10 // runs per version of the speed comparison, one after the other
	speedDays  = DaysPerYear
)

// TestExactAgainstReference runs the reference and the current code with the same seeds and requires the same
// value on every day in every column both have (phase 1, see README.md).
func TestExactAgainstReference(t *testing.T) {
	ref, cur := referenceRuns(t), currentRuns(t)
	common, missing, added := columns(t, ref, cur)

	failures := []string{}
	for i := range ref {
		if err := Exact(ref[i].Table, cur[i].Table, common); err != nil {
			failures = append(failures, fmt.Sprintf("seed %d: %v", ref[i].Seed, err))
		}
	}

	reference := "the reference " + commit(referenceCommitEnv)
	c := Question("1-changed")
	c.Scope = runScope(common, added, "the reference")
	if len(failures) == 0 {
		c.Status = "✅"
		c.Result = fmt.Sprintf("Equal to %s: %d of %d runs identical in all %d columns", reference, len(ref),
			len(ref), len(dataColumns(common)))
	} else {
		c.Status = "❌"
		c.Result = fmt.Sprintf("%d of %d runs differ from %s", len(failures), len(ref), reference)
		var b strings.Builder
		fmt.Fprintf(&b, "The first difference of each run (at most 10 runs):\n\n")
		for _, f := range failures[:min(10, len(failures))] {
			fmt.Fprintf(&b, "- %s\n", f)
		}
		c.Details = b.String()
	}
	markMissing(&c, missing)
	writeCheck(t, c)
	for _, f := range failures {
		t.Error(f)
	}
}

// TestStatisticalAgainstReference compares the distributions over the seeds with those of the reference (phases 2
// and 3, see README.md): the daily median of the current code must lie within the 5 to 95 % band of the reference,
// and no yearly summary may differ in a two-sample Kolmogorov-Smirnov test at the Bonferroni-corrected level.
func TestStatisticalAgainstReference(t *testing.T) {
	ref, cur := referenceRuns(t), currentRuns(t)
	common, missing, added := columns(t, ref, cur)
	c := statistical(t, Question("1-changed"), ref, cur, common, "the reference "+commit(referenceCommitEnv))
	c.Scope = runScope(common, added, "the reference")
	markMissing(&c, missing)
	writeCheck(t, c)
}

// TestDriftAgainstBase compares the distributions over the seeds with those of the base commit (gobees), like
// TestStatisticalAgainstReference. The reference moves with the steps, the base stays fixed: differences too small
// to show in a single step show here once they add up.
func TestDriftAgainstBase(t *testing.T) {
	base, cur := baseRuns(t), currentRuns(t)
	common, missing, added := columns(t, base, cur)
	c := statistical(t, Question("2-drift"), base, cur, common, "gobees "+commit(baseCommitEnv))
	c.Scope = fmt.Sprintf("The same seeds and days, against gobees, %s", columnScope(common, added, "gobees"))
	markMissing(&c, missing)
	writeCheck(t, c)
}

// statistical compares the runs cur with the runs ref on the given columns (see TestStatisticalAgainstReference),
// fills the result and the details of c and fails t on every difference. against names ref in the result.
func statistical(t *testing.T, c Check, ref, cur []Run, columns []string, against string) Check {
	t.Helper()
	columns = dataColumns(columns)
	checks, violations := Band(ref, cur, columns)
	tests := CompareSummaries(ref, cur, columns)
	alpha := FamilyAlpha / float64(max(1, len(tests)))
	significant := slices.DeleteFunc(slices.Clone(tests), func(s SummaryTest) bool { return s.P >= alpha })

	counts := fmt.Sprintf("%d of %d days outside the band, %d of %d Kolmogorov-Smirnov tests significant",
		len(violations), checks, len(significant), len(tests))
	if len(violations) == 0 && len(significant) == 0 {
		c.Status, c.Result = "✅", fmt.Sprintf("Statistically equal to %s: %s", against, counts)
	} else {
		c.Status, c.Result = "❌", fmt.Sprintf("Differs from %s: %s", against, counts)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "**Daily band:** on %d of %d days and columns, the median of the current code lies outside the "+
		"5 to 95 %% band of %s.\n\n", len(violations), checks, against)
	for _, v := range violations[:min(10, len(violations))] {
		fmt.Fprintf(&b, "- day %d, %s: median %v, band [%v, %v] (%d and %d runs)\n", v.Day, v.Column, v.Median,
			v.Q05, v.Q95, v.Ref, v.Actual)
	}
	if len(violations) > 0 {
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "**Yearly values:** %d Kolmogorov-Smirnov tests of the maximum and the mean of each column "+
		"and year, %d with p < %.2g (%.2g / %d).", len(tests), len(significant), alpha, FamilyAlpha, len(tests))
	if len(tests) > 0 {
		shown := min(5, len(tests))
		fmt.Fprintf(&b, " The %d with the smallest p:\n\n| Summary | D | p |\n| --- | --- | --- |\n", shown)
		for _, s := range tests[:shown] {
			fmt.Fprintf(&b, "| %s | %.3f | %.3g |\n", s.Summary, s.D, s.P)
		}
	}
	c.Details = b.String()

	for _, v := range violations {
		t.Errorf("day %d, %s: median %v outside [%v, %v]", v.Day, v.Column, v.Median, v.Q05, v.Q95)
	}
	for _, s := range significant {
		t.Errorf("%s: D = %.3f, p = %.3g < %.2g", s.Summary, s.D, s.P, alpha)
	}
	return c
}

// TestSpeedAgainstBase times the base commit and the current code on the same machine, one run after the other.
// It reports the times and fails only if a program fails.
func TestSpeedAgainstBase(t *testing.T) {
	base, cur := os.Getenv(baseEnv), os.Getenv(currentEnv)
	if base == "" || cur == "" {
		t.Skipf("set %s and %s (see go/compare/README.md)", baseEnv, currentEnv)
	}
	dir := inputDir(t)
	baseRuns, err := RunPrograms(base, dir, Seeds(speedSeeds), speedDays, 1)
	if err != nil {
		t.Fatal(err)
	}
	curRuns, err := RunPrograms(cur, dir, Seeds(speedSeeds), speedDays, 1)
	if err != nil {
		t.Fatal(err)
	}
	baseMedian, curMedian := medianDuration(baseRuns), medianDuration(curRuns)

	c := Question("4-speed")
	c.Result = fmt.Sprintf("%.2f× as fast as gobees %s (%v → %v per run of %d days)",
		float64(baseMedian)/float64(curMedian), commit(baseCommitEnv), baseMedian.Round(time.Millisecond),
		curMedian.Round(time.Millisecond), speedDays)
	c.Scope = fmt.Sprintf("Seeds 1 to %d, %d days per version, one run after the other, median wall time per run "+
		"(incl. reading the input files).", speedSeeds, speedDays)
	writeCheck(t, c)
}

// runScope describes the runs of the exact and the statistical comparison and their columns.
func runScope(common, added []string, other string) string {
	return fmt.Sprintf("Example farm, seeds 1 to %d, %d days (%d years) per version, %s", seeds, days,
		days/DaysPerYear, columnScope(common, added, other))
}

// columnScope describes the compared columns: other names the version the current code is compared with.
func columnScope(common, added []string, other string) string {
	s := fmt.Sprintf("the %d output columns both versions write.", len(dataColumns(common)))
	if len(added) > 0 {
		s += fmt.Sprintf(" Not compared: %d columns %s does not write (%s).", len(added), other,
			strings.Join(added, ", "))
	}
	return s
}

// markMissing marks c as failed if the current code lacks columns of the other version (see columns).
func markMissing(c *Check, missing []string) {
	if len(missing) > 0 {
		c.Status = "❌"
		c.Result += fmt.Sprintf("; columns missing in the current code: %s", strings.Join(missing, ", "))
	}
}

// commit returns the commit in the environment variable env as code, or an empty string.
func commit(env string) string {
	if v := os.Getenv(env); v != "" {
		return "`" + v + "`"
	}
	return ""
}

// writeCheck logs the check and writes it to the folder of COMPARE_CHECKS, if set.
func writeCheck(t *testing.T, c Check) {
	t.Helper()
	t.Logf("\n%s %s %s\n%s\n\n%s", c.Status, c.Question, c.Result, c.Scope, c.Details)
	dir := os.Getenv(checksEnv)
	if dir == "" {
		return
	}
	if err := WriteCheck(dir, c); err != nil {
		t.Fatal(err)
	}
}

// runCache holds the runs of a program, so that the comparisons use the same runs. run_compare.sh passes the same
// program as reference and base if both are the same commit, so that it runs only once.
var runCache = struct {
	sync.Mutex
	runs map[string][]Run
}{runs: map[string][]Run{}}

func referenceRuns(t *testing.T) []Run { return cachedRuns(t, referenceEnv) }
func baseRuns(t *testing.T) []Run      { return cachedRuns(t, baseEnv) }
func currentRuns(t *testing.T) []Run   { return cachedRuns(t, currentEnv) }

// cachedRuns returns the runs of the program in the environment variable env, with the seeds 1 to seeds.
func cachedRuns(t *testing.T, env string) []Run {
	t.Helper()
	bin := os.Getenv(env)
	if bin == "" || os.Getenv(currentEnv) == "" {
		t.Skipf("set %s and %s (see go/compare/README.md)", env, currentEnv)
	}
	runCache.Lock()
	defer runCache.Unlock()
	if runs, ok := runCache.runs[bin]; ok {
		return runs
	}
	runs, err := RunPrograms(bin, inputDir(t), Seeds(seeds), days, threads(t))
	if err != nil {
		t.Fatal(err)
	}
	runCache.runs[bin] = runs
	return runs
}

// columns returns the columns of both versions. A column of the other version missing in the current code is an
// error: the column names stay the same in all steps (obs.DailyOutputExtended).
func columns(t *testing.T, other, cur []Run) (common, missing, added []string) {
	t.Helper()
	common, missing, added = Columns(other[0].Table, cur[0].Table)
	if len(missing) > 0 {
		t.Errorf("columns missing in the current code: %v", missing)
	}
	return common, missing, added
}

// dataColumns returns the columns without the time columns.
func dataColumns(columns []string) []string {
	return slices.DeleteFunc(slices.Clone(columns), func(c string) bool { return c == "t" || c == "ticks" })
}

// inputDir returns a temporary folder with the input files of the example farm.
func inputDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gobees-compare-input")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	if err := WriteInputFiles(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func threads(t *testing.T) int {
	t.Helper()
	s := os.Getenv(threadsEnv)
	if s == "" {
		return runtime.NumCPU()
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		t.Fatalf("%s must be a positive number, got %q", threadsEnv, s)
	}
	return n
}

func medianDuration(runs []Run) time.Duration {
	d := make([]float64, len(runs))
	for i, r := range runs {
		d[i] = float64(r.Duration)
	}
	return time.Duration(Quantile(d, 0.5))
}
