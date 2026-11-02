// Package event holds the custom Ark event types of the model. The model emits them where the output needs a value
// that the state does not hold, e.g. the number of foraging trips. No system observes them: they are for the
// observers of the output (package obs) only. Without a registered observer, emitting an event costs one check.
//
// Deaths and new colonies need no custom event: the observers use Ark's OnRemoveEntity, which is emitted before the
// entity is removed, and OnCreateEntity.
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
