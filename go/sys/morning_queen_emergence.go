package sys

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// queenEmergence lets queens emerge from hibernation, applies winter mortality, lets every queen without
// a colony search a nest site and founds colonies.
//
// Part of the morning routine, after flowerRefill and seasonalEvents.
// At ticks = 0 (Setup) no queen emerges, so it does not run then.
//
// NetLogo: EmergenceNewQueensProc, NestSitesSearchingProc, NestSiteFoodSourceREP, PsithyrusNestSearchProc,
// CreateColoniesProc.
type queenEmergence struct {
	rng        *rand.Rand
	ticks      *globals.Ticks
	parameters *params.Parameters
	constants  *params.Constants
	derived    *params.Derived
	landscape  *globals.Landscape
	nestSites  *globals.NestSites
	world      *params.World
	death      beeDeath
	colonies   colonyFounder

	beeFilter     *ecs.Filter1[comp.Bee]
	colonyFilter  *ecs.Filter2[comp.Position, comp.Colony]
	beeMap        *beeMapper
	speciesMap    *speciesTable
	foodsourceMap *ecs.Map2[comp.Position, comp.Foodsource]
}

func (s *queenEmergence) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.parameters = ecs.GetResource[params.Parameters](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.derived = ecs.GetResource[params.Derived](w)
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.nestSites = ecs.GetResource[globals.NestSites](w)
	s.world = ecs.GetResource[params.World](w)
	s.death = newBeeDeath(w)
	s.colonies = newColonyFounder(w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter2[comp.Position, comp.Colony](w)
	s.beeMap = newBeeMapper(w)
	s.speciesMap = newSpeciesTable(w)
	s.foodsourceMap = ecs.NewMap2[comp.Position, comp.Foodsource](w)
}

func (s *queenEmergence) Update(w *ecs.World) {
	ticks := s.ticks.Ticks

	// emergence: cohorts are split into identical copies
	for _, e := range s.collectBees(func(b *comp.Bee) bool { return b.EmergenceDate.Is(ticks) }) {
		_, bee, _, th, _, _ := s.beeMap.Get(e)
		if bee.Caste != caste.Queen {
			fmt.Println("WARNING! Something else than a queen emerged from hibernation! (1)")
		}
		bee.Activity = activity.Emerging
		*th = thresholds(&s.constants.Thresholds, sitQueenInitiationPhase)
		hatchlings := int(bee.Number) - 1
		bee.Number = 1
		s.copyBee(e, hatchlings)
	}

	// ORDER: irrelevant, each queen only changes herself.
	for _, e := range s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.Emerging }) {
		_, bee, _, _, _, _ := s.beeMap.Get(e)
		if bee.Caste != caste.Queen {
			fmt.Println("WARNING! Something else than a queen emerged from hibernation! (2)")
		}
		if s.parameters.WinterMortality && s.rng.Float64() > winterMortalityProb(bee.Weight, s.speciesMap.Get(bee.Species)) {
			s.death.die(w, e, "winter mortality")
			continue
		}
		bee.Activity = activity.Resting
		bee.Colony = ecs.Entity{}
	}

	// ORDER: MATTERS. Irrelevant for social species, the queens choose nest sites independently (no check whether a
	// site is taken). For cuckoo bees the order decides which cuckoo bee may kill another one in the same host first.
	// Shuffled, as NetLogo asks the queens in random order.
	searching := s.collectBees(func(b *comp.Bee) bool {
		return b.Caste == caste.Queen && b.Colony.IsZero() && b.Activity != activity.Hibernate
	})
	if len(searching) == 0 {
		return
	}
	shuffle(s.rng, searching)
	for _, e := range searching {
		if !w.Alive(e) {
			continue // killed by a cuckoo bee of this loop
		}
		_, bee, _, _, _, _ := s.beeMap.Get(e)
		sp := s.speciesMap.Get(bee.Species)
		ns := &s.nestSites.BySpecies[bee.Species]
		if len(ns.Sites) == 0 {
			s.death.die(w, e, "no suitable foodsources for nesting exist for "+sp.Name)
			continue
		}
		if !sp.Parasitic { // not Psithyrus
			s.searchNestSite(w, e, sp, ns)
		} else {
			s.searchHostColony(w, e)
		}
	}

	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		s.colonies.found(w)
	}
	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		fmt.Println("That's weird - no constructions here, please!")
	}
}

func (s *queenEmergence) Finalize(w *ecs.World) {}

// collectBees returns all bees for which keep is true, in storage order. NetLogo asks them in random order. The
// emerging cohorts, the emerged queens and the victims of a cuckoo bee each change only themselves with their own
// draws, the counts of the queens in nest construction only use the length: not shuffled. Update shuffles the
// searching queens itself.
func (s *queenEmergence) collectBees(keep func(b *comp.Bee) bool) []ecs.Entity {
	result := []ecs.Entity{}
	query := s.beeFilter.Query()
	for query.Next() {
		if keep(query.Get()) {
			result = append(result, query.Entity())
		}
	}
	return result
}

