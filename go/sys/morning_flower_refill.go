package sys

import (
	"fmt"

	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// flowerRefill fills every layer in flower up to its maximum and empties all others. It sets the flags
// comp.FlowerPatch.PollenInFlower and NectarInFlower, which colonyFounder.flowersInRange reads.
//
// Part of the morning routine. Initialize runs it also at ticks = 0, i.e. Day = 0.
//
// NetLogo: UpdateFoodsourcesProc.
type flowerRefill struct {
	calendar *globals.Calendar
	today    globals.Flowering

	filter *ecs.Filter1[comp.FlowerPatch]
}

func (s *flowerRefill) Initialize(w *ecs.World) {
	s.calendar = ecs.GetResource[globals.Calendar](w)
	ecs.AddResource(w, &s.today)
	s.filter = ecs.NewFilter1[comp.FlowerPatch](w)
	s.Update(w) // run at ticks = 0
}

func (s *flowerRefill) Update(w *ecs.World) {
	day := s.calendar.Day
	s.today.Changed = false

	// ORDER: irrelevant, each layer only changes itself. Not shuffled (NetLogo: random order).
	// NetLogo: a flower patch without layers is a masterpatch without food with StartDay = StopDay = 0. It would
	// set Changed on day 0 only, the morning at ticks = 0, when no colony exists.
	query := s.filter.Query()
	for query.Next() {
		fp := query.Get()
		fp.PollenInFlower, fp.NectarInFlower = false, false
		for i := range fp.Layers {
			l := &fp.Layers[i]
			if l.StartDay > l.StopDay {
				panic(fmt.Sprintf("Foodsource: startDay > StopDay! (flowerRefill): startDay %v, stopDay %v", l.StartDay, l.StopDay))
			}
			if day == l.StartDay || day == l.StopDay {
				s.today.Changed = true
			}
			if day >= l.StartDay && day < l.StopDay {
				l.Nectar = l.NectarMax
				l.Pollen = l.PollenMax
			} else {
				l.Nectar = 0
				l.Pollen = 0
			}
			fp.PollenInFlower = fp.PollenInFlower || l.Pollen > 0
			fp.NectarInFlower = fp.NectarInFlower || l.Nectar > 0
		}
	}
}

func (s *flowerRefill) Finalize(w *ecs.World) {}
