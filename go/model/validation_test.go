package model

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"gobees/params"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// netLogoComparison describes the comparison of the full model with a NetLogo experiment on the example farm.
type netLogoComparison struct {
	reference   string // the embedded NetLogo reference, e.g. validation.NetLogoStep7
	order       string // sort key of the report section (validation.Section.Order), e.g. "07.1"
	step        string // heading of the step in the report, shared by the experiments of the step
	title       string // heading of the experiment below the step heading
	experiment  string // name of the experiment, e.g. "step7_badgers": file name of the plot
	description string // text of the report section
	metrics     []string
	steps       int                    // days per run
	shownStep   int                    // example day of the report
	parameters  string                 // parameter file in validation.ParameterFiles instead of the example farm one, may be empty
	change      func(p *params.Params) // changes the example farm parameters as the experiment does, may be nil
	measure     func(w *ecs.World) map[string]float64
}

// compareWithNetLogo runs model.Default 100 times (seeds 1 to 100) with StopExtinct? = false, as the experiments do
// [GO-20], and requires the gobees median of every metric to lie within the NetLogo 5-95 % band on every day.
// It logs the result and writes the report (validation.WriteReport).
func compareWithNetLogo(t *testing.T, c netLogoComparison) {
	t.Helper()
	if testing.Short() {
		t.Skipf("100 runs of %d days", c.steps)
	}
	reference, err := validation.Parse(c.reference)
	if err != nil {
		t.Fatal(err)
	}
	const runs = 100
	var file params.ParameterFile
	if c.parameters != "" {
		b, err := validation.ParameterFiles.ReadFile(c.parameters)
		if err != nil {
			t.Fatal(err)
		}
		if file, err = params.ParseParameterFile(strings.NewReader(string(b))); err != nil {
			t.Fatal(err)
		}
	}

	values := map[string][][]float64{} // values[metric][step][run]
	for _, name := range c.metrics {
		values[name] = make([][]float64, c.steps+1)
	}
	record := func(step int, m map[string]float64) {
		for _, name := range c.metrics {
			values[name][step] = append(values[name][step], m[name])
		}
	}
	for run := 1; run <= runs; run++ {
		p := exampleFarm(t, float64(run))
		if c.parameters != "" {
			p.File = file
		}
		p.Termination.MaxTicks = c.steps
		p.File.StopExtinct = false // as the experiments [GO-20]
		if c.change != nil {
			c.change(&p)
		}
		a := Default(p, nil)
		a.Initialize()
		record(0, c.measure(a.World))
		for step := 1; step <= c.steps; step++ {
			a.Update()
			record(step, c.measure(a.World))
		}
	}

	rows := []validation.ReportRow{}
	failures := []string{}
	for _, metric := range c.metrics {
		outside := []string{}
		worst, worstStep, worstGobees := 0.0, 0, 0.0 // largest |median difference| / width of the NetLogo band
		medians, bands := []float64{}, []validation.Quantiles{}
		for step := 0; step <= c.steps; step++ {
			q, ok := reference[validation.Key{Step: step, Metric: metric}]
			if !ok {
				t.Fatalf("no NetLogo reference for %s at step %d", metric, step)
			}
			// a relative tolerance of 1e-9 for the float sums (stores, landscape), whose last digits depend on the order
			tol := 1e-9 * math.Max(1, math.Abs(q.Q50))
			median := validation.Median(values[metric][step])
			medians, bands = append(medians, median), append(bands, q)
			if width := q.Q95 - q.Q05; width > tol && math.Abs(median-q.Q50)/width > worst {
				worst, worstStep, worstGobees = math.Abs(median-q.Q50)/width, step, median
			}
			if median < q.Q05-tol || median > q.Q95+tol {
				outside = append(outside, fmt.Sprintf("step %d: gobees median %.6g, NetLogo 5-95 %% [%.6g, %.6g]", step, median, q.Q05, q.Q95))
			}
		}
		t.Logf("%-18s %4d of %4d steps within the band", metric, c.steps+1-len(outside), c.steps+1)
		mid := reference[validation.Key{Step: c.shownStep, Metric: metric}]
		t.Logf("%-18s day %d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g]",
			metric, c.shownStep, validation.Median(values[metric][c.shownStep]), mid.Q50, mid.Q05, mid.Q95)
		w := reference[validation.Key{Step: worstStep, Metric: metric}]
		t.Logf("%-18s largest deviation on day %4d: gobees median %.5g, NetLogo median %.5g [%.5g, %.5g] (%.2f of the band width)",
			metric, worstStep, worstGobees, w.Q50, w.Q05, w.Q95, worst)
		rows = append(rows, validation.ReportRow{Metric: metric, Steps: c.steps + 1, Within: c.steps + 1 - len(outside),
			Step: c.shownStep, Gobees: validation.Median(values[metric][c.shownStep]), NetLogo: mid,
			Medians: medians, Bands: bands})
		if len(outside) > 0 {
			failures = append(failures, fmt.Sprintf("%s: gobees median outside the NetLogo 5-95 %% band at %d of %d steps, e.g. %s",
				metric, len(outside), c.steps+1, outside[0]))
		}
	}

	if err := validation.WriteReport(validation.Section{Order: c.order, Step: c.step, Title: c.title, Experiment: c.experiment,
		Description: c.description, Runs: runs, Rows: rows, Values: values}); err != nil {
		t.Fatal(err)
	}
	for _, f := range failures {
		t.Error(f)
	}
}
