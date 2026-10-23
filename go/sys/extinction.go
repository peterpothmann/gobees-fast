package sys

import (
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Extinction ends the run [GO-20].
//
// Stops when no colonies and no bees are left and StopExtinct? is true,
// or after params.Termination.MaxTicks days (replaces the "Run for period" button).
//
// NetLogo: end of Go.
type Extinction struct {
	termRes    *resource.Termination
	termParams *params.Termination
	file       *params.ParameterFile
	ticks      *globals.Ticks
}

func (s *Extinction) Initialize(w *ecs.World) {
	s.termRes = ecs.GetResource[resource.Termination](w)
	s.termParams = ecs.GetResource[params.Termination](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.ticks = ecs.GetResource[globals.Ticks](w)
}

func (s *Extinction) Update(w *ecs.World) {
	// The extinction check (count Colonies + count Bees = 0, TimeToExtinction, Survived?, s.file.StopExtinct)
	// is not yet translated, as there are no colonies and bees yet.
	if s.termParams.MaxTicks > 0 && s.ticks.Ticks >= int64(s.termParams.MaxTicks) {
		s.termRes.Terminate = true
	}
}

func (s *Extinction) Finalize(w *ecs.World) {}
