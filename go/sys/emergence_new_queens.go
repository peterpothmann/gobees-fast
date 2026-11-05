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
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// EmergenceNewQueens lets queens emerge from hibernation, applies winter mortality, lets every queen without
// a colony search a nest site and founds colonies [EMER].
//
// Part of the morning routine [MORN-07], after UpdateFoodsources and UpdateSeasonalEvents.
// At ticks = 0 (Setup, [INIT-10]) no queen emerges, so it does not run in Initialize.
//
// NetLogo: EmergenceNewQueensProc, NestSitesSearchingProc, NestSiteFoodSourceREP, PsithyrusNestSearchProc,
// CreateColoniesProc.
type EmergenceNewQueens struct {
	rng       *rand.Rand
	ticks     *globals.Ticks
	file      *params.ParameterFile
	constants *params.Constants
	derived   *globals.Derived
	landscape *globals.Landscape
	world     *params.World
	who       *globals.WhoCounter
	death     beeDeath
	colonies  colonyFounder

	beeFilter     *ecs.Filter1[comp.Bee]
	colonyFilter  *ecs.Filter2[comp.Position, comp.Colony]
	beeMap        *beeMapper
	speciesMap    *ecs.Map1[comp.Species]
	foodsourceMap *ecs.Map2[comp.Position, comp.Foodsource]
}

func (s *EmergenceNewQueens) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.derived = ecs.GetResource[globals.Derived](w)
	s.landscape = ecs.GetResource[globals.Landscape](w)
	s.world = ecs.GetResource[params.World](w)
	s.who = ecs.GetResource[globals.WhoCounter](w)
	s.death = newBeeDeath(w)
	s.colonies = newColonyFounder(w, s.rng)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter2[comp.Position, comp.Colony](w)
	s.beeMap = newBeeMapper(w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
	s.foodsourceMap = ecs.NewMap2[comp.Position, comp.Foodsource](w)
}

func (s *EmergenceNewQueens) Update(w *ecs.World) {
	ticks := s.ticks.Ticks

	// [EMER-01] emergence, cohorts are split into identical copies (finding 32, see trans/netlogo-findings.md#finding-32)
	for _, e := range s.collectBees(func(b *comp.Bee) bool { return b.EmergingDate.Is(ticks) }) {
		_, _, bee, _, th, _, _ := s.beeMap.Get(e)
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
		_, _, bee, _, _, _, _ := s.beeMap.Get(e)
		if bee.Caste != caste.Queen {
			fmt.Println("WARNING! Something else than a queen emerged from hibernation! (2)")
		}
		// [EMER-03] finding 33 (see trans/netlogo-findings.md#finding-33)
		if s.file.WinterMortality && s.rng.Float64() > WintermortalityProbREP(bee.Weight_mg, s.speciesMap.Get(bee.SpeciesID)) {
			s.death.DieProc(w, e, "winter mortality")
			continue // finding 28 (see trans/netlogo-findings.md#finding-28)
		}
		bee.Activity = activity.Resting
		bee.ColonyID = ecs.Entity{}
	}

	// [EMER-04] ORDER: irrelevant for social species, see the annotation.
	searching := s.collectBees(func(b *comp.Bee) bool {
		return b.Caste == caste.Queen && b.ColonyID.IsZero() && b.Activity != activity.Hibernate
	})
	if len(searching) == 0 {
		return
	}
	for _, e := range searching {
		if !w.Alive(e) {
			continue // killed by a cuckoo bee of this loop
		}
		_, _, bee, _, _, _, _ := s.beeMap.Get(e)
		sp := s.speciesMap.Get(bee.SpeciesID)
		if len(sp.NestsiteFoodsourceList) == 0 {
			s.death.DieProc(w, e, "no suitable foodsources for nesting exist for "+bee.SpeciesName)
			continue
		}
		if bee.SpeciesName != "Psithyrus" {
			s.NestSitesSearchingProc(w, e, sp)
		} else {
			s.PsithyrusNestSearchProc(w, e)
		}
	}

	// [EMER-05]
	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		s.colonies.CreateColoniesProc(w)
	}
	if len(s.collectBees(func(b *comp.Bee) bool { return b.Activity == activity.NestConstruction })) > 0 {
		fmt.Println("That's weird - no constructions here, please!")
	}
}

func (s *EmergenceNewQueens) Finalize(w *ecs.World) {}

// collectBees returns all bees for which keep is true, shuffled (decision C, NetLogo: ask bees with [...]).
func (s *EmergenceNewQueens) collectBees(keep func(b *comp.Bee) bool) []ecs.Entity {
	result := []ecs.Entity{}
	query := s.beeFilter.Query()
	for query.Next() {
		if keep(query.Get()) {
			result = append(result, query.Entity())
		}
	}
	util.Shuffle(s.rng, result)
	return result
}

