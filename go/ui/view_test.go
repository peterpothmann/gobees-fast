//go:build ui || windows || (darwin && cgo)

package ui

import "testing"

// TestViewBackForward: ← goes back through the kept days and stops at the oldest, → goes forward until the latest
// day is shown again.
func TestViewBackForward(t *testing.T) {
	g := &Game{}
	for tick := int64(5); tick <= 15; tick++ { // kept days: ticks 5 to 15
		g.days = append(g.days, snapshot{Tick: tick})
	}
	g.snap = g.days[len(g.days)-1]

	g.viewBack()
	if !g.viewing || g.shown().Tick != 14 {
		t.Fatalf("one back: viewing %v, tick %d, want true, 14", g.viewing, g.shown().Tick)
	}
	for range 20 {
		g.viewBack()
	}
	if g.shown().Tick != 5 {
		t.Fatalf("back to the oldest kept day: tick %d, want 5", g.shown().Tick)
	}
	for i := 0; i < 9; i++ {
		if !g.viewForward() {
			t.Fatalf("forward %d: latest day reached too early", i)
		}
	}
	if !g.viewing || g.shown().Tick != 14 {
		t.Fatalf("9 forward: viewing %v, tick %d, want true, 14", g.viewing, g.shown().Tick)
	}
	if !g.viewForward() || g.viewing || g.shown().Tick != 15 {
		t.Fatalf("10th forward: viewing %v, tick %d, want false, 15", g.viewing, g.shown().Tick)
	}
	if g.viewForward() {
		t.Fatal("forward at the latest day must simulate, not view")
	}
}
