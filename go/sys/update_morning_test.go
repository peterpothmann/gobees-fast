package sys

import (
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/globals"
	"gobees/internal/examplefiles"
	"gobees/params"
	"gobees/util"

	"github.com/mlange-42/ark-tools/resource"
	"github.com/mlange-42/ark/ecs"
)

func TestDayOfYear(t *testing.T) {
	// [MORN-02] finding 36 (see trans/netlogo-findings.md#finding-36): the worked examples of the annotation
	cases := map[int64]float64{0: 0, 1: 1, 364: 364, 365: 365, 366: 1, 730: 365, 731: 1}
	for ticks, want := range cases {
		if got := dayOfYear(ticks); got != want {
			t.Errorf("dayOfYear(%d) = %v, want %v", ticks, got, want)
		}
	}
	// the integer formula equals the NetLogo formula round (ticks mod 365.00000001) for 100 years
	for ticks := int64(0); ticks <= 100*365; ticks++ {
		netlogo := util.Round(math.Mod(float64(ticks), 365.00000001))
		if got := dayOfYear(ticks); got != netlogo {
			t.Fatalf("ticks %d: dayOfYear %v, NetLogo formula %v", ticks, got, netlogo)
		}
	}
}

// weatherExampleFile is the file name of the example weather year (examplefiles.WeatherForagingHoursExample).
const weatherExampleFile = "_SYSTEM_Weather_foragingHoursExample.csv"

// morningWorld is a world with the resources of the morning systems, without foodsources and bees. weather is the
// parameter Weather: a "Constant N hrs" option or weatherExampleFile.
func morningWorld(t *testing.T, weather string) (*ecs.World, *globals.Ticks) {
	t.Helper()
	w := ecs.NewWorld()
	ecs.AddResource(w, &resource.Rand{Source: rand.NewPCG(1, 2)})
	p := params.Default()
	p.File.Weather = weather
	ecs.AddResource(w, &p.File)
	ticks := &globals.Ticks{}
	ecs.AddResource(w, ticks)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, weatherExampleFile), []byte(examplefiles.WeatherForagingHoursExample), 0o644); err != nil {
		t.Fatal(err)
	}
	weatherResource, err := weatherFromParameter(weather, dir)
	if err != nil {
		t.Fatal(err)
	}
	ecs.AddResource(w, &weatherResource)
	return w, ticks
}

func TestUpdateMorningForagingPeriod(t *testing.T) {
	w, ticks := morningWorld(t, "Constant 8 hrs")
	s := &UpdateMorning{}
	s.Initialize(w) // Setup, ticks = 0 [INIT-10]
	cal := ecs.GetResource[globals.Calendar](w)
	// finding 4 (see trans/netlogo-findings.md#finding-4): Day 0, no foraging period in Setup
	if cal.Day != 0 || cal.DailyForagingPeriod_s != 0 {
		t.Errorf("Setup: Day %v, DailyForagingPeriod_s %v, want 0 and 0", cal.Day, cal.DailyForagingPeriod_s)
	}
	ticks.Ticks = 1
	s.Update(w)
	if cal.Day != 1 || cal.DailyForagingPeriod_s != 8*3600 {
		t.Errorf("day 1: Day %v, DailyForagingPeriod_s %v, want 1 and 28800", cal.Day, cal.DailyForagingPeriod_s)
	}
	ticks.Ticks = 366
	s.Update(w)
	if cal.Day != 1 {
		t.Errorf("ticks 366: Day %v, want 1", cal.Day)
	}
}

func TestUpdateMorningForagingHoursExample(t *testing.T) {
	// [MORN-04] with the example weather year, day 43 is the ".5" of the NetLogo list (finding 1, see
	// trans/netlogo-findings.md#finding-1)
	w, ticks := morningWorld(t, weatherExampleFile)
	s := &UpdateMorning{}
	s.Initialize(w)
	ticks.Ticks = 365 + 43 // second year: a file of one year repeats, read by Day, not by ticks
	s.Update(w)
	if got := ecs.GetResource[globals.Calendar](w).DailyForagingPeriod_s; got != 0.5*3600 {
		t.Errorf("ticks %d (Day 43): DailyForagingPeriod_s %v, want 1800", ticks.Ticks, got)
	}
}

