package obs

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// emit emits the custom event evt on the bee e.
func emit(w *ecs.World, evt ecs.EventType, e ecs.Entity) {
	w.Event(evt).For(ecs.C[comp.Bee]()).Emit(e)
}

func TestDailyOutputExtended(t *testing.T) {
	w := newWorld(t)
	calendar := ecs.GetResource[globals.Calendar](w)
	calendar.ForagingPeriod = 28800

	// species indices: the bees without Species belong to species 0, B_terrestris
	const terrestris, hortorum, pascuorum, psithyrus = 0, 1, 2, 3
	ecs.AddResource(w, &params.Input{Species: []params.Species{{SpeciesFile: params.SpeciesFile{Name: "B_terrestris"}}, {SpeciesFile: params.SpeciesFile{Name: "B_hortorum"}},
		{SpeciesFile: params.SpeciesFile{Name: "B_pascuorum"}}, {SpeciesFile: params.SpeciesFile{Name: "Psithyrus", Parasitic: true}}}})

	// activity.Hibernate is the zero value: all other bees rest.
	bees := []ecs.Entity{}
	for _, b := range []comp.Bee{
		{Number: 12, Stage: stage.Egg, Activity: activity.Resting},
		{Number: 10, Stage: stage.Larva, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 5, Stage: stage.Pupa, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 3, Stage: stage.Pupa, Caste: caste.Male, Activity: activity.Resting},
		{Number: 20, Stage: stage.Adult, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Worker, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Queen, Activity: activity.Resting},
		{Number: 1, Stage: stage.Adult, Caste: caste.Queen, Mated: true, Activity: activity.Hibernate},
		{Number: 2, Stage: stage.Adult, Caste: caste.Queen, Mated: true, Species: psithyrus, Activity: activity.Resting},
		{Number: 4, Stage: stage.Adult, Caste: caste.Male, Activity: activity.Resting},
	} {
		bees = append(bees, addBee(w, b))
	}
	// diploid male pupae (2 alleles), haploid adult males (1 allele)
	genetics := ecs.NewMap1[comp.Genetics](w)
	genetics.Get(bees[3]).Alleles = []float64{10, 20}
	genetics.Get(bees[9]).Alleles = []float64{10}
	colonies := ecs.NewMap2[comp.Colony, comp.ColonyStores](w)
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: day.At(150), Species: terrestris},
		&comp.ColonyStores{Energy: 100, Pollen: 2})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: day.At(160), Species: hortorum},
		&comp.ColonyStores{Energy: 50, Pollen: 1})
	colonies.NewEntity(&comp.Colony{CompetitionPointDate: day.Date{}, Species: terrestris},
		&comp.ColonyStores{Energy: 10})
	// one flower patch with two layers: two foodsources
	ecs.NewMap1[comp.FlowerPatch](w).NewEntity(&comp.FlowerPatch{Layers: []comp.Layer{
		{Nectar: 1500000, Pollen: 2000}, {Nectar: 500000, Pollen: 500}}})
	ecs.NewMap1[comp.Badger](w).NewEntity(&comp.Badger{})

	// Setup ends here: the counters start
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

	// day 1: 2 trips of the squadron of 20 workers, the single worker dies on her trip, the queen pupa emerges, a
	// colony is founded. MaxHibernatingQueens is 0, so the 1 hibernating queen of the setup is planned for removal
	emit(w, event.ForagingTrip, bees[4])
	emit(w, event.ForagingTrip, bees[4])
	emit(w, event.ForagingTrip, bees[5])
	emit(w, event.DiedForaging, bees[5])
	w.RemoveEntity(bees[5])
	emit(w, event.QueenEmerged, bees[6])
	founded := colonies.NewEntity(&comp.Colony{CompetitionPointDate: day.Date{}, Species: pascuorum},
		&comp.ColonyStores{})
	o.Update(w)
	want := map[string]float64{
		"ticks": 1, "TotalColonies": 4,
		"colonies": 4, "dead_colonies": 0, "colonies_founded": 1,
		"eggs": 12, "larvae": 10, "pupae": 8, "workers": 20, "males": 4, "young_queens": 1, "hibernating_queens": 1,
		"energy_kJ": 160, "pollen_g": 3, "trips": 41, "nectar_left_l": 2, "pollen_left_kg": 2.5,
		"badgers": 1, "foodsources": 2, "bee_agents": 9, "diploid_males": 3,
		"colonies_B_terrestris": 2, "colonies_B_hortorum": 1, "colonies_B_pascuorum": 1, "colonies_B_pratorum": 0,
		"psithyrus": 2, "foraging_period_s": 28800, "queens_removed": 1,
		"ForagerDeathsToday": 1, "TotalAdultQueensEverProduced": 1, "MeanCompetitionPointDate": 155,
	}
	check(t, header, o.Values(w), want)

	// day 2: the trips and the foragers that died are counted per day, the founded colony dies
	ecs.GetResource[globals.Ticks](w).Ticks = 2
	calendar.Day = 2
	emit(w, event.ForagingTrip, bees[9])
	w.RemoveEntity(founded)
	o.Update(w)
	want["ticks"], want["trips"], want["ForagerDeathsToday"] = 2, 4, 0
	want["TotalColonies"], want["colonies"], want["dead_colonies"], want["colonies_B_pascuorum"] = 3, 3, 1, 0
	check(t, header, o.Values(w), want)
}

func TestDailyOutputExtendedNoCompetitionPoint(t *testing.T) {
	w := newWorld(t)
	ecs.AddResource(w, &params.Input{Species: []params.Species{{SpeciesFile: params.SpeciesFile{Name: "B_terrestris"}}}})
	ecs.NewMap2[comp.Colony, comp.ColonyStores](w).NewEntity(
		&comp.Colony{CompetitionPointDate: day.Date{}, Species: 0}, &comp.ColonyStores{})

	o := DailyOutputExtended{}
	o.Initialize(w)
	o.Update(w)
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
