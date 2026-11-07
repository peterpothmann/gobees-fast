package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"gobees/input"
	"gobees/params"
)

// exampleFarmConfig is the key run.config of a run without a run file.
const exampleFarmConfig = "built-in example farm"

// randomSeed returns a random seed other than 0, so that a run with seed 0 can be repeated with the seed recorded in
// the run info.
func randomSeed() int64 {
	for {
		if s := rand.Int64(); s != 0 {
			return s
		}
	}
}

// runInfoPath returns the run info of the daily output file output: its name without extension plus
// "_run_info.yaml", in the same folder (daily_output.csv -> daily_output_run_info.yaml).
func runInfoPath(output string) string {
	return strings.TrimSuffix(output, filepath.Ext(output)) + "_run_info.yaml"
}

// runInfo is the run info of a run: a run file that repeats the run, with the record of the run in the section run,
// and the copies of its landscape files.
type runInfo struct {
	Config input.Config
	Files  map[string][]byte // the copies of the landscape files by their path
}

// newRunInfo returns the run info of a run with the files f (read from the run file config, "" for the example farm)
// and the parameters p, which hold the seed and the days actually used. output is the daily output file with the
// columns columns (the key output.columns of the run info, always written in full). The landscape files are copied
// next to output, byte for byte.
func newRunInfo(f input.Files, config string, p *params.Params, output string, columns []string,
	started time.Time) runInfo {
	c := f.Config
	if config == "" {
		config = exampleFarmConfig
	}
	days := p.Termination.MaxTicks
	run := &input.Run{
		Seed:    p.Seed,
		Days:    &days,
		Started: started.Format(time.RFC3339),
		Command: slices.Clone(os.Args),
		Version: buildVersion(),
		Config:  config,
		Output:  output,
	}
	c.Output.Columns = slices.Clone(columns)

	// The landscape files in the order of the run file. Their copies get the name of the output in front, so that
	// the run infos of several outputs in one folder do not overwrite each other's copies.
	sources := []string{c.Landscape}
	sources = append(sources, c.CropRotation...)
	stem := strings.TrimSuffix(filepath.Base(output), filepath.Ext(output))
	copies := map[string]string{} // name in the run info by name in the run file
	used := map[string]bool{}
	info := runInfo{Files: map[string][]byte{}}
	for _, source := range sources {
		raw, ok := f.Raw[source]
		if !ok || copies[source] != "" {
			continue // read only with a crop rotation of 2 files or more (input.Read), or copied already
		}
		name := stem + "_" + filepath.Base(source)
		for i := 2; used[name]; i++ {
			name = fmt.Sprintf("%s_%d_%s", stem, i, filepath.Base(source))
		}
		used[name] = true
		copies[source] = name
		sum := sha256.Sum256(raw)
		run.Landscapes = append(run.Landscapes, input.LandscapeCopy{File: name, Source: source,
			SHA256: hex.EncodeToString(sum[:])})
		info.Files[filepath.Join(filepath.Dir(output), name)] = raw
	}

	c.Run = run
	c.Landscape = copies[c.Landscape]
	if len(c.CropRotation) > 0 {
		rotation := make([]string, len(c.CropRotation))
		for i, n := range c.CropRotation {
			rotation[i] = n
			if name, ok := copies[n]; ok {
				rotation[i] = name
			}
		}
		c.CropRotation = rotation
	}
	info.Config = c
	return info
}

// write writes the run info to path and the copies of the landscape files. It creates a missing folder.
func (r runInfo) write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b bytes.Buffer
	if err := input.WriteConfig(&b, r.Config); err != nil {
		return err
	}
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		return err
	}
	for name, raw := range r.Files {
		if err := os.WriteFile(name, raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// buildVersion returns the build of this program. The keys of the version control system are only known in a
// binary built with go build inside the repository, not with go run.
func buildVersion() *input.Version {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	v := &input.Version{Module: bi.Main.Path, Version: bi.Main.Version, Go: bi.GoVersion}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			v.Revision = s.Value
		case "vcs.time":
			v.Time = s.Value
		case "vcs.modified":
			v.Modified = s.Value
		}
	}
	return v
}