// flowerPeriods returns startDay and stopDay of flower species from the example FlowerspeciesFile.
func flowerPeriods(t *testing.T, names ...string) map[string][2]float64 {
	t.Helper()
	table, err := parseFlowerspeciesFile(strings.NewReader(examplefiles.Flowerspecies))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][2]float64{}
	for _, n := range names {
		row, err := table.get(n)
		if err != nil {
			t.Fatal(err)
		}
		result[n] = [2]float64{row.StartDay, row.StopDay}
	}
	return result
}

func TestUpdateFoodsources(t *testing.T) {
	// [UFS-02] worked examples of the annotation from netlogo/_SYSTEM_Flowerspecies.csv, columns startDay, stopDay
	periods := flowerPeriods(t, "Dandelion", "Daisy", "Crop_Cereals")
	if periods["Dandelion"] != [2]float64{1, 364} || periods["Daisy"] != [2]float64{1, 365} || periods["Crop_Cereals"] != [2]float64{0, 0} {
		t.Fatalf("periods from the example file: %v", periods)
	}

	w, ticks := morningWorld(t, "Constant 8 hrs")
	fsMap := ecs.NewMap2[comp.Foodsource, comp.FoodsourceState](w)
	names := []string{"Dandelion", "Daisy", "Crop_Cereals"}
	entities := map[string]ecs.Entity{}
	for i, n := range names {
		fs := comp.Foodsource{FlowerSpecies: n, StartDay: periods[n][0], StopDay: periods[n][1],
			NectarMax_myl: float64(i+1) * 1e6, PollenMax_g: float64(i+1) * 1000}
		// state left over from the day before: is overwritten every morning
		entities[n] = fsMap.NewEntity(&fs, &comp.FoodsourceState{Nectar_myl: 7, Pollen_g: 7})
	}
	morning := &UpdateMorning{}
	morning.Initialize(w)
	s := &UpdateFoodsources{}
	s.Initialize(w)
	today := ecs.GetResource[globals.FoodsourcesToday](w)

	inFlower := func(n string) bool {
		fs, st := fsMap.Get(entities[n])
		full := st.Nectar_myl == fs.NectarMax_myl && st.Pollen_g == fs.PollenMax_g
		empty := st.Nectar_myl == 0 && st.Pollen_g == 0
		if full == empty {
			t.Fatalf("%s: nectar %v, pollen %v is neither full nor empty", n, st.Nectar_myl, st.Pollen_g)
		}
		return full
	}
	cases := []struct {
		ticks                    int64
		dandelion, daisy, cereal bool
		update                   bool
		nectar_l, pollen_kg      float64
	}{
		{0, false, false, false, true, 0, 0}, // Setup: Crop_Cereals sets the flag (startDay = stopDay = 0)
		{1, true, true, false, true, 3, 3},   // 1 + 2 million microlitres = 3 l, 1 + 2 kg
		{2, true, true, false, false, 3, 3},
		{363, true, true, false, false, 3, 3},
		{364, false, true, false, true, 2, 2}, // stopDay of Dandelion
		{365, false, false, false, true, 0, 0},
		{366, true, true, false, true, 3, 3}, // Day 1 of year 2
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
		if today.FoodsourcesInFlowerUpdate != c.update {
			t.Errorf("ticks %d: FoodsourcesInFlowerUpdate %v, want %v", c.ticks, today.FoodsourcesInFlowerUpdate, c.update)
		}
		if today.NectarAvailableTotalToday_l != c.nectar_l || today.PollenAvailableTotalToday_kg != c.pollen_kg {
			t.Errorf("ticks %d: totals %v l, %v kg; want %v, %v", c.ticks,
				today.NectarAvailableTotalToday_l, today.PollenAvailableTotalToday_kg, c.nectar_l, c.pollen_kg)
		}
	}

	// assertion: no flowering period across the turn of the year
	fsMap.NewEntity(&comp.Foodsource{StartDay: 300, StopDay: 20}, &comp.FoodsourceState{})
	expectPanic(t, "startDay > stopDay", func() { s.Update(w) })
}
