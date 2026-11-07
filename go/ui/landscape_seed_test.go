//go:build ui || windows || (darwin && cgo)

package ui

import (
	"reflect"
	"testing"

	"gobees/input"
	"gobees/obs"
	"gobees/params"
	"gobees/schedule"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// exampleFarm returns the run of the example farm with seed and initial queens, set up like gobees does it: the
// model systems, then obs.Counters. queens nil keeps those of the run file.
func exampleFarm(t *testing.T, seed int64, queens []params.SpeciesQueens) (*app.App, *obs.Counters) {
	t.Helper()
	f := input.ExampleFarmFiles()
	c := f.Config
	run := *c.Run
	run.Seed = seed
	c.Run = &run
	if queens != nil {
		c.InitialQueens = queens
	}
	p, err := input.Build(c, f.Landscapes)
	if err != nil {
		t.Fatal(err)
	}
	a := schedule.Default(p, nil)
	counters := &obs.Counters{}
	a.AddSystem(counters)
	return a, counters
}

// patchLayout returns position and size of every flower patch of the example farm after setup.
func patchLayout(t *testing.T, seed int64, queens []params.SpeciesQueens) []flowerPatch {
	t.Helper()
	a, _ := exampleFarm(t, seed, queens)
	a.Initialize()
	out := newLandscape(a.World).Patches
	for i := range out { // entities differ between worlds
		out[i].Entity = ecs.Entity{}
		for j := range out[i].Layers {
			out[i].Layers[j].Entity = ecs.Entity{}
		}
	}
	return out
}

// TestLandscapeSeed: what the map shows of the landscape depends neither on the seed nor on the initial queens.
// The landscape setup draws random numbers, but they change only the order of the layers.
func TestLandscapeSeed(t *testing.T) {
	terrestris := []params.SpeciesQueens{{Species: "B_terrestris", Queens: 500}}
	a := patchLayout(t, 7, terrestris)
	three := []params.SpeciesQueens{{Species: "B_terrestris", Queens: 500}, {Species: "B_pascuorum", Queens: 300},
		{Species: "Psithyrus", Queens: 100}}
	if b := patchLayout(t, 7, three); !reflect.DeepEqual(a, b) {
		t.Error("same seed, other queens: the landscape differs")
	}
	for seed := int64(8); seed < 12; seed++ {
		if b := patchLayout(t, seed, terrestris); !reflect.DeepEqual(a, b) {
			t.Errorf("seed %v: the landscape differs from seed 7", seed)
		}
	}
}

// TestLandscapeFlowers: every layer of the example farm names its flower species (comp.Foodsource.Flower), and the
// flowering period of the layer is the one of that species in the run file.
func TestLandscapeFlowers(t *testing.T) {
	rows := map[string]params.FlowerSpecies{}
	for _, r := range input.ExampleFarmFiles().Config.Flowers {
		if _, ok := rows[r.Name]; !ok { // the model takes the first row of a name
			rows[r.Name] = r
		}
	}
	layers := 0
	for _, p := range patchLayout(t, 1, nil) {
		for _, ly := range p.Layers {
			layers++
			r, ok := rows[ly.FlowerSpecies]
			if !ok {
				t.Fatalf("patch %s: layer with flower species %q, not in the run file", p.PatchType, ly.FlowerSpecies)
			}
			if ly.StartDay != float64(r.StartDay) || ly.StopDay != float64(r.StopDay) {
				t.Errorf("%s: flowering %v to %v, run file %v to %v", ly.FlowerSpecies, ly.StartDay, ly.StopDay,
					r.StartDay, r.StopDay)
			}
		}
	}
	if layers == 0 {
		t.Fatal("no layers")
	}
}

// TestRunDays steps the example farm through spring as the window does: the UI numbers the colonies once, in the
// order it first sees them, and the day of foundation is that day.
func TestRunDays(t *testing.T) {
	g := &Game{}
	if err := g.setup(exampleFarm(t, 1, nil)); err != nil {
		t.Fatal(err)
	}
	for range 150 {
		g.step()
	}
	if g.runErr != nil {
		t.Fatal(g.runErr)
	}
	if len(g.snap.Colonies) == 0 {
		t.Fatal("no colonies after 150 days")
	}
	seen := map[int64]bool{}
	for _, c := range g.snap.Colonies {
		if seen[c.Who] {
			t.Errorf("colony number %d twice", c.Who)
		}
		seen[c.Who] = true
		h := g.colHist[c.Who]
		if h == nil || float64(h.Start) != c.Founded {
			t.Errorf("colony %d: founded %v, history from %v", c.Who, c.Founded, h)
		}
	}
	if g.snap.Tick != 150 || len(g.hist.Colonies) != 151 {
		t.Errorf("tick %d, %d days of history, want 150, 151", g.snap.Tick, len(g.hist.Colonies))
	}
}
