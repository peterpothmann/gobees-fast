package obs

import (
	"os"
	"path/filepath"
	"testing"

	"gobees/globals"

	"github.com/mlange-42/ark-tools/reporter"
	"github.com/mlange-42/ark/ecs"
)

func TestDailyOutputCSV(t *testing.T) {
	w := ecs.NewWorld()
	ticks := globals.Ticks{Ticks: 1}
	ecs.AddResource(w, &ticks)
	output := globals.DailyOutput{TotalQueens: 500, TotalHibernatingQueens: 500, TotalMapArea_km2: 1.25}
	ecs.AddResource(w, &output)

	file := filepath.Join(t.TempDir(), "out", "daily_output.csv")
	csv := reporter.CSV{Observer: &DailyOutput{}, File: file, Sep: "\t"}
	csv.Initialize(w)
	csv.Update(w) // day 1

	ticks.Ticks = 2
	output.TotalColonies, output.TotalHibernatingQueens, output.MeanQueenWeight_mg = 2, 498, 600.5
	output.ColonyDensity_km2 = 1.6
	csv.Update(w) // day 2
	csv.Finalize(w)

	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	// TotalIBMColonies of NetLogo is not written: gobees-fast has no individual-based colonies (step F2)
	want := "t\tticks\tTotalQueens\tTotalMatedQueens\tTotalUnmatedQueens\tTotalHibernatingQueens\t" +
		"TotalColonies\tTotalBeeAgents\tTotalMales\tTotalActiveBees\tMeanWorkerWeight_mg\tMeanQueenWeight_mg\t" +
		"MeanAdultWeight_mg\tTotalMapArea_km2\tColonyDensity_km2\n" +
		"0\t1\t500\t0\t0\t500\t0\t0\t0\t0\t0\t0\t0\t1.25\t0\n" +
		"1\t2\t500\t0\t0\t498\t2\t0\t0\t0\t0\t600.5\t0\t1.25\t1.6\n"
	if got := string(b); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
