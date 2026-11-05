package event

import "github.com/mlange-42/ark/ecs"

var registry ecs.EventRegistry

var (
	// ForagingTrip is emitted on a bee at the start of each foraging trip [FOR-01]. The trip counts Bee.Number
	// times (a cohort flies as a squadron).
	ForagingTrip = registry.NewEventType()
	// DiedForaging is emitted on a bee that dies on a foraging trip, just before DieProc [COST-01], [COST-02].
	// Not in NetLogo.
	DiedForaging = registry.NewEventType()
	// QueenEmerged is emitted on a queen pupa that has become an adult [DEVP-04].
	QueenEmerged = registry.NewEventType()
)
