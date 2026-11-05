package sys

import (
	"log"
	"math"
	"math/rand/v2"

	"gobees/comp"
	"gobees/globals"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// CropRotation replaces the whole landscape by the next foodsources file of the rotation [CROP].
//
// Part of the morning routine [MORN-09]: on Day 365, if CropRotationList has more than 1 item (finding 3, see
// trans/netlogo-findings.md#finding-3). CropRotationListAsString is empty in
// netlogo/_SYSTEM_Example_Farm_Parameters.csv, so by default it never runs. Not run in Setup (Day 0).
// Not translated: Report_name (output file of a GUI button).
//
// NetLogo: the call in UpdateMorning_Proc and CropRotationProc.
type CropRotation struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	calendar  *globals.Calendar
	derived   *globals.Derived
	landscape *globals.Landscape

	fsFilter      *ecs.Filter1[comp.Foodsource]
	beeFilter     *ecs.Filter1[comp.Bee]
	colonyFilter  *ecs.Filter1[comp.Colony]
	speciesFilter *ecs.Filter1[comp.Species]
	foragerMap    *ecs.Map1[comp.Forager]
	rangeMap      *ecs.Map1[comp.ColonyRange]
	speciesMap    *ecs.Map1[comp.Species]
}

func (s *CropRotation) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.derived = ecs.GetResource[globals.Derived](w)
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.fsFilter = ecs.NewFilter1[comp.Foodsource](w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.speciesFilter = ecs.NewFilter1[comp.Species](w)
	s.foragerMap = ecs.NewMap1[comp.Forager](w)
	s.rangeMap = ecs.NewMap1[comp.ColonyRange](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
}

func (s *CropRotation) Update(w *ecs.World) {
	// [MORN-09]
	if s.calendar.Day == 365 && len(s.derived.CropRotationList) > 1 {
		s.CropRotationProc(w)
	}
}

func (s *CropRotation) Finalize(w *ecs.World) {}

// CropRotationProc replaces all foodsources by the ones of the next file of the rotation [CROP].
func (s *CropRotation) CropRotationProc(w *ecs.World) {
	// [CROP-01] Worked example (list chosen by hand): 2 files, ticks 365: year 1, item 1; ticks 730: item 0.
	year := math.Ceil(float64(s.ticks.Ticks) / 365)
	nMaps := len(s.derived.CropRotationList)
	mapToUse := int(year) % nMaps // remainder: ticks > 0, so no negative values
	s.derived.FoodsourcesFile = s.derived.CropRotationList[mapToUse]

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
	for _, e := range askAllBees(s.beeFilter, s.rng) {
		f := s.foragerMap.Get(e)
		f.KnownMasterpatchesNectarList = []ecs.Entity{}
		f.KnownMasterpatchesPollenList = []ecs.Entity{}
		f.NectarsourceToGoTo = ecs.Entity{}
		f.PollensourceToGoTo = ecs.Entity{}
		f.CurrentFoodsource = ecs.Entity{}
	}
	// every colony empties its in-flower lists of masterpatches, but NOT AllPatchesInRangeList [PRNG]: it keeps the
	// dead foodsources. FoodsourcesInFlowerAndRangeProc [FRNG] reads
	// AllPatchesInRangeList on the next flowering change and panics there (Ark: "can't get components of a dead
	// entity", NetLogo: runtime error on a dead turtle). Needs a colony alive on Day 365, which the species of
	// netlogo/_SYSTEM_BumbleSpecies_UK_01.csv cannot have (seasonStop 305).
	// masterpatchesInRangeList is not translated (finding 23, see trans/netlogo-findings.md#finding-23).
	// ORDER: irrelevant, each colony only changes itself. Shuffled (decision C).
	for _, e := range askColonies(s.colonyFilter, s.rng) {
		r := s.rangeMap.Get(e)
		r.MasterpatchesWithNectarlayersInFlowerAndRangeList = []ecs.Entity{}
		r.MasterpatchesWithPollenlayersInFlowerAndRangeList = []ecs.Entity{}
	}

	// [CROP-03] the new landscape: new who numbers, random numbers for the hedges and layers
	CreateFoodsourcesProc(w, s.derived.FoodsourcesFile, s.landscape)

	// per species the nest sites and their area as in [SPEC-06], minPollenStore_g as in [SPEC-07].
	// nestHabitatsList is NOT expanded again [SPEC-04]: it keeps the habitat types of the first landscape.
	// ORDER: irrelevant, each species only changes itself. Shuffled (decision C).
	species := []ecs.Entity{}
	spQuery := s.speciesFilter.Query()
	for spQuery.Next() {
		species = append(species, spQuery.Entity())
	}
	util.Shuffle(s.rng, species)
	masterpatches := collectMasterpatches(w)
	for _, e := range species {
		sp := s.speciesMap.Get(e)
		sp.NestsiteFoodsourceList, sp.NestSiteArea = nestSites(sp.NestHabitatsList, masterpatches)
		sp.MinPollenStore_g = 0.5 * 0.001 * sp.DevWeightPupationMin_mg * float64(sp.Batchsize) / sp.PollenToBodymassFactor
		if len(sp.NestsiteFoodsourceList) == 0 && !sp.Parasitic {
			log.Printf("%s has no suitable nesting habitat", sp.Name)
		}
	}
}
