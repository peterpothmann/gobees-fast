package input

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"gobees/params"
)

// exampleFarm holds the example farm as run file and landscape file, converted from the NetLogo files with
// tools/netlogo2gobees (the tool's tests check that the files are up to date).
//
//go:embed all:examplefarm
var exampleFarm embed.FS

// ExampleFarmFile is the run file of the example farm in exampleFarm.
const ExampleFarmFile = "run.yaml"

// Load reads the run file at path and its landscape files, relative to the folder of the run file, and returns the
// parameters of the run (Build).
func Load(path string) (params.Params, error) {
	dir := filepath.Dir(path)
	return load(path, func(name string) (io.ReadCloser, error) {
		if !filepath.IsAbs(name) {
			name = filepath.Join(dir, name)
		}
		return os.Open(name)
	})
}

// ExampleFarm returns the parameters of the example farm, the run without a run file.
func ExampleFarm() params.Params {
	p, err := load(ExampleFarmFile, openExampleFarm)
	if err != nil {
		panic(err) // the embedded files are checked by the tests
	}
	return p
}

// ExampleFarmConfig returns the run file of the example farm and its landscape files by name, e.g. for tests that
// change them before Build.
func ExampleFarmConfig() (Config, map[string]Landscape) {
	c, landscapes, err := readAll(ExampleFarmFile, openExampleFarm)
	if err != nil {
		panic(err)
	}
	return c, landscapes
}

func openExampleFarm(name string) (io.ReadCloser, error) {
	return exampleFarm.Open(path.Join("examplefarm", name))
}

func load(name string, open func(string) (io.ReadCloser, error)) (params.Params, error) {
	c, landscapes, err := readAll(name, open)
	if err != nil {
		return params.Params{}, err
	}
	return Build(c, landscapes)
}

// readAll reads the run file name and the landscape files it names (the start and, with a crop rotation, every file
// of the crop rotation, each once).
func readAll(name string, open func(string) (io.ReadCloser, error)) (Config, map[string]Landscape, error) {
	c := Config{}
	if err := readFile(name, open, func(r io.Reader) (err error) { c, err = ReadConfig(r); return }); err != nil {
		return c, nil, err
	}
	names := []string{c.Landscape}
	if len(c.CropRotation) > 1 {
		names = append(names, c.CropRotation...)
	}
	landscapes := map[string]Landscape{}
	for _, n := range names {
		if _, ok := landscapes[n]; ok {
			continue
		}
		var l Landscape
		if err := readFile(n, open, func(r io.Reader) (err error) { l, err = ReadLandscape(r); return }); err != nil {
			return c, nil, err
		}
		landscapes[n] = l
	}
	return c, landscapes, nil
}

func readFile(name string, open func(string) (io.ReadCloser, error), read func(io.Reader) error) error {
	f, err := open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := read(f); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
