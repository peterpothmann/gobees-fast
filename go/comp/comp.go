// Package comp contains the ECS components.
//
// Naming: fields keep the NetLogo variable names with the first letter in upper case, so they can be found
// in the NetLogo code. "?" is dropped, "/" becomes "_per_" (as in package params).
// References to other agents (NetLogo: who numbers) are ecs.Entity. NetLogo -1 ("nobody") is the zero entity.
package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// Who is the NetLogo who number of an agent. NetLogo numbers all turtles of all breeds with one counter.
// Some model code sorts by who, so gobees keeps it (see trans/data-model.md).
type Who struct {
	Who int64
}

// Position of a turtle in NetLogo patch units (xcor, ycor).
type Position struct {
	X float64
	Y float64
}

// Badger marks a badger entity. A badger has the components Who, Position and Badger.
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
	LayersInPatchList []ecs.Entity // all layers of the same flower patch, sorted by who [LAYR-10], finding 15 (see trans/netlogo-findings.md#finding-15), fixed in step F5
}

// Species holds the parameters of one bee species.
//
// NetLogo: species-own (oneSpecies agents), created in CreateSpeciesProc [SPEC].
type Species struct {
	Name                               string
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
	NestHabitatsList                   []string // expanded by NestHabitatsListREP [SPEC-04]
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
	NestsiteFoodsourceList             []ecs.Entity // [SPEC-06] masterpatches, sorted by who (an agentset in NetLogo)
	NestSiteArea                       float64
	MinPollenStore_g                   float64 // [SPEC-07]
}
