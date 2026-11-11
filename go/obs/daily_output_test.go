package obs

import (
	"math"
	"os"
	"path/filepath"
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

// newWorld returns a world with the resources that the observers read: ticks 1, day 1, an empty landscape, an empty
// parameter file.
func newWorld(t *testing.T) *ecs.World {
	t.Helper()
	w := ecs.NewWorld()
	ecs.AddResource(w, &globals.Ticks{Ticks: 1})
	ecs.AddResource(w, &globals.Calendar{Day: 1})
	ecs.AddResource(w, &globals.Landscape{})
	ecs.AddResource(w, &globals.ActiveBees{})
	ecs.AddResource(w, &params.Parameters{})
	ecs.AddResource(w, &params.Output{})
	return w
}

// addBee adds a bee with the given values and empty genetics.
func addBee(w *ecs.World, b comp.Bee) ecs.Entity {
	return ecs.NewMap2[comp.Bee, comp.Genetics](w).NewEntity(&b, &comp.Genetics{})
}

// addFoodsourceAt adds a flower patch with the given radius at (x, y), all in [m].
func addFoodsourceAt(w *ecs.World, x, y, radius float64) ecs.Entity {
	return ecs.NewMap2[comp.Position, comp.FlowerPatch](w).NewEntity(&comp.Position{X: x, Y: y}, &comp.FlowerPatch{Radius: radius})
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Abs(b))
}

func TestDailyOutputTotals(t *testing.T) {
	w := newWorld(t)
	colonies := ecs.NewMap1[comp.Colony](w)
	colonies.NewEntity(&comp.Colony{})
	colonies.NewEntity(&comp.Colony{})
	rest := activity.Resting // the zero value of Activity is Hibernate
	for _, b := range []comp.Bee{
		{Stage: stage.Adult, Caste: caste.Queen, Number: 3, Mated: true, Activity: activity.Hibernate, Weight: 600},
		{Stage: stage.Adult, Caste: caste.Queen, Number: 1, Mated: true, Activity: rest, Weight: 800}, // mother queen
		{Stage: stage.Adult, Caste: caste.Queen, Number: 1, Activity: rest, Weight: 500},              // young queen, unmated
		{Stage: stage.Pupa, Caste: caste.Queen, Number: 2, Activity: rest},                            // queen brood, unmated
		{Stage: stage.Adult, Caste: caste.Male, Number: 4, Activity: rest, Weight: 100},
		{Stage: stage.Larva, Caste: caste.Male, Number: 2, Activity: rest},
		{Stage: stage.Adult, Caste: caste.Worker, Number: 1, Activity: rest, Weight: 100},
		{Stage: stage.Adult, Caste: caste.Worker, Number: 5, Activity: rest, Weight: 200},
		{Stage: stage.Egg, Caste: caste.Undefined, Number: 5, Activity: rest},
	} {
		addBee(w, b)
	}
	ecs.GetResource[globals.ActiveBees](w).Set(make([]ecs.Entity, 3), func(ecs.Entity) float64 { return 0 })

	o := &DailyOutput{}
	o.Initialize(w)
	o.Values(w)
	// the mean weights are means over adult agents, not weighted by number
	want := DailyTotals{
		TotalColonies: 2, TotalQueens: 7, TotalMatedQueens: 4, TotalUnmatedQueens: 3, TotalHibernatingQueens: 3,
		TotalBeeAgents: 9, TotalMales: 6, TotalActiveBees: 3,
		MeanWorkerWeight_mg: 150, MeanQueenWeight_mg: (600 + 800 + 500) / 3.0,
		MeanAdultWeight_mg: (600 + 800 + 500 + 100 + 100 + 200) / 6.0,
	}
	if o.Totals != want {
		t.Errorf("totals:\n got %+v\nwant %+v", o.Totals, want)
	}

	// without adults the mean weights are 0
	adults := []ecs.Entity{}
	query := ecs.NewFilter1[comp.Bee](w).Query()
	for query.Next() {
		if query.Get().Stage == stage.Adult {
			adults = append(adults, query.Entity())
		}
	}
	for _, e := range adults {
		w.RemoveEntity(e)
	}
	o.Values(w)
	if d := o.Totals; d.MeanWorkerWeight_mg != 0 || d.MeanQueenWeight_mg != 0 || d.MeanAdultWeight_mg != 0 {
		t.Errorf("mean weights without adults: %v, %v, %v; want 0", d.MeanWorkerWeight_mg, d.MeanQueenWeight_mg, d.MeanAdultWeight_mg)
	}
}

