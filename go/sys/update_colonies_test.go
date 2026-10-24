package sys

import (
	"math"
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// colonyWorld is a small world for the tests of step 4a: B_terrestris, colonies and bees created by hand,
// and the resources of the morning systems. Ticks and Day are 200 (before seasonStop 305).
type colonyWorld struct {
	w         *ecs.World
	ticks     *globals.Ticks
	cal       *globals.Calendar
	today     *globals.FoodsourcesToday
	file      *params.ParameterFile
	constants *params.Constants
	counters  *globals.Counters
	dead      *globals.DeadColonies
	sys       *UpdateColonies
	species   ecs.Entity
	colonyMap *colonyMapper
	beeMap    *beeMapper
	who       *globals.WhoCounter
}

func newColonyWorld(t *testing.T, seed uint64) *colonyWorld {
	t.Helper()
	w := ecs.NewWorld()
	cw := &colonyWorld{w: w}
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(seed, 3)})
	p := params.Default()
	cw.file, cw.constants = &p.File, &p.Constants
	ecs.AddResource(w, cw.file)
	ecs.AddResource(w, cw.constants)
	cw.ticks = &globals.Ticks{Ticks: 200}
	ecs.AddResource(w, cw.ticks)
	cw.cal = &globals.Calendar{Day: 200}
	ecs.AddResource(w, cw.cal)
	cw.today = &globals.FoodsourcesToday{}
	ecs.AddResource(w, cw.today)
	ecs.AddResource(w, &globals.Landscape{Scaling_NLpatches_per_m: 0.1})
	cw.who = &globals.WhoCounter{}
	ecs.AddResource(w, cw.who)
	cw.counters = &globals.Counters{}
	ecs.AddResource(w, cw.counters)
	ecs.AddResource(w, &globals.ActiveBees{})

	sp := *terrestris(t) // seasonStop 305
	cw.species = ecs.NewMap2[comp.Who, comp.Species](w).NewEntity(&comp.Who{Who: cw.who.Next()}, &sp)
	cw.colonyMap = ecs.NewMap7[comp.Who, comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w)
	cw.beeMap = newBeeMapper(w)

	cw.sys = &UpdateColonies{}
	cw.sys.Initialize(w)
	cw.dead = ecs.GetResource[globals.DeadColonies](w)
	return cw
}

// addColony adds a colony with the start values of CreateColoniesProc [CCOL-04]. change may modify them.
func (cw *colonyWorld) addColony(change func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts)) ecs.Entity {
	c := comp.Colony{SpeciesIDcolony: cw.species, SpeciesNameColony: "B_terrestris", CohortBased: true, Queenright: true,
		QueenProductionDate: params.NotSetHigh, SwitchPointDate: params.NotSetHigh,
		CompetitionPointDate: params.NotSetHigh, EusocialPhaseDate: params.NotSetHigh}
	s := comp.ColonyStores{EnergyStore_kJ: 0.873}
	n := comp.ColonyCounts{Colonysize: 1}
	if change != nil {
		change(&c, &s, &n)
	}
	return cw.colonyMap.NewEntity(&comp.Who{Who: cw.who.Next()}, &comp.Position{}, &c, &s, &comp.ColonyStimuli{}, &n, &comp.ColonyRange{})
}

// addBee adds a bee or cohort to a colony. Adult queens are mated.
func (cw *colonyWorld) addBee(col ecs.Entity, st stage.Stage, c caste.Caste, number, adultAge float64) ecs.Entity {
	b := comp.Bee{Number: number, Stage: st, Brood: st != stage.Adult, Caste: c, Activity: activity.Resting,
		ColonyID: col, SpeciesID: cw.species, SpeciesName: "B_terrestris", AdultAge: adultAge, Weight_mg: 100,
		Mated: st == stage.Adult && c == caste.Queen}
	return cw.beeMap.NewEntity(&comp.Who{Who: cw.who.Next()}, &comp.Position{}, &b, &comp.Genetics{},
		&comp.Thresholds{}, &comp.Forager{}, &comp.Development{})
}

func (cw *colonyWorld) colony(e ecs.Entity) (*comp.Colony, *comp.ColonyStores, *comp.ColonyCounts) {
	_, _, c, s, _, n, _ := cw.colonyMap.Get(e)
	return c, s, n
}

// nextMorning advances ticks and Day by one and runs UpdateColonies.
func (cw *colonyWorld) nextMorning() {
	cw.ticks.Ticks++
	cw.cal.Day = dayOfYear(cw.ticks.Ticks)
	cw.sys.Update(cw.w)
}

