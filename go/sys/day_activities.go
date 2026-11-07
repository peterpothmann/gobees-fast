package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// activities is the event-driven loop of the day.
//
// It repeatedly takes the active bee with the smallest personalTime_s, lets it perform exactly one activity and
// re-inserts it, until the smallest personalTime_s exceeds Constants.CallItADay, no active adult worker or queen is
// left, or no colony is left.
//
// Foraging: day_foraging.go.
//
// NetLogo: while loop in Go, NextActiveBeeREP, ActivityProc, BroodIncubationProc.
type activities struct {
	// CheckActiveCount makes the loop count the active adult workers and queens in every turn, as NetLogo does,
	// and panic if the number differs from the length of the queue. For tests.
	CheckActiveCount bool

	rng       *rand.Rand
	constants *params.Constants
	calendar  *globals.Calendar
	active    *globals.ActiveBees
	daytime   *globals.Daytime
	hooks     *Hooks // nil without the resource (the window of --ui)
	eggs      eggLayer
	foraging  forager

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	beeMap       *beeMapper
	colonyMap    *ecs.Map4[comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts]
	speciesMap   *speciesTable

	// broodOf holds the brood (eggs, larvae and pupae) of each colony for incubateBrood. Built at the start of the
	// loop. During the day brood neither develops nor dies, and new eggs are added when they are laid.
	// NetLogo: bees with [ colonyID = x and (stage = "egg" or stage = "larva" or stage = "pupa") ].
	broodOf map[ecs.Entity][]ecs.Entity
}

func (s *activities) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.active = ecs.GetResource[globals.ActiveBees](w)
	s.daytime = ecs.GetResource[globals.Daytime](w)
	s.hooks = ecs.GetResource[Hooks](w)
	s.eggs = newEggLayer(w, s.rng)
	s.foraging = newForager(w, s.rng)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeMap = newBeeMapper(w)
	s.colonyMap = ecs.NewMap4[comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts](w)
	s.speciesMap = newSpeciesTable(w)
	s.broodOf = map[ecs.Entity][]ecs.Entity{}
}

func (s *activities) Update(w *ecs.World) {
	s.broodOf = beesByColony(s.broodOf, s.beeFilter, func(b *comp.Bee) bool { return b.Stage != stage.Adult })

	// the queue holds exactly the active adult workers and queens: wakeUp put all of them into it, no bee becomes
	// one during the day, and beeDeath.die removes dying bees from it. Its length is therefore NetLogo's count,
	// without counting all bees in every turn.
	continueWorking := true
	for continueWorking && s.activeAdults() > 0 && s.colonies() > 0 {
		e := s.nextActiveBee()
		_, b, _, _, _, _ := s.beeMap.Get(e)
		if b.PersonalTime > s.constants.Activities.CallItADay {
			continueWorking = false
		} else {
			s.daytime.Time = b.PersonalTime // Model has no global time, each bee has her own personalTime_s and the one with the lowest is next
			s.act(w, e)
			if s.hooks != nil {
				s.hooks.AfterTurn(e)
			}
			if !w.Alive(e) {
				continue // died during her task (foraging): the rest of her turn is skipped
			}
		}
		// get the bee again: act may have created entities
		_, b, _, _, _, _ = s.beeMap.Get(e)
		// Queens without a colony are inactive for the rest of the day.
		if b.Colony.IsZero() {
			b.PersonalTime = s.constants.Activities.CallItADay + 1
		}
	}
}

func (s *activities) Finalize(w *ecs.World) {}

// activeAdults returns the number of active adult workers and queens.
func (s *activities) activeAdults() int {
	n := len(s.active.List)
	if s.CheckActiveCount {
		count := 0
		query := s.beeFilter.Query()
		for query.Next() {
			if isActiveAdult(query.Get()) {
				count++
			}
		}
		if count != n {
			panic(fmt.Sprintf("%d active adult workers and queens, but %d bees in the activity queue", count, n))
		}
	}
	return n
}

// colonies returns the number of colonies.
func (s *activities) colonies() int {
	query := s.colonyFilter.Query()
	n := query.Count()
	query.Close()
	return n
}

