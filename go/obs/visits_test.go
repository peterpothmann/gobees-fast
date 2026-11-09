package obs

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/event"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestVisits(t *testing.T) {
	w := newWorld(t)
	ecs.GetResource[globals.Landscape](w).File = "farm.yaml"
	ecs.AddResource(w, &params.Input{Flowers: &params.FlowerSpeciesTable{Names: []string{"Bramble", "Dandelion"}}})
	patches := ecs.NewMap1[comp.FlowerPatch](w)
	// the flower patch 7 with a bramble and a dandelion layer, the patch 3 with a dandelion layer
	patch7 := patches.NewEntity(&comp.FlowerPatch{ID: 7, Layers: []comp.Layer{{Flower: 0}, {Flower: 1}}})
	patch3 := patches.NewEntity(&comp.FlowerPatch{ID: 3, Layers: []comp.Layer{{Flower: 1}}})
	bramble, dandelion := comp.Source{Patch: patch7, Layer: 0}, comp.Source{Patch: patch7, Layer: 1}
	other := comp.Source{Patch: patch3, Layer: 0}
	v := Visits{}
	v.Initialize(w)

	bees := ecs.NewMap2[comp.Bee, comp.Forager](w)
	collected := w.Event(event.Collected).For(ecs.C[comp.Bee](), ecs.C[comp.Forager]())
	visit := func(src comp.Source, a activity.Activity, number int) {
		collected.Emit(bees.NewEntity(&comp.Bee{Activity: a, Number: number}, &comp.Forager{CurrentFoodsource: src}))
	}
	// day 1: a cohort of 3 takes nectar and one of 2 takes pollen at the dandelion of patch 7. Patch 3 is visited and
	// then removed (as by a crop rotation), its visit stays.
	visit(dandelion, activity.BringingNectar, 3)
	visit(dandelion, activity.BringingPollen, 2)
	visit(other, activity.BringingNectar, 1)
	visit(dandelion, activity.BringingNectar, 1)
	w.RemoveEntity(patch3)
	v.Update(w)
	want := []Visit{
		{Landscape: "farm.yaml", ID: 3, Flower: 1, Nectar: 1},
		{Landscape: "farm.yaml", ID: 7, Flower: 1, Nectar: 4, Pollen: 2},
	}
	if !slices.Equal(v.Day, want) {
		t.Errorf("day 1: %+v, want %+v (sorted by landscape, id, flower species)", v.Day, want)
	}
	if v.FlowerName(v.Day[0]) != "Dandelion" || v.FlowerName(v.Day[1]) != "Dandelion" {
		t.Errorf("flower species %q, %q, want Dandelion, Dandelion", v.FlowerName(v.Day[0]), v.FlowerName(v.Day[1]))
	}

	// day 2: counted per day, the bramble layer of the same flower patch comes before the dandelion
	visit(dandelion, activity.BringingPollen, 5)
	visit(bramble, activity.BringingNectar, 2)
	v.Update(w)
	want = []Visit{
		{Landscape: "farm.yaml", ID: 7, Flower: 0, Nectar: 2},
		{Landscape: "farm.yaml", ID: 7, Flower: 1, Pollen: 5},
	}
	if !slices.Equal(v.Day, want) {
		t.Errorf("day 2: %+v, want %+v", v.Day, want)
	}

	// day 3: no visit, no item
	v.Update(w)
	if len(v.Day) != 0 {
		t.Errorf("day 3: %+v, want no visits", v.Day)
	}
}
