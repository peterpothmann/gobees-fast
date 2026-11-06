package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/globals"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// flowerRefill fills every foodsource in flower up to its maximum and empties all others [UFS].
//
// Part of the morning routine [MORN-05]. Initialize runs it also at ticks = 0, i.e. Day = 0 [INIT-10].
//
// NetLogo: UpdateFoodsourcesProc.
type flowerRefill struct {
	rng      *rand.Rand
	calendar *globals.Calendar
	today    globals.Flowering

	filter *ecs.Filter2[comp.Foodsource, comp.FoodsourceState]
	fsMap  *ecs.Map2[comp.Foodsource, comp.FoodsourceState]
}

func (s *flowerRefill) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.calendar = ecs.GetResource[globals.Calendar](w)
	ecs.AddResource(w, &s.today)
	s.filter = ecs.NewFilter2[comp.Foodsource, comp.FoodsourceState](w)
	s.fsMap = ecs.NewMap2[comp.Foodsource, comp.FoodsourceState](w)
	s.Update(w) // run at ticks = 0 [INIT-10]
}

func (s *flowerRefill) Update(w *ecs.World) {
	day := s.calendar.Day
	// [UFS-01] pollenAvailableTotalToday_kg, nectarAvailableTotalToday_l: output only, removed in step F4
	s.today.Changed = false

	// [UFS-02] ORDER: irrelevant, each foodsource only changes itself.
	foodsources := []ecs.Entity{}
	query := s.filter.Query()
	for query.Next() {
		foodsources = append(foodsources, query.Entity())
	}
	util.Shuffle(s.rng, foodsources)

	for _, e := range foodsources {
		fs, state := s.fsMap.Get(e)
		if fs.StartDay > fs.StopDay {
			panic(fmt.Sprintf("Foodsource: startDay > StopDay! (UpdateFoodsourcesProc): startDay %v, stopDay %v", fs.StartDay, fs.StopDay))
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
