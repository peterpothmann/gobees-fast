package sys

import (
	"fmt"

	"gobees/day"
	"gobees/globals"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

// calendarUpdate sets the day of the year and today's foraging period.
//
// First part of the morning routine. The other parts follow in the group Morning, in this order (see schedule.go):
// flowerRefill, seasonalEvents, queenEmergence, wakeUp,
// cropRotation, colonyUpdate, broodTotals.
// Initialize runs it also at ticks = 0, i.e. Day = 0.
//
// NetLogo: UpdateMorning_Proc.
type calendarUpdate struct {
	ticks    *globals.Ticks
	weather  *params.ForagingPeriod
	calendar globals.Calendar
}

func (s *calendarUpdate) Initialize(w *ecs.World) {
	s.ticks = ecs.GetResource[globals.Ticks](w)
	s.weather = ecs.GetResource[params.ForagingPeriod](w)
	ecs.AddResource(w, &s.calendar)
	s.Update(w) // run at ticks = 0
}

func (s *calendarUpdate) Update(w *ecs.World) {
	s.calendar.Day = dayOfYear(s.ticks.Ticks)
	if s.ticks.Ticks > 0 && (s.calendar.Day < 1 || s.calendar.Day > 365) {
		panic(fmt.Sprintf("Error in calculation of Day! (calendarUpdate): ticks %d, Day %v", s.ticks.Ticks, s.calendar.Day))
	}
	// no foraging period on Day 0
	// Update the calendar with today's foraging period and the Day of the year.
	if s.calendar.Day > 0 {
		// year of the run: 1 for ticks 1 to 365, 2 for 366 to 730,...
		year := int((s.ticks.Ticks-1)/365) + 1
		s.calendar.ForagingPeriod = s.weather.Get(year, int(s.calendar.Day))
	}
}

func (s *calendarUpdate) Finalize(w *ecs.World) {}

// dayOfYear returns the day of the year, 1 to 365, and 0 for ticks = 0.
//
// NetLogo computes round (ticks mod 365.00000001). The integer formula gives the same result.
func dayOfYear(ticks day.Abs) day.OfYear {
	if ticks <= 0 {
		return 0
	}
	return day.OfYear((ticks-1)%365 + 1)
}
