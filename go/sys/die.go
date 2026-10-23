package sys

import (
	"slices"

	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// beeDeath removes bees from the world [DIE-01].
//
// NetLogo's die ends the dying agent's code at once: the caller must not use the bee afterwards
// (finding 28, see trans/netlogo-findings.md#finding-28).
// Must not be called while a query is open (Ark locks the world during a query).
//
// NetLogo: DieProc for breed bees. Colonies [DIE-02] and badgers [DIE-03] follow in later steps.
type beeDeath struct {
	counters *globals.Counters
	active   *globals.ActiveBees
	beeMap   *ecs.Map1[comp.Bee]
}

func newBeeDeath(w *ecs.World) beeDeath {
	return beeDeath{
		counters: ecs.GetResource[globals.Counters](w),
		active:   ecs.GetResource[globals.ActiveBees](w),
		beeMap:   ecs.NewMap1[comp.Bee](w),
	}
}

// DieProc removes the bee e. causeOfDeath is not used, as in NetLogo (only printed for badgers).
func (d *beeDeath) DieProc(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	// [DIE-01]
	d.counters.TotalBeesEverDied += d.beeMap.Get(e).Number
	d.active.List = slices.DeleteFunc(d.active.List, func(x ecs.Entity) bool { return x == e })
	w.RemoveEntity(e)
}
