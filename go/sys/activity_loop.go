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

// ActivityLoop [GO-06 to GO-10].
//
// Event-driven loop of the day. Repeatedly takes the active bee with the smallest personalTime_s
// [NEXT], lets it perform exactly one activity [ACT] and re-inserts it, until the smallest personalTime_s
// exceeds CallItaDay_s, no active adult worker or queen is left, or no colony is left.
//
// Foraging is not translated yet (step 6): see ForagingProc.
//
// NetLogo: while loop in Go, NextActiveBeeREP, ActivityProc, BroodIncubationProc.
type ActivityLoop struct {
	// CheckActiveCount makes the loop count the active adult workers and queens in every turn, as NetLogo does
	// in [GO-06], and panic if the number differs from the length of the queue. For tests.
	CheckActiveCount bool

	rng       *rand.Rand
	constants *params.Constants
	calendar  *globals.Calendar
	active    *globals.ActiveBees
	daytime   *globals.Daytime
	eggs      eggLayer

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.Colony]
	beeMap       *beeMapper
	colonyMap    *ecs.Map4[comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts]
	speciesMap   *ecs.Map1[comp.Species]

	// broodOf holds the brood of each colony for BroodIncubationProc. NetLogo: bees with [ colonyID = x and
	// (stage = "egg" or stage = "larva" or stage = "pupa") ]. Built at the start of the loop. During the day
	// brood neither develops nor dies, and new eggs are added when they are laid.
	broodOf map[ecs.Entity][]ecs.Entity
}

func (s *ActivityLoop) Initialize(w *ecs.World) {
	s.rng = rand.New(ecs.GetResource[resource.Rand](w))
	s.constants = ecs.GetResource[params.Constants](w)
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.active = ecs.GetResource[globals.ActiveBees](w)
	s.daytime = ecs.GetResource[globals.Daytime](w)
	s.eggs = newEggLayer(w, s.rng)

	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.Colony](w)
	s.beeMap = newBeeMapper(w)
	s.colonyMap = ecs.NewMap4[comp.Colony, comp.ColonyStores, comp.ColonyStimuli, comp.ColonyCounts](w)
	s.speciesMap = ecs.NewMap1[comp.Species](w)
	s.broodOf = map[ecs.Entity][]ecs.Entity{}
}

func (s *ActivityLoop) Update(w *ecs.World) {
	s.broodOf = beesByColony(s.beeFilter, func(b *comp.Bee) bool { return b.Stage != stage.Adult })

	// [GO-06] the queue holds exactly the active adult workers and queens ([NEXT-01]): all of them were put
	// into it in [MORN-08], no bee becomes one during the day, and DieProc removes dying bees from it. Its
	// length is therefore NetLogo's count, without counting all bees in every turn.
	continueWorking := true
	for continueWorking && s.activeAdults() > 0 && s.colonies() > 0 {
		// [GO-07]
		e := s.NextActiveBeeREP()
		_, _, b, _, _, _, _ := s.beeMap.Get(e)
		// [GO-08]
		if b.PersonalTime_s > s.constants.CallItaDay_s {
			continueWorking = false
		} else {
			// [GO-09] activityList: GUI log, not translated.
			s.daytime.Daytime_s = b.PersonalTime_s // Model has no global time, each bee has her own personalTime_s and the one with the lowest is next
			s.ActivityProc(w, e)
			if !w.Alive(e) {
				continue // died during her task (foraging, step 6): DieProc ends her code, [GO-10] is skipped (finding 28)
			}
		}
		// [GO-10] get the bee again: ActivityProc may have created entities
		_, _, b, _, _, _, _ = s.beeMap.Get(e)
		// Queens without a colony are inactive for the rest of the day.
		if b.ColonyID.IsZero() {
			b.PersonalTime_s = s.constants.CallItaDay_s + 1
		}
	}
}

func (s *ActivityLoop) Finalize(w *ecs.World) {}

