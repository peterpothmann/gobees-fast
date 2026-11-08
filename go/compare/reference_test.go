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
	referenceEnv       = "COMPARE_REFERENCE_BIN"    // program of the reference commit
	currentEnv         = "COMPARE_CURRENT_BIN"      // program of the current code
	baseEnv            = "COMPARE_BASE_BIN"         // program of the base commit (gobees)
	referenceCommitEnv = "COMPARE_REFERENCE_COMMIT" // REFERENCE_COMMIT of commits.env, shown in the comment
	baseCommitEnv      = "COMPARE_BASE_COMMIT"      // BASE_COMMIT of commits.env, shown in the comment
	phaseEnv           = "COMPARE_PHASE"            // PHASE of commits.env, default 1
	threadsEnv         = "COMPARE_THREADS"          // runs at the same time, default: number of CPUs
	commentEnv         = "COMPARE_COMMENT"          // file the pull request comment is written to, none if empty
	reportDirEnv       = "COMPARE_REPORT_DIR"       // folder of the NetLogo report and its plots, none if empty
	reportURLEnv       = "COMPARE_REPORT_URL"       // link to the NetLogo report in the comment, may be empty
	runURLEnv          = "COMPARE_RUN_URL"          // link to the workflow run in the comment, may be empty
	removedEnv         = "COMPARE_REMOVED_COLUMNS"  // REMOVED_COLUMNS of commits.env, space-separated
	referenceInputEnv  = "COMPARE_REFERENCE_INPUT"  // REFERENCE_INPUT of commits.env: netlogo (default) or yaml
)

const (
	seeds = 100 // runs per version and scenario, seeds 1 to 100

	speedSeeds = 10 // runs per version of the speed comparison, one after the other
	speedDays  = DaysPerYear
)

// results collects the results of the tests. TestMain writes the comment and the report from them.
var results struct {
	sync.Mutex
	ran       bool // a comparison ran, i.e. the programs were set
	scenarios []ScenarioResult
	speed     *Speed
}

