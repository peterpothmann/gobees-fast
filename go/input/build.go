package input

import (
	"fmt"
	"slices"

	"gobees/params"
	"gobees/util"
)

// daysPerYear is the number of foraging periods of one year in a run file [FPER].
const daysPerYear = 365

// Build returns the parameters of a run: params.Default with the run file c and the values derived from it
// [INIT-05]. landscapes holds the landscape files by their name in c, at least c.Landscape and, with a crop rotation,
// every file of c.CropRotation.
func Build(c Config, landscapes map[string]Landscape) (params.Params, error) {
	p := params.Default()
	p.File = c.Parameters
	p.Output = c.Output
	in := params.Input{}

	// [PAR-01], [QINI]
	species := make([]string, len(c.InitialQueens))
	for i, q := range c.InitialQueens {
		species[i] = q.Species
	}
	// [PAR-04]
	mortality, err := mortalityForager(c.Parameters.ForagingMortalityModel)
	if err != nil {
		return p, err
	}
	in.Derived = params.Derived{
		BeeSpeciesInitialQueensList: slices.Clone(c.InitialQueens),
		SpeciesList:                 species,
		MortalityForager_per_s:      mortality,
		FoodsourcesFile:             c.Landscape,                        // [PAR-05]
		CSS_OptionsList:             slices.Clone(c.StewardshipOptions), // [PAR-06]
		CropRotationList:            slices.Clone(c.CropRotation),       // [PAR-07]
	}

	// [FPER]
	if in.Weather, err = weather(c.ForagingPeriod_s); err != nil {
		return p, err
	}

	// [SPEC-01] to [SPEC-03], [SPEC-05], [SPEC-07]
	for _, s := range c.Species {
		if slices.Contains(species, s.Name) { // [SPEC-02] case-sensitive, in the order of the run file
			in.Species = append(in.Species, deriveSpecies(s, &p.Constants))
		}
	}

	// [FSRC-12], [LAYR-02]
	if len(c.Flowers) > 0 {
		t := params.FlowerSpeciesTable{}
		for _, f := range c.Flowers {
			t.Names = append(t.Names, f.Name)
			t.Rows = append(t.Rows, f)
		}
		in.Flowers = &t
	}

	// [FSRC-02] to [FSRC-04] the start and, if the crop rotation runs (more than one file [MORN-09]), every file of
	// the crop rotation [CROP-01]
	names := []string{c.Landscape}
	if len(c.CropRotation) > 1 {
		names = append(names, c.CropRotation...)
	}
	in.Foodsources = map[string]params.Foodsources{}
	for _, name := range names {
		if _, ok := in.Foodsources[name]; ok {
			continue
		}
		l, ok := landscapes[name]
		if !ok {
			return p, fmt.Errorf("landscape file %q not read", name)
		}
		in.Foodsources[name] = foodsources(l)
	}

	p.Input = in
	return p, nil
}

// mortalityForager returns the foraging mortality per second [PAR-04].
// Any other model name is an error. NetLogo gives no message and keeps the 0 of clear-all, so foragers and
// queens searching a nest never die (finding 5, see trans/netlogo-findings.md#finding-5). Fixed in step F5.
func mortalityForager(model string) (float64, error) {
	switch model {
	case "high":
		return 1.0e-05, nil // BEEHAVE value, Visscher & Dukas 1997
	case "intermediate":
		return 2.14e-06, nil // Schmid-Hempel & Heeb 1991
	case "low":
		return 2.75e-07, nil // Stelzer et al. 2010
	}
	return 0, fmt.Errorf("unknown ForagingMortalityModel %q (expected \"high\", \"intermediate\" or \"low\")", model)
}

// weather checks the foraging periods of the run file: at least one year, 365 values per year, each from 0 to
// 86400 s.
func weather(years []Year) (params.Weather, error) {
	if len(years) == 0 {
		return params.Weather{}, fmt.Errorf("foraging_period_s: at least one year is needed")
	}
	w := params.Weather{Years_s: make([][]float64, len(years))}
	for i, y := range years {
		if len(y) != daysPerYear {
			return w, fmt.Errorf("foraging_period_s: year %d has %d days, expected %d", i+1, len(y), daysPerYear)
		}
		for d, v := range y {
			if v < 0 || v > 86400 {
				return w, fmt.Errorf("foraging_period_s: year %d, day %d: %v s, expected 0 to 86400", i+1, d+1, v)
			}
		}
		w.Years_s[i] = slices.Clone(y)
	}
	return w, nil
}

// deriveSpecies returns the species of the run file with the derived values [SPEC-03], [SPEC-05], [SPEC-07].
// NestHabitatsList is not yet expanded [SPEC-04].
func deriveSpecies(s Species, constants *params.Constants) params.Species {
	// the values of the run file as they are (a copy, NestHabitatsList cloned), then the derived values (step F14)
	sp := params.Species{SpeciesFile: s}
	sp.NestHabitatsList = slices.Clone(s.NestHabitatsList)
	sp.EmergingDay_sd = util.Round(sp.EmergingDay_sd) // finding 20 (see trans/netlogo-findings.md#finding-20)
	sp.FlightCosts_kJ_per_m_per_mg = constants.MetabolicRateFlight_W_per_kg / sp.FlightVelocity_m_per_s / (1000 * 1000 * 1000)

	// [SPEC-05]
	sp.DevQuotaIncubationToday_kJ = 10 / (1.5 * float64(sp.Batchsize))
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
	sp.MinPollenStore_g = 0.5 * 0.001 * sp.DevWeightPupationMin_mg * float64(sp.Batchsize) / sp.PollenToBodymassFactor
	return sp
}

// foodsources returns the flower patches of a landscape file.
func foodsources(l Landscape) params.Foodsources {
	f := params.Foodsources{MetresPerPatch: l.MetresPerPatch, Patches: make([]params.FlowerPatch, len(l.Patches))}
	for i, p := range l.Patches {
		list := make([]params.FlowerEntry, len(p.FlowerSpeciesList))
		for j, e := range p.FlowerSpeciesList {
			list[j] = params.FlowerEntry(e)
		}
		f.Patches[i] = params.FlowerPatch{Id_Beescout: p.Id, PatchType: p.PatchType, X: p.X, Y: p.Y,
			Area_sqm: p.Area_sqm, FlowerSpeciesList: list}
	}
	return f
}
