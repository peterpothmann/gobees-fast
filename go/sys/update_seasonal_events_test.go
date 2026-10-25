package sys

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// seasWorld is a colonyWorld with the totals of the previous evening (OutputDaily) and morning (CheckNumbers),
// set by hand, and the UpdateSeasonalEvents system.
func seasWorld(t *testing.T, day float64) (*colonyWorld, *UpdateSeasonalEvents, *globals.DailyOutput, *globals.Totals) {
	t.Helper()
	cw := newColonyWorld(t, 1)
	cw.ticks.Ticks = int64(day)
	cw.cal.Day = day
	out := &globals.DailyOutput{}
	ecs.AddResource(cw.w, out)
	totals := &globals.Totals{}
	ecs.AddResource(cw.w, totals)
	s := &UpdateSeasonalEvents{}
	s.Initialize(cw.w)
	return cw, s, out, totals
}

// addHibernatingQueen adds a mated, hibernating queen (cohort) without colony.
func (cw *colonyWorld) addHibernatingQueen(number float64) ecs.Entity {
	e := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, number, 0)
	_, _, b, _, _, _, _ := cw.beeMap.Get(e)
	b.Activity = activity.Hibernate
	return e
}

func TestSeasonalEventsSeasonStop(t *testing.T) {
	// [SEAS-01] on Day seasonStop (305 for B_terrestris) every bee of the species that does not hibernate dies
	for _, day := range []float64{304, 305} {
		cw, s, _, _ := seasWorld(t, day)
		col := cw.addColony(func(c *comp.Colony, st *comp.ColonyStores, n *comp.ColonyCounts) {
			n.TotalEggsProduced = 9 // the 5 eggs and 4 larvae below
		})
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
		_, _, n := cw.colony(col)
		if day == 304 {
			if alive != 8 || n.BroodDeathEndSeason != 0 {
				t.Errorf("day 304: %d bee agents, broodDeathEndSeason %v; want 8 and 0", alive, n.BroodDeathEndSeason)
			}
			continue
		}
		if alive != 1 || !cw.w.Alive(hibernating) || cw.w.Alive(searching) {
			t.Errorf("day 305: %d bee agents alive, want only the hibernating queen", alive)
		}
		if n.BroodDeathEndSeason != 9 {
			t.Errorf("broodDeathEndSeason %v, want 9", n.BroodDeathEndSeason)
		}
		// the colony dies later this morning in [UCOL-11], counted as end of season. Its eggs are accounted for.
		cw.sys.Update(cw.w)
		if cw.w.Alive(col) || cw.counters.ColonyDeathsEndSeason != 1 || cw.counters.ColonyDeathsNoBees != 0 {
			t.Errorf("colony alive %v, ColonyDeathsEndSeason %v, ColonyDeathsNoBees %v; want false, 1, 0",
				cw.w.Alive(col), cw.counters.ColonyDeathsEndSeason, cw.counters.ColonyDeathsNoBees)
		}
	}
}

func TestSeasonalEventsSeasonStopOtherSpecies(t *testing.T) {
	// [SEAS-01] per species: a species with a later seasonStop keeps its bees
	cw, s, _, _ := seasWorld(t, 305)
	other := cw.speciesEntityNamed(t, "late")
	ecs.NewMap1[comp.Species](cw.w).Get(other).SeasonStop = 330
	e := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0)
	_, _, b, _, _, _, _ := cw.beeMap.Get(e)
	b.SpeciesID = other
	s.Update(cw.w)
	if !cw.w.Alive(e) {
		t.Errorf("queen of a species with seasonStop 330 died on day 305")
	}
}

