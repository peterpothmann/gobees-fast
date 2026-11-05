package globals

import (
	"gobees/day"
)

// Calendar holds the day of the year and today's foraging period [MORN-02], [MORN-04].
type Calendar struct {
	Day                   day.OfYear // day of the year, 1 to 365, 0 in Setup
	DailyForagingPeriod_s float64    // length of today's foraging period [s], 0 until day 1 (finding 4)
}

// FoodsourcesToday holds the globals that UpdateFoodsourcesProc sets [UFS].
// NectarAvailableTotalToday_l and PollenAvailableTotalToday_kg: output only, removed in step F4. The NetLogo
// comparison of step 4a sums them itself.
type FoodsourcesToday struct {
	FoodsourcesInFlowerUpdate bool // true on days when a foodsource starts or stops flowering, read by [UCOL-01]
}

// Totals are the daily totals of the brood that CheckNumbersProc sets in the morning [CHK-02] (finding 39, see
// trans/netlogo-findings.md#finding-39). UpdateSeasonalEventsProc reads TotalEggs + TotalLarvae + TotalPupae on
// the next morning [SEAS-02].
// TotalAdultWorkers, TotalAdultMales (read only by the check [CHK-03]) and TotalForagingTripsToday (output only,
// counted by the observers from event.ForagingTrip): removed in step F4.
type Totals struct {
	TotalEggs   int
	TotalLarvae int
	TotalPupae  int
}
