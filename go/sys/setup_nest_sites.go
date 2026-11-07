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

// nestSitesSetup finds the nest sites of each simulated species and adds them as the resource
// globals.NestSites. Runs after landscapeSetup, because the nest sites are foodsources.
//
// NetLogo: CreateSpeciesProc, NestHabitatsListREP.
type nestSitesSetup struct {
	nestSites globals.NestSites
}

func (s *nestSitesSetup) Initialize(w *ecs.World) {
	in := ecs.GetResource[params.Input](w)
	rng := rand.New(ecs.GetResource[resource.Rand](w))

	// one species per species of the run file, in FILE order. Package input has read the rows into
	// params.Input.Species and set the derived values. Bees and colonies hold the index of their species.

	// Collect all masterpatches, in NetLogo's who order, to expand the nest habitats list and find the nest sites.
	masterpatches := collectMasterpatches(w)
	s.nestSites.BySpecies = make([]globals.SpeciesNestSites, len(in.Species))
	for i := range in.Species {
		sp := &in.Species[i]
		ns := &s.nestSites.BySpecies[i]
		expanded, err := expandNestHabitats(sp.NestHabitats, masterpatches, rng)
		if err != nil {
			panic(err)
		}
		ns.Habitats = expanded
		ns.Sites, ns.Area = nestSites(ns.Habitats, masterpatches)
		// message only
		if len(ns.Sites) == 0 && !sp.Parasitic {
			log.Printf("%s has no suitable nesting foodsources. No colonies will form", sp.Name)
		}
	}
	ecs.AddResource(w, &s.nestSites)
}

func (s *nestSitesSetup) Update(w *ecs.World) {}

func (s *nestSitesSetup) Finalize(w *ecs.World) {}

// speciesTable gives a species by its index, the Species of bees and colonies.
// It reads params.Input.Species at each call, so species added later (tests) are found too.
type speciesTable struct {
	in *params.Input
}

func newSpeciesTable(w *ecs.World) *speciesTable {
	return &speciesTable{in: ecs.GetResource[params.Input](w)}
}

// Get returns the species with index i.
func (t *speciesTable) Get(i int) *params.Species {
	return &t.in.Species[i]
}

// all returns the indices of all species in file order. NetLogo: ask species (shuffled by the caller).
func (t *speciesTable) all() []int {
	result := make([]int, len(t.in.Species))
	for i := range result {
		result[i] = i
	}
	return result
}

// masterpatch is a masterpatch foodsource with the values expandNestHabitats and nestSites need.
type masterpatch struct {
	entity    ecs.Entity
	patchType string
	area      float64 // [m2]
}

// collectMasterpatches returns all masterpatches in NetLogo's who order, which is the order of
// globals.Landscape.Masterpatches.
func collectMasterpatches(w *ecs.World) []masterpatch {
	landscape := ecs.GetResource[globals.Landscape](w)
	fsMap := ecs.NewMap1[comp.Foodsource](w)
	result := make([]masterpatch, len(landscape.Masterpatches))
	for i, e := range landscape.Masterpatches {
		fs := fsMap.Get(e)
		result[i] = masterpatch{e, fs.PatchType, fs.Area}
	}
	return result
}

// expandNestHabitats expands entries ending with "*" to all habitat types of the masterpatches that contain
// the text before the "*".
func expandNestHabitats(list []string, masterpatches []masterpatch, rng *rand.Rand) ([]string, error) {
	// ORDER: irrelevant, the nest habitats list is only used with member?. Shuffled anyway, as NetLogo collects the
	// habitat types of the masterpatches in random order.
	order := slices.Clone(masterpatches)
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	allHabitats := []string{}
	for _, m := range order {
		if !slices.Contains(allHabitats, m.patchType) {
			allHabitats = append(allHabitats, m.patchType)
		}
	}

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

// nestSites returns all masterpatches whose patchType is in the nest habitat list (exact match), in the order of
// masterpatches, and the sum of their areas.
func nestSites(habitats []string, masterpatches []masterpatch) ([]ecs.Entity, float64) {
	sites := []ecs.Entity{}
	area := 0.0
	for _, m := range masterpatches {
		if slices.Contains(habitats, m.patchType) {
			sites = append(sites, m.entity)
			area += m.area
		}
	}
	return sites, area
}
