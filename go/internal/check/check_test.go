package check

import (
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// world returns a world on Day 200 with a colony (its queen, brood of every stage, a worker and a male) and a
// hibernating queen without colony.
func world() (*ecs.World, *globals.Calendar, ecs.Entity) {
	w := ecs.NewWorld()
	cal := &globals.Calendar{Day: 200}
	ecs.AddResource(w, cal)
	col := ecs.NewMap1[comp.Colony](w).NewEntity(&comp.Colony{})
	bees := ecs.NewMap1[comp.Bee](w)
	for _, b := range []comp.Bee{
		{Stage: stage.Adult, Caste: caste.Queen, Colony: col, Activity: activity.Egglaying},
		{Stage: stage.Egg, Caste: caste.Undefined, Colony: col},
		{Stage: stage.Larva, Caste: caste.Queen, Colony: col},
		{Stage: stage.Pupa, Caste: caste.Male, Colony: col},
		{Stage: stage.Adult, Caste: caste.Worker, Colony: col},
		{Stage: stage.Adult, Caste: caste.Male, Colony: col},
		{Stage: stage.Adult, Caste: caste.Queen, Activity: activity.Hibernate},
	} {
		b.Number = 1
		bees.NewEntity(&b)
	}
	return w, cal, col
}

func TestDayValid(t *testing.T) {
	w, _, _ := world()
	if err := Day(w, true); err != nil {
		t.Error(err)
	}
}

func TestDayMembers(t *testing.T) {
	// brood, workers and males without a living colony
	for _, b := range []comp.Bee{
		{Stage: stage.Adult, Caste: caste.Worker},
		{Stage: stage.Adult, Caste: caste.Male},
		{Stage: stage.Egg, Caste: caste.Undefined},
	} {
		w, _, _ := world()
		ecs.NewMap1[comp.Bee](w).NewEntity(&b)
		if err := Day(w, false); err == nil {
			t.Errorf("stage %v, caste %v without colony: no error", b.Stage, b.Caste)
		}
	}
	// the colony was removed while bees still belong to it
	w, _, col := world()
	w.RemoveEntity(col)
	if err := Day(w, false); err == nil {
		t.Error("bees of a removed colony: no error")
	}
}

func TestDayColoniesOnDay364(t *testing.T) {
	// a living colony on Day 364, only with seasonEnd
	w, cal, _ := world()
	cal.Day = 364
	if err := Day(w, true); err == nil {
		t.Error("living colony on Day 364: no error")
	}
	if err := Day(w, false); err != nil {
		t.Errorf("without seasonEnd: %v", err)
	}
	cal.Day = 363
	if err := Day(w, true); err != nil {
		t.Errorf("Day 363: %v", err)
	}
}
