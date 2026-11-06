package main

import (
	"os"
	"path"
	"path/filepath"

	"gobees/compare"
	"gobees/examplefiles"
)

// Folders of the generated files in the repository, relative to its root.
const (
	exampleFarmFolder = "go/input/examplefarm" // embedded in gobees-fast, the run without a run file
	scenariosFolder   = "go/compare/scenarios" // the scenarios of go/compare
)

// Generate converts the NetLogo files that gobees-fast needs as run files: the example farm and the scenarios of
// go/compare (compare.Scenarios). It returns the files by their path relative to the root of the repository and the
// report of the converter for each run file.
//
// The NetLogo files are the copies in gobees/examplefiles and the experiment files in go/validation/experiments, so
// that the result does not depend on the files in netlogo/.
func Generate() (map[string][]byte, []string, error) {
	dir, err := os.MkdirTemp("", "netlogo2gobees")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(dir)
	if err := compare.WriteInputFiles(dir); err != nil {
		return nil, nil, err
	}

	result := map[string][]byte{}
	report := []string{}
	add := func(folder, run, parameters string) error {
		file := filepath.Join(dir, "parameters.csv")
		if err := os.WriteFile(file, []byte(parameters), 0o644); err != nil {
			return err
		}
		c, err := Convert(dir, file, DefaultMap)
		if err != nil {
			return err
		}
		files, err := c.Files(run)
		if err != nil {
			return err
		}
		for name, content := range files {
			result[path.Join(folder, name)] = content
		}
		report = append(report, path.Join(folder, run)+":")
		for _, line := range c.Report {
			report = append(report, "  "+line)
		}
		return nil
	}

	if err := add(exampleFarmFolder, "run.yaml", examplefiles.ExampleFarmParameters); err != nil {
		return nil, nil, err
	}
	for _, s := range compare.Scenarios {
		text, err := s.ParameterFile()
		if err != nil {
			return nil, nil, err
		}
		if err := add(scenariosFolder, s.ConfigFile(), text); err != nil {
			return nil, nil, err
		}
	}
	return result, report, nil
}
