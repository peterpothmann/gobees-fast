package sys

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"strings"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CreateSpecies creates one species entity per simulated species [INIT-07].
// Runs after CreateFoodsources, because the nest sites are foodsources.
//
// NetLogo: CreateSpeciesProc, NestHabitatsListREP.
type CreateSpecies struct{}

func (s *CreateSpecies) Initialize(w *ecs.World) {
	in := ecs.GetResource[params.Input](w)
	who := ecs.GetResource[globals.WhoCounter](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// [SPEC-01] one species entity per species in speciesList, in FILE order of the BeespeciesFile. Package input
	// has read the rows and set the derived values [SPEC-02], [SPEC-03], [SPEC-05], [SPEC-07].
	species := slices.Clone(in.Species)

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
