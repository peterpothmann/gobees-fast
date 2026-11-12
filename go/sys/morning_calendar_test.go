package sys

import (
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/day"
	"gobees/examplefiles"
	"gobees/globals"
	"gobees/input"
	"gobees/params"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

func TestDayOfYear(t *testing.T) {
	// worked examples
	cases := map[day.Abs]day.OfYear{0: 0, 1: 1, 364: 364, 365: 365, 366: 1, 730: 365, 731: 1}
	for ticks, want := range cases {
		if got := dayOfYear(ticks); got != want {
			t.Errorf("dayOfYear(%d) = %v, want %v", ticks, got, want)
		}
	}
	// the integer formula equals the NetLogo formula round (ticks mod 365.00000001) for 100 years
	for ticks := day.Abs(0); ticks <= 100*365; ticks++ {
		netlogo := round(math.Mod(float64(ticks), 365.00000001))
		if got := dayOfYear(ticks); float64(got) != netlogo {
			t.Fatalf("ticks %d: dayOfYear %v, NetLogo formula %v", ticks, got, netlogo)
		}
	}
}

// exampleWeatherYear returns the example weather year (examplefiles.WeatherForagingHoursExample, the columns Day and
// ForagingPeriod_s), the foraging periods of step8_weather.
func exampleWeatherYear(t *testing.T) input.Year {
	t.Helper()
	year := make(input.Year, 365)
	lines := strings.Split(strings.TrimSpace(examplefiles.WeatherForagingHoursExample), "\n")
	for _, line := range lines[1:] {
		day, period, _ := strings.Cut(strings.TrimSpace(line), ",")
		d, err1 := strconv.Atoi(day)
		v, err2 := strconv.ParseFloat(period, 64)
		if err1 != nil || err2 != nil {
			t.Fatalf("example weather year: line %q", line)
		}
		year[d-1] = v
	}
	return year
}

// morningWorld is a world with the resources of the morning systems, without foodsources and bees. years are the
// foraging periods of the run file, nil for those of the example farm (8 hours on every day).
func morningWorld(t *testing.T, years []input.Year) (*ecs.World, *globals.Ticks) {
	t.Helper()
	w := ecs.NewWorld()
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(1, 2)})
	ecs.AddResource(w, &params.Parameters{})
	ticks := &globals.Ticks{}
	ecs.AddResource(w, ticks)
	p := exampleFarmParams(t, nil, func(c *input.Config) {
		if years != nil {
			c.ForagingPeriod = years
		}
	})
	ecs.AddResource(w, &p.Input.ForagingPeriod)
	return w, ticks
}

func TestCalendarUpdateForagingPeriod(t *testing.T) {
	w, ticks := morningWorld(t, nil)
	s := &calendarUpdate{}
	s.Initialize(w) // Setup, ticks = 0
	cal := ecs.GetResource[globals.Calendar](w)
	// Day 0, no foraging period in Setup
	if cal.Day != 0 || cal.ForagingPeriod != 0 {
		t.Errorf("Setup: Day %v, ForagingPeriod %v, want 0 and 0", cal.Day, cal.ForagingPeriod)
	}
	ticks.Ticks = 1
	s.Update(w)
	if cal.Day != 1 || cal.ForagingPeriod != 8*3600 {
		t.Errorf("day 1: Day %v, ForagingPeriod %v, want 1 and 28800", cal.Day, cal.ForagingPeriod)
	}
	ticks.Ticks = 366
	s.Update(w)
	if cal.Day != 1 {
		t.Errorf("ticks 366: Day %v, want 1", cal.Day)
	}
}

func TestCalendarUpdateForagingHoursExample(t *testing.T) {
	// with the example weather year, day 43 is the ".5" of the NetLogo list
	w, ticks := morningWorld(t, []input.Year{exampleWeatherYear(t)})
	s := &calendarUpdate{}
	s.Initialize(w)
	ticks.Ticks = 365 + 43 // second year: a file of one year repeats, read by Day, not by ticks
	s.Update(w)
	if got := ecs.GetResource[globals.Calendar](w).ForagingPeriod; got != 0.5*3600 {
		t.Errorf("ticks %d (Day 43): ForagingPeriod %v, want 1800", ticks.Ticks, got)
	}
}

