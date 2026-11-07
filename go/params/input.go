package params

import (
	"fmt"
	"slices"

	"gobees/day"
)

// Input holds the content of the run file and its landscape files, and the values that ParametersProc derives from
// the parameters [INIT-05]. Package input reads it before the setup, so no system opens a file. Read-only during the
// run.
type Input struct {
	Derived Derived // [PAR] the derived globals, Params.Apply adds them as resource
	Weather Weather // [FPER] daily foraging period of the parameter Weather, Params.Apply adds it as resource

	// Species are the species of the run file whose name is in Derived.SpeciesList, in file order, with the
	// derived values [SPEC-01] to [SPEC-03], [SPEC-05], [SPEC-07]. NestHabitatsList is as in the file: the
	// expansion [SPEC-04] and the nest sites [SPEC-06] need the foodsources and happen in sys.CreateSpecies.
	Species []Species

	// Flowers are the flower species of the run file [LAYR-02], nil without them (NetLogo: FlowerspeciesFile is
	// "No Input File" [FSRC-12]).
	Flowers *FlowerSpeciesTable

	// Foodsources holds the landscape files by their name in the run file [FSRC-02] to [FSRC-04]: the one of the
	// start (Derived.FoodsourcesFile, the current one is globals.Landscape.FoodsourcesFile) and, if there is a crop rotation, every file of Derived.CropRotationList
	// [CROP-01].
	Foodsources map[string]Foodsources
}

// Foodsources is the content of one landscape file (NetLogo: foodsources file).
type Foodsources struct {
	MetresPerPatch float64       // metres per NetLogo patch [FSRC-02]
	Patches        []FlowerPatch // in file order [FSRC-04]
}

// FlowerPatch is one flower patch of a landscape file: a whole flower patch before it is split into layers [FSRC-04].
// Of the columns of the NetLogo foodsources file, patchColour, perimeter_m and info are not read since step F7: only
// the NetLogo GUI uses them.
type FlowerPatch struct {
	Id_Beescout       float64
	PatchType         string
	X, Y              float64
	Area_sqm          float64
	FlowerSpeciesList []FlowerEntry
}

// FlowerEntry is one item of a flowerSpeciesList, e.g. ["Bramble" 0.331] or ["Margin"].
type FlowerEntry struct {
	Name     string
	Value    float64
	HasValue bool
}

// FlowerSpecies is the data per flower of one flower species [LAYR-02] (NetLogo: one row of the FlowerspeciesFile).
// The keys of the run file (tag `yaml`) are the column names of the NetLogo FlowerspeciesFile.
type FlowerSpecies struct {
	Name                    string     `yaml:"Flowerspecies"`
	Pollen_g_per_flower     float64    `yaml:"pollen_g/flower"`
	Nectar_ml_per_flower    float64    `yaml:"nectar_ml/flower"`
	Concentration_mol_per_l float64    `yaml:"concentration_mol/l"`
	StartDay                day.OfYear `yaml:"startDay"`
	StopDay                 day.OfYear `yaml:"stopDay"`
	CorollaDepth_mm         float64    `yaml:"corollaDepth_mm"`
	IntFlowerTime_s         float64    `yaml:"intFlowerTime_s"`
}

// FlowerSpeciesTable holds all flower species of the run file in file order.
type FlowerSpeciesTable struct {
	Names []string // allFlowerspeciesList
	Rows  []FlowerSpecies
}

// Get returns the row of a flower species. NetLogo: item (position mySpecies allFlowerspeciesList + 1),
// i.e. the FIRST row with this name. A missing species is an error (NetLogo: message, then runtime error).
func (t *FlowerSpeciesTable) Get(name string) (*FlowerSpecies, error) {
	i := slices.Index(t.Names, name)
	if i < 0 {
		return nil, fmt.Errorf("flower species %q not in the flower species of the run file", name)
	}
	return &t.Rows[i], nil
}
