package sys

import "github.com/mlange-42/ark/ecs"

// CreateSpecies [INIT-07].
//
// Creates one species entity per simulated species, incl. its nest-site foodsources.
//
// NetLogo: CreateSpeciesProc.
type CreateSpecies struct{}

func (s *CreateSpecies) Initialize(w *ecs.World) {}

func (s *CreateSpecies) Update(w *ecs.World) {}

func (s *CreateSpecies) Finalize(w *ecs.World) {}
