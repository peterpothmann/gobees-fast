package sys

import "github.com/mlange-42/ark/ecs"

// FeedLarvae [GO-13].
//
// Feeds the larvae from the colony stores and calculates their weight gain.
//
// NetLogo: FeedLarvaeProc.
type FeedLarvae struct{}

func (s *FeedLarvae) Initialize(w *ecs.World) {}

func (s *FeedLarvae) Update(w *ecs.World) {}

func (s *FeedLarvae) Finalize(w *ecs.World) {}
