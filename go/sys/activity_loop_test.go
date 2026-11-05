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

	"github.com/mlange-42/ark/ecs"
)

// activityWorld extends broodWorld (step 4b) with the systems of steps 5 and 6. The foraging period is 8 h
// (Weather = "Constant 8 hrs" in netlogo/_SYSTEM_Example_Farm_Parameters.csv), the foraging mortality "high".
type activityWorld struct {
	*broodWorld
	wake    *WakeUpBees
	loop    *ActivityLoop
	eod     *EndOfDayResting
	active  *globals.ActiveBees
	daytime *globals.Daytime
}

func newActivityWorld(t *testing.T, seed uint64) *activityWorld {
	t.Helper()
	aw := &activityWorld{broodWorld: newBroodWorld(t, seed), wake: &WakeUpBees{},
		loop: &ActivityLoop{CheckActiveCount: true}, eod: &EndOfDayResting{}}
	aw.cal.DailyForagingPeriod_s = 8 * 3600
	ecs.AddResource(aw.w, &globals.Derived{MortalityForager_per_s: 1e-05}) // ForagingMortalityModel "high" [PAR-04]
	aw.wake.Initialize(aw.w)
	aw.loop.Initialize(aw.w)
	aw.eod.Initialize(aw.w)
	aw.active = ecs.GetResource[globals.ActiveBees](aw.w)
	aw.daytime = ecs.GetResource[globals.Daytime](aw.w)
	return aw
}

// adult adds an adult bee with the thresholds of a situation [THR].
func (aw *activityWorld) adult(col ecs.Entity, c caste.Caste, number int, weight_mg float64, situation string) ecs.Entity {
	e := aw.addBee(col, stage.Adult, c, number, 5)
	b, _, th, _, _ := aw.bee(e)
	b.Weight_mg = weight_mg
	*th = thresholds(situation)
	return e
}

func TestNextActiveBeeREP(t *testing.T) {
	// [NEXT-04] worked examples
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
			b.PersonalTime_s = time
			names[e] = string(rune('A' + i))
			list = append(list, e)
		}
		aw.active.List = list
		first := aw.loop.NextActiveBeeREP()
		got := ""
		for _, e := range aw.active.List {
			got += names[e]
		}
		if got != tc.want || first != aw.active.List[0] {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestWakeUpBees(t *testing.T) {
	// [MORN-08] active adult workers and queens, also without colony, wake up at 1 to 1800 s
	aw := newActivityWorld(t, 2)
	col := aw.addColony(nil)
	want := []ecs.Entity{aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)}
	for range 200 {
		want = append(want, aw.adult(col, caste.Worker, 1, 100, sitWorker))
	}
	young := aw.adult(col, caste.Queen, 1, 800, sitYoungQueen)
	searching := aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	b, _, _, _, _ := aw.bee(searching)
	b.ColonyID = ecs.Entity{}
	want = append(want, young, searching)
	aw.adult(col, caste.Male, 1, 100, sitEgg)
	hibernating := aw.adult(ecs.Entity{}, caste.Queen, 1, 800, sitYoungQueen)
	b, _, _, _, _ = aw.bee(hibernating)
	b.Activity = activity.Hibernate
	aw.addBrood(col, stage.Pupa, caste.Worker, 5, 20, 100, 10)

	aw.wake.Update(aw.w)

	if len(aw.active.List) != len(want) {
		t.Fatalf("%d bees in the queue, want %d", len(aw.active.List), len(want))
	}
	last := 0.0
	for _, e := range aw.active.List {
		if !slices.Contains(want, e) {
			t.Errorf("bee %v should not be in the queue", e)
		}
		b, _, _, _, _ := aw.bee(e)
		if b.PersonalTime_s < 1 || b.PersonalTime_s > 1800 || b.PersonalTime_s != float64(int(b.PersonalTime_s)) {
			t.Errorf("personalTime_s %v, want a whole second from 1 to 1800", b.PersonalTime_s)
		}
		if b.PersonalTime_s < last {
			t.Errorf("queue not sorted: %v after %v", b.PersonalTime_s, last)
		}
		last = b.PersonalTime_s
	}
}

