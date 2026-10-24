package globals

// Calendar holds the day of the year and today's foraging period [MORN-02], [MORN-04].
type Calendar struct {
	Day                   float64 // day of the year, 1 to 365, 0 in Setup
	DailyForagingPeriod_s float64 // length of today's foraging period [s], 0 until day 1 (finding 4)
}

// FoodsourcesToday holds the globals that UpdateFoodsourcesProc sets [UFS].
type FoodsourcesToday struct {
	NectarAvailableTotalToday_l  float64 // output only
	PollenAvailableTotalToday_kg float64 // output only
	FoodsourcesInFlowerUpdate    bool    // true on days when a foodsource starts or stops flowering, read by [UCOL-01]
}

// Totals are the daily totals that CheckNumbersProc sets in the morning [CHK-02] (finding 39, see
// trans/netlogo-findings.md#finding-39). UpdateSeasonalEventsProc reads TotalEggs + TotalLarvae + TotalPupae on
// the next morning (step 7). The others are output only.
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
