package sys

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/internal/check"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// activityWorld extends broodWorld with wakeUp, activities and nightRest. The foraging period is 8 h
// (Weather = "Constant 8 hrs" in netlogo/_SYSTEM_Example_Farm_Parameters.csv), the foraging mortality "high".
type activityWorld struct {
	*broodWorld
	wake    *wakeUp
	loop    *activities
	eod     *nightRest
	active  *globals.ActiveBees
	daytime *globals.Daytime
}

func newActivityWorld(t *testing.T, seed uint64) *activityWorld {
	t.Helper()
	aw := &activityWorld{broodWorld: newBroodWorld(t, seed), wake: &wakeUp{},
		loop: &activities{CheckActiveCount: true}, eod: &nightRest{}}
	aw.cal.ForagingPeriod = 8 * 3600
	ecs.AddResource(aw.w, &params.Derived{ForagerMortality: 1e-05}) // ForagingMortalityModel "high"
	aw.wake.Initialize(aw.w)
	aw.loop.Initialize(aw.w)
	aw.eod.Initialize(aw.w)
	aw.active = ecs.GetResource[globals.ActiveBees](aw.w)
	aw.daytime = ecs.GetResource[globals.Daytime](aw.w)
	return aw
}

// adult adds an adult bee with the thresholds of a situation.
// weight is in [mg].
func (aw *activityWorld) adult(col ecs.Entity, c caste.Caste, number int, weight float64, situation string) ecs.Entity {
	e := aw.addBee(col, stage.Adult, c, number, 5)
	b, _, th, _, _ := aw.bee(e)
	b.Weight = weight
	*th = thresholds(&defaults.Thresholds, situation)
	return e
}