// nextActiveBee puts the first bee of the queue back at its sorted position and returns the new first bee.
// The binary search follows NetLogo line by line, including the tie rule, so bees with equal times keep
// NetLogo's order.
func (s *activities) nextActiveBee() ecs.Entity {
	time := func(e ecs.Entity) float64 {
		_, b, _, _, _, _ := s.beeMap.Get(e)
		return b.PersonalTime
	}
	list := s.active.List
	activeBee := list[0]
	activeBeeTime := time(activeBee) // [s]
	minPosition := 0.0
	maxPosition := float64(len(list) - 1)
	currentPosition := round(maxPosition / 2)
	for maxPosition-minPosition > 1 {
		if time(list[int(currentPosition)]) > activeBeeTime {
			maxPosition = currentPosition
			currentPosition = round((currentPosition + minPosition) / 2)
		} else {
			minPosition = currentPosition
			currentPosition = round((maxPosition + currentPosition) / 2)
		}
	}
	beginningList := list[:int(currentPosition)]
	endList := list[int(currentPosition):]
	if len(list) > 1 {
		beginningList = beginningList[1:]
	} else {
		endList = endList[1:]
	}
	// in place, without a new list: beginningList moves one position to the front into the place
	// of activeBee, endList stays where it is. Same order as NetLogo's sentence.
	// SIMPLIFY: a priority queue.
	if len(endList) == 1 && activeBeeTime > time(endList[0]) {
		copy(list, list[1:])
		list[len(list)-1] = activeBee
	} else {
		copy(list, beginningList)
		list[len(beginningList)] = activeBee
	}
	return list[0]
}

// act lets the bee e choose one task and do it.
func (s *activities) act(w *ecs.World, e ecs.Entity) {
	pause := s.constants.Activities.Pause // [s]
	_, b, _, th, _, _ := s.beeMap.Get(e)
	b.Activity = activity.Resting
	if !b.Colony.IsZero() {
		// ORDER of the four: irrelevant.
		colony, stores, stim, counts := s.colonyMap.Get(b.Colony)
		sp := s.speciesMap.Get(colony.Species)
		stim.Egglaying = egglayingStimulus(stores, counts, sp)
		stim.NectarForaging = nectarStimulus(stores, &s.constants.Stimuli, s.daytime.Time, s.constants.Activities.Sunrise,
			s.calendar.ForagingPeriod)
		stim.PollenForaging = pollenStimulus(stores, sp, &s.constants.Stimuli, s.daytime.Time, s.constants.Activities.Sunrise,
			s.calendar.ForagingPeriod)
		stim.Nursing = nursingStimulus(stores, counts, sp)

		// the last matching task wins: no "else if".
		if stim.Egglaying > th.Egglaying {
			b.Activity = activity.Egglaying
		}
		if stim.Nursing > th.Nursing {
			b.Activity = activity.Nursing
		}
		if stim.PollenForaging > th.PollenForaging {
			b.Activity = activity.PollenForaging
		}
		if stim.NectarForaging > th.NectarForaging {
			b.Activity = activity.NectarForaging
		}
	}

	// a bee of a parasitic species (Psithyrus) never nurses or forages. The task is tested first, so the
	// species is looked up only for nursing and foraging (no side effects, same result).
	if (b.Activity == activity.Nursing || b.Activity == activity.PollenForaging || b.Activity == activity.NectarForaging) &&
		s.speciesMap.Get(b.Species).Parasitic {
		b.Activity = activity.Resting
	}

	// switch: at most one task runs per turn
	switch b.Activity {
	case activity.Resting:
		b.PersonalTime = b.PersonalTime + pause
	case activity.Egglaying:
		col := b.Colony
		eggs := s.eggs.layEggs(e)
		s.broodOf[col] = append(s.broodOf[col], eggs...)
	case activity.Nursing:
		s.incubateBrood(w, e)
	case activity.PollenForaging, activity.NectarForaging:
		s.foraging.forage(w, e)
	}
}

// incubateBrood lets the nurse e heat the brood of her colony for 48 min.
func (s *activities) incubateBrood(w *ecs.World, e ecs.Entity) {
	heatingPeriod := s.constants.Activities.NursingTime         // [s]
	incubationEnergy := s.constants.Activities.IncubationEnergy // [kJ/(mg s)]
	_, b, _, _, _, _ := s.beeMap.Get(e)
	heatProvided := heatingPeriod * incubationEnergy * b.Weight * float64(b.Number) // [kJ]
	// the divisor uses this morning's counts, brood laid today is not in it
	heatPerBrood := 0.0 // [kJ]
	_, stores, _, counts := s.colonyMap.Get(b.Colony)
	if counts.Eggs+counts.Larvae+counts.Pupae > 0 {
		heatPerBrood = heatProvided / float64(counts.Eggs+counts.Larvae+counts.Pupae)
	}
	// ORDER: irrelevant, each brood agent only changes itself. Shuffled, as NetLogo asks the agents in random order.
	brood := s.broodOf[b.Colony]
	shuffle(s.rng, brood)
	for _, x := range brood {
		if !w.Alive(x) {
			continue
		}
		_, _, _, _, _, dev := s.beeMap.Get(x)
		dev.IncubationReceived = dev.IncubationReceived + heatPerBrood
	}
	stores.IncubationToday = stores.IncubationToday + heatProvided
	stores.Energy = stores.Energy - heatProvided
	b.PersonalTime = b.PersonalTime + heatingPeriod
}
