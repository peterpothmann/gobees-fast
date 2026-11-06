package sys

import (
	"math/rand/v2"
	"slices"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// wakeUp gives every active adult worker and queen her wake-up time and fills the daily activity queue
// [MORN-08].
//
// Part of the morning routine, after queenEmergence [MORN-07] and before colonyUpdate [MORN-10].
// Initialize runs it also at ticks = 0 [INIT-10] (all queens hibernate then, the queue stays empty).
// Not translated: activityList (GUI log).
//
// NetLogo: part of UpdateMorning_Proc.
type wakeUp struct {
	rng       *rand.Rand
	constants *params.Constants
	active    *globals.ActiveBees

	beeFilter *ecs.Filter1[comp.Bee]
	beeMap    *ecs.Map1[comp.Bee]

	bees []ecs.Entity // buffer of askBees
}

func (s *wakeUp) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)
	s.active = ecs.GetResource[globals.ActiveBees](w)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.beeMap = ecs.NewMap1[comp.Bee](w)
	s.Update(w) // run at ticks = 0 [INIT-10]
}

func (s *wakeUp) Update(w *ecs.World) {
	// [MORN-08] ORDER: MATTERS for ties. NetLogo puts the bees in the random order of ask into the list (fput,
	// i.e. in reverse), then sorts it with the stable sort-by. Here: shuffle in askBees, then a stable sort.
	randomTimeToGetUp_s := 1800.0
	s.bees = askBees(s.bees, s.beeFilter, s.rng, isActiveAdult)
	list := make([]ecs.Entity, 0, len(s.bees))
	for _, e := range s.bees {
		b := s.beeMap.Get(e)
		b.PersonalTime = s.constants.GetUpTime_s + util.RandomInt(s.rng, randomTimeToGetUp_s)
		list = append(list, e)
	}
	slices.Reverse(list) // fput
	slices.SortStableFunc(list, func(a, b ecs.Entity) int {
		ta, tb := s.beeMap.Get(a).PersonalTime, s.beeMap.Get(b).PersonalTime
		if ta < tb {
			return -1
		}
		if tb < ta {
			return 1
		}
		return 0
	})
	s.active.List = list
}

func (s *wakeUp) Finalize(w *ecs.World) {}

// isActiveAdult is true for an active adult worker or queen. NetLogo: (caste = "worker" or caste = "queen") and
// (activity != "hibernate") and stage = "adult", used in [MORN-08] and [GO-06].
func isActiveAdult(b *comp.Bee) bool {
	return (b.Caste == caste.Worker || b.Caste == caste.Queen) && b.Activity != activity.Hibernate && b.Stage == stage.Adult
}
