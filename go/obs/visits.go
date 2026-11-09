package obs

import (
	"cmp"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// VisitColumns are the columns of the visits output after those that the run adds in front (run, seed, the varied
// keys, t).
var VisitColumns = []string{"ticks", "landscape", "id", "flowerSpecies", "nectarVisits", "pollenVisits"}

// Visit holds the visits of the foragers to one foodsource (a layer of a flower patch) on one day.
type Visit struct {
	Landscape string // the landscape file of the foodsource, as named in the run file (globals.Landscape.File)
	ID        int    // the id of its flower patch in the landscape file (comp.FlowerPatch.ID)
	Flower    int    // the index of its flower species in params.Input.Flowers (FlowerName gives the name)
	Nectar    int    // visits that took nectar
	Pollen    int    // visits that took pollen
}

// Visits counts the visits of the foragers per foodsource and day: event.Collected, one visit per bee of the cohort.
// A trip to an empty foodsource is no visit. NetLogo: cumulNectarVisits and cumulPollenVisits, written by
// BBCT_WriteNPvisitsProc. Unlike NetLogo, the visits are counted per day, not summed since the start.
//
// The foodsource is read at the visit, so a foodsource removed later in the day (crop rotation, empty foodsources)
// keeps its visits. Initialize must run after Setup. Update must run once at the end of every day, after all model
// systems. Day holds the visits of the day that ended last.
type Visits struct {
	// Day holds the visits of the day that ended last, one item per foodsource with at least one visit, sorted by
	// landscape, ID and Flower.
	Day []Visit

	ticks     *globals.Ticks
	landscape *globals.Landscape
	input     *params.Input
	patches   *ecs.Map1[comp.FlowerPatch]
	today     map[comp.Source]int // index in visits of a foodsource visited today
	visits    []Visit             // the visits of today
}

func (v *Visits) Initialize(w *ecs.World) {
	v.ticks = ecs.GetResource[globals.Ticks](w)
	v.landscape = ecs.GetResource[globals.Landscape](w)
	v.input = ecs.GetResource[params.Input](w)
	v.patches = ecs.NewMap1[comp.FlowerPatch](w)
	v.today = map[comp.Source]int{}

	ecs.Observe2[comp.Bee, comp.Forager](event.Collected).
		Do(func(e ecs.Entity, b *comp.Bee, f *comp.Forager) { v.add(b, f) }).
		Register(w)
}

// add counts the visit of the forager b, f.
func (v *Visits) add(b *comp.Bee, f *comp.Forager) {
	src := f.CurrentFoodsource
	i, ok := v.today[src]
	if !ok {
		fp := v.patches.Get(src.Patch)
		i = len(v.visits)
		v.today[src] = i
		v.visits = append(v.visits, Visit{Landscape: v.landscape.File, ID: fp.ID, Flower: fp.Layers[src.Layer].Flower})
	}
	switch b.Activity {
	case activity.BringingNectar:
		v.visits[i].Nectar += b.Number
	case activity.BringingPollen:
		v.visits[i].Pollen += b.Number
	default:
		panic("obs: event.Collected on a bee that brings neither nectar nor pollen")
	}
}

func (v *Visits) Update(w *ecs.World) {
	// sorted, so that the rows do not depend on the order of the visits
	slices.SortStableFunc(v.visits, func(a, b Visit) int {
		return cmp.Or(cmp.Compare(a.Landscape, b.Landscape), cmp.Compare(a.ID, b.ID), cmp.Compare(a.Flower, b.Flower))
	})
	v.Day = append(v.Day[:0], v.visits...)
	v.visits = v.visits[:0]
	clear(v.today)
}

func (v *Visits) Finalize(w *ecs.World) {}

// FlowerName returns the name of the flower species of the visit. A forager visits only a layer, and every layer has
// a flower species (comp.FlowerPatch.Layers).
func (v *Visits) FlowerName(visit Visit) string {
	return v.input.Flowers.Names[visit.Flower]
}

// Ticks returns the ticks of the day that ended last.
func (v *Visits) Ticks() int64 {
	return int64(v.ticks.Ticks)
}
