package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/internal/check"

	"github.com/mlange-42/ark/ecs"
)

// checkWorld is a colonyWorld with a hibernating queen without colony and a colony with its queen, brood, a worker
// and a male. The colony statistics are set by one colonyUpdate morning.
func checkWorld(t *testing.T) (*colonyWorld, ecs.Entity) {
	t.Helper()
	cw := newColonyWorld(t, 1)
	hibernating := cw.addBee(ecs.Entity{}, stage.Adult, caste.Queen, 1, 100)
	_, b, _, _, _, _ := cw.beeMap.Get(hibernating)
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

func TestBroodTotals(t *testing.T) {
	cw, col := checkWorld(t)
	s := &broodTotals{}
	s.Initialize(cw.w)
	got := ecs.GetResource[globals.BroodTotals](cw.w)
	// [CHK-02] the totals of the colonies of this morning, finding 39 (see trans/netlogo-findings.md#finding-39):
	// an egg laid after the morning statistics is not counted
	cw.addBee(col, stage.Egg, caste.Undefined, 2, 0)
	s.Update(cw.w)
	want := globals.BroodTotals{Eggs: 5, Larvae: 4, Pupae: 3}
	if *got != want {
		t.Errorf("totals:\n got %+v\nwant %+v", *got, want)
	}
}

func TestMorningExampleFarm(t *testing.T) {
	// The morning systems on the example farm for one year, with the queens of step 3. Without brood and
	// foraging every colony consists of its queen only and keeps its start energy, so no colony dies.
	a := newExampleFarmSetup(t, 1, &initialQueensSetup{}, &tick{}, &calendarUpdate{}, &flowerRefill{},
		&queenEmergence{}, &colonyUpdate{}, &broodTotals{})
	w := a.World
	tally := newTally(w) // Setup kills no colony
	cal := ecs.GetResource[globals.Calendar](w)
	// the nectar of all foodsources (NetLogo: NectarAvailableTotalToday_l, removed in step F4). Without foraging the
	// value of the morning stays until the evening.
	states := ecs.NewFilter1[comp.FoodsourceState](w)
	nectar := func() float64 {
		sum := 0.0
		q := states.Query()
		for q.Next() {
			sum += q.Get().Nectar
		}
		return sum
	}
	if cal.Day != 0 || nectar() != 0 {
		t.Fatalf("Setup: Day %v, nectar %v; want 0, 0", cal.Day, nectar())
	}
	filter := ecs.NewFilter3[comp.Colony, comp.ColonyCounts, comp.ColonyRange](w)
	bees := ecs.NewFilter1[comp.Bee](w)
	maxNectar, withLists := 0.0, 0
	for range 400 {
		a.Update()
		if err := check.Day(w, false); err != nil {
			t.Fatalf("day %v: %v", cal.Day, err)
		}
		maxNectar = max(maxNectar, nectar())
		q := filter.Query()
		for q.Next() {
			c, n, r := q.Get()
			if n.Adults != 1 || c.EusocialPhaseDate.IsSet() {
				t.Fatalf("colony: allAdults %v, eusocialPhaseDate %v", n.Adults, c.EusocialPhaseDate)
			}
			if len(r.PollenPatches)+
				len(r.NectarPatches) > 0 {
				withLists++
			}
		}
		if tally.coloniesDied != 0 {
			t.Fatalf("day %v: %d colonies died", cal.Day, tally.coloniesDied)
		}
		bq := bees.Query()
		for bq.Next() {
			if b := bq.Get(); b.Caste != caste.Queen {
				bq.Close()
				t.Fatalf("day %v: a bee of caste %v; only queens exist", cal.Day, b.Caste)
			}
		}
	}
	if maxNectar == 0 || withLists == 0 {
		t.Errorf("max nectar %v myl, colony days with foodsources in flower and range %d; want > 0", maxNectar, withLists)
	}
}
