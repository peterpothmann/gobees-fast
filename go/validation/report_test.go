package validation

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWriteReportOrder(t *testing.T) {
	// The comparison tests of different packages finish in any order. The report is sorted by Order, a step with
	// several experiments gets one heading, and a section written again replaces the old one.
	path := filepath.Join(t.TempDir(), "report.md")
	t.Setenv(ReportEnv, path)
	row := []ReportRow{{Metric: "colonies", Steps: 3, Within: 3}}
	sections := []Section{
		{Order: "07.2", Step: "Step 7: the year cycle", Title: "Badgers", Description: "badgers", Runs: 1, Rows: row},
		{Order: "04a", Step: "Step 4a: morning routine", Description: "morning", Runs: 1, Rows: row},
		{Order: "07.1", Step: "Step 7: the year cycle", Title: "Example farm", Description: "farm", Runs: 1, Rows: row},
		{Order: "03", Step: "Steps 1 to 3", Description: "queens", Runs: 1, Rows: row},
	}
	// written at the same time, as by the test processes of go/sys and go/model
	var wg sync.WaitGroup
	for _, s := range sections {
		wg.Go(func() {
			if err := WriteReport(s); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	// written again with a new description
	sections[1].Description = "morning, second run"
	if err := WriteReport(sections[1]); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	want := []string{"# Comparison with NetLogo", "## Steps 1 to 3", "queens", "## Step 4a: morning routine",
		"morning, second run", "## Step 7: the year cycle", "### Example farm", "farm", "### Badgers", "badgers"}
	pos := 0
	for _, w := range want {
		i := strings.Index(text[pos:], w)
		if i < 0 {
			t.Fatalf("%q missing or out of order in:\n%s", w, text)
		}
		pos += i + len(w)
	}
	if n := strings.Count(text, "## Step 7: the year cycle"); n != 1 {
		t.Errorf("step 7 heading %d times, want once", n)
	}
	if strings.Contains(text, "morning\n") {
		t.Errorf("the old section was not replaced:\n%s", text)
	}
	if _, err := os.Stat(path + ".lock"); !os.IsNotExist(err) {
		t.Errorf("lock file left behind")
	}
}
