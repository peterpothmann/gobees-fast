package sys

import (
	"log"
	"slices"

	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// beeDeath removes bees from the world [DIE-01].
//
// NetLogo's die ends the dying agent's code at once: the caller must not use the bee afterwards
// (finding 28, see trans/netlogo-findings.md#finding-28).
// Must not be called while a query is open (Ark locks the world during a query).
// totalBeesEverDied: output and [CHK-01] only, removed in step F4. The observers count the removed bees
// (ecs.OnRemoveEntity, obs.Counters).
//
// NetLogo: DieProc for breed bees. Colonies [DIE-02]: colonyDeath. Badgers [DIE-03]: badgerDeath.
type beeDeath struct {
	active *globals.ActiveBees
}

func newBeeDeath(w *ecs.World) beeDeath {
	return beeDeath{
		active: ecs.GetResource[globals.ActiveBees](w),
	}
}

// DieProc removes the bee e. causeOfDeath is not used, as in NetLogo (only printed for badgers).
func (d *beeDeath) DieProc(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	// [DIE-01]
	d.active.List = slices.DeleteFunc(d.active.List, func(x ecs.Entity) bool { return x == e })
	w.RemoveEntity(e)
}

// colonyDeath removes colonies from the world [DIE-02].
//
// NetLogo turns a dead colony into a deadCol, which only the output reads. gobees deletes the entity (decision E in
// trans/data-model.md, finding 9, see trans/netlogo-findings.md#finding-9). The observers count the dead colonies
// (obs.Counters). The colony must have no bees left: NetLogo calls DieProc for a colony only in [UCOL-11], when no
// bee has its colonyID.
// The record of the dead colony (colonyDeathDay) and the egg accounting at its death [CHK-04]: removed in step F4.
// Must not be called while a query is open.
//
// NetLogo: DieProc for breed colonies. Not translated: the store bars and the display (GUI).
type colonyDeath struct{}

// DieProc removes the colony e. causeOfDeath is not used, as in NetLogo.
func (colonyDeath) DieProc(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	// [DIE-02]
	w.RemoveEntity(e)
}

// badgerDeath removes badgers from the world [DIE-03].
// Must not be called while a query is open.
//
// NetLogo: DieProc for breed badgers.
type badgerDeath struct{}

// DieProc prints the cause and removes the badger e (NetLogo: output-show causeOfDeath, die).
func (badgerDeath) DieProc(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	// [DIE-03]
	log.Print(causeOfDeath)
	w.RemoveEntity(e)
}
