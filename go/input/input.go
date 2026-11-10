package input

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	Landscapes map[string]Landscape        // the landscape files by their name in Config
	Tables     map[string][]LandscapeTable // the tables of each landscape file, by its name in Config
	Raw        map[string][]byte           // the bytes of each landscape file (by its name in Config) and table (by LandscapeTable.Name)
	Sites      []Site                      // the sites of the variation sites, in the order of its patterns

	source []byte                                   // the run file as read
	open   func(name string) (io.ReadCloser, error) // opens a file relative to the folder of the run file
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
	// absolute, because the opener also opens the run file itself and would put its folder in front a second time
	path, err := filepath.Abs(path)
	if err != nil {
		return Files{}, err
	}
	dir := filepath.Dir(path)
	open := func(name string) (io.ReadCloser, error) {
		if !filepath.IsAbs(name) {
			name = filepath.Join(dir, name)
		}
		return os.Open(name)
	}
	// the matches of a relative pattern are relative to the folder of the run file, as the paths in the run file
	glob := func(pattern string) ([]string, error) {
		if filepath.IsAbs(pattern) {
			return filepath.Glob(pattern)
		}
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		for i, m := range matches {
			if matches[i], err = filepath.Rel(dir, m); err != nil {
				return nil, err
			}
		}
		return matches, err
	}
	return readAll(path, open, glob)
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
	f, err := readAll(ExampleFarmFile, openExampleFarm, globExampleFarm)
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

func globExampleFarm(pattern string) ([]string, error) {
	sub, err := fs.Sub(exampleFarm, "examplefarm")
	if err != nil {
		return nil, err
	}
	return fs.Glob(sub, pattern)
}

// readAll reads the run file name and the landscape files it names (the start and, with a crop rotation, every file
// of the crop rotation, each once) or, with sites, the site files that the patterns match (glob). The landscape files
// of the sites are read by each combination.
func readAll(name string, open func(string) (io.ReadCloser, error), glob func(string) ([]string, error)) (Files,
	error) {
	r, err := open(name)
	if err != nil {
		return Files{}, err
	}
	data, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		return Files{}, fmt.Errorf("%s: %w", name, err)
	}
	f, err := readFrom(data, open, Files{})
	if err != nil {
		var landscape landscapeError
		if errors.As(err, &landscape) {
			return f, err // names its file
		}
		return f, fmt.Errorf("%s: %w", name, err)
	}
	if f.Sites, err = resolveSites(f.Config.Variations, glob, open); err != nil {
		return f, fmt.Errorf("%s: %w", name, err)
	}
	return f, nil
}

// landscapeNames returns the landscape files of the run file c: the start and, with a crop rotation, every file of
// the crop rotation.
func landscapeNames(c Config) []string {
	names := []string{c.Landscape}
	if len(c.CropRotation) > 1 {
		names = append(names, c.CropRotation...)
	}
	return names
}

// landscapeError is an error in a landscape file. Its message names the file.
type landscapeError struct{ error }

func (e landscapeError) Unwrap() error { return e.error }

// readLandscape reads the landscape file name into f, once.
func (f *Files) readLandscape(name string) error {
	if _, ok := f.Landscapes[name]; ok {
		return nil
	}
	l, tables, raw, err := ReadLandscape(name, f.open)
	if err != nil {
		return landscapeError{err}
	}
	f.Landscapes[name] = l
	f.Tables[name] = tables
	for n, b := range raw {
		f.Raw[n] = b
	}
	return nil
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