func TestTotalMapArea(t *testing.T) {
	// Worked example (positions chosen by hand, in m):
	// A (1000, 1000), radius 100; B (1500, 1200), radius 50.
	// top: B 1200 + 50 = 1250, bottom: A 1000 - 100 = 900, right: B 1500 + 50 = 1550, left: A 1000 - 100 = 900.
	// (1250 - 900) * (1550 - 900) = 227500 m2 = 0.2275 km2.
	w := newWorld(t)
	addFoodsourceAt(w, 1000, 1000, 100)
	addFoodsourceAt(w, 1500, 1200, 50)
	colonies := ecs.NewMap1[comp.Colony](w)
	colonies.NewEntity(&comp.Colony{})
	colonies.NewEntity(&comp.Colony{})
	o := &DailyOutput{}
	o.Initialize(w) // the area of the landscape after Setup
	o.Values(w)
	if d := o.Totals; !almostEqual(d.TotalMapArea_km2, 0.2275) || !almostEqual(d.ColonyDensity_km2, 2/0.2275) {
		t.Errorf("area %v km2, density %v; want 0.2275 and %v", d.TotalMapArea_km2, d.ColonyDensity_km2, 2/0.2275)
	}
	// computed only once: a new foodsource does not change it
	addFoodsourceAt(w, 100, 100, 100)
	o.Values(w)
	if !almostEqual(o.Totals.TotalMapArea_km2, 0.2275) {
		t.Errorf("area changed to %v", o.Totals.TotalMapArea_km2)
	}

	// not limited (NetLogo: limited to its world, 0-3000 m x 0-2100 m at 10 m per patch): C (2950, 50), radius
	// 1000: right 3950, bottom -950. (1250 + 950) * (3950 - 900) = 6710000 m2 = 6.71 km2
	w = newWorld(t)
	addFoodsourceAt(w, 1000, 1000, 100)
	addFoodsourceAt(w, 1500, 1200, 50)
	addFoodsourceAt(w, 2950, 50, 1000)
	o = &DailyOutput{}
	o.Initialize(w)
	if got := o.totalMapArea(); !almostEqual(got, 6.71) {
		t.Errorf("area beyond the NetLogo world: %v km2, want 6.71", got)
	}

	// no foodsources: no area, no density
	w = newWorld(t)
	o = &DailyOutput{}
	o.Initialize(w)
	o.Values(w)
	if d := o.Totals; d.TotalMapArea_km2 != 0 || d.ColonyDensity_km2 != 0 {
		t.Errorf("without foodsources: area %v, density %v; want 0", d.TotalMapArea_km2, d.ColonyDensity_km2)
	}
}

func TestDailyOutputCSV(t *testing.T) {
	w := newWorld(t)
	addFoodsourceAt(w, 100, 100, 100) // 200 m x 200 m = 0.04 km2
	queens := addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 500, Mated: true, Activity: activity.Hibernate})

	file := filepath.Join(t.TempDir(), "out", "daily_output.csv")
	csv := reporter.CSV{Observer: &DailyOutput{}, File: file, Sep: "\t"}
	csv.Initialize(w)
	csv.Update(w) // day 1

	ecs.GetResource[globals.Ticks](w).Ticks = 2
	ecs.NewMap1[comp.Bee](w).Get(queens).Number = 498
	colonies := ecs.NewMap1[comp.Colony](w)
	colonies.NewEntity(&comp.Colony{})
	colonies.NewEntity(&comp.Colony{})
	addBee(w, comp.Bee{Stage: stage.Adult, Caste: caste.Queen, Number: 1, Mated: true, Weight: 600, Activity: activity.Resting})
	csv.Update(w) // day 2
	csv.Finalize(w)

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := "t\tticks\tTotalQueens\tTotalMatedQueens\tTotalUnmatedQueens\tTotalHibernatingQueens\t" +
		"TotalColonies\tTotalBeeAgents\tTotalMales\tTotalActiveBees\tMeanWorkerWeight_mg\tMeanQueenWeight_mg\t" +
		"MeanAdultWeight_mg\tTotalMapArea_km2\tColonyDensity_km2\n" +
		"0\t1\t500\t500\t0\t500\t0\t1\t0\t0\t0\t0\t0\t0.04\t0\n" +
		"1\t2\t499\t499\t0\t498\t2\t2\t0\t0\t0\t300\t300\t0.04\t50\n"
	if got := string(b); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
