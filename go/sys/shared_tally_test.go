package sys

import (
	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/event"

	"github.com/mlange-42/ark/ecs"
)

// tally counts what the model reports through the removal of entities and through the custom events of package
// event, as obs.Counters does. The model does not count it itself.
type tally struct {
	beesDied      int // sum of Number of the removed bees
	coloniesDied  int
	trips         int // event.ForagingTrip, sum of Number
	foragersDied  int // event.DiedForaging, sum of Number
	adultsEmerged int // event.AdultEmerged, sum of Number
	queensEmerged int // event.AdultEmerged of queens, sum of Number
	nectarVisits  int // event.Collected of bees bringing nectar, sum of Number
	pollenVisits  int // event.Collected of bees bringing pollen, sum of Number
}

// newTally registers the observers of a new tally.
func newTally(w *ecs.World) *tally {
	c := &tally{}
	ecs.Observe1[comp.Bee](ecs.OnRemoveEntity).
		Do(func(e ecs.Entity, b *comp.Bee) { c.beesDied += b.Number }).
		Register(w)
	ecs.Observe(ecs.OnRemoveEntity).With(ecs.C[comp.Colony]()).
		Do(func(e ecs.Entity) { c.coloniesDied++ }).
		Register(w)
	ecs.Observe1[comp.Bee](event.ForagingTrip).
		Do(func(e ecs.Entity, b *comp.Bee) { c.trips += b.Number }).
		Register(w)
	ecs.Observe1[comp.Bee](event.DiedForaging).
		Do(func(e ecs.Entity, b *comp.Bee) { c.foragersDied += b.Number }).
		Register(w)
	ecs.Observe1[comp.Bee](event.AdultEmerged).
		Do(func(e ecs.Entity, b *comp.Bee) {
			c.adultsEmerged += b.Number
			if b.Caste == caste.Queen {
				c.queensEmerged += b.Number
			}
		}).
		Register(w)
	ecs.Observe1[comp.Bee](event.Collected).
		Do(func(e ecs.Entity, b *comp.Bee) {
			switch b.Activity {
			case activity.BringingNectar:
				c.nectarVisits += b.Number
			case activity.BringingPollen:
				c.pollenVisits += b.Number
			default:
				panic("event.Collected on a bee that brings nothing")
			}
		}).
		Register(w)
	return c
}
