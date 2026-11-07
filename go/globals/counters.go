package globals

import (
	"github.com/mlange-42/ark/ecs"
)

// ActiveBees is the daily activity queue: the active bees sorted by their personal time. Filled in the morning and
// used by the activity loop. A dying bee is removed from it. NetLogo: ActiveBeesSortedList.
type ActiveBees struct {
	List []ecs.Entity
}

// Daytime is the time of day of the activity loop: the personal time of the bee whose turn it is. Read by the
// foraging stimuli.
type Daytime struct {
	Time float64 // [s]. NetLogo: Daytime_s
}
