package sys

import (
	"math"
	"slices"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// broodWorld extends colonyWorld (step 4a) with the brood systems of step 4b. Brood is laid with EgglayingProc
// or created by hand, because ActivityProc (step 5) does not exist yet.
type broodWorld struct {
	*colonyWorld
	need *NeedNectarPollenLarvaeToday
	feed *FeedLarvae
	qpd  *QueenProductionDate
	dev  *Development
	morb *MortalityBrood
	eggs eggLayer
	sp   *comp.Species
}

func newBroodWorld(t *testing.T, seed uint64) *broodWorld {
	t.Helper()
	bw := &broodWorld{colonyWorld: newColonyWorld(t, seed),
		need: &NeedNectarPollenLarvaeToday{}, feed: &FeedLarvae{}, qpd: &QueenProductionDate{},
		dev: &Development{}, morb: &MortalityBrood{}}
	bw.need.Initialize(bw.w)
	bw.feed.Initialize(bw.w)
	bw.qpd.Initialize(bw.w)
	bw.dev.Initialize(bw.w)
	bw.morb.Initialize(bw.w)
	bw.eggs = newEggLayer(bw.w, bw.dev.rng)
	bw.sp = bw.speciesMap(bw.species)
	return bw
}

func (bw *broodWorld) speciesMap(e ecs.Entity) *comp.Species {
	return bw.sys.speciesMap.Get(e)
}

// addParasiticSpecies adds a copy of the species sp named Psithyrus with Parasitic set and returns its entity.
func addParasiticSpecies(w *ecs.World, sp *comp.Species) ecs.Entity {
	psith := *sp
	psith.Name, psith.Parasitic = "Psithyrus", true
	return ecs.NewMap1[comp.Species](w).NewEntity(&psith)
}

func (bw *broodWorld) beeFilter() *ecs.Filter1[comp.Bee] {
	return ecs.NewFilter1[comp.Bee](bw.w)
}

// eggsOf returns the egg cohorts of a colony.
func (bw *broodWorld) eggsOf(col ecs.Entity) []ecs.Entity {
	result := []ecs.Entity{}
	query := bw.beeFilter().Query()
	for query.Next() {
		if b := query.Get(); b.Stage == stage.Egg && b.ColonyID == col {
			result = append(result, query.Entity())
		}
	}
	return result
}

// bee returns the components of a bee.
func (bw *broodWorld) bee(e ecs.Entity) (*comp.Bee, *comp.Genetics, *comp.Thresholds, *comp.Forager, *comp.Development) {
	_, _, b, g, th, f, d := bw.beeMap.Get(e)
	return b, g, th, f, d
}

// addBrood adds brood of a colony with the given caste, broodAge, weight and incubation.
func (bw *broodWorld) addBrood(col ecs.Entity, st stage.Stage, c caste.Caste, number, broodAge int, weight_mg, incubation_kJ float64) ecs.Entity {
	e := bw.addBee(col, st, c, number, 0)
	b, _, _, _, d := bw.bee(e)
	b.BroodAge, b.Weight_mg, d.CumulIncubationReceived_kJ = broodAge, weight_mg, incubation_kJ
	return e
}

// evening runs the evening systems [GO-13] to [GO-16] in the order of model/default.go.
func (bw *broodWorld) evening() {
	bw.feed.Update(bw.w)
	bw.qpd.Update(bw.w)
	bw.dev.Update(bw.w)
	bw.morb.Update(bw.w)
}

func TestMaxWeightGainToday_mg_REP(t *testing.T) {
	// [WGAIN-01] worked example: 1.5 * 1.88^k, workers capped at 249.5 mg on the ninth feeding, queens at 980 mg
	// on the eleventh.
	sp := terrestris(t)
	for _, tc := range []struct {
		caste  caste.Caste
		capped int // feeding on which the cap binds
		cap    float64
	}{{caste.Worker, 9, 249.5}, {caste.Male, 9, 249.5}, {caste.Undefined, 9, 249.5}, {caste.Queen, 11, 980}} {
		b := comp.Bee{Caste: tc.caste, Weight_mg: sp.DevWeightEgg_mg}
		for k := 1; k <= 12; k++ {
			b.Weight_mg += MaxWeightGainToday_mg_REP(&b, sp)
			want := math.Min(1.5*math.Pow(1.88, float64(k)), tc.cap)
			if !almostEqual(b.Weight_mg, want) {
				t.Errorf("%v after %d feedings: %v mg, want %v", tc.caste, k, b.Weight_mg, want)
			}
			if k == tc.capped-1 && b.Weight_mg >= tc.cap {
				t.Errorf("%v reached the cap before feeding %d", tc.caste, tc.capped)
			}
		}
	}
}

func TestNeedNectarPollenLarvaeToday(t *testing.T) {
	// [NEED-01] worked example: 12 larvae of 10 mg: 0.1056 g pollen, 0.65472 kJ
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	other := bw.addColony(nil)
	bw.addBrood(col, stage.Larva, caste.Undefined, 12, 6, 10, 0)
	bw.addBrood(col, stage.Egg, caste.Undefined, 12, 2, 1.5, 0)  // eggs need nothing
	bw.addBrood(col, stage.Pupa, caste.Worker, 5, 15, 200, 0)    // pupae need nothing
	bw.addBrood(other, stage.Larva, caste.Worker, 1, 10, 100, 0) // other colony: 88 mg
	bw.need.Update(bw.w)
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenNeedLarvaeToday_g, 0.1056) || !almostEqual(s.EnergyNeedToday_kJ, 0.1056*6.2) {
		t.Errorf("pollen need %v g, energy need %v kJ; want 0.1056, 0.65472", s.PollenNeedLarvaeToday_g, s.EnergyNeedToday_kJ)
	}
	_, s, _ = bw.colony(other)
	if !almostEqual(s.PollenNeedLarvaeToday_g, 0.088) {
		t.Errorf("other colony: pollen need %v g, want 0.088", s.PollenNeedLarvaeToday_g)
	}
}

