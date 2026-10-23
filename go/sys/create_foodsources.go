package sys

import "github.com/mlange-42/ark/ecs"

// CreateFoodsources [INIT-06].
//
// Reads the Foodsources file and splits each foodsource into one layer per flower species.
//
// NetLogo: CreateFoodsourcesProc.
type CreateFoodsources struct{}

func (s *CreateFoodsources) Initialize(w *ecs.World) {}

func (s *CreateFoodsources) Update(w *ecs.World) {}

func (s *CreateFoodsources) Finalize(w *ecs.World) {}
