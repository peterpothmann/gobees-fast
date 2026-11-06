package sys

import (
	"log"
	"math"
	"math/rand/v2"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// cropRotation replaces the whole landscape by the next foodsources file of the rotation [CROP].
//
// Part of the morning routine [MORN-09]: on Day 365, if CropRotationList has more than 1 item (finding 3, see
// trans/netlogo-findings.md#finding-3). CropRotationListAsString is empty in
// netlogo/_SYSTEM_Example_Farm_Parameters.csv, so by default it never runs. Not run in Setup (Day 0).
// Not translated: Report_name (output file of a GUI button).
//
// NetLogo: the call in UpdateMorning_Proc and CropRotationProc.
type cropRotation struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	calendar  *globals.Calendar
	derived   *params.Derived
	landscape *globals.Landscape
	nestSites *globals.NestSites

	fsFilter     *ecs.Filter1[comp.Foodsource]
	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	foragerMap   *ecs.Map1[comp.Forager]
	rangeMap     *ecs.Map1[comp.ColonyRange]
	speciesMap   *speciesTable

	bees []ecs.Entity // buffer of askAllBees
}

func (s *cropRotation) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.derived = ecs.GetResource[params.Derived](w)
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.nestSites = ecs.GetResource[globals.NestSites](w)
	s.fsFilter = ecs.NewFilter1[comp.Foodsource](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.foragerMap = ecs.NewMap1[comp.Forager](w)
	s.rangeMap = ecs.NewMap1[comp.ColonyRange](w)
	s.speciesMap = newSpeciesTable(w)
}

func (s *cropRotation) Update(w *ecs.World) {
	// [MORN-09]
	if s.calendar.Day == 365 && len(s.derived.CropRotationList) > 1 {
		s.rotate(w)
	}
}

func (s *cropRotation) Finalize(w *ecs.World) {}

// rotate replaces all foodsources by the ones of the next file of the rotation [CROP].
func (s *cropRotation) rotate(w *ecs.World) {
	// [CROP-01] Worked example (list chosen by hand): 2 files, ticks 365: year 1, item 1; ticks 730: item 0.
	year := math.Ceil(float64(s.ticks.Ticks) / 365)
	nMaps := len(s.derived.CropRotationList)
	mapToUse := int(year) % nMaps // remainder: ticks > 0, so no negative values
	s.landscape.FoodsourcesFile = s.derived.CropRotationList[mapToUse]

	// [CROP-02] all foodsources die
	dead := []ecs.Entity{}
	fsQuery := s.fsFilter.Query()
	for fsQuery.Next() {
		dead = append(dead, fsQuery.Entity())
	}
	for _, e := range dead {
		w.RemoveEntity(e)
	}
	// every bee forgets her foodsources. ORDER: irrelevant, each bee only changes herself. Shuffled (decision C).
	s.bees = askAllBees(s.bees, s.beeFilter, s.rng)
	for _, e := range s.bees {
		f := s.foragerMap.Get(e)
		f.KnownNectarPatches = []ecs.Entity{}
		f.KnownPollenPatches = []ecs.Entity{}
		f.NectarSource = ecs.Entity{}
		f.PollenSource = ecs.Entity{}
		f.CurrentFoodsource = ecs.Entity{}
	}
	// every colony empties its in-flower lists of masterpatches, but NOT ColonyRange.Layers [PRNG]: it keeps the
	// dead foodsources. FoodsourcesInFlowerAndRangeProc [FRNG] reads
	// Layers on the next flowering change and panics there (Ark: "can't get components of a dead
	// entity", NetLogo: runtime error on a dead turtle). Needs a colony alive on Day 365, which the species of
	// netlogo/_SYSTEM_BumbleSpecies_UK_01.csv cannot have (seasonStop 305).
	// masterpatchesInRangeList is not translated (finding 23, see trans/netlogo-findings.md#finding-23).
	// ORDER: irrelevant, each colony only changes itself. Shuffled (decision C).
	for _, e := range askColonies(s.colonyFilter, s.rng) {
		r := s.rangeMap.Get(e)
		r.NectarPatches = []ecs.Entity{}
		r.PollenPatches = []ecs.Entity{}
	}

	// [CROP-03] the new landscape: new who numbers, random numbers for the hedges and layers
	createFoodsources(w, s.landscape.FoodsourcesFile, s.landscape)

	// per species the nest sites and their area as in [SPEC-06].
	// nestHabitatsList is NOT expanded again [SPEC-04]: it keeps the habitat types of the first landscape.
	// minPollenStore_g as in [SPEC-07]: not computed again (step F11), the formula reads only species values, which
	// do not change.
	// ORDER: irrelevant, each species only changes itself. Shuffled (decision C).
	species := s.speciesMap.all()
	util.Shuffle(s.rng, species)
	masterpatches := collectMasterpatches(w)
	for _, i := range species {
		ns := &s.nestSites.BySpecies[i]
		ns.Sites, ns.Area = nestSites(ns.Habitats, masterpatches)
		if sp := s.speciesMap.Get(i); len(ns.Sites) == 0 && !sp.Parasitic {
			log.Printf("%s has no suitable nesting habitat", sp.Name)
		}
	}
}
