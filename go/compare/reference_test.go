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
	referenceEnv = "COMPARE_REFERENCE_BIN" // program of the reference commit (exact and statistical comparison)
	currentEnv   = "COMPARE_CURRENT_BIN"   // program of the current code
	baseEnv      = "COMPARE_BASE_BIN"      // program of the base commit (speed)
	threadsEnv   = "COMPARE_THREADS"       // runs at the same time, default: number of CPUs
	reportEnv    = "COMPARE_REPORT"        // Markdown file the results are appended to, e.g. $GITHUB_STEP_SUMMARY
)

const (
	seeds = 100             // runs per version of the exact and the statistical comparison, seeds 1 to 100
	days  = 3 * DaysPerYear // days per run (1095)

	speedSeeds = 10 // runs per version of the speed comparison, one after the other
	speedDays  = DaysPerYear
)

// TestExactAgainstReference runs the reference and the current code with the same seeds and requires the same
// value on every day in every column both have (phase 1, see README.md).
func TestExactAgainstReference(t *testing.T) {
	ref, cur := referenceRuns(t), currentRuns(t)
	common, removed, added := columns(t, ref, cur)

	failures := []string{}
	for i := range ref {
		if err := Exact(ref[i].Table, cur[i].Table, common); err != nil {
			failures = append(failures, fmt.Sprintf("seed %d: %v", ref[i].Seed, err))
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Exact comparison with the reference\n\n")
	fmt.Fprintf(&b, "%d seeds, %d days, %d columns compared value by value.\n\n", seeds, days, len(common)-1)
	describeColumns(&b, removed, added)
	if len(failures) == 0 {
		fmt.Fprintf(&b, "**Result:** all %d runs are equal to the reference.\n\n", len(ref))
	} else {
		fmt.Fprintf(&b, "**Result:** %d of %d runs differ:\n\n", len(failures), len(ref))
		for _, f := range failures[:min(10, len(failures))] {
			fmt.Fprintf(&b, "- %s\n", f)
		}
		b.WriteString("\n")
	}
	report(t, b.String())
	for _, f := range failures {
		t.Error(f)
	}
}

// TestStatisticalAgainstReference compares the distributions over the seeds (phases 2 and 3, see README.md):
// the daily median of the current code must lie within the 5 to 95 % band of the reference, and no yearly summary
// may differ in a two-sample Kolmogorov-Smirnov test at the Bonferroni-corrected level.
func TestStatisticalAgainstReference(t *testing.T) {
	ref, cur := referenceRuns(t), currentRuns(t)
	common, removed, added := columns(t, ref, cur)
	common = slices.DeleteFunc(common, func(c string) bool { return c == "t" || c == "ticks" })

	checks, violations := Band(ref, cur, common)
	tests := CompareSummaries(ref, cur, common)
	alpha := FamilyAlpha / float64(max(1, len(tests)))
	significant := slices.DeleteFunc(slices.Clone(tests), func(s SummaryTest) bool { return s.P >= alpha })

	var b strings.Builder
	fmt.Fprintf(&b, "## Statistical comparison with the reference\n\n")
	fmt.Fprintf(&b, "%d seeds per version, %d days, %d columns.\n\n", seeds, days, len(common))
	describeColumns(&b, removed, added)
	fmt.Fprintf(&b, "**Daily band:** %d of %d days and columns outside the reference 5 to 95 %% band.\n\n",
		len(violations), checks)
	for _, v := range violations[:min(10, len(violations))] {
		fmt.Fprintf(&b, "- day %d, %s: median %v, band [%v, %v] (%d and %d runs)\n", v.Day, v.Column, v.Median,
			v.Q05, v.Q95, v.Ref, v.Actual)
	}
	fmt.Fprintf(&b, "\n**Yearly summaries:** %d Kolmogorov-Smirnov tests, %d with p < %.2g (%.2g / %d).\n\n",
		len(tests), len(significant), alpha, FamilyAlpha, len(tests))
	if len(tests) > 0 {
		b.WriteString("| Summary | D | p |\n| --- | --- | --- |\n")
		for _, s := range tests[:min(5, len(tests))] {
			fmt.Fprintf(&b, "| %s | %.3f | %.3g |\n", s.Summary, s.D, s.P)
		}
		b.WriteString("\n")
	}
	report(t, b.String())

	for _, v := range violations {
		t.Errorf("day %d, %s: median %v outside [%v, %v]", v.Day, v.Column, v.Median, v.Q05, v.Q95)
	}
	for _, s := range significant {
		t.Errorf("%s: D = %.3f, p = %.3g < %.2g", s.Summary, s.D, s.P, alpha)
	}
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

	var b strings.Builder
	fmt.Fprintf(&b, "## Speed against the base commit\n\n")
	fmt.Fprintf(&b, "%d runs of %d days per version, one after the other, median wall time per run (incl. "+
		"reading the input files).\n\n", speedSeeds, speedDays)
	b.WriteString("| Version | Median per run |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| base | %v |\n| current | %v |\n\n", baseMedian.Round(time.Millisecond),
		curMedian.Round(time.Millisecond))
	fmt.Fprintf(&b, "**Speed-up:** %.2f\n\n", float64(baseMedian)/float64(curMedian))
	report(t, b.String())
}

// runCache holds the runs of a program, so that the exact and the statistical comparison use the same runs.
var runCache = struct {
	sync.Mutex
	runs map[string][]Run
}{runs: map[string][]Run{}}

func referenceRuns(t *testing.T) []Run { return cachedRuns(t, referenceEnv) }
func currentRuns(t *testing.T) []Run   { return cachedRuns(t, currentEnv) }

// cachedRuns returns the runs of the program in the environment variable env, with the seeds 1 to seeds.
func cachedRuns(t *testing.T, env string) []Run {
	t.Helper()
	bin := os.Getenv(env)
	if bin == "" || os.Getenv(currentEnv) == "" {
		t.Skipf("set %s and %s (see go/compare/README.md)", referenceEnv, currentEnv)
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

// columns returns the columns of both versions. A column of the reference missing in the current code is an
// error: the column names stay the same in all steps (obs.DailyOutputExtended).
func columns(t *testing.T, ref, cur []Run) (common, removed, added []string) {
	t.Helper()
	common, removed, added = Columns(ref[0].Table, cur[0].Table)
	if len(removed) > 0 {
		t.Errorf("columns of the reference missing in the current code: %v", removed)
	}
	return common, removed, added
}

func describeColumns(b *strings.Builder, removed, added []string) {
	if len(added) > 0 {
		fmt.Fprintf(b, "Only in the current code (not compared): %s.\n\n", strings.Join(added, ", "))
	}
	if len(removed) > 0 {
		fmt.Fprintf(b, "Missing in the current code: %s.\n\n", strings.Join(removed, ", "))
	}
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

// report appends the section to the report file, if one is set, and logs it.
func report(t *testing.T, section string) {
	t.Helper()
	t.Log("\n" + section)
	path := os.Getenv(reportEnv)
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(section); err != nil {
		t.Fatal(err)
	}
}
