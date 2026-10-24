package sys

import (
	"fmt"
	"math"
	"testing"

	"gobees/globals"
	"gobees/internal/validation"

	"github.com/mlange-42/ark/ecs"
)

func TestValidationStep4aAgainstNetLogo(t *testing.T) {
	// Experiment trans/validation/step4a_experiment.xml: Day, NectarAvailableTotalToday_l, PollenAvailableTotalToday_kg.
	// The values depend only on the landscape, so all NetLogo runs agree up to the last digits of the float sums
	// (random order of ask). The band is therefore widened by a relative tolerance of 1e-9.
	if testing.Short() {
		t.Skip("100 runs of 400 days")
	}
	reference, err := validation.Parse(validation.NetLogoStep4a)
	if err != nil {
		t.Fatal(err)
	}
	const runs, steps = 100, 400
	metrics := []string{"day", "nectar_l", "pollen_kg"}

	values := map[string][][]float64{} // values[metric][step][run]
	for _, m := range metrics {
		values[m] = make([][]float64, steps+1)
	}
	for run := 1; run <= runs; run++ {
		a := newExampleFarmSetup(t, float64(run), &CreateInitialQueens{}, &Tick{}, &UpdateMorning{}, &UpdateFoodsources{},
			&EmergenceNewQueens{}, &UpdateColonies{}, &CheckNumbers{SkipColonyCheck: true})
		cal := ecs.GetResource[globals.Calendar](a.World)
		today := ecs.GetResource[globals.FoodsourcesToday](a.World)
		for step := 0; step <= steps; step++ {
			if step > 0 {
				a.Update()
			}
			values["day"][step] = append(values["day"][step], cal.Day)
			values["nectar_l"][step] = append(values["nectar_l"][step], today.NectarAvailableTotalToday_l)
			values["pollen_kg"][step] = append(values["pollen_kg"][step], today.PollenAvailableTotalToday_kg)
		}
	}

	rows := []validation.ReportRow{}
	failures := []string{}
	for _, m := range metrics {
		outside := []string{}
		for step := 0; step <= steps; step++ {
			q, ok := reference[validation.Key{Step: step, Metric: m}]
			if !ok {
				t.Fatalf("no NetLogo reference for %s at step %d", m, step)
			}
			tol := 1e-9 * math.Max(1, math.Abs(q.Q50))
			median := validation.Median(values[m][step])
			if median < q.Q05-tol || median > q.Q95+tol {
				outside = append(outside, fmt.Sprintf("step %d: gobees median %.10g, NetLogo 5-95 %% [%.10g, %.10g]", step, median, q.Q05, q.Q95))
			}
		}
		final := reference[validation.Key{Step: steps, Metric: m}]
		gobees := validation.Median(values[m][steps])
		t.Logf("%-10s step %3d: gobees median %10.6g | NetLogo median %10.6g", m, steps, gobees, final.Q50)
		rows = append(rows, validation.ReportRow{Metric: m, Steps: steps + 1, Within: steps + 1 - len(outside),
			LastStep: steps, Gobees: gobees, NetLogo: final})
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: outside the NetLogo band at %d of %d steps, e.g. %s",
				m, len(outside), steps+1, outside[0]))
		}
	}

	err = validation.WriteReport("Step 4a: morning routine",
		"Experiment `trans/validation/step4a_experiment.xml`: example farm, 400 days. Day of the year, nectar and "+
			"pollen available in the landscape. These values depend only on the landscape, so the NetLogo band is "+
			"widened by a relative tolerance of 1e-9 (float sums in random order).",
		runs, rows)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range failures {
		t.Error(f)
	}
}
