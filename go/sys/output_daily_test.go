package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// outputWorld is a colonyWorld with the world size and the OutputDaily system (not yet initialised).
func outputWorld(t *testing.T) *colonyWorld {
	t.Helper()
	cw := newColonyWorld(t, 1)
	world := params.Default().World // 0-300 x 0-210
	ecs.AddResource(cw.w, &world)
	return cw
}

// addFoodsourceAt adds a foodsource with the given radius at (x, y).
func addFoodsourceAt(w *ecs.World, x, y, radius_m float64) ecs.Entity {
	return ecs.NewMap2[comp.Position, comp.Foodsource](w).NewEntity(&comp.Position{X: x, Y: y}, &comp.Foodsource{Radius_m: radius_m})
}

func TestOutputDailyTotals(t *testing.T) {
	// [OUT-01]
	cw := outputWorld(t)
	col := cw.addColony(nil)
	col2 := cw.addColony(nil)
	weight := func(e ecs.Entity, mg float64) {
		_, _, b, _, _, _, _ := cw.beeMap.Get(e)
		b.Weight_mg = mg
	}
	weight(cw.addHibernatingQueen(3), 600)                       // mated, hibernating cohort
	weight(cw.addBee(col, stage.Adult, caste.Queen, 1, 30), 800) // mother queen, mated
	weight(cw.addYoungQueen(col, 1), 500)                        // unmated
	cw.addBee(col, stage.Pupa, caste.Queen, 2, 0)                // queen brood, unmated
	weight(cw.addBee(col, stage.Adult, caste.Male, 4, 1), 100)
	cw.addBee(col, stage.Larva, caste.Male, 2, 0)
	weight(cw.addBee(col2, stage.Adult, caste.Worker, 1, 3), 100)
	weight(cw.addBee(col, stage.Adult, caste.Worker, 5, 3), 200)
	cw.addBee(col, stage.Egg, caste.Undefined, 5, 0)
	active := ecs.GetResource[globals.ActiveBees](cw.w)
	active.List = make([]ecs.Entity, 3)

	s := &OutputDaily{}
	s.Initialize(cw.w) // [INIT-12] runs once
	got := *ecs.GetResource[globals.DailyOutput](cw.w)
	// the mean weights are means over adult agents, not weighted by number
	want := globals.DailyOutput{
		TotalColonies: 2, TotalQueens: 7, TotalMatedQueens: 4, TotalUnmatedQueens: 3, TotalHibernatingQueens: 3,
		TotalBeeAgents: 9, TotalMales: 6, TotalActiveBees: 3,
		MeanWorkerWeight_mg: 150, MeanQueenWeight_mg: (600 + 800 + 500) / 3.0,
		MeanAdultWeight_mg: (600 + 800 + 500 + 100 + 100 + 200) / 6.0,
	}
	if got != want {
		t.Errorf("totals:\n got %+v\nwant %+v", got, want)
	}

	// without adults the mean weights are 0
	adults := askBees(ecs.NewFilter1[comp.Bee](cw.w), cw.sys.rng, func(b *comp.Bee) bool { return !b.Brood })
	for _, e := range adults {
		cw.w.RemoveEntity(e)
	}
	s.Update(cw.w)
	got = *ecs.GetResource[globals.DailyOutput](cw.w)
	if got.MeanWorkerWeight_mg != 0 || got.MeanQueenWeight_mg != 0 || got.MeanAdultWeight_mg != 0 {
		t.Errorf("mean weights without adults: %v, %v, %v; want 0", got.MeanWorkerWeight_mg, got.MeanQueenWeight_mg, got.MeanAdultWeight_mg)
	}
}

func TestTotalMapAreaREP(t *testing.T) {
	// [OUT-02], [AREA]. Worked example (positions chosen by hand), Scaling 0.1 patches per m:
	// A (100, 100), radius 100 m = 10 patches; B (150, 120), radius 50 m = 5 patches.
	// top: B 120 + 5 = 125, bottom: A 100 - 10 = 90, right: B 150 + 5 = 155, left: A 100 - 10 = 90.
	// (125 - 90) * (155 - 90) = 2275 patches^2 = 227500 m2 = 0.2275 km2.
	cw := outputWorld(t)
	addFoodsourceAt(cw.w, 100, 100, 100)
	addFoodsourceAt(cw.w, 150, 120, 50)
	cw.addColony(nil)
	cw.addColony(nil)
	s := &OutputDaily{}
	s.Initialize(cw.w)
	out := ecs.GetResource[globals.DailyOutput](cw.w)
	if !almostEqual(out.TotalMapArea_km2, 0.2275) || !almostEqual(out.ColonyDensity_km2, 2/0.2275) {
		t.Errorf("area %v km2, density %v; want 0.2275 and %v", out.TotalMapArea_km2, out.ColonyDensity_km2, 2/0.2275)
	}
	// computed only once: a new foodsource does not change it
	addFoodsourceAt(cw.w, 10, 10, 10)
	s.Update(cw.w)
	if !almostEqual(out.TotalMapArea_km2, 0.2275) {
		t.Errorf("area changed to %v", out.TotalMapArea_km2)
	}

	// limited to the world: C (295, 5), radius 100 m: right 305 -> 300, bottom -5 -> 0.
	// (125 - 0) * (300 - 90) = 26250 patches^2 = 2.625 km2
	cw = outputWorld(t)
	addFoodsourceAt(cw.w, 100, 100, 100)
	addFoodsourceAt(cw.w, 150, 120, 50)
	addFoodsourceAt(cw.w, 295, 5, 100)
	s = &OutputDaily{}
	s.Initialize(cw.w)
	if got := s.TotalMapAreaREP(); !almostEqual(got, 2.625) {
		t.Errorf("area limited to the world: %v km2, want 2.625", got)
	}

	// no foodsources: no area, no density
	cw = outputWorld(t)
	s = &OutputDaily{}
	s.Initialize(cw.w)
	if out := ecs.GetResource[globals.DailyOutput](cw.w); out.TotalMapArea_km2 != 0 || out.ColonyDensity_km2 != 0 {
		t.Errorf("without foodsources: area %v, density %v; want 0", out.TotalMapArea_km2, out.ColonyDensity_km2)
	}
}
