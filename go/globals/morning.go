package globals

import (
	"gobees/day"
)

// Calendar holds the day of the year and today's foraging period.
type Calendar struct {
	Day            day.OfYear // day of the year, 1 to 365, 0 in Setup
	ForagingPeriod float64    // length of today's foraging period [s], 0 until day 1. NetLogo: DailyForagingPeriod_s
}

// Flowering tells whether the flowering changed today, set by the flower refill in the morning.
type Flowering struct {
	Changed bool // true on days when a foodsource starts or stops flowering, read by the colony update. NetLogo: FoodsourcesInFlowerUpdate?
}

// BroodTotals are the daily totals of the brood, set in the morning by broodTotals. The seasonal events read
// Eggs + Larvae + Pupae on the next morning, as they run before broodTotals.
// NetLogo: TotalEggs, TotalLarvae, TotalPupae.
type BroodTotals struct {
	Eggs   int
	Larvae int
	Pupae  int
}
