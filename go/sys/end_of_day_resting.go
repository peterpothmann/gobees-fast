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

// EndOfDayResting [GO-11].
//
// Sets all adult, non-hibernating females to resting, also the ones whose day ended earlier.
// Not translated: activityList (GUI log).
//
// NetLogo: ask bees ... set activity "resting" (in Go).
type EndOfDayResting struct {
	rng       *rand.Rand
	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map1[comp.Bee]
}

func (s *EndOfDayResting) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *EndOfDayResting) Update(w *ecs.World) {
	// [GO-11] ORDER: irrelevant, each bee only sets itself. Shuffled (decision C).
	for _, e := range askBees(s.beeFilter, s.rng, func(b *comp.Bee) bool {
		return b.Stage == stage.Adult && b.Activity != activity.Hibernate && b.Caste != caste.Male
	}) {
		s.beeMap.Get(e).Activity = activity.Resting
	}
}

func (s *EndOfDayResting) Finalize(w *ecs.World) {}