func TestBeesByColonyReusesGroups(t *testing.T) {
	// step F10: the second call empties the groups of the first and drops the colony without larvae
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	other := bw.addColony(nil)
	a := bw.addBrood(col, stage.Larva, caste.Undefined, 1, 6, 10, 0)
	b := bw.addBrood(other, stage.Larva, caste.Undefined, 1, 6, 10, 0)
	isLarva := func(x *comp.Bee) bool { return x.Stage == stage.Larva }
	groups := beesByColony(nil, bw.beeFilter(), isLarva)
	bw.w.RemoveEntity(b)
	c := bw.addBrood(col, stage.Larva, caste.Undefined, 1, 6, 10, 0)
	groups = beesByColony(groups, bw.beeFilter(), isLarva)
	if _, ok := groups[other]; ok || len(groups) != 1 || len(groups[col]) != 2 ||
		!slices.Contains(groups[col], a) || !slices.Contains(groups[col], c) {
		t.Errorf("groups %v, want only colony %v with %v and %v", groups, col, a, c)
	}
}

func TestFeedLarvae(t *testing.T) {
	// [FEED-01] worked example: pollen 0.3 of 0.6 g, energy 20 of 10 kJ: factor 0.5
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.IdealPollenStore_g, s.EnergyStore_kJ, s.IdealEnergyStore_kJ = 0.3, 0.6, 20, 10
	})
	larva := bw.addBrood(col, stage.Larva, caste.Undefined, 12, 6, 10, 0)
	pupa := bw.addBrood(col, stage.Pupa, caste.Worker, 3, 15, 200, 0)
	// [FEED-01] an ideal store of 0: no feeding
	noIdeal := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ, s.IdealEnergyStore_kJ = 5, 50, 10
	})
	unfed := bw.addBrood(noIdeal, stage.Larva, caste.Worker, 1, 10, 10, 0)
	// [FEED-02] the cap compares ONE larva with the whole store: each of 12 larvae gets the full 1 mg
	small := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.IdealPollenStore_g, s.EnergyStore_kJ, s.IdealEnergyStore_kJ = 0.001, 0.001, 50, 10
	})
	capped := bw.addBrood(small, stage.Larva, caste.Worker, 12, 10, 10, 0)

	bw.feed.Update(bw.w)

	// gain 8.8 mg, half of it: 4.4 mg per larva, 52.8 mg for the colony
	if b, _, _, _, _ := bw.bee(larva); !almostEqual(b.Weight_mg, 14.4) {
		t.Errorf("larva weight %v mg, want 14.4", b.Weight_mg)
	}
	if b, _, _, _, _ := bw.bee(pupa); b.Weight_mg != 200 {
		t.Errorf("pupa was fed: %v mg", b.Weight_mg)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenStore_g, 0.3-0.0528) || !almostEqual(s.EnergyStore_kJ, 20-0.0528*6.2) {
		t.Errorf("stores %v g, %v kJ; want %v, %v", s.PollenStore_g, s.EnergyStore_kJ, 0.3-0.0528, 20-0.0528*6.2)
	}
	if b, _, _, _, _ := bw.bee(unfed); b.Weight_mg != 10 {
		t.Errorf("larva without ideal store was fed: %v mg", b.Weight_mg)
	}
	if b, _, _, _, _ := bw.bee(capped); !almostEqual(b.Weight_mg, 11) {
		t.Errorf("capped larva %v mg, want 11", b.Weight_mg)
	}
	if _, s, _ := bw.colony(small); !almostEqual(s.PollenStore_g, 0.001-0.012) {
		t.Errorf("small store %v g, want %v (negative)", s.PollenStore_g, 0.001-0.012)
	}
}

