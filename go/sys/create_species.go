package sys

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CreateSpecies creates one species entity per simulated species [INIT-07].
// Runs after CreateFoodsources, because the nest sites are foodsources.
//
// NetLogo: CreateSpeciesProc, NestHabitatsListREP.
type CreateSpecies struct{}

func (s *CreateSpecies) Initialize(w *ecs.World) {
	file := ecs.GetResource[params.ParameterFile](w)
	dir := ecs.GetResource[params.WorkingDirectory](w)
	constants := ecs.GetResource[params.Constants](w)
	derived := ecs.GetResource[globals.Derived](w)
	who := ecs.GetResource[globals.WhoCounter](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// [SPEC-01] Read the species parameters from the BeespeciesFile, in FILE order, and create one species entity per species in speciesList.
	path := filepath.Join(dir.Path, file.BeespeciesFile)
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("BeespeciesFile: %w", err))
	}
	defer f.Close()
	species, err := parseSpeciesFile(f, derived.SpeciesList, constants)
	if err != nil {
		panic(fmt.Errorf("BeespeciesFile %s: %w", path, err))
	}

	// Collect all masterpatches, sorted by who, to expand the nest habitats list and find the nest sites.
	masterpatches := collectMasterpatches(w)
	mapper := ecs.NewMap2[comp.Who, comp.Species](w)
	for i := range species {
		sp := &species[i]
		// [SPEC-04]
		expanded, err := NestHabitatsListREP(sp.NestHabitatsList, masterpatches, rng)
		if err != nil {
			panic(err)
		}
		sp.NestHabitatsList = expanded
		// [SPEC-06]
		sp.NestsiteFoodsourceList, sp.NestSiteArea = nestSites(sp.NestHabitatsList, masterpatches)
		// [SPEC-08] message only
		if len(sp.NestsiteFoodsourceList) == 0 && sp.Name != "Psithyrus" {
			log.Printf("%s has no suitable nesting foodsources. No colonies will form", sp.Name)
		}
		mapper.NewEntity(&comp.Who{Who: who.Next()}, sp)
	}
}

func (s *CreateSpecies) Update(w *ecs.World) {}

func (s *CreateSpecies) Finalize(w *ecs.World) {}

// masterpatch is a masterpatch foodsource with the values NestHabitatsListREP and the nest sites need.
type masterpatch struct {
	entity    ecs.Entity
	who       int64
	patchType string
	area_sqm  float64
}

// collectMasterpatches returns all masterpatches, sorted by who.
func collectMasterpatches(w *ecs.World) []masterpatch {
	result := []masterpatch{}
	query := ecs.NewFilter3[comp.Who, comp.Foodsource, comp.Layer](w).Query()
	for query.Next() {
		who, fs, layer := query.Get()
		if layer.Masterpatch {
			result = append(result, masterpatch{query.Entity(), who.Who, fs.PatchType, fs.Area_sqm})
		}
	}
	slices.SortFunc(result, func(a, b masterpatch) int { return int(a.who - b.who) })
	return result
}

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

// NestHabitatsListREP expands entries ending with "*" to all habitat types of the masterpatches that contain
// the text before the "*" [NHAB], finding 20 (see trans/netlogo-findings.md#finding-20).
func NestHabitatsListREP(list []string, masterpatches []masterpatch, rng *rand.Rand) ([]string, error) {
	// [NHAB-01] ORDER: irrelevant, nestHabitatsList is only used with member?. Shuffled anyway (decision C).
	order := slices.Clone(masterpatches)
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	allHabitats := []string{}
	for _, m := range order {
		if !slices.Contains(allHabitats, m.patchType) {
			allHabitats = append(allHabitats, m.patchType)
		}
	}

	// [NHAB-02]
	result := []string{}
	for _, entry := range list {
		if !strings.Contains(entry, "*") {
			result = append(result, entry)
			continue
		}
		general := entry[:len(entry)-1] // but-last
		if strings.Contains(general, "*") {
			return nil, fmt.Errorf("Wrong format of nesting habitat in BB species input file: %q", entry)
		}
		for _, h := range allHabitats {
			if strings.Contains(h, general) {
				result = append(result, h)
			}
		}
	}
	return result, nil
}

// nestSites returns all masterpatches whose patchType is in the nest habitat list (exact match), sorted by who,
// and the sum of their areas [SPEC-06].
func nestSites(habitats []string, masterpatches []masterpatch) ([]ecs.Entity, float64) {
	sites := []ecs.Entity{}
	area := 0.0
	for _, m := range masterpatches {
		if slices.Contains(habitats, m.patchType) {
			sites = append(sites, m.entity)
			area += m.area_sqm
		}
	}
	return sites, area
}

// stringList reads a NetLogo list literal of strings, e.g. ["Grassland*" "Garden"]. NetLogo: read-from-string.
func stringList(s string) ([]string, error) {
	v, err := util.ReadFromString(s)
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
