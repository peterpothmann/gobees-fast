// Package input reads the NetLogo input files of a run into params.Params before the setup [INIT-04], [INIT-05].
// No system opens a file afterwards.
//
// The files keep the NetLogo format, so that NetLogo, gobees and gobees-fast read the same files (validation,
// benchmark, comparisons).
package input

import (
	"fmt"
	"os"
	"path/filepath"

	"gobees/globals"
	"gobees/params"
)

// StartFoodsourcesFile is the foodsources file of the start [PAR-05].
//
// NetLogo derives it from the interface setting MyMap ("X.png" -> "X_Foodsources.txt"), and the .nlogo stores
// MyMap = "_SYSTEM_Example_Farm.png". gobees-fast has no MyMap since step F6, so every run starts on the example farm
// until gobees-fast has its own input format. Other maps are only reached through the crop rotation [CROP-01].
const StartFoodsourcesFile = "_SYSTEM_Example_Farm_Foodsources.txt"

// noFlowerspeciesFile is the value of FlowerspeciesFile without a file [FSRC-12].
const noFlowerspeciesFile = "No Input File"

// Load reads the parameter file at path [INIT-04] and all input files it names (Read). dir is the folder of the
// input files, the folder of the .nlogo in NetLogo. The other values are the ones of params.Default.
func Load(dir, path string) (params.Params, error) {
	p := params.Default()
	file, err := ReadParameterFile(path)
	if err != nil {
		return p, err
	}
	p.File = file
	if p.Input, err = Read(dir, &p.File, &p.Constants); err != nil {
		return p, err
	}
	return p, nil
}

// Read reads all input files that file names, relative to dir, and derives the values of ParametersProc
// [INIT-05]. A missing or invalid file is an error.
func Read(dir string, file *params.ParameterFile, constants *params.Constants) (params.Input, error) {
	in := params.Input{}

	// [PAR-01]
	queens, species, err := QueensInitialProc(file.BeeSpeciesInitialQueensListAsString)
	if err != nil {
		return in, err
	}
	// [PAR-06]
	cso, err := AllCSOlistREP(filepath.Join(dir, file.HabitatsFile))
	if err != nil {
		return in, err
	}
	// [FPER] NetLogo: Foraging_periodREP every morning
	if in.Weather, err = weatherFromParameter(file.Weather, dir); err != nil {
		return in, err
	}
	// [PAR-04]
	mortality, err := mortalityForager(file.ForagingMortalityModel)
	if err != nil {
		return in, err
	}
	in.Derived = globals.Derived{
		BeeSpeciesInitialQueensList: queens,
		SpeciesList:                 species,
		MortalityForager_per_s:      mortality,
		FoodsourcesFile:             StartFoodsourcesFile, // [PAR-05]
		CSS_OptionsList:             cso,
		CropRotationList:            ListFromStringREP(file.CropRotationListAsString), // [PAR-07]
	}

	// [SPEC-01] to [SPEC-03], [SPEC-05], [SPEC-07]
	if in.Species, err = readSpeciesFile(filepath.Join(dir, file.BeespeciesFile), species, constants); err != nil {
		return in, err
	}

	// [FSRC-12] NetLogo reads the file in every CreateFoodsourcesProc, i.e. again in each crop rotation.
	if file.FlowerspeciesFile != noFlowerspeciesFile {
		t, err := readFlowerspeciesFile(filepath.Join(dir, file.FlowerspeciesFile))
		if err != nil {
			return in, err
		}
		in.Flowers = &t
	}

	// [FSRC-02] to [FSRC-04] the start map and, if the crop rotation runs (more than one file [MORN-09]), every file
	// of the crop rotation [CROP-01]. NetLogo reads a file of the crop rotation only on its Day 365, so a missing one
	// stops the run there instead of in Setup.
	names := []string{StartFoodsourcesFile}
	if len(in.Derived.CropRotationList) > 1 {
		names = append(names, in.Derived.CropRotationList...)
	}
	in.Foodsources = map[string]params.Foodsources{}
	for _, name := range names {
		if _, ok := in.Foodsources[name]; ok {
			continue
		}
		fs, err := readFoodsourcesFile(filepath.Join(dir, name))
		if err != nil {
			return in, err
		}
		in.Foodsources[name] = fs
	}
	return in, nil
}

// readFoodsourcesFile reads a foodsources file [FSRC-02] to [FSRC-04]. NetLogo shows a message and Setup continues
// without foodsources [FSRC-09]. Here it is an error.
func readFoodsourcesFile(path string) (params.Foodsources, error) {
	f, err := os.Open(path)
	if err != nil {
		return params.Foodsources{}, fmt.Errorf("FoodsourcesFile: %w", err)
	}
	defer f.Close()
	m, p, err := parseFoodsourcesFile(f)
	if err != nil {
		return params.Foodsources{}, fmt.Errorf("FoodsourcesFile %s: %w", path, err)
	}
	return params.Foodsources{MetresPerPatch: m, Patches: p}, nil
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