func TestQueenProductionDate(t *testing.T) {
	// [QPD-02] worked example: switch point on day 100
	for _, tc := range []struct {
		name       string
		ticks      day.Abs
		larvalAge3 int // broodAge of a larva at larval day 3, 0 for none
		ratio      float64
		switchDay  day.Date
		want       bool
		wantDate   day.Date // unset without queen production
	}{
		{"no larvae, day 104", 104, 0, 1, day.At(100), false, day.Date{}},
		{"no larvae, day 105", 105, 0, 1, day.At(100), true, day.At(95)},
		{"larvae of broodAge 8, day 102", 102, 8, 1, day.At(100), false, day.Date{}},
		{"larvae of broodAge 8, day 103", 103, 8, 1, day.At(100), true, day.At(95)},
		{"larva:worker ratio 3", 120, 0, 3, day.At(100), false, day.Date{}},
		{"no switch point", 120, 0, 1, day.Date{}, false, day.Date{}},
	} {
		bw := newBroodWorld(t, 1)
		bw.ticks.Ticks = tc.ticks
		col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			c.SwitchPointDate = tc.switchDay
			n.LarvaWorkerRatio = tc.ratio
		})
		if tc.larvalAge3 > 0 {
			e := bw.addBrood(col, stage.Larva, caste.Worker, 12, tc.larvalAge3, 50, 0)
			_, _, _, _, d := bw.bee(e)
			d.CumulTimeLarva_d = 3
			// a larva at another larval day is ignored
			e = bw.addBrood(col, stage.Larva, caste.Undefined, 12, 20, 50, 0)
			_, _, _, _, d = bw.bee(e)
			d.CumulTimeLarva_d = 2
		}
		bw.qpd.Update(bw.w)
		c, _, _ := bw.colony(col)
		if c.QueenProduction != tc.want || c.QueenProductionDate != tc.wantDate {
			t.Errorf("%s: queenProduction %v, date %v; want %v, %v", tc.name, c.QueenProduction, c.QueenProductionDate, tc.want, tc.wantDate)
		}
	}
}

func TestQueenProductionDateBeforeDevelopment(t *testing.T) {
	// finding 42 (see trans/netlogo-findings.md#finding-42): the date is the laying day of the last worker larvae
	bw := newBroodWorld(t, 1)
	bw.ticks.Ticks = 110
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.SwitchPointDate = day.At(100)
		n.LarvaWorkerRatio = 1
	})
	worker := bw.addBrood(col, stage.Larva, caste.Worker, 12, 8, 30, 0) // larval day 3 since last evening
	_, _, _, _, d := bw.bee(worker)
	d.CumulTimeLarva_d = 3
	female := bw.addBrood(col, stage.Larva, caste.Undefined, 12, 7, 20, 0) // reaches larval day 3 tonight
	_, _, _, _, d = bw.bee(female)
	d.CumulTimeLarva_d = 2

	bw.qpd.Update(bw.w)
	bw.dev.Update(bw.w)

	c, _, _ := bw.colony(col)
	if !c.QueenProduction || c.QueenProductionDate != day.At(102) {
		t.Errorf("queenProduction %v, date %v; want true, 102", c.QueenProduction, c.QueenProductionDate)
	}
	if b, _, _, _, _ := bw.bee(worker); b.Caste != caste.Worker {
		t.Errorf("larva laid on day 102 is %v, want worker", b.Caste)
	}
	if b, _, _, _, _ := bw.bee(female); b.Caste != caste.Queen {
		t.Errorf("larva laid on day 103 is %v, want queen", b.Caste)
	}
}

