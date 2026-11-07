package sys

import (
	"gobees/comp"
	"gobees/day"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// extinction checks whether the population is extinct and ends the run [GO-20].
//
// Stops when no colonies and no bees are left and StopExtinct? is true,
// or after params.Termination.MaxTicks days (replaces the "Run for period" button).
// NetLogo's stop is the last command of Go, so the run ends after this day, as Terminate does.
// TimeToExtinction: output only, set by the observers (obs.Counters, step F4). Survived? [PAR-08]: never read,
// removed in step F3.
//
// NetLogo: end of Go.
type extinction struct {
	termRes    *resource.Termination
	termParams *params.Termination
	parameters *params.Parameters
	ticks      *globals.Ticks

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
}

func (s *extinction) Initialize(w *ecs.World) {
	s.termRes = ecs.GetResource[resource.Termination](w)
	s.termParams = ecs.GetResource[params.Termination](w)
	s.parameters = ecs.GetResource[params.Parameters](w)
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
}

func (s *extinction) Update(w *ecs.World) {
	// [GO-20] count Colonies + count Bees = 0
	colonies := s.colonyFilter.Query()
	n := colonies.Count()
	colonies.Close()
	bees := s.beeFilter.Query()
	n += bees.Count()
	bees.Close()
	if n == 0 && s.parameters.StopExtinct {
		s.termRes.Terminate = true
	}
	// not NetLogo: the run length of the "Run for period" button or BehaviorSpace
	if s.termParams.MaxTicks > 0 && s.ticks.Ticks >= day.Abs(s.termParams.MaxTicks) {
		s.termRes.Terminate = true
	}
}

func (s *extinction) Finalize(w *ecs.World) {}
