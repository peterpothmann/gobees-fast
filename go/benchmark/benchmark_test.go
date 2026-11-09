package benchmark

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Environment of TestCompareWithNetLogo, set by the CI (.github/workflows/benchmark.yml).
const (
	tableEnv   = "BENCHMARK_NETLOGO_TABLE"   // BehaviorSpace table of run_netlogo.sh
	consoleEnv = "BENCHMARK_NETLOGO_CONSOLE" // console output of run_netlogo.sh
	threadsEnv = "BENCHMARK_THREADS"         // runs at the same time, as given to run_netlogo.sh
	reportEnv  = "BENCHMARK_REPORT"          // path of the Markdown report
	javaEnv    = "BENCHMARK_JAVA"            // Java version that ran NetLogo, for the report
	// the command line programs of the base commit (gobees) and of the current code (gobees-fast), and the base
	// commit for the report
	baseBinEnv    = "BENCHMARK_BASE_BIN"
	currentBinEnv = "BENCHMARK_CURRENT_BIN"
	baseCommitEnv = "BENCHMARK_BASE_COMMIT"
	// optional, all four or none: the command line program of the branch below the pull request (its base
	// branch), the name of that branch and its commit for the report, and the runs per seed (RunAlternating)
	belowBinEnv    = "BENCHMARK_BELOW_BIN"
	belowBranchEnv = "BENCHMARK_BELOW_BRANCH"
	belowCommitEnv = "BENCHMARK_BELOW_COMMIT"
	roundsEnv      = "BENCHMARK_ROUNDS"
)

// TestCompareWithNetLogo runs gobees and then gobees-fast as programs with the seeds and days of the NetLogo
// experiment, and writes the report. With the program of the branch below (belowBinEnv), it then runs that program
// and gobees-fast alternately (RunAlternating) for the section "Against the branch below". Skipped unless the CI (or a user) sets the environment above, so that
// `go test ./...` does not time anything.
func TestCompareWithNetLogo(t *testing.T) {
	tablePath, consolePath, reportPath := os.Getenv(tableEnv), os.Getenv(consoleEnv), os.Getenv(reportEnv)
	baseBin, currentBin := os.Getenv(baseBinEnv), os.Getenv(currentBinEnv)
	if tablePath == "" || consolePath == "" || reportPath == "" || baseBin == "" || currentBin == "" {
		t.Skipf("set %s, %s, %s, %s and %s (see go/benchmark/README.md)", tableEnv, consoleEnv, reportEnv,
			baseBinEnv, currentBinEnv)
	}
	threads, err := strconv.Atoi(os.Getenv(threadsEnv))
	if err != nil || threads < 1 {
		t.Fatalf("%s must be the number of runs at the same time, as given to run_netlogo.sh", threadsEnv)
	}
	table, err := os.Open(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()
	console, err := os.Open(consolePath)
	if err != nil {
		t.Fatal(err)
	}
	defer console.Close()
	netlogo, err := ParseNetLogo(table, console)
	if err != nil {
		t.Fatal(err)
	}

	seeds := make([]float64, len(netlogo.Runs))
	for i, run := range netlogo.Runs {
		seeds[i] = run.Seed
	}
	base, err := RunPrograms(baseBin, true, seeds, netlogo.Days, threads)
	if err != nil {
		t.Fatal(err)
	}
	current, err := RunPrograms(currentBin, false, seeds, netlogo.Days, threads)
	if err != nil {
		t.Fatal(err)
	}

	below, err := runBelow(seeds, netlogo.Days, currentBin)
	if err != nil {
		t.Fatal(err)
	}

	commit := os.Getenv("GITHUB_SHA")
	if len(commit) > 7 {
		commit = commit[:7]
	}
	machine := Machine{
		CPU:    CPUModel(),
		Cores:  runtime.NumCPU(),
		OS:     runtime.GOOS + "/" + runtime.GOARCH,
		Go:     runtime.Version(),
		Java:   os.Getenv(javaEnv),
		Commit: commit,
	}
	r := Report{NetLogo: netlogo, BaseCommit: os.Getenv(baseCommitEnv), Base: base, Current: current,
		Threads: threads, Machine: machine, Below: below}
	if err := WriteReport(reportPath, r); err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + r.Markdown())
}

