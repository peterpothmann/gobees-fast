package params

import (
	"math/rand/v2"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// Termination criteria.
type Termination struct {
	MaxTicks int // Maximum number of ticks to run [d]. A value <= 0 means no limit.
}

// Params contains all parameters of the model.
//
// Of the NetLogo interface settings, only RAND_SEED (Seed) remains. The world size of NetLogo is gone: the landscape
// has no edge.
type Params struct {
	Seed        int64      // Random seed, used as it is (the run file needs 1 or more). NetLogo: RAND_SEED
	Parameters  Parameters // The parameters of the run file, read with input.Load.
	Output      Output     // The output options of the run file.
	Input       Input      // The other content of the run file and its landscape files, read with input.Load.
	Constants   Constants
	Termination Termination
}

// Default returns the default parameters. Seed is the value stored in the .nlogo file.
//
// Parameters, Output and Input are left empty. input.Load reads them from a run file, input.ExampleFarm returns
// the example farm.
func Default() Params {
	return Params{
		Seed:      0,
		Constants: DefaultConstants(),
		Termination: Termination{
			MaxTicks: 365,
		},
	}
}

// Apply seeds the random number generator with Seed and adds all parameters to the world as resources.
func (p *Params) Apply(w *ecs.World) {
	rng := ecs.GetResource[resource.Rand](w)
	rng.Source = rand.NewPCG(0, uint64(p.Seed))

	pCopy := *p
	ecs.AddResource(w, &pCopy.Parameters)
	ecs.AddResource(w, &pCopy.Output)
	ecs.AddResource(w, &pCopy.Input)
	ecs.AddResource(w, &pCopy.Input.Derived)
	ecs.AddResource(w, &pCopy.Input.ForagingPeriod)
	ecs.AddResource(w, &pCopy.Constants)
	ecs.AddResource(w, &pCopy.Termination)
}
