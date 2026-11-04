package sys

import (
	"gobees/day"
	"gobees/globals"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Tick advances the NetLogo day counter [GO-03].
//
// NetLogo: tick (in Go).
type Tick struct {
	tick  *resource.Tick
	ticks globals.Ticks
}

func (s *Tick) Initialize(w *ecs.World) {
	s.tick = ecs.GetResource[resource.Tick](w)
	s.ticks = globals.Ticks{Ticks: 0}
	ecs.AddResource(w, &s.ticks)
}

func (s *Tick) Update(w *ecs.World) {
	s.ticks.Ticks = day.Abs(s.tick.Tick + 1)
}

func (s *Tick) Finalize(w *ecs.World) {}
