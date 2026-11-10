package sys

import (
	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"

	"github.com/mlange-42/ark/ecs"
)

// nightRest sets all adult, non-hibernating females to resting, also the ones whose day ended earlier.
//
// NetLogo: ask bees ... set activity "resting" (in Go).
type nightRest struct {
	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map1[comp.Bee]

	bees []ecs.Entity // buffer of beesWhere
}

func (s *nightRest) Initialize(w *ecs.World) {
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
}

func (s *nightRest) Update(w *ecs.World) {
	// ORDER: irrelevant, each bee only sets itself. Not shuffled (NetLogo: random order).
	s.bees = beesWhere(s.bees, s.beeFilter, func(b *comp.Bee) bool {
		return b.Stage == stage.Adult && b.Activity != activity.Hibernate && b.Caste != caste.Male
	})
	for _, e := range s.bees {
		s.beeMap.Get(e).Activity = activity.Resting
	}
}

func (s *nightRest) Finalize(w *ecs.World) {}
