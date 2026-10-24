package validation

import (
	"fmt"
	"os"
	"strings"
)

// ReportEnv is the environment variable with the path of the Markdown report. If it is set, the comparison tests
// append their results to this file. The CI sets it and adds the report to the summary of the run.
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

// WriteReport appends a Markdown section with the rows to the file in ReportEnv. Does nothing if it is not set.
func WriteReport(title, description string, runs int, rows []ReportRow) error {
	path := os.Getenv(ReportEnv)
	if path == "" {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## %s\n\n%s\n\n", title, description)
	fmt.Fprintf(&b, "%d runs each. Every metric is checked on each of its days (all days of the experiment, unless "+
		"the description above says otherwise): the median of the gobees runs on a day "+
		"must lie within the 5-95 %% band of the NetLogo runs on the same day. \"Days inside the band\" counts the days "+
		"that pass. A metric passes (✅) only if all days pass. The last three columns show the values of one example "+
		"day.\n\n", runs)
	b.WriteString("| | Metric | Days inside the band | Example day | gobees median | NetLogo median [5-95 %] |\n" +
		"| --- | --- | --- | ---: | ---: | --- |\n")
	for _, r := range rows {
		mark := "✅"
		if r.Within < r.Steps {
			mark = "❌"
		}
		fmt.Fprintf(&b, "| %s | %s | %d of %d days | %d | %.6g | %.6g [%.6g, %.6g] |\n",
			mark, r.Metric, r.Within, r.Steps, r.Step, r.Gobees, r.NetLogo.Q50, r.NetLogo.Q05, r.NetLogo.Q95)
	}
	b.WriteString("\n")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(b.String())
	return err
}
