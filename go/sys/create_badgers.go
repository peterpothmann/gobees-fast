package sys

import "github.com/mlange-42/ark/ecs"

// CreateBadgers [INIT-08].
//
// Creates up to N_Badgers badgers on "Scrub" masterpatches, at least 300 m apart.
//
// NetLogo: CreateBadgersProc.
type CreateBadgers struct{}

func (s *CreateBadgers) Initialize(w *ecs.World) {}

func (s *CreateBadgers) Update(w *ecs.World) {}

func (s *CreateBadgers) Finalize(w *ecs.World) {}
