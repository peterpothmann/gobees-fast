package sys

import "github.com/mlange-42/ark/ecs"

// MortalityBrood [GO-16].
//
// Kills brood that exceeded its maximum development age, and all eggs after the competition point.
//
// NetLogo: MortalityBroodProc.
type MortalityBrood struct{}

func (s *MortalityBrood) Initialize(w *ecs.World) {}

func (s *MortalityBrood) Update(w *ecs.World) {}

func (s *MortalityBrood) Finalize(w *ecs.World) {}
