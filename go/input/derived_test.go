package input

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobees/globals"
	"gobees/internal/examplefiles"
	"gobees/params"
)

// wantCSOList are the stewardship option names in _SYSTEM_Habitats.csv, in file order [CSO].
var wantCSOList = []string{"Legume", "Margin", "newCSO_Legume", "BirdBumble_Margin", "GameCover_Plot", "Plot"}

// exampleFarmDir writes the input files of the example farm and the files in extra to a temporary folder and
// returns it.
func exampleFarmDir(t *testing.T, extra map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"_SYSTEM_Example_Farm_Parameters.csv":  examplefiles.ExampleFarmParameters,
		"_SYSTEM_Habitats.csv":                 examplefiles.Habitats,
		"_SYSTEM_Example_Farm_Foodsources.txt": examplefiles.ExampleFarmFoodsources,
		"_SYSTEM_Flowerspecies.csv":            examplefiles.Flowerspecies,
		"_SYSTEM_BumbleSpecies_UK_01.csv":      examplefiles.BumbleSpecies,
	}
	for name, text := range extra {
		files[name] = text
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// exampleFarmFile returns the parameters of _SYSTEM_Example_Farm_Parameters.csv.
func exampleFarmFile(t *testing.T) params.ParameterFile {
	t.Helper()
	file, err := ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestLoadExampleFarm(t *testing.T) {
	dir := exampleFarmDir(t, nil)
	p, err := Load(dir, filepath.Join(dir, "_SYSTEM_Example_Farm_Parameters.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if p.File != wantExampleFarmParameters {
		t.Errorf("File:\n got  %+v\n want %+v", p.File, wantExampleFarmParameters)
	}
	// the other values of params.Default
	def := params.Default()
	if p.Seed != def.Seed || p.World != def.World || p.Constants != def.Constants || p.Termination != def.Termination {
		t.Errorf("Seed, World, Constants or Termination differ from params.Default")
	}

	in := p.Input
	// [PAR-01], [PAR-04] to [PAR-07]
	want := globals.Derived{
		BeeSpeciesInitialQueensList: []globals.SpeciesQueens{{Species: "B_terrestris", Queens: 500}},
		SpeciesList:                 []string{"B_terrestris"},
		MortalityForager_per_s:      1.0e-05,
		FoodsourcesFile:             "_SYSTEM_Example_Farm_Foodsources.txt",
		CSS_OptionsList:             wantCSOList,
		CropRotationList:            []string{""}, // finding 3 (see trans/netlogo-findings.md#finding-3): length 1, no crop rotation
	}
	if !reflect.DeepEqual(in.Derived, want) {
		t.Errorf("Derived:\n got  %+v\n want %+v", in.Derived, want)
	}
	// [FPER] Weather "Constant 8 hrs"
	if in.Weather.Constant_s != 28800 || in.Weather.Years_s != nil {
		t.Errorf("Weather = %+v, want a constant 28800 s", in.Weather)
	}
	// [SPEC-01] only the species of SpeciesList; the values are checked in TestParseSpeciesFileExample
	if len(in.Species) != 1 || in.Species[0].Name != "B_terrestris" {
		t.Errorf("Species = %d, want B_terrestris only", len(in.Species))
	}
	// [FSRC-12], [LAYR-02]; the rows are checked in TestParseFlowerspeciesFileExample
	if in.Flowers == nil || len(in.Flowers.Rows) != 79 {
		t.Errorf("Flowers: want the 79 rows of _SYSTEM_Flowerspecies.csv")
	}
	// [FSRC-02] to [FSRC-04] only the start map, without crop rotation; the lines are checked in
	// TestParseFoodsourcesFileExampleFarm
	fs, ok := in.Foodsources[StartFoodsourcesFile]
	if len(in.Foodsources) != 1 || !ok || fs.MetresPerPatch != 10 || len(fs.Patches) != 73 {
		t.Errorf("Foodsources: want only %s with 10 m per patch and 73 flower patches, got %d files",
			StartFoodsourcesFile, len(in.Foodsources))
	}
}

func TestReadCropRotationFiles(t *testing.T) {
	// [CROP-01] every file of the crop rotation is read before the setup, each file once.
	dir := exampleFarmDir(t, map[string]string{"Example_Farm_PlotTest_Foodsources.txt": examplefiles.ExampleFarmPlotTestFoodsources})
	file := exampleFarmFile(t)
	file.CropRotationListAsString = "_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt"
	constants := params.DefaultConstants()
	in, err := Read(dir, &file, &constants)
	if err != nil {
		t.Fatal(err)
	}
	if len(in.Foodsources) != 2 || len(in.Foodsources["Example_Farm_PlotTest_Foodsources.txt"].Patches) != 36 ||
		len(in.Foodsources[StartFoodsourcesFile].Patches) != 73 {
		t.Errorf("want the start map (73 flower patches) and the PlotTest map (36), got %d files", len(in.Foodsources))
	}

	// A missing file of the crop rotation is an error before the setup (NetLogo: on Day 365).
	file.CropRotationListAsString = "_SYSTEM_Example_Farm_Foodsources.txt missing_Foodsources.txt"
	if _, err := Read(dir, &file, &constants); err == nil {
		t.Error("expected an error for a missing file of the crop rotation")
	}
	// Without crop rotation (one item), the item is not read: "" is no file (finding 3).
	file.CropRotationListAsString = "missing_Foodsources.txt"
	if _, err := Read(dir, &file, &constants); err != nil {
		t.Errorf("a crop rotation list with one item must not be read: %v", err)
	}
}

func TestReadNoFlowerspeciesFile(t *testing.T) {
	// [FSRC-12] "No Input File": no table, no error.
	dir := exampleFarmDir(t, nil)
	file := exampleFarmFile(t)
	file.FlowerspeciesFile = "No Input File"
	constants := params.DefaultConstants()
	in, err := Read(dir, &file, &constants)
	if err != nil || in.Flowers != nil {
		t.Errorf("Flowers = %v, %v, want nil without error", in.Flowers, err)
	}
}

func TestReadMissingFiles(t *testing.T) {
	// Every missing input file is an error of Read, before the setup.
	constants := params.DefaultConstants()
	for _, name := range []string{"_SYSTEM_Habitats.csv", "_SYSTEM_Example_Farm_Foodsources.txt",
		"_SYSTEM_Flowerspecies.csv", "_SYSTEM_BumbleSpecies_UK_01.csv"} {
		dir := exampleFarmDir(t, nil)
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
		file := exampleFarmFile(t)
		if _, err := Read(dir, &file, &constants); err == nil {
			t.Errorf("expected an error without %s", name)
		}
	}
	if _, err := Load(t.TempDir(), "missing.csv"); err == nil {
		t.Error("expected an error for a missing parameter file")
	}
}

func TestParseCSOListExample(t *testing.T) {
	// The original file has CRLF line endings, the copy LF. Both must give the same list.
	for _, lineEnd := range []string{"\n", "\r\n"} {
		got, err := parseCSOList(strings.NewReader(strings.ReplaceAll(examplefiles.Habitats, "\n", lineEnd)))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, wantCSOList) {
			t.Errorf("line end %q: got %q, want %q", lineEnd, got, wantCSOList)
		}
	}
}

func TestAllCSOListMissingFile(t *testing.T) {
	// NetLogo shows a message and continues with an empty list. gobees stops with an error.
	if _, err := AllCSOlistREP(filepath.Join(t.TempDir(), "missing.csv")); err == nil {
		t.Error("expected an error for a missing HabitatsFile")
	}
}

func TestMortalityForagerUnknownModel(t *testing.T) {
	// Finding 5 (see trans/netlogo-findings.md#finding-5), fixed in step F5: case-sensitive, any other value is an
	// error (NetLogo: 0).
	for _, m := range []string{"High", "medium", ""} {
		if got, err := mortalityForager(m); err == nil {
			t.Errorf("mortalityForager(%q) = %v, want an error", m, got)
		}
	}
	if got, err := mortalityForager("high"); err != nil || got != 1.0e-05 {
		t.Errorf("mortalityForager(\"high\") = %v, %v, want 1e-05", got, err)
	}
}

func TestQueensInitialProc(t *testing.T) {
	cases := []struct {
		input   string
		queens  []globals.SpeciesQueens
		species []string
	}{
		{"B_hortorum 100 B_terrestris 500",
			[]globals.SpeciesQueens{{Species: "B_hortorum", Queens: 100}, {Species: "B_terrestris", Queens: 500}},
			[]string{"B_hortorum", "B_terrestris"}},
		{"A 1 B", []globals.SpeciesQueens{{Species: "A", Queens: 1}}, []string{"A"}}, // trailing name is dropped
		{"", []globals.SpeciesQueens{}, []string{}},                                  // [""] gives no species
	}
	for _, c := range cases {
		q, s, err := QueensInitialProc(c.input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(q, c.queens) || !reflect.DeepEqual(s, c.species) {
			t.Errorf("QueensInitialProc(%q) = %v, %v; want %v, %v", c.input, q, s, c.queens, c.species)
		}
	}
	if _, _, err := QueensInitialProc("B_terrestris many"); err == nil {
		t.Error("expected an error for a non-numeric number of queens")
	}
}
