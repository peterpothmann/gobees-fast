package util

// ForagingPeriod returns the length of today's foraging period [s] for the parameter Weather [FPER].
//
// day is NetLogo's Day (1 to 365). An unknown Weather gives -1 [FPER-01].
// NetLogo calls the reporter only for Day > 0, see [INIT-10].
//
// NetLogo: Foraging_periodREP.
func ForagingPeriod(weather string, day int) float64 {
	foragingPeriod_s := -1.0
	// [FPER-02]
	if weather == "foragingHoursExample" {
		foragingPeriod_s = foragingHoursExample[day-1] * 3600
	}
	// [FPER-03]
	if hours, ok := constantForagingHours[weather]; ok {
		foragingPeriod_s = hours * 3600
	}
	return foragingPeriod_s
}

// constantForagingHours are the "Constant N hrs" options of Weather [FPER-03].
var constantForagingHours = map[string]float64{
	"Constant 24 hrs": 24,
	"Constant 20 hrs": 20,
	"Constant 16 hrs": 16,
	"Constant 12 hrs": 12,
	"Constant 11 hrs": 11,
	"Constant 10 hrs": 10,
	"Constant 9 hrs":  9,
	"Constant 8 hrs":  8,
	"Constant 7 hrs":  7,
	"Constant 6 hrs":  6,
	"Constant 5 hrs":  5,
	"Constant 4 hrs":  4,
	"Constant 3 hrs":  3,
	"Constant 2 hrs":  2,
	"Constant 1 hrs":  1,
}

// foragingHoursExample are the daily foraging hours for Weather = "foragingHoursExample" [FPER-02].
//
// Copied token by token from the NetLogo code. The list has 366 items: item 42 is ".5" after a "0",
// probably meant as "0.5". Kept as it is, see finding 1 in trans/netlogo-findings.md#finding-1.
var foragingHoursExample = []float64{
	0, 3.1, 0, 0, 0, 1.5, 0, 0.1, 0, 0, 1.7, 1.6, 0, 0, 0, 0, 0, 0, 0, 0,
	1.5, 5, 0, 3.2, 0, 0, 0, 0.2, 0, 0, 0, 0.1, 0.9, 5.9, 3.5, 6.9, 1.3, 7.7, 2.3, 4.6,
	2.2, 0, .5, 9.2, 0, 8, 3.2, 4.1, 0, 9, 9.1, 7.3, 5.7, 4.9, 0, 12.1, 6.5, 7.9, 7.9, 11.1,
	2.8, 0, 2.8, 6, 5.7, 0, 4, 10.1, 2.9, 10.1, 0, 11.4, 6.3, 9.9, 4.4, 7.5, 8, 12.3, 8.7, 10.3,
	3.7, 11.3, 13.2, 14, 4.2, 7.7, 8.2, 7.2, 9.2, 5, 13.1, 10.5, 3.5, 11.1, 13.6, 6.2, 8.4, 7.8, 8.5, 9.8,
	6.5, 4.1, 10.8, 12.5, 15.1, 10.1, 4.3, 7, 9.4, 8.9, 7.5, 7.8, 6.6, 11.4, 12.1, 12.4, 11.9, 10.1, 14.7, 7.8,
	13.1, 3.3, 16.6, 14.8, 17.9, 5.7, 0.2, 2.9, 10, 14.7, 16.2, 15.8, 5.3, 5.8, 2.5, 6, 15.2, 1.3, 13.1, 11.2,
	2, 12.9, 9.7, 2.1, 17.3, 5.7, 8.5, 13.1, 18.5, 1.7, 6.7, 13.8, 0.5, 0.8, 15.7, 4.9, 11.4, 11.9, 3.8, 11.7,
	7.1, 21.2, 17.7, 1.8, 12.3, 15.7, 16.9, 16.8, 9.9, 3.6, 20.4, 13, 5.1, 0.6, 11.7, 2.1, 4.7, 13.9, 13.8, 1.4,
	0.3, 18.4, 14.8, 12.8, 3.7, 13.5, 4.7, 0.3, 5.5, 4, 17.5, 1.7, 0.3, 14.9, 12.4, 11.6, 8.5, 4.5, 11.1, 16,
	13.2, 13.8, 0.7, 7.1, 14.3, 3.4, 2.2, 5.6, 10.6, 3.4, 15.5, 15.6, 12.8, 15, 14, 5.9, 15.5, 9.1, 2, 1,
	3.2, 9.3, 3, 3.1, 14, 10.2, 1, 9.7, 8.8, 3.8, 1.9, 11.9, 9.3, 6.5, 6.6, 8.4, 4.3, 7.2, 1.5, 11.4,
	10.4, 13.5, 1.2, 6, 4.4, 13.5, 12.4, 8, 9.3, 5.9, 0.9, 6.8, 5.9, 9.1, 10.5, 6, 7.9, 2.3, 0.8, 0,
	7.9, 11, 1.3, 8.7, 6.5, 6.6, 7.6, 0, 0, 9.4, 7.1, 6.4, 4, 6.6, 0, 2.7, 0, 0, 7.8, 0,
	8.7, 0.3, 2, 4.8, 1.8, 0.9, 0, 0, 7.2, 5.8, 6.5, 0, 1.1, 0, 0, 0.7, 6.3, 1.3, 0, 5.5,
	1.4, 2.8, 0, 0, 0, 4, 0, 1.4, 5.1, 0, 0, 2.1, 0, 0.5, 0, 1, 0, 0, 2.3, 0,
	0, 0, 1.4, 0.6, 0, 0, 0, 0, 0.8, 0, 0, 1, 0.9, 0, 0, 0, 0, 0, 0, 2.3,
	0, 0, 1.9, 1.4, 0, 0, 0, 1.5, 0, 0, 0, 1, 1.9, 0, 0, 3.4, 0, 0, 1, 0,
	0, 0, 0, 0, 0, 1.6,
}
