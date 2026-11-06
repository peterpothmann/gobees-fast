package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gobees/examplefiles"
	"gobees/input"
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
func exampleFarmFile(t *testing.T) netLogoParameters {
	t.Helper()
	file, err := ParseParameterFile(strings.NewReader(examplefiles.ExampleFarmParameters))
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func TestConvertExampleFarm(t *testing.T) {
	dir := exampleFarmDir(t, nil)
	conv, err := Convert(dir, filepath.Join(dir, "_SYSTEM_Example_Farm_Parameters.csv"), DefaultMap)
	if err != nil {
		t.Fatal(err)
	}
	c := conv.Config
	if c.Parameters != wantExampleFarmParameters.Parameters {
		t.Errorf("Parameters:\n got  %+v\n want %+v", c.Parameters, wantExampleFarmParameters.Parameters)
	}
	// [QINI] "B_terrestris 500"
	if !reflect.DeepEqual(c.InitialQueens, []params.SpeciesQueens{{Species: "B_terrestris", Queens: 500}}) {
		t.Errorf("InitialQueens = %+v", c.InitialQueens)
	}
	if !reflect.DeepEqual(c.StewardshipOptions, wantCSOList) {
		t.Errorf("StewardshipOptions = %q, want %q", c.StewardshipOptions, wantCSOList)
	}
	// [FPER] "Constant 8 hrs"
	if len(c.ForagingPeriod) != 1 || len(c.ForagingPeriod[0]) != 365 || c.ForagingPeriod[0][100] != 28800 {
		t.Errorf("ForagingPeriod: want one year of 28800 s")
	}
	// all rows of the species and flower files; their values are checked in TestParseSpeciesFileExample and
	// TestParseFlowerspeciesFileExample
	if len(c.Species) != 7 || len(c.Flowers) != 79 {
		t.Errorf("%d species and %d flower species, want 7 and 79", len(c.Species), len(c.Flowers))
	}
	// [PAR-05] MyMap _SYSTEM_Example_Farm.png -> _SYSTEM_Example_Farm_Foodsources.txt, as a landscape file
	l, ok := conv.Landscapes[c.Landscape]
	if c.Landscape != "_SYSTEM_Example_Farm_Foodsources.yaml" || len(conv.Landscapes) != 1 || !ok ||
		l.MetresPerPatch != 10 || len(l.Patches) != 73 || c.CropRotation != nil {
		t.Errorf("Landscape %q, %d landscape files, crop rotation %q", c.Landscape, len(conv.Landscapes), c.CropRotation)
	}
	// Line 4: 0 "Hedgerow" 105 130 175 200 [["Average_Willow" 1.275] ...] 20  "no info": colour, perimeter and info
	// are not converted
	p := l.Patches[0]
	if p.ID != 0 || p.PatchType != "Hedgerow" || p.X != 130 || p.Y != 175 || p.Area != 200 ||
		len(p.Flowers) != 33 || p.Flowers[0] != (input.FlowerEntry{Name: "Average_Willow", Value: 1.275, HasValue: true}) {
		t.Errorf("first flower patch = %+v", p)
	}
}

func TestConvertCropRotation(t *testing.T) {
	// [CROP-01] every file of the crop rotation is converted, each once.
	dir := exampleFarmDir(t, map[string]string{"Example_Farm_PlotTest_Foodsources.txt": examplefiles.ExampleFarmPlotTestFoodsources})
	file := exampleFarmFile(t)
	file.CropRotationListAsString = "_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt"
	conv, err := convert(dir, &file, DefaultMap)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"_SYSTEM_Example_Farm_Foodsources.yaml", "Example_Farm_PlotTest_Foodsources.yaml", "Example_Farm_PlotTest_Foodsources.yaml"}
	if !reflect.DeepEqual(conv.Config.CropRotation, want) || len(conv.Landscapes) != 2 ||
		len(conv.Landscapes["Example_Farm_PlotTest_Foodsources.yaml"].Patches) != 36 {
		t.Errorf("CropRotation %q, %d landscape files", conv.Config.CropRotation, len(conv.Landscapes))
	}

	// A missing file of the crop rotation is an error.
	file.CropRotationListAsString = "_SYSTEM_Example_Farm_Foodsources.txt missing_Foodsources.txt"
	if _, err := convert(dir, &file, DefaultMap); err == nil {
		t.Error("expected an error for a missing file of the crop rotation")
	}
	// Without crop rotation (one item), the item is not converted: "" is no file (finding 3).
	file.CropRotationListAsString = "missing_Foodsources.txt"
	if conv, err := convert(dir, &file, DefaultMap); err != nil || conv.Config.CropRotation != nil {
		t.Errorf("a crop rotation list with one item must not be converted: %v", err)
	}
}

func TestConvertOtherMap(t *testing.T) {
	// MyMap gives the foodsources file of the start [PAR-05].
	dir := exampleFarmDir(t, map[string]string{"Example_Farm_PlotTest_Foodsources.txt": examplefiles.ExampleFarmPlotTestFoodsources})
	file := exampleFarmFile(t)
	conv, err := convert(dir, &file, "Example_Farm_PlotTest.png")
	if err != nil || conv.Config.Landscape != "Example_Farm_PlotTest_Foodsources.yaml" ||
		len(conv.Landscapes[conv.Config.Landscape].Patches) != 36 {
		t.Errorf("Landscape %q, %v", conv.Config.Landscape, err)
	}
}

func TestConvertNoFlowerspeciesFile(t *testing.T) {
	// [FSRC-12] "No Input File": no flower species, no error.
	dir := exampleFarmDir(t, nil)
	file := exampleFarmFile(t)
	file.FlowerspeciesFile = "No Input File"
	conv, err := convert(dir, &file, DefaultMap)
	if err != nil || conv.Config.Flowers != nil {
		t.Errorf("Flowers = %v, %v, want none without error", conv.Config.Flowers, err)
	}
}

func TestConvertMissingFiles(t *testing.T) {
	// Every missing input file is an error.
	for _, name := range []string{"_SYSTEM_Habitats.csv", "_SYSTEM_Example_Farm_Foodsources.txt",
		"_SYSTEM_Flowerspecies.csv", "_SYSTEM_BumbleSpecies_UK_01.csv"} {
		dir := exampleFarmDir(t, nil)
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
		file := exampleFarmFile(t)
		if _, err := convert(dir, &file, DefaultMap); err == nil {
			t.Errorf("expected an error without %s", name)
		}
	}
	if _, err := Convert(t.TempDir(), "missing.csv", DefaultMap); err == nil {
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

func TestQueensInitialProc(t *testing.T) {
	cases := []struct {
		input   string
		queens  []params.SpeciesQueens
		species []string
	}{
		{"B_hortorum 100 B_terrestris 500",
			[]params.SpeciesQueens{{Species: "B_hortorum", Queens: 100}, {Species: "B_terrestris", Queens: 500}},
			[]string{"B_hortorum", "B_terrestris"}},
		{"A 1 B", []params.SpeciesQueens{{Species: "A", Queens: 1}}, []string{"A"}}, // trailing name is dropped
		{"", []params.SpeciesQueens{}, []string{}},                                  // [""] gives no species
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
