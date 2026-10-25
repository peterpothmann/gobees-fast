package model

import (
	"slices"
	"testing"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/validation"

	"github.com/mlange-42/ark/ecs"
)

// Step 8: the model branches that the example farm does not use. Each experiment changes a few parameters of the
// example farm (go/validation/experiments/step8_*_Parameters.csv, written by scripts/parameter_files.py) and runs
// the full model for 2 years. The metrics are those of step 7 plus a few per experiment.

// step8Comparison returns the comparison of the experiment go/validation/experiments/<name>.xml.
func step8Comparison(order, name, reference, title, description string, extra []string,
	measure func(w *ecs.World, m map[string]float64)) netLogoComparison {
	return netLogoComparison{
		reference: reference,
		order:     order,
		step:      "Step 8: model branches the example farm does not use",
		title:     title,
		description: "Experiment `go/validation/experiments/" + name + ".xml`: example farm, full model, 730 days, " +
			"with the parameter file `" + name + "_Parameters.csv`: " + description + " Both run with StopExtinct? " +
			"= false. The table shows day 150.",
		metrics:    append(slices.Clone(step7MetricNames), extra...),
		steps:      730,
		shownStep:  150,
		parameters: "experiments/" + name + "_Parameters.csv",
		measure: func(w *ecs.World) map[string]float64 {
			m := step7Metrics(w)
			measure(w, m)
			return m
		},
	}
}

func TestValidationStep8IBMAgainstNetLogo(t *testing.T) {
	compareWithNetLogo(t, step8Comparison("08.1", "step8_ibm", validation.NetLogoStep8IBM,
		"Individual-based colonies",
		"COLONIES_IBM = 10, the first 10 colonies are individual-based [CCOL-05].",
		[]string{"ibm_colonies", "bee_agents"},
		func(w *ecs.World, m map[string]float64) {
			q := ecs.NewFilter1[comp.Colony](w).Query()
			for q.Next() {
				if !q.Get().CohortBased {
					m["ibm_colonies"]++
				}
			}
			bees := ecs.NewFilter1[comp.Bee](w).Query()
			m["bee_agents"] = float64(bees.Count())
			bees.Close()
		}))
}

func TestValidationStep8GeneticsAgainstNetLogo(t *testing.T) {
	compareWithNetLogo(t, step8Comparison("08.2", "step8_genetics", validation.NetLogoStep8Genetics,
		"Genetics",
		"SexLocus? = true (diploid males), UnlimitedMales? = false (a queen without a male dies).",
		[]string{"diploid_males"},
		func(w *ecs.World, m map[string]float64) {
			q := ecs.NewFilter1[comp.Bee](w).Query()
			for q.Next() {
				if b := q.Get(); b.Caste == caste.Male && b.Ploidy == 2 {
					m["diploid_males"] += b.Number
				}
			}
		}))
}

func TestValidationStep8SpeciesAgainstNetLogo(t *testing.T) {
	species := []string{"B_terrestris", "B_pascuorum", "B_lapidarius", "B_hortorum", "B_pratorum", "B_hypnorum"}
	extra := []string{}
	for _, s := range species {
		extra = append(extra, "colonies_"+s)
	}
	compareWithNetLogo(t, step8Comparison("08.3", "step8_species", validation.NetLogoStep8Species,
		"Several species",
		"all 7 species, 150 queens of each Bombus species and 100 Psithyrus queens.",
		append(extra, "psithyrus"),
		func(w *ecs.World, m map[string]float64) {
			for _, s := range species {
				m["colonies_"+s] = 0
			}
			q := ecs.NewFilter1[comp.Colony](w).Query()
			for q.Next() {
				m["colonies_"+q.Get().SpeciesNameColony]++
			}
			bees := ecs.NewFilter1[comp.Bee](w).Query()
			for bees.Next() {
				if b := bees.Get(); b.SpeciesName == "Psithyrus" {
					m["psithyrus"] += b.Number
				}
			}
		}))
}

func TestValidationStep8WeatherAgainstNetLogo(t *testing.T) {
	compareWithNetLogo(t, step8Comparison("08.4", "step8_weather", validation.NetLogoStep8Weather,
		"Weather and foraging mortality",
		`Weather = "foragingHoursExample" (foraging hours per day of the year), ForagingMortalityModel = "low".`,
		[]string{"foraging_period_s"},
		func(w *ecs.World, m map[string]float64) {
			m["foraging_period_s"] = ecs.GetResource[globals.Calendar](w).DailyForagingPeriod_s
		}))
}

func TestValidationStep8SettingsAgainstNetLogo(t *testing.T) {
	compareWithNetLogo(t, step8Comparison("08.5", "step8_settings", validation.NetLogoStep8Settings,
		"Other settings",
		"WinterMortality?, MergeHedges?, MinSizeFoodSources?, RemoveEmptyFoodSources?, KeepDeadColonies? = false, "+
			"MaxHibernatingQueens = 150 (the limit takes effect on every 1 January, also on the first day).",
		[]string{"foodsources", "queens_removed"},
		func(w *ecs.World, m map[string]float64) {
			q := ecs.NewFilter1[comp.Foodsource](w).Query()
			m["foodsources"] = float64(q.Count())
			q.Close()
			m["queens_removed"] = ecs.GetResource[globals.Counters](w).TotalHibernatingQueensEverRemoved
		}))
}
