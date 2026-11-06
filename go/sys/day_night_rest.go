package sys

import (
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// nightRest sets all adult, non-hibernating females to resting, also the ones whose day ended earlier.
//
// NetLogo: ask bees ... set activity "resting" (in Go).
type nightRest struct {
	rng       *rand.Rand
	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map1[comp.Bee]

	bees []ecs.Entity // buffer of askBees
}

func (s *nightRest) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *nightRest) Update(w *ecs.World) {
	// ORDER: irrelevant, each bee only sets itself. Shuffled, as NetLogo asks the bees in random order.
	s.bees = askBees(s.bees, s.beeFilter, s.rng, func(b *comp.Bee) bool {
		return b.Stage == stage.Adult && b.Activity != activity.Hibernate && b.Caste != caste.Male
	})
	for _, e := range s.bees {
		s.beeMap.Get(e).Activity = activity.Resting
	}
}

func (s *nightRest) Finalize(w *ecs.World) {}
