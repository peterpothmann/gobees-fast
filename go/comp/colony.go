package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// A colony entity has the components Position, Colony, ColonyStores, ColonyStimuli, ColonyCounts and
// ColonyRange. NetLogo: colonies-own (56 variables), split by topic (trans/data-model.md).
// Not translated: masterpatchesInRangeList (never read, finding 23, see trans/netlogo-findings.md#finding-23).

// Colony holds identity, phenology and flags of a colony.
type Colony struct {
	Species              int      // index in params.Input.Species (step F11). NetLogo: speciesIDcolony
	QueenProduction      bool     // queenProduction?
	QueenProductionDate  day.Date // unset until set (NetLogo: NotSetHigh)
	SwitchPointDate      day.Date // unset until set (NetLogo: NotSetHigh)
	CompetitionPointDate day.Date // unset until set (NetLogo: NotSetHigh)
	EusocialPhaseDate    day.Date // unset until set (NetLogo: NotSetHigh)
}

// ColonyStores holds the stores and the daily needs of a colony.
type ColonyStores struct {
	Energy                float64 // energy store [kJ]. NetLogo: energyStore_kJ
	Pollen                float64 // pollen store [g]. NetLogo: pollenStore_g
	IdealEnergy           float64 // [kJ]. NetLogo: idealEnergyStore_kJ
	IdealPollen           float64 // [g]. NetLogo: idealPollenStore_g
	EnergyNeedToday       float64 // [kJ]. NetLogo: energyNeedToday_kJ
	LarvaePollenNeedToday float64 // [g]. NetLogo: pollenNeedLarvaeToday_g
	IncubationToday       float64 // incubation provided today [kJ]. NetLogo: summedIncubationToday_kJ
}

// ColonyStimuli holds the task stimuli of a colony (ActivityProc, step 5). NetLogo: stimEgglaying,
// stimNectarForaging, stimNursing, stimPollenForaging.
type ColonyStimuli struct {
	Egglaying      float64
	NectarForaging float64
	Nursing        float64
	PollenForaging float64
}

// ColonyCounts holds the numbers of bees of a colony, set in [UCOL-06].
// The brood death counters (broodDeath..., ...DeathsIncubation, larvaDeathsWeight), totalEggsProduced and
// totalAdultsProduced: read only by the egg accounting [CHK-04], allAdultMales: read only by [CHK-03]. Removed in
// step F4 together with the checks.
// NetLogo: allEggs, allLarvae, allPupae, allAdults, allAdultWorkers.
type ColonyCounts struct {
	LarvaWorkerRatio float64
	Eggs             int
	Larvae           int
	Pupae            int
	Adults           int
	AdultWorkers     int
}

// ColonyRange holds the foodsources within the foraging range of a colony [PRNG], [FRNG].
type ColonyRange struct {
	Layers        []ecs.Entity // all foodsources (layers) in range, random order (finding 29). NetLogo: allPatchesInRangeList
	PollenPatches []ecs.Entity // masterpatches in range with a pollen layer in flower. NetLogo: masterpatchesWithPollenlayersInFlowerAndRangeList
	NectarPatches []ecs.Entity // masterpatches in range with a nectar layer in flower. NetLogo: masterpatchesWithNectarlayersInFlowerAndRangeList
}
