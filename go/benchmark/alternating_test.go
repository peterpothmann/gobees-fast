package benchmark

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeEnv makes the test binary act as the command line program (TestMain): it appends its file name and its
// arguments to the file named by fakeEnv, prints a summary for 365 days and exits.
const fakeEnv = "GOBEES_BENCHMARK_FAKE_LOG"

func TestMain(m *testing.M) {
	if log := os.Getenv(fakeEnv); log != "" {
		f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(2)
		}
		fmt.Fprintln(f, strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), strings.Join(os.Args[1:], " "))
		f.Close()
		fmt.Print("colonies founded: 7, bees died: 1\nfinished after 365 days\n")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeProgram copies the test binary to dir/name, as a program for RunAlternating.
func fakeProgram(t *testing.T, dir, name string) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(self)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	// Windows runs only files ending in .exe. Elsewhere no extension: the test binary itself is benchmark.test.
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(dir, name)
	dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunAlternating(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls.txt")
	t.Setenv(fakeEnv, log)
	a, b := fakeProgram(t, dir, "a"), fakeProgram(t, dir, "b")

	runsA, runsB, err := RunAlternating(a, b, []float64{1, 2}, 365, 2)
	if err != nil {
		t.Fatal(err)
	}
	// one untimed run each, then round 1: a before b, round 2: b before a
	want := `a -s 1 --days 365
b -s 1 --days 365
a -s 1 --days 365
b -s 1 --days 365
a -s 2 --days 365
b -s 2 --days 365
b -s 1 --days 365
a -s 1 --days 365
b -s 2 --days 365
a -s 2 --days 365
`
	calls, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.ReplaceAll(string(calls), "\r\n", "\n"); got != want {
		t.Errorf("calls:\n%s\nwant:\n%s", got, want)
	}
	for _, runs := range [][]Run{runsA, runsB} {
		if len(runs) != 2 {
			t.Fatalf("got %d runs, want 2", len(runs))
		}
		for i, run := range runs {
			if run.Number != i+1 || run.Seed != float64(i+1) || run.ColoniesFounded != 7 || run.Program <= 0 {
				t.Errorf("run %d: %+v", i+1, run)
			}
		}
	}

	if _, _, err := RunAlternating(a, b, []float64{1}, 365, 0); err == nil {
		t.Error("want an error for 0 rounds")
	}
}
