package sys

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"
)

func TestQueenProductionDate(t *testing.T) {
	// [QPD-02] worked example: switch point on day 100
	for _, tc := range []struct {
		name       string
		ticks      day.Abs
		larvalAge3 int // broodAge of a larva at larval day 3, 0 for none
		ratio      float64
		switchDay  day.Date
		want       bool
		wantDate   day.Date // unset without queen production
	}{
		{"no larvae, day 104", 104, 0, 1, day.At(100), false, day.Date{}},
		{"no larvae, day 105", 105, 0, 1, day.At(100), true, day.At(95)},
		{"larvae of broodAge 8, day 102", 102, 8, 1, day.At(100), false, day.Date{}},
		{"larvae of broodAge 8, day 103", 103, 8, 1, day.At(100), true, day.At(95)},
		{"larva:worker ratio 3", 120, 0, 3, day.At(100), false, day.Date{}},
		{"no switch point", 120, 0, 1, day.Date{}, false, day.Date{}},
	} {
		bw := newBroodWorld(t, 1)
		bw.ticks.Ticks = tc.ticks
		col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
			c.SwitchPointDate = tc.switchDay
			n.LarvaWorkerRatio = tc.ratio
		})
		if tc.larvalAge3 > 0 {
			e := bw.addBrood(col, stage.Larva, caste.Worker, 12, tc.larvalAge3, 50, 0)
			_, _, _, _, d := bw.bee(e)
			d.CumulTimeLarva_d = 3
			// a larva at another larval day is ignored
			e = bw.addBrood(col, stage.Larva, caste.Undefined, 12, 20, 50, 0)
			_, _, _, _, d = bw.bee(e)
			d.CumulTimeLarva_d = 2
		}
		bw.qpd.Update(bw.w)
		c, _, _ := bw.colony(col)
		if c.QueenProduction != tc.want || c.QueenProductionDate != tc.wantDate {
			t.Errorf("%s: queenProduction %v, date %v; want %v, %v", tc.name, c.QueenProduction, c.QueenProductionDate, tc.want, tc.wantDate)
		}
	}
}

func TestQueenProductionDateBeforeDevelopment(t *testing.T) {
	// finding 42 (see trans/netlogo-findings.md#finding-42): the date is the laying day of the last worker larvae
	bw := newBroodWorld(t, 1)
	bw.ticks.Ticks = 110
	col := bw.addColony(func(c *comp.Colony, s *comp.ColonyStores, n *comp.ColonyCounts) {
		c.SwitchPointDate = day.At(100)
		n.LarvaWorkerRatio = 1
	})
	worker := bw.addBrood(col, stage.Larva, caste.Worker, 12, 8, 30, 0) // larval day 3 since last evening
	_, _, _, _, d := bw.bee(worker)
	d.CumulTimeLarva_d = 3
	female := bw.addBrood(col, stage.Larva, caste.Undefined, 12, 7, 20, 0) // reaches larval day 3 tonight
	_, _, _, _, d = bw.bee(female)
	d.CumulTimeLarva_d = 2

	bw.qpd.Update(bw.w)
	bw.dev.Update(bw.w)

	c, _, _ := bw.colony(col)
	if !c.QueenProduction || c.QueenProductionDate != day.At(102) {
		t.Errorf("queenProduction %v, date %v; want true, 102", c.QueenProduction, c.QueenProductionDate)
	}
	if b, _, _, _, _ := bw.bee(worker); b.Caste != caste.Worker {
		t.Errorf("larva laid on day 102 is %v, want worker", b.Caste)
	}
	if b, _, _, _, _ := bw.bee(female); b.Caste != caste.Queen {
		t.Errorf("larva laid on day 103 is %v, want queen", b.Caste)
	}
}
