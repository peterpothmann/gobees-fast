package obs

import (
	"slices"
	"strings"
	"testing"

	"gobees/comp"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func TestColumnDocs(t *testing.T) {
	// every column has a description, and no description is left over
	all := AllColumns()
	for _, name := range all {
		if ColumnDoc(name) == "" {
			t.Errorf("column %s has no description", name)
		}
	}
	for name := range columnDocs {
		if !slices.Contains(all, name) {
			t.Errorf("description of %s, which is no column", name)
		}
	}
	// no name twice
	seen := map[string]bool{}
	for _, name := range all {
		if seen[name] {
			t.Errorf("column %s twice", name)
		}
		seen[name] = true
	}
	if all[0] != "ticks" {
		t.Errorf("first column = %s, want ticks", all[0])
	}
}

func TestSelectColumns(t *testing.T) {
	for _, tc := range []struct {
		name      string
		requested []string
		want      []string
	}{
		{"none: all", nil, AllColumns()},
		{"empty: all", []string{}, AllColumns()},
		{"ticks added in front", []string{"eggs", "colonies"}, []string{"ticks", "eggs", "colonies"}},
		{"ticks moved to the front", []string{"eggs", "ticks", "colonies"}, []string{"ticks", "eggs", "colonies"}},
		{"only ticks", []string{"ticks"}, []string{"ticks"}},
	} {
		got, err := SelectColumns(tc.requested)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}

	// an unknown name: the error lists the available columns
	_, err := SelectColumns([]string{"eggs", "egg"})
	if err == nil || !strings.Contains(err.Error(), `"egg"`) || !strings.Contains(err.Error(), "larvae") {
		t.Errorf("unknown name: error = %v, want the name and the list of columns", err)
	}
	// a repeated name
	if _, err := SelectColumns([]string{"eggs", "eggs"}); err == nil {
		t.Error("want an error for a column listed twice")
	}
}

func TestDailyOutputExtendedColumns(t *testing.T) {
	// the selected columns hold the values of the same columns of the full output, also with a counter that is a
	// difference to the previous day (ForagerDeathsToday)
	w := newWorld(t)
	ecs.AddResource(w, &params.Input{Species: []params.Species{{SpeciesFile: params.SpeciesFile{Name: "B_terrestris"}}}})
	ecs.NewMap2[comp.Colony, comp.ColonyStores](w).NewEntity(&comp.Colony{}, &comp.ColonyStores{Energy: 7})

	all := DailyOutputExtended{}
	all.Initialize(w)
	columns, err := SelectColumns([]string{"energy_kJ", "ForagerDeathsToday", "colonies"})
	if err != nil {
		t.Fatal(err)
	}
	some := DailyOutputExtended{Columns: columns}
	some.Initialize(w)

	if got, want := some.Header(), []string{"ticks", "energy_kJ", "ForagerDeathsToday", "colonies"}; !slices.Equal(got, want) {
		t.Fatalf("Header = %v, want %v", got, want)
	}
	if got := all.Header(); !slices.Equal(got, AllColumns()) {
		t.Fatalf("Header without Columns = %v, want all columns", got)
	}
	all.Update(w)
	some.Update(w)
	full, part := all.Values(w), some.Values(w)
	if len(full) != len(AllColumns()) || len(part) != 4 {
		t.Fatalf("%d and %d values, want %d and 4", len(full), len(part), len(AllColumns()))
	}
	for i, name := range some.Header() {
		if want := full[slices.Index(AllColumns(), name)]; part[i] != want {
			t.Errorf("%s = %v, want %v", name, part[i], want)
		}
	}
	if part[1] != 7 || part[3] != 1 {
		t.Errorf("energy_kJ, colonies = %v, %v, want 7, 1", part[1], part[3])
	}
}
