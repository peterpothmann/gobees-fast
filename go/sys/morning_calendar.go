package sys

import (
	"fmt"

	"gobees/day"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// calendarUpdate sets the day of the year and today's foraging period [MORN-01] to [MORN-04].
//
// First part of UpdateMorning_Proc. The other parts follow in the group Morning, in this order (see schedule.go):
// flowerRefill [MORN-05], seasonalEvents [MORN-06], queenEmergence [MORN-07], wakeUp [MORN-08],
// cropRotation [MORN-09], colonyUpdate [MORN-10], broodTotals [MORN-12].
// Initialize runs it also at ticks = 0, i.e. Day = 0 [INIT-10].
// Not translated: the assertion stop sign [MORN-01] (assertions panic where they occur) and Date [MORN-03]
// (display only).
//
// NetLogo: UpdateMorning_Proc.
type calendarUpdate struct {
	ticks    *globals.Ticks
	weather  *params.Weather
	calendar globals.Calendar
}

func (s *calendarUpdate) Initialize(w *ecs.World) {
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.weather = ecs.GetResource[params.Weather](w)
	ecs.AddResource(w, &s.calendar)
	s.Update(w) // run at ticks = 0 [INIT-10]
}

func (s *calendarUpdate) Update(w *ecs.World) {
	// [MORN-02] finding 36 (see trans/netlogo-findings.md#finding-36)
	s.calendar.Day = dayOfYear(s.ticks.Ticks)
	if s.ticks.Ticks > 0 && (s.calendar.Day < 1 || s.calendar.Day > 365) {
		panic(fmt.Sprintf("Error in calculation of Day! (Updates_Proc): ticks %d, Day %v", s.ticks.Ticks, s.calendar.Day))
	}
	// [MORN-04] finding 4, no foraging period on Day 0 (see trans/netlogo-findings.md#finding-4)
	// Update the calender with today's foraging period and the Day of the year.
	if s.calendar.Day > 0 {
		// [FPER] year of the run: 1 for ticks 1 to 365, 2 for 366 to 730, ...
		year := int((s.ticks.Ticks-1)/365) + 1
		s.calendar.ForagingPeriod = s.weather.ForagingPeriod_s(year, int(s.calendar.Day))
	}
}

func (s *calendarUpdate) Finalize(w *ecs.World) {}

// dayOfYear returns the day of the year, 1 to 365, and 0 for ticks = 0 [MORN-02].
//
// NetLogo: round (ticks mod 365.00000001). Computed with integers, which gives the same result
// (finding 36, see trans/netlogo-findings.md#finding-36).
func dayOfYear(ticks day.Abs) day.OfYear {
	if ticks <= 0 {
		return 0
	}
	return day.OfYear((ticks-1)%365 + 1)
}
