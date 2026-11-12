package params

import (
	"fmt"
	"slices"

	"gobees/day"
)

// Input holds the content of the run file and its landscape files, and the values derived from them. Package input
// reads it before the setup, so no system opens a file. Read-only during the run.
type Input struct {
	Derived        Derived        // the derived values, Params.Apply adds them as resource
	ForagingPeriod ForagingPeriod // daily foraging period, Params.Apply adds it as resource. NetLogo: the parameter Weather

	// Species are the species of the run file whose name is in Derived.SpeciesNames, in file order, with the
	// derived values. NestHabitats is as in the file: the expansion and the nest sites need the foodsources and
	// happen in sys.nestSitesSetup.
	Species []Species

	// Flowers are the flower species of the run file, nil without them.
	Flowers *FlowerSpeciesTable

	// Foodsources holds the landscape files by their name in the run file: the one of the
	// start (Derived.StartLandscape, the current one is globals.Landscape.File) and, if there is a crop rotation, every
	// file of Derived.CropRotation.
	Foodsources map[string]Foodsources
}

// Foodsources is the content of one landscape file (NetLogo: foodsources file).
type Foodsources struct {
	Patches []FlowerPatch // in file order
}

// FlowerPatch is one flower patch of a landscape file: a whole flower patch before it is split into layers.
type FlowerPatch struct {
	ID        int // column id of the landscape file. NetLogo: id_Beescout
	PatchType string
	X, Y      float64       // centre [m]. NetLogo: xcor, ycor in NetLogo patches
	Area      float64       // [m2]. NetLogo: size_sqm
	Flowers   []FlowerEntry // the flower species and their relative abundance. NetLogo: flowerSpeciesList
}

// FlowerEntry is one flower species of a flower patch and its relative abundance, e.g. ["Bramble" 0.331] of a NetLogo
// flowerSpeciesList.
type FlowerEntry struct {
	Name  string
	Value float64
}

// FlowerSpecies is the data per flower of one flower species (NetLogo: one row of the FlowerspeciesFile).
// The keys of the run file (tag `yaml`) are the Go names in lowerCamelCase, the comments name the columns
// of the NetLogo FlowerspeciesFile.
type FlowerSpecies struct {
	Name            string     `yaml:"name"`            // NetLogo: Flowerspecies
	PollenPerFlower float64    `yaml:"pollenPerFlower"` // [g]. NetLogo: pollen_g/flower
	NectarPerFlower float64    `yaml:"nectarPerFlower"` // [ml]. NetLogo: nectar_ml/flower
	Concentration   float64    `yaml:"concentration"`   // sugar concentration of the nectar [mol/l]. NetLogo: concentration_mol/l
	StartDay        day.OfYear `yaml:"startDay"`        // first day of the flowering period
	StopDay         day.OfYear `yaml:"stopDay"`         // first day AFTER the flowering period
	CorollaDepth    float64    `yaml:"corollaDepth"`    // [mm]. NetLogo: corollaDepth_mm
	InterFlowerTime float64    `yaml:"interFlowerTime"` // time from one flower to the next [s]. NetLogo: intFlowerTime_s
}

// FlowerSpeciesTable holds all flower species of the run file in file order.
type FlowerSpeciesTable struct {
	Names []string // allFlowerspeciesList
	Rows  []FlowerSpecies
}

// Get returns the FIRST row with the name of a flower species. A missing species is an error.
func (t *FlowerSpeciesTable) Get(name string) (*FlowerSpecies, error) {
	i := slices.Index(t.Names, name)
	if i < 0 {
		return nil, fmt.Errorf("flower species %q not in the flower species of the run file", name)
	}
	return &t.Rows[i], nil
}
