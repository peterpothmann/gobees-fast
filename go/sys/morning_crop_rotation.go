package sys

import (
	"log"
	"math"

	"gobees/comp"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// cropRotation replaces the whole landscape by the next foodsources file of the rotation.
//
// Part of the morning routine: on Day 365, if CropRotationList has more than 1 item. CropRotationListAsString is
// empty in netlogo/_SYSTEM_Example_Farm_Parameters.csv, so by default it never runs. Not run in Setup (Day 0).
//
// NetLogo: the call in UpdateMorning_Proc and CropRotationProc.
type cropRotation struct {
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

	bees []ecs.Entity // buffer of allBees
}

func (s *cropRotation) Initialize(w *ecs.World) {
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
	if s.calendar.Day == 365 && len(s.derived.CropRotation) > 1 {
		s.rotate(w)
	}
}

func (s *cropRotation) Finalize(w *ecs.World) {}

// rotate replaces all foodsources by the ones of the next file of the rotation.
func (s *cropRotation) rotate(w *ecs.World) {
	// Worked example (list chosen by hand): 2 files, ticks 365: year 1, item 1. Ticks 730: item 0.
	year := math.Ceil(float64(s.ticks.Ticks) / 365)
	nMaps := len(s.derived.CropRotation)
	mapToUse := int(year) % nMaps // remainder: ticks > 0, so no negative values
	s.landscape.File = s.derived.CropRotation[mapToUse]

	// all foodsources die
	dead := []ecs.Entity{}
	fsQuery := s.fsFilter.Query()
	for fsQuery.Next() {
		dead = append(dead, fsQuery.Entity())
	}
	for _, e := range dead {
		w.RemoveEntity(e)
	}
	// every bee forgets her foodsources. ORDER: irrelevant, each bee only changes herself. Not shuffled (NetLogo:
	// random order).
	s.bees = allBees(s.bees, s.beeFilter)
	for _, e := range s.bees {
		f := s.foragerMap.Get(e)
		f.KnownNectarPatches = []ecs.Entity{}
		f.KnownPollenPatches = []ecs.Entity{}
		f.NectarSource = ecs.Entity{}
		f.PollenSource = ecs.Entity{}
		f.CurrentFoodsource = ecs.Entity{}
	}
	// every colony empties its in-flower lists of masterpatches, but NOT ColonyRange.Layers: it keeps the
	// dead foodsources. colonyFounder.flowersInRange reads
	// Layers on the next flowering change and panics there (Ark: "can't get components of a dead
	// entity", NetLogo: runtime error on a dead turtle). Needs a colony alive on Day 365, which the species of
	// netlogo/_SYSTEM_BumbleSpecies_UK_01.csv cannot have (seasonStop 305).
	// ORDER: irrelevant, each colony only changes itself. Not shuffled (NetLogo: random order).
	for _, e := range allColonies(s.colonyFilter) {
		r := s.rangeMap.Get(e)
		r.NectarPatches = []ecs.Entity{}
		r.PollenPatches = []ecs.Entity{}
		r.NectarDetection, r.PollenDetection = nil, nil
	}

	// the new landscape: new entities, random numbers for the hedges and layers
	createFoodsources(w, s.landscape.File, s.landscape)

	// per species the nest sites and their area as in nestSitesSetup.
	// nestHabitatsList is NOT expanded again: it keeps the habitat types of the first landscape.
	// minPollenStore_g is not computed again: the formula reads only species values, which do not change.
	// ORDER: irrelevant, each species only changes itself. Not shuffled (NetLogo: random order).
	species := s.speciesMap.all()
	masterpatches := collectMasterpatches(w)
	for _, i := range species {
		ns := &s.nestSites.BySpecies[i]
		ns.Sites, ns.Area = nestSites(ns.Habitats, masterpatches)
		if sp := s.speciesMap.Get(i); len(ns.Sites) == 0 && !sp.Parasitic {
			log.Printf("%s has no suitable nesting habitat", sp.Name)
		}
	}
}