func TestUpdateColoniesStatistics(t *testing.T) {
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.SummedIncubationToday_kJ = 5
	})
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	young := cw.addBee(col, stage.Adult, caste.Queen, 1, 2)
	_, _, b, _, _, _, _ := cw.beeMap.Get(young)
	b.Activity = activity.Hibernate
	b.Weight_mg = 800
	cw.addBee(col, stage.Adult, caste.Worker, 3, 5)
	cw.addBee(col, stage.Adult, caste.Male, 2, 1)
	cw.addBee(col, stage.Egg, caste.Undefined, 10, 0)
	cw.addBee(col, stage.Larva, caste.Worker, 6, 0)
	cw.addBee(col, stage.Pupa, caste.Male, 4, 0)
	empty := cw.addColony(nil) // only the queen, no workers
	cw.addBee(empty, stage.Adult, caste.Queen, 1, 300)

	cw.nextMorning()
	c, s, n := cw.colony(col)
	want := comp.ColonyCounts{Colonysize: 27, ColonyWeight_mg: 26*100 + 800, LarvaWorkerRatio: 2,
		AllEggs: 10, AllLarvae: 6, AllPupae: 4, AllAdults: 7, AllAdultWorkers: 3,
		AllAdultQueens: 2, AllAdultActiveQueens: 1, AllAdultMales: 2}
	if *n != want {
		t.Errorf("statistics [UCOL-06]:\n got %+v\nwant %+v", *n, want)
	}
	if s.SummedIncubationToday_kJ != 0 {
		t.Errorf("[UCOL-05] summedIncubationToday_kJ = %v, want 0", s.SummedIncubationToday_kJ)
	}
	// [UCOL-07] the first adult worker starts the eusocial phase
	if c.EusocialPhaseDate != 201 || c.ColonyAge != 1 {
		t.Errorf("eusocialPhaseDate %v, colonyAge %v; want 201, 1", c.EusocialPhaseDate, c.ColonyAge)
	}
	c, _, n = cw.colony(empty)
	if n.LarvaWorkerRatio != params.NotSetHigh || c.EusocialPhaseDate != params.NotSetHigh {
		t.Errorf("colony without workers: larvaWorkerRatio %v, eusocialPhaseDate %v; want NotSetHigh", n.LarvaWorkerRatio, c.EusocialPhaseDate)
	}
}

func TestUpdateColoniesDeathAfterCP(t *testing.T) {
	// [UCOL-02] finding 37 (see trans/netlogo-findings.md#finding-37) and finding 38 (see trans/netlogo-findings.md#finding-38)
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.CompetitionPointDate = 150
		n.AllPupae = 4 // counted on the previous morning, the pupae have emerged since
	})
	queen := cw.addBee(col, stage.Adult, caste.Queen, 1, 400)
	oldWorker := cw.addBee(col, stage.Adult, caste.Worker, 1, 20)
	youngWorker := cw.addBee(col, stage.Adult, caste.Worker, 4, 5)
	male := cw.addBee(col, stage.Adult, caste.Male, 1, 3)

	cw.nextMorning() // 201: the stale pupae protect the adults for one more day
	if !cw.w.Alive(queen) || !cw.w.Alive(oldWorker) {
		t.Fatal("day 201: adults must survive, the brood counts of the previous morning are not 0")
	}
	if _, _, n := cw.colony(col); n.AllPupae != 0 {
		t.Fatalf("day 201: allPupae %v, want 0 after the statistics", n.AllPupae)
	}
	cw.nextMorning() // 202: workers and queens older than 10 days die
	if cw.w.Alive(queen) || cw.w.Alive(oldWorker) || !cw.w.Alive(youngWorker) || !cw.w.Alive(male) {
		t.Fatal("day 202: queen and old worker must die, young worker and male survive")
	}
	if cw.counters.TotalBeesEverDied != 2 {
		t.Errorf("TotalBeesEverDied %v, want 2", cw.counters.TotalBeesEverDied)
	}
	_, _, b, _, _, _, _ := cw.beeMap.Get(youngWorker)
	b.AdultAge = 11
	cw.nextMorning() // 203: the young worker is now older than 10 days, only the male is left
	if cw.w.Alive(youngWorker) || !cw.w.Alive(col) {
		t.Fatal("day 203: the worker must die, the colony must stay alive because of the male")
	}
	if c, _, _ := cw.colony(col); c.Queenright || c.ColonyAge != 3 {
		t.Errorf("day 203: queenright %v, colonyAge %v; want false (no mated queen, [UCOL-10]), 3", c.Queenright, c.ColonyAge)
	}

	// the male dies (e.g. after MaxLifespanMales). The colony dies the next morning [UCOL-11].
	death := newBeeDeath(cw.w)
	death.DieProc(cw.w, male, "Male: adultAge > MaxLifespanMales")
	cw.nextMorning() // 204
	if cw.w.Alive(col) {
		t.Fatal("day 204: the colony without bees must die")
	}
	if cw.counters.ColonyDeathsNoBees != 1 || cw.counters.ColonyDeathsEndSeason != 0 {
		t.Errorf("ColonyDeathsNoBees %v, ColonyDeathsEndSeason %v; want 1, 0", cw.counters.ColonyDeathsNoBees, cw.counters.ColonyDeathsEndSeason)
	}
	if len(cw.dead.List) != 1 || cw.dead.List[0].ColonyDeathDay != 204 || cw.dead.List[0].Colony.SpeciesNameColony != "B_terrestris" {
		t.Errorf("dead colonies: %+v, want one record with colonyDeathDay 204 and the species name", cw.dead.List)
	}
}