func TestNextActiveBee(t *testing.T) {
	// worked examples
	for _, tc := range []struct {
		name  string
		times []float64 // A first
		want  string
	}{
		{"A 1810", []float64{1810, 900, 1200, 2000}, "BCAD"},
		{"A 2500", []float64{2500, 900, 1200, 2000}, "BCDA"},
		{"A 2000, tie with the last bee", []float64{2000, 900, 1200, 2000}, "BCAD"},
		{"three ties", []float64{10, 10, 10}, "BAC"},
		{"A smallest", []float64{5, 10, 20}, "ABC"},
		{"two bees", []float64{30, 20}, "BA"},
		{"one bee", []float64{30}, "A"},
	} {
		aw := newActivityWorld(t, 1)
		col := aw.addColony(nil)
		names := map[ecs.Entity]string{}
		list := []ecs.Entity{}
		for i, time := range tc.times {
			e := aw.adult(col, caste.Worker, 1, 100, sitWorker)
			b, _, _, _, _ := aw.bee(e)
			b.PersonalTime = time
			names[e] = string(rune('A' + i))
			list = append(list, e)
		}
		// A is the first bee of the queue, then she acts and gets her new time
		b, _, _, _, _ := aw.bee(list[0])
		b.PersonalTime = 0
		aw.active.Set(list, aw.loop.personalTime)
		b.PersonalTime = tc.times[0]
		first := aw.loop.nextActiveBee()
		got := ""
		for _, e := range aw.active.Entities() {
			got += names[e]
		}
		if got != tc.want || first != aw.active.Entities()[0] {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestWakeUp(t *testing.T) {
	// active adult workers and queens, also without colony, wake up at 1 to 1800 s
	aw := newActivityWorld(t, 2)
	col := aw.addColony(nil)
	want := []ecs.Entity{aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)}
	for range 200 {
		want = append(want, aw.adult(col, caste.Worker, 1, 100, sitWorker))
	}
	young := aw.adult(col, caste.Queen, 1, 800, sitYoungQueen)
	searching := aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	b, _, _, _, _ := aw.bee(searching)
	b.Colony = ecs.Entity{}
	want = append(want, young, searching)
	aw.adult(col, caste.Male, 1, 100, sitEgg)
	hibernating := aw.adult(ecs.Entity{}, caste.Queen, 1, 800, sitYoungQueen)
	b, _, _, _, _ = aw.bee(hibernating)
	b.Activity = activity.Hibernate
	aw.addBrood(col, stage.Pupa, caste.Worker, 5, 20, 100, 10)

	aw.wake.Update(aw.w)

	if aw.active.Len() != len(want) {
		t.Fatalf("%d bees in the queue, want %d", aw.active.Len(), len(want))
	}
	last := 0.0
	for _, e := range aw.active.Entities() {
		if !slices.Contains(want, e) {
			t.Errorf("bee %v should not be in the queue", e)
		}
		b, _, _, _, _ := aw.bee(e)
		if b.PersonalTime < 1 || b.PersonalTime > 1800 || b.PersonalTime != float64(int(b.PersonalTime)) {
			t.Errorf("PersonalTime %v, want a whole second from 1 to 1800", b.PersonalTime)
		}
		if b.PersonalTime < last {
			t.Errorf("queue not sorted: %v after %v", b.PersonalTime, last)
		}
		last = b.PersonalTime
	}
}

func TestStimuli(t *testing.T) {
	// worked examples
	sp := terrestris(t)
	if !almostEqual(sp.MinPollenStore, 0.3744) {
		t.Fatalf("MinPollenStore %v, want 0.3744", sp.MinPollenStore)
	}
	for _, tc := range []struct {
		name   string
		pollen float64
		counts comp.ColonyCounts
		want   float64
	}{
		{"founding, pollen above minimum", 0.375, comp.ColonyCounts{Adults: 1}, 1},
		{"founding, pollen at minimum", sp.MinPollenStore, comp.ColonyCounts{Adults: 1}, 0},
		{"founding, larvae present", 1, comp.ColonyCounts{Adults: 1, Larvae: 12}, 0},
		{"workers present", 0, comp.ColonyCounts{Adults: 2, Eggs: 12}, 1},
		// a male counts like a worker
		{"queen and one male", 0, comp.ColonyCounts{Adults: 2, Eggs: 12}, 1},
	} {
		stores := comp.ColonyStores{Pollen: tc.pollen}
		if got := egglayingStimulus(&stores, &tc.counts, sp); got != tc.want {
			t.Errorf("egglayingStimulus %s: %v, want %v", tc.name, got, tc.want)
		}
	}

	eggs := comp.ColonyCounts{Eggs: 12}
	for _, tc := range []struct{ summed, want float64 }{{6.6, 1}, {6.7, 0}} {
		stores := comp.ColonyStores{IncubationToday: tc.summed}
		if got := nursingStimulus(&stores, &eggs, sp); got != tc.want {
			t.Errorf("nursingStimulus, 12 eggs need 6.67 kJ, %v kJ received: %v, want %v", tc.summed, got, tc.want)
		}
	}

	const sunrise, period = 8 * 3600, 8 * 3600
	for _, tc := range []struct{ store, daytime, want float64 }{
		{39.44, 36000, 1}, {39.45, 36000, 0}, {0, 28799, 0}, {0, 28800, 1}, {0, 57600, 1}, {0, 57601, 0},
	} {
		stores := comp.ColonyStores{EnergyNeedToday: 0.65472, Energy: tc.store}
		got := nectarStimulus(&stores, &defaults.Stimuli, tc.daytime, sunrise, period)
		if got != tc.want || !almostEqual(stores.IdealEnergy, 39.6416) {
			t.Errorf("nectarStimulus store %v at %v s: %v (ideal %v), want %v (ideal 39.6416)",
				tc.store, tc.daytime, got, stores.IdealEnergy, tc.want)
		}
	}
	for _, tc := range []struct{ store, daytime, want float64 }{{0.8978, 36000, 1}, {0.898, 36000, 0}, {0, 7200, 0}} {
		stores := comp.ColonyStores{LarvaePollenNeedToday: 0.1056, Pollen: tc.store}
		got := pollenStimulus(&stores, sp, &defaults.Stimuli, tc.daytime, sunrise, period)
		if got != tc.want || !almostEqual(stores.IdealPollen, 0.9024) {
			t.Errorf("pollenStimulus store %v at %v s: %v (ideal %v), want %v (ideal 0.9024)",
				tc.store, tc.daytime, got, stores.IdealPollen, tc.want)
		}
	}
}

func TestActChoice(t *testing.T) {
	// the last matching task wins
	for _, tc := range []struct {
		name      string
		situation string
		psithyrus bool
		noColony  bool
		lowStores bool    // both foraging stimuli 1 during the foraging period
		heatNeed  bool    // nursing stimulus 1
		daytime   float64 // 10:00 is in the foraging period, 3:00 is not
		want      activity.Activity
	}{
		{"founding queen, all stimuli 1", sitQueenInitiationPhase, false, false, true, true, 36000, activity.NectarForaging},
		{"founding queen at night", sitQueenInitiationPhase, false, false, true, true, 10800, activity.Nursing},
		{"founding queen at night, brood warm", sitQueenInitiationPhase, false, false, true, false, 10800, activity.Egglaying},
		{"social queen, low stores", sitQueenSocialPhase, false, false, true, true, 36000, activity.Nursing},
		{"social queen, brood warm", sitQueenSocialPhase, false, false, true, false, 36000, activity.Egglaying},
		{"worker, low stores", sitWorker, false, false, true, true, 36000, activity.NectarForaging},
		{"worker at night", sitWorker, false, false, true, true, 10800, activity.Nursing},
		{"worker at night, brood warm", sitWorker, false, false, true, false, 10800, activity.Resting},
		{"young queen", sitYoungQueen, false, false, true, true, 36000, activity.Resting},
		{"queen without colony", sitQueenInitiationPhase, false, true, true, true, 36000, activity.Resting},
		{"Psithyrus with Psith thresholds", sitPsith, true, false, true, true, 36000, activity.Egglaying},
		// nursing wins and becomes resting
		{"Psithyrus with social thresholds", sitQueenSocialPhase, true, false, true, true, 36000, activity.Resting},
		{"Psithyrus with social thresholds, brood warm", sitQueenSocialPhase, true, false, true, false, 36000, activity.Egglaying},
	} {
		aw := newActivityWorld(t, 1)
		col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			s.Pollen, s.Energy = 1000, 1000
			if tc.lowStores {
				s.Pollen, s.Energy = 0.3, 10
			}
			n.Eggs, n.Adults = 12, 3
			if !tc.heatNeed {
				s.IncubationToday = 100
			}
		})
		e := aw.adult(col, caste.Queen, 1, 800, tc.situation)
		psithSpecies := addParasiticSpecies(aw.w, aw.sp)
		b, g, _, _, _ := aw.bee(e)
		g.Alleles, g.Spermatheca = []float64{10, 20}, []float64{30}
		if tc.situation == sitWorker {
			b.Caste = caste.Worker
		}
		if tc.psithyrus {
			b.Species = psithSpecies // in the host colony of species aw.species
		}
		if tc.noColony {
			b.Colony = ecs.Entity{}
		}
		b.PersonalTime = tc.daytime
		aw.daytime.Time = tc.daytime
		aw.loop.act(aw.w, e)
		if !aw.w.Alive(e) {
			continue // died on a search flight
		}
		// a forager ends her trip in another activity: without known patches she returns empty
		b, _, _, fo, _ := aw.bee(e)
		got := b.Activity
		if got == activity.ReturningEmpty {
			got = activity.NectarForaging
			if fo.PollenForager {
				got = activity.PollenForaging
			}
		}
		if got != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestActTime(t *testing.T) {
	// resting 1800 s, nursing 2880 s, egg laying 24 h, a failed search 2500 m / 5 m/s
	for _, tc := range []struct {
		name      string
		situation string
		daytime   float64
		heatNeed  bool
		want      float64
	}{
		{"resting", sitYoungQueen, 3600, true, 3600 + 1800},
		{"nursing", sitQueenSocialPhase, 3600, true, 3600 + 2880},
		{"egg laying", sitQueenSocialPhase, 3600, false, 3600 + 86400},
		{"foraging without known patch: failed search", sitQueenInitiationPhase, 36000, true, 36000 + 2500/5},
	} {
		aw := newActivityWorld(t, 1)
		col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			n.Eggs, n.Adults = 12, 3
			if !tc.heatNeed {
				s.IncubationToday = 100
			}
		})
		e := aw.adult(col, caste.Queen, 1, 800, tc.situation)
		b, _, _, _, _ := aw.bee(e)
		b.PersonalTime = tc.daytime
		aw.daytime.Time = tc.daytime
		aw.loop.act(aw.w, e)
		if b, _, _, _, _ := aw.bee(e); b.PersonalTime != tc.want {
			t.Errorf("%s: PersonalTime %v, want %v", tc.name, b.PersonalTime, tc.want)
		}
	}
}

