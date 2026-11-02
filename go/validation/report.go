package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ReportRow is the comparison of one metric.
type ReportRow struct {
	Metric  string
	Steps   int       // number of compared days (steps), e.g. 365 for days 0 to 364
	Within  int       // days on which the median of the runs is within the NetLogo 5-95 % band of the same day
	Step    int       // example step, logged by the tests
	Gobees  float64   // median of the runs at Step
	NetLogo Quantiles // NetLogo quantiles at Step
	// Medians and Bands hold the median of the runs and the NetLogo quantiles of every compared day, from day 0 on.
	// They are drawn in the plot of the experiment. A row with fewer than 2 days gets no panel.
	Medians []float64
	Bands   []Quantiles
}

// Section is the comparison of one experiment.
type Section struct {
	// Order is the sort key of the experiment in the report, zero-padded, e.g. "03", "04a", "06", "07.1", "07.2".
	Order string
	// Step is the heading of the translation step, e.g. "Step 7: the year cycle". The experiments of one step share
	// it, the report shows it once.
	Step string
	// Title is the heading of the experiment below the step heading. Empty if the step has only one experiment.
	Title string
	// Experiment is the name of the experiment, e.g. "step7_badgers": the name of its NetLogo reference, of its
	// recorded results and of its plot.
	Experiment  string
	Description string
	Runs        int
	Rows        []ReportRow
	// Values holds the metrics of all runs, Values[metric][step][run]. WriteReport records them.
	Values Values
}

// Meta describes a recorded experiment: what cmd/report needs besides the values.
type Meta struct {
	Order       string
	Step        string
	Title       string
	Experiment  string
	Description string
	Runs        int
	Metrics     []MetricMeta
}

// MetricMeta is a compared metric of an experiment: it is compared on the days 0 to Days-1.
type MetricMeta struct {
	Metric string
	Days   int
}

// WriteReport records the results of an experiment in the folder of DumpEnv: the quantiles per day and the run
// summaries (WriteDump), and <experiment>_meta.json (Meta). Does nothing if DumpEnv is not set. cmd/report compares
// the records of gobees-fast with NetLogo and writes the report.
func WriteReport(s Section) error {
	dir := os.Getenv(DumpEnv)
	if dir == "" || s.Values == nil {
		return nil
	}
	if err := WriteDump(dir, s.Experiment, s.Values); err != nil {
		return err
	}
	meta := Meta{Order: s.Order, Step: s.Step, Title: s.Title, Experiment: s.Experiment, Description: s.Description,
		Runs: s.Runs}
	for _, r := range s.Rows {
		meta.Metrics = append(meta.Metrics, MetricMeta{Metric: r.Metric, Days: r.Steps})
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, s.Experiment+"_meta.json"), b, 0o644)
}