func TestUpdateColoniesEnergyDepleted(t *testing.T) {
	// [UCOL-03] all bees die, the brood is counted, the colony dies on the same morning [UCOL-11]
	cw := newColonyWorld(t, 1)
	cw.ticks.Ticks = 309 // Day 310 >= seasonStop 305: counted as end of season
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.EnergyStore_kJ = -0.1
		n.TotalEggsProduced = 7 // for the egg accounting at death [CHK-04]
	})
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	cw.addBee(col, stage.Larva, caste.Undefined, 5, 0)
	cw.addBee(col, stage.Egg, caste.Undefined, 2, 0)

	cw.nextMorning()
	if cw.w.Alive(col) || cw.counters.TotalBeesEverDied != 8 {
		t.Fatalf("colony alive %v, TotalBeesEverDied %v; want false, 8", cw.w.Alive(col), cw.counters.TotalBeesEverDied)
	}
	if cw.counters.ColonyDeathsEndSeason != 1 || cw.counters.ColonyDeathsNoBees != 0 {
		t.Errorf("ColonyDeathsEndSeason %v, ColonyDeathsNoBees %v; want 1, 0", cw.counters.ColonyDeathsEndSeason, cw.counters.ColonyDeathsNoBees)
	}
	if got := cw.dead.List[0].Counts.BroodDeathsEnergyStores; got != 7 {
		t.Errorf("broodDeathsEnergyStores %v, want 7", got)
	}
}

func TestUpdateColoniesNoAdults(t *testing.T) {
	// [UCOL-04] brood without any adult dies. An adult male keeps the brood alive (finding 38, see
	// trans/netlogo-findings.md#finding-38).
	cw := newColonyWorld(t, 1)
	orphan := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { n.TotalEggsProduced = 3 })
	cw.addBee(orphan, stage.Egg, caste.Undefined, 3, 0)
	withMale := cw.addColony(nil)
	brood := cw.addBee(withMale, stage.Pupa, caste.Worker, 6, 0)
	cw.addBee(withMale, stage.Adult, caste.Male, 1, 3)

	cw.nextMorning()
	if cw.w.Alive(orphan) {
		t.Fatal("the colony with brood only must die")
	}
	if got := cw.dead.List[0].Counts.BroodDeathsNoAdults; got != 3 {
		t.Errorf("broodDeathsNoAdults %v, want 3", got)
	}
	if !cw.w.Alive(brood) || !cw.w.Alive(withMale) {
		t.Error("brood with an adult male must survive")
	}
}

func TestUpdateColoniesEggCheck(t *testing.T) {
	// [CHK-04] at death (decision E): eggs that are not accounted for are an assertion
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { n.TotalEggsProduced = 5 })
	cw.addBee(col, stage.Egg, caste.Undefined, 3, 0)
	expectPanic(t, "5 eggs produced, 3 died", func() { cw.nextMorning() })
}

func TestUpdateColoniesSwitchPoint(t *testing.T) {
	// [UCOL-08] in the eusocial phase and with larvaWorkerRatio < 3: p = 0.13 per day
	cw := newColonyWorld(t, 5)
	const n = 2000
	cols := make([]ecs.Entity, n)
	for i := range cols {
		cols[i] = cw.addColony(nil)
		cw.addBee(cols[i], stage.Adult, caste.Worker, 1, 1) // eusocial phase starts this morning [UCOL-07]
	}
	highRatio := cw.addColony(nil)
	cw.addBee(highRatio, stage.Adult, caste.Worker, 2, 1)
	cw.addBee(highRatio, stage.Larva, caste.Undefined, 6, 0) // ratio 3, not < 3

	cw.nextMorning()
	switched := 0
	for _, e := range cols {
		if c, _, _ := cw.colony(e); c.SwitchPointDate == 201 {
			switched++
		} else if c.SwitchPointDate != params.NotSetHigh {
			t.Fatalf("switchPointDate %v", c.SwitchPointDate)
		}
	}
	if p := float64(switched) / n; math.Abs(p-0.13) > 0.03 {
		t.Errorf("switched %d of %d colonies (%.3f), want about 0.13", switched, n, p)
	}
	if c, _, _ := cw.colony(highRatio); c.SwitchPointDate != params.NotSetHigh {
		t.Error("larvaWorkerRatio 3 must not switch")
	}
	// a switch point is set only once
	date := map[ecs.Entity]float64{}
	for _, e := range cols {
		c, _, _ := cw.colony(e)
		date[e] = c.SwitchPointDate
	}
	cw.nextMorning()
	for _, e := range cols {
		if c, _, _ := cw.colony(e); date[e] != params.NotSetHigh && c.SwitchPointDate != date[e] {
			t.Fatal("switch point changed after it was set")
		}
	}
}

