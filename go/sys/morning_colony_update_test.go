package sys

import (
	"math"
	"math/rand/v2"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// colonyWorld is a small world for the tests of colonyUpdate: B_terrestris, colonies and bees created by hand,
// and the resources of the morning systems. Ticks and Day are 200 (before seasonStop 305).
type colonyWorld struct {
	w          *ecs.World
	ticks      *globals.Ticks
	cal        *globals.Calendar
	today      *globals.Flowering
	parameters *params.Parameters
	constants  *params.Constants
	tally      *tally
	sys        *colonyUpdate
	input      *params.Input // the species
	species    int           // index of B_terrestris
	colonyMap  *colonyMapper
	beeMap     *beeMapper
}

func newColonyWorld(t *testing.T, seed uint64) *colonyWorld {
	t.Helper()
	w := ecs.NewWorld()
	cw := &colonyWorld{w: w}
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(seed, 3)})
	p := params.Default()
	cw.parameters, cw.constants = &p.Parameters, &p.Constants
	ecs.AddResource(w, cw.parameters)
	ecs.AddResource(w, cw.constants)
	cw.ticks = &globals.Ticks{Ticks: 200}
	ecs.AddResource(w, cw.ticks)
	cw.cal = &globals.Calendar{Day: 200}
	ecs.AddResource(w, cw.cal)
	cw.today = &globals.Flowering{}
	ecs.AddResource(w, cw.today)
	ecs.AddResource(w, &globals.Landscape{Scale: 0.1})
	cw.tally = newTally(w)
	ecs.AddResource(w, &globals.ActiveBees{})
	ecs.AddResource(w, &globals.Daytime{})

	cw.input = &params.Input{Species: []params.Species{*terrestris(t)}} // seasonStop 305
	ecs.AddResource(w, cw.input)
	cw.species = 0
	cw.colonyMap = ecs.NewMap6[comp.Position, comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts, comp.ColonyRange](w)
	cw.beeMap = newBeeMapper(w)

	cw.sys = &colonyUpdate{}
	cw.sys.Initialize(w)
	return cw
}

// addColony adds a colony with the start values of colonyFounder.found. change may modify them.
func (cw *colonyWorld) addColony(change func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts)) ecs.Entity {
	c := comp.Colony{Species: cw.species} // the four dates unset
	s := comp.ColonyStores{Energy: 0.873}
	n := comp.ColonyCounts{}
	if change != nil {
		change(&c, &s, &n)
	}
	return cw.colonyMap.NewEntity(&comp.Position{}, &c, &s, &comp.ColonyStimuli{}, &n, &comp.ColonyRange{})
}

// addBee adds a bee or cohort to a colony. Adult queens are mated.
func (cw *colonyWorld) addBee(col ecs.Entity, st stage.Stage, c caste.Caste, number, adultAge int) ecs.Entity {
	b := comp.Bee{Number: number, Stage: st, Caste: c, Activity: activity.Resting,
		Colony: col, Species: cw.species, AdultAge: adultAge, Weight: 100,
		Mated: st == stage.Adult && c == caste.Queen}
	return cw.beeMap.NewEntity(&comp.Position{}, &b, &comp.Genetics{},
		&comp.Thresholds{}, &comp.Forager{}, &comp.Development{})
}

func (cw *colonyWorld) colony(e ecs.Entity) (*comp.Colony, *comp.ColonyStores, *comp.ColonyCounts) {
	_, c, s, _, n, _ := cw.colonyMap.Get(e)
	return c, s, n
}

// nextMorning advances ticks and Day by one and runs colonyUpdate.
func (cw *colonyWorld) nextMorning() {
	cw.ticks.Ticks++
	cw.cal.Day = dayOfYear(cw.ticks.Ticks)
	cw.sys.Update(cw.w)
}

func TestColonyUpdateStatistics(t *testing.T) {
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.IncubationToday = 5
	})
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	young := cw.addBee(col, stage.Adult, caste.Queen, 1, 2)
	_, b, _, _, _, _ := cw.beeMap.Get(young)
	b.Activity = activity.Hibernate
	b.Weight = 800
	cw.addBee(col, stage.Adult, caste.Worker, 3, 5)
	cw.addBee(col, stage.Adult, caste.Male, 2, 1)
	cw.addBee(col, stage.Egg, caste.Undefined, 10, 0)
	cw.addBee(col, stage.Larva, caste.Worker, 6, 0)
	cw.addBee(col, stage.Pupa, caste.Male, 4, 0)
	empty := cw.addColony(nil) // only the queen, no workers
	cw.addBee(empty, stage.Adult, caste.Queen, 1, 300)

	cw.nextMorning()
	c, s, n := cw.colony(col)
	want := comp.ColonyCounts{LarvaWorkerRatio: 2, Eggs: 10, Larvae: 6, Pupae: 4, Adults: 7,
		AdultWorkers: 3}
	if *n != want {
		t.Errorf("statistics:\n got %+v\nwant %+v", *n, want)
	}
	if s.IncubationToday != 0 {
		t.Errorf("IncubationToday = %v, want 0", s.IncubationToday)
	}
	// the first adult worker starts the eusocial phase
	if c.EusocialPhaseDate != day.At(201) {
		t.Errorf("eusocialPhaseDate %v, want 201", c.EusocialPhaseDate)
	}
	c, _, n = cw.colony(empty)
	if n.LarvaWorkerRatio != params.NotSetHigh || c.EusocialPhaseDate.IsSet() {
		t.Errorf("colony without workers: larvaWorkerRatio %v, eusocialPhaseDate %v; want NotSetHigh, unset", n.LarvaWorkerRatio, c.EusocialPhaseDate)
	}
}

