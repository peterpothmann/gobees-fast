package sys

import "github.com/mlange-42/ark/ecs"

// BadgersOnTheProwl [GO-17].
//
// Badgers dig up colonies within their range.
//
// NetLogo: BadgersOnTheProwlProc.
type BadgersOnTheProwl struct{}

func (s *BadgersOnTheProwl) Initialize(w *ecs.World) {}

func (s *BadgersOnTheProwl) Update(w *ecs.World) {}

func (s *BadgersOnTheProwl) Finalize(w *ecs.World) {}
