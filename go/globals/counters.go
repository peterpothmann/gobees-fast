package globals

import (
	"github.com/mlange-42/ark/ecs"
)

// The NetLogo globals that only the output or the checks read are not part of the model (step F4): the counters of
// bees and colonies ([DIE-01], [EGGP-02], [CCOL-08], [UCOL-11], [DEVP-04], [SEAS-03]), the record of the dead
// colonies (deadCols, [DIE-02], [UCOL-12]) and the run summaries. The observers in package obs count them from Ark
// events (package event).

// ActiveBees is the daily activity queue: the active bees sorted by personalTime_s.
// NetLogo: ActiveBeesSortedList. Filled in the morning [MORN-08] and used by the activity loop (both step 5).
// DieProc removes dying bees from it [DIE-01].
type ActiveBees struct {
	List []ecs.Entity
}

// Daytime is the time of day of the event loop: the personalTime_s of the bee whose turn it is [GO-09].
// Read by the foraging stimuli [STNE-03], [STPO-03].
type Daytime struct {
	Time float64 // [s]. NetLogo: Daytime_s
}
