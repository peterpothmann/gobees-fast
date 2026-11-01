package comp

import "github.com/mlange-42/ark/ecs"

// A colony entity has the components Who, Position, Colony, ColonyStores, ColonyStimuli, ColonyCounts and
// ColonyRange. NetLogo: colonies-own (56 variables), split by topic (trans/data-model.md).
// Not translated: masterpatchesInRangeList (never read, finding 23, see trans/netlogo-findings.md#finding-23).

// Colony holds identity, phenology and flags of a colony.
type Colony struct {
	SpeciesIDcolony      ecs.Entity
	QueenProduction      bool    // queenProduction?
	QueenProductionDate  float64 // NotSetHigh until set
	SwitchPointDate      float64 // NotSetHigh until set
	CompetitionPointDate float64 // NotSetHigh until set
	EusocialPhaseDate    float64 // NotSetHigh until set
}

// ColonyStores holds the stores and the daily needs of a colony.
type ColonyStores struct {
	EnergyStore_kJ           float64
	PollenStore_g            float64
	IdealEnergyStore_kJ      float64
	IdealPollenStore_g       float64
	EnergyNeedToday_kJ       float64
	PollenNeedLarvaeToday_g  float64
	SummedIncubationToday_kJ float64
}

// ColonyStimuli holds the task stimuli of a colony (ActivityProc, step 5).
type ColonyStimuli struct {
	StimEgglaying      float64
	StimNectarForaging float64
	StimNursing        float64
	StimPollenForaging float64
}

// ColonyCounts holds the numbers of bees, deaths and production of a colony.
type ColonyCounts struct {
	LarvaWorkerRatio        float64
	AllEggs                 float64
	AllLarvae               float64
	AllPupae                float64
	AllAdults               float64
	AllAdultWorkers         float64
	AllAdultMales           float64
	BroodDeathBadger        float64
	BroodDeathEndSeason     float64
	BroodDeathsCP           float64
	BroodDeathsEnergyStores float64
	BroodDeathsNoAdults     float64
	EggDeathsIncubation     float64
	LarvaDeathsIncubation   float64
	LarvaDeathsWeight       float64
	PupaDeathsIncubation    float64
	TotalEggsProduced       float64
	TotalAdultsProduced     float64
}

// ColonyRange holds the foodsources within the foraging range of a colony [PRNG], [FRNG].
type ColonyRange struct {
	AllPatchesInRangeList                             []ecs.Entity // all foodsources (layers) in range, random order (finding 29)
	MasterpatchesWithPollenlayersInFlowerAndRangeList []ecs.Entity
	MasterpatchesWithNectarlayersInFlowerAndRangeList []ecs.Entity
}
