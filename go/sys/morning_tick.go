package sys

import (
	"gobees/day"
	"gobees/globals"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// tick advances the NetLogo day counter [GO-03].
//
// NetLogo: tick (in Go).
type tick struct {
	tick  *resource.Tick
	ticks globals.Ticks
}

func (s *tick) Initialize(w *ecs.World) {
	s.tick = ecs.GetResource[resource.Tick](w)
	s.ticks = globals.Ticks{Ticks: 0}
	ecs.AddResource(w, &s.ticks)
}

func (s *tick) Update(w *ecs.World) {
	s.ticks.Ticks = day.Abs(s.tick.Tick + 1)
}

func (s *tick) Finalize(w *ecs.World) {}
