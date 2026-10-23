package globals

import "github.com/mlange-42/ark/ecs"

// Counters holds global counters of the run (NetLogo globals).
type Counters struct {
	TotalBeesEverDied         float64 // [DIE-01]
	TotalColoniesEverProduced float64 // [CCOL-08]
}

// ActiveBees is the daily activity queue: the active bees sorted by personalTime_s.
// NetLogo: ActiveBeesSortedList. Filled in the morning (step 4), used by the activity loop (step 5).
// DieProc removes dying bees from it [DIE-01].
type ActiveBees struct {
	List []ecs.Entity
}
