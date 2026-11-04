package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobees/input"
	"gobees/params"
)

// DefaultMap is the map of the NetLogo interface setting MyMap stored in the .nlogo [INIT-02].
const DefaultMap = "_SYSTEM_Example_Farm.png"

// noFlowerspeciesFile is the value of FlowerspeciesFile without a file [FSRC-12].
const noFlowerspeciesFile = "No Input File"

// Converted is a run file and its landscape files by their name in the run file.
type Converted struct {
	Config     input.Config
	Landscapes map[string]input.Landscape
}

// Convert reads the NetLogo parameter file at path [INIT-04] and all files it names, relative to dir (the folder of
// the .nlogo), and returns them as run file and landscape files. mapName is the NetLogo setting MyMap, which gives the
// foodsources file of the start [PAR-05].
//
// Only the format changes: the values are those that NetLogo reads, the run file has no derived values.
func Convert(dir, path, mapName string) (Converted, error) {
	file, err := ReadParameterFile(path)
	if err != nil {
		return Converted{}, err
	}
	return convert(dir, &file, mapName)
}

func convert(dir string, file *netLogoParameters, mapName string) (Converted, error) {
	c := input.Config{Parameters: file.ParameterFile}

	// [PAR-01], [QINI]
	queens, _, err := QueensInitialProc(file.BeeSpeciesInitialQueensListAsString)
	if err != nil {
		return Converted{}, err
	}
	c.InitialQueens = queens
	// [PAR-06]
	if c.StewardshipOptions, err = AllCSOlistREP(filepath.Join(dir, file.HabitatsFile)); err != nil {
		return Converted{}, err
	}
	// [FPER]
	years, err := weatherFromParameter(file.Weather, dir)
	if err != nil {
		return Converted{}, err
	}
	for _, y := range years {
		c.ForagingPeriod_s = append(c.ForagingPeriod_s, input.Year(y))
	}
	// [SPEC-01], [SPEC-03]: all species of the file
	if c.Species, err = readSpeciesFile(filepath.Join(dir, file.BeespeciesFile)); err != nil {
		return Converted{}, err
	}
	// [FSRC-12], [LAYR-02]
	if file.FlowerspeciesFile != noFlowerspeciesFile {
		t, err := readFlowerspeciesFile(filepath.Join(dir, file.FlowerspeciesFile))
		if err != nil {
			return Converted{}, err
		}
		c.Flowers = t.Rows
	}

	// [PAR-05] "X.png" -> "X_Foodsources.txt", [PAR-07]. A crop rotation list with one item is no crop rotation
	// [MORN-09] and is not converted (by default [""], finding 3).
	start := strings.ReplaceAll(mapName, ".png", "") + "_Foodsources.txt"
	names := []string{start}
	c.Landscape = landscapeName(start)
	if rotation := ListFromStringREP(file.CropRotationListAsString); len(rotation) > 1 {
		names = append(names, rotation...)
		for _, n := range rotation {
			c.CropRotation = append(c.CropRotation, landscapeName(n))
		}
	}
	landscapes := map[string]input.Landscape{}
	for _, n := range names {
		if _, ok := landscapes[landscapeName(n)]; ok {
			continue
		}
		l, err := readFoodsourcesFile(filepath.Join(dir, n))
		if err != nil {
			return Converted{}, err
		}
		landscapes[landscapeName(n)] = l
	}
	return Converted{Config: c, Landscapes: landscapes}, nil
}

// landscapeName is the name of the landscape file of a NetLogo foodsources file: ".txt" becomes ".yaml".
func landscapeName(foodsourcesFile string) string {
	return strings.TrimSuffix(foodsourcesFile, ".txt") + ".yaml"
}

// readFoodsourcesFile reads a foodsources file [FSRC-02] to [FSRC-04] as a landscape file. NetLogo shows a message
// and Setup continues without foodsources [FSRC-09]. Here it is an error.
func readFoodsourcesFile(path string) (input.Landscape, error) {
	f, err := os.Open(path)
	if err != nil {
		return input.Landscape{}, fmt.Errorf("FoodsourcesFile: %w", err)
	}
	defer f.Close()
	m, patches, err := parseFoodsourcesFile(f)
	if err != nil {
		return input.Landscape{}, fmt.Errorf("FoodsourcesFile %s: %w", path, err)
	}
	l := input.Landscape{MetresPerPatch: m, Patches: make([]input.Patch, len(patches))}
	for i, p := range patches {
		list := make([]input.FlowerEntry, len(p.FlowerSpeciesList))
		for j, e := range p.FlowerSpeciesList {
			list[j] = input.FlowerEntry(e)
		}
		l.Patches[i] = input.Patch{Id: p.Id_Beescout, PatchType: p.PatchType, X: p.X, Y: p.Y, Area_sqm: p.Area_sqm,
			FlowerSpeciesList: list}
	}
	return l, nil
}

// readFlowerspeciesFile reads the FlowerspeciesFile [LAYR-02]. NetLogo shows a message and creates no layers.
// Here it is an error.
func readFlowerspeciesFile(path string) (params.FlowerSpeciesTable, error) {
	f, err := os.Open(path)
	if err != nil {
		return params.FlowerSpeciesTable{}, fmt.Errorf("FlowerspeciesFile: %w", err)
	}
	defer f.Close()
	t, err := parseFlowerspeciesFile(f)
	if err != nil {
		return t, fmt.Errorf("FlowerspeciesFile %s: %w", path, err)
	}
	return t, nil
}

// Files returns the run file under the name run and the landscape files by their names, as YAML.
func (c Converted) Files(run string) (map[string][]byte, error) {
	files := map[string][]byte{}
	var b bytes.Buffer
	if err := input.WriteConfig(&b, c.Config); err != nil {
		return nil, err
	}
	files[run] = slices.Clone(b.Bytes())
	for name, l := range c.Landscapes {
		b.Reset()
		if err := input.WriteLandscape(&b, l); err != nil {
			return nil, err
		}
		files[name] = slices.Clone(b.Bytes())
	}
	return files, nil
}

// writeFiles writes files (name: content) into dir.
func writeFiles(dir string, files map[string][]byte) error {
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
			return err
		}
	}
	return nil
}
