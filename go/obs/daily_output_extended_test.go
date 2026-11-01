package obs

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestDailyOutputExtended(t *testing.T) {
	w := ecs.NewWorld()
	ticks := globals.Ticks{Ticks: 1}
	ecs.AddResource(w, &ticks)
	output := globals.DailyOutput{TotalQueens: 500}
	ecs.AddResource(w, &output)
	totals := globals.Totals{TotalForagingTripsToday: 40}
	ecs.AddResource(w, &totals)
	counters := globals.Counters{TotalAdultQueensEverProduced: 7, TotalForagersEverDiedForaging: 3}
	ecs.AddResource(w, &counters)

	bees := ecs.NewMap1[comp.Bee](w)
	for _, b := range []comp.Bee{
		{Number: 12, Stage: stage.Egg},
		{Number: 10, Stage: stage.Larva, Caste: caste.Worker},
		{Number: 5, Stage: stage.Pupa, Caste: caste.Worker},
		{Number: 20, Stage: stage.Adult, Caste: caste.Worker},
		{Number: 1, Stage: stage.Adult, Caste: caste.Worker},
		{Number: 1, Stage: stage.Adult, Caste: caste.Queen},
		{Number: 4, Stage: stage.Adult, Caste: caste.Male},
	} {
		bees.NewEntity(&b)
	}
	colonies := ecs.NewMap2[comp.Colony, comp.ColonyStores](w)
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: 150}, &comp.ColonyStores{EnergyStore_kJ: 100, PollenStore_g: 2})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: 160}, &comp.ColonyStores{EnergyStore_kJ: 50, PollenStore_g: 1})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: params.NotSetHigh}, &comp.ColonyStores{EnergyStore_kJ: 10})

	o := DailyOutputExtended{}
	o.Initialize(w)
	header := o.Header()
	daily := (&DailyOutput{}).Header()
	if got := header[:len(daily)]; !slices.Equal(got, daily) {
		t.Fatalf("the first columns must be those of DailyOutput, got %v", got)
	}

	want := map[string]float64{
		"ticks": 1, "TotalQueens": 500,
		"Eggs": 12, "Larvae": 10, "Pupae": 5, "AdultWorkers": 21,
		"EnergyStore_kJ": 160, "PollenStore_g": 3,
		"ForagingTripsToday": 40, "ForagerDeathsToday": 3, "TotalAdultQueensEverProduced": 7,
		"MeanCompetitionPointDate": 155,
	}
	check(t, header, o.Values(w), want)

	// day 2: the foragers that died are counted per day
	ticks.Ticks = 2
	counters.TotalForagersEverDiedForaging = 5
	want["ticks"], want["ForagerDeathsToday"] = 2, 2
	check(t, header, o.Values(w), want)
}

func TestDailyOutputExtendedNoCompetitionPoint(t *testing.T) {
	w := ecs.NewWorld()
	ecs.AddResource(w, &globals.Ticks{})
	ecs.AddResource(w, &globals.DailyOutput{})
	ecs.AddResource(w, &globals.Totals{})
	ecs.AddResource(w, &globals.Counters{})
	ecs.NewMap2[comp.Colony, comp.ColonyStores](w).NewEntity(
		&comp.Colony{CompetitionPointDate: params.NotSetHigh}, &comp.ColonyStores{})

	o := DailyOutputExtended{}
	o.Initialize(w)
	check(t, o.Header(), o.Values(w), map[string]float64{"MeanCompetitionPointDate": 0})
}

// check compares the values of the columns in want.
func check(t *testing.T, header []string, values []float64, want map[string]float64) {
	t.Helper()
	if len(values) != len(header) {
		t.Fatalf("%d values for %d columns", len(values), len(header))
	}
	for name, v := range want {
		i := slices.Index(header, name)
		if i < 0 {
			t.Errorf("no column %s", name)
			continue
		}
		if values[i] != v {
			t.Errorf("%s: got %v, want %v", name, values[i], v)
		}
	}
}