func TestSeasonalEventsMalesDie(t *testing.T) {
	// [SEAS-02] all males die if every queen hibernates, there is no brood and males exist, with the totals of the
	// previous evening (TotalQueens, TotalHibernatingQueens, TotalMales) and morning (brood)
	cases := []struct {
		name                string
		queens, hib, males  float64
		eggs, larvae, pupae float64
		die                 bool
	}{
		{"all queens hibernate", 4, 4, 3, 0, 0, 0, true},
		{"a queen is active", 4, 3, 3, 0, 0, 0, false},
		{"brood left", 4, 4, 3, 0, 0, 1, false},
		{"no males", 4, 4, 0, 0, 0, 0, false},
	}
	for _, c := range cases {
		cw, s, out, totals := seasWorld(t, 250)
		out.TotalQueens, out.TotalHibernatingQueens, out.TotalMales = c.queens, c.hib, c.males
		totals.TotalEggs, totals.TotalLarvae, totals.TotalPupae = c.eggs, c.larvae, c.pupae
		col := cw.addColony(nil)
		males := []ecs.Entity{cw.addMale(col, 1), cw.addMale(ecs.Entity{}, 2)}
		queen := cw.addHibernatingQueen(4)
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

func TestSeasonalEventsHibernatingQueensLimit(t *testing.T) {
	// [SEAS-03] on Day 1: queensToKill = TotalHibernatingQueens (previous evening) - MaxHibernatingQueens
	cw, s, out, _ := seasWorld(t, 366)
	cw.cal.Day = 1
	cw.file.MaxHibernatingQueens = 5
	queens := []ecs.Entity{}
	for range 8 {
		queens = append(queens, cw.addHibernatingQueen(1))
	}
	active := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 0) // not hibernating: never killed here
	out.TotalHibernatingQueens = 8
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
	if cw.counters.TotalHibernatingQueensEverRemoved != 3 {
		t.Errorf("TotalHibernatingQueensEverRemoved %v, want 3", cw.counters.TotalHibernatingQueensEverRemoved)
	}
	// output only: the number of bees before the queens were removed
	if !slices.Equal(cw.counters.PopulationSizeDay365List, []float64{9}) {
		t.Errorf("PopulationSizeDay365List %v, want [9]", cw.counters.PopulationSizeDay365List)
	}
}

func TestSeasonalEventsHibernatingCohort(t *testing.T) {
	// [SEAS-03] a cohort can remove more queens than needed. TotalHibernatingQueensEverRemoved adds the planned
	// number.
	cw, s, out, _ := seasWorld(t, 1)
	cw.file.MaxHibernatingQueens = 7
	cohort := cw.addHibernatingQueen(10)
	out.TotalHibernatingQueens = 10
	s.Update(cw.w)
	if cw.w.Alive(cohort) {
		t.Errorf("cohort of 10 survived with 3 queens to kill")
	}
	if cw.counters.TotalHibernatingQueensEverRemoved != 3 || cw.counters.TotalBeesEverDied != 10 {
		t.Errorf("removed %v, died %v; want 3 and 10", cw.counters.TotalHibernatingQueensEverRemoved, cw.counters.TotalBeesEverDied)
	}
}

func TestSeasonalEventsCumulVisits(t *testing.T) {
	// [SEAS-03] the visit counters are reset on Day 1 with CumulVisitsOnlyLastYear?
	for _, reset := range []bool{false, true} {
		for _, day := range []float64{1, 2} {
			cw, s, _, _ := seasWorld(t, day)
			cw.file.CumulVisitsOnlyLastYear = reset
			e := ecs.NewMap1[comp.FoodsourceState](cw.w).NewEntity(&comp.FoodsourceState{CumulNectarVisits: 4, CumulPollenVisits: 2})
			s.Update(cw.w)
			st := ecs.NewMap1[comp.FoodsourceState](cw.w).Get(e)
			wantZero := reset && day == 1
			if (st.CumulNectarVisits == 0) != wantZero || (st.CumulPollenVisits == 0) != wantZero {
				t.Errorf("reset %v, day %v: visits %v/%v", reset, day, st.CumulNectarVisits, st.CumulPollenVisits)
			}
		}
	}
}

// countEntities returns the number of entities of a filter.
func countEntities[T any](f *ecs.Filter1[T]) int {
	q := f.Query()
	n := q.Count()
	q.Close()
	return n
}
