package params

import (
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// World is the size of the NetLogo world (interface: graphics window). It does not wrap.
type World struct {
	MinPxcor float64
	MaxPxcor float64
	MinPycor float64
	MaxPycor float64
}

// Termination criteria.
type Termination struct {
	MaxTicks int // Maximum number of ticks to run [d]. A value <= 0 means no limit.
}

// Params contains all parameters of the model.
//
// Of the NetLogo interface settings, only RAND_SEED (Seed) and the world size (World) remain.
type Params struct {
	Seed        int64 // Random seed. 0 means a random seed. NetLogo: RAND_SEED
	World       World
	Parameters  Parameters // The parameters of the run file, read with input.Load.
	Output      Output     // The output options of the run file.
	Input       Input      // The other content of the run file and its landscape files, read with input.Load.
	Constants   Constants
	Termination Termination
}

// Default returns the default parameters. Seed and World are the values stored in the .nlogo file.
//
// Parameters, Output and Input are left empty. input.Load reads them from a run file, input.ExampleFarm returns
// the example farm.
func Default() Params {
	return Params{
		Seed:      0,
		World:     World{MinPxcor: 0, MaxPxcor: 300, MinPycor: 0, MaxPycor: 210},
		Constants: DefaultConstants(),
		Termination: Termination{
			MaxTicks: 365,
		},
	}
}

// Apply seeds the random number generator and adds all parameters to the world as resources.
func (p *Params) Apply(w *ecs.World) {
	seed := uint64(p.Seed)
	if p.Seed == 0 {
		seed = rand.Uint64()
	}
	rng := ecs.GetResource[resource.Rand](w)
	rng.Source = rand.NewPCG(0, seed)

	pCopy := *p
	ecs.AddResource(w, &pCopy.World)
	ecs.AddResource(w, &pCopy.Parameters)
	ecs.AddResource(w, &pCopy.Output)
	ecs.AddResource(w, &pCopy.Input)
	ecs.AddResource(w, &pCopy.Input.Derived)
	ecs.AddResource(w, &pCopy.Input.ForagingPeriod)
	ecs.AddResource(w, &pCopy.Constants)
	ecs.AddResource(w, &pCopy.Termination)
}
