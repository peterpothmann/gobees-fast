package sys

import "github.com/mlange-42/ark/ecs"

// NeedNectarPollenLarvaeToday [GO-05].
//
// Calculates the nectar and pollen the larvae of each colony need today.
//
// NetLogo: NeedNectarPollenLarvaeTodayProc.
type NeedNectarPollenLarvaeToday struct{}

func (s *NeedNectarPollenLarvaeToday) Initialize(w *ecs.World) {}

func (s *NeedNectarPollenLarvaeToday) Update(w *ecs.World) {}

func (s *NeedNectarPollenLarvaeToday) Finalize(w *ecs.World) {}
