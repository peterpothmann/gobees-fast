package globals

// Calendar holds the day of the year and today's foraging period [MORN-02], [MORN-04].
type Calendar struct {
	Day                   float64 // day of the year, 1 to 365, 0 in Setup
	DailyForagingPeriod_s float64 // length of today's foraging period [s], 0 until day 1 (finding 4)
}

// Weather holds the daily foraging period of the parameter Weather [FPER], set in Setup (sys.Parameters): one value
// for every day ("Constant N hrs"), or the period per day of the year from a weather file, for one or more years.
type Weather struct {
	Constant_s float64     // foraging period of every day [s], used when Years_s is nil
	Years_s    [][]float64 // foraging period [year-1][day-1] of a weather file [s], nil for a constant
}

// ForagingPeriod_s returns the foraging period [s] of the day of the year day (1 to 365) in the year year of the run
// (1, 2, ...). The years of a weather file repeat.
func (w *Weather) ForagingPeriod_s(year, day int) float64 {
	if w.Years_s == nil {
		return w.Constant_s
	}
	return w.Years_s[(year-1)%len(w.Years_s)][day-1]
}

// FoodsourcesToday holds the globals that UpdateFoodsourcesProc sets [UFS].
type FoodsourcesToday struct {
	NectarAvailableTotalToday_l  float64 // output only
	PollenAvailableTotalToday_kg float64 // output only
	FoodsourcesInFlowerUpdate    bool    // true on days when a foodsource starts or stops flowering, read by [UCOL-01]
}

// Totals are the daily totals that CheckNumbersProc sets in the morning [CHK-02] (finding 39, see
// trans/netlogo-findings.md#finding-39). UpdateSeasonalEventsProc reads TotalEggs + TotalLarvae + TotalPupae on
// the next morning [SEAS-02]. The others are output only.
type Totals struct {
	TotalEggs               float64
	TotalLarvae             float64
	TotalPupae              float64
	TotalAdultWorkers       float64
	TotalAdultQueens        float64 // all adult queens, also without colony and hibernating
	TotalAdultMales         float64
	TotalAdults             float64 // all adults, also without colony and hibernating
	TotalForagingTripsToday float64 // reset every morning, counted in ForagingProc (step 6)
}
