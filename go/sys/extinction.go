package sys

import (
	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Extinction checks whether the population is extinct and ends the run [GO-20].
//
// Stops when no colonies and no bees are left and StopExtinct? is true,
// or after params.Termination.MaxTicks days (replaces the "Run for period" button).
// NetLogo's stop is the last command of Go, so the run ends after this day, as Terminate does.
// Initialize sets Survived? = 1, as ParametersProc does [PAR-08].
//
// NetLogo: end of Go.
type Extinction struct {
	termRes    *resource.Termination
	termParams *params.Termination
	file       *params.ParameterFile
	ticks      *globals.Ticks
	extinction globals.Extinction

	colonyFilter *ecs.Filter1[comp.Colony]
	beeFilter    *ecs.Filter1[comp.Bee]
}

func (s *Extinction) Initialize(w *ecs.World) {
	s.termRes = ecs.GetResource[resource.Termination](w)
	s.termParams = ecs.GetResource[params.Termination](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.extinction = globals.Extinction{Survived: 1} // [PAR-08]
	ecs.AddResource(w, &s.extinction)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
}

func (s *Extinction) Update(w *ecs.World) {
	// [GO-20] count Colonies + count Bees = 0
	colonies := s.colonyFilter.Query()
	n := colonies.Count()
	colonies.Close()
	bees := s.beeFilter.Query()
	n += bees.Count()
	bees.Close()
	if n == 0 {
		if s.extinction.TimeToExtinction <= 0 {
			s.extinction.TimeToExtinction = float64(s.ticks.Ticks)
		}
		s.extinction.Survived = 0
		if s.file.StopExtinct {
			s.termRes.Terminate = true
		}
	}
	// not NetLogo: the run length of the "Run for period" button or BehaviorSpace
	if s.termParams.MaxTicks > 0 && s.ticks.Ticks >= int64(s.termParams.MaxTicks) {
		s.termRes.Terminate = true
	}
}

func (s *Extinction) Finalize(w *ecs.World) {}
