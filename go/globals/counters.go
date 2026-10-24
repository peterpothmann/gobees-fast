package globals

import (
	"gobees/comp"

	"github.com/mlange-42/ark/ecs"
)

// Counters holds global counters of the run (NetLogo globals).
type Counters struct {
	TotalBeesEverDied         float64 // [DIE-01]
	TotalBeesEverProduced     float64 // [EGGP-02], read by [CHK-01]
	TotalColoniesEverProduced float64 // [CCOL-08]
	ColonyDeathsEndSeason     float64 // [UCOL-11], read by [CHK-04]
	ColonyDeathsNoBees        float64 // [UCOL-11], read by [CHK-04]

	TotalAdultsEverProduced      float64 // [DEVP-02] to [DEVP-04], output only (BehaviorSpace)
	TotalAdultMalesEverProduced  float64 // [DEVP-03], output only (BehaviorSpace)
	TotalAdultQueensEverProduced float64 // [DEVP-04], output only (BehaviorSpace)
	// QueensProducingColoniesList holds every colony in which a queen emerged, in the order of the first
	// emergence [DEVP-04]. Never read. SIMPLIFY: not needed.
	QueensProducingColoniesList []ecs.Entity
}

// ActiveBees is the daily activity queue: the active bees sorted by personalTime_s.
// NetLogo: ActiveBeesSortedList. Filled in the morning [MORN-08] and used by the activity loop (both step 5).
// DieProc removes dying bees from it [DIE-01].
type ActiveBees struct {
	List []ecs.Entity
}

// Daytime is the time of day of the event loop: the personalTime_s of the bee whose turn it is [GO-09].
// Read by the foraging stimuli [STNE-03], [STPO-03].
type Daytime struct {
	Daytime_s float64
}

// DeadColony is the record of a dead colony: the colony's data at death, as NetLogo keeps it in deadCols-own,
// plus colonyDeathDay [DIE-02]. Unlike deadCols it keeps SpeciesNameColony (finding 9, see
// trans/netlogo-findings.md#finding-9).
type DeadColony struct {
	Who            int64
	Position       comp.Position
	Colony         comp.Colony
	Stores         comp.ColonyStores
	Counts         comp.ColonyCounts
	ColonyDeathDay float64 // ticks
}

// DeadColonies records the dead colonies (decision E in trans/data-model.md). NetLogo: the agentset deadCols.
// Output only, no model procedure reads it.
type DeadColonies struct {
	List []DeadColony
}
