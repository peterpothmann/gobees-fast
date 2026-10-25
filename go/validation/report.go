package validation

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"time"
)

// ReportEnv is the environment variable with the path of the Markdown report. If it is set, the comparison tests
// write their results to this file. The CI sets it and adds the report to the summary of the run.
const ReportEnv = "VALIDATION_REPORT"

// ReportRow is the comparison of one metric.
type ReportRow struct {
	Metric  string
	Steps   int       // number of compared days (steps), e.g. 365 for days 0 to 364
	Within  int       // days on which the gobees median is within the NetLogo 5-95 % band of the same day
	Step    int       // step shown in the report (all Steps are compared)
	Gobees  float64   // gobees median at Step
	NetLogo Quantiles // NetLogo quantiles at Step
}

// Section is the comparison of one experiment in the report.
type Section struct {
	// Order is the sort key of the section in the report, zero-padded, e.g. "03", "04a", "06", "07.1", "07.2".
	// The tests of different packages run at the same time and finish in any order, the report is sorted by Order.
	Order string
	// Step is the heading of the translation step, e.g. "Step 7: the year cycle". The experiments of one step share
	// it, the report shows it once.
	Step string
	// Title is the heading of the experiment below the step heading. Empty if the step has only one experiment.
	Title       string
	Description string
	Runs        int
	Rows        []ReportRow
}

const (
	reportHeader = "# Comparison with NetLogo\n\n" +
		"Every metric is checked on each day of its experiment (unless the description says otherwise): the median " +
		"of the gobees runs on a day must lie within the 5-95 % band of the NetLogo runs on the same day. \"Days " +
		"inside the band\" counts the days that pass. A metric passes (✅) only if all days pass. The last three " +
		"columns show the values of one example day. Details: go/validation/README.md.\n\n"
	sectionMarker = "<!-- section " // followed by the order, "|", the step heading and " -->"
)

// WriteReport adds the section to the report file in ReportEnv, or replaces the section with the same Order, and
// writes the report sorted by Order. Does nothing if ReportEnv is not set.
// Several test processes can write at the same time: a lock file next to the report serialises them.
func WriteReport(s Section) error {
	path := os.Getenv(ReportEnv)
	if path == "" {
		return nil
	}
	unlock, err := lockFile(path + ".lock")
	if err != nil {
		return err
	}
	defer unlock()

	sections := map[string]storedSection{}
	if b, err := os.ReadFile(path); err == nil {
		sections = parseReport(string(b))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	sections[s.Order] = storedSection{order: s.Order, step: s.Step, body: renderSection(s)}
	return os.WriteFile(path, []byte(renderReport(sections)), 0o644)
}

// storedSection is a section as it is stored in the report file: a marker line with order and step heading,
// then the body (experiment heading, description, table).
type storedSection struct {
	order, step, body string
}

// renderSection returns the body of a section.
func renderSection(s Section) string {
	var b strings.Builder
	if s.Title != "" {
		fmt.Fprintf(&b, "### %s\n\n", s.Title)
	}
	fmt.Fprintf(&b, "%s\n\n%d runs each.\n\n", s.Description, s.Runs)
	b.WriteString("| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |\n" +
		"| --- | --- | --- | ---: | ---: | --- |\n")
	for _, r := range s.Rows {
		mark := "✅"
		if r.Within < r.Steps {
			mark = "❌"
		}
		fmt.Fprintf(&b, "| %s | %s | %d of %d days | %d | %.6g | %.6g [%.6g, %.6g] |\n",
			mark, r.Metric, r.Within, r.Steps, r.Step, r.Gobees, r.NetLogo.Q50, r.NetLogo.Q05, r.NetLogo.Q95)
	}
	b.WriteString("\n")
	return b.String()
}

// renderReport returns the whole report: the header, then the sections sorted by order, with each step heading
// once before the first section of the step.
func renderReport(sections map[string]storedSection) string {
	orders := make([]string, 0, len(sections))
	for o := range sections {
		orders = append(orders, o)
	}
	slices.Sort(orders)

	var b strings.Builder
	b.WriteString(reportHeader)
	lastStep := ""
	for _, o := range orders {
		s := sections[o]
		if s.step != lastStep {
			fmt.Fprintf(&b, "## %s\n\n", s.step)
			lastStep = s.step
		}
		fmt.Fprintf(&b, "%s%s|%s -->\n%s", sectionMarker, s.order, s.step, s.body)
	}
	return b.String()
}

// parseReport reads the sections of a report written by renderReport. A section body ends at the next marker or
// step heading ("## ").
func parseReport(text string) map[string]storedSection {
	sections := map[string]storedSection{}
	var current *storedSection
	var body strings.Builder
	finish := func() {
		if current != nil {
			current.body = body.String()
			sections[current.order] = *current
			current = nil
		}
		body.Reset()
	}
	for _, line := range strings.SplitAfter(text, "\n") {
		switch {
		case strings.HasPrefix(line, sectionMarker):
			finish()
			order, step, _ := strings.Cut(strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, sectionMarker)), "-->"), "|")
			current = &storedSection{order: strings.TrimSpace(order), step: strings.TrimSpace(step)}
		case strings.HasPrefix(line, "## "):
			finish()
		case current != nil:
			body.WriteString(line)
		}
	}
	finish()
	return sections
}

// lockFile creates the lock file exclusively, waiting up to 2 minutes for another test process to remove it.
// It returns the function that removes it.
func lockFile(path string) (func(), error) {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(path) }, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("report lock %s still exists after 2 minutes (delete it if no test is running)", path)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
