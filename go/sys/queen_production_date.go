package sys

import "github.com/mlange-42/ark/ecs"

// QueenProductionDate [GO-14].
//
// Determines when female larvae can start to develop into queens.
//
// NetLogo: QueenProductionDateProc.
type QueenProductionDate struct{}

func (s *QueenProductionDate) Initialize(w *ecs.World) {}

func (s *QueenProductionDate) Update(w *ecs.World) {}

func (s *QueenProductionDate) Finalize(w *ecs.World) {}
