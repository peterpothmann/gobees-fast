package sys

import (
	"fmt"
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
// NetLogo: DieProc for breed bees. Colonies [DIE-02]: colonyDeath. Badgers [DIE-03] follow in step 7.
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

// colonyDeath removes colonies from the world [DIE-02].
//
// NetLogo turns a dead colony into a deadCol. gobees records its data and deletes the entity (decision E in
// trans/data-model.md, finding 9, see trans/netlogo-findings.md#finding-9). The colony must have no bees left:
// NetLogo calls DieProc for a colony only in [UCOL-11], when no bee has its colonyID.
// Must not be called while a query is open.
//
// NetLogo: DieProc for breed colonies. Not translated: the store bars and the display (GUI).
type colonyDeath struct {
	ticks     *globals.Ticks
	dead      *globals.DeadColonies
	colonyMap *ecs.Map5[comp.Who, comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyCounts]
}

func newColonyDeath(w *ecs.World) colonyDeath {
	return colonyDeath{
		ticks:     ecs.GetResource[globals.Ticks](w),
		dead:      ecs.GetResource[globals.DeadColonies](w),
		colonyMap: ecs.NewMap5[comp.Who, comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyCounts](w),
	}
}

// DieProc records the colony e and removes it. causeOfDeath is not used, as in NetLogo.
func (d *colonyDeath) DieProc(w *ecs.World, e ecs.Entity, causeOfDeath string) {
	// [DIE-02] colonyDeathDay = ticks
	who, pos, colony, stores, counts := d.colonyMap.Get(e)
	record := globals.DeadColony{Who: who.Who, Position: *pos, Colony: *colony, Stores: *stores, Counts: *counts,
		ColonyDeathDay: float64(d.ticks.Ticks)}
	// [CHK-04] egg accounting, here at death instead of on day 364 (decision E). The counters of a dead colony
	// do not change any more.
	if eggcheck := eggCheck(&record.Counts); eggcheck != 0 {
		panic(fmt.Sprintf("Dead Colony %d cannot account for all eggs produced %v", who.Who, eggcheck))
	}
	d.dead.List = append(d.dead.List, record)
	w.RemoveEntity(e)
}

// eggCheck returns the eggs of a colony that are neither adults nor counted in a brood death counter [CHK-04].
// It is 0 for a dead colony.
func eggCheck(c *comp.ColonyCounts) float64 {
	return c.TotalEggsProduced - (c.TotalAdultsProduced +
		c.BroodDeathsEnergyStores +
		c.BroodDeathsNoAdults +
		c.BroodDeathsCP +
		c.BroodDeathEndSeason +
		c.BroodDeathBadger +
		c.EggDeathsIncubation +
		c.LarvaDeathsIncubation +
		c.LarvaDeathsWeight +
		c.PupaDeathsIncubation)
}