func TestDetermineCaste_REP(t *testing.T) {
	// finding 40 (see trans/netlogo-findings.md#finding-40): every female larva becomes a queen
	sp := terrestris(t)
	if sp.Dev_Q_DeterminationWeight_mg != 0 {
		t.Fatalf("dev_Q_DeterminationWeight_mg = %v, want 0", sp.Dev_Q_DeterminationWeight_mg)
	}
	producing := comp.Colony{QueenProduction: true}
	for _, w := range []float64{0, 1.5, 10, 600} {
		if got := DetermineCaste_REP(w, sp, &producing); got != caste.Queen {
			t.Errorf("weight %v mg, queen production: %v, want queen", w, got)
		}
		if got := DetermineCaste_REP(w, sp, &comp.Colony{}); got != caste.Worker {
			t.Errorf("weight %v mg, no queen production: %v, want worker", w, got)
		}
	}
	heavy := *sp
	heavy.Dev_Q_DeterminationWeight_mg = 50
	if got := DetermineCaste_REP(49, &heavy, &producing); got != caste.Worker {
		t.Errorf("weight criterion: %v, want worker", got)
	}
}

func TestDevelopmentLarvaePrecedence(t *testing.T) {
	// finding 41 (see trans/netlogo-findings.md#finding-41): (worker or male) and incubation and age and weight
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	light := bw.addBrood(col, stage.Larva, caste.Worker, 1, 13, 50, 100) // below 62.4 mg
	lightMale := bw.addBrood(col, stage.Larva, caste.Male, 1, 13, 50, 100)
	young := bw.addBrood(col, stage.Larva, caste.Worker, 1, 10, 100, 100) // broodAge 11 < 12.9
	cold := bw.addBrood(col, stage.Larva, caste.Worker, 1, 13, 100, 7)    // below 7.17 kJ
	ready := bw.addBrood(col, stage.Larva, caste.Worker, 1, 13, 63, 7.2)
	readyMale := bw.addBrood(col, stage.Larva, caste.Male, 1, 13, 63, 7.2)
	bw.dev.Update(bw.w)
	for _, tc := range []struct {
		name string
		e    ecs.Entity
		want stage.Stage
	}{{"light worker", light, stage.Larva}, {"light male", lightMale, stage.Larva}, {"young worker", young, stage.Larva},
		{"cold worker", cold, stage.Larva}, {"worker", ready, stage.Pupa}, {"male", readyMale, stage.Pupa}} {
		if b, _, _, _, _ := bw.bee(tc.e); b.Stage != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, b.Stage, tc.want)
		}
	}
}

// layAndRaise lays one batch with a queen on day 200 and raises it with full food and slightly more than the
// incubation quota every day. It returns the evening (days after laying) of each stage change.
func layAndRaise(t *testing.T, queenProduction bool) (bw *broodWorld, eggs ecs.Entity, events map[string]int) {
	t.Helper()
	bw = newBroodWorld(t, 3)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.QueenProduction = queenProduction
		s.PollenStore_g, s.IdealPollenStore_g, s.EnergyStore_kJ, s.IdealEnergyStore_kJ = 1000, 1, 1000, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	_, g, _, _, _ := bw.bee(queen)
	g.AllelesList, g.SpermathecaList = []float64{10, 20}, []float64{30}

	bw.eggs.EgglayingProc(queen)
	eggs = ecs.Entity{}
	query := bw.beeFilter().Query()
	for query.Next() {
		if query.Get().Stage == stage.Egg {
			eggs = query.Entity()
		}
	}
	if eggs.IsZero() {
		t.Fatal("no eggs laid")
	}

	events = map[string]int{}
	quota := bw.sp.DevQuotaIncubationToday_kJ * 1.01
	for k := 0; k <= 40; k++ {
		if !bw.w.Alive(eggs) {
			t.Fatalf("brood died on the evening of day L + %d", k)
		}
		b, _, _, _, d := bw.bee(eggs)
		if b.Stage == stage.Adult {
			break
		}
		d.CumulIncubationReceived_kJ += quota // during the day (BroodIncubationProc, step 5)
		before := *b
		bw.evening()
		b, _, _, _, _ = bw.bee(eggs)
		if before.Stage != b.Stage {
			events[b.Stage.String()] = k
		}
		if before.Caste != b.Caste {
			events["caste"] = k
			events[b.Caste.String()] = k
		}
		if b.Stage == stage.Pupa && before.Stage == stage.Larva {
			events["pupationWeight"] = int(math.Round(b.Weight_mg * 100))
		}
		bw.ticks.Ticks++
	}
	return bw, eggs, events
}

