package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"gobees/input"
	"gobees/params"
)

// The files in the output folder of a run file (-o DIR: DIR/<name>/), next to the copies of its landscape files.
const (
	dailyOutputFile = "daily.csv"       // the daily output
	provenanceFile  = "provenance.yaml" // the provenance, a run file that repeats the run
)

// exampleFarmConfig is the key run.config of a run without a run file.
const exampleFarmConfig = "built-in example farm"

// provenance is the provenance of the output of a run file: a run file that repeats the run, with the record of the
// run in the section run, and the copies of its landscape files.
type provenance struct {
	Config input.Config
	Files  map[string][]byte // the copies of the landscape files by their name in the output folder
}

// newProvenance returns the provenance of a run with the files f (read from the run file config, "" for the example
// farm) and the parameters p, which hold the seed (of the first repetition) and the days actually used. repetitions
// and threads are the numbers of repetitions and threads used. output is the daily output file with the columns
// columns (the key output.columns of the provenance, always written in full). The landscape files keep their name in
// the output folder. A second file of the same name gets _2 in front of the extension (a.yaml, a_2.yaml).
func newProvenance(f input.Files, config string, p *params.Params, repetitions, threads int, output string,
	columns []string, started time.Time) provenance {
	c := f.Config
	if config == "" {
		config = exampleFarmConfig
	}
	days := p.Termination.MaxTicks
	run := &input.Run{
		Seed:        p.Seed,
		Days:        &days,
		Repetitions: repetitions,
		Threads:     threads,
		Started:     started.Format(time.RFC3339),
		Command:     slices.Clone(os.Args),
		Version:     buildVersion(),
		Config:      config,
		Output:      output,
	}
	c.Output.Columns = slices.Clone(columns)

	// The landscape files in the order of the run file. Two of them may have the same name in different folders, and
	// none may take the name of the daily output or the provenance. Lower case, because the file system of Windows
	// and macOS does not tell a.yaml and A.yaml apart.
	sources := []string{c.Landscape}
	sources = append(sources, c.CropRotation...)
	copies := map[string]string{} // name in the output folder by name in the run file
	used := map[string]bool{dailyOutputFile: true, provenanceFile: true}
	prov := provenance{Files: map[string][]byte{}}
	for _, source := range sources {
		raw, ok := f.Raw[source]
		if !ok || copies[source] != "" {
			continue // read only with a crop rotation of 2 files or more (input.Read), or copied already
		}
		name := filepath.Base(source)
		ext := filepath.Ext(name)
		for i := 2; used[strings.ToLower(name)]; i++ {
			name = fmt.Sprintf("%s_%d%s", strings.TrimSuffix(filepath.Base(source), ext), i, ext)
		}
		used[strings.ToLower(name)] = true
		copies[source] = name
		sum := sha256.Sum256(raw)
		run.Landscapes = append(run.Landscapes, input.LandscapeCopy{File: name, Source: source,
			SHA256: hex.EncodeToString(sum[:])})
		prov.Files[name] = raw
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
	prov.Config = c
	return prov
}

// write writes the provenance (provenanceFile) and the copies of the landscape files into the folder dir. It creates a
// missing folder.
func (p provenance) write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var b bytes.Buffer
	if err := input.WriteConfig(&b, p.Config); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, provenanceFile), b.Bytes(), 0o644); err != nil {
		return err
	}
	for name, raw := range p.Files {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o644); err != nil {
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
