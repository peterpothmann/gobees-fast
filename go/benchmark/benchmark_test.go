package benchmark

import (
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
)

// TestCompareWithNetLogo runs gobees with the seeds and days of the NetLogo experiment and writes the report.
// Skipped unless the CI (or a user) sets the environment above, so that `go test ./...` does not time anything.
func TestCompareWithNetLogo(t *testing.T) {
	tablePath, consolePath, reportPath := os.Getenv(tableEnv), os.Getenv(consoleEnv), os.Getenv(reportEnv)
	if tablePath == "" || consolePath == "" || reportPath == "" {
		t.Skipf("set %s, %s and %s (see go/benchmark/README.md)", tableEnv, consoleEnv, reportEnv)
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
	gobees, err := RunGobees(seeds, netlogo.Days, threads)
	if err != nil {
		t.Fatal(err)
	}

	commit := os.Getenv("GITHUB_SHA")
	if len(commit) > 7 {
		commit = commit[:7]
	}
	r := Report{NetLogo: netlogo, Gobees: gobees, Threads: threads, Machine: Machine{
		CPU:    CPUModel(),
		Cores:  runtime.NumCPU(),
		OS:     runtime.GOOS + "/" + runtime.GOARCH,
		Go:     runtime.Version(),
		Java:   os.Getenv(javaEnv),
		Commit: commit,
	}}
	if err := WriteReport(reportPath, r); err != nil {
		t.Fatal(err)
	}
	t.Log("\n" + r.Markdown())
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

func TestRunGobees(t *testing.T) {
	runs, err := RunGobees([]float64{1, 2, 3}, 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	for i, run := range runs {
		if run.Number != i+1 || run.Seed != float64(i+1) || run.Setup <= 0 || run.Days <= 0 {
			t.Errorf("run %d: %+v", i+1, run)
		}
	}
}

func TestMarkdown(t *testing.T) {
	r := Report{
		NetLogo: NetLogo{Days: 365, Runs: []Run{
			{Number: 1, Seed: 1, Setup: 1200 * time.Millisecond, Days: 20 * time.Second, ColoniesFounded: 150},
			{Number: 2, Seed: 2, Setup: 400 * time.Millisecond, Days: 30 * time.Second, ColoniesFounded: 132},
		}},
		Gobees: []Run{
			{Number: 1, Seed: 1, Setup: 4 * time.Millisecond, Days: 100 * time.Millisecond, ColoniesFounded: 140},
			{Number: 2, Seed: 2, Setup: 6 * time.Millisecond, Days: 150 * time.Millisecond, ColoniesFounded: 145},
		},
		Threads: 2,
		Machine: Machine{OS: "linux/amd64"},
	}
	md := r.Markdown()
	for _, want := range []string{
		"| Setup | 800 ms (400 ms-1.2 s) | 5 ms (4 ms-6 ms) | 160 |", // medians of 2 runs: the mean of both
		"| 365 days | 25.0 s (20.0 s-30.0 s) | 125 ms (100 ms-150 ms) | 200 |",
		"| 2 | 2 | 400 ms | 30.0 s | 132 | 6 ms | 150 ms | 145 |",
		"- OS: linux/amd64\n",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q:\n%s", want, md)
		}
	}
	if strings.Contains(md, "- CPU") {
		t.Errorf("empty CPU must be left out:\n%s", md)
	}
}