func TestIncubateBrood(t *testing.T) {
	// worked example: a queen of 800 mg and 12 eggs: about 0.5 kJ, 0.0417 kJ per egg.
	// The NetLogo constant is truncated, so the heat is 0.499999998 kJ.
	heat := 2880 * 0.000000217013888 * 800.0
	aw := newActivityWorld(t, 1)
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.Energy = 10
		n.Eggs = 12
	})
	other := aw.addColony(nil)
	eggs := aw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0.1)
	foreign := aw.addBrood(other, stage.Egg, caste.Undefined, 12, 2, 1.5, 0.1)
	queen := aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)
	aw.loop.broodOf = beesByColony(aw.loop.broodOf, aw.loop.beeFilter, func(b *comp.Bee) bool { return b.Stage != stage.Adult })

	aw.loop.incubateBrood(aw.w, queen)

	if _, _, _, _, d := aw.bee(eggs); !almostEqual(d.IncubationReceived, 0.1+heat/12) {
		t.Errorf("egg incubation %v kJ, want %v", d.IncubationReceived, 0.1+heat/12)
	}
	if _, _, _, _, d := aw.bee(foreign); d.IncubationReceived != 0.1 {
		t.Errorf("egg of another colony was heated: %v kJ", d.IncubationReceived)
	}
	_, s, _ := aw.colony(col)
	if heat < 0.4999999 || heat > 0.5 {
		t.Errorf("heat %v kJ, want about 0.5", heat)
	}
	if !almostEqual(s.IncubationToday, heat) || !almostEqual(s.Energy, 10-heat) {
		t.Errorf("IncubationToday %v, energy store %v; want %v, %v", s.IncubationToday, s.Energy, heat, 10-heat)
	}
	if b, _, _, _, _ := aw.bee(queen); b.PersonalTime != 2880 {
		t.Errorf("PersonalTime %v, want 2880", b.PersonalTime)
	}
}

