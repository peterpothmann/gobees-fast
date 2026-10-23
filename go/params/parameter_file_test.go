package params

import (
	"strings"
	"testing"
)

func TestReadParameterFileSystemDefaults(t *testing.T) {
	for _, file := range []string{"_SYSTEM_Parameters.csv", "_SYSTEM_Example_Farm_Parameters.csv"} {
		p, err := ReadParameterFile("../../netlogo/" + file)
		if err != nil {
			t.Fatal(err)
		}
		checks := []struct {
			name      string
			got, want any
		}{
			{"BeeSpeciesInitialQueensListAsString", p.BeeSpeciesInitialQueensListAsString, "B_terrestris 500"},
			{"CropRotationListAsString", p.CropRotationListAsString, ""},
			{"MyFarmlandPatches", p.MyFarmlandPatches, ""},
			{"FlowerspeciesFile (column name with trailing blank)", p.FlowerspeciesFile, "_SYSTEM_Flowerspecies.csv"},
			{"ForagingMortalityModel", p.ForagingMortalityModel, "high"},
			{"Weather", p.Weather, "Constant 8 hrs"},
			{"Lambda_detectProb", p.Lambda_detectProb, -0.005},
			{"MaxHibernatingQueens", p.MaxHibernatingQueens, 10000.0},
			{"CumulVisitsOnlyLastYear? (FALSE/false)", p.CumulVisitsOnlyLastYear, false},
			{"StopExtinct? (TRUE/true)", p.StopExtinct, true},
		}
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("%s: %s = %#v, want %#v", file, c.name, c.got, c.want)
			}
		}
	}
}

func TestParseParameterFileErrors(t *testing.T) {
	cases := map[string]string{
		"unknown parameter (NetLogo: run fails)": "Foo\n1\n",
		"string without quotes":                  "Weather\nConstant\n",
		"boolean expected":                       "StopExtinct?\n1\n",
		"no value row":                           "Weather\n",
	}
	for name, csv := range cases {
		if _, err := ParseParameterFile(strings.NewReader(csv)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
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
