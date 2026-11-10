package comp

import (
	"gobees/day"

	"github.com/mlange-42/ark/ecs"
)

// A colony entity has the components Position, Colony, ColonyStores, ColonyStimuli, ColonyCounts and
// ColonyRange. NetLogo: colonies-own, split by topic.

// Colony holds identity, phenology and flags of a colony.
type Colony struct {
	Species              int      // index in params.Input.Species. NetLogo: speciesIDcolony
	QueenProduction      bool     // queenProduction?
	QueenProductionDate  day.Date // unset until set
	SwitchPointDate      day.Date // unset until set
	CompetitionPointDate day.Date // unset until set
	EusocialPhaseDate    day.Date // unset until set
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

// ColonyStimuli holds the task stimuli of a colony. NetLogo: stimEgglaying, stimNectarForaging, stimNursing,
// stimPollenForaging.
type ColonyStimuli struct {
	Egglaying      float64
	NectarForaging float64
	Nursing        float64
	PollenForaging float64
}

// ColonyCounts holds the numbers of bees of a colony, set by the colony update in the morning.
// NetLogo: allEggs, allLarvae, allPupae, allAdults, allAdultWorkers.
type ColonyCounts struct {
	LarvaWorkerRatio float64
	Eggs             int
	Larvae           int
	Pupae            int
	Adults           int
	AdultWorkers     int
}

// ColonyRange holds the foodsources within the foraging range of a colony.
type ColonyRange struct {
	Patches       []ecs.Entity // all flower patches in range, in storage order. NetLogo: allPatchesInRangeList (all layers, random order)
	PollenPatches []ecs.Entity // flower patches in range with a pollen layer in flower. NetLogo: masterpatchesWithPollenlayersInFlowerAndRangeList
	NectarPatches []ecs.Entity // flower patches in range with a nectar layer in flower. NetLogo: masterpatchesWithNectarlayersInFlowerAndRangeList

	// The detection probability of each flower patch of PollenPatches and NectarPatches, in the same order. Colony,
	// flower patches and their radius do not move, so the search computes it once per list (forager.detections).
	// nil: not computed yet. Whoever sets PollenPatches or NectarPatches sets the matching list to nil.
	PollenDetection []float64
	NectarDetection []float64
}
