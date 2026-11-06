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
// a colony search a nest site and founds colonies [EMER].
//
// Part of the morning routine [MORN-07], after flowerRefill and seasonalEvents.
// At ticks = 0 (Setup, [INIT-10]) no queen emerges, so it does not run then.
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
	s.colonies = newColonyFounder(w, s.rng)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter2[comp.Position, comp.Colony](w)
	s.beeMap = newBeeMapper(w)
	s.speciesMap = newSpeciesTable(w)
	s.foodsourceMap = ecs.NewMap2[comp.Position, comp.Foodsource](w)
}

func (s *queenEmergence) Update(w *ecs.World) {
	ticks := s.ticks.Ticks

	// [EMER-01] emergence, cohorts are split into identical copies (finding 32, see trans/netlogo-findings.md#finding-32)
	for _, e := range s.collectBees(func(b *comp.Bee) bool { return b.EmergenceDate.Is(ticks) }) {
		_, bee, _, th, _, _ := s.beeMap.Get(e)
		if bee.Caste != caste.Queen {
			fmt.Println("WARNING! Something else than a queen emerged from hibernation! (1)")
		}
		bee.Activity = activity.Emerging
		*th = thresholds(sitQueenInitiationPhase)
		hatchlings := int(bee.Number) - 1
		bee.Number = 1
		s.copyBee(e, hatchlings)
	}

	// [EMER-02] ORDER: irrelevant, each queen only changes herself.
	for _, e := range s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.Emerging }) {
		_, bee, _, _, _, _ := s.beeMap.Get(e)
		if bee.Caste != caste.Queen {
			fmt.Println("WARNING! Something else than a queen emerged from hibernation! (2)")
		}
		// [EMER-03] finding 33 (see trans/netlogo-findings.md#finding-33)
		if s.parameters.WinterMortality && s.rng.Float64() > winterMortalityProb(bee.Weight, s.speciesMap.Get(bee.Species)) {
			s.death.die(w, e, "winter mortality")
			continue // finding 28 (see trans/netlogo-findings.md#finding-28)
		}
		bee.Activity = activity.Resting
		bee.Colony = ecs.Entity{}
	}

	// [EMER-04] ORDER: irrelevant for social species, see the annotation.
	searching := s.collectBees(func(b *comp.Bee) bool {
		return b.Caste == caste.Queen && b.Colony.IsZero() && b.Activity != activity.Hibernate
	})
	if len(searching) == 0 {
		return
	}
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
		if !sp.Parasitic { // speciesName != "Psithyrus" (step F9)
			s.searchNestSite(w, e, sp, ns)
		} else {
			s.searchHostColony(w, e)
		}
	}

	// [EMER-05]
	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		s.colonies.found(w)
	}
	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		fmt.Println("That's weird - no constructions here, please!")
	}
}

func (s *queenEmergence) Finalize(w *ecs.World) {}

// collectBees returns all bees for which keep is true, shuffled (decision C, NetLogo: ask bees with [...]).
func (s *queenEmergence) collectBees(keep func(b *comp.Bee) bool) []ecs.Entity {
	result := []ecs.Entity{}
	query := s.beeFilter.Query()
	for query.Next() {
		if keep(query.Get()) {
			result = append(result, query.Entity())
		}
	}
	shuffle(s.rng, result)
	return result
}

// copyBee creates n copies of a bee with all its values. NetLogo: hatch n (with a new who number each, not kept
// since step F12).
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

// searchNestSite lets a queen search a nest site for one day [NEST].
func (s *queenEmergence) searchNestSite(w *ecs.World, e ecs.Entity, sp *params.Species, ns *globals.SpeciesNestSites) {
	// [NEST-01]
	if s.rng.Float64() <= sp.DailyNestSiteChance {
		// [NEST-02]
		fsE := s.chooseNestSite(ns)
		fsPos, fs := s.foodsourceMap.Get(fsE)
		// [NEST-03] a random patch centre within the circle, finding 25 (see trans/netlogo-findings.md#finding-25)
		candidates := patchCentresWithin(fsPos.X, fsPos.Y, fs.Radius*s.landscape.Scale,
			int(s.world.MinPxcor), int(s.world.MaxPxcor), int(s.world.MinPycor), int(s.world.MaxPycor))
		// No patch centre within the circle: the nearest patch centre, the one of the patch cell of the
		// foodsource, without a random number. NetLogo: runtime error. Fixed in step F5.
		c := [2]int{pcor(fsPos.X), pcor(fsPos.Y)}
		if len(candidates) > 0 {
			c = candidates[s.rng.IntN(len(candidates))]
		}
		// [NEST-04]
		pos, bee, _, _, _, _ := s.beeMap.Get(e)
		pos.X, pos.Y = float64(c[0]), float64(c[1])
		bee.Activity = activity.NestConstruction
		return
	}
	// [NEST-04] finding 26 (see trans/netlogo-findings.md#finding-26)
	if s.rng.Float64() < s.searchMortality() {
		s.death.die(w, e, "Queen: died while searching nest site")
	}
}

// searchMortality returns the probability to die during one day of nest search [NEST-04], [PSI-04].
func (s *queenEmergence) searchMortality() float64 {
	return 1 - math.Pow(1-s.derived.ForagerMortality, s.constants.NestSearchTime*60*60)
}

// chooseNestSite chooses a nest site, weighted by area [NSFS].
func (s *queenEmergence) chooseNestSite(ns *globals.SpeciesNestSites) ecs.Entity {
	// [NSFS-01] Sites is sorted by who. SIMPLIFY: shuffle not needed for the distribution.
	list := slices.Clone(ns.Sites)
	shuffle(s.rng, list)
	p := s.rng.Float64()
	probsSummedUp := 0.0
	for _, e := range list {
		_, fs := s.foodsourceMap.Get(e)
		probsSummedUp += fs.Area / ns.Area
		if probsSummedUp > p {
			return e
		}
	}
	// The sum ended below p by rounding (finding 27, see trans/netlogo-findings.md#finding-27): the last site.
	// NetLogo: runtime error. Fixed in step F5.
	return list[len(list)-1]
}

// searchHostColony lets a cuckoo bee search a host colony for one day [PSI].
func (s *queenEmergence) searchHostColony(w *ecs.World, e ecs.Entity) {
	// [PSI-01] arbitrary values (comments in the NetLogo code)
	const (
		findSingleNestProb = 0.05
		getAccessProb      = 0.25
		getKilledProb      = 0.25
		killQueenProb      = 0.5
	)
	colonies := []ecs.Entity{}
	colonyPos := map[ecs.Entity]comp.Position{}
	query := s.colonyFilter.Query()
	for query.Next() {
		pos, _ := query.Get()
		colonies = append(colonies, query.Entity())
		colonyPos[query.Entity()] = *pos
	}
	findAnyNestProb := 1 - math.Pow(1-findSingleNestProb, float64(len(colonies)))

	// [PSI-02] finding 35 (see trans/netlogo-findings.md#finding-35)
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
			*th = thresholds(sitPsith)
		}
	}

	if !host.IsZero() {
		// [PSI-03] ORDER: irrelevant, each queen draws independently.
		victims := s.collectBees(func(b *comp.Bee) bool { return b.Colony == host && b.Caste == caste.Queen && b.Mated })
		for _, v := range victims {
			if v != e && s.rng.Float64() < killQueenProb {
				s.death.die(w, v, "Queen killed by cuckoo bee!")
			}
		}
		return
	}
	// [PSI-04]
	if s.rng.Float64() < s.searchMortality() {
		s.death.die(w, e, "Psithyrus: died while searching nest")
	}
}