func TestActivitiesDay(t *testing.T) {
	// one day of a social colony: the queen and 20 workers heat 12 eggs, then the queen lays.
	// The new eggs get no heat today.
	aw := newActivityWorld(t, 4)
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.Pollen, s.Energy = 1000, 1000 // no foraging
		n.Eggs, n.Adults, n.AdultWorkers = 12, 21, 20
	})
	old := aw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0)
	queen := aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)
	_, g, _, _, _ := aw.bee(queen)
	g.Alleles, g.Spermatheca = []float64{10, 20}, []float64{30}
	for range 20 {
		aw.adult(col, caste.Worker, 1, 100, sitWorker)
	}

	aw.wake.Update(aw.w)
	aw.loop.Update(aw.w)
	aw.eod.Update(aw.w)

	eggs := aw.eggsOf(col)
	if len(eggs) != 2 {
		t.Fatalf("%d egg cohorts, want 2 (the old one and one new batch)", len(eggs))
	}
	for _, e := range eggs {
		_, _, _, _, d := aw.bee(e)
		if e == old && d.IncubationReceived < aw.sp.DailyIncubationQuota {
			t.Errorf("old eggs got %v kJ, want at least the quota %v", d.IncubationReceived, aw.sp.DailyIncubationQuota)
		}
		if e != old && d.IncubationReceived != 0 {
			t.Errorf("new eggs got %v kJ, want 0", d.IncubationReceived)
		}
	}
	// the day ends when the first bee of the queue is beyond 24 h, then everybody rests
	for _, e := range aw.active.Entities() {
		b, _, _, _, _ := aw.bee(e)
		if b.PersonalTime <= aw.constants.Activities.CallItADay || b.Activity != activity.Resting {
			t.Errorf("after the loop: PersonalTime %v, activity %v", b.PersonalTime, b.Activity)
		}
	}
}

