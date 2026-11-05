package sys

import (
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

func TestExtinction(t *testing.T) {
	// [GO-20] the run stops on the first day without colonies and bees, only with StopExtinct?.
	// TimeToExtinction: output, see obs.Counters.
	for _, stop := range []bool{false, true} {
		cw := newColonyWorld(t, 1)
		cw.file.StopExtinct = stop
		term := &resource.Termination{}
		ecs.AddResource(cw.w, term)
		ecs.AddResource(cw.w, &params.Termination{MaxTicks: 0})
		s := &extinction{}
		s.Initialize(cw.w)

		// a hibernating queen alone keeps the population alive, a colony alone too
		queen := cw.addHibernatingQueen(1)
		col := cw.addColony(nil)
		cw.ticks.Ticks = 200
		s.Update(cw.w)
		cw.w.RemoveEntity(queen)
		cw.ticks.Ticks = 201
		s.Update(cw.w)
		if term.Terminate {
			t.Fatal("population alive: terminate")
		}

		cw.w.RemoveEntity(col)
		cw.ticks.Ticks = 202
		s.Update(cw.w)
		cw.ticks.Ticks = 203
		s.Update(cw.w) // NetLogo with StopExtinct? = false runs on
		if term.Terminate != stop {
			t.Errorf("StopExtinct? %v: terminate %v", stop, term.Terminate)
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
	s := &extinction{}
	s.Initialize(cw.w)
	for _, ticks := range []day.Abs{364, 365} {
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
