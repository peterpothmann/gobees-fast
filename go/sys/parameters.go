package sys

import (
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// Parameters adds the globals of ParametersProc as resources [INIT-05].
//
// Package input derives their values before the setup (params.Input): [PAR-01], [PAR-04] to [PAR-07] and the
// weather [FPER]. The hard-coded constants [PAR-03] are in params.Constants. The output variables [PAR-08]
// (PopulationSizeDay365List, Survived?) are never read and were removed in step F3.
// Display-only values and the check loop [PAR-09] are not translated.
//
// NetLogo: ParametersProc.
type Parameters struct {
	derived globals.Derived
	weather globals.Weather
}

func (s *Parameters) Initialize(w *ecs.World) {
	in := ecs.GetResource[params.Input](w)

	// [FPER] the weather, read once (NetLogo: Foraging_periodREP every morning)
	s.weather = in.Weather
	ecs.AddResource(w, &s.weather)

	// a copy, because CropRotationProc changes FoodsourcesFile [CROP-01]
	s.derived = in.Derived
	ecs.AddResource(w, &s.derived)
}

func (s *Parameters) Update(w *ecs.World) {}

func (s *Parameters) Finalize(w *ecs.World) {}
