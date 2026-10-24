package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// checkWorld is a colonyWorld with 2 initial queens (BeeSpeciesInitialQueensList [["B_terrestris" 2]]):
// one hibernating without colony, one in a colony with brood, a worker and a male. Eggs, larvae, pupae, worker
// and male (14 bees) were produced in the run. The colony statistics are set by one UpdateColonies morning.
func checkWorld(t *testing.T) (*colonyWorld, ecs.Entity) {
	t.Helper()
	cw := newColonyWorld(t, 1)
	ecs.AddResource(cw.w, &globals.Derived{BeeSpeciesInitialQueensList: []globals.SpeciesQueens{{Species: "B_terrestris", Queens: 2}}})
	cw.file.InitialQueensModifierBehaviorSpace = 1
	cw.counters.TotalBeesEverProduced = 14

	hibernating := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 100)
	_, _, b, _, _, _, _ := cw.beeMap.Get(hibernating)
	b.Activity = activity.Hibernate
	col := cw.addColony(nil)
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	cw.addBee(col, stage.Egg, caste.Undefined, 5, 0)
	cw.addBee(col, stage.Larva, caste.Undefined, 4, 0)
	cw.addBee(col, stage.Pupa, caste.Worker, 3, 0)
	cw.addBee(col, stage.Adult, caste.Worker, 1, 2)
	cw.addBee(col, stage.Adult, caste.Male, 1, 2)
	cw.nextMorning()
	return cw, col
}

func TestCheckNumbersTotals(t *testing.T) {
	cw, _ := checkWorld(t)
	s := &CheckNumbers{}
	s.Initialize(cw.w)
	got := ecs.GetResource[globals.Totals](cw.w)
	got.TotalForagingTripsToday = 99
	s.Update(cw.w)
	// [CHK-02] finding 39 (see trans/netlogo-findings.md#finding-39)
	want := globals.Totals{TotalEggs: 5, TotalLarvae: 4, TotalPupae: 3, TotalAdultWorkers: 1,
		TotalAdultQueens: 2, TotalAdultMales: 1, TotalAdults: 4, TotalForagingTripsToday: 0}
	if *got != want {
		t.Errorf("totals:\n got %+v\nwant %+v", *got, want)
	}
}

func TestCheckNumbersAssertions(t *testing.T) {
	// [CHK-01] a bee that is neither initial nor produced
	cw, col := checkWorld(t)
	s := &CheckNumbers{}
	s.Initialize(cw.w)
	cw.addBee(col, stage.Adult, caste.Worker, 1, 2)
	expectPanic(t, "bee not produced", func() { s.Update(cw.w) })

	// [CHK-03] colony statistics differ from the bees (an egg laid after the morning statistics)
	cw, col = checkWorld(t)
	s = &CheckNumbers{}
	s.Initialize(cw.w)
	cw.addBee(col, stage.Egg, caste.Undefined, 2, 0)
	cw.counters.TotalBeesEverProduced += 2
	expectPanic(t, "colony statistics", func() { s.Update(cw.w) })
}

func TestCheckNumbersColonyCheck(t *testing.T) {
	// [CHK-04] day 364: every colony ever produced must have died
	for _, skip := range []bool{false, true} {
		cw, _ := checkWorld(t)
		cw.counters.TotalColoniesEverProduced = 1
		s := &CheckNumbers{SkipColonyCheck: skip}
		s.Initialize(cw.w) // Day 201
		cw.cal.Day = 364
		if skip {
			s.Update(cw.w) // must not panic
		} else {
			expectPanic(t, "living colony on day 364", func() { s.Update(cw.w) })
		}
	}
	cw, _ := checkWorld(t)
	cw.counters.TotalColoniesEverProduced = 3
	cw.counters.ColonyDeathsEndSeason = 2
	cw.counters.ColonyDeathsNoBees = 1
	s := &CheckNumbers{}
	s.Initialize(cw.w)
	cw.cal.Day = 364
	s.Update(cw.w) // all colonies accounted for
}

func TestMorningExampleFarm(t *testing.T) {
	// The morning systems on the example farm for one year, with the queens of step 3. Without brood and
	// foraging every colony consists of its queen only and keeps its start energy, so no colony dies.
	a := newExampleFarmSetup(t, 1, &CreateInitialQueens{}, &Tick{}, &UpdateMorning{}, &UpdateFoodsources{},
		&EmergenceNewQueens{}, &UpdateColonies{}, &CheckNumbers{SkipColonyCheck: true})
	w := a.World
	counters := ecs.GetResource[globals.Counters](w)
	totals := ecs.GetResource[globals.Totals](w)
	cal := ecs.GetResource[globals.Calendar](w)
	today := ecs.GetResource[globals.FoodsourcesToday](w)
	if cal.Day != 0 || today.NectarAvailableTotalToday_l != 0 {
		t.Fatalf("Setup: Day %v, nectar %v; want 0, 0", cal.Day, today.NectarAvailableTotalToday_l)
	}
	filter := ecs.NewFilter3[comp.Colony, comp.ColonyCounts, comp.ColonyRange](w)
	maxNectar, withLists := 0.0, 0
	for range 400 {
		a.Update() // panics on a failed assertion
		maxNectar = max(maxNectar, today.NectarAvailableTotalToday_l)
		colonies := 0
		q := filter.Query()
		for q.Next() {
			c, n, r := q.Get()
			colonies++
			if n.Colonysize != 1 || n.AllAdults != 1 || c.EusocialPhaseDate != params.NotSetHigh {
				t.Fatalf("colony: colonysize %v, allAdults %v, eusocialPhaseDate %v", n.Colonysize, n.AllAdults, c.EusocialPhaseDate)
			}
			if len(r.AllSourcesInFlowerAndRangeList) > 0 {
				withLists++
			}
		}
		if float64(colonies) != counters.TotalColoniesEverProduced {
			t.Fatalf("day %v: %d living colonies, %v founded", cal.Day, colonies, counters.TotalColoniesEverProduced)
		}
		if totals.TotalAdults != totals.TotalAdultQueens {
			t.Fatalf("day %v: TotalAdults %v, TotalAdultQueens %v; only queens exist", cal.Day, totals.TotalAdults, totals.TotalAdultQueens)
		}
	}
	if maxNectar == 0 || withLists == 0 {
		t.Errorf("max nectar %v l, colony days with foodsources in flower and range %d; want > 0", maxNectar, withLists)
	}
}
