package input

import (
	"fmt"
	"io"
	"os"
	"slices"

	"gobees/comp"
	"gobees/params"
	"gobees/util"
)

// parseSpeciesFile reads the rows of the BeespeciesFile whose first column ("species") is in speciesList [SPEC-02],
// in FILE order, and sets the derived values [SPEC-03], [SPEC-05], [SPEC-07].
// nestHabitatsList is returned as read from the file, not yet expanded.
func parseSpeciesFile(r io.Reader, speciesList []string, constants *params.Constants) ([]comp.Species, error) {
	t, err := readCSVTable(r)
	if err != nil {
		return nil, err
	}
	result := []comp.Species{}
	for _, row := range t.rows {
		if !slices.Contains(speciesList, row[0]) { // case-sensitive
			continue
		}
		sp := comp.Species{}
		if sp.Name, err = t.cell(row, "name"); err != nil {
			return nil, err
		}
		// [SPEC-03]
		fields := []struct {
			column string
			value  *float64
		}{
			{"maxLifespanWorkers", &sp.MaxLifespanWorkers},
			{"emergingDay_mean", &sp.EmergingDay_mean},
			{"emergingDay_sd", &sp.EmergingDay_sd},
			{"batchsize", &sp.Batchsize},
			{"flightVelocity_m/s", &sp.FlightVelocity_m_per_s},
			{"searchLength_m", &sp.SearchLength_m},
			{"seasonStop", &sp.SeasonStop},
			{"timeUnloading", &sp.TimeUnloading},
			{"specMax_cropVolume_myl", &sp.SpecMax_cropVolume_myl},
			{"specMax_pollenPellets_g", &sp.SpecMax_pollenPellets_g},
			{"minToMaxFactor", &sp.MinToMaxFactor},
			{"devAgeHatchingMin_d", &sp.DevAgeHatchingMin_d},
			{"devAgePupationMin_d", &sp.DevAgePupationMin_d},
			{"devAgeEmergingMin_d", &sp.DevAgeEmergingMin_d},
			{"devWeightEgg_mg", &sp.DevWeightEgg_mg},
			{"devWeightPupationMin_mg", &sp.DevWeightPupationMin_mg},
			{"devWeightPupationMax_mg", &sp.DevWeightPupationMax_mg},
			{"pollenToBodymassFactor", &sp.PollenToBodymassFactor},
			{"dev_Q_DeterminationWeight_mg", &sp.Dev_Q_DeterminationWeight_mg},
			{"devAge_Q_PupationMin_d", &sp.DevAge_Q_PupationMin_d},
			{"devWeight_Q_PupationMin_mg", &sp.DevWeight_Q_PupationMin_mg},
			{"devWeight_Q_PupationMax_mg", &sp.DevWeight_Q_PupationMax_mg},
			{"devAge_Q_EmergingMin_d", &sp.DevAge_Q_EmergingMin_d},
			{"growthFactor", &sp.GrowthFactor},
			{"proboscis_min_mm", &sp.Proboscis_min_mm},
			{"proboscis_max_mm", &sp.Proboscis_max_mm},
			{"dailyNestSiteChance", &sp.ChanceFindNest},
		}
		for _, f := range fields {
			if *f.value, err = t.number(row, f.column); err != nil {
				return nil, err
			}
		}
		sp.EmergingDay_sd = util.Round(sp.EmergingDay_sd) // finding 20 (see trans/netlogo-findings.md#finding-20)
		sp.FlightCosts_kJ_per_m_per_mg = constants.MetabolicRateFlight_W_per_kg / sp.FlightVelocity_m_per_s / (1000 * 1000 * 1000)

		// [SPEC-04] list literal from the file, expanded later
		cell, err := t.cell(row, "nestHabitatsList")
		if err != nil {
			return nil, err
		}
		if sp.NestHabitatsList, err = stringList(cell); err != nil {
			return nil, fmt.Errorf("%s: nestHabitatsList: %w", row[0], err)
		}

		// [SPEC-05]
		sp.DevQuotaIncubationToday_kJ = 10 / (1.5 * sp.Batchsize)
		sp.DevAgeHatchingMax_d = sp.DevAgeHatchingMin_d * sp.MinToMaxFactor
		sp.DevAgePupationMax_d = sp.DevAgePupationMin_d*sp.MinToMaxFactor - sp.DevAgeHatchingMin_d
		sp.DevAgeEmergingMax_d = sp.DevAgeEmergingMin_d*sp.MinToMaxFactor - sp.DevAgePupationMin_d
		sp.DevIncubationHatchingTH_kJ = sp.DevQuotaIncubationToday_kJ * sp.DevAgeHatchingMin_d
		sp.DevIncubationPupationTH_kJ = sp.DevQuotaIncubationToday_kJ * sp.DevAgePupationMin_d
		sp.DevIncubationEmergingTH_kJ = sp.DevQuotaIncubationToday_kJ * sp.DevAgeEmergingMin_d
		sp.Dev_larvalAge_QueenDetermination_d = 3
		sp.DevAge_Q_PupationMax_d = sp.DevAge_Q_PupationMin_d*sp.MinToMaxFactor - sp.DevAgeHatchingMin_d
		sp.DevAge_Q_EmergingMax_d = sp.DevAge_Q_EmergingMin_d*sp.MinToMaxFactor - sp.DevAge_Q_PupationMin_d
		sp.DevIncubation_Q_PupationTH_kJ = sp.DevQuotaIncubationToday_kJ * sp.DevAge_Q_PupationMin_d
		sp.DevIncubation_Q_EmergingTH_kJ = sp.DevQuotaIncubationToday_kJ * sp.DevAge_Q_EmergingMin_d

		// [SPEC-07]
		sp.MinPollenStore_g = 0.5 * 0.001 * sp.DevWeightPupationMin_mg * sp.Batchsize / sp.PollenToBodymassFactor

		result = append(result, sp)
	}
	return result, nil
}

// stringList reads a NetLogo list literal of strings, e.g. ["Grassland*" "Garden"]. NetLogo: read-from-string.
func stringList(s string) ([]string, error) {
	v, err := ReadFromString(s)
	if err != nil {
		return nil, err
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected a list")
	}
	result := make([]string, len(items))
	for i, item := range items {
		if result[i], ok = item.(string); !ok {
			return nil, fmt.Errorf("expected strings in the list")
		}
	}
	return result, nil
}

// readSpeciesFile reads the BeespeciesFile (parseSpeciesFile) [SPEC-01].
func readSpeciesFile(path string, speciesList []string, constants *params.Constants) ([]comp.Species, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("BeespeciesFile: %w", err)
	}
	defer f.Close()
	species, err := parseSpeciesFile(f, speciesList, constants)
	if err != nil {
		return nil, fmt.Errorf("BeespeciesFile %s: %w", path, err)
	}
	return species, nil
}