func TestColonyUpdateDeathAfterCP(t *testing.T) {
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.CompetitionPointDate = day.At(150)
		n.Pupae = 4 // counted on the previous morning, the pupae have emerged since
	})
	queen := cw.addBee(col, stage.Adult, caste.Queen, 1, 400)
	oldWorker := cw.addBee(col, stage.Adult, caste.Worker, 1, 20)
	youngWorker := cw.addBee(col, stage.Adult, caste.Worker, 4, 5)
	male := cw.addBee(col, stage.Adult, caste.Male, 1, 3)

	cw.nextMorning() // 201: the stale pupae protect the adults for one more day
	if !cw.w.Alive(queen) || !cw.w.Alive(oldWorker) {
		t.Fatal("day 201: adults must survive, the brood counts of the previous morning are not 0")
	}
	if _, _, n := cw.colony(col); n.Pupae != 0 {
		t.Fatalf("day 201: allPupae %v, want 0 after the statistics", n.Pupae)
	}
	cw.nextMorning() // 202: workers and queens older than 10 days die
	if cw.w.Alive(queen) || cw.w.Alive(oldWorker) || !cw.w.Alive(youngWorker) || !cw.w.Alive(male) {
		t.Fatal("day 202: queen and old worker must die, young worker and male survive")
	}
	if cw.tally.beesDied != 2 {
		t.Errorf("%v bees died, want 2", cw.tally.beesDied)
	}
	_, b, _, _, _, _ := cw.beeMap.Get(youngWorker)
	b.AdultAge = 11
	cw.nextMorning() // 203: the young worker is now older than 10 days, only the male is left
	if cw.w.Alive(youngWorker) || !cw.w.Alive(col) {
		t.Fatal("day 203: the worker must die, the colony must stay alive because of the male")
	}

	// the male dies (e.g. after MaxLifespanMales). The colony dies the next morning.
	death := newBeeDeath(cw.w)
	death.die(cw.w, male, "Male: adultAge > MaxLifespanMales")
	cw.nextMorning() // 204
	if cw.w.Alive(col) {
		t.Fatal("day 204: the colony without bees must die")
	}
	if cw.tally.coloniesDied != 1 {
		t.Errorf("%d colonies died, want 1", cw.tally.coloniesDied)
	}
}

func TestColonyUpdateEnergyDepleted(t *testing.T) {
	// all bees die, the colony dies on the same morning
	cw := newColonyWorld(t, 1)
	col := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { s.Energy = -0.1 })
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	cw.addBee(col, stage.Larva, caste.Undefined, 5, 0)
	cw.addBee(col, stage.Egg, caste.Undefined, 2, 0)

	cw.nextMorning()
	if cw.w.Alive(col) || cw.tally.beesDied != 8 {
		t.Fatalf("colony alive %v, %v bees died; want false, 8", cw.w.Alive(col), cw.tally.beesDied)
	}
}

func TestColonyUpdateNoAdults(t *testing.T) {
	// brood without any adult dies. An adult male keeps the brood alive.
	cw := newColonyWorld(t, 1)
	orphan := cw.addColony(nil)
	cw.addBee(orphan, stage.Egg, caste.Undefined, 3, 0)
	withMale := cw.addColony(nil)
	brood := cw.addBee(withMale, stage.Pupa, caste.Worker, 6, 0)
	cw.addBee(withMale, stage.Adult, caste.Male, 1, 3)

	cw.nextMorning()
	if cw.w.Alive(orphan) {
		t.Fatal("the colony with brood only must die")
	}
	if cw.tally.beesDied != 3 {
		t.Errorf("%v bees died, want the 3 eggs", cw.tally.beesDied)
	}
	if !cw.w.Alive(brood) || !cw.w.Alive(withMale) {
		t.Error("brood with an adult male must survive")
	}
}

