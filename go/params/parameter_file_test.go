package params

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobees/internal/examplefiles"
)

// wantExampleFarmParameters holds every value of _SYSTEM_Example_Farm_Parameters.csv [RPAR].
// Strings are stored as """text""" in the file, """""" means an empty string.
var wantExampleFarmParameters = ParameterFile{
	AbundanceBoost:                      1,
	BeeSpeciesInitialQueensListAsString: "B_terrestris 500",
	Backgroundcolour:                    5,
	BackgroundImage:                     "none",
	Black_th:                            1,
	CropRotationListAsString:            "",
	CumulVisitsOnlyLastYear:             false,
	FlowerspeciesFile:                   "_SYSTEM_Flowerspecies.csv", // column name "FlowerspeciesFile " with trailing blank
	FoodSourceLimit:                     25,
	ForagingMortalityFactor:             1,
	ForagingMortalityModel:              "high",
	Gridsize:                            500,
	HabitatsFile:                        "_SYSTEM_Habitats.csv",
	InitialQueensModifierBehaviorSpace:  1,
	KeepDeadColonies:                    true,
	Lambda_detectProb:                   -0.005,
	MapAreaIncluded:                     "complete",
	MasterSizeFactor:                    1,
	MaxForagingRange_m:                  10000,
	MaxHibernatingQueens:                10000,
	MaxPatchRadius_m:                    500,
	MergeHedges:                         true,
	MinSizeFoodSources:                  true,
	MyFarmlandPatches:                   "",
	N_Badgers:                           0,
	RemoveEmptyFoodSources:              true,
	SatelliteFile:                       "No satellite image",
	SetColour:                           "Magenta",
	SexLocus:                            false,
	ShowCohorts:                         true,
	ShowDeadCols:                        false,
	ShowFoodsources:                     true,
	ShowGrid:                            false,
	ShowInspectedColony:                 false,
	ShowMasterpatchesOnly:               false,
	ShowNests:                           true,
	ShowPlots:                           true,
	ShowQueens:                          true,
	ShowSearchingQueens:                 true,
	ShowWeather:                         true,
	BeespeciesFile:                      "_SYSTEM_BumbleSpecies_UK_01.csv",
	StopExtinct:                         true,
	UnlimitedMales:                      true,
	Weather:                             "Constant 8 hrs",
	White_th:                            9,
	WinterMortality:                     true,
}

// wantSystemParameters holds every value of _SYSTEM_Parameters.csv. It differs from the
// example farm file only in SetColour (display only) and in the spelling TRUE/FALSE.
func wantSystemParameters() ParameterFile {
	p := wantExampleFarmParameters
	p.SetColour = "Yellow"
	return p
}

func TestParseParameterFileExamples(t *testing.T) {
	cases := []struct {
		name string
		text string
		want ParameterFile
	}{
		{"_SYSTEM_Parameters.csv", examplefiles.SystemParameters, wantSystemParameters()},
		{"_SYSTEM_Example_Farm_Parameters.csv", examplefiles.ExampleFarmParameters, wantExampleFarmParameters},
	}
	for _, c := range cases {
		// The original files have CRLF line endings, the copies LF. Both must give the same values.
		for _, lineEnd := range []string{"\n", "\r\n"} {
			got, err := ParseParameterFile(strings.NewReader(strings.ReplaceAll(c.text, "\n", lineEnd)))
			if err != nil {
				t.Fatalf("%s (line end %q): %v", c.name, lineEnd, err)
			}
			if got != c.want {
				t.Errorf("%s (line end %q):\n got  %+v\n want %+v", c.name, lineEnd, got, c.want)
			}
		}
	}
}

func TestReadParameterFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "_SYSTEM_Example_Farm_Parameters.csv")
	if err := os.WriteFile(path, []byte(examplefiles.ExampleFarmParameters), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadParameterFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != wantExampleFarmParameters {
		t.Errorf("got  %+v\nwant %+v", got, wantExampleFarmParameters)
	}
	if _, err := ReadParameterFile(filepath.Join(t.TempDir(), "missing.csv")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestParseParameterFileErrors(t *testing.T) {
	cases := map[string]string{
		"unknown parameter (NetLogo: run fails)": "Foo\n1\n",
		"string without quotes":                  "Weather\nConstant\n",
		"boolean expected":                       "StopExtinct?\n1\n",
		"no value row":                           "Weather\n",
		"removed parameter with another value":   "COLONIES_IBM\n10\n",
	}
	for name, csv := range cases {
		if _, err := ParseParameterFile(strings.NewReader(csv)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseParameterFileRemovedParameter(t *testing.T) {
	// COLONIES_IBM: individual-based colonies are removed (step F1). The NetLogo files keep the column with 0.
	p, err := ParseParameterFile(strings.NewReader("COLONIES_IBM,N_Badgers\n0,2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if p.N_Badgers != 2 {
		t.Errorf("N_Badgers = %v, want 2", p.N_Badgers)
	}
}

func TestParseParameterFileFirstColumnWins(t *testing.T) {
	// [RPAR-03] position finds the first column with that name.
	p, err := ParseParameterFile(strings.NewReader("N_Badgers,N_Badgers\n1,2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if p.N_Badgers != 1 {
		t.Errorf("N_Badgers = %v, want 1", p.N_Badgers)
	}
}
