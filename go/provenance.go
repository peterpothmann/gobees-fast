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
	Files  map[string][]byte // the copies of the landscape files and their tables by their name in the output folder
}

// newProvenance returns the provenance of a run with the files f (read from the run file config, "" for the example
// farm) and the parameters p, which hold the seed (of the first repetition) and the days actually used. repetitions
// and threads are the numbers of repetitions and threads used. output is the daily output file with the columns
// columns (the key output.columns of the provenance, always written in full). The landscape files and their tables
// keep their name in the output folder. A second file of the same name gets _2 in front of the extension (a.yaml,
// a_2.yaml). The tables are copied byte for byte. A landscape file is copied byte for byte if its tables keep their
// names, else it is written again with the names of the copies. The landscape files of sites are not copied: sums
// holds their SHA-256 at the start, by their name in the run file.
func newProvenance(f input.Files, config string, p *params.Params, repetitions, threads int, output string,
	columns []string, started time.Time, sums map[string]string) provenance {
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
	if len(f.Sites) > 0 {
		run.Sites, c.Variations = siteRecords(f, config, filepath.Dir(output), sums)
	}

	// The landscape files in the order of the run file, each with its tables. Two of them may have the same name in
	// different folders, and none may take the name of the daily output or the provenance. Lower case, because the
	// file system of Windows and macOS does not tell a.yaml and A.yaml apart. A table of several landscape files
	// (e.g. ../habitats.csv) is copied once.
	sources := []string{c.Landscape}
	sources = append(sources, c.CropRotation...)
	copies := map[string]string{} // name in the output folder by name in the run file (landscape) or as read (table)
	used := map[string]bool{dailyOutputFile: true, provenanceFile: true}
	prov := provenance{Files: map[string][]byte{}}
	reserve := func(source string) string {
		name := filepath.Base(source)
		ext := filepath.Ext(name)
		for i := 2; used[strings.ToLower(name)]; i++ {
			name = fmt.Sprintf("%s_%d%s", strings.TrimSuffix(filepath.Base(source), ext), i, ext)
		}
		used[strings.ToLower(name)] = true
		copies[source] = name
		return name
	}
	record := func(name, source string, raw, content []byte) {
		run.Landscapes = append(run.Landscapes, input.LandscapeCopy{File: name, Source: source,
			SHA256: sha256Hex(raw)})
		prov.Files[name] = content
	}
	for _, source := range sources {
		raw, ok := f.Raw[source]
		if !ok || copies[source] != "" {
			continue // read only with a crop rotation of 2 files or more (input.Read), or copied already
		}
		name := reserve(source)
		// the tables byte for byte, the landscape file names their copies. Listed after the landscape file.
		renamed := map[string]string{}
		changed := false
		tables := []input.LandscapeTable{}
		for _, t := range f.Tables[source] {
			tableCopy, ok := copies[t.Name]
			if !ok {
				tableCopy = reserve(t.Name)
				tables = append(tables, t)
			}
			renamed[t.Ref] = tableCopy
			changed = changed || tableCopy != t.Ref
		}
		content := raw
		if changed {
			var err error
			if content, err = input.RewriteLandscapeFile(raw, renamed); err != nil {
				panic(err) // read by input.Read already
			}
		}
		record(name, source, raw, content)
		for _, t := range tables {
			record(copies[t.Name], t.Name, f.Raw[t.Name], f.Raw[t.Name])
		}
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

// siteRecords returns the record of the sites of f (read from the run file config) and the variations of the
// provenance in the folder dir. The patterns of the sites become the site files they matched, relative to dir (or
// absolute without a relative path, e.g. on another drive), so that the provenance runs the same sites also when the
// folder of the sites gets more files. The provenance does not copy the sites: it needs them in their place.
func siteRecords(f input.Files, config, dir string, sums map[string]string) ([]input.SiteRecord, []input.Variation) {
	base := filepath.Dir(config)
	files := make([]string, len(f.Sites))
	records := make([]input.SiteRecord, len(f.Sites))
	for i, s := range f.Sites {
		files[i] = relativeTo(dir, base, s.File)
		records[i] = input.SiteRecord{Site: s.Name, File: files[i]}
		// the files read for the site: the start and, with a crop rotation of 2 files or more, the rotation
		names := []string{s.Landscape}
		if len(s.CropRotation) > 1 {
			names = append(names, s.CropRotation...)
		}
		seen := map[string]bool{}
		for _, n := range names {
			if !seen[n] {
				seen[n] = true
				records[i].Landscapes = append(records[i].Landscapes, input.LandscapeSum{Source: n, SHA256: sums[n]})
			}
		}
	}
	vs := slices.Clone(f.Config.Variations)
	for i := range vs {
		if len(vs[i].Sites) > 0 {
			vs[i].Sites = files
		}
	}
	return records, vs
}

// relativeTo returns the file name, relative to the folder base, as a path relative to the folder dir with / as
// separator, or absolute if there is none.
func relativeTo(dir, base, name string) string {
	abs := name
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(base, name)
	}
	abs, err := filepath.Abs(abs)
	if err != nil {
		return filepath.ToSlash(name)
	}
	if d, err := filepath.Abs(dir); err == nil {
		if r, err := filepath.Rel(d, abs); err == nil {
			return filepath.ToSlash(r)
		}
	}
	return filepath.ToSlash(abs)
}

// sha256Hex returns the SHA-256 of data, hexadecimal.
func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
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