// runBelow runs the program of the branch below and gobees-fast alternately, if belowBinEnv is set. Without it, the
// result is empty and the report has no section on the branch below.
func runBelow(seeds []float64, days int, currentBin string) (Below, error) {
	bin := os.Getenv(belowBinEnv)
	if bin == "" {
		return Below{}, nil
	}
	below := Below{Branch: os.Getenv(belowBranchEnv), Commit: os.Getenv(belowCommitEnv)}
	rounds, err := strconv.Atoi(os.Getenv(roundsEnv))
	if err != nil || rounds < 1 || below.Branch == "" || below.Commit == "" {
		return Below{}, fmt.Errorf("with %s set, %s and %s must be set and %s must be 1 or more", belowBinEnv,
			belowBranchEnv, belowCommitEnv, roundsEnv)
	}
	below.Rounds = rounds
	below.Runs, below.Current, err = RunAlternating(bin, currentBin, seeds, days, rounds)
	return below, err
}

// The first lines of the console output and of the table of run_netlogo.sh (netlogo.xml, 2 threads, local run of
// 2026-10-26), verbatim: the version line of Java, the setup lines of runs 1 and 2, and the table up to the first two
// runs. The runs finish in any order.
const (
	exampleConsole = `openjdk version "21.0.12.1" 2026-08-18 LTS
gobees-benchmark setup 1 3.167
gobees-benchmark setup 2 3.17
`
	exampleTable = `"BehaviorSpace results (NetLogo 5.3.1)"
"C:/Users/pothmann/01-projects/gobees/netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo"
"benchmark"
"10/26/2026 07:21:24:068 +0100"
"min-pxcor","max-pxcor","min-pycor","max-pycor"
"0","300","0","210"
"[run number]","RAND_SEED","[step]","timer","TotalColoniesEverProduced"
"2","2","365","43.721","150"
"1","1","365","54.481","154"
`
)

func TestParseNetLogo(t *testing.T) {
	got, err := ParseNetLogo(strings.NewReader(exampleTable), strings.NewReader(exampleConsole))
	if err != nil {
		t.Fatal(err)
	}
	want := NetLogo{Days: 365, Runs: []Run{
		{Number: 1, Seed: 1, Setup: 3167 * time.Millisecond, Days: 54481 * time.Millisecond, ColoniesFounded: 154},
		{Number: 2, Seed: 2, Setup: 3170 * time.Millisecond, Days: 43721 * time.Millisecond, ColoniesFounded: 150},
	}}
	if got.Days != want.Days || len(got.Runs) != len(want.Runs) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want.Runs {
		g, w := got.Runs[i], want.Runs[i]
		// NetLogo's timer has millisecond resolution, the float conversion may be off by a nanosecond
		if g.Number != w.Number || g.Seed != w.Seed || (g.Setup-w.Setup).Abs() > time.Microsecond ||
			(g.Days-w.Days).Abs() > time.Microsecond || g.ColoniesFounded != w.ColoniesFounded {
			t.Errorf("run %d: got %+v, want %+v", i+1, g, w)
		}
	}
}

func TestParseNetLogoMissingSetup(t *testing.T) {
	console := "gobees-benchmark setup 2 3.17\n" // no line for run 1
	if _, err := ParseNetLogo(strings.NewReader(exampleTable), strings.NewReader(console)); err == nil {
		t.Error("want an error for a run without setup time")
	}
}

// exampleSummary is the output of go/main.go at the end of a run (seed 1, 365 days, local run), verbatim.
const exampleSummary = `species: [{B_terrestris 500}]
colonies founded: 132, bees died: 9884
finished after 365 days
`

func TestParseSummary(t *testing.T) {
	run := Run{Seed: 1}
	if err := parseSummary(&run, exampleSummary, 365); err != nil {
		t.Fatal(err)
	}
	if run.ColoniesFounded != 132 {
		t.Errorf("colonies founded: got %g, want 132", run.ColoniesFounded)
	}
	if err := parseSummary(&run, exampleSummary, 730); err == nil {
		t.Error("want an error for a run that ended before the last day")
	}
	if err := parseSummary(&run, "species: []\n", 365); err == nil {
		t.Error("want an error for an output without summary")
	}
}

