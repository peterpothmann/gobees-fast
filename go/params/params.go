// Package params contains the model parameters as ECS resources.
//
// Naming: fields keep the NetLogo names, so they can be found in the NetLogo code.
// Characters that are invalid in Go are replaced: "?" is dropped, "/" becomes "_per_".
// Numbers are float64 like in NetLogo. Conversions to integers happen where NetLogo converts.
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
//
// NetLogo has no maximum run length inside Go; it is set by the "Run for period" button or BehaviorSpace.
type Termination struct {
	MaxTicks int // Maximum number of ticks to run [d]. A value <= 0 means no limit.
}

// Params contains all parameters of the model.
//
// Of the NetLogo interface settings, only RAND_SEED (Seed) and the world size (World) remain. The landscape of the
// start (NetLogo: MyMap) is in the run file.
type Params struct {
	Seed        int64 // Random seed (RAND_SEED). 0 means a random seed [INIT-01].
	World       World
	File        ParameterFile // The parameters of the run file, read with input.Load [INIT-04].
	Output      Output        // The output options of the run file (step F13).
	Input       Input         // The other content of the run file and its landscape files, read with input.Load.
	Constants   Constants
	Termination Termination
}

// Default returns the default parameters. Seed and World are the values stored in the .nlogo file.
//
// File and Input are left empty. input.Load reads both from a run file, input.ExampleFarm returns the example farm
// (NetLogo: _SYSTEM_Example_Farm_Parameters.csv with the default interface settings, MyParametersFile [INIT-02]).
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

// Apply seeds the random number generator [INIT-01] and adds all parameters to the world as resources.
func (p *Params) Apply(w *ecs.World) {
	// [INIT-01] NetLogo seeds only if RAND_SEED != 0. NetLogo's random sequence cannot be reproduced.
	seed := uint64(p.Seed)
	if p.Seed == 0 {
		seed = rand.Uint64()
	}
	rng := ecs.GetResource[resource.Rand](w)
	rng.Source = rand.NewPCG(0, seed)

	pCopy := *p
	ecs.AddResource(w, &pCopy.World)
	ecs.AddResource(w, &pCopy.File)
	ecs.AddResource(w, &pCopy.Output)
	ecs.AddResource(w, &pCopy.Input)
	ecs.AddResource(w, &pCopy.Constants)
	ecs.AddResource(w, &pCopy.Termination)
}
