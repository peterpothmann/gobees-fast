package input

import (
	"fmt"
	"math"
	"slices"

	"gobees/params"
)

// daysPerYear is the number of foraging periods of one year in a run file.
const daysPerYear = 365

// Build returns the parameters of a run: params.Default with the run file c and the values derived from it. The
// run file needs a name (CheckName). The section run (needed, checkRun) sets the seed and the number of days.
// landscapes holds the landscape files by their name in c, at least c.Landscape and, with a crop rotation,
// every file of c.CropRotation.
func Build(c Config, landscapes map[string]Landscape) (params.Params, error) {
	p := params.Default()
	if err := CheckName(c.Name); err != nil {
		return p, err
	}
	if err := checkRun(c.Run); err != nil {
		return p, err
	}
	p.Seed = c.Run.Seed
	if c.Run.Days != nil {
		p.Termination.MaxTicks = *c.Run.Days
	}
	p.Parameters = c.Parameters
	p.Output = c.Output
	p.Constants = c.Constants
	in := params.Input{}

	species := make([]string, len(c.InitialQueens))
	for i, q := range c.InitialQueens {
		species[i] = q.Species
	}
	mortality, err := mortalityForager(c.Parameters.ForagingMortalityModel, &c.Constants.Foraging.Mortality)
	if err != nil {
		return p, err
	}
	in.Derived = params.Derived{
		InitialQueens:      slices.Clone(c.InitialQueens),
		SpeciesNames:       species,
		ForagerMortality:   mortality,
		StartLandscape:     c.Landscape,
		StewardshipOptions: slices.Clone(c.StewardshipOptions),
		CropRotation:       slices.Clone(c.CropRotation),
	}

	if in.ForagingPeriod, err = weather(c.ForagingPeriod); err != nil {
		return p, err
	}

	for _, s := range c.Species {
		if slices.Contains(species, s.Name) { // case-sensitive, in the order of the run file
			in.Species = append(in.Species, deriveSpecies(s, &p.Constants))
		}
	}

	if len(c.Flowers) > 0 {
		t := params.FlowerSpeciesTable{}
		for _, f := range c.Flowers {
			t.Names = append(t.Names, f.Name)
			t.Rows = append(t.Rows, f)
		}
		in.Flowers = &t
	}

	// the start and, if the crop rotation runs (more than one file), every file of the crop rotation
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

// mortalityForager returns the foraging mortality per second of the model "high", "intermediate" or "low".
// Any other model name is an error.
func mortalityForager(model string, m *params.ForagingMortality) (float64, error) {
	switch model {
	case "high":
		return m.High, nil
	case "intermediate":
		return m.Intermediate, nil
	case "low":
		return m.Low, nil
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

// deriveSpecies returns the species of the run file with the derived values.
// NestHabitats is not yet expanded.
func deriveSpecies(s Species, constants *params.Constants) params.Species {
	// the values of the run file as they are (a copy, NestHabitats cloned), then the derived values
	sp := params.Species{SpeciesFile: s}
	sp.NestHabitats = slices.Clone(s.NestHabitats)
	sp.EmergenceDaySD = math.Floor(sp.EmergenceDaySD + 0.5) // rounded to a whole day, .5 up
	sp.FlightCosts = constants.Foraging.FlightMetabolicRate / sp.FlightVelocity / (1000 * 1000 * 1000)

	sp.DailyIncubationQuota = constants.Colony.DailyIncubation / (constants.Colony.IncubationBatches * float64(sp.BatchSize))
	sp.HatchingAgeMax = sp.HatchingAgeMin * sp.MinToMaxFactor
	sp.PupationAgeMax = sp.PupationAgeMin*sp.MinToMaxFactor - sp.HatchingAgeMin
	sp.EmergenceAgeMax = sp.EmergenceAgeMin*sp.MinToMaxFactor - sp.PupationAgeMin
	sp.HatchingIncubation = sp.DailyIncubationQuota * sp.HatchingAgeMin
	sp.PupationIncubation = sp.DailyIncubationQuota * sp.PupationAgeMin
	sp.EmergenceIncubation = sp.DailyIncubationQuota * sp.EmergenceAgeMin
	sp.QueenDeterminationAge = constants.Colony.QueenDeterminationAge
	sp.QueenPupationAgeMax = sp.QueenPupationAgeMin*sp.MinToMaxFactor - sp.HatchingAgeMin
	sp.QueenEmergenceAgeMax = sp.QueenEmergenceAgeMin*sp.MinToMaxFactor - sp.QueenPupationAgeMin
	sp.QueenPupationIncubation = sp.DailyIncubationQuota * sp.QueenPupationAgeMin
	sp.QueenEmergenceIncubation = sp.DailyIncubationQuota * sp.QueenEmergenceAgeMin

	sp.MinPollenStore = constants.Colony.MinPollenStoreBatches * 0.001 * sp.PupationWeightMin * float64(sp.BatchSize) / sp.PollenToBodymassFactor
	return sp
}

// foodsources returns the flower patches of a landscape file.
func foodsources(l Landscape) params.Foodsources {
	f := params.Foodsources{MetresPerPatch: l.MetresPerPatch, Patches: make([]params.FlowerPatch, len(l.Patches))}
	for i, p := range l.Patches {
		f.Patches[i] = params.FlowerPatch{ID: p.ID, PatchType: p.PatchType, X: p.X, Y: p.Y,
			Area: p.Area, Flowers: slices.Clone(p.Flowers)}
	}
	return f
}