func TestWhole(t *testing.T) {
	if got := (Run{Setup: time.Second, Days: 2 * time.Second}).Whole(); got != 3*time.Second {
		t.Errorf("setup and days: got %v, want 3s", got)
	}
	if got := (Run{Setup: time.Second, Program: 5 * time.Second}).Whole(); got != 5*time.Second {
		t.Errorf("program: got %v, want 5s", got)
	}
}

func TestMarkdown(t *testing.T) {
	r := Report{
		NetLogo: NetLogo{Days: 365, Runs: []Run{
			{Number: 1, Seed: 1, Setup: 1200 * time.Millisecond, Days: 20 * time.Second, ColoniesFounded: 150},
			{Number: 2, Seed: 2, Setup: 400 * time.Millisecond, Days: 30 * time.Second, ColoniesFounded: 132},
		}},
		BaseCommit: "40dc301",
		Base: []Run{
			{Number: 1, Seed: 1, Program: 300 * time.Millisecond, ColoniesFounded: 140},
			{Number: 2, Seed: 2, Program: 500 * time.Millisecond, ColoniesFounded: 145},
		},
		Current: []Run{
			{Number: 1, Seed: 1, Program: 100 * time.Millisecond, ColoniesFounded: 140},
			{Number: 2, Seed: 2, Program: 100 * time.Millisecond, ColoniesFounded: 145},
		},
		Threads: 2,
		Machine: Machine{OS: "linux/amd64"},
	}
	md := r.Markdown()
	for _, want := range []string{
		"gobees is the base commit 40dc301",
		// NetLogo: setup and days, 21.2 s and 30.4 s, median 25.8 s (the mean of both); gobees 400 ms; gobees-fast
		// 100 ms
		"| Whole run | 25.8 s (21.2 s-30.4 s) | 400 ms (300 ms-500 ms) | 100 ms (100 ms-100 ms) | 258 | 4.00 |",
		"| 2 | 2 | 30.4 s | 132 | 500 ms | 145 | 100 ms | 145 |",
		"- OS: linux/amd64\n",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q:\n%s", want, md)
		}
	}
	if strings.Contains(md, "- CPU") {
		t.Errorf("empty CPU must be left out:\n%s", md)
	}
	if strings.Contains(md, "## Against the branch below") {
		t.Errorf("no section on the branch below without its runs:\n%s", md)
	}

	// with the branch below: step-S4 200 and 300 ms (median 250 ms), gobees-fast 100 and 300 ms (median 200 ms)
	r.Below = Below{Branch: "step-S4", Commit: "45c6b6c", Rounds: 5,
		Runs: []Run{
			{Number: 1, Seed: 1, Program: 200 * time.Millisecond, ColoniesFounded: 140},
			{Number: 2, Seed: 2, Program: 300 * time.Millisecond, ColoniesFounded: 145},
		},
		Current: []Run{
			{Number: 1, Seed: 1, Program: 100 * time.Millisecond, ColoniesFounded: 141},
			{Number: 2, Seed: 2, Program: 300 * time.Millisecond, ColoniesFounded: 146},
		},
	}
	md = r.Markdown()
	for _, want := range []string{
		"## Against the branch below\n\ngobees-fast against the base branch of the pull request, step-S4 (commit 45c6b6c)",
		"each seed 5 times per program",
		"| | step-S4 median (min-max) | gobees-fast median (min-max) | Change |",
		"| Whole run | 250 ms (200 ms-300 ms) | 200 ms (100 ms-300 ms) | -20.0 % |",
		"| 1 | 200 ms | 140 | 100 ms | 141 | -50.0 % |",
		"| 2 | 300 ms | 145 | 300 ms | 146 | +0.0 % |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q:\n%s", want, md)
		}
	}
	if i, j := strings.Index(md, "## Against the branch below"), strings.Index(md, "## Machine"); i > j {
		t.Errorf("the section on the branch below must come before the machine:\n%s", md)
	}
}
