package sys

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// queensLeavingNest lets every young (unmated) adult queen in a colony leave it, mate and start to hibernate.
// A young queen emerges in development, which runs after this system, so she stays in the colony for
// one whole day and leaves on the next evening.
//
// NetLogo: QueensLeavingNestProc.
type queensLeavingNest struct {
	rng        *rand.Rand
	parameters *params.Parameters
	constants  *params.Constants
	death      beeDeath

	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map3[comp.Bee, comp.Genetics, comp.Thresholds]

	queens []ecs.Entity // buffer of askBees
}

func (s *queensLeavingNest) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.parameters = ecs.GetResource[params.Parameters](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.death = newBeeDeath(w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap3[comp.Bee, comp.Genetics, comp.Thresholds](w)
}

func (s *queensLeavingNest) Update(w *ecs.World) {
	// ORDER: irrelevant. Each queen only changes herself, and mating does not use up the male.
	s.queens = askBees(s.queens, s.beeFilter, s.rng, func(b *comp.Bee) bool {
		return b.Stage == stage.Adult && b.Caste == caste.Queen && !b.Mated && !b.Colony.IsZero()
	})
	if len(s.queens) == 0 {
		return
	}
	// the adult males per species, anywhere in the landscape. Collected once: no male is created or dies
	// in this system, and a queen that dies is not a male. NetLogo evaluates the agentset per queen.
	males := map[int][]ecs.Entity{}
	query := s.beeFilter.Query()
	for query.Next() {
		if b := query.Get(); b.Caste == caste.Male && b.Stage == stage.Adult {
			males[b.Species] = append(males[b.Species], query.Entity())
		}
	}

	for _, e := range s.queens {
		bee, gen, th := s.beeMap.Get(e)
		if candidates := males[bee.Species]; len(candidates) > 0 {
			// one-of: a random male of her species, haploid or diploid
			_, maleGen, _ := s.beeMap.Get(candidates[s.rng.IntN(len(candidates))])
			gen.Spermatheca = slices.Clone(maleGen.Alleles)
		} else {
			// mating with a male from outside
			if !s.parameters.UnlimitedMales {
				s.death.die(w, e, "Queen: no mating")
				continue
			}
			foreignAllele := -1*randomInt(s.rng, s.constants.ForeignAlleles) - 1 // -1, -2, ... -ForeignAlleles
			gen.Spermatheca = append([]float64{foreignAllele}, gen.Spermatheca...)
			if len(gen.Spermatheca) > 1 {
				panic(fmt.Sprintf("Assertion violated in queensLeavingNest: too many alleles here! %v", gen.Spermatheca))
			}
		}
		// the other three thresholds stay "youngQueen" until she emerges next year
		bee.Mated = true
		th.Egglaying = thresholdLevel(thEggLaying, sitQueenInitiationPhase)
		bee.Activity = activity.Hibernate
		bee.Colony = ecs.Entity{}
		// mated with a diploid male: she could not found a colony
		if len(gen.Spermatheca) == 2 {
			s.death.die(w, e, "Queen: mating with diploid male")
		}
	}
}

func (s *queensLeavingNest) Finalize(w *ecs.World) {}
