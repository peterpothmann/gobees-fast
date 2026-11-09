package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// constantForagingPeriods are the "Constant N hrs" options of Weather [FPER-03], in seconds.
var constantForagingPeriods = map[string]float64{
	"Constant 24 hrs": 86400,
	"Constant 20 hrs": 72000,
	"Constant 16 hrs": 57600,
	"Constant 12 hrs": 43200,
	"Constant 11 hrs": 39600,
	"Constant 10 hrs": 36000,
	"Constant 9 hrs":  32400,
	"Constant 8 hrs":  28800,
	"Constant 7 hrs":  25200,
	"Constant 6 hrs":  21600,
	"Constant 5 hrs":  18000,
	"Constant 4 hrs":  14400,
	"Constant 3 hrs":  10800,
	"Constant 2 hrs":  7200,
	"Constant 1 hrs":  3600,
}

// netLogoWeatherFiles are the weather files of the NetLogo options of Weather whose values NetLogo has in the code.
//
// NetLogo has the daily foraging hours of one year in the code (Weather = "foragingHoursExample", [FPER-02]).
// The converter reads them from a file instead: netlogo/_SYSTEM_Weather_foragingHoursExample.csv holds the same
// values in seconds. The NetLogo name stays valid, so that the parameter files of NetLogo convert unchanged (e.g. the
// scenario step8_weather of go/compare).
var netLogoWeatherFiles = map[string]string{
	"foragingHoursExample": "_SYSTEM_Weather_foragingHoursExample.csv",
}

// weatherFromParameter returns the foraging period per year and day [s] of the parameter Weather [FPER]: one of
// the "Constant N hrs" options (365 equal values), a NetLogo option with a weather file (netLogoWeatherFiles), or
// otherwise the name of a weather file in dir (readWeatherFile). An unknown Weather is an error, NetLogo gives a
// foraging period of -1 [FPER-01].
func weatherFromParameter(weather, dir string) ([][]float64, error) {
	if period, ok := constantForagingPeriods[weather]; ok {
		year := make([]float64, 365)
		for d := range year {
			year[d] = period
		}
		return [][]float64{year}, nil
	}
	if file, ok := netLogoWeatherFiles[weather]; ok {
		weather = file
	}
	return readWeatherFile(filepath.Join(dir, weather))
}

// readWeatherFile reads a weather file: the foraging period of each day of the year, for one or more years.
//
// The file is a CSV table with a header row and the columns Day (1 to 365) and ForagingPeriod_s (0 to 86400), and
// optionally Year (1, 2, ...). Without Year, the file holds one year, which repeats in every year of the run. With
// Year, the years of the file follow each other and then repeat. Every year needs every day 1 to 365 once, in any
// order.
func readWeatherFile(path string) ([][]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("weather file: %w", err)
	}
	defer f.Close()
	years, err := parseWeatherFile(f)
	if err != nil {
		return nil, fmt.Errorf("weather file %s: %w", path, err)
	}
	return years, nil
}

// parseWeatherFile parses a weather file (readWeatherFile) and returns the foraging period [s] per year and day,
// [year-1][day-1].
func parseWeatherFile(r io.Reader) ([][]float64, error) {
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("expected a header row and at least one value row")
	}
	yearCol, dayCol, periodCol := -1, -1, -1
	for i, name := range rows[0] {
		switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, string(rune(0xFEFF))))) {
		case "year":
			yearCol = i
		case "day":
			dayCol = i
		case "foragingperiod_s":
			periodCol = i
		default:
			return nil, fmt.Errorf("unknown column %q (expected Year, Day and ForagingPeriod_s)", name)
		}
	}
	if dayCol < 0 || periodCol < 0 {
		return nil, fmt.Errorf("columns Day and ForagingPeriod_s are needed")
	}

	type value struct {
		set    bool
		period float64 // [s]
	}
	years := [][]value{}
	for i, row := range rows[1:] {
		line := i + 2
		number := func(col int) (float64, error) {
			v, err := strconv.ParseFloat(strings.TrimSpace(row[col]), 64)
			if err != nil {
				return 0, fmt.Errorf("line %d, column %s: %w", line, rows[0][col], err)
			}
			return v, nil
		}
		year := 1.0
		if yearCol >= 0 {
			if year, err = number(yearCol); err != nil {
				return nil, err
			}
		}
		day, err := number(dayCol)
		if err != nil {
			return nil, err
		}
		period, err := number(periodCol) // [s]
		if err != nil {
			return nil, err
		}
		if year < 1 || year != float64(int(year)) {
			return nil, fmt.Errorf("line %d: year %v, expected 1, 2, ...", line, year)
		}
		if day < 1 || day > 365 || day != float64(int(day)) {
			return nil, fmt.Errorf("line %d: day %v, expected 1 to 365", line, day)
		}
		if period < 0 || period > 86400 {
			return nil, fmt.Errorf("line %d: foraging period %v s, expected 0 to 86400", line, period)
		}
		for len(years) < int(year) {
			years = append(years, make([]value, 365))
		}
		v := &years[int(year)-1][int(day)-1]
		if v.set {
			return nil, fmt.Errorf("line %d: year %v, day %v given twice", line, year, day)
		}
		*v = value{set: true, period: period}
	}

	result := make([][]float64, len(years))
	for y, days := range years {
		result[y] = make([]float64, 365)
		for d, v := range days {
			if !v.set {
				return nil, fmt.Errorf("year %d, day %d is missing", y+1, d+1)
			}
			result[y][d] = v.period
		}
	}
	return result, nil
}
