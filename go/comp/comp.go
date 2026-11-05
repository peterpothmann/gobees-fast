// Package comp contains the ECS components.
//
// Naming: fields keep the NetLogo variable names with the first letter in upper case, so they can be found
// in the NetLogo code. "?" is dropped, "/" becomes "_per_" (as in package params).
// References to other agents (NetLogo: who numbers) are ecs.Entity. Agents keep no who number (step F12). NetLogo -1 ("nobody") is the zero entity.
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
// step F3). Id_Beescout is read only by an error message.
type Foodsource struct {
	Id_Beescout                   float64 // column 1 of the foodsources file
	PatchType                     string  // habitat type, e.g. "Hedgerow"
	Area_sqm                      float64 // area of the whole flower patch [m2]
	Radius_m                      float64 // radius of a circle with this area [m]
	NectarMax_myl                 float64
	PollenMax_g                   float64
	NectarFlowerVolume_myl        float64
	NectarConcentration_mol_per_l float64
	ProteinPollenProp             float64
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

// Species holds the parameters of one bee species. Not a component since step F11: the species lie in
// params.Input.Species, bees and colonies hold their index. Read-only during the run. The values that depend on the
// landscape are in globals.NestSites.
//
// NetLogo: species-own (oneSpecies agents), created in CreateSpeciesProc [SPEC].
type Species struct {
	Name string
	// Parasitic is true for the cuckoo bumblebees (Psithyrus): they rest instead of nursing and foraging [ACT-04]
	// and take over a host colony [EMER-04], [PSI]. NetLogo compares the name with "Psithyrus" at these places.
	// Derived from the name in input.deriveSpecies (step F9).
	Parasitic                          bool
	MaxLifespanWorkers                 int
	EmergingDay_mean                   float64
	EmergingDay_sd                     float64 // rounded [SPEC-03]
	Batchsize                          int
	FlightVelocity_m_per_s             float64
	FlightCosts_kJ_per_m_per_mg        float64 // derived from MetabolicRateFlight_W/kg
	SearchLength_m                     float64
	SeasonStop                         day.OfYear
	TimeUnloading                      float64
	SpecMax_cropVolume_myl             float64
	SpecMax_pollenPellets_g            float64
	NestHabitatsList                   []string // as in the file, the expansion [SPEC-04] is in globals.NestSites
	MinToMaxFactor                     float64
	DevAgeHatchingMin_d                float64
	DevAgePupationMin_d                float64
	DevAgeEmergingMin_d                float64
	DevWeightEgg_mg                    float64
	DevWeightPupationMin_mg            float64
	DevWeightPupationMax_mg            float64
	PollenToBodymassFactor             float64
	Dev_Q_DeterminationWeight_mg       float64
	DevAge_Q_PupationMin_d             float64
	DevWeight_Q_PupationMin_mg         float64
	DevWeight_Q_PupationMax_mg         float64
	DevAge_Q_EmergingMin_d             float64
	GrowthFactor                       float64
	Proboscis_min_mm                   float64
	Proboscis_max_mm                   float64
	ChanceFindNest                     float64 // column dailyNestSiteChance
	DevQuotaIncubationToday_kJ         float64 // [SPEC-05] and following: derived
	DevAgeHatchingMax_d                float64
	DevAgePupationMax_d                float64
	DevAgeEmergingMax_d                float64
	DevIncubationHatchingTH_kJ         float64
	DevIncubationPupationTH_kJ         float64
	DevIncubationEmergingTH_kJ         float64
	Dev_larvalAge_QueenDetermination_d int
	DevAge_Q_PupationMax_d             float64
	DevAge_Q_EmergingMax_d             float64
	DevIncubation_Q_PupationTH_kJ      float64
	DevIncubation_Q_EmergingTH_kJ      float64
	MinPollenStore_g                   float64 // [SPEC-07]
}
