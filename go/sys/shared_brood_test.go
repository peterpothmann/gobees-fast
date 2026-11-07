package sys

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// broodWorld extends colonyWorld with the brood systems. Brood is laid with eggLayer or created by hand
// (addBrood), without the activities of the bees.
type broodWorld struct {
	*colonyWorld
	need *larvaeNeeds
	feed *larvaeFeeding
	qpd  *queenProductionDate
	dev  *development
	morb *broodMortality
	eggs eggLayer
	sp   *params.Species
}

func newBroodWorld(t *testing.T, seed uint64) *broodWorld {
	t.Helper()
	bw := &broodWorld{colonyWorld: newColonyWorld(t, seed),
		need: &larvaeNeeds{}, feed: &larvaeFeeding{}, qpd: &queenProductionDate{},
		dev: &development{}, morb: &broodMortality{}}
	bw.need.Initialize(bw.w)
	bw.feed.Initialize(bw.w)
	bw.qpd.Initialize(bw.w)
	bw.dev.Initialize(bw.w)
	bw.morb.Initialize(bw.w)
	bw.eggs = newEggLayer(bw.w, bw.dev.rng)
	bw.sp = &bw.input.Species[bw.species] // read only: addParasiticSpecies may move the slice
	return bw
}

// addParasiticSpecies adds a copy of the species sp named Psithyrus with Parasitic set and returns its index.
func addParasiticSpecies(w *ecs.World, sp *params.Species) int {
	psith := *sp
	psith.Name, psith.Parasitic = "Psithyrus", true
	in := ecs.GetResource[params.Input](w)
	in.Species = append(in.Species, psith)
	return len(in.Species) - 1
}

func (bw *broodWorld) beeFilter() *ecs.Filter1[comp.Bee] {
	return ecs.NewFilter1[comp.Bee](bw.w)
}

// eggsOf returns the egg cohorts of a colony.
func (bw *broodWorld) eggsOf(col ecs.Entity) []ecs.Entity {
	result := []ecs.Entity{}
	query := bw.beeFilter().Query()
	for query.Next() {
		if b := query.Get(); b.Stage == stage.Egg && b.Colony == col {
			result = append(result, query.Entity())
		}
	}
	return result
}

// bee returns the components of a bee.
func (bw *broodWorld) bee(e ecs.Entity) (*comp.Bee, *comp.Genetics, *comp.Thresholds, *comp.Forager, *comp.Development) {
	_, b, g, th, f, d := bw.beeMap.Get(e)
	return b, g, th, f, d
}

// addBrood adds brood of a colony with the given caste, broodAge, weight and incubation.
// weight [mg], incubation [kJ].
func (bw *broodWorld) addBrood(col ecs.Entity, st stage.Stage, c caste.Caste, number, broodAge int, weight, incubation float64) ecs.Entity {
	e := bw.addBee(col, st, c, number, 0)
	b, _, _, _, d := bw.bee(e)
	b.BroodAge, b.Weight, d.IncubationReceived = broodAge, weight, incubation
	return e
}

// evening runs the evening systems in the order of the group Evening (schedule.go).
func (bw *broodWorld) evening() {
	bw.feed.Update(bw.w)
	bw.qpd.Update(bw.w)
	bw.dev.Update(bw.w)
	bw.morb.Update(bw.w)
}

func TestBeesByColonyReusesGroups(t *testing.T) {
	// the second call empties the groups of the first and drops the colony without larvae
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
