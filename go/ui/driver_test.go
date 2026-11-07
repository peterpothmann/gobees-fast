//go:build ui || windows || (darwin && cgo)

package ui

import (
	"reflect"
	"testing"
)

// TestStepParts: "Next part" stops after each part of the day in schedule order, the part count goes from 1 to 19,
// and after the last part the day ends.
func TestStepParts(t *testing.T) {
	g := newTestGame(t, 1)
	for range 100 {
		g.step()
	}
	var parts []string
	for range 30 {
		g.stepPart()
		if g.stopAt == nil {
			break // end of the day
		}
		if g.stopAt.partIndex != len(parts)+1 {
			t.Fatalf("part %d reported as %d", len(parts)+1, g.stopAt.partIndex)
		}
		parts = append(parts, g.stopAt.system+"."+g.stopAt.part)
	}
	if len(parts) != 19 || parts[0] != "Morning.tick" || parts[10] != "Day.activities" ||
		parts[18] != "Evening.extinction" {
		t.Fatalf("parts of a day: %v", parts)
	}
	if g.snap.Tick != 101 || len(g.drv.lastDay) != 19 || len(g.drv.order) != 19 {
		t.Errorf("tick %d, %d parts timed, %d in order, want 101, 19, 19", g.snap.Tick, len(g.drv.lastDay),
			len(g.drv.order))
	}
}

// TestStepTurns: "Next turn" stops after each bee turn of the activity loop, in the order of the personal times, and
// shows the bee, her task and the stimuli of her colony.
func TestStepTurns(t *testing.T) {
	g := newTestGame(t, 1)
	for range 120 { // spring: colonies with workers
		g.step()
	}
	last := 0.0
	for i := 1; i <= 50; i++ {
		g.stepTurn()
		s := g.stopAt
		if s == nil || s.kind != stopTurn {
			t.Fatalf("turn %d: no stop after a turn", i)
		}
		if s.turn != i || s.part != "activities" || s.bee.IsZero() {
			t.Fatalf("turn %d: stop %+v", i, s)
		}
		b := g.dbg.bee
		if b == nil {
			t.Fatalf("turn %d: no bee", i)
		}
		if b.start < last {
			t.Fatalf("turn %d at %v after a turn at %v", i, b.start, last)
		}
		last = b.start
		if b.alive && (b.task == "" || b.end < b.start) {
			t.Errorf("turn %d: task %q, time %v to %v", i, b.task, b.start, b.end)
		}
	}
	if g.live == nil || len(g.dbg.archetypes) == 0 || len(g.dbg.resources) == 0 {
		t.Error("no live state, entities or resources in the middle of the day")
	}
	g.step() // the rest of the day
	if g.stopAt != nil || g.live != nil || g.snap.Tick != 121 {
		t.Errorf("after the rest of the day: stop %v, tick %d", g.stopAt, g.snap.Tick)
	}
}

// TestSteppingSameRun: stepping by parts and turns gives the same run as stepping by days.
func TestSteppingSameRun(t *testing.T) {
	const days = 140
	byDay := newTestGame(t, 3)
	for range days {
		byDay.step()
	}
	stepped := newTestGame(t, 3)
	// within one day: n steps, fewer if the day ends before (no bee turns in winter), then the rest of the day
	within := func(n int, next func()) {
		for range n {
			next()
			if stepped.stopAt == nil {
				return
			}
		}
		stepped.step()
	}
	for day := 1; day <= days; day++ {
		switch day % 3 {
		case 1:
			stepped.step()
		case 2:
			within(7, stepped.stepPart)
		case 0:
			within(25, stepped.stepTurn)
		}
	}
	if stepped.snap.Tick != days || !reflect.DeepEqual(stepped.hist, byDay.hist) {
		t.Errorf("stepping by parts and turns differs from stepping by days (tick %d)", stepped.snap.Tick)
	}
}
