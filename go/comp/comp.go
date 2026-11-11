package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// Position of an entity [m]. NetLogo: xcor, ycor in NetLogo patches.
type Position struct {
	X float64
	Y float64
}

// Badger marks a badger entity. A badger has the components Position and Badger. NetLogo: breed badgers.
type Badger struct{}

// A flower patch entity has the components Position and FlowerPatch. It is one flower patch of the landscape file
// with one layer per flower species. NetLogo: the masterpatch and the other layers of its flower patch, each a
// foodsource turtle (foodsources-own).

// FlowerPatch holds a flower patch and its layers.
type FlowerPatch struct {
	ID        int     // column id of the landscape file. NetLogo: id_Beescout
	PatchType string  // habitat type, e.g. "Hedgerow"
	Area      float64 // [m2]. NetLogo: area_sqm
	Radius    float64 // radius of a circle with this area [m]. NetLogo: radius_m
	// Layers holds the layers that survived the setup, in the order of the flower list of the patch (NetLogo: the who
	// order of the layers, layersInPatchList). Empty for a patch without flower species, which is not split (NetLogo:
	// a masterpatch without food).
	Layers []Layer
	// PollenInFlower and NectarInFlower: a layer has pollen or nectar after the morning refill (flowerRefill). The
	// food of a layer is above 0 after the refill exactly if it is in flower and its maximum is above 0.
	PollenInFlower bool
	NectarInFlower bool
}

// Layer is one flower species on a flower patch: its static data and its food. NetLogo: a foodsource layer.
type Layer struct {
	Flower              int        // index of the flower species in params.Input.Flowers, read only by the user interface (ui). NetLogo: flowerSpecies
	NectarMax           float64    // [µl]. NetLogo: nectarMax_myl
	PollenMax           float64    // [g]. NetLogo: pollenMax_g
	NectarPerFlower     float64    // nectar volume of one flower [µl]. NetLogo: nectarFlowerVolume_myl
	NectarConcentration float64    // [mol/l]. NetLogo: nectarConcentration_mol_per_l
	StartDay            day.OfYear // first day of the flowering period
	StopDay             day.OfYear // first day AFTER the flowering period
	CorollaDepth        float64    // [mm]. NetLogo: corollaDepth_mm
	InterFlowerTime     float64    // [s]. NetLogo: interFlowerTime_s
	Nectar              float64    // [µl]. NetLogo: nectar_myl
	Pollen              float64    // [g]. NetLogo: pollen_g
}

// Source is one layer of a flower patch, e.g. the foodsource of a forager. The zero value is no layer.
// NetLogo: the who number of a foodsource layer.
type Source struct {
	Patch ecs.Entity // the flower patch
	Layer int        // index in FlowerPatch.Layers
}

// IsZero is true for no layer.
func (s Source) IsZero() bool { return s.Patch.IsZero() }