// activeAdults returns the number of active adult workers and queens [GO-06].
func (s *ActivityLoop) activeAdults() int {
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
func (s *ActivityLoop) colonies() int {
	query := s.colonyFilter.Query()
	n := query.Count()
	query.Close()
	return n
}

// NextActiveBeeREP puts the first bee of the queue back at its sorted position and returns the new first bee
// [NEXT]. Ported line by line, including the tie rule [NEXT-04].
func (s *ActivityLoop) NextActiveBeeREP() ecs.Entity {
	time := func(e ecs.Entity) float64 {
		_, _, b, _, _, _, _ := s.beeMap.Get(e)
		return b.PersonalTime_s
	}
	list := s.active.List
	// [NEXT-01]
	activeBee := list[0]
	persTime_activeBee := time(activeBee)
	// [NEXT-02]
	minPosition := 0.0
	maxPosition := float64(len(list) - 1)
	currentPosition := util.Round(maxPosition / 2)
	for maxPosition-minPosition > 1 {
		if time(list[int(currentPosition)]) > persTime_activeBee {
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
	// [NEXT-04] a new slice: beginningList and endList share the memory of the old list.
	// SIMPLIFY: a priority queue (trans/data-model.md).
	newList := make([]ecs.Entity, 0, len(list))
	if len(endList) == 1 && persTime_activeBee > time(endList[0]) {
		newList = append(append(append(newList, beginningList...), endList...), activeBee)
	} else {
		newList = append(append(append(newList, beginningList...), activeBee), endList...)
	}
	s.active.List = newList
	return newList[0]
}

// ActivityProc lets the bee e choose one task and do it [ACT].
func (s *ActivityLoop) ActivityProc(w *ecs.World, e ecs.Entity) {
	// [ACT-01]
	break_s := 0.5 * 3600
	_, _, b, _, th, _, _ := s.beeMap.Get(e)
	b.Activity = activity.Resting
	if !b.ColonyID.IsZero() {
		// [ACT-02] ORDER of the four: irrelevant.
		colony, stores, stim, counts := s.colonyMap.Get(b.ColonyID)
		sp := s.speciesMap.Get(colony.SpeciesIDcolony)
		stim.StimEgglaying = StimEgglayingREP(stores, counts, sp)
		stim.StimNectarForaging = StimForagingNectarREP(stores, s.daytime.Daytime_s, s.constants.Sunrise_s, s.calendar.DailyForagingPeriod_s)
		stim.StimPollenForaging = StimForagingPollenREP(stores, sp, s.daytime.Daytime_s, s.constants.Sunrise_s, s.calendar.DailyForagingPeriod_s)
		stim.StimNursing = StimNursingREP(stores, counts, sp)

		// [ACT-03] the last matching task wins (finding 11, see trans/netlogo-findings.md#finding-11): no "else if".
		if stim.StimEgglaying > th.ThEgglaying {
			b.Activity = activity.Egglaying
		}
		if stim.StimNursing > th.ThNursing {
			b.Activity = activity.Nursing
		}
		if stim.StimPollenForaging > th.ThForagingPollen {
			b.Activity = activity.PollenForaging
		}
		if stim.StimNectarForaging > th.ThForagingNectar {
			b.Activity = activity.NectarForaging
		}
	}

	// [ACT-04]
	if b.SpeciesName == "Psithyrus" && (b.Activity == activity.Nursing || b.Activity == activity.PollenForaging || b.Activity == activity.NectarForaging) {
		b.Activity = activity.Resting
	}

	// [ACT-05] at most one task runs, see the annotation
	switch b.Activity {
	case activity.Resting:
		b.PersonalTime_s = b.PersonalTime_s + break_s
	case activity.Egglaying:
		col := b.ColonyID
		eggs := s.eggs.EgglayingProc(e)
		s.broodOf[col] = append(s.broodOf[col], eggs...)
	case activity.Nursing:
		s.BroodIncubationProc(w, e)
	case activity.PollenForaging, activity.NectarForaging:
		s.ForagingProc(w, e)
	}
}

// BroodIncubationProc lets the nurse e heat the brood of her colony for 48 min [INC].
func (s *ActivityLoop) BroodIncubationProc(w *ecs.World, e ecs.Entity) {
	heatingPeriod_s := 2880.0
	incubationEnergy_kJ_per_mg_s := 0.000000217013888
	_, _, b, _, _, _, _ := s.beeMap.Get(e)
	// [INC-01]
	heatProvided_kJ := heatingPeriod_s * incubationEnergy_kJ_per_mg_s * b.Weight_mg * b.Number
	// [INC-02] the divisor uses this morning's counts, brood laid today is not in it (finding 46, see
	// trans/netlogo-findings.md#finding-46)
	heatProvidedPerBrood_kJ := 0.0
	_, stores, _, counts := s.colonyMap.Get(b.ColonyID)
	if counts.AllEggs+counts.AllLarvae+counts.AllPupae > 0 {
		heatProvidedPerBrood_kJ = heatProvided_kJ / (counts.AllEggs + counts.AllLarvae + counts.AllPupae)
	}
	// ORDER: irrelevant, each brood agent only changes itself. Shuffled (decision C).
	brood := s.broodOf[b.ColonyID]
	util.Shuffle(s.rng, brood)
	for _, x := range brood {
		if !w.Alive(x) {
			continue
		}
		_, _, _, _, _, _, dev := s.beeMap.Get(x)
		dev.CumulIncubationReceived_kJ = dev.CumulIncubationReceived_kJ + heatProvidedPerBrood_kJ
	}
	stores.SummedIncubationToday_kJ = stores.SummedIncubationToday_kJ + heatProvided_kJ
	stores.EnergyStore_kJ = stores.EnergyStore_kJ - heatProvided_kJ
	// [INC-03]
	b.PersonalTime_s = b.PersonalTime_s + heatingPeriod_s
}

// ForagingProc is TEMPORARY until step 6 translates the foraging. The bee spends 30 min (like resting) and
// collects nothing, so the colony stores do not grow. Without any increase of personalTime_s the loop would not
// end.
//
// NetLogo: ForagingProc (step 6).
func (s *ActivityLoop) ForagingProc(w *ecs.World, e ecs.Entity) {
	_, _, b, _, _, _, _ := s.beeMap.Get(e)
	b.PersonalTime_s = b.PersonalTime_s + 0.5*3600
}
