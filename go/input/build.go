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
	p.Parameters = c.Parameters
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
		InitialQueens:      slices.Clone(c.InitialQueens),
		SpeciesNames:       species,
		ForagerMortality:   mortality,
		StartLandscape:     c.Landscape,                        // [PAR-05]
		StewardshipOptions: slices.Clone(c.StewardshipOptions), // [PAR-06]
		CropRotation:       slices.Clone(c.CropRotation),       // [PAR-07]
	}

	// [FPER]
	if in.ForagingPeriod, err = weather(c.ForagingPeriod); err != nil {
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
func weather(years []Year) (params.ForagingPeriod, error) {
	if len(years) == 0 {
		return params.ForagingPeriod{}, fmt.Errorf("foragingPeriod: at least one year is needed")
	}
	w := params.ForagingPeriod{Years: make([][]float64, len(years))}
	for i, y := range years {
		if len(y) != daysPerYear {
			return w, fmt.Errorf("foragingPeriod: year %d has %d days, expected %d", i+1, len(y), daysPerYear)
		}
		for d, v := range y {
			if v < 0 || v > 86400 {
				return w, fmt.Errorf("foragingPeriod: year %d, day %d: %v s, expected 0 to 86400", i+1, d+1, v)
			}
		}
		w.Years[i] = slices.Clone(y)
	}
	return w, nil
}

// deriveSpecies returns the species of the run file with the derived values [SPEC-03], [SPEC-05], [SPEC-07].
// NestHabitats is not yet expanded [SPEC-04].
func deriveSpecies(s Species, constants *params.Constants) params.Species {
	// the values of the run file as they are (a copy, NestHabitats cloned), then the derived values (step F14)
	sp := params.Species{SpeciesFile: s}
	sp.NestHabitats = slices.Clone(s.NestHabitats)
	sp.EmergenceDaySD = util.Round(sp.EmergenceDaySD) // finding 20 (see trans/netlogo-findings.md#finding-20)
	sp.FlightCosts = constants.FlightMetabolicRate / sp.FlightVelocity / (1000 * 1000 * 1000)

	// [SPEC-05]
	sp.DailyIncubationQuota = 10 / (1.5 * float64(sp.BatchSize))
	sp.HatchingAgeMax = sp.HatchingAgeMin * sp.MinToMaxFactor
	sp.PupationAgeMax = sp.PupationAgeMin*sp.MinToMaxFactor - sp.HatchingAgeMin
	sp.EmergenceAgeMax = sp.EmergenceAgeMin*sp.MinToMaxFactor - sp.PupationAgeMin
	sp.HatchingIncubation = sp.DailyIncubationQuota * sp.HatchingAgeMin
	sp.PupationIncubation = sp.DailyIncubationQuota * sp.PupationAgeMin
	sp.EmergenceIncubation = sp.DailyIncubationQuota * sp.EmergenceAgeMin
	sp.QueenDeterminationAge = 3
	sp.QueenPupationAgeMax = sp.QueenPupationAgeMin*sp.MinToMaxFactor - sp.HatchingAgeMin
	sp.QueenEmergenceAgeMax = sp.QueenEmergenceAgeMin*sp.MinToMaxFactor - sp.QueenPupationAgeMin
	sp.QueenPupationIncubation = sp.DailyIncubationQuota * sp.QueenPupationAgeMin
	sp.QueenEmergenceIncubation = sp.DailyIncubationQuota * sp.QueenEmergenceAgeMin

	// [SPEC-07]
	sp.MinPollenStore = 0.5 * 0.001 * sp.PupationWeightMin * float64(sp.BatchSize) / sp.PollenToBodymassFactor
	return sp
}

// foodsources returns the flower patches of a landscape file.
func foodsources(l Landscape) params.Foodsources {
	f := params.Foodsources{MetresPerPatch: l.MetresPerPatch, Patches: make([]params.FlowerPatch, len(l.Patches))}
	for i, p := range l.Patches {
		list := make([]params.FlowerEntry, len(p.Flowers))
		for j, e := range p.Flowers {
			list[j] = params.FlowerEntry(e)
		}
		f.Patches[i] = params.FlowerPatch{ID: p.ID, PatchType: p.PatchType, X: p.X, Y: p.Y,
			Area: p.Area, Flowers: list}
	}
	return f
}
