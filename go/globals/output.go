package globals

// DailyOutput holds the daily totals that OutputDailyProc sets every evening [OUT-01] and once in Setup [INIT-12].
//
// Read by the model: TotalHibernatingQueens, TotalQueens and TotalMales, on the next morning in
// UpdateSeasonalEventsProc [SEAS-02] (TotalHibernatingQueens also in [SEAS-03]). The others are output only.
// SIMPLIFY: the output-only totals are not needed for the model.
type DailyOutput struct {
	TotalQueens            float64 // all queens, all stages (queen brood too)
	TotalMatedQueens       float64
	TotalUnmatedQueens     float64
	TotalHibernatingQueens float64 // all bees with activity "hibernate" (only queens hibernate)
	TotalColonies          float64
	TotalBeeAgents         float64 // number of bee agents (cohorts count once)
	TotalMales             float64 // all males, all stages (male brood too)
	TotalActiveBees        float64 // length of ActiveBeesSortedList after the event loop
	MeanWorkerWeight_mg    float64 // 0 without adult workers
	MeanQueenWeight_mg     float64 // 0 without adult queens
	MeanAdultWeight_mg     float64 // 0 without adults
	TotalMapArea_km2       float64 // set once [OUT-02], [AREA]
	ColonyDensity_km2      float64 // [OUT-02]
}

// Extinction holds the output of the extinction check [GO-20]. Output only (BehaviorSpace).
// SIMPLIFY: not needed for the model.
type Extinction struct {
	Survived         float64 // Survived?: 1 from Setup on [PAR-08], 0 once no colony and no bee is left
	TimeToExtinction float64 // ticks of the first day without colonies and bees, 0 before
}
