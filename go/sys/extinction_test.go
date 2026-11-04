package sys

import (
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

func TestExtinction(t *testing.T) {
	// [GO-20] TimeToExtinction is the first day without colonies and bees, and the run stops only with
	// StopExtinct?.
	for _, stop := range []bool{false, true} {
		cw := newColonyWorld(t, 1)
		cw.file.StopExtinct = stop
		term := &resource.Termination{}
		ecs.AddResource(cw.w, term)
		ecs.AddResource(cw.w, &params.Termination{MaxTicks: 0})
		s := &Extinction{}
		s.Initialize(cw.w)
		ext := ecs.GetResource[globals.Extinction](cw.w)
		if ext.TimeToExtinction != 0 {
			t.Fatalf("after Setup: TimeToExtinction %v, want 0", ext.TimeToExtinction)
		}

		// a hibernating queen alone keeps the population alive, a colony alone too
		queen := cw.addHibernatingQueen(1)
		col := cw.addColony(nil)
		cw.ticks.Ticks = 200
		s.Update(cw.w)
		cw.w.RemoveEntity(queen)
		cw.ticks.Ticks = 201
		s.Update(cw.w)
		if ext.TimeToExtinction != 0 || term.Terminate {
			t.Fatalf("population alive: TimeToExtinction %v, terminate %v", ext.TimeToExtinction, term.Terminate)
		}

		cw.w.RemoveEntity(col)
		cw.ticks.Ticks = 202
		s.Update(cw.w)
		cw.ticks.Ticks = 203
		s.Update(cw.w) // NetLogo with StopExtinct? = false runs on: the first day is kept
		if ext.TimeToExtinction != 202 || term.Terminate != stop {
			t.Errorf("StopExtinct? %v: TimeToExtinction %v, terminate %v; want 202, %v",
				stop, ext.TimeToExtinction, term.Terminate, stop)
		}
	}
}

func TestExtinctionMaxTicks(t *testing.T) {
	// not NetLogo: the run length of the "Run for period" button
	cw := newColonyWorld(t, 1)
	cw.addBee(cw.addColony(nil), stage.Adult, caste.Queen, 1, 0)
	term := &resource.Termination{}
	ecs.AddResource(cw.w, term)
	ecs.AddResource(cw.w, &params.Termination{MaxTicks: 365})
	s := &Extinction{}
	s.Initialize(cw.w)
	for _, ticks := range []int64{364, 365} {
		cw.ticks.Ticks = ticks
		s.Update(cw.w)
		if term.Terminate != (ticks == 365) {
			t.Errorf("ticks %d: terminate %v", ticks, term.Terminate)
		}
	}
	if n := countEntities(ecs.NewFilter1[comp.Bee](cw.w)); n != 1 {
		t.Errorf("%d bees", n)
	}
}
