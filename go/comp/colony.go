package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// A colony entity has the components Who, Position, Colony, ColonyStores, ColonyStimuli, ColonyCounts and
// ColonyRange. NetLogo: colonies-own (56 variables), split by topic (trans/data-model.md).
// Not translated: masterpatchesInRangeList (never read, finding 23, see trans/netlogo-findings.md#finding-23).

// Colony holds identity, phenology and flags of a colony.
type Colony struct {
	SpeciesIDcolony      ecs.Entity
	QueenProduction      bool     // queenProduction?
	QueenProductionDate  day.Date // unset until set (NetLogo: NotSetHigh)
	SwitchPointDate      day.Date // unset until set (NetLogo: NotSetHigh)
	CompetitionPointDate day.Date // unset until set (NetLogo: NotSetHigh)
	EusocialPhaseDate    day.Date // unset until set (NetLogo: NotSetHigh)
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

// ColonyCounts holds the numbers of bees of a colony, set in [UCOL-06].
// The brood death counters (broodDeath..., ...DeathsIncubation, larvaDeathsWeight), totalEggsProduced and
// totalAdultsProduced: read only by the egg accounting [CHK-04], allAdultMales: read only by [CHK-03]. Removed in
// step F4 together with the checks.
type ColonyCounts struct {
	LarvaWorkerRatio float64
	AllEggs          int
	AllLarvae        int
	AllPupae         int
	AllAdults        int
	AllAdultWorkers  int
}

// ColonyRange holds the foodsources within the foraging range of a colony [PRNG], [FRNG].
type ColonyRange struct {
	AllPatchesInRangeList                             []ecs.Entity // all foodsources (layers) in range, random order (finding 29)
	MasterpatchesWithPollenlayersInFlowerAndRangeList []ecs.Entity
	MasterpatchesWithNectarlayersInFlowerAndRangeList []ecs.Entity
}
