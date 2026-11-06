package sys

import (
	"fmt"
	"math/rand/v2"

	"gobees/comp"
	"gobees/enum/activity"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

// activities [GO-06 to GO-10].
//
// Event-driven loop of the day. Repeatedly takes the active bee with the smallest personalTime_s
// [NEXT], lets it perform exactly one activity [ACT] and re-inserts it, until the smallest personalTime_s
// exceeds Constants.CallItADay, no active adult worker or queen is left, or no colony is left.
//
// Foraging: sys/foraging.go [FOR].
//
// NetLogo: while loop in Go, NextActiveBeeREP, ActivityProc, BroodIncubationProc.
type activities struct {
	// CheckActiveCount makes the loop count the active adult workers and queens in every turn, as NetLogo does
	// in [GO-06], and panic if the number differs from the length of the queue. For tests.
	CheckActiveCount bool

	rng       *rand.Rand
	constants *params.Constants
	calendar  *globals.Calendar
	active    *globals.ActiveBees
	daytime   *globals.Daytime
	eggs      eggLayer
	foraging  forager

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	beeMap       *beeMapper
	colonyMap    *ecs.Map4[comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts]
	speciesMap   *speciesTable

	// broodOf holds the brood of each colony for BroodIncubationProc. NetLogo: bees with [ colonyID = x and
	// (stage = "egg" or stage = "larva" or stage = "pupa") ]. Built at the start of the loop. During the day
	// brood neither develops nor dies, and new eggs are added when they are laid.
	broodOf map[ecs.Entity][]ecs.Entity
}

func (s *activities) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.active = ecs.GetResource[globals.ActiveBees](w)
	s.daytime = ecs.GetResource[globals.Daytime](w)
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

	// [GO-06] the queue holds exactly the active adult workers and queens ([NEXT-01]): all of them were put
	// into it in [MORN-08], no bee becomes one during the day, and DieProc removes dying bees from it. Its
	// length is therefore NetLogo's count, without counting all bees in every turn.
	continueWorking := true
	for continueWorking && s.activeAdults() > 0 && s.colonies() > 0 {
		// [GO-07]
		e := s.nextActiveBee()
		_, b, _, _, _, _ := s.beeMap.Get(e)
		// [GO-08]
		if b.PersonalTime > s.constants.CallItADay {
			continueWorking = false
		} else {
			// [GO-09] activityList: GUI log, not translated.
			s.daytime.Time = b.PersonalTime // Model has no global time, each bee has her own personalTime_s and the one with the lowest is next
			s.act(w, e)
			if !w.Alive(e) {
				continue // died during her task (foraging, step 6): DieProc ends her code, [GO-10] is skipped (finding 28)
			}
		}
		// [GO-10] get the bee again: ActivityProc may have created entities
		_, b, _, _, _, _ = s.beeMap.Get(e)
		// Queens without a colony are inactive for the rest of the day.
		if b.Colony.IsZero() {
			b.PersonalTime = s.constants.CallItADay + 1
		}
	}
}

func (s *activities) Finalize(w *ecs.World) {}

// activeAdults returns the number of active adult workers and queens [GO-06].
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

// colonies returns the number of colonies [GO-06].
func (s *activities) colonies() int {
	query := s.colonyFilter.Query()
	n := query.Count()
	query.Close()
	return n
}

