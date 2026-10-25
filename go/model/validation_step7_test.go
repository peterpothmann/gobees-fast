package model

import (
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// step7MetricNames are the metrics of go/validation/experiments/step7.xml, in the order of
// go/validation/scripts/aggregate_netlogo.py.
var step7MetricNames = []string{"colonies", "dead_colonies", "colonies_founded", "eggs", "larvae", "pupae", "workers",
	"males", "young_queens", "hibernating_queens", "energy_kJ", "pollen_g", "trips", "nectar_left_l", "pollen_left_kg"}

// step7Metrics returns the metrics of go/validation/experiments/step7.xml for the current state of the world,
// with the same definitions as the NetLogo reporters in the experiment file.
func step7Metrics(w *ecs.World) map[string]float64 {
	m := map[string]float64{}
	cols := ecs.NewFilter1[comp.ColonyStores](w).Query()
	for cols.Next() {
		s := cols.Get()
		m["colonies"]++
		m["energy_kJ"] += s.EnergyStore_kJ
		m["pollen_g"] += s.PollenStore_g
	}
	m["dead_colonies"] = float64(len(ecs.GetResource[globals.DeadColonies](w).List))
	m["colonies_founded"] = ecs.GetResource[globals.Counters](w).TotalColoniesEverProduced

	bees := ecs.NewFilter1[comp.Bee](w).Query()
	for bees.Next() {
		b := bees.Get()
		switch b.Stage {
		case stage.Egg:
			m["eggs"] += b.Number
		case stage.Larva:
			m["larvae"] += b.Number
		case stage.Pupa:
			m["pupae"] += b.Number
		case stage.Adult:
			switch {
			case b.Caste == caste.Worker:
				m["workers"] += b.Number
			case b.Caste == caste.Male:
				m["males"] += b.Number
			case b.Caste == caste.Queen && !b.Mated:
				m["young_queens"] += b.Number
			}
		}
		if b.Activity == activity.Hibernate {
			m["hibernating_queens"] += b.Number
		}
	}
	m["trips"] = ecs.GetResource[globals.Totals](w).TotalForagingTripsToday

	fs := ecs.NewFilter1[comp.FoodsourceState](w).Query()
	for fs.Next() {
		s := fs.Get()
		m["nectar_left_l"] += s.Nectar_myl
		m["pollen_left_kg"] += s.Pollen_g
	}
	m["nectar_left_l"] /= 1000000
	m["pollen_left_kg"] /= 1000
	return m
}

func TestValidationStep7AgainstNetLogo(t *testing.T) {
	// Experiment go/validation/experiments/step7.xml: the full model (NetLogo Go, gobees model.Default) over 3 years
	// on the example farm. Both run with StopExtinct? = false [GO-20].
	compareWithNetLogo(t, netLogoComparison{
		reference: validation.NetLogoStep7,
		order:     "07.1",
		step:      "Step 7: the year cycle",
		title:     "Example farm, 3 years",
		description: "Experiment `go/validation/experiments/step7.xml`: example farm, 500 B. terrestris queens, 1095 days, the " +
			"full model (NetLogo `Go`, gobees `model.Default`) with the year cycle: young queens leave the nest and " +
			"hibernate, the season ends, the queens emerge again. Both run with StopExtinct? = false. Compared " +
			"at every step: the gobees median must lie in the NetLogo 5-95 % band. The table shows day 515 (day 150 " +
			"of year 2).",
		metrics:   step7MetricNames,
		steps:     1095,
		shownStep: 515, // day 150 of year 2
		measure:   step7Metrics,
	})
}
