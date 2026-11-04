package globals

// Calendar holds the day of the year and today's foraging period [MORN-02], [MORN-04].
type Calendar struct {
	Day                   float64 // day of the year, 1 to 365, 0 in Setup
	DailyForagingPeriod_s float64 // length of today's foraging period [s], 0 until day 1 (finding 4)
}

// Weather holds the daily foraging period [FPER] of the run file, set in Setup (sys.Parameters): the period per day of
// the year, for one or more years. Since step F7 the NetLogo options "Constant N hrs" are 365 equal values.
type Weather struct {
	Years_s [][]float64 // foraging period [year-1][day-1] [s]
}

// ForagingPeriod_s returns the foraging period [s] of the day of the year day (1 to 365) in the year year of the run
// (1, 2, ...). The years repeat.
func (w *Weather) ForagingPeriod_s(year, day int) float64 {
	return w.Years_s[(year-1)%len(w.Years_s)][day-1]
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
	TotalEggs   float64
	TotalLarvae float64
	TotalPupae  float64
}
