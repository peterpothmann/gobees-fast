package event

import "github.com/mlange-42/ark/ecs"

var registry ecs.EventRegistry

var (
	// ForagingTrip is emitted on a bee at the start of each foraging trip. The trip counts Bee.Number
	// times (a cohort flies as a squadron).
	ForagingTrip = registry.NewEventType()
	// DiedForaging is emitted on a bee that dies on a foraging trip, just before she is removed.
	DiedForaging = registry.NewEventType()
	// QueenEmerged is emitted on a queen pupa that has become an adult.
	QueenEmerged = registry.NewEventType()
)
