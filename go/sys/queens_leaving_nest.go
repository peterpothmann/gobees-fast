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
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// QueensLeavingNest [GO-12].
//
// Every young (unmated) adult queen in a colony leaves it, mates and starts to hibernate.
// A young queen emerges in Development [DEVP-04], which runs after this system, so she stays in the colony for
// one whole day and leaves on the next evening [QLN-01].
// Not translated: size, shape, color (GUI).
//
// NetLogo: QueensLeavingNestProc.
type QueensLeavingNest struct {
	rng       *rand.Rand
	file      *params.ParameterFile
	constants *params.Constants
	death     beeDeath

	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map3[comp.Bee, comp.Genetics, comp.Thresholds]
}

func (s *QueensLeavingNest) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.constants = ecs.GetResource[params.Constants](w)
	s.death = newBeeDeath(w)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap3[comp.Bee, comp.Genetics, comp.Thresholds](w)
}

func (s *QueensLeavingNest) Update(w *ecs.World) {
	// [QLN-01] ORDER: irrelevant. Each queen only changes herself, and mating does not use up the male.
	queens := askBees(s.beeFilter, s.rng, func(b *comp.Bee) bool {
		return b.Stage == stage.Adult && b.Caste == caste.Queen && !b.Mated && !b.ColonyID.IsZero()
	})
	if len(queens) == 0 {
		return
	}
	// [QLN-02] the adult males per species, anywhere in the landscape. Collected once: no male is created or dies
	// in this system, and a queen that dies is not a male. NetLogo evaluates the agentset per queen.
	males := map[ecs.Entity][]ecs.Entity{}
	query := s.beeFilter.Query()
	for query.Next() {
		if b := query.Get(); b.Caste == caste.Male && b.Stage == stage.Adult {
			males[b.SpeciesID] = append(males[b.SpeciesID], query.Entity())
		}
	}

	for _, e := range queens {
		bee, gen, th := s.beeMap.Get(e)
		// [QLN-02]
		if candidates := males[bee.SpeciesID]; len(candidates) > 0 {
			// one-of: a random male of her species, haploid or diploid
			_, maleGen, _ := s.beeMap.Get(candidates[s.rng.IntN(len(candidates))])
			gen.SpermathecaList = slices.Clone(maleGen.AllelesList)
		} else {
			// mating with a male from outside
			if !s.file.UnlimitedMales {
				s.death.DieProc(w, e, "Queen: no mating")
				continue // finding 28 (see trans/netlogo-findings.md#finding-28)
			}
			foreignAllele := -1*util.RandomInt(s.rng, s.constants.N_ForeignAlleles) - 1 // -1, -2, ... -N_ForeignAlleles
			gen.SpermathecaList = append([]float64{foreignAllele}, gen.SpermathecaList...)
			if len(gen.SpermathecaList) > 1 {
				panic(fmt.Sprintf("Assertion violated in QueensLeavingNestProc: too many alleles here! %v", gen.SpermathecaList))
			}
		}
		// [QLN-03] the other three thresholds stay "youngQueen" until she emerges next year [EMER-01]
		bee.Mated = true
		th.ThEgglaying = ThresholdLevelREP(thEggLaying, sitQueenInitiationPhase)
		bee.Activity = activity.Hibernate
		bee.ColonyID = ecs.Entity{}
		// [QLN-04] mated with a diploid male: she could not found a colony
		if len(gen.SpermathecaList) == 2 {
			s.death.DieProc(w, e, "Queen: mating with diploid male")
		}
	}
}

func (s *QueensLeavingNest) Finalize(w *ecs.World) {}
