package sys

import (
	"testing"

	"gobees/day"
	"gobees/globals"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

func TestTickMatchesNetLogo(t *testing.T) {
	a := app.New()
	a.AddSystem(&tick{})
	a.Initialize()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	if ticks.Ticks != 0 {
		t.Fatalf("after Setup: ticks = %d, NetLogo has 0", ticks.Ticks)
	}

	for d := day.Abs(1); d <= 3; d++ {
		a.Update()
		if ticks.Ticks != d {
			t.Fatalf("during Go call %d: ticks = %d, NetLogo has %d", d, ticks.Ticks, d)
		}
	}
}
