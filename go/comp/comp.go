package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// Position of a turtle in NetLogo patch units (xcor, ycor).
type Position struct {
	X float64
	Y float64
}

// Badger marks a badger entity. A badger has the components Position and Badger.
// NetLogo: breed badgers, without own variables. Created in CreateBadgersProc [BADC].
type Badger struct{}

// Foodsource holds the static data of a foodsource layer: one flower species on one flower patch.
//
// NetLogo: foodsources-own. Not translated: colorMemo (GUI), farmland? (never read, finding 17, see trans/netlogo-findings.md#finding-17),
// stewardshipSpeciesList (GUI), perimeter_m (GUI), cso_area_m2 (never set), flowerSpeciesList (only needed to
// create the layers), and the never read patchInfo, occurrence, flowerSpecies and relativeAbundance (removed in
// step F3), proteinPollenProp (set in [LAYR-05], never read, removed in step F13). ID is read only by an error
// message.
type Foodsource struct {
	ID                  float64    // column 1 of the foodsources file. NetLogo: id_Beescout
	PatchType           string     // habitat type, e.g. "Hedgerow"
	Area                float64    // area of the whole flower patch [m2]. NetLogo: area_sqm
	Radius              float64    // radius of a circle with this area [m]. NetLogo: radius_m
	NectarMax           float64    // [µl]. NetLogo: nectarMax_myl
	PollenMax           float64    // [g]. NetLogo: pollenMax_g
	NectarPerFlower     float64    // nectar volume of one flower [µl]. NetLogo: nectarFlowerVolume_myl
	NectarConcentration float64    // [mol/l]. NetLogo: nectarConcentration_mol_per_l
	StartDay            day.OfYear // first day of the flowering period
	StopDay             day.OfYear // first day AFTER the flowering period
	CorollaDepth        float64    // [mm]. NetLogo: corollaDepth_mm
	InterFlowerTime     float64    // [s]. NetLogo: interFlowerTime_s
}

// FoodsourceState holds the changing data of a foodsource layer.
type FoodsourceState struct {
	Nectar float64 // [µl]. NetLogo: nectar_myl
	Pollen float64 // [g]. NetLogo: pollen_g
}

// Layer links a foodsource to the other foodsources of its flower patch.
type Layer struct {
	IsMasterpatch bool         // true for the first surviving layer of a flower patch [LAYR-08]. NetLogo: masterpatch?
	Masterpatch   ecs.Entity   // the masterpatch of this layer's flower patch, itself for a masterpatch. NetLogo: masterpatchID
	PatchLayers   []ecs.Entity // all layers of the same flower patch, sorted by who (= order of creation) [LAYR-10], finding 15 (see trans/netlogo-findings.md#finding-15), fixed in step F5. NetLogo: layersInPatchList
}
