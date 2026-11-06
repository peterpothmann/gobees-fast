package obs

import (
	"gobees/comp"
	"gobees/enum/activity"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// Counters counts what happened during the run and is not held by the state: the NetLogo globals that only the
// output reads, and further counters of gobees-fast. It registers Ark observers in Initialize: OnRemoveEntity for
// dead bees and colonies, OnCreateEntity for new colonies, and the custom events of package event.
//
// Initialize must run after Setup, e.g. as a system added after the model systems, or in the Initialize of an
// observer of reporter.CSV. Setup creates no colony, kills no bee and makes no foraging trip, so no event is missed.
// Update must run once at the end of every day, after all model systems. Counters is an app system and is also
// used by other observers (DailyOutputExtended).
type Counters struct {
	TotalBeesEverDied            float64 // [DIE-01], sum of Number of the removed bees
	TotalColoniesEverProduced    float64 // [CCOL-08]
	TotalAdultQueensEverProduced float64 // [DEVP-04]
	// TotalHibernatingQueensEverRemoved adds up the planned number of queens to kill on 1 January [SEAS-03], not
	// the removed number: max(0, hibernating queens of the previous evening - MaxHibernatingQueens). Added at the
	// end of day 1, as NetLogo adds it in the morning of day 1.
	TotalHibernatingQueensEverRemoved int
	// TotalForagersEverDiedForaging adds up the bees that died on a foraging trip [COST-01], [COST-02]. Not in
	// NetLogo.
	TotalForagersEverDiedForaging float64
	// TotalForagingTripsToday is the number of foraging trips of the day that ended last [FOR-01] (NetLogo resets
	// it every morning in CheckNumbersProc [CHK-02]).
	TotalForagingTripsToday float64
	// DeadColonies is the number of dead colonies that NetLogo keeps as deadCols [DIE-02]: since the last 1 January,
	// or since the start with the output option KeepDeadColonies? [UCOL-12] (params.Output, step F13).
	DeadColonies int
	// TimeToExtinction is the ticks of the first day without colonies and bees, 0 before [GO-20].
	TimeToExtinction float64

	ticks      *globals.Ticks
	calendar   *globals.Calendar
	parameters *params.Parameters
	output     *params.Output

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]

	tripsToday        float64 // trips since the last Update
	deadColoniesToday int     // colonies removed since the last Update
	queensToBeRemoved int     // planned number of [SEAS-03] for the next day 1
}

func (c *Counters) Initialize(w *ecs.World) {
	c.ticks = ecs.GetResource[globals.Ticks](w)
	c.calendar = ecs.GetResource[globals.Calendar](w)
	c.parameters = ecs.GetResource[params.Parameters](w)
	c.output = ecs.GetResource[params.Output](w)
	c.beeFilter = ecs.NewFilter1[comp.Bee](w)
	c.colonyFilter = ecs.NewFilter1[comp.Colony](w)

	ecs.Observe1[comp.Bee](ecs.OnRemoveEntity).
		Do(func(e ecs.Entity, b *comp.Bee) { c.TotalBeesEverDied += float64(b.Number) }).
		Register(w)
	ecs.Observe(ecs.OnCreateEntity).With(ecs.C[comp.Colony]()).
		Do(func(e ecs.Entity) { c.TotalColoniesEverProduced++ }).
		Register(w)
	ecs.Observe(ecs.OnRemoveEntity).With(ecs.C[comp.Colony]()).
		Do(func(e ecs.Entity) { c.deadColoniesToday++ }).
		Register(w)
	ecs.Observe1[comp.Bee](event.ForagingTrip).
		Do(func(e ecs.Entity, b *comp.Bee) { c.tripsToday += float64(b.Number) }).
		Register(w)
	ecs.Observe1[comp.Bee](event.DiedForaging).
		Do(func(e ecs.Entity, b *comp.Bee) { c.TotalForagersEverDiedForaging += float64(b.Number) }).
		Register(w)
	ecs.Observe1[comp.Bee](event.QueenEmerged).
		Do(func(e ecs.Entity, b *comp.Bee) { c.TotalAdultQueensEverProduced += float64(b.Number) }).
		Register(w)

	c.planQueenRemoval() // Setup is the evening before day 1 of the first year
}

func (c *Counters) Update(w *ecs.World) {
	c.TotalForagingTripsToday, c.tripsToday = c.tripsToday, 0

	// [UCOL-12] colonyUpdate removes the dead colonies on day 1 after the colonies of that day have died
	if c.calendar.Day == 1 && !c.output.KeepDeadColonies {
		c.DeadColonies = 0
	} else {
		c.DeadColonies += c.deadColoniesToday
	}
	c.deadColoniesToday = 0

	// [SEAS-03]
	if c.calendar.Day == 1 {
		c.TotalHibernatingQueensEverRemoved += c.queensToBeRemoved
		c.queensToBeRemoved = 0
	}
	if c.calendar.Day == 365 {
		c.planQueenRemoval()
	}

	// [GO-20] count Colonies + count Bees = 0
	if c.TimeToExtinction <= 0 {
		colonies := c.colonyFilter.Query()
		n := colonies.Count()
		colonies.Close()
		bees := c.beeFilter.Query()
		n += bees.Count()
		bees.Close()
		if n == 0 {
			c.TimeToExtinction = float64(c.ticks.Ticks)
		}
	}
}

func (c *Counters) Finalize(w *ecs.World) {}

// planQueenRemoval sets the number of queens that [SEAS-03] plans to kill on the next day 1, from the hibernating
// queens of this evening (TotalHibernatingQueens of OutputDailyProc).
func (c *Counters) planQueenRemoval() {
	hibernating := 0
	query := c.beeFilter.Query()
	for query.Next() {
		if b := query.Get(); b.Activity == activity.Hibernate {
			hibernating += b.Number
		}
	}
	c.queensToBeRemoved = max(0, hibernating-c.parameters.MaxHibernatingQueens)
}
