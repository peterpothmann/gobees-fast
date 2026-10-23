package sys

import "github.com/mlange-42/ark/ecs"

// Development [GO-15].
//
// Ages all bees by one day, develops brood into the next stage and applies adult mortality.
//
// NetLogo: DevelopmentProc.
type Development struct{}

func (s *Development) Initialize(w *ecs.World) {}

func (s *Development) Update(w *ecs.World) {}

func (s *Development) Finalize(w *ecs.World) {}
