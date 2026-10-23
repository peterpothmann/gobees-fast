package sys

import "github.com/mlange-42/ark/ecs"

// ActivityLoop [GO-06 to GO-10].
//
// Event-driven loop of the day. Repeatedly takes the active bee with the smallest personalTime_s,
// lets it perform exactly one activity and re-inserts it, until the smallest personalTime_s
// exceeds CallItaDay_s, no active adult worker or queen is left, or no colony is left.
//
// NetLogo: while loop in Go, NextActiveBeeREP, ActivityProc.
type ActivityLoop struct{}

func (s *ActivityLoop) Initialize(w *ecs.World) {}

func (s *ActivityLoop) Update(w *ecs.World) {}

func (s *ActivityLoop) Finalize(w *ecs.World) {}
