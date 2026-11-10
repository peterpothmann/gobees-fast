package sys

import (
	"fmt"

	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// flowerRefill fills every foodsource in flower up to its maximum and empties all others.
//
// Part of the morning routine. Initialize runs it also at ticks = 0, i.e. Day = 0.
//
// NetLogo: UpdateFoodsourcesProc.
type flowerRefill struct {
	calendar *globals.Calendar
	today    globals.Flowering

	filter *ecs.Filter2[comp.Foodsource, comp.FoodsourceState]
	fsMap  *ecs.Map2[comp.Foodsource, comp.FoodsourceState]
}

func (s *flowerRefill) Initialize(w *ecs.World) {
	s.calendar = ecs.GetResource[globals.Calendar](w)
	ecs.AddResource(w, &s.today)
	s.filter = ecs.NewFilter2[comp.Foodsource, comp.FoodsourceState](w)
	s.fsMap = ecs.NewMap2[comp.Foodsource, comp.FoodsourceState](w)
	s.Update(w) // run at ticks = 0
}

func (s *flowerRefill) Update(w *ecs.World) {
	day := s.calendar.Day
	s.today.Changed = false

	// ORDER: irrelevant, each foodsource only changes itself. Not shuffled (NetLogo: random order).
	foodsources := []ecs.Entity{}
	query := s.filter.Query()
	for query.Next() {
		foodsources = append(foodsources, query.Entity())
	}

	for _, e := range foodsources {
		fs, state := s.fsMap.Get(e)
		if fs.StartDay > fs.StopDay {
			panic(fmt.Sprintf("Foodsource: startDay > StopDay! (flowerRefill): startDay %v, stopDay %v", fs.StartDay, fs.StopDay))
		}
		if day == fs.StartDay || day == fs.StopDay {
			s.today.Changed = true
		}
		if day >= fs.StartDay && day < fs.StopDay {
			state.Nectar = fs.NectarMax
			state.Pollen = fs.PollenMax
		} else {
			state.Nectar = 0
			state.Pollen = 0
		}
	}
}

func (s *flowerRefill) Finalize(w *ecs.World) {}