// copyBee creates n copies of a bee with all its values (hatch n in NetLogo).
func (s *queenEmergence) copyBee(e ecs.Entity, n int) {
	if n <= 0 {
		return
	}
	pos, bee, gen, th, forager, dev := s.beeMap.Get(e)
	// copy the values first: creating entities can move the components in memory
	p, b, g, t, f, d := *pos, *bee, *gen, *th, *forager, *dev
	for range n {
		gc, fc := g, f
		gc.Alleles = slices.Clone(g.Alleles)
		gc.Spermatheca = slices.Clone(g.Spermatheca)
		fc.KnownNectarPatches = slices.Clone(f.KnownNectarPatches)
		fc.KnownPollenPatches = slices.Clone(f.KnownPollenPatches)
		pc, bc, tc, dc := p, b, t, d
		s.beeMap.NewEntity(&pc, &bc, &gc, &tc, &fc, &dc)
	}
}

// searchNestSite lets a queen search a nest site for one day.
func (s *queenEmergence) searchNestSite(w *ecs.World, e ecs.Entity, sp *params.Species, ns *globals.SpeciesNestSites) {
	if s.rng.Float64() <= sp.DailyNestSiteChance {
		fsE := s.chooseNestSite(ns)
		fsPos, fs := s.foodsourceMap.Get(fsE)
		// a random patch centre within the circle
		candidates := patchCentresWithin(fsPos.X, fsPos.Y, fs.Radius*s.landscape.Scale,
			int(s.world.MinPxcor), int(s.world.MaxPxcor), int(s.world.MinPycor), int(s.world.MaxPycor))
		// No patch centre within the circle: the nearest patch centre, the one of the patch cell of the
		// foodsource, without a random number. NetLogo stops here with a runtime error.
		c := [2]int{pcor(fsPos.X), pcor(fsPos.Y)}
		if len(candidates) > 0 {
			c = candidates[s.rng.IntN(len(candidates))]
		}
		pos, bee, _, _, _, _ := s.beeMap.Get(e)
		pos.X, pos.Y = float64(c[0]), float64(c[1])
		bee.Activity = activity.NestConstruction
		return
	}
	if s.rng.Float64() < s.searchMortality() {
		s.death.die(w, e, "Queen: died while searching nest site")
	}
}

// searchMortality returns the probability to die during one day of nest search.
func (s *queenEmergence) searchMortality() float64 {
	return 1 - math.Pow(1-s.derived.ForagerMortality, s.constants.Bees.NestSearchTime)
}

// chooseNestSite chooses a nest site, weighted by area.
func (s *queenEmergence) chooseNestSite(ns *globals.SpeciesNestSites) ecs.Entity {
	// Sites is in the order of globals.Landscape.Masterpatches. ORDER: irrelevant, with one uniform number p each
	// site is chosen with the probability area / total area in any order. Not shuffled (NetLogo: shuffle).
	list := ns.Sites
	p := s.rng.Float64()
	probsSummedUp := 0.0
	for _, e := range list {
		_, fs := s.foodsourceMap.Get(e)
		probsSummedUp += fs.Area / ns.Area
		if probsSummedUp > p {
			return e
		}
	}
	// The sum ended below p by rounding: the last site.
	// NetLogo stops here with a runtime error.
	return list[len(list)-1]
}

// searchHostColony lets a cuckoo bee search a host colony for one day.
func (s *queenEmergence) searchHostColony(w *ecs.World, e ecs.Entity) {
	findSingleNestProb := s.constants.Psithyrus.FindNestProb
	getAccessProb := s.constants.Psithyrus.AccessProb
	getKilledProb := s.constants.Psithyrus.KilledProb
	killQueenProb := s.constants.Psithyrus.KillQueenProb
	colonies := []ecs.Entity{}
	colonyPos := map[ecs.Entity]comp.Position{}
	query := s.colonyFilter.Query()
	for query.Next() {
		pos, _ := query.Get()
		colonies = append(colonies, query.Entity())
		colonyPos[query.Entity()] = *pos
	}
	findAnyNestProb := 1 - math.Pow(1-findSingleNestProb, float64(len(colonies)))

	var host ecs.Entity
	if s.rng.Float64() < findAnyNestProb {
		if s.rng.Float64() < getAccessProb {
			if s.rng.Float64() < getKilledProb {
				s.death.die(w, e, "Psithyrus: killed by Bombus queen")
				return
			}
			host = colonies[s.rng.IntN(len(colonies))]
			pos, bee, _, th, _, _ := s.beeMap.Get(e)
			bee.Colony = host
			*pos = colonyPos[host]
			*th = thresholds(&s.constants.Thresholds, sitPsith)
		}
	}

	if !host.IsZero() {
		// ORDER: irrelevant, each queen draws independently.
		victims := s.collectBees(func(b *comp.Bee) bool { return b.Colony == host && b.Caste == caste.Queen && b.Mated })
		for _, v := range victims {
			if v != e && s.rng.Float64() < killQueenProb {
				s.death.die(w, v, "Queen killed by cuckoo bee!")
			}
		}
		return
	}
	if s.rng.Float64() < s.searchMortality() {
		s.death.die(w, e, "Psithyrus: died while searching nest")
	}
}
