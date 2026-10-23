package sys

import "github.com/mlange-42/ark/ecs"

// EndOfDayResting [GO-11].
//
// Sets all adult, non-hibernating females to resting.
//
// NetLogo: ask bees ... set activity "resting" (in Go).
type EndOfDayResting struct{}

func (s *EndOfDayResting) Initialize(w *ecs.World) {}

func (s *EndOfDayResting) Update(w *ecs.World) {}

func (s *EndOfDayResting) Finalize(w *ecs.World) {}