// flowerPeriods returns startDay and stopDay of flower species from the example FlowerspeciesFile.
func flowerPeriods(t *testing.T, names ...string) map[string][2]day.OfYear {
	t.Helper()
	table := exampleFarmParams(t, nil, func(c *input.Config) {}).Input.Flowers
	result := map[string][2]day.OfYear{}
	for _, n := range names {
		row, err := table.Get(n)
		if err != nil {
			t.Fatal(err)
		}
		result[n] = [2]day.OfYear{row.StartDay, row.StopDay}
	}
	return result
}

func TestFlowerRefill(t *testing.T) {
	// worked examples from netlogo/_SYSTEM_Flowerspecies.csv, columns startDay, stopDay
	periods := flowerPeriods(t, "Dandelion", "Daisy", "Crop_Cereals")
	if periods["Dandelion"] != [2]day.OfYear{1, 364} || periods["Daisy"] != [2]day.OfYear{1, 365} || periods["Crop_Cereals"] != [2]day.OfYear{0, 0} {
		t.Fatalf("periods from the example file: %v", periods)
	}

	w, ticks := morningWorld(t, nil)
	// one flower patch per flower species, each with one layer
	fpMap := ecs.NewMap1[comp.FlowerPatch](w)
	names := []string{"Dandelion", "Daisy", "Crop_Cereals"}
	entities := map[string]ecs.Entity{}
	for i, n := range names {
		// state left over from the day before: is overwritten every morning
		l := comp.Layer{StartDay: periods[n][0], StopDay: periods[n][1],
			NectarMax: float64(i+1) * 1e6, PollenMax: float64(i+1) * 1000, Nectar: 7, Pollen: 7}
		entities[n] = fpMap.NewEntity(&comp.FlowerPatch{Layers: []comp.Layer{l}, PollenInFlower: true, NectarInFlower: true})
	}
	morning := &calendarUpdate{}
	morning.Initialize(w)
	s := &flowerRefill{}
	s.Initialize(w)
	today := ecs.GetResource[globals.Flowering](w)

	// inFlower also checks the flags of the flower patch: its only layer has food above 0 exactly when it is full
	inFlower := func(n string) bool {
		fp := fpMap.Get(entities[n])
		st := &fp.Layers[0]
		full := st.Nectar == st.NectarMax && st.Pollen == st.PollenMax
		empty := st.Nectar == 0 && st.Pollen == 0
		if full == empty {
			t.Fatalf("%s: nectar %v, pollen %v is neither full nor empty", n, st.Nectar, st.Pollen)
		}
		if fp.PollenInFlower != full || fp.NectarInFlower != full {
			t.Fatalf("%s: full %v, flags pollen %v, nectar %v", n, full, fp.PollenInFlower, fp.NectarInFlower)
		}
		return full
	}
	cases := []struct {
		ticks                    day.Abs
		dandelion, daisy, cereal bool
		update                   bool
	}{
		{0, false, false, false, true}, // Setup: Crop_Cereals sets Changed (startDay = stopDay = 0)
		{1, true, true, false, true},
		{2, true, true, false, false},
		{363, true, true, false, false},
		{364, false, true, false, true}, // stopDay of Dandelion
		{365, false, false, false, true},
		{366, true, true, false, true}, // Day 1 of year 2
	}
	for _, c := range cases {
		ticks.Ticks = c.ticks
		if c.ticks > 0 {
			morning.Update(w)
			s.Update(w)
		}
		if inFlower("Dandelion") != c.dandelion || inFlower("Daisy") != c.daisy || inFlower("Crop_Cereals") != c.cereal {
			t.Errorf("ticks %d: in flower Dandelion %v, Daisy %v, Crop_Cereals %v; want %v, %v, %v", c.ticks,
				inFlower("Dandelion"), inFlower("Daisy"), inFlower("Crop_Cereals"), c.dandelion, c.daisy, c.cereal)
		}
		if today.Changed != c.update {
			t.Errorf("ticks %d: Flowering.Changed %v, want %v", c.ticks, today.Changed, c.update)
		}
	}

	// assertion: no flowering period across the turn of the year
	fpMap.NewEntity(&comp.FlowerPatch{Layers: []comp.Layer{{StartDay: 300, StopDay: 20}}})
	expectPanic(t, "startDay > stopDay", func() { s.Update(w) })
}
