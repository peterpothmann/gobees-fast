package sys

import "github.com/mlange-42/ark/ecs"

// CreateInitialQueens [INIT-09].
//
// Creates the hibernating queens with their emergence dates.
//
// NetLogo: CreateInitialQueensProc.
type CreateInitialQueens struct{}

func (s *CreateInitialQueens) Initialize(w *ecs.World) {}

func (s *CreateInitialQueens) Update(w *ecs.World) {}

func (s *CreateInitialQueens) Finalize(w *ecs.World) {}
