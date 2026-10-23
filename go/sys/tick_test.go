package sys

import (
	"testing"

	"gobees/globals"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

func TestTickMatchesNetLogo(t *testing.T) {
	a := app.New()
	a.AddSystem(&Tick{})
	a.Initialize()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	if ticks.Ticks != 0 {
		t.Fatalf("after Setup: ticks = %d, NetLogo has 0", ticks.Ticks)
	}

	for day := int64(1); day <= 3; day++ {
		a.Update()
		if ticks.Ticks != day {
			t.Fatalf("during Go call %d: ticks = %d, NetLogo has %d", day, ticks.Ticks, day)
		}
	}
}
