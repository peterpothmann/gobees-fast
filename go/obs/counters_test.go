package obs

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestCountersEvents(t *testing.T) {
	w := newWorld(t)
	c := Counters{}
	c.Initialize(w)

	// removed bees count with their number new colonies queens emerged
	colonies := ecs.NewMap1[comp.Colony](w)
	col := colonies.NewEntity(&comp.Colony{})
	colonies.NewEntity(&comp.Colony{})
	squadron := addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Worker, Number: 5})
	queen := addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 3})
	emit(w, event.QueenEmerged, queen)
	emit(w, event.ForagingTrip, squadron)
	emit(w, event.DiedForaging, squadron)
	w.RemoveEntity(squadron)
	w.RemoveEntity(queen)
	if c.TotalBeesEverDied != 8 || c.TotalColoniesEverProduced != 2 || c.TotalAdultQueensEverProduced != 3 ||
		c.TotalForagersEverDiedForaging != 5 {
		t.Errorf("died %v, colonies %v, queens %v, foragers died %v; want 8, 2, 3, 5", c.TotalBeesEverDied,
			c.TotalColoniesEverProduced, c.TotalAdultQueensEverProduced, c.TotalForagersEverDiedForaging)
	}
	// the trips of a day show after its Update
	if c.TotalForagingTripsToday != 0 {
		t.Errorf("trips before the end of the day: %v", c.TotalForagingTripsToday)
	}
	c.Update(w)
	if c.TotalForagingTripsToday != 5 {
		t.Errorf("trips %v, want 5", c.TotalForagingTripsToday)
	}
	c.Update(w)
	if c.TotalForagingTripsToday != 0 {
		t.Errorf("trips of the next day %v, want 0", c.TotalForagingTripsToday)
	}

	// the dead colonies (on day 2, day 1 removes them)
	ecs.GetResource[globals.Calendar](w).Day = 2
	w.RemoveEntity(col)
	c.Update(w)
	if c.DeadColonies != 1 {
		t.Errorf("%d dead colonies, want 1", c.DeadColonies)
	}
}

func TestCountersDeadColoniesNewYear(t *testing.T) {
	// on 1 January the dead colonies are removed unless KeepDeadColonies?, also those of that day
	for _, keep := range []bool{false, true} {
		w := newWorld(t)
		ecs.GetResource[params.Output](w).KeepDeadColonies = keep
		calendar := ecs.GetResource[globals.Calendar](w)
		c := Counters{}
		c.Initialize(w)
		colonies := ecs.NewMap1[comp.Colony](w)

		calendar.Day = 365
		w.RemoveEntity(colonies.NewEntity(&comp.Colony{}))
		c.Update(w)
		calendar.Day = 1
		w.RemoveEntity(colonies.NewEntity(&comp.Colony{}))
		c.Update(w)
		want := 0
		if keep {
			want = 2
		}
		if c.DeadColonies != want {
			t.Errorf("KeepDeadColonies %v: %d dead colonies on 1 January, want %d", keep, c.DeadColonies, want)
		}
		calendar.Day = 2
		w.RemoveEntity(colonies.NewEntity(&comp.Colony{}))
		c.Update(w)
		if c.DeadColonies != want+1 {
			t.Errorf("KeepDeadColonies %v: %d dead colonies on 2 January, want %d", keep, c.DeadColonies, want+1)
		}
	}
}

func TestCountersHibernatingQueensRemoved(t *testing.T) {
	// the planned number from the hibernating queens of the evening before day 1, added on day 1
	w := newWorld(t)
	ecs.GetResource[params.Parameters](w).MaxHibernatingQueens = 10
	calendar := ecs.GetResource[globals.Calendar](w)
	calendar.Day = 0 // Setup
	hibernating := addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 12, Activity: activity.Hibernate})
	addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 50, Activity: activity.Resting})
	c := Counters{}
	c.Initialize(w)
	if c.TotalHibernatingQueensEverRemoved != 0 {
		t.Fatalf("after Setup: %v, want 0", c.TotalHibernatingQueensEverRemoved)
	}
	calendar.Day = 1
	c.Update(w)
	if c.TotalHibernatingQueensEverRemoved != 2 {
		t.Errorf("day 1 of year 1: %v, want 2", c.TotalHibernatingQueensEverRemoved)
	}

	// year 2: 25 hibernating queens on the evening of day 365, 15 planned
	ecs.NewMap1[comp.Bee](w).Get(hibernating).Number = 25
	for _, d := range []day.OfYear{364, 365} {
		calendar.Day = d
		c.Update(w)
		if c.TotalHibernatingQueensEverRemoved != 2 {
			t.Errorf("day %v: %v, want 2", d, c.TotalHibernatingQueensEverRemoved)
		}
	}
	calendar.Day = 1
	c.Update(w)
	if c.TotalHibernatingQueensEverRemoved != 17 {
		t.Errorf("day 1 of year 2: %v, want 17", c.TotalHibernatingQueensEverRemoved)
	}

	// fewer queens than the limit: nothing planned
	ecs.NewMap1[comp.Bee](w).Get(hibernating).Number = 4
	calendar.Day = 365
	c.Update(w)
	calendar.Day = 1
	c.Update(w)
	if c.TotalHibernatingQueensEverRemoved != 17 {
		t.Errorf("day 1 of year 3: %v, want 17", c.TotalHibernatingQueensEverRemoved)
	}
}

func TestCountersTimeToExtinction(t *testing.T) {
	// the first day without colonies and bees
	w := newWorld(t)
	ticks := ecs.GetResource[globals.Ticks](w)
	bee := addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 1})
	c := Counters{}
	c.Initialize(w)
	c.Update(w)
	if c.TimeToExtinction != 0 {
		t.Fatalf("population alive: TimeToExtinction %v", c.TimeToExtinction)
	}
	w.RemoveEntity(bee)
	ticks.Ticks = 202
	c.Update(w)
	ticks.Ticks = 203
	c.Update(w)
	if c.TimeToExtinction != 202 {
		t.Errorf("TimeToExtinction %v, want 202", c.TimeToExtinction)
	}
}
