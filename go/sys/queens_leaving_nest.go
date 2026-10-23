package sys

import "github.com/mlange-42/ark/ecs"

// QueensLeavingNest [GO-12].
//
// Young queens leave the nest, mate and start hibernation.
//
// NetLogo: QueensLeavingNestProc.
type QueensLeavingNest struct{}

func (s *QueensLeavingNest) Initialize(w *ecs.World) {}

func (s *QueensLeavingNest) Update(w *ecs.World) {}

func (s *QueensLeavingNest) Finalize(w *ecs.World) {}
