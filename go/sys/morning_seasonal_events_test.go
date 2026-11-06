package sys

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// seasWorld is a colonyWorld with the brood totals of the previous morning (broodTotals), set by hand, and the
// seasonalEvents system. The totals of the previous evening (the daily output of NetLogo) are counted from the bees.
func seasWorld(t *testing.T, d day.OfYear) (*colonyWorld, *seasonalEvents, *globals.BroodTotals) {
	t.Helper()
	cw := newColonyWorld(t, 1)
	cw.ticks.Ticks = day.Abs(d) // the first year
	cw.cal.Day = d
	totals := &globals.BroodTotals{}
	ecs.AddResource(cw.w, totals)
	s := &seasonalEvents{}
	s.Initialize(cw.w)
	return cw, s, totals
}

// addHibernatingQueen adds a mated, hibernating queen (cohort) without colony.
func (cw *colonyWorld) addHibernatingQueen(number int) ecs.Entity {
	e := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, number, 0)
	_, b, _, _, _, _ := cw.beeMap.Get(e)
	b.Activity = activity.Hibernate
	return e
}

func TestSeasonalEventsSeasonStop(t *testing.T) {
	// on Day seasonStop (305 for B_terrestris) every bee of the species that does not hibernate dies
	for _, d := range []day.OfYear{304, 305} {
		cw, s, _ := seasWorld(t, d)
		col := cw.addColony(nil)
		cw.addBee(col, stage.Adult, caste.Queen, 1, 60)
		cw.addBee(col, stage.Adult, caste.Worker, 10, 5)
		cw.addBee(col, stage.Adult, caste.Male, 2, 1)
		cw.addBee(col, stage.Egg, caste.Undefined, 5, 0)
		cw.addBee(col, stage.Larva, caste.Worker, 4, 0)
		cw.addYoungQueen(col, 1)                                             // has not left yet
		searching := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0) // without colony
		hibernating := cw.addHibernatingQueen(3)

		s.Update(cw.w)
		alive := countEntities(ecs.NewFilter1[comp.Bee](cw.w))
		if d == 304 {
			if alive != 8 {
				t.Errorf("day 304: %d bee agents, want 8", alive)
			}
			continue
		}
		if alive != 1 || !cw.w.Alive(hibernating) || cw.w.Alive(searching) {
			t.Errorf("day 305: %d bee agents alive, want only the hibernating queen", alive)
		}
		// the colony dies later this morning in colonyUpdate
		cw.sys.Update(cw.w)
		if cw.w.Alive(col) {
			t.Error("the colony without bees is alive")
		}
	}
}

func TestSeasonalEventsSeasonStopOtherSpecies(t *testing.T) {
	// per species: a species with a later seasonStop keeps its bees
	cw, s, _ := seasWorld(t, 305)
	other := cw.speciesNamed(t, "late")
	cw.input.Species[other].SeasonStop = 330
	e := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0)
	_, b, _, _, _, _ := cw.beeMap.Get(e)
	b.Species = other
	s.Update(cw.w)
	if !cw.w.Alive(e) {
		t.Errorf("queen of a species with seasonStop 330 died on day 305")
	}
}

func TestSeasonalEventsMalesDie(t *testing.T) {
	// all males die if every queen hibernates, there is no brood and males exist, with the totals of the
	// previous evening (TotalQueens, TotalHibernatingQueens, TotalMales, counted from the bees) and morning (brood)
	cases := []struct {
		name                string
		activeQueen         bool // a queen that does not hibernate
		males               bool
		eggs, larvae, pupae int
		die                 bool
	}{
		{"all queens hibernate", false, true, 0, 0, 0, true},
		{"a queen is active", true, true, 0, 0, 0, false},
		{"brood left", false, true, 0, 0, 1, false},
		{"no males", false, false, 0, 0, 0, false},
	}
	for _, c := range cases {
		cw, s, totals := seasWorld(t, 250)
		totals.Eggs, totals.Larvae, totals.Pupae = c.eggs, c.larvae, c.pupae
		col := cw.addColony(nil)
		males := []ecs.Entity{}
		if c.males {
			males = append(males, cw.addMale(col, 1), cw.addMale(ecs.Entity{}, 2))
		}
		queen := cw.addHibernatingQueen(4)
		if c.activeQueen {
			cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0)
		}
		s.Update(cw.w)
		for _, m := range males {
			if cw.w.Alive(m) == c.die {
				t.Errorf("%s: male alive %v, want %v", c.name, cw.w.Alive(m), !c.die)
			}
		}
		if !cw.w.Alive(queen) {
			t.Errorf("%s: queen died", c.name)
		}
	}
}

func TestSeasonalEventsTotalsBeforeSeasonStop(t *testing.T) {
	// uses the totals of the previous evening: a queen that dies at the end of the season (Day seasonStop)
	// still counts as not hibernating, so the males survive this morning.
	cw, s, _ := seasWorld(t, 305)
	other := cw.speciesNamed(t, "late")
	cw.input.Species[other].SeasonStop = 330
	male := cw.addMale(ecs.Entity{}, 2)
	_, b, _, _, _, _ := cw.beeMap.Get(male)
	b.Species = other // survives
	cw.addHibernatingQueen(4)
	cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0) // dies at the end of the season
	s.Update(cw.w)
	if !cw.w.Alive(male) {
		t.Error("the males died on the day the last active queen died")
	}
	s.Update(cw.w) // the next morning: all queens hibernate
	if cw.w.Alive(male) {
		t.Error("the males survived the next morning")
	}
}

func TestSeasonalEventsHibernatingQueensLimit(t *testing.T) {
	// on Day 1: queensToKill = TotalHibernatingQueens (previous evening) - MaxHibernatingQueens
	cw, s, _ := seasWorld(t, 366)
	cw.cal.Day = 1
	cw.parameters.MaxHibernatingQueens = 5
	queens := []ecs.Entity{}
	for range 8 {
		queens = append(queens, cw.addHibernatingQueen(1))
	}
	active := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0) // not hibernating: never killed here
	s.Update(cw.w)

	alive := 0
	for _, e := range queens {
		if cw.w.Alive(e) {
			alive++
		}
	}
	if alive != 5 || !cw.w.Alive(active) {
		t.Errorf("%d hibernating queens alive, active queen alive %v; want 5 and true", alive, cw.w.Alive(active))
	}
}

func TestSeasonalEventsHibernatingCohort(t *testing.T) {
	// a cohort can remove more queens than needed. The planned number (TotalHibernatingQueensEverRemoved)
	// is output, see obs.Counters.
	cw, s, _ := seasWorld(t, 1)
	cw.parameters.MaxHibernatingQueens = 7
	cohort := cw.addHibernatingQueen(10)
	s.Update(cw.w)
	if cw.w.Alive(cohort) {
		t.Errorf("cohort of 10 survived with 3 queens to kill")
	}
	if cw.tally.beesDied != 10 {
		t.Errorf("%v bees died, want 10", cw.tally.beesDied)
	}
}

// countEntities returns the number of entities of a filter.
func countEntities[T any](f *ecs.Filter1[T]) int {
	q := f.Query()
	n := q.Count()
	q.Close()
	return n
}
