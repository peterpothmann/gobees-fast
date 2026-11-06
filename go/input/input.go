package input

import (
	"bytes"
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

// Files is a run file and its landscape files as read, before Build.
type Files struct {
	Config     Config
	Landscapes map[string]Landscape // the landscape files by their name in Config
	Raw        map[string][]byte    // the bytes of each landscape file, by its name in Config
}

// Load reads the run file at path and its landscape files, relative to the folder of the run file, and returns the
// parameters of the run (Build).
func Load(path string) (params.Params, error) {
	f, err := Read(path)
	if err != nil {
		return params.Params{}, err
	}
	return Build(f.Config, f.Landscapes)
}

// Read reads the run file at path and its landscape files, relative to the folder of the run file.
func Read(path string) (Files, error) {
	dir := filepath.Dir(path)
	return readAll(path, func(name string) (io.ReadCloser, error) {
		if !filepath.IsAbs(name) {
			name = filepath.Join(dir, name)
		}
		return os.Open(name)
	})
}

// ExampleFarm returns the parameters of the example farm, the run without a run file.
func ExampleFarm() params.Params {
	f := ExampleFarmFiles()
	p, err := Build(f.Config, f.Landscapes)
	if err != nil {
		panic(err) // the embedded files are checked by the tests
	}
	return p
}

// ExampleFarmFiles returns the run file and the landscape file of the example farm, as read.
func ExampleFarmFiles() Files {
	f, err := readAll(ExampleFarmFile, openExampleFarm)
	if err != nil {
		panic(err) // the embedded files are checked by the tests
	}
	return f
}

// ExampleFarmConfig returns the run file of the example farm and its landscape files by name, e.g. for tests that
// change them before Build.
func ExampleFarmConfig() (Config, map[string]Landscape) {
	f := ExampleFarmFiles()
	return f.Config, f.Landscapes
}

func openExampleFarm(name string) (io.ReadCloser, error) {
	return exampleFarm.Open(path.Join("examplefarm", name))
}

// readAll reads the run file name and the landscape files it names (the start and, with a crop rotation, every file
// of the crop rotation, each once).
func readAll(name string, open func(string) (io.ReadCloser, error)) (Files, error) {
	f := Files{Landscapes: map[string]Landscape{}, Raw: map[string][]byte{}}
	if _, err := readFile(name, open, func(r io.Reader) (err error) { f.Config, err = ReadConfig(r); return }); err != nil {
		return f, err
	}
	names := []string{f.Config.Landscape}
	if len(f.Config.CropRotation) > 1 {
		names = append(names, f.Config.CropRotation...)
	}
	for _, n := range names {
		if _, ok := f.Landscapes[n]; ok {
			continue
		}
		var l Landscape
		raw, err := readFile(n, open, func(r io.Reader) (err error) { l, err = ReadLandscape(r); return })
		if err != nil {
			return f, err
		}
		f.Landscapes[n] = l
		f.Raw[n] = raw
	}
	return f, nil
}

// readFile reads the file name with read and returns its bytes.
func readFile(name string, open func(string) (io.ReadCloser, error), read func(io.Reader) error) ([]byte, error) {
	f, err := open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := read(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return data, nil
}
