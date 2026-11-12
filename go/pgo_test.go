package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"

	"gobees/input"
)

// pgoEnv names the file for the CPU profile of TestWritePGOProfile.
const pgoEnv = "GOBEES_PGO"

// TestWritePGOProfile writes the CPU profile for the profile-guided optimisation (default.pgo, README): the example
// farm, 20 runs of 730 days, one thread. Skipped unless pgoEnv is set, so that `go test ./...` writes nothing.
func TestWritePGOProfile(t *testing.T) {
	path := os.Getenv(pgoEnv)
	if path == "" {
		t.Skipf("set %s to the file of the profile, e.g. default.pgo (see README)", pgoEnv)
	}
	cfg, landscapes := input.ExampleFarmConfig()
	days := 730
	cfg.Run = &input.Run{Seed: 1, Days: &days, Repetitions: 20, Threads: 1}
	dir := t.TempDir()
	writeRunFile(t, dir, "run.yaml", cfg, map[string]input.Landscape{cfg.Landscape: landscapes[cfg.Landscape]})
	c := parse(t, "-c", filepath.Join(dir, "run.yaml"))

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		t.Fatal(err)
	}
	err = c.run(&bytes.Buffer{})
	pprof.StopCPUProfile()
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
