package obs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/reporter"
	"github.com/mlange-42/ark/ecs"
)

// newWorld returns a world with the resources DailyOutput reads.
func newWorld(ticks *globals.Ticks, output *globals.DailyOutput, totals *globals.Totals,
	counters *globals.Counters) *ecs.World {
	w := ecs.NewWorld()
	ecs.AddResource(w, ticks)
	ecs.AddResource(w, output)
	ecs.AddResource(w, totals)
	ecs.AddResource(w, counters)
	ecs.AddResource(w, &globals.DeadColonies{List: make([]globals.DeadColony, 4)})
	ecs.AddResource(w, &globals.Calendar{DailyForagingPeriod_s: 28800})
	return w
}

func TestDailyOutputCSV(t *testing.T) {
	ticks := globals.Ticks{Ticks: 1}
	output := globals.DailyOutput{TotalQueens: 500, TotalHibernatingQueens: 500, TotalMapArea_km2: 1.25}
	w := newWorld(&ticks, &output, &globals.Totals{}, &globals.Counters{})

	file := filepath.Join(t.TempDir(), "out", "daily_output.csv")
	csv := reporter.CSV{Observer: &DailyOutput{}, File: file, Sep: "\t"}
	csv.Initialize(w)
	csv.Update(w) // day 1

	ticks.Ticks = 2
	output.TotalColonies, output.TotalHibernatingQueens, output.MeanQueenWeight_mg = 2, 498, 600.5
	output.ColonyDensity_km2 = 1.6
	csv.Update(w) // day 2
	csv.Finalize(w)

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want a header and 2 rows, got\n%s", b)
	}
	columns := 1 + len((&DailyOutput{}).Header()) // t and the columns of the observer
	// the columns t and ticks and the totals of OutputDailyProc
	want := []string{
		"t\tticks\tTotalIBMColonies\tTotalQueens\tTotalMatedQueens\tTotalUnmatedQueens\tTotalHibernatingQueens\t" +
			"TotalColonies\tTotalBeeAgents\tTotalMales\tTotalActiveBees\tMeanWorkerWeight_mg\tMeanQueenWeight_mg\t" +
			"MeanAdultWeight_mg\tTotalMapArea_km2\tColonyDensity_km2",
		"0\t1\t0\t500\t0\t0\t500\t0\t0\t0\t0\t0\t0\t0\t1.25\t0",
		"1\t2\t0\t500\t0\t0\t498\t2\t0\t0\t0\t0\t600.5\t0\t1.25\t1.6",
	}
	for i, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) != columns {
			t.Errorf("line %d: %d columns, want %d", i+1, len(fields), columns)
			continue
		}
		if got := strings.Join(fields[:16], "\t"); got != want[i] {
			t.Errorf("line %d: got\n%s\nwant\n%s", i+1, got, want[i])
		}
	}
}

func TestDailyOutputNetLogoMetrics(t *testing.T) {
	counters := globals.Counters{TotalAdultQueensEverProduced: 7, TotalColoniesEverProduced: 9,
		TotalHibernatingQueensEverRemoved: 11}
	w := newWorld(&globals.Ticks{Ticks: 1}, &globals.DailyOutput{TotalQueens: 500},
		&globals.Totals{TotalForagingTripsToday: 40}, &counters)

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

	o := DailyOutput{}
	o.Initialize(w)
	header := o.Header()
	n := len(outputDailyColumns)
	if got := header[n : n+len(NetLogoMetrics)]; !slices.Equal(got, NetLogoMetrics) {
		t.Fatalf("the NetLogo metrics must follow the totals of OutputDailyProc, got %v", got)
	}

	check(t, header, o.Values(w), map[string]float64{
		"ticks": 1, "TotalQueens": 500,
		"colonies": 3, "dead_colonies": 4, "colonies_founded": 9,
		"eggs": 12, "larvae": 10, "pupae": 8, "workers": 21, "males": 4, "young_queens": 1, "hibernating_queens": 1,
		"energy_kJ": 160, "pollen_g": 3, "trips": 40, "nectar_left_l": 2, "pollen_left_kg": 2.5,
		"badgers": 1, "foodsources": 2, "ibm_colonies": 1, "bee_agents": 10, "diploid_males": 3,
		"colonies_B_terrestris": 2, "colonies_B_hortorum": 1, "colonies_B_pascuorum": 0, "psithyrus": 2,
		"foraging_period_s": 28800, "queens_removed": 11,
		"TotalAdultQueensEverProduced": 7, "MeanCompetitionPointDate": 155,
	})
}

func TestDailyOutputNoCompetitionPoint(t *testing.T) {
	w := newWorld(&globals.Ticks{}, &globals.DailyOutput{}, &globals.Totals{}, &globals.Counters{})
	ecs.NewMap2[comp.Colony, comp.ColonyStores](w).NewEntity(
		&comp.Colony{CompetitionPointDate: params.NotSetHigh}, &comp.ColonyStores{})

	o := DailyOutput{}
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
