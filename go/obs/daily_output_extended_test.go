package obs

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// newWorld returns a world with the resources DailyOutputExtended reads.
func newWorld(ticks *globals.Ticks, totals *globals.Totals, counters *globals.Counters) *ecs.World {
	w := ecs.NewWorld()
	ecs.AddResource(w, ticks)
	ecs.AddResource(w, &globals.DailyOutput{TotalQueens: 500})
	ecs.AddResource(w, totals)
	ecs.AddResource(w, counters)
	ecs.AddResource(w, &globals.DeadColonies{List: make([]globals.DeadColony, 4)})
	ecs.AddResource(w, &globals.Calendar{DailyForagingPeriod_s: 28800})
	return w
}

func TestDailyOutputExtended(t *testing.T) {
	ticks := globals.Ticks{Ticks: 1}
	counters := globals.Counters{TotalAdultQueensEverProduced: 7, TotalForagersEverDiedForaging: 3,
		TotalColoniesEverProduced: 9, TotalHibernatingQueensEverRemoved: 11}
	w := newWorld(&ticks, &globals.Totals{TotalForagingTripsToday: 40}, &counters)

	// activity.Hibernate is the zero value: all other bees rest.
	bees := ecs.NewMap1[comp.Bee](w)
	for _, b := range []comp.Bee{
		{Number: 12, Stage: stage.Egg, Activity: activity.Resting},
		{Number: 10, Stage: stage.Larva, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 5, Stage: stage.Pupa, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 3, Stage: stage.Pupa, Caste: caste.Male, Ploidy: 2, Activity: activity.Resting},
		{Number: 20, Stage: stage.Adult, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Queen, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Queen, Mated: true, Activity: activity.Hibernate},
		{Number: 2, Stage: stage.Adult, Caste: caste.Queen, Mated: true, SpeciesName: "Psithyrus", Activity: activity.Resting},
		{Number: 4, Stage: stage.Adult, Caste: caste.Male, Ploidy: 1, Activity: activity.Resting},
	} {
		bees.NewEntity(&b)
	}
	colonies := ecs.NewMap2[comp.Colony, comp.ColonyStores](w)
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: 150, CohortBased: true, SpeciesNameColony: "B_terrestris"},
		&comp.ColonyStores{EnergyStore_kJ: 100, PollenStore_g: 2})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: 160, CohortBased: true, SpeciesNameColony: "B_hortorum"},
		&comp.ColonyStores{EnergyStore_kJ: 50, PollenStore_g: 1})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: params.NotSetHigh, SpeciesNameColony: "B_terrestris"},
		&comp.ColonyStores{EnergyStore_kJ: 10})
	foodsources := ecs.NewMap2[comp.Foodsource, comp.FoodsourceState](w)
	foodsources.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{Nectar_myl: 1500000, Pollen_g: 2000})
	foodsources.NewEntity(&comp.Foodsource{}, &comp.FoodsourceState{Nectar_myl: 500000, Pollen_g: 500})
	ecs.NewMap1[comp.Badger](w).NewEntity(&comp.Badger{})

	o := DailyOutputExtended{}
	o.Initialize(w)
	header := o.Header()
	daily := (&DailyOutput{}).Header()
	if got := header[:len(daily)]; !slices.Equal(got, daily) {
		t.Fatalf("the first columns must be those of DailyOutput, got %v", got)
	}
	if got := header[len(daily) : len(daily)+len(NetLogoMetrics)]; !slices.Equal(got, NetLogoMetrics) {
		t.Fatalf("the NetLogo metrics must follow, got %v", got)
	}

	want := map[string]float64{
		"ticks": 1, "TotalQueens": 500,
		"colonies": 3, "dead_colonies": 4, "colonies_founded": 9,
		"eggs": 12, "larvae": 10, "pupae": 8, "workers": 21, "males": 4, "young_queens": 1, "hibernating_queens": 1,
		"energy_kJ": 160, "pollen_g": 3, "trips": 40, "nectar_left_l": 2, "pollen_left_kg": 2.5,
		"badgers": 1, "foodsources": 2, "ibm_colonies": 1, "bee_agents": 10, "diploid_males": 3,
		"colonies_B_terrestris": 2, "colonies_B_hortorum": 1, "colonies_B_pascuorum": 0, "psithyrus": 2,
		"foraging_period_s": 28800, "queens_removed": 11,
		"ForagerDeathsToday": 3, "TotalAdultQueensEverProduced": 7, "MeanCompetitionPointDate": 155,
	}
	check(t, header, o.Values(w), want)

	// day 2: the foragers that died are counted per day
	ticks.Ticks = 2
	counters.TotalForagersEverDiedForaging = 5
	want["ticks"], want["ForagerDeathsToday"] = 2, 2
	check(t, header, o.Values(w), want)
}

func TestDailyOutputExtendedNoCompetitionPoint(t *testing.T) {
	w := newWorld(&globals.Ticks{}, &globals.Totals{}, &globals.Counters{})
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