func TestDevelopmentWorkerTimeline(t *testing.T) {
	// worked example [DEVL-02], [DEVL-03], [DEVP-01]: egg laid on day L, larva in the evening of L + 4, caste on
	// L + 7, pupa on L + 12 (8 feedings, 234.07 mg), adult on L + 23
	bw, e, events := layAndRaise(t, false)
	want := map[string]int{"larva": 4, "caste": 7, "worker": 7, "pupa": 12, "adult": 23, "pupationWeight": 23407}
	for k, v := range want {
		if events[k] != v {
			t.Errorf("%s: %v, want %v (all: %v)", k, events[k], v, events)
		}
	}
	b, _, th, f, _ := bw.bee(e)
	if b.Stage != stage.Adult || b.Number != 12 || b.AdultAge != 0 || *th != thresholds(sitWorker) {
		t.Errorf("new workers: stage %v, number %v, adultAge %v, thresholds %+v", b.Stage, b.Number, b.AdultAge, *th)
	}
	if f.Cropvolume_myl != CropAndPelletSizeREP("nectar", b.Weight_mg, bw.sp) || f.GlossaLength_mm != ProboscisLengthREP(b.Weight_mg, bw.sp) {
		t.Errorf("crop %v, glossa %v", f.Cropvolume_myl, f.GlossaLength_mm)
	}
	if bw.tally.queensEmerged != 0 {
		t.Errorf("%v queens emerged, want 0", bw.tally.queensEmerged)
	}
}

func TestDevelopmentQueenTimeline(t *testing.T) {
	// worked example [DEVL-03], [DEVP-04]: queen pupa on L + 16 (capped at 980 mg), adult on L + 31, emerging
	// date in year 2 between Day 1 and seasonStop 305
	bw, e, events := layAndRaise(t, true)
	want := map[string]int{"larva": 4, "caste": 7, "queen": 7, "pupa": 16, "adult": 31, "pupationWeight": 98000}
	for k, v := range want {
		if events[k] != v {
			t.Errorf("%s: %v, want %v (all: %v)", k, events[k], v, events)
		}
	}
	b, _, th, f, _ := bw.bee(e)
	if !b.EmergingDate.IsSet() || b.EmergingDate.Day() <= 365 || b.EmergingDate.Day() > 365+305 {
		t.Errorf("emergingDate %v, want a day in (365, 670]", b.EmergingDate)
	}
	if *th != thresholds(sitYoungQueen) || f.PollenPellets_g != CropAndPelletSizeREP("pollen", 980, bw.sp) {
		t.Errorf("thresholds %+v, pollen load %v", *th, f.PollenPellets_g)
	}
	// event.QueenEmerged for the output (obs.Counters.TotalAdultQueensEverProduced)
	if bw.tally.queensEmerged != 12 {
		t.Errorf("%v queens emerged, want 12", bw.tally.queensEmerged)
	}
}

func TestQueenEmergingDateDistribution(t *testing.T) {
	// [DEVP-04] worked example: emergingDate = 365 + round random-normal 91 28, mean about 456
	bw := newBroodWorld(t, 5)
	col := bw.addColony(nil)
	sum := 0.0
	n := 2000
	for range n {
		e := bw.addBrood(col, stage.Pupa, caste.Queen, 1, 40, 980, 100)
		b, _, _, _, _ := bw.bee(e)
		b.EmergingDate = day.Date{}
		bw.dev.Development_PupaeProc(e)
		b, _, _, _, _ = bw.bee(e)
		if !b.EmergingDate.IsSet() || b.EmergingDate.Day() <= 365 || b.EmergingDate.Day() > 670 {
			t.Fatalf("emergingDate %v outside (365, 670]", b.EmergingDate)
		}
		sum += float64(b.EmergingDate.Day())
	}
	if mean := sum / float64(n); math.Abs(mean-456) > 2 {
		t.Errorf("mean emergingDate %v, want about 456", mean)
	}
}

