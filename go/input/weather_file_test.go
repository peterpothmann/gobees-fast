package input

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gobees/globals"
	"gobees/internal/examplefiles"
)

// netLogoForagingHoursExample returns the list foragingHoursExample of Foraging_periodREP [FPER-02] from the NetLogo
// code, item by item.
func netLogoForagingHoursExample(t *testing.T) []float64 {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "netlogo", "__BEE-STEWARD_RC_ColourVersionMB.nlogo"))
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(b), "let foragingHoursExample [")
	if !ok {
		t.Fatal("no list foragingHoursExample in the NetLogo code")
	}
	list, _, _ := strings.Cut(rest, "]")
	result := []float64{}
	for _, token := range strings.Fields(list) {
		v, err := strconv.ParseFloat(token, 64)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, v)
	}
	return result
}

func TestParseWeatherFileExample(t *testing.T) {
	// [FPER-02] the example weather year holds item 0 to 364 of the NetLogo list, which has 366 items: ".5" after a
	// "0" is item 42, so day 43 gets 0.5 h (finding 1, see trans/netlogo-findings.md#finding-1). NetLogo reads item
	// Day - 1 with Day <= 365, item 365 never. The file holds hours * 3600 in whole seconds: on 15 days NetLogo's
	// product differs in the last bit (e.g. 16.9 * 3600 = 60839.99999999999). go/compare rounds the output column
	// foraging_period_s for this reason (RoundedColumns in go/compare/compare.go).
	netlogo := netLogoForagingHoursExample(t)
	if len(netlogo) != 366 || netlogo[41] != 0 || netlogo[42] != 0.5 {
		t.Fatalf("NetLogo list: %d items, items 41 and 42 = %v, %v; want 366, 0, 0.5", len(netlogo), netlogo[41],
			netlogo[42])
	}
	for _, ending := range []string{"\n", "\r\n"} {
		text := strings.ReplaceAll(examplefiles.WeatherForagingHoursExample, "\n", ending)
		years, err := parseWeatherFile(strings.NewReader(text))
		if err != nil {
			t.Fatal(err)
		}
		if len(years) != 1 {
			t.Fatalf("%d years, want 1", len(years))
		}
		lastBit := 0
		for day := 1; day <= 365; day++ {
			product := netlogo[day-1] * 3600
			if years[0][day-1] != math.Round(product) {
				t.Errorf("day %d: %v s, NetLogo item %d: %v h", day, years[0][day-1], day-1, netlogo[day-1])
			}
			if years[0][day-1] != product {
				lastBit++
			}
		}
		if lastBit != 15 {
			t.Errorf("%d days differ from NetLogo's product in the last bit, want 15", lastBit)
		}
	}
}

// weatherFile returns a weather file with the header and one row per year and day, period(year, day) seconds.
func weatherFile(header string, years int, period func(year, day int) float64) string {
	var b strings.Builder
	b.WriteString(header + "\n")
	for year := 1; year <= years; year++ {
		for day := 1; day <= 365; day++ {
			if strings.HasPrefix(header, "Year") {
				fmt.Fprintf(&b, "%d,", year)
			}
			fmt.Fprintf(&b, "%d,%v\n", day, period(year, day))
		}
	}
	return b.String()
}

func TestParseWeatherFileYears(t *testing.T) {
	// two years, the rows of year 2 first and in reverse order: the order of the rows does not matter
	text := weatherFile("Year,Day,ForagingPeriod_s", 2, func(year, day int) float64 { return float64(year) + float64(day)/1000 })
	header, rows, _ := strings.Cut(text, "\n")
	lines := strings.Split(strings.TrimSpace(rows), "\n")
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	years, err := parseWeatherFile(strings.NewReader(header + "\n" + strings.Join(lines, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	if len(years) != 2 || years[0][0] != 1.001 || years[0][364] != 1.365 || years[1][0] != 2.001 {
		t.Errorf("years %d, year 1 day 1 %v, day 365 %v, year 2 day 1 %v; want 2, 1.001, 1.365, 2.001", len(years),
			years[0][0], years[0][364], years[1][0])
	}
}

func TestParseWeatherFileErrors(t *testing.T) {
	full := weatherFile("Day,ForagingPeriod_s", 1, func(_, _ int) float64 { return 8 })
	cases := map[string]string{
		"unknown column":   strings.Replace(full, "ForagingPeriod_s", "ForagingHours_h", 1),
		"no Day column":    "ForagingPeriod_s\n8\n",
		"no value row":     "Day,ForagingPeriod_s\n",
		"day missing":      strings.Replace(full, "200,8\n", "", 1),
		"day twice":        strings.Replace(full, "200,8\n", "199,8\n", 1),
		"day 366":          full + "366,8\n",
		"more than 24 h":   strings.Replace(full, "200,8\n", "200,86401\n", 1),
		"not a number":     strings.Replace(full, "200,8\n", "200,eight\n", 1),
		"year 2 not full":  weatherFile("Year,Day,ForagingPeriod_s", 1, func(_, _ int) float64 { return 8 }) + "2,1,8\n",
		"year not natural": "Year,Day,ForagingPeriod_s\n0.5,1,8\n",
	}
	for name, text := range cases {
		if _, err := parseWeatherFile(strings.NewReader(text)); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestWeatherFromParameter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "weather.csv"), []byte(examplefiles.WeatherForagingHoursExample), 0o644); err != nil {
		t.Fatal(err)
	}
	// [FPER-03] a constant option
	w, err := weatherFromParameter("Constant 8 hrs", dir)
	if err != nil || w.Years_s != nil || w.ForagingPeriod_s(1, 1) != 8*3600 || w.ForagingPeriod_s(5, 365) != 8*3600 {
		t.Errorf("Constant 8 hrs: %+v, %v", w, err)
	}
	// a weather file of one year: the same in every year
	w, err = weatherFromParameter("weather.csv", dir)
	if err != nil || w.ForagingPeriod_s(1, 43) != 0.5*3600 || w.ForagingPeriod_s(3, 43) != 0.5*3600 {
		t.Errorf("weather file: day 43 %v s, %v", w.ForagingPeriod_s(1, 43), err)
	}
	// an unknown Weather is an error ([FPER-01] gives -1), also NetLogo's name of the list without its file
	for _, weather := range []string{"foragingHoursExample", "Constant 13 hrs", ""} {
		if _, err := weatherFromParameter(weather, dir); err == nil {
			t.Errorf("Weather %q: want an error", weather)
		}
	}
	// [FPER-02] NetLogo's name of the list reads its weather file
	file := filepath.Join(dir, "_SYSTEM_Weather_foragingHoursExample.csv")
	if err := os.WriteFile(file, []byte(examplefiles.WeatherForagingHoursExample), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err = weatherFromParameter("foragingHoursExample", dir)
	if err != nil || w.ForagingPeriod_s(1, 43) != 0.5*3600 {
		t.Errorf("foragingHoursExample: day 43 %v s, %v", w.ForagingPeriod_s(1, 43), err)
	}
}

func TestWeatherYearsRepeat(t *testing.T) {
	// [FPER] the years of a weather file follow each other and then repeat
	years, err := parseWeatherFile(strings.NewReader(weatherFile("Year,Day,ForagingPeriod_s", 2,
		func(year, _ int) float64 { return float64(year) })))
	if err != nil {
		t.Fatal(err)
	}
	w := globals.Weather{Years_s: years}
	for year, want := range map[int]float64{1: 1, 2: 2, 3: 1, 4: 2} {
		if got := w.ForagingPeriod_s(year, 100); got != want {
			t.Errorf("year %d: %v s, want %v s", year, got, want)
		}
	}
}
