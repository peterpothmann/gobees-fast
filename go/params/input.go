package params

import (
	"fmt"
	"slices"

	"gobees/comp"
	"gobees/globals"
)

// Input holds the content of all input files that the parameter file names, and the values that ParametersProc
// derives from the parameters [INIT-05]. Package input reads it before the setup, so no system opens a file.
// Read-only during the run.
type Input struct {
	Derived globals.Derived // [PAR] start values of the derived globals, sys.Parameters adds a copy as resource
	Weather globals.Weather // [FPER] daily foraging period of the parameter Weather

	// Species are the rows of the BeespeciesFile whose name is in Derived.SpeciesList, in file order, with the
	// derived values [SPEC-01] to [SPEC-03], [SPEC-05], [SPEC-07]. NestHabitatsList is as in the file: the
	// expansion [SPEC-04] and the nest sites [SPEC-06] need the foodsources and happen in sys.CreateSpecies.
	Species []comp.Species

	// Flowers is the FlowerspeciesFile [LAYR-02], nil if FlowerspeciesFile is "No Input File" [FSRC-12].
	Flowers *FlowerSpeciesTable

	// Foodsources holds the foodsources files by file name [FSRC-02] to [FSRC-04]: the one of the start
	// (Derived.FoodsourcesFile) and, if there is a crop rotation, every file of Derived.CropRotationList [CROP-01].
	Foodsources map[string]Foodsources
}

// Foodsources is the content of one foodsources file.
type Foodsources struct {
	MetresPerPatch float64       // line 1: metres per NetLogo patch [FSRC-02]
	Patches        []FlowerPatch // one per data line, in file order [FSRC-04]
}

// FlowerPatch is one data line of a foodsources file: a whole flower patch before it is split into layers [FSRC-04].
type FlowerPatch struct {
	Id_Beescout       float64
	PatchType         string
	PatchColour       float64 // GUI only
	X, Y              float64
	Area_sqm          float64
	FlowerSpeciesList []FlowerEntry
	Perimeter_m       float64 // GUI only
	PatchInfo         string
}

// FlowerEntry is one item of a flowerSpeciesList, e.g. ["Bramble" 0.331] or ["Margin"].
type FlowerEntry struct {
	Name     string
	Value    float64
	HasValue bool
}

// FlowerSpecies is one row of the FlowerspeciesFile, i.e. the data per flower of one flower species [LAYR-02].
type FlowerSpecies struct {
	Name                    string
	Pollen_g_per_flower     float64
	Nectar_ml_per_flower    float64
	ProteinPollenProp       float64
	Concentration_mol_per_l float64
	StartDay                float64
	StopDay                 float64
	CorollaDepth_mm         float64
	IntFlowerTime_s         float64
}

// FlowerSpeciesTable holds all rows of the FlowerspeciesFile in file order.
type FlowerSpeciesTable struct {
	Names []string // allFlowerspeciesList
	Rows  []FlowerSpecies
}

// Get returns the row of a flower species. NetLogo: item (position mySpecies allFlowerspeciesList + 1),
// i.e. the FIRST row with this name. A missing species is an error (NetLogo: message, then runtime error).
func (t *FlowerSpeciesTable) Get(name string) (*FlowerSpecies, error) {
	i := slices.Index(t.Names, name)
	if i < 0 {
		return nil, fmt.Errorf("flower species %q not in the FlowerspeciesFile", name)
	}
	return &t.Rows[i], nil
}
