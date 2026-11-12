package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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
	Report     []string          // what the converter did beyond copying values, one line each
	Notes      map[string]string // notes at keys of the section constants, by the path below constants
}

// Convert reads the NetLogo parameter file at path [INIT-04] and all files it names, relative to dir (the folder of
// the .nlogo), and returns them as run file and landscape files. mapName is the NetLogo setting MyMap, which gives the
// foodsources file of the start [PAR-05]. name is the key name of the run file (input.CheckName), NetLogo has none.
//
// Only the format changes: the values are those that NetLogo reads, the run file has no derived values.
func Convert(dir, path, mapName, name string) (Converted, error) {
	if err := input.CheckName(name); err != nil {
		return Converted{}, err
	}
	file, err := ReadParameterFile(path)
	if err != nil {
		return Converted{}, err
	}
	return convert(dir, &file, mapName, name)
}

func convert(dir string, file *netLogoParameters, mapName, name string) (Converted, error) {
	// NetLogo's RAND_SEED is a setting of the interface, not of the parameter file: seed 1. days: the default of
	// gobees, NetLogo runs until it is stopped. One repetition on one thread.
	days := params.Default().Termination.MaxTicks
	c := input.Config{Name: name, Run: &input.Run{Seed: 1, Days: &days, Repetitions: 1, Threads: 1},
		Parameters: file.Parameters}
	c.Output.KeepDeadColonies = file.KeepDeadColonies
	// the constants come from the code, as in NetLogo, never from the parameter file
	c.Constants = params.DefaultConstants()
	report, notes, err := reportSkipped(file.Skipped, &c.Constants)
	if err != nil {
		return Converted{}, err
	}

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
		c.ForagingPeriod = append(c.ForagingPeriod, input.Year(y))
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
		l, left, err := readFoodsourcesFile(filepath.Join(dir, n), c.StewardshipOptions)
		if err != nil {
			return Converted{}, err
		}
		landscapes[landscapeName(n)] = l
		for _, id := range left {
			report = append(report, fmt.Sprintf("%s: flower patch %d has only stewardship options and is left out "+
				"(NetLogo removes it at the setup)", n, id))
		}
	}
	return Converted{Config: c, Landscapes: landscapes, Report: report, Notes: notes}, nil
}

// reportSkipped returns the report lines and the notes for the constants: one line for the constants, one line for
// each column not taken over (the GUI columns in one line), and for a shadowed column whose value differs from
// the constant both values and what was written.
func reportSkipped(skipped []skippedColumn, constants *params.Constants) ([]string, map[string]string, error) {
	report := []string{"constants: the NetLogo values of the code (params.DefaultConstants), " +
		"not from the parameter file, which cannot set them in NetLogo"}
	notes := map[string]string{}
	gui := []string{}
	for _, s := range skipped {
		if s.GUI {
			gui = append(gui, s.Name)
			continue
		}
		if sh, ok := shadowedParameters[strings.ToLower(strings.TrimSpace(s.Name))]; ok {
			v, err := strconv.ParseFloat(strings.TrimSpace(s.Value), 64)
			if err != nil {
				return nil, nil, fmt.Errorf("parameter %q: expected a number, got %s", s.Name, s.Value)
			}
			if c := sh.value(constants); v != c {
				report = append(report, fmt.Sprintf("%s = %s has no effect in NetLogo, which uses %s = %v. "+
					"Written: constants.%s: %v", s.Name, s.Value, sh.constant, c, sh.key, c))
				notes[sh.key] = fmt.Sprintf("NetLogo parameter file: %s = %s, without effect in NetLogo", s.Name,
					s.Value)
				continue
			}
		}
		report = append(report, fmt.Sprintf("%s = %s not taken over: %s", s.Name, s.Value, s.Reason))
	}
	if len(gui) > 0 {
		report = append(report, fmt.Sprintf("%d columns not taken over, only the NetLogo GUI reads them: %s", len(gui),
			strings.Join(gui, ", ")))
	}
	return report, notes, nil
}

// landscapeName is the name of the landscape file of a NetLogo foodsources file: ".txt" becomes ".yaml".
func landscapeName(foodsourcesFile string) string {
	return strings.TrimSuffix(foodsourcesFile, ".txt") + ".yaml"
}

// readFoodsourcesFile reads a foodsources file [FSRC-02] to [FSRC-04] as a landscape file. NetLogo shows a message
// and Setup continues without foodsources [FSRC-09]. Here it is an error.
//
// The stewardship options of the flower lists (items without a value, e.g. ["Plot"]) are not converted: they mark
// which flower species grow in the option, which only the NetLogo GUI reads [LAYR-11], [LAYR-12]. An option must be
// in options (stewardshipOptions), as NetLogo stops with a runtime error at any other item without a value
// [LAYR-04]. A flower patch with only options is left out, as NetLogo splits it into no layer and removes it
// [LAYR-09]. Returns the ids of the patches left out.
//
// The coordinates are converted from NetLogo patches to metres with the metres per patch of line 1, rounded to
// millimetres. The id must be an integer.
func readFoodsourcesFile(path string, options []string) (input.Landscape, []int, error) {
	f, err := os.Open(path)
	if err != nil {
		return input.Landscape{}, nil, fmt.Errorf("FoodsourcesFile: %w", err)
	}
	defer f.Close()
	m, patches, err := parseFoodsourcesFile(f)
	if err != nil {
		return input.Landscape{}, nil, fmt.Errorf("FoodsourcesFile %s: %w", path, err)
	}
	l := input.Landscape{Patches: make([]input.Patch, 0, len(patches))}
	var left []int
	metres := func(v float64) float64 { return math.Round(v*m*1000) / 1000 }
	for _, p := range patches {
		id := int(p.Id_Beescout)
		if float64(id) != p.Id_Beescout {
			return input.Landscape{}, nil, fmt.Errorf("FoodsourcesFile %s: flower patch %v: the id must be an "+
				"integer", path, p.Id_Beescout)
		}
		for _, o := range p.Options {
			if !slices.Contains(options, o) {
				return input.Landscape{}, nil, fmt.Errorf("FoodsourcesFile %s: flower patch %v: [%q] is no "+
					"stewardship option (HabitatsFile) and no flower species with a value", path, p.Id_Beescout, o)
			}
		}
		if len(p.FlowerSpeciesList) == 0 && len(p.Options) > 0 {
			left = append(left, id)
			continue
		}
		l.Patches = append(l.Patches, input.Patch{ID: id, PatchType: p.PatchType, X: metres(p.X), Y: metres(p.Y),
			Area: p.Area_sqm, Flowers: p.FlowerSpeciesList})
	}
	return l, left, nil
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

// Files returns the run file under the name run and the landscape files with their tables by their names.
func (c Converted) Files(run string) (map[string][]byte, error) {
	files := map[string][]byte{}
	var b bytes.Buffer
	if err := input.WriteConfigNotes(&b, c.Config, c.Notes); err != nil {
		return nil, err
	}
	files[run] = slices.Clone(b.Bytes())
	for name, l := range c.Landscapes {
		lf, err := input.WriteLandscape(name, l)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		for n, content := range lf {
			if _, ok := files[n]; ok {
				return nil, fmt.Errorf("%s: file %s written twice", name, n)
			}
			files[n] = content
		}
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
