package sys

import "github.com/mlange-42/ark/ecs"

// OutputDaily [GO-18].
//
// Calculates the daily totals (Total... globals). Some of them are read by model procedures.
// Plots are not translated.
//
// NetLogo: OutputDailyProc.
type OutputDaily struct{}

func (s *OutputDaily) Initialize(w *ecs.World) {}

func (s *OutputDaily) Update(w *ecs.World) {}

func (s *OutputDaily) Finalize(w *ecs.World) {}
