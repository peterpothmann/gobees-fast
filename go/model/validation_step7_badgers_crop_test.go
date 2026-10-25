package model

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/params"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// Step 7: the parts of the year cycle that do not run on the example farm (badgers, crop rotation), each in
// its own experiment.

// step7BadgersMetrics returns the metrics of go/validation/experiments/step7_badgers.xml: those of step 7 and the
// number of badgers.
func step7BadgersMetrics(w *ecs.World) map[string]float64 {
	m := step7Metrics(w)
	q := ecs.NewFilter1[comp.Badger](w).Query()
	m["badgers"] = float64(q.Count())
	q.Close()
	return m
}

func TestValidationStep7BadgersAgainstNetLogo(t *testing.T) {
	// Experiment go/validation/experiments/step7_badgers.xml: as step 7, with 2 badgers. NetLogo reads a parameter
	// file with N_Badgers = 2 (otherwise the example farm file), gobees sets the value.
	compareWithNetLogo(t, netLogoComparison{
		reference: validation.NetLogoStep7Badgers,
		title:     "Step 7: badgers",
		description: "Experiment `go/validation/experiments/step7_badgers.xml`: as step 7 (example farm, 1095 days, " +
			"full model), with 2 badgers (N_Badgers = 2). They live on the two Scrub patches of the example farm and dig " +
			"up each colony within 735 m with p = 0.019 per day. Both run with StopExtinct? = false. Compared at " +
			"every step: the gobees median must lie in the NetLogo 5-95 % band. The table shows day 515 (day 150 of " +
			"year 2).",
		metrics:   append(slices.Clone(step7MetricNames), "badgers"),
		steps:     1095,
		shownStep: 515,
		change:    func(p *params.Params) { p.File.N_Badgers = 2 },
		measure:   step7BadgersMetrics,
	})
}

// step7CropMetrics returns the metrics of go/validation/experiments/step7_crop.xml: those of step 7 and the number
// of foodsources.
func step7CropMetrics(w *ecs.World) map[string]float64 {
	m := step7Metrics(w)
	q := ecs.NewFilter1[comp.Foodsource](w).Query()
	m["foodsources"] = float64(q.Count())
	q.Close()
	return m
}

func TestValidationStep7CropRotationAgainstNetLogo(t *testing.T) {
	// Experiment go/validation/experiments/step7_crop.xml: as step 7, with a crop rotation of two maps. NetLogo sets
	// CropRotationList after Setup, gobees sets CropRotationListAsString, which gives the same list [PAR-07].
	compareWithNetLogo(t, netLogoComparison{
		reference: validation.NetLogoStep7Crop,
		title:     "Step 7: crop rotation",
		description: "Experiment `go/validation/experiments/step7_crop.xml`: as step 7 (example farm, 1095 days, full " +
			"model), with a crop rotation of two maps. On Day 365 the landscape is replaced: year 2 runs on " +
			"`Example_Farm_PlotTest_Foodsources.txt` (36 flower patches), year 3 on the example farm again. Both run " +
			"with StopExtinct? = false. Compared at every step: the gobees median must lie in the NetLogo 5-95 % band. " +
			"The table shows day 515 (day 150 of year 2, on the second map).",
		metrics:   append(slices.Clone(step7MetricNames), "foodsources"),
		steps:     1095,
		shownStep: 515,
		change: func(p *params.Params) {
			p.File.CropRotationListAsString = "_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt"
		},
		measure: step7CropMetrics,
	})
}