func TestDevelopmentAdultMortality(t *testing.T) {
	// [MORA-01] workers die at adultAge 61, males at 31, queens never
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	w60 := bw.addBee(col, stage.Adult, caste.Worker, 1, 60)
	w59 := bw.addBee(col, stage.Adult, caste.Worker, 1, 59)
	m30 := bw.addBee(col, stage.Adult, caste.Male, 1, 30)
	m29 := bw.addBee(col, stage.Adult, caste.Male, 1, 29)
	q := bw.addBee(col, stage.Adult, caste.Queen, 1, 600)
	bw.dev.Update(bw.w)
	for _, tc := range []struct {
		name  string
		e     ecs.Entity
		alive bool
	}{{"worker 61", w60, false}, {"worker 60", w59, true}, {"male 31", m30, false}, {"male 30", m29, true}, {"queen 601", q, true}} {
		if bw.w.Alive(tc.e) != tc.alive {
			t.Errorf("%s: alive %v, want %v", tc.name, bw.w.Alive(tc.e), tc.alive)
		}
	}
	if bw.tally.beesDied != 2 {
		t.Errorf("%v bees died, want 2", bw.tally.beesDied)
	}
	if b, _, _, _, _ := bw.bee(q); b.AdultAge != 601 {
		t.Errorf("queen adultAge %v, want 601", b.AdultAge)
	}
	expectPanic(t, "adult older than 700 d", func() {
		bw.addBee(col, stage.Adult, caste.Queen, 1, 700)
		bw.dev.Update(bw.w)
	})
}

func TestDevelopmentSocialPhaseThresholds(t *testing.T) {
	// [MORA-02] finding 44 (see trans/netlogo-findings.md#finding-44): also a Psithyrus queen
	bw := newBroodWorld(t, 1)
	social := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { n.AllAdultWorkers = 5 })
	founding := bw.addColony(nil)
	host := bw.addBee(social, stage.Adult, caste.Queen, 1, 30)
	psith := bw.addBee(social, stage.Adult, caste.Queen, 1, 30)
	alone := bw.addBee(founding, stage.Adult, caste.Queen, 1, 30)
	psithSpecies := addParasiticSpecies(bw.w, bw.sp)
	b, _, th, _, _ := bw.bee(psith)
	b.SpeciesID = psithSpecies
	*th = thresholds(sitPsith)
	for _, e := range []ecs.Entity{host, alone} {
		_, _, th, _, _ := bw.bee(e)
		*th = thresholds(sitQueenInitiationPhase)
	}
	bw.dev.Update(bw.w)
	for _, tc := range []struct {
		name string
		e    ecs.Entity
		want string
	}{{"host queen", host, sitQueenSocialPhase}, {"Psithyrus queen", psith, sitQueenSocialPhase},
		{"queen without workers", alone, sitQueenInitiationPhase}} {
		if _, _, th, _, _ := bw.bee(tc.e); *th != thresholds(tc.want) {
			t.Errorf("%s: %+v, want %s", tc.name, *th, tc.want)
		}
	}
}

func TestDevelopmentBackgroundMortality(t *testing.T) {
	// [MORA-03] finding 22 (see trans/netlogo-findings.md#finding-22): 0 kills nobody, 0.01 kills about 99 %
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	for range 1000 {
		bw.addBee(col, stage.Adult, caste.Worker, 1, 1)
	}
	bw.dev.Update(bw.w)
	if bw.tally.beesDied != 0 {
		t.Errorf("value 0: %v bees died", bw.tally.beesDied)
	}
	bw.constants.MortalityAdultsBackground_daily = 0.01
	bw.dev.Update(bw.w)
	if died := bw.tally.beesDied; died < 975 || died > 1000 {
		t.Errorf("value 0.01: %v of 1000 bees died, want about 990", died)
	}
}