// TestMain runs the tests and then writes the comment (COMPARE_COMMENT) and the NetLogo report (COMPARE_REPORT_DIR),
// also if a comparison failed.
func TestMain(m *testing.M) {
	code := m.Run()
	if err := writeResults(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}

func writeResults() error {
	if !results.ran {
		return nil
	}
	if path := os.Getenv(commentEnv); path != "" {
		p, _ := phase()
		info := CommentInfo{Phase: p, Reference: os.Getenv(referenceCommitEnv), Base: os.Getenv(baseCommitEnv),
			Seeds: seeds, ReportURL: os.Getenv(reportURLEnv), RunURL: os.Getenv(runURLEnv)}
		if err := os.WriteFile(path, []byte(Comment(results.scenarios, results.speed, info)), 0o644); err != nil {
			return err
		}
	}
	if dir := os.Getenv(reportDirEnv); dir != "" {
		return WriteReport(dir, results.scenarios)
	}
	return nil
}

// TestScenarios runs every scenario with the current code, the reference and the base commit and compares the
// current code with each of them and with NetLogo (see README.md). Each scenario is a subtest that fails if a
// comparison fails.
func TestScenarios(t *testing.T) {
	cur, ref, base := os.Getenv(currentEnv), os.Getenv(referenceEnv), os.Getenv(baseEnv)
	if cur == "" || ref == "" || base == "" {
		t.Skipf("set %s, %s and %s (see go/compare/README.md)", currentEnv, referenceEnv, baseEnv)
	}
	p, err := phase()
	if err != nil {
		t.Fatal(err)
	}
	results.Lock()
	results.ran = true
	results.Unlock()

	dir := inputDir(t)
	for _, s := range Scenarios {
		t.Run(s.Name, func(t *testing.T) {
			r := runScenario(t, s, dir, cur, ref, base, p)
			results.Lock()
			results.scenarios = append(results.scenarios, r)
			results.Unlock()
			t.Logf("previous step: %s | gobees: %s | NetLogo: %s", cellText(r.Changed), cellText(r.Drift),
				cellText(r.NetLogo))
			for _, c := range []struct {
				against string
				cell    Cell
			}{{"the previous step", r.Changed}, {"gobees", r.Drift}, {"NetLogo", r.NetLogo}} {
				if !c.cell.Passed {
					t.Errorf("against %s: %s\n%s", c.against, c.cell.Text, c.cell.Details)
				}
			}
		})
	}
}

// runScenario runs the scenario with the programs and compares the current code with the others.
func runScenario(t *testing.T, s Scenario, dir, cur, ref, base string, phase int) ScenarioResult {
	t.Helper()
	run := func(bin string, in Input) []Run {
		runs, err := RunPrograms(bin, in, in.Args(dir, &s), Seeds(seeds), s.Days, threads(t), true)
		if err != nil {
			t.Fatal(err)
		}
		return runs
	}
	// gobees (base) reads the NetLogo files, the current code its run files, the reference its REFERENCE_INPUT.
	current, reference := run(cur, YAMLInput), run(ref, referenceInput(t))
	baseRuns := reference // run_compare.sh passes the same program for the same commit
	if base != ref {
		baseRuns = run(base, NetLogoInput)
	}

	r := ScenarioResult{Scenario: s, Removed: strings.Fields(os.Getenv(removedEnv))}
	common, missing, added := Columns(reference[0].Table, current[0].Table)
	missing = notRemoved(missing, r.Removed)
	r.ReferenceColumns, r.NotInReference = dataColumns(common), added
	switch {
	case len(missing) > 0:
		r.Changed = MissingColumnsCell(missing)
	case phase == 1:
		r.Changed = ExactCell(reference, current, common)
	default:
		r.Changed = StatisticalCell(reference, current, common)
	}

	common, missing, added = Columns(baseRuns[0].Table, current[0].Table)
	missing = notRemoved(missing, r.Removed)
	r.BaseColumns, r.NotInBase = dataColumns(common), added
	if len(missing) > 0 {
		r.Drift = MissingColumnsCell(missing)
	} else {
		r.Drift = StatisticalCell(baseRuns, current, common)
	}

	var err error
	if r.NetLogoResult, err = CompareNetLogo(s, current, baseRuns); err != nil {
		r.NetLogo = Cell{Ran: true, Text: err.Error()}
	} else {
		r.NetLogo = NetLogoCell(r.NetLogoResult)
	}
	return r
}

// notRemoved returns the missing columns that are not removed on purpose (REMOVED_COLUMNS of commits.env). A column
// is removed only together with its model branch (obs.DailyOutputExtended), any other missing column is an error.
func notRemoved(missing, removed []string) []string {
	return slices.DeleteFunc(missing, func(c string) bool { return slices.Contains(removed, c) })
}

// TestSpeedAgainstBase times the base commit and the current code on the same machine, one run after the other and
// alternating between them, so that a change of the load of the machine affects both alike. On the example farm and
// without output. It reports the times and fails only if a program fails.
func TestSpeedAgainstBase(t *testing.T) {
	base, cur := os.Getenv(baseEnv), os.Getenv(currentEnv)
	if base == "" || cur == "" {
		t.Skipf("set %s and %s (see go/compare/README.md)", baseEnv, currentEnv)
	}
	results.Lock()
	results.ran = true
	results.Unlock()

	dir := inputDir(t)
	var baseRuns, curRuns []Run
	for _, seed := range Seeds(speedSeeds) {
		b, err := RunProgram(base, NetLogoInput, NetLogoInput.Args(dir, nil), seed, speedDays, false)
		if err != nil {
			t.Fatal(err)
		}
		c, err := RunProgram(cur, YAMLInput, YAMLInput.Args(dir, nil), seed, speedDays, false)
		if err != nil {
			t.Fatal(err)
		}
		baseRuns, curRuns = append(baseRuns, b), append(curRuns, c)
	}
	speed := Speed{Base: medianDuration(baseRuns), Current: medianDuration(curRuns), Days: speedDays}
	t.Logf("median per run: base %v, current %v", speed.Base, speed.Current)
	results.Lock()
	results.speed = &speed
	results.Unlock()
}

// phase returns the phase of COMPARE_PHASE, 1 if it is not set.
func phase() (int, error) {
	s := os.Getenv(phaseEnv)
	if s == "" {
		return 1, nil
	}
	p, err := strconv.Atoi(s)
	if err != nil || p < 1 || p > 3 {
		return 1, fmt.Errorf("%s must be 1, 2 or 3, got %q", phaseEnv, s)
	}
	return p, nil
}

// referenceInput returns the input format of the reference program (COMPARE_REFERENCE_INPUT).
func referenceInput(t *testing.T) Input {
	t.Helper()
	in, err := ParseInput(os.Getenv(referenceInputEnv))
	if err != nil {
		t.Fatal(err)
	}
	return in
}

// inputDir returns a temporary folder with the input files of WriteInputFiles.
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