func TestStimuli(t *testing.T) {
	// worked examples of [STEG-01], [STNU-01], [STNE-01] to [STNE-03], [STPO-01] to [STPO-03]
	sp := terrestris(t)
	if !almostEqual(sp.MinPollenStore_g, 0.3744) {
		t.Fatalf("minPollenStore_g %v, want 0.3744", sp.MinPollenStore_g)
	}
	for _, tc := range []struct {
		name   string
		pollen float64
		counts comp.ColonyCounts
		want   float64
	}{
		{"founding, pollen above minimum", 0.375, comp.ColonyCounts{AllAdults: 1}, 1},
		{"founding, pollen at minimum", sp.MinPollenStore_g, comp.ColonyCounts{AllAdults: 1}, 0},
		{"founding, larvae present", 1, comp.ColonyCounts{AllAdults: 1, AllLarvae: 12}, 0},
		{"workers present", 0, comp.ColonyCounts{AllAdults: 2, AllEggs: 12}, 1},
		// finding 45 (see trans/netlogo-findings.md#finding-45): a male counts like a worker
		{"queen and one male", 0, comp.ColonyCounts{AllAdults: 2, AllEggs: 12}, 1},
	} {
		stores := comp.ColonyStores{PollenStore_g: tc.pollen}
		if got := StimEgglayingREP(&stores, &tc.counts, sp); got != tc.want {
			t.Errorf("StimEgglayingREP %s: %v, want %v", tc.name, got, tc.want)
		}
	}

	eggs := comp.ColonyCounts{AllEggs: 12}
	for _, tc := range []struct{ summed, want float64 }{{6.6, 1}, {6.7, 0}} {
		stores := comp.ColonyStores{SummedIncubationToday_kJ: tc.summed}
		if got := StimNursingREP(&stores, &eggs, sp); got != tc.want {
			t.Errorf("StimNursingREP, 12 eggs need 6.67 kJ, %v kJ received: %v, want %v", tc.summed, got, tc.want)
		}
	}

	const sunrise, period = 8 * 3600, 8 * 3600
	for _, tc := range []struct{ store, daytime, want float64 }{
		{39.44, 36000, 1}, {39.45, 36000, 0}, {0, 28799, 0}, {0, 28800, 1}, {0, 57600, 1}, {0, 57601, 0},
	} {
		stores := comp.ColonyStores{EnergyNeedToday_kJ: 0.65472, EnergyStore_kJ: tc.store}
		got := StimForagingNectarREP(&stores, tc.daytime, sunrise, period)
		if got != tc.want || !almostEqual(stores.IdealEnergyStore_kJ, 39.6416) {
			t.Errorf("StimForagingNectarREP store %v at %v s: %v (ideal %v), want %v (ideal 39.6416)",
				tc.store, tc.daytime, got, stores.IdealEnergyStore_kJ, tc.want)
		}
	}
	for _, tc := range []struct{ store, daytime, want float64 }{{0.8978, 36000, 1}, {0.898, 36000, 0}, {0, 7200, 0}} {
		stores := comp.ColonyStores{PollenNeedLarvaeToday_g: 0.1056, PollenStore_g: tc.store}
		got := StimForagingPollenREP(&stores, sp, tc.daytime, sunrise, period)
		if got != tc.want || !almostEqual(stores.IdealPollenStore_g, 0.9024) {
			t.Errorf("StimForagingPollenREP store %v at %v s: %v (ideal %v), want %v (ideal 0.9024)",
				tc.store, tc.daytime, got, stores.IdealPollenStore_g, tc.want)
		}
	}
}

func TestActivityProcChoice(t *testing.T) {
	// [ACT-03] the last matching task wins (finding 11, see trans/netlogo-findings.md#finding-11), [ACT-04]
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
		// finding 44 (see trans/netlogo-findings.md#finding-44): nursing wins and becomes resting
		{"Psithyrus with social thresholds", sitQueenSocialPhase, true, false, true, true, 36000, activity.Resting},
		{"Psithyrus with social thresholds, brood warm", sitQueenSocialPhase, true, false, true, false, 36000, activity.Egglaying},
	} {
		aw := newActivityWorld(t, 1)
		col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			s.PollenStore_g, s.EnergyStore_kJ = 1000, 1000
			if tc.lowStores {
				s.PollenStore_g, s.EnergyStore_kJ = 0.3, 10
			}
			n.AllEggs, n.AllAdults = 12, 3
			if !tc.heatNeed {
				s.SummedIncubationToday_kJ = 100
			}
		})
		e := aw.adult(col, caste.Queen, 1, 800, tc.situation)
		b, g, _, _, _ := aw.bee(e)
		g.AllelesList, g.SpermathecaList = []float64{10, 20}, []float64{30}
		if tc.situation == sitWorker {
			b.Caste = caste.Worker
		}
		if tc.psithyrus {
			b.SpeciesName = "Psithyrus"
		}
		if tc.noColony {
			b.ColonyID = ecs.Entity{}
		}
		b.PersonalTime_s = tc.daytime
		aw.daytime.Daytime_s = tc.daytime
		aw.loop.ActivityProc(aw.w, e)
		if !aw.w.Alive(e) {
			continue // died on a search flight [COST-01]
		}
		// a forager ends her trip in another activity: without known patches she returns empty [SRCH-03]
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

