package sys

import (
	"fmt"

	"gobees/comp"
	"gobees/enum/caste"
	"gobees/enum/stage"
	"gobees/globals"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark/ecs"
)

// CheckNumbers checks the bee numbers and sets the daily totals [CHK].
//
// Not only a check: it sets globals.Totals and resets TotalForagingTripsToday (finding 39, see
// trans/netlogo-findings.md#finding-39). The assertions panic.
// Part of the morning routine [MORN-12]. Initialize runs it at ticks = 0 [INIT-10].
//
// NetLogo: CheckNumbersProc.
type CheckNumbers struct {
	// SkipColonyCheck switches off the colony check of [CHK-04]. TEMPORARY until step 7: the check needs all
	// colonies to be dead on day 364, which requires the end of season (UpdateSeasonalEventsProc, not yet
	// translated). model/default.go sets it.
	SkipColonyCheck bool

	calendar *globals.Calendar
	file     *params.ParameterFile
	derived  *globals.Derived
	counters *globals.Counters
	totals   globals.Totals

	beeFilter    *ecs.Filter1[comp.Bee]
	colonyFilter *ecs.Filter1[comp.ColonyCounts]
}

func (s *CheckNumbers) Initialize(w *ecs.World) {
	s.calendar = ecs.GetResource[globals.Calendar](w)
	s.file = ecs.GetResource[params.ParameterFile](w)
	s.derived = ecs.GetResource[globals.Derived](w)
	s.counters = ecs.GetResource[globals.Counters](w)
	ecs.AddResource(w, &s.totals)
	s.beeFilter = ecs.NewFilter1[comp.Bee](w)
	s.colonyFilter = ecs.NewFilter1[comp.ColonyCounts](w)
	s.Update(w)
}

// Update: ORDER irrelevant, only sums and checks.
func (s *CheckNumbers) Update(w *ecs.World) {
	// [CHK-01]
	totalInitalBees := 0.0
	for _, sq := range s.derived.BeeSpeciesInitialQueensList {
		totalInitalBees += util.Round(sq.Queens * s.file.InitialQueensModifierBehaviorSpace)
	}

	// sums over all bees
	var totalBeesPresent, eggs, larvae, pupae, adultWorkers, adultMales, adultQueens, adults float64
	query := s.beeFilter.Query()
	for query.Next() {
		b := query.Get()
		totalBeesPresent += b.Number
		switch b.Stage {
		case stage.Egg:
			eggs += b.Number
		case stage.Larva:
			larvae += b.Number
		case stage.Pupa:
			pupae += b.Number
		case stage.Adult:
			adults += b.Number
			switch b.Caste {
			case caste.Worker:
				adultWorkers += b.Number
			case caste.Male:
				adultMales += b.Number
			case caste.Queen:
				adultQueens += b.Number
			}
		}
	}
	if totalInitalBees+s.counters.TotalBeesEverProduced-s.counters.TotalBeesEverDied != totalBeesPresent {
		panic(fmt.Sprintf("Error in number of bees ever produced/died! (CheckNumbersProc): initial %v + produced %v - died %v != present %v",
			totalInitalBees, s.counters.TotalBeesEverProduced, s.counters.TotalBeesEverDied, totalBeesPresent))
	}

	// [CHK-02] finding 39 (see trans/netlogo-findings.md#finding-39)
	t := &s.totals
	t.TotalEggs, t.TotalLarvae, t.TotalPupae, t.TotalAdultWorkers, t.TotalAdultMales = 0, 0, 0, 0, 0
	colonies := s.colonyFilter.Query()
	for colonies.Next() {
		c := colonies.Get()
		t.TotalEggs += c.AllEggs
		t.TotalLarvae += c.AllLarvae
		t.TotalPupae += c.AllPupae
		t.TotalAdultWorkers += c.AllAdultWorkers
		t.TotalAdultMales += c.AllAdultMales
	}
	t.TotalAdultQueens = adultQueens
	t.TotalAdults = adults
	t.TotalForagingTripsToday = 0

	// [CHK-03] finding 38 (see trans/netlogo-findings.md#finding-38)
	if t.TotalEggs != eggs || t.TotalLarvae != larvae || t.TotalPupae != pupae ||
		t.TotalAdultWorkers != adultWorkers || t.TotalAdultMales != adultMales {
		panic(fmt.Sprintf("Error in total numbers of bees! (CheckNumbersProc): colonies %v/%v/%v/%v/%v, bees %v/%v/%v/%v/%v (eggs/larvae/pupae/workers/males)",
			t.TotalEggs, t.TotalLarvae, t.TotalPupae, t.TotalAdultWorkers, t.TotalAdultMales,
			eggs, larvae, pupae, adultWorkers, adultMales))
	}

	// [CHK-04] colony check. The egg accounting runs at the death of each colony (colonyDeath, decision E).
	if s.calendar.Day == 364 && !s.SkipColonyCheck {
		colCheck := s.counters.TotalColoniesEverProduced - (s.counters.ColonyDeathsEndSeason + s.counters.ColonyDeathsNoBees)
		if colCheck != 0 {
			panic(fmt.Sprintf("Colonies produced cannot be accounted for: %v colonies", colCheck))
		}
	}
}

func (s *CheckNumbers) Finalize(w *ecs.World) {}
