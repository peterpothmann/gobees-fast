package sys

import (
	"fmt"
	"math"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/internal/check"
	"gobees/obs"
	"gobees/params"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// step6MetricNames are the metrics of go/validation/experiments/step6.xml, in the order of go/validation/scripts/aggregate_netlogo.py.
var step6MetricNames = []string{"colonies", "dead_colonies", "eggs", "larvae", "pupae", "workers", "males",
	"young_queens", "energy_kJ", "pollen_g", "trips", "switch_point", "queen_production", "nectar_left_l",
	"pollen_left_kg"}

// step6Metrics returns the metrics of go/validation/experiments/step6.xml for the current state of the world and
// the counters of the run, with the same definitions as the NetLogo reporters in the experiment file.
func step6Metrics(w *ecs.World, counters *obs.Counters) map[string]float64 {
	m := map[string]float64{}
	cols := ecs.NewFilter2[comp.Colony, comp.ColonyStores](w).Query()
	for cols.Next() {
		c, s := cols.Get()
		m["colonies"]++
		m["energy_kJ"] += s.EnergyStore_kJ
		m["pollen_g"] += s.PollenStore_g
		if c.SwitchPointDate < params.NotSetHigh {
			m["switch_point"]++
		}
		if c.QueenProduction {
			m["queen_production"]++
		}
	}
	m["dead_colonies"] = float64(counters.DeadColonies)

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
	}
	m["trips"] = counters.TotalForagingTripsToday

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

func TestValidationStep6AgainstNetLogo(t *testing.T) {
	// Experiment go/validation/experiments/step6.xml: the colony dynamics of year 1 with brood, the daily activity
	// queue and foraging (steps 4b to 6). NetLogo runs Go without QueensLeavingNestProc and UpdateSeasonalEventsProc
	// (step 7), and so does gobees: the systems below are model/default.go without the step 7 stubs.
	if testing.Short() {
		t.Skip("100 runs of 364 days")
	}
	reference, err := validation.Parse(validation.NetLogoStep6)
	if err != nil {
		t.Fatal(err)
	}
	const runs, steps = 100, 364
	const shownStep = 150

	values := map[string][][]float64{} // values[metric][step][run]
	for _, name := range step6MetricNames {
		values[name] = make([][]float64, steps+1)
	}
	record := func(step int, m map[string]float64) {
		for _, name := range step6MetricNames {
			values[name][step] = append(values[name][step], m[name])
		}
	}
	for run := 1; run <= runs; run++ {
		counters := &obs.Counters{} // after the model systems
		a := newExampleFarmSetup(t, float64(run), &CreateInitialQueens{}, &Tick{}, &UpdateMorning{}, &UpdateFoodsources{},
			&EmergenceNewQueens{}, &WakeUpBees{}, &UpdateColonies{}, &CheckNumbers{},
			&NeedNectarPollenLarvaeToday{}, &ActivityLoop{}, &EndOfDayResting{}, &FeedLarvae{}, &QueenProductionDate{},
			&Development{}, &MortalityBrood{}, counters)
		record(0, step6Metrics(a.World, counters))
		for step := 1; step <= steps; step++ {
			a.Update()
			if err := check.Day(a.World, false); err != nil {
				t.Fatalf("run %d, step %d: %v", run, step, err)
			}
			record(step, step6Metrics(a.World, counters))
		}
	}

	failures := []string{}
	for _, metric := range step6MetricNames {
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
		t.Logf("%-17s %3d of %3d steps within the band", metric, steps+1-len(outside), steps+1)
		mid := reference[validation.Key{Step: shownStep, Metric: metric}]
		t.Logf("%-17s day %d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g]",
			metric, shownStep, validation.Median(values[metric][shownStep]), mid.Q50, mid.Q05, mid.Q95)
		w := reference[validation.Key{Step: worstStep, Metric: metric}]
		t.Logf("%-17s largest deviation on day %3d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g] (%.2f of the band width)",
			metric, worstStep, worstGobees, w.Q50, w.Q05, w.Q95, worst)
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: gobees median outside the NetLogo 5-95 %% band at %d of %d steps, e.g. %s",
				metric, len(outside), steps+1, outside[0]))
		}
	}

	for _, f := range failures {
		t.Error(f)
	}
}