func TestActivityProcTime(t *testing.T) {
	// [ACT-05] resting 1800 s, nursing 2880 s, egg laying 24 h, a failed search 2500 m / 5 m/s [COST-01]
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
			n.AllEggs, n.AllAdults = 12, 3
			if !tc.heatNeed {
				s.SummedIncubationToday_kJ = 100
			}
		})
		e := aw.adult(col, caste.Queen, 1, 800, tc.situation)
		b, _, _, _, _ := aw.bee(e)
		b.PersonalTime_s = tc.daytime
		aw.daytime.Daytime_s = tc.daytime
		aw.loop.ActivityProc(aw.w, e)
		if b, _, _, _, _ := aw.bee(e); b.PersonalTime_s != tc.want {
			t.Errorf("%s: personalTime_s %v, want %v", tc.name, b.PersonalTime_s, tc.want)
		}
	}
}

func TestBroodIncubationProc(t *testing.T) {
	// [INC-01], [INC-02] worked example: a queen of 800 mg and 12 eggs: about 0.5 kJ, 0.0417 kJ per egg.
	// The NetLogo constant is truncated, so the heat is 0.499999998 kJ.
	heat := 2880 * 0.000000217013888 * 800.0
	aw := newActivityWorld(t, 1)
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.EnergyStore_kJ = 10
		n.AllEggs = 12
	})
	other := aw.addColony(nil)
	eggs := aw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0.1)
	foreign := aw.addBrood(other, stage.Egg, caste.Undefined, 12, 2, 1.5, 0.1)
	queen := aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)
	aw.loop.broodOf = beesByColony(aw.loop.beeFilter, func(b *comp.Bee) bool { return b.Stage != stage.Adult })

	aw.loop.BroodIncubationProc(aw.w, queen)

	if _, _, _, _, d := aw.bee(eggs); !almostEqual(d.CumulIncubationReceived_kJ, 0.1+heat/12) {
		t.Errorf("egg incubation %v kJ, want %v", d.CumulIncubationReceived_kJ, 0.1+heat/12)
	}
	if _, _, _, _, d := aw.bee(foreign); d.CumulIncubationReceived_kJ != 0.1 {
		t.Errorf("egg of another colony was heated: %v kJ", d.CumulIncubationReceived_kJ)
	}
	_, s, _ := aw.colony(col)
	if heat < 0.4999999 || heat > 0.5 {
		t.Errorf("heat %v kJ, want about 0.5", heat)
	}
	if !almostEqual(s.SummedIncubationToday_kJ, heat) || !almostEqual(s.EnergyStore_kJ, 10-heat) {
		t.Errorf("summedIncubationToday_kJ %v, energy store %v; want %v, %v", s.SummedIncubationToday_kJ, s.EnergyStore_kJ, heat, 10-heat)
	}
	if b, _, _, _, _ := aw.bee(queen); b.PersonalTime_s != 2880 {
		t.Errorf("personalTime_s %v, want 2880", b.PersonalTime_s)
	}
}

