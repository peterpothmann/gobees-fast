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
// step F3), proteinPollenProp (set in [LAYR-05], never read, removed in step F13). Id_Beescout is read only by an
// error message.
type Foodsource struct {
	Id_Beescout                   float64 // column 1 of the foodsources file
	PatchType                     string  // habitat type, e.g. "Hedgerow"
	Area_sqm                      float64 // area of the whole flower patch [m2]
	Radius_m                      float64 // radius of a circle with this area [m]
	NectarMax_myl                 float64
	PollenMax_g                   float64
	NectarFlowerVolume_myl        float64
	NectarConcentration_mol_per_l float64
	StartDay                      day.OfYear // first day of the flowering period
	StopDay                       day.OfYear // first day AFTER the flowering period
	CorollaDepth_mm               float64
	InterFlowerTime_s             float64
}

// FoodsourceState holds the changing data of a foodsource layer.
type FoodsourceState struct {
	Nectar_myl float64
	Pollen_g   float64
}

// Layer links a foodsource to the other foodsources of its flower patch.
type Layer struct {
	Masterpatch       bool         // true for the first surviving layer of a flower patch [LAYR-08]
	MasterpatchID     ecs.Entity   // the masterpatch of this layer's flower patch, itself for a masterpatch
	LayersInPatchList []ecs.Entity // all layers of the same flower patch, sorted by who (= order of creation) [LAYR-10], finding 15 (see trans/netlogo-findings.md#finding-15), fixed in step F5
}
