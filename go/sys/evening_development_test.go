package sys

import (
	"math"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

func TestDetermineCaste(t *testing.T) {
	// QueenDeterminationWeight is 0: every female larva becomes a queen once the colony produces queens
	sp := terrestris(t)
	if sp.QueenDeterminationWeight != 0 {
		t.Fatalf("QueenDeterminationWeight = %v, want 0", sp.QueenDeterminationWeight)
	}
	producing := comp.Colony{QueenProduction: true}
	for _, w := range []float64{0, 1.5, 10, 600} {
		if got := determineCaste(w, sp, &producing); got != caste.Queen {
			t.Errorf("weight %v mg, queen production: %v, want queen", w, got)
		}
		if got := determineCaste(w, sp, &comp.Colony{}); got != caste.Worker {
			t.Errorf("weight %v mg, no queen production: %v, want worker", w, got)
		}
	}
	heavy := *sp
	heavy.QueenDeterminationWeight = 50
	if got := determineCaste(49, &heavy, &producing); got != caste.Worker {
		t.Errorf("weight criterion: %v, want worker", got)
	}
}

func TestDevelopmentLarvaePrecedence(t *testing.T) {
	// the pupation condition is (worker or male) and incubation and age and weight
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
		s.Pollen, s.IdealPollen, s.Energy, s.IdealEnergy = 1000, 1, 1000, 1
	})
	queen := bw.addBee(col, stage.Adult, caste.Queen, 1, 30)
	_, g, _, _, _ := bw.bee(queen)
	g.Alleles, g.Spermatheca = []float64{10, 20}, []float64{30}

	bw.eggs.layEggs(queen)
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
	quota := bw.sp.DailyIncubationQuota * 1.01
	for k := 0; k <= 40; k++ {
		if !bw.w.Alive(eggs) {
			t.Fatalf("brood died on the evening of day L + %d", k)
		}
		b, _, _, _, d := bw.bee(eggs)
		if b.Stage == stage.Adult {
			break
		}
		d.IncubationReceived += quota // during the day (activities.incubateBrood)
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
			events["pupationWeight"] = int(math.Round(b.Weight * 100))
		}
		bw.ticks.Ticks++
	}
	return bw, eggs, events
}

func TestDevelopmentWorkerTimeline(t *testing.T) {
	// worked example: egg laid on day L, larva in the evening of L + 4, caste on
	// L + 7, pupa on L + 12 (8 feedings, 234.07 mg), adult on L + 23
	bw, e, events := layAndRaise(t, false)
	want := map[string]int{"larva": 4, "caste": 7, "worker": 7, "pupa": 12, "adult": 23, "pupationWeight": 23407}
	for k, v := range want {
		if events[k] != v {
			t.Errorf("%s: %v, want %v (all: %v)", k, events[k], v, events)
		}
	}
	b, _, th, f, _ := bw.bee(e)
	if b.Stage != stage.Adult || b.Number != 12 || b.AdultAge != 0 || *th != thresholds(&defaults.Thresholds, sitWorker) {
		t.Errorf("new workers: stage %v, number %v, adultAge %v, thresholds %+v", b.Stage, b.Number, b.AdultAge, *th)
	}
	if f.CropVolume != cropVolume(b.Weight, bw.sp, &defaults.Foraging) || f.GlossaLength != proboscisLength(b.Weight, bw.sp) {
		t.Errorf("crop %v, glossa %v", f.CropVolume, f.GlossaLength)
	}
	if bw.tally.queensEmerged != 0 {
		t.Errorf("%v queens emerged, want 0", bw.tally.queensEmerged)
	}
}

