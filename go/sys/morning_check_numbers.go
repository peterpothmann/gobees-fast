package sys

import (
	"gobees/comp"
	"gobees/globals"

	"github.com/mlange-42/ark/ecs"
)

// CheckNumbers sets the daily totals of the brood [CHK-02], which UpdateSeasonalEvents reads on the next morning
// [SEAS-02] (finding 39, see trans/netlogo-findings.md#finding-39).
// Part of the morning routine [MORN-12]. Initialize runs it also at ticks = 0 [INIT-10].
//
// The checks are not part of the model (step F4). [CHK-01] (initial + produced - died = present bees) and the egg
// accounting of [CHK-04] checked only NetLogo's own counters and are removed with them. [CHK-03] (the totals of the
// colonies equal the totals of the bees) and the colony check of [CHK-04] (no colony alive on day 364) are tests
// of the model state (internal/check), run by the tests at the end of every day. The totals of adult workers and
// males were read only by [CHK-03].
//
// NetLogo: CheckNumbersProc.
type CheckNumbers struct {
	totals globals.Totals

	colonyFilter *ecs.Filter1[comp.ColonyCounts]
}

func (s *CheckNumbers) Initialize(w *ecs.World) {
	ecs.AddResource(w, &s.totals)
	s.colonyFilter = ecs.NewFilter1[comp.ColonyCounts](w)
	s.Update(w) // run at ticks = 0 [INIT-10]
}

// Update: ORDER irrelevant, only sums.
func (s *CheckNumbers) Update(w *ecs.World) {
	// [CHK-02] finding 39 (see trans/netlogo-findings.md#finding-39). totalForagingTripsToday: output only, counted
	// by the observers (step F4).
	t := &s.totals
	t.TotalEggs, t.TotalLarvae, t.TotalPupae = 0, 0, 0
	colonies := s.colonyFilter.Query()
	for colonies.Next() {
		c := colonies.Get()
		t.TotalEggs += c.AllEggs
		t.TotalLarvae += c.AllLarvae
		t.TotalPupae += c.AllPupae
	}
}

func (s *CheckNumbers) Finalize(w *ecs.World) {}
