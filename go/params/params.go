package params

import "github.com/mlange-42/ark/ecs"

// Termination criteria.
//
// NetLogo has no maximum run length inside Go; it is set by the "Run for period" button or BehaviorSpace.
type Termination struct {
	MaxTicks    int  // Maximum number of ticks to run [d]. A value <= 0 means no limit.
	StopExtinct bool // NetLogo StopExtinct?: stop when no colonies and no bees are left [GO-20].
}

// Params contains all parameters of the model.
type Params struct {
	Termination Termination
}

// Default returns the default parameter set, taken from _SYSTEM_Parameters.csv.
func Default() Params {
	return Params{
		Termination: Termination{
			MaxTicks:    365,
			StopExtinct: true,
		},
	}
}

// Apply adds all parameters to the world as resources.
func (p *Params) Apply(w *ecs.World) {
	pCopy := *p
	ecs.AddResource(w, &pCopy.Termination)
}