func TestCompetitionPointDateREP(t *testing.T) {
	// [CPD-01] worked examples of the annotation, and the bounds
	cases := []struct{ eus, qpd, want float64 }{
		{120, 130, 143}, // y = 22.5, round 142.5 = 143
		{120, 180, 165}, // 178 - 120 = 58 > 45
		{120, 115, 132}, // x = -5: y = 12
		{120, 120, 136}, // x = 0: y = 15.5, round 135.5 = 136
		{120, 162, 165}, // x = 42: y = 44.9, round 164.9 = 165
		{120, 163, 165}, // x = 43: y = 45.6, round 165.6 = 166 > 165
	}
	for _, c := range cases {
		col := comp.Colony{EusocialPhaseDate: c.eus, QueenProductionDate: c.qpd}
		if got := CompetitionPointDateREP(&col); got != c.want {
			t.Errorf("CompetitionPointDateREP(eus %v, qpd %v) = %v, want %v", c.eus, c.qpd, got, c.want)
		}
	}

	// [UCOL-09] set once both dates are set
	cw := newColonyWorld(t, 1)
	set := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.EusocialPhaseDate, c.QueenProductionDate = 120, 130
	})
	unset := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.EusocialPhaseDate = 120 })
	for _, e := range []ecs.Entity{set, unset} {
		cw.addBee(e, stage.Adult, caste.Queen, 1, 300)
	}
	cw.nextMorning()
	if c, _, _ := cw.colony(set); c.CompetitionPointDate != 143 {
		t.Errorf("competitionPointDate %v, want 143", c.CompetitionPointDate)
	}
	if c, _, _ := cw.colony(unset); c.CompetitionPointDate != params.NotSetHigh {
		t.Errorf("without queenProductionDate: competitionPointDate %v, want NotSetHigh", c.CompetitionPointDate)
	}
}

func TestUpdateColoniesRangeLists(t *testing.T) {
	// [UCOL-01] the lists in flower and range are rebuilt only if FoodsourcesInFlowerUpdate
	cw := newColonyWorld(t, 1)
	fsMap := ecs.NewMap3[comp.Foodsource, comp.FoodsourceState, comp.Layer](cw.w)
	inFlower := fsMap.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{Pollen_g: 1, Nectar_myl: 1}, &comp.Layer{Masterpatch: true})
	notInFlower := fsMap.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{}, &comp.Layer{Masterpatch: true})
	col := cw.addColony(nil)
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	_, _, _, _, _, _, r := cw.colonyMap.Get(col)
	r.AllPatchesInRangeList = []ecs.Entity{inFlower, notInFlower}

	cw.nextMorning() // no update
	if _, _, _, _, _, _, r := cw.colonyMap.Get(col); len(r.PollenInFlowerAndRangeList) != 0 {
		t.Fatal("lists must not change without FoodsourcesInFlowerUpdate")
	}
	cw.today.FoodsourcesInFlowerUpdate = true
	cw.nextMorning()
	_, _, _, _, _, _, r = cw.colonyMap.Get(col)
	if len(r.PollenInFlowerAndRangeList) != 1 || r.PollenInFlowerAndRangeList[0] != inFlower ||
		len(r.NectarInFlowerAndRangeList) != 1 || len(r.AllSourcesInFlowerAndRangeList) != 1 {
		t.Errorf("lists after the update: %+v", *r)
	}
}

func TestUpdateColoniesDeadColoniesNewYear(t *testing.T) {
	// [UCOL-12] on 1 January the dead colonies are removed unless KeepDeadColonies?
	for _, keep := range []bool{true, false} {
		cw := newColonyWorld(t, 1)
		cw.file.KeepDeadColonies = keep
		cw.addColony(nil) // no bees: dies on the next morning
		cw.ticks.Ticks = 364
		cw.nextMorning() // Day 365
		if len(cw.dead.List) != 1 {
			t.Fatalf("keep %v: %d dead colonies on day 365, want 1", keep, len(cw.dead.List))
		}
		cw.nextMorning() // Day 1
		want := 1
		if !keep {
			want = 0
		}
		if len(cw.dead.List) != want {
			t.Errorf("KeepDeadColonies %v: %d dead colonies on 1 January, want %d", keep, len(cw.dead.List), want)
		}
	}
}
