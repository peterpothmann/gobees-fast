package sys

import "github.com/mlange-42/ark/ecs"

// UpdateMorning [GO-04].
//
// Morning routine: Day and Date, daily foraging period, flowering state of foodsources,
// seasonal events, queens emerging from hibernation, wake-up times of active bees,
// crop rotation and colony updates. Will be split into several systems once annotated [MORN-xx].
//
// NetLogo: UpdateMorning_Proc.
type UpdateMorning struct{}

func (s *UpdateMorning) Initialize(w *ecs.World) {}

// Update is not yet translated.
func (s *UpdateMorning) Update(w *ecs.World) {}

func (s *UpdateMorning) Finalize(w *ecs.World) {}
