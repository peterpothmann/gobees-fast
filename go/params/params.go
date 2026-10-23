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

// WorkingDirectory is the directory of the NetLogo input files.
// File names in the parameter file (e.g. FlowerspeciesFile) are relative to it.
type WorkingDirectory struct {
	Path string
}

// Interface contains the model-relevant settings of the NetLogo interface widgets.
// All other widgets only affect the GUI.
type Interface struct {
	MyMap     string  // Map image. Only its name is used, to derive FoodsourcesFile [PAR-05].
	RAND_SEED float64 // Random seed. 0 means no seeding [INIT-01].
}

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
type Params struct {
	WorkingDirectory WorkingDirectory
	Interface        Interface
	World            World
	File             ParameterFile // Read with ReadParameterFile [INIT-04].
	Constants        Constants
	Termination      Termination
}

// Default returns the default parameters. Interface values are the ones stored in the .nlogo file.
//
// File is left empty. It must be read with ReadParameterFile. With the default interface settings,
// NetLogo reads _SYSTEM_Example_Farm_Parameters.csv (MyParametersFile, see [INIT-02]).
func Default() Params {
	return Params{
		WorkingDirectory: WorkingDirectory{Path: "."},
		Interface: Interface{
			MyMap:     "_SYSTEM_Example_Farm.png",
			RAND_SEED: 0,
		},
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
	seed := uint64(int64(p.Interface.RAND_SEED))
	if p.Interface.RAND_SEED == 0 {
		seed = rand.Uint64()
	}
	rng := ecs.GetResource[resource.Rand](w)
	rng.Source = rand.NewPCG(0, seed)

	pCopy := *p
	ecs.AddResource(w, &pCopy.WorkingDirectory)
	ecs.AddResource(w, &pCopy.Interface)
	ecs.AddResource(w, &pCopy.World)
	ecs.AddResource(w, &pCopy.File)
	ecs.AddResource(w, &pCopy.Constants)
	ecs.AddResource(w, &pCopy.Termination)
}