func TestActivityLoopDay(t *testing.T) {
	// one day of a social colony: the queen and 20 workers heat 12 eggs, then the queen lays.
	// Finding 46 (see trans/netlogo-findings.md#finding-46): the new eggs get no heat today.
	aw := newActivityWorld(t, 4)
	col := aw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 1000, 1000 // no foraging
		n.AllEggs, n.AllAdults, n.AllAdultWorkers = 12, 21, 20
	})
	old := aw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0)
	queen := aw.adult(col, caste.Queen, 1, 800, sitQueenSocialPhase)
	_, g, _, _, _ := aw.bee(queen)
	g.AllelesList, g.SpermathecaList = []float64{10, 20}, []float64{30}
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
		if e == old && d.CumulIncubationReceived_kJ < aw.sp.DevQuotaIncubationToday_kJ {
			t.Errorf("old eggs got %v kJ, want at least the quota %v", d.CumulIncubationReceived_kJ, aw.sp.DevQuotaIncubationToday_kJ)
		}
		if e != old && d.CumulIncubationReceived_kJ != 0 {
			t.Errorf("new eggs got %v kJ, want 0", d.CumulIncubationReceived_kJ)
		}
	}
	// [GO-08] the day ends when the first bee of the queue is beyond 24 h; [GO-11] everybody rests
	for _, e := range aw.active.List {
		b, _, _, _, _ := aw.bee(e)
		if b.PersonalTime_s <= aw.constants.CallItaDay_s || b.Activity != activity.Resting {
			t.Errorf("after the loop: personalTime_s %v, activity %v", b.PersonalTime_s, b.Activity)
		}
	}
}

func TestActivityLoopQueenWithoutColony(t *testing.T) {
	// [GO-10] one rest, then her day is over. The loop needs a colony [GO-06].
	aw := newActivityWorld(t, 1)
	col := aw.addColony(nil)
	aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	searching := aw.adult(col, caste.Queen, 1, 800, sitQueenInitiationPhase)
	b, _, _, _, _ := aw.bee(searching)
	b.ColonyID = ecs.Entity{}
	aw.wake.Update(aw.w)
	aw.loop.Update(aw.w)
	if b, _, _, _, _ := aw.bee(searching); b.PersonalTime_s != aw.constants.CallItaDay_s+1 {
		t.Errorf("queen without colony: personalTime_s %v, want %v", b.PersonalTime_s, aw.constants.CallItaDay_s+1)
	}
}

func TestActivityLoopExampleFarm(t *testing.T) {
	// the full schedule of steps 0 to 6 on the example farm for 200 days. CheckActiveCount compares the length of
	// the queue with NetLogo's count in every turn [GO-06]. The checks of the state [CHK-03] run every day.
	if testing.Short() {
		t.Skip("200 days of the example farm")
	}
	loop := &ActivityLoop{CheckActiveCount: true}
	a := newExampleFarmSetup(t, 1, &CreateInitialQueens{}, &Tick{}, &UpdateMorning{}, &UpdateFoodsources{},
		&EmergenceNewQueens{}, &WakeUpBees{}, &UpdateColonies{}, &CheckNumbers{},
		&NeedNectarPollenLarvaeToday{}, loop, &EndOfDayResting{}, &FeedLarvae{}, &QueenProductionDate{},
		&Development{}, &MortalityBrood{})
	active := ecs.GetResource[globals.ActiveBees](a.World)
	maxQueue := 0
	for range 200 {
		a.Update()
		if err := check.Day(a.World, false); err != nil {
			t.Fatal(err)
		}
		maxQueue = max(maxQueue, len(active.List))
	}
	if maxQueue == 0 {
		t.Error("no bee was ever active")
	}
	totals := ecs.GetResource[globals.Totals](a.World)
	t.Logf("largest activity queue: %d bees. Day 200: brood %v/%v/%v",
		maxQueue, totals.TotalEggs, totals.TotalLarvae, totals.TotalPupae)
}

func TestNextActiveBeeREPAfterDeath(t *testing.T) {
	// [NEXT-01] the previous bee died: the next bee has not been active yet and is re-inserted. With equal times
	// it moves behind the next bee (trans/data-model.md, "Daily activity queue").
	aw := newActivityWorld(t, 1)
	col := aw.addColony(nil)
	names := map[ecs.Entity]string{}
	list := []ecs.Entity{}
	for i := range 4 {
		e := aw.adult(col, caste.Worker, 1, 100, sitWorker)
		b, _, _, _, _ := aw.bee(e)
		b.PersonalTime_s = 10
		names[e] = string(rune('A' + i))
		list = append(list, e)
	}
	aw.active.List = list
	death := newBeeDeath(aw.w)
	death.DieProc(aw.w, list[0], "test") // A
	aw.loop.NextActiveBeeREP()
	got := ""
	for _, e := range aw.active.List {
		got += names[e]
	}
	if got != "CBD" {
		t.Errorf("queue %s, want CBD", got)
	}
}

func TestEndOfDayResting(t *testing.T) {
	// [GO-11] all adult, non-hibernating females rest, also bees whose day ended earlier
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
