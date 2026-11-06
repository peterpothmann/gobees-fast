package sys

import (
	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// broodTotals sets the daily totals of the brood, which seasonalEvents reads on the next morning.
// Part of the morning routine. Initialize runs it also at ticks = 0.
//
// The checks of the bee numbers are not part of the model. The checks of the model state (the totals of the colonies
// equal the totals of the bees, no colony alive on day 364) are in internal/check, run by the tests at the end of
// every day.
//
// NetLogo: CheckNumbersProc.
type broodTotals struct {
	totals globals.BroodTotals

	colonyFilter *ecs.Filter1[comp.ColonyCounts]
}

func (s *broodTotals) Initialize(w *ecs.World) {
	ecs.AddResource(w, &s.totals)
	s.colonyFilter = ecs.NewFilter1[comp.ColonyCounts](w)
	s.Update(w) // run at ticks = 0
}

// Update: ORDER irrelevant, only sums.
func (s *broodTotals) Update(w *ecs.World) {
	t := &s.totals
	t.Eggs, t.Larvae, t.Pupae = 0, 0, 0
	colonies := s.colonyFilter.Query()
	for colonies.Next() {
		c := colonies.Get()
		t.Eggs += c.Eggs
		t.Larvae += c.Larvae
		t.Pupae += c.Pupae
	}
}

func (s *broodTotals) Finalize(w *ecs.World) {}