func TestActivitiesQueenWithoutColony(t *testing.T) {
	// one rest, then her day is over. The loop needs a colony.
	aw := newActivityWorld(t, 1)
	col := aw.addColony(nil)
	aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	searching := aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	b, _, _, _, _ := aw.bee(searching)
	b.Colony = ecs.Entity{}
	aw.wake.Update(aw.w)
	aw.loop.Update(aw.w)
	if b, _, _, _, _ := aw.bee(searching); b.PersonalTime != aw.constants.Activities.CallItADay+1 {
		t.Errorf("queen without colony: PersonalTime %v, want %v", b.PersonalTime, aw.constants.Activities.CallItADay+1)
	}
}

func TestActivitiesExampleFarm(t *testing.T) {
	// the parts below on the example farm for 200 days. CheckActiveCount compares the length of
	// the queue with NetLogo's count in every turn. The checks of the state run every day.
	if testing.Short() {
		t.Skip("200 days of the example farm")
	}
	loop := &activities{CheckActiveCount: true}
	a := newExampleFarmSetup(t, 1, &initialQueensSetup{}, &tick{}, &calendarUpdate{}, &flowerRefill{},
		&queenEmergence{}, &wakeUp{}, &colonyUpdate{}, &broodTotals{},
		&larvaeNeeds{}, loop, &nightRest{}, &larvaeFeeding{}, &queenProductionDate{},
		&development{}, &broodMortality{})
	active := ecs.GetResource[globals.ActiveBees](a.World)
	maxQueue := 0
	for range 200 {
		a.Update()
		if err := check.Day(a.World, false); err != nil {
			t.Fatal(err)
		}
		maxQueue = max(maxQueue, active.Len())
	}
	if maxQueue == 0 {
		t.Error("no bee was ever active")
	}
	totals := ecs.GetResource[globals.BroodTotals](a.World)
	t.Logf("largest activity queue: %d bees. Day 200: brood %v/%v/%v",
		maxQueue, totals.Eggs, totals.Larvae, totals.Pupae)
}

func TestNextActiveBeeAfterDeath(t *testing.T) {
	// the previous bee died: the next bee has not been active yet and is re-inserted. With equal times
	// it moves behind the next bee.
	aw := newActivityWorld(t, 1)
	col := aw.addColony(nil)
	names := map[ecs.Entity]string{}
	list := []ecs.Entity{}
	for i := range 4 {
		e := aw.adult(col, caste.Worker, 1, 100, sitWorker)
		b, _, _, _, _ := aw.bee(e)
		b.PersonalTime = 10
		names[e] = string(rune('A' + i))
		list = append(list, e)
	}
	aw.active.Set(list, aw.loop.personalTime)
	death := newBeeDeath(aw.w)
	death.die(aw.w, list[0], "test") // A
	aw.loop.nextActiveBee()
	got := ""
	for _, e := range aw.active.Entities() {
		got += names[e]
	}
	if got != "CBD" {
		t.Errorf("queue %s, want CBD", got)
	}
}

func TestNightRest(t *testing.T) {
	// all adult, non-hibernating females rest, also bees whose day ended earlier
	aw := newActivityWorld(t, 1)
	col := aw.addColony(nil)
	worker := aw.adult(col, caste.Worker, 1, 100, sitWorker)
	queen := aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)
	male := aw.adult(col, caste.Male, 1, 100, sitEgg)
	hibernating := aw.adult(ecs.Entity{}, caste.Queen, 1, 800, sitYoungQueen)
	pupa := aw.addBrood(col, stage.Pupa, caste.Worker, 5, 20, 100, 10)
	set := map[ecs.Entity]activity.Activity{worker: activity.CollectNectar, queen: activity.Egglaying,
		male: activity.Nursing, hibernating: activity.Hibernate, pupa: activity.Nursing}
	for e, a := range set {
		b, _, _, _, _ := aw.bee(e)
		b.Activity = a
	}
	aw.eod.Update(aw.w)
	want := map[ecs.Entity]activity.Activity{worker: activity.Resting, queen: activity.Resting,
		male: activity.Nursing, hibernating: activity.Hibernate, pupa: activity.Nursing}
	for e, a := range want {
		if b, _, _, _, _ := aw.bee(e); b.Activity != a {
			t.Errorf("bee %v: %v, want %v", e, b.Activity, a)
		}
	}
}
