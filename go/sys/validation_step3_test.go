package sys

import (
	"fmt"
	"testing"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// step3Metrics returns the metrics of go/validation/experiments/step3.xml for the current state of the world,
// with the same definitions as in NetLogo. Initial queens are the queens with 0 < emergingDate < 365.
func step3Metrics(w *ecs.World) map[string]float64 {
	m := map[string]float64{}
	fs := ecs.NewFilter2[comp.Foodsource, comp.Layer](w).Query()
	for fs.Next() {
		_, l := fs.Get()
		m["foodsources"]++
		if l.Masterpatch {
			m["masterpatches"]++
		}
	}
	sp := ecs.NewFilter1[comp.Species](w).Query()
	for sp.Next() {
		if s := sp.Get(); s.Name == "B_terrestris" {
			m["nestsites"] = float64(len(s.NestsiteFoodsourceList))
			m["nestsite_area"] = s.NestSiteArea
		}
	}
	// TotalColoniesEverProduced: without UpdateColonies no colony dies, so the living colonies are all colonies
	// founded (the counter is output only since step F4, obs.Counters)
	cols := ecs.NewFilter1[comp.Colony](w).Query()
	m["colonies_founded"] = float64(cols.Count())
	cols.Close()

	n, sumWeight, sumEmerging := 0.0, 0.0, 0.0
	bees := ecs.NewFilter1[comp.Bee](w).Query()
	for bees.Next() {
		b := bees.Get()
		if b.Caste != caste.Queen || !b.EmergingDate.IsSet() || b.EmergingDate.Day() <= 0 || b.EmergingDate.Day() >= 365 {
			continue
		}
		n++
		sumWeight += b.Weight_mg
		sumEmerging += float64(b.EmergingDate.Day())
		if b.Activity == activity.Hibernate {
			m["queens_hibernating"]++
		} else if b.ColonyID.IsZero() {
			m["queens_searching"]++
		}
	}
	if n > 0 {
		m["mean_queen_weight"] = sumWeight / n
		m["mean_emerging_date"] = sumEmerging / n
	}
	return m
}

func TestValidationStep3AgainstNetLogo(t *testing.T) {
	if testing.Short() {
		t.Skip("100 runs of 250 days")
	}
	reference, err := validation.Parse(validation.NetLogoStep3)
	if err != nil {
		t.Fatal(err)
	}
	const runs, steps = 100, 250

	// gobees: values[metric][step][run]
	values := map[string][][]float64{}
	record := func(step int, m map[string]float64) {
		for _, name := range []string{"foodsources", "masterpatches", "nestsites", "nestsite_area", "colonies_founded",
			"queens_searching", "queens_hibernating", "mean_queen_weight", "mean_emerging_date"} {
			if values[name] == nil {
				values[name] = make([][]float64, steps+1)
			}
			values[name][step] = append(values[name][step], m[name])
		}
	}
	for run := 1; run <= runs; run++ {
		a := newExampleFarmSetup(t, float64(run), &CreateInitialQueens{}, &Tick{}, &EmergenceNewQueens{})
		record(0, step3Metrics(a.World))
		for step := 1; step <= steps; step++ {
			a.Update()
			record(step, step3Metrics(a.World))
		}
	}

	// Setup values and the queens' traits: only at step 0 (later, NetLogo colonies die with their queen).
	// Colony foundation and queen counts: at every step.
	checks := []struct {
		metric   string
		allSteps bool
	}{
		{"foodsources", false}, {"masterpatches", false}, {"nestsites", false}, {"nestsite_area", false},
		{"mean_queen_weight", false}, {"mean_emerging_date", false},
		{"colonies_founded", true}, {"queens_searching", true}, {"queens_hibernating", true},
	}
	failures := []string{}
	for _, c := range checks {
		last := 0
		if c.allSteps {
			last = steps
		}
		outside := []string{}
		for step := 0; step <= last; step++ {
			q, ok := reference[validation.Key{Step: step, Metric: c.metric}]
			if !ok {
				t.Fatalf("no NetLogo reference for %s at step %d", c.metric, step)
			}
			median := validation.Median(values[c.metric][step])
			if median < q.Q05 || median > q.Q95 {
				outside = append(outside, fmt.Sprintf("step %d: gobees median %.4g, NetLogo 5-95 %% [%.4g, %.4g]", step, median, q.Q05, q.Q95))
			}
		}
		final := reference[validation.Key{Step: last, Metric: c.metric}]
		gobees := validation.Median(values[c.metric][last])
		t.Logf("%-20s step %3d: gobees median %10.4g | NetLogo median %10.4g [%.4g, %.4g]",
			c.metric, last, gobees, final.Q50, final.Q05, final.Q95)
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: gobees median outside the NetLogo 5-95 %% band at %d of %d steps, e.g. %s",
				c.metric, len(outside), last+1, outside[0]))
		}
	}

	for _, f := range failures {
		t.Error(f)
	}
}
