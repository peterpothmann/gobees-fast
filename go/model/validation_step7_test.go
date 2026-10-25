package model

import (
	"fmt"
	"math"
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
	if testing.Short() {
		t.Skip("100 runs of 1095 days")
	}
	reference, err := validation.Parse(validation.NetLogoStep7)
	if err != nil {
		t.Fatal(err)
	}
	const runs, steps = 100, 1095
	const shownStep = 515 // day 150 of year 2

	values := map[string][][]float64{} // values[metric][step][run]
	for _, name := range step7MetricNames {
		values[name] = make([][]float64, steps+1)
	}
	record := func(step int, m map[string]float64) {
		for _, name := range step7MetricNames {
			values[name][step] = append(values[name][step], m[name])
		}
	}
	for run := 1; run <= runs; run++ {
		p := exampleFarm(t, float64(run))
		p.Termination.MaxTicks = steps
		p.File.StopExtinct = false // as the experiment [GO-20]
		a := Default(p, nil)
		a.Initialize()
		record(0, step7Metrics(a.World))
		for step := 1; step <= steps; step++ {
			a.Update()
			record(step, step7Metrics(a.World))
		}
	}

	rows := []validation.ReportRow{}
	failures := []string{}
	for _, metric := range step7MetricNames {
		outside := []string{}
		worst, worstStep, worstGobees := 0.0, 0, 0.0 // largest |median difference| / width of the NetLogo band
		for step := 0; step <= steps; step++ {
			q, ok := reference[validation.Key{Step: step, Metric: metric}]
			if !ok {
				t.Fatalf("no NetLogo reference for %s at step %d", metric, step)
			}
			// a relative tolerance of 1e-9 for the float sums (stores, landscape), whose last digits depend on the order
			tol := 1e-9 * math.Max(1, math.Abs(q.Q50))
			median := validation.Median(values[metric][step])
			if width := q.Q95 - q.Q05; width > tol && math.Abs(median-q.Q50)/width > worst {
				worst, worstStep, worstGobees = math.Abs(median-q.Q50)/width, step, median
			}
			if median < q.Q05-tol || median > q.Q95+tol {
				outside = append(outside, fmt.Sprintf("step %d: gobees median %.6g, NetLogo 5-95 %% [%.6g, %.6g]", step, median, q.Q05, q.Q95))
			}
		}
		t.Logf("%-18s %4d of %4d steps within the band", metric, steps+1-len(outside), steps+1)
		mid := reference[validation.Key{Step: shownStep, Metric: metric}]
		t.Logf("%-18s day %d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g]",
			metric, shownStep, validation.Median(values[metric][shownStep]), mid.Q50, mid.Q05, mid.Q95)
		w := reference[validation.Key{Step: worstStep, Metric: metric}]
		t.Logf("%-18s largest deviation on day %4d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g] (%.2f of the band width)",
			metric, worstStep, worstGobees, w.Q50, w.Q05, w.Q95, worst)
		rows = append(rows, validation.ReportRow{Metric: metric, Steps: steps + 1, Within: steps + 1 - len(outside),
			Step: shownStep, Gobees: validation.Median(values[metric][shownStep]), NetLogo: mid})
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: gobees median outside the NetLogo 5-95 %% band at %d of %d steps, e.g. %s",
				metric, len(outside), steps+1, outside[0]))
		}
	}

	err = validation.WriteReport("Step 7: the full model over 3 years",
		"Experiment `go/validation/experiments/step7.xml`: example farm, 500 B. terrestris queens, 1095 days, the "+
			"full model (NetLogo `Go`, gobees `model.Default`) with the year cycle: young queens leave the nest and "+
			"hibernate, the season ends, the queens emerge again. Both run with StopExtinct? = false. Compared "+
			"at every step: the gobees median must lie in the NetLogo 5-95 % band. The table shows day 515 (day 150 "+
			"of year 2).",
		runs, rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range failures {
		t.Error(f)
	}
}