func TestDevelopmentQueenTimeline(t *testing.T) {
	// worked example: queen pupa on L + 16 (capped at 980 mg), adult on L + 31, emerging
	// date in year 2 between Day 1 and seasonStop 305
	bw, e, events := layAndRaise(t, true)
	want := map[string]int{"larva": 4, "caste": 7, "queen": 7, "pupa": 16, "adult": 31, "pupationWeight": 98000}
	for k, v := range want {
		if events[k] != v {
			t.Errorf("%s: %v, want %v (all: %v)", k, events[k], v, events)
		}
	}
	b, _, th, f, _ := bw.bee(e)
	if !b.EmergenceDate.IsSet() || b.EmergenceDate.Day() <= 365 || b.EmergenceDate.Day() > 365+305 {
		t.Errorf("EmergenceDate %v, want a day in (365, 670]", b.EmergenceDate)
	}
	if *th != thresholds(&defaults.Thresholds, sitYoungQueen) || f.PollenPellets != pollenLoad(980, bw.sp, &defaults.Foraging) {
		t.Errorf("thresholds %+v, pollen load %v", *th, f.PollenPellets)
	}
	// event.QueenEmerged for the output (obs.Counters.TotalAdultQueensEverProduced)
	if bw.tally.queensEmerged != 12 {
		t.Errorf("%v queens emerged, want 12", bw.tally.queensEmerged)
	}
}

func TestQueenEmergingDateDistribution(t *testing.T) {
	// worked example: emergingDate = 365 + round random-normal 91 28, mean about 456
	bw := newBroodWorld(t, 5)
	col := bw.addColony(nil)
	sum := 0.0
	n := 2000
	for range n {
		e := bw.addBrood(col, stage.Pupa, caste.Queen, 1, 40, 980, 100)
		b, _, _, _, _ := bw.bee(e)
		b.EmergenceDate = day.Date{}
		bw.dev.developPupa(e)
		b, _, _, _, _ = bw.bee(e)
		if !b.EmergenceDate.IsSet() || b.EmergenceDate.Day() <= 365 || b.EmergenceDate.Day() > 670 {
			t.Fatalf("EmergenceDate %v outside (365, 670]", b.EmergenceDate)
		}
		sum += float64(b.EmergenceDate.Day())
	}
	if mean := sum / float64(n); math.Abs(mean-456) > 2 {
		t.Errorf("mean EmergenceDate %v, want about 456", mean)
	}
}

func TestDevelopmentAdultMortality(t *testing.T) {
	// workers die at adultAge 61, males at 31, queens never
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
	// no species check: a Psithyrus queen also gets the social-phase thresholds
	bw := newBroodWorld(t, 1)
	social := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) { n.AdultWorkers = 5 })
	founding := bw.addColony(nil)
	host := bw.addBee(social, stage.Adult, caste.Queen, 1, 30)
	psith := bw.addBee(social, stage.Adult, caste.Queen, 1, 30)
	alone := bw.addBee(founding, stage.Adult, caste.Queen, 1, 30)
	psithSpecies := addParasiticSpecies(bw.w, bw.sp)
	b, _, th, _, _ := bw.bee(psith)
	b.Species = psithSpecies
	*th = thresholds(&defaults.Thresholds, sitPsith)
	for _, e := range []ecs.Entity{host, alone} {
		_, _, th, _, _ := bw.bee(e)
		*th = thresholds(&defaults.Thresholds, sitQueenInitiationPhase)
	}
	bw.dev.Update(bw.w)
	for _, tc := range []struct {
		name string
		e    ecs.Entity
		want string
	}{{"host queen", host, sitQueenSocialPhase}, {"Psithyrus queen", psith, sitQueenSocialPhase},
		{"queen without workers", alone, sitQueenInitiationPhase}} {
		if _, _, th, _, _ := bw.bee(tc.e); *th != thresholds(&defaults.Thresholds, tc.want) {
			t.Errorf("%s: %+v, want %s", tc.name, *th, tc.want)
		}
	}
}

func TestDevelopmentBackgroundMortality(t *testing.T) {
	// inverted condition: 0 kills nobody, 0.01 kills about 99 %
	bw := newBroodWorld(t, 1)
	col := bw.addColony(nil)
	for range 1000 {
		bw.addBee(col, stage.Adult, caste.Worker, 1, 1)
	}
	bw.dev.Update(bw.w)
	if bw.tally.beesDied != 0 {
		t.Errorf("value 0: %v bees died", bw.tally.beesDied)
	}
	bw.constants.Bees.AdultBackgroundMortality = 0.01
	bw.dev.Update(bw.w)
	if died := bw.tally.beesDied; died < 975 || died > 1000 {
		t.Errorf("value 0.01: %v of 1000 bees died, want about 990", died)
	}
}