func TestMortalityBrood(t *testing.T) {
	// [MORB-01] to [MORB-04], maximum ages from the worked example: eggs 10, larvae 20.8, pupae 34.7, queen
	// larvae 29, queen pupae 47
	bw := newBroodWorld(t, 1)
	sp := bw.sp
	for _, x := range []struct{ got, want float64 }{{sp.DevAgeHatchingMax_d, 10}, {sp.DevAgePupationMax_d, 20.8},
		{sp.DevAgeEmergingMax_d, 34.7}, {sp.DevAge_Q_PupationMax_d, 29}, {sp.DevAge_Q_EmergingMax_d, 47}} {
		if !almostEqual(x.got, x.want) {
			t.Errorf("maximum age %v, want %v", x.got, x.want)
		}
	}
	col := bw.addColony(nil)
	oldEgg := bw.addBrood(col, stage.Egg, caste.Undefined, 12, 11, 1.5, 0)
	egg := bw.addBrood(col, stage.Egg, caste.Undefined, 12, 10, 1.5, 0)
	lightLarva := bw.addBrood(col, stage.Larva, caste.Worker, 3, 21, 31.2, 7.17) // weight 0.5, incubation 1
	coldQueen := bw.addBrood(col, stage.Larva, caste.Queen, 2, 30, 590, 4.72)    // weight 1, incubation 0.5
	queenLarva := bw.addBrood(col, stage.Larva, caste.Queen, 1, 29, 100, 1)      // not too old
	oldPupa := bw.addBrood(col, stage.Pupa, caste.Male, 4, 35, 100, 1)
	queenPupa := bw.addBrood(col, stage.Pupa, caste.Queen, 1, 47, 900, 1)
	oldQueenPupa := bw.addBrood(col, stage.Pupa, caste.Queen, 5, 48, 900, 1)
	bw.addBee(bw.addColony(nil), stage.Adult, caste.Queen, 1, 300)
	queenWithoutColony := bw.addBee(col, stage.Adult, caste.Queen, 1, 300)
	b, _, _, _, _ := bw.bee(queenWithoutColony)
	b.ColonyID = ecs.Entity{}

	bw.morb.Update(bw.w)

	for _, tc := range []struct {
		name  string
		e     ecs.Entity
		alive bool
	}{{"egg 11", oldEgg, false}, {"egg 10", egg, true}, {"light larva", lightLarva, false},
		{"cold queen larva", coldQueen, false}, {"queen larva 29", queenLarva, true}, {"male pupa 35", oldPupa, false},
		{"queen pupa 47", queenPupa, true}, {"queen pupa 48", oldQueenPupa, false}, {"queen without colony", queenWithoutColony, true}} {
		if bw.w.Alive(tc.e) != tc.alive {
			t.Errorf("%s: alive %v, want %v", tc.name, bw.w.Alive(tc.e), tc.alive)
		}
	}
	if bw.tally.beesDied != 12+3+2+4+5 {
		t.Errorf("%v bees died, want 26", bw.tally.beesDied)
	}

	// [MORB-02] assertion: a larva that would have pupated
	expectPanic(t, "larva with enough weight and incubation", func() {
		bw.addBrood(col, stage.Larva, caste.Worker, 1, 21, 100, 100)
		bw.morb.Update(bw.w)
	})
}

func TestMortalityBroodCompetitionPoint(t *testing.T) {
	// [MORB-04] after the competition point every egg dies, larvae survive
	bw := newBroodWorld(t, 1)
	after := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.CompetitionPointDate = day.At(199) })
	onCP := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { c.CompetitionPointDate = day.At(200) })
	eggAfter := bw.addBrood(after, stage.Egg, caste.Undefined, 12, 1, 1.5, 0)
	larvaAfter := bw.addBrood(after, stage.Larva, caste.Undefined, 12, 6, 10, 3)
	eggOnCP := bw.addBrood(onCP, stage.Egg, caste.Undefined, 12, 1, 1.5, 0)
	bw.morb.Update(bw.w)
	if bw.w.Alive(eggAfter) || !bw.w.Alive(larvaAfter) || !bw.w.Alive(eggOnCP) {
		t.Errorf("alive: egg after CP %v, larva after CP %v, egg on CP day %v; want false, true, true",
			bw.w.Alive(eggAfter), bw.w.Alive(larvaAfter), bw.w.Alive(eggOnCP))
	}
	if bw.tally.beesDied != 12 {
		t.Errorf("%v bees died, want 12", bw.tally.beesDied)
	}
}

func TestEgglaying(t *testing.T) {
	// [EGG-01] worked example: 12 eggs cost 0.018 g pollen and 0.1116 kJ. [EGGP-01] to [EGGP-03]
	bw := newBroodWorld(t, 7)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 1, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	qb, g, _, _, _ := bw.bee(queen)
	qb.PersonalTime_s = 1000
	qb.EmergingDate = day.At(120)
	g.MtDNA, g.AllelesList, g.SpermathecaList = 42, []float64{10, 20}, []float64{30}

	bw.eggs.EgglayingProc(queen)

	eggs := bw.eggsOf(col)
	if len(eggs) != 1 {
		t.Fatalf("%d egg cohorts, want 1", len(eggs))
	}
	b, g, th, f, d := bw.bee(eggs[0])
	if b.Number != 12 || b.Stage != stage.Egg || b.Caste != caste.Undefined ||
		b.ColonyID != col || b.SpeciesID != bw.species || b.Weight_mg != 1.5 || b.EmergingDate.IsSet() ||
		b.AdultAge != 0 || b.BroodAge != 0 || b.PersonalTime_s != 0 || b.Mated {
		t.Errorf("egg %+v", *b)
	}
	if g.MtDNA != 42 || len(g.AllelesList) != 2 || g.AllelesList[0] != 30 || (g.AllelesList[1] != 10 && g.AllelesList[1] != 20) ||
		len(g.SpermathecaList) != 0 {
		t.Errorf("egg genetics %+v: want mtDNA 42, alleles [30 10|20], no spermatheca", *g)
	}
	if *th != thresholds(sitEgg) || f.Cropvolume_myl != 0 || f.KnownMasterpatchesNectarList == nil || *d != (comp.Development{}) {
		t.Errorf("egg thresholds %+v, forager %+v, development %+v", *th, *f, *d)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenStore_g, 1-0.018) || !almostEqual(s.EnergyStore_kJ, 1-0.1116) {
		t.Errorf("stores %v g, %v kJ; want 0.982, 0.8884", s.PollenStore_g, s.EnergyStore_kJ)
	}
	// [EGG-03]
	if qb, _, _, _, _ := bw.bee(queen); qb.PersonalTime_s != 1000+86400 || qb.EmergingDate != day.At(120) {
		t.Errorf("queen personalTime_s %v, emergingDate %v", qb.PersonalTime_s, qb.EmergingDate)
	}
}

