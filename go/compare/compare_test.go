package compare

import (
	"slices"
	"strings"
	"testing"
)

// gobeesOutput is the beginning of the output of gobees (fork point 40dc301) with -s 7: the 15 columns of
// DailyOutput after the column t of reporter.CSV.
const gobeesOutput = "t\tticks\tTotalIBMColonies\tTotalQueens\tTotalMatedQueens\tTotalUnmatedQueens\t" +
	"TotalHibernatingQueens\tTotalColonies\tTotalBeeAgents\tTotalMales\tTotalActiveBees\tMeanWorkerWeight_mg\t" +
	"MeanQueenWeight_mg\tMeanAdultWeight_mg\tTotalMapArea_km2\tColonyDensity_km2\n" +
	"198\t199\t0\t283\t223\t60\t216\t11\t81\t216\t19\t171.69201312015804\t780.5708633722423\t456.7383321970625\t" +
	"1.9978001665033709\t5.50605620343532\r\n" +
	"398\t399\t0\t300\t300\t0\t300\t0\t25\t0\t0\t0\t781.5547737416106\t781.5547737416106\t1.9978001665033709\t0\n"

func TestReadTable(t *testing.T) {
	table, err := ReadTable(strings.NewReader(gobeesOutput))
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Header) != 16 || len(table.Rows) != 2 {
		t.Fatalf("got %d columns and %d rows, want 16 and 2", len(table.Header), len(table.Rows))
	}
	if got := table.Rows[0][table.Column("MeanWorkerWeight_mg")]; got != 171.69201312015804 {
		t.Errorf("MeanWorkerWeight_mg on day 199: got %v", got)
	}
	if got := table.Rows[1][table.Column("ticks")]; got != 399 {
		t.Errorf("ticks of row 2: got %v", got)
	}
	if table.Column("Eggs") != -1 {
		t.Error("Eggs is not a column")
	}
}

func TestReadTableRounded(t *testing.T) {
	// day 41 of step8_weather: NetLogo's 2.2 h * 3600 against the whole seconds of the weather file
	netlogo := table(t, "ticks\tforaging_period_s\tMeanWorkerWeight_mg\n41\t7920.000000000001\t171.69201312015804\n")
	fast := table(t, "ticks\tforaging_period_s\tMeanWorkerWeight_mg\n41\t7920\t171.69201312015804\n")
	if err := Exact(netlogo, fast, []string{"ticks", "foraging_period_s", "MeanWorkerWeight_mg"}); err != nil {
		t.Errorf("foraging_period_s is rounded: %v", err)
	}
	if got := netlogo.Rows[0][netlogo.Column("MeanWorkerWeight_mg")]; got != 171.69201312015804 {
		t.Errorf("other columns are not rounded: got %v", got)
	}
}

func TestReadTableErrors(t *testing.T) {
	for name, text := range map[string]string{
		"empty":         "",
		"short row":     "a\tb\n1\n",
		"not a number":  "a\tb\n1\tx\n",
		"no header row": "\n\n",
	} {
		if _, err := ReadTable(strings.NewReader(text)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func table(t *testing.T, text string) Table {
	t.Helper()
	tab, err := ReadTable(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return tab
}

func TestColumns(t *testing.T) {
	ref := table(t, "t\tticks\tA\tB\n0\t1\t2\t3\n")
	cur := table(t, "t\tticks\tB\tC\n0\t1\t3\t4\n")
	common, removed, added := Columns(ref, cur)
	if !slices.Equal(common, []string{"t", "ticks", "B"}) || !slices.Equal(removed, []string{"A"}) ||
		!slices.Equal(added, []string{"C"}) {
		t.Errorf("got common %v, removed %v, added %v", common, removed, added)
	}
}

func TestExact(t *testing.T) {
	ref := table(t, "ticks\tA\tB\n1\t2\t3\n2\t2.5\t3\n")
	// other column order and a further column: equal on the common columns
	if err := Exact(ref, table(t, "B\tticks\tA\tC\n3\t1\t2\t9\n3\t2\t2.5\t9\n"), []string{"ticks", "A", "B"}); err != nil {
		t.Errorf("equal tables: %v", err)
	}

	err := Exact(ref, table(t, "ticks\tA\tB\n1\t2\t3\n2\t2.5000000000000004\t3\n"), []string{"ticks", "A", "B"})
	if err == nil || !strings.Contains(err.Error(), "day 2, column A") {
		t.Errorf("one bit differs on day 2: got %v", err)
	}

	err = Exact(ref, table(t, "ticks\tA\tB\n1\t2\t3\n"), []string{"ticks", "A", "B"})
	if err == nil || !strings.Contains(err.Error(), "2 days in the reference, 1 in the current") {
		t.Errorf("a shorter run: got %v", err)
	}
}

func TestSeeds(t *testing.T) {
	if got := Seeds(3); !slices.Equal(got, []int64{1, 2, 3}) {
		t.Errorf("got %v", got)
	}
}
