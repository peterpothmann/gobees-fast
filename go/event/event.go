package event

import "github.com/mlange-42/ark/ecs"

var registry ecs.EventRegistry

var (
	// ForagingTrip is emitted on a bee at the start of each foraging trip. The trip counts Bee.Number
	// times (a cohort flies as a squadron).
	ForagingTrip = registry.NewEventType()
	// DiedForaging is emitted on a bee that dies on a foraging trip, just before she is removed.
	DiedForaging = registry.NewEventType()
	// AdultEmerged is emitted on a pupa of any caste that has become an adult. Bee.Caste tells the caste.
	AdultEmerged = registry.NewEventType()
	// Collected is emitted on a forager that has taken a load from her foodsource (Forager.CurrentFoodsource). Her
	// Bee.Activity is then BringingNectar or BringingPollen. The visit counts Bee.Number times (a cohort takes
	// Number loads at once).
	Collected = registry.NewEventType()
)