func TestEgglayingNotEnoughStores(t *testing.T) {
	// [EGG-02] strict >: a store equal to the cost lays nothing. [EGG-03] the 24 h are used anyway
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 0.018, 10
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	bw.eggs.EgglayingProc(queen)
	if len(bw.eggsOf(col)) != 0 {
		t.Error("eggs laid without enough pollen")
	}
	if b, _, _, _, _ := bw.bee(queen); b.PersonalTime_s != 86400 {
		t.Errorf("personalTime_s %v, want 86400", b.PersonalTime_s)
	}
	if _, s, _ := bw.colony(col); s.PollenStore_g != 0.018 {
		t.Errorf("pollen store %v, want 0.018", s.PollenStore_g)
	}
}

func TestEgglayingPloidy(t *testing.T) {
	// [EGG-02] haploid eggs only after the switch point day. SexLocus? with equal alleles: diploid male [EGGP-01].
	// The ploidy is the number of alleles (step F9).
	for _, tc := range []struct {
		name      string
		switchDay day.Date
		sexLocus  bool
		caste     caste.Caste
		alleles   int
	}{
		{"on switch point day", day.At(200), false, caste.Undefined, 2},
		{"after switch point", day.At(199), false, caste.Male, 1},
		{"sex locus, homozygous", day.Date{}, true, caste.Male, 2},
	} {
		bw := newBroodWorld(t, 1)
		bw.file.SexLocus = tc.sexLocus
		col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			c.SwitchPointDate = tc.switchDay
			s.PollenStore_g, s.EnergyStore_kJ = 1, 1
		})
		queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
		_, g, _, _, _ := bw.bee(queen)
		g.AllelesList, g.SpermathecaList = []float64{5, 5}, []float64{5}
		bw.eggs.EgglayingProc(queen)
		b, g, _, _, _ := bw.bee(bw.eggsOf(col)[0])
		if b.Caste != tc.caste || len(g.AllelesList) != tc.alleles {
			t.Errorf("%s: caste %v, alleles %v; want %v, %d alleles", tc.name, b.Caste, g.AllelesList, tc.caste, tc.alleles)
		}
	}
}

func TestEgglayingOneCohort(t *testing.T) {
	// [EGG-01] one cohort of batchsize (12) eggs for the cost of 12 eggs (0.0015 g pollen each). gobees-fast has
	// no individual-based colonies, which in NetLogo lay 12 cohorts of 1 egg for the cost of 1 egg (finding 43).
	bw := newBroodWorld(t, 1)
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		s.PollenStore_g, s.EnergyStore_kJ = 1, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	_, g, _, _, _ := bw.bee(queen)
	g.AllelesList, g.SpermathecaList = []float64{10, 20}, []float64{30}
	bw.eggs.EgglayingProc(queen)
	eggs := bw.eggsOf(col)
	if len(eggs) != 1 {
		t.Fatalf("%d egg cohorts, want 1", len(eggs))
	}
	if b, _, _, _, _ := bw.bee(eggs[0]); b.Number != 12 {
		t.Errorf("cohort of %v eggs, want 12", b.Number)
	}
	_, s, _ := bw.colony(col)
	if !almostEqual(s.PollenStore_g, 1-12*0.0015) || !almostEqual(s.EnergyStore_kJ, 1-12*0.0015*6.2) {
		t.Errorf("stores %v g, %v kJ; want the cost of 12 eggs", s.PollenStore_g, s.EnergyStore_kJ)
	}
}