// nextActiveBee puts the first bee of the queue back at its sorted position and returns the new first bee
// [NEXT]. Ported line by line, including the tie rule [NEXT-04].
func (s *activities) nextActiveBee() ecs.Entity {
	time := func(e ecs.Entity) float64 {
		_, b, _, _, _, _ := s.beeMap.Get(e)
		return b.PersonalTime
	}
	list := s.active.List
	// [NEXT-01]
	activeBee := list[0]
	activeBeeTime := time(activeBee) // [s]
	// [NEXT-02]
	minPosition := 0.0
	maxPosition := float64(len(list) - 1)
	currentPosition := util.Round(maxPosition / 2)
	for maxPosition-minPosition > 1 {
		if time(list[int(currentPosition)]) > activeBeeTime {
			maxPosition = currentPosition
			currentPosition = util.Round((currentPosition + minPosition) / 2)
		} else {
			minPosition = currentPosition
			currentPosition = util.Round((maxPosition + currentPosition) / 2)
		}
	}
	beginningList := list[:int(currentPosition)]
	endList := list[int(currentPosition):]
	// [NEXT-03]
	if len(list) > 1 {
		beginningList = beginningList[1:]
	} else {
		endList = endList[1:]
	}
	// [NEXT-04] in place, without a new list (step F10): beginningList moves one position to the front into the place
	// of activeBee, endList stays where it is. Same order as NetLogo's sentence.
	// SIMPLIFY: a priority queue (trans/data-model.md).
	if len(endList) == 1 && activeBeeTime > time(endList[0]) {
		copy(list, list[1:])
		list[len(list)-1] = activeBee
	} else {
		copy(list, beginningList)
		list[len(beginningList)] = activeBee
	}
	return list[0]
}

// act lets the bee e choose one task and do it [ACT].
func (s *activities) act(w *ecs.World, e ecs.Entity) {
	// [ACT-01]
	pause := 0.5 * 3600 // [s]
	_, b, _, th, _, _ := s.beeMap.Get(e)
	b.Activity = activity.Resting
	if !b.Colony.IsZero() {
		// [ACT-02] ORDER of the four: irrelevant.
		colony, stores, stim, counts := s.colonyMap.Get(b.Colony)
		sp := s.speciesMap.Get(colony.Species)
		stim.Egglaying = egglayingStimulus(stores, counts, sp)
		stim.NectarForaging = nectarStimulus(stores, s.daytime.Time, s.constants.Sunrise, s.calendar.ForagingPeriod)
		stim.PollenForaging = pollenStimulus(stores, sp, s.daytime.Time, s.constants.Sunrise, s.calendar.ForagingPeriod)
		stim.Nursing = nursingStimulus(stores, counts, sp)

		// [ACT-03] the last matching task wins (finding 11, see trans/netlogo-findings.md#finding-11): no "else if".
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

	// [ACT-04] speciesName = "Psithyrus": the flag of the bee's species (step F9). The task is tested first, so the
	// species is looked up only for nursing and foraging (no side effects, same result).
	if (b.Activity == activity.Nursing || b.Activity == activity.PollenForaging || b.Activity == activity.NectarForaging) &&
		s.speciesMap.Get(b.Species).Parasitic {
		b.Activity = activity.Resting
	}

	// [ACT-05] at most one task runs, see the annotation
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

// incubateBrood lets the nurse e heat the brood of her colony for 48 min [INC].
func (s *activities) incubateBrood(w *ecs.World, e ecs.Entity) {
	heatingPeriod := 2880.0               // [s]
	incubationEnergy := 0.000000217013888 // [kJ/(mg s)]
	_, b, _, _, _, _ := s.beeMap.Get(e)
	// [INC-01]
	heatProvided := heatingPeriod * incubationEnergy * b.Weight * float64(b.Number) // [kJ]
	// [INC-02] the divisor uses this morning's counts, brood laid today is not in it (finding 46, see
	// trans/netlogo-findings.md#finding-46)
	heatPerBrood := 0.0 // [kJ]
	_, stores, _, counts := s.colonyMap.Get(b.Colony)
	if counts.Eggs+counts.Larvae+counts.Pupae > 0 {
		heatPerBrood = heatProvided / float64(counts.Eggs+counts.Larvae+counts.Pupae)
	}
	// ORDER: irrelevant, each brood agent only changes itself. Shuffled (decision C).
	brood := s.broodOf[b.Colony]
	util.Shuffle(s.rng, brood)
	for _, x := range brood {
		if !w.Alive(x) {
			continue
		}
		_, _, _, _, _, dev := s.beeMap.Get(x)
		dev.IncubationReceived = dev.IncubationReceived + heatPerBrood
	}
	stores.IncubationToday = stores.IncubationToday + heatProvided
	stores.Energy = stores.Energy - heatProvided
	// [INC-03]
	b.PersonalTime = b.PersonalTime + heatingPeriod
}
