package schedule

import (
	"gobees/globals"
	"gobees/params"
	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// Default sets up the default gobees model: the parameters and the global state as resources, then the four systems
// sys.Setup, sys.Morning, sys.Day and sys.Evening. Setup runs in the systems' Initialize, the other three each day.
// The run file must already be read into p (input.Load).
//
// If the argument app is nil, a new model instance is created.
// If it is non-nil, the model is reset and re-used.
func Default(p params.Params, a *app.App) *app.App {
	if a == nil {
		a = app.New()
	} else {
		a.Reset()
	}

	p.Apply(a.World) // random seed, parameters as resources
	activeBees := globals.ActiveBees{}
	ecs.AddResource(a.World, &activeBees)
	daytime := globals.Daytime{}
	ecs.AddResource(a.World, &daytime)
	a.AddSystem(&sys.Setup{})

	// one update = one day
	a.AddSystem(&sys.Morning{}) // its Initialize runs the morning once at day 0
	a.AddSystem(&sys.Day{})
	a.AddSystem(&sys.Evening{})
	return a
}
