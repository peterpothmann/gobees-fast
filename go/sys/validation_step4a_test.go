package sys

import (
	"fmt"
	"math"
	"testing"

	"gobees/comp"
	"gobees/globals"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

func TestValidationStep4aAgainstNetLogo(t *testing.T) {
	// Experiment go/validation/experiments/step4a.xml: Day, NectarAvailableTotalToday_l, PollenAvailableTotalToday_kg.
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
		today := &foodsourceTotals{}
		a := newExampleFarmSetup(t, float64(run), &initialQueensSetup{}, &tick{}, &calendarUpdate{}, &flowerRefill{},
			today, &queenEmergence{}, &colonyUpdate{}, &broodTotals{})
		cal := ecs.GetResource[globals.Calendar](a.World)
		for step := 0; step <= steps; step++ {
			if step > 0 {
				a.Update()
			}
			values["day"][step] = append(values["day"][step], float64(cal.Day))
			values["nectar_l"][step] = append(values["nectar_l"][step], today.nectar_l)
			values["pollen_kg"][step] = append(values["pollen_kg"][step], today.pollen_kg)
		}
	}

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
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: outside the NetLogo band at %d of %d steps, e.g. %s",
				m, len(outside), steps+1, outside[0]))
		}
	}

	for _, f := range failures {
		t.Error(f)
	}
}

// foodsourceTotals sums the nectar [l] and the pollen [kg] of all foodsources, as UpdateFoodsourcesProc sets
// NectarAvailableTotalToday_l and PollenAvailableTotalToday_kg [UFS-01], [UFS-02]. Output only, not in the model
// since step F4. Added right after flowerRefill: the values of the morning, before foraging.
type foodsourceTotals struct {
	nectar_l, pollen_kg float64
	filter              *ecs.Filter1[comp.FoodsourceState]
}

func (s *foodsourceTotals) Initialize(w *ecs.World) {
	s.filter = ecs.NewFilter1[comp.FoodsourceState](w)
	s.Update(w) // flowerRefill runs in Setup too [INIT-10]
}

func (s *foodsourceTotals) Update(w *ecs.World) {
	s.nectar_l, s.pollen_kg = 0, 0
	query := s.filter.Query()
	for query.Next() {
		st := query.Get()
		s.nectar_l += st.Nectar / (1000 * 1000)
		s.pollen_kg += st.Pollen / 1000
	}
}

func (s *foodsourceTotals) Finalize(w *ecs.World) {}