func TestColonyUpdateSwitchPoint(t *testing.T) {
	// in the eusocial phase and with larvaWorkerRatio < 3: p = 0.13 per day
	cw := newColonyWorld(t, 5)
	const n = 2000
	cols := make([]ecs.Entity, n)
	for i := range cols {
		cols[i] = cw.addColony(nil)
		cw.addBee(cols[i], stage.Adult, caste.Worker, 1, 1) // eusocial phase starts this morning
	}
	highRatio := cw.addColony(nil)
	cw.addBee(highRatio, stage.Adult, caste.Worker, 2, 1)
	cw.addBee(highRatio, stage.Larva, caste.Undefined, 6, 0) // ratio 3, not < 3

	cw.nextMorning()
	switched := 0
	for _, e := range cols {
		if c, _, _ := cw.colony(e); c.SwitchPointDate == day.At(201) {
			switched++
		} else if c.SwitchPointDate.IsSet() {
			t.Fatalf("switchPointDate %v", c.SwitchPointDate)
		}
	}
	if p := float64(switched) / n; math.Abs(p-0.13) > 0.03 {
		t.Errorf("switched %d of %d colonies (%.3f), want about 0.13", switched, n, p)
	}
	if c, _, _ := cw.colony(highRatio); c.SwitchPointDate.IsSet() {
		t.Error("larvaWorkerRatio 3 must not switch")
	}
	// a switch point is set only once
	date := map[ecs.Entity]day.Date{}
	for _, e := range cols {
		c, _, _ := cw.colony(e)
		date[e] = c.SwitchPointDate
	}
	cw.nextMorning()
	for _, e := range cols {
		if c, _, _ := cw.colony(e); date[e].IsSet() && c.SwitchPointDate != date[e] {
			t.Fatal("switch point changed after it was set")
		}
	}
}

func TestCompetitionPointDate(t *testing.T) {
	// worked examples and the bounds
	cases := []struct{ eus, qpd, want day.Abs }{
		{120, 130, 143}, // y = 22.5, round 142.5 = 143
		{120, 180, 165}, // 178 - 120 = 58 > 45
		{120, 115, 132}, // x = -5: y = 12
		{120, 120, 136}, // x = 0: y = 15.5, round 135.5 = 136
		{120, 162, 165}, // x = 42: y = 44.9, round 164.9 = 165
		{120, 163, 165}, // x = 43: y = 45.6, round 165.6 = 166 > 165
	}
	for _, c := range cases {
		col := comp.Colony{EusocialPhaseDate: day.At(c.eus), QueenProductionDate: day.At(c.qpd)}
		if got := competitionPointDate(&col, &defaults.Colony); got != c.want {
			t.Errorf("competitionPointDate(eus %v, qpd %v) = %v, want %v", c.eus, c.qpd, got, c.want)
		}
	}

	// set once both dates are set
	cw := newColonyWorld(t, 1)
	set := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.EusocialPhaseDate, c.QueenProductionDate = day.At(120), day.At(130)
	})
	unset := cw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.EusocialPhaseDate = day.At(120) })
	for _, e := range []ecs.Entity{set, unset} {
		cw.addBee(e, stage.Adult, caste.Queen, 1, 300)
	}
	cw.nextMorning()
	if c, _, _ := cw.colony(set); c.CompetitionPointDate != day.At(143) {
		t.Errorf("competitionPointDate %v, want 143", c.CompetitionPointDate)
	}
	if c, _, _ := cw.colony(unset); c.CompetitionPointDate.IsSet() {
		t.Errorf("without queenProductionDate: competitionPointDate %v, want unset", c.CompetitionPointDate)
	}
}

func TestColonyUpdateRangeLists(t *testing.T) {
	// the lists in flower and range are rebuilt only if Flowering.Changed
	cw := newColonyWorld(t, 1)
	fsMap := ecs.NewMap3[comp.Foodsource, comp.FoodsourceState, comp.Layer](cw.w)
	inFlower := fsMap.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{Pollen: 1, Nectar: 1}, &comp.Layer{IsMasterpatch: true})
	notInFlower := fsMap.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{}, &comp.Layer{IsMasterpatch: true})
	for _, e := range []ecs.Entity{inFlower, notInFlower} {
		_, _, l := fsMap.Get(e)
		l.Masterpatch = e
	}
	col := cw.addColony(nil)
	cw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	_, _, _, _, _, r := cw.colonyMap.Get(col)
	r.Layers = []ecs.Entity{inFlower, notInFlower}

	cw.nextMorning() // no update
	if _, _, _, _, _, r := cw.colonyMap.Get(col); len(r.PollenPatches) != 0 {
		t.Fatal("lists must not change without Flowering.Changed")
	}
	cw.today.Changed = true
	cw.nextMorning()
	_, _, _, _, _, r = cw.colonyMap.Get(col)
	if len(r.PollenPatches) != 1 || r.PollenPatches[0] != inFlower ||
		len(r.NectarPatches) != 1 {
		t.Errorf("lists after the update: %+v", *r)
	}
}