// copyBee creates n copies of a bee with all its values and a new who number each. NetLogo: hatch n.
func (s *EmergenceNewQueens) copyBee(e ecs.Entity, n int) {
	if n <= 0 {
		return
	}
	_, pos, bee, gen, th, forager, dev := s.beeMap.Get(e)
	// copy the values first: creating entities can move the components in memory
	p, b, g, t, f, d := *pos, *bee, *gen, *th, *forager, *dev
	for range n {
		gc, fc := g, f
		gc.AllelesList = slices.Clone(g.AllelesList)
		gc.SpermathecaList = slices.Clone(g.SpermathecaList)
		fc.KnownMasterpatchesNectarList = slices.Clone(f.KnownMasterpatchesNectarList)
		fc.KnownMasterpatchesPollenList = slices.Clone(f.KnownMasterpatchesPollenList)
		pc, bc, tc, dc := p, b, t, d
		s.beeMap.NewEntity(&comp.Who{Who: s.who.Next()}, &pc, &bc, &gc, &tc, &fc, &dc)
	}
}

// NestSitesSearchingProc lets a queen search a nest site for one day [NEST].
func (s *EmergenceNewQueens) NestSitesSearchingProc(w *ecs.World, e ecs.Entity, sp *comp.Species) {
	// [NEST-01]
	if s.rng.Float64() <= sp.ChanceFindNest {
		// [NEST-02]
		fsE := s.NestSiteFoodSourceREP(sp)
		fsPos, fs := s.foodsourceMap.Get(fsE)
		// [NEST-03] a random patch centre within the circle, finding 25 (see trans/netlogo-findings.md#finding-25)
		candidates := util.PatchCentresWithin(fsPos.X, fsPos.Y, fs.Radius_m*s.landscape.Scaling_NLpatches_per_m,
			int(s.world.MinPxcor), int(s.world.MaxPxcor), int(s.world.MinPycor), int(s.world.MaxPycor))
		// No patch centre within the circle: the nearest patch centre, the one of the patch cell of the
		// foodsource, without a random number. NetLogo: runtime error. Fixed in step F5.
		c := [2]int{util.Pcor(fsPos.X), util.Pcor(fsPos.Y)}
		if len(candidates) > 0 {
			c = candidates[s.rng.IntN(len(candidates))]
		}
		// [NEST-04]
		_, pos, bee, _, _, _, _ := s.beeMap.Get(e)
		pos.X, pos.Y = float64(c[0]), float64(c[1])
		bee.Activity = activity.NestConstruction
		return
	}
	// [NEST-04] finding 26 (see trans/netlogo-findings.md#finding-26)
	if s.rng.Float64() < s.searchMortality() {
		s.death.DieProc(w, e, "Queen: died while searching nest site")
	}
}

// searchMortality returns the probability to die during one day of nest search [NEST-04], [PSI-04].
func (s *EmergenceNewQueens) searchMortality() float64 {
	return 1 - math.Pow(1-s.derived.MortalityForager_per_s, s.constants.NestSearchTime_h*60*60)
}

// NestSiteFoodSourceREP chooses a nest site, weighted by area [NSFS].
func (s *EmergenceNewQueens) NestSiteFoodSourceREP(sp *comp.Species) ecs.Entity {
	// [NSFS-01] NestsiteFoodsourceList is sorted by who. SIMPLIFY: shuffle not needed for the distribution.
	list := slices.Clone(sp.NestsiteFoodsourceList)
	util.Shuffle(s.rng, list)
	p := s.rng.Float64()
	probsSummedUp := 0.0
	for _, e := range list {
		_, fs := s.foodsourceMap.Get(e)
		probsSummedUp += fs.Area_sqm / sp.NestSiteArea
		if probsSummedUp > p {
			return e
		}
	}
	// The sum ended below p by rounding (finding 27, see trans/netlogo-findings.md#finding-27): the last site.
	// NetLogo: runtime error. Fixed in step F5.
	return list[len(list)-1]
}

// PsithyrusNestSearchProc lets a cuckoo bee search a host colony for one day [PSI].
func (s *EmergenceNewQueens) PsithyrusNestSearchProc(w *ecs.World, e ecs.Entity) {
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
				s.death.DieProc(w, e, "Psithyrus: killed by Bombus queen")
				return
			}
			host = colonies[s.rng.IntN(len(colonies))]
			_, pos, bee, _, th, _, _ := s.beeMap.Get(e)
			bee.ColonyID = host
			*pos = colonyPos[host]
			*th = thresholds(sitPsith)
		}
	}

	if !host.IsZero() {
		// [PSI-03] ORDER: irrelevant, each queen draws independently.
		victims := s.collectBees(func(b *comp.Bee) bool { return b.ColonyID == host && b.Caste == caste.Queen && b.Mated })
		for _, v := range victims {
			if v != e && s.rng.Float64() < killQueenProb {
				s.death.DieProc(w, v, "Queen killed by cuckoo bee!")
			}
		}
		return
	}
	// [PSI-04]
	if s.rng.Float64() < s.searchMortality() {
		s.death.DieProc(w, e, "Psithyrus: died while searching nest")
	}
}
