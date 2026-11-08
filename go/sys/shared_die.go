package sys

import (
	"log"

	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// beeDeath removes bees from the world.
//
// As NetLogo's die ends the dying agent's code at once, the caller must not use the bee afterwards.
// Must not be called while a query is open (Ark locks the world during a query).
// The observers count the removed bees (ecs.OnRemoveEntity, obs.Counters).
//
// NetLogo: DieProc for breed bees. Colonies: colonyDeath. Badgers: badgerDeath.
type beeDeath struct {
	active *globals.ActiveBees
}

func newBeeDeath(w *ecs.World) beeDeath {
	return beeDeath{
		active: ecs.GetResource[globals.ActiveBees](w),
	}
}

// die removes the bee e. causeOfDeath is not used, as in NetLogo (only printed for badgers).
func (d *beeDeath) die(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	d.active.Remove(e)
	w.RemoveEntity(e)
}

// colonyDeath removes colonies from the world.
//
// The entity is deleted. NetLogo keeps a dead colony as a deadCol, which only the output reads. The observers count
// the dead colonies (obs.Counters). The colony must have no bees left: a colony dies only in colonyUpdate, when no
// bee has its colony.
// Must not be called while a query is open.
//
// NetLogo: DieProc for breed colonies.
type colonyDeath struct{}

// die removes the colony e. causeOfDeath is not used, as in NetLogo.
func (colonyDeath) die(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	w.RemoveEntity(e)
}

// badgerDeath removes badgers from the world.
// Must not be called while a query is open.
//
// NetLogo: DieProc for breed badgers.
type badgerDeath struct{}

// die prints the cause and removes the badger e (NetLogo: output-show causeOfDeath, die).
func (badgerDeath) die(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	log.Print(causeOfDeath)
	w.RemoveEntity(e)
}
