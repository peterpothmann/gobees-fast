package benchmark

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

// Machine describes where the benchmark ran. Empty fields are left out of the report.
type Machine struct {
	CPU    string
	Cores  int // logical cores
	OS     string
	Go     string
	Java   string
	Commit string
}

// Report is the comparison of the run times. Base (gobees) and Current (gobees-fast) have the same seeds and days as
// NetLogo, in the same order.
type Report struct {
	NetLogo    NetLogo
	BaseCommit string
	Base       []Run
	Current    []Run
	Threads    int // runs at the same time, in all models
	Machine    Machine

	// Below is the comparison with the branch below the pull request (its base branch), from RunAlternating with
	// the same seeds and days. Left out of the report if Below.Runs is empty.
	Below Below
}

// Below is the comparison of gobees-fast with the base branch of its pull request (RunAlternating).
type Below struct {
	Branch  string // name of the base branch, e.g. main or step-S4
	Commit  string // its commit
	Rounds  int    // runs per seed and program, the fastest counts
	Runs    []Run  // the program of the base branch
	Current []Run  // the program of the current code
}

// WriteReport writes the report as Markdown to path.
func WriteReport(path string, r Report) error {
	return os.WriteFile(path, []byte(r.Markdown()), 0o644)
}

// Markdown returns the report as Markdown.
func (r Report) Markdown() string {
	var b strings.Builder
	n := r.NetLogo
	fmt.Fprintf(&b, "# Run time of NetLogo, gobees and gobees-fast\n\n"+
		"The example farm with %d seeds, %d days each, %d runs at the same time in all models. gobees is the base "+
		"commit %s, gobees-fast the current code. NetLogo 5.3.1 runs headless with BehaviorSpace and the GUI "+
		"switches off, its time is its Setup and its days, without the start of Java. gobees and gobees-fast run "+
		"as command line programs without output, one version after the other, and are timed from the start of "+
		"the program to its end, including reading the input files. Details: go/benchmark/README.md.\n\n",
		len(n.Runs), n.Days, r.Threads, r.BaseCommit)

	whole := func(run Run) time.Duration { return run.Whole() }
	nl, base, cur := durations(n.Runs, whole), durations(r.Base, whole), durations(r.Current, whole)
	b.WriteString("| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | " +
		"NetLogo / gobees-fast | gobees / gobees-fast |\n| --- | ---: | ---: | ---: | ---: | ---: |\n")
	fmt.Fprintf(&b, "| Whole run | %s | %s | %s | %.0f | %.2f |\n", spread(nl), spread(base), spread(cur),
		float64(median(nl))/float64(median(cur)), float64(median(base))/float64(median(cur)))

	b.WriteString("\n## Runs\n\nThe runs of NetLogo and of the Go versions with the same seed are not the same " +
		"run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a " +
		"comparable amount of work.\n\n" +
		"| Run | Seed | NetLogo | NetLogo colonies founded | gobees | gobees colonies founded | gobees-fast | " +
		"gobees-fast colonies founded |\n| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for i, nl := range n.Runs {
		base, cur := r.Base[i], r.Current[i]
		fmt.Fprintf(&b, "| %d | %g | %s | %g | %s | %g | %s | %g |\n", nl.Number, nl.Seed, format(nl.Whole()),
			nl.ColoniesFounded, format(base.Whole()), base.ColoniesFounded, format(cur.Whole()), cur.ColoniesFounded)
	}

	if below := r.Below; len(below.Runs) > 0 {
		writeBelow(&b, below, n.Days)
	}

	b.WriteString("\n## Machine\n\n")
	m := r.Machine
	for _, line := range []struct{ name, value string }{
		{"CPU", m.CPU},
		{"Logical cores", fmt.Sprint(m.Cores)},
		{"OS", m.OS},
		{"Go", m.Go},
		{"Java", m.Java},
		{"Commit", m.Commit},
	} {
		if line.value != "" && line.value != "0" {
			fmt.Fprintf(&b, "- %s: %s\n", line.name, line.value)
		}
	}
	return b.String()
}

// writeBelow writes the section of the comparison with the base branch of the pull request.
func writeBelow(b *strings.Builder, below Below, days int) {
	whole := func(run Run) time.Duration { return run.Whole() }
	old, cur := durations(below.Runs, whole), durations(below.Current, whole)
	fmt.Fprintf(b, "\n## Against the branch below\n\n"+
		"gobees-fast against the base branch of the pull request, %s (commit %s): the same %d seeds, %d days "+
		"each. Both programs run one run at a time and alternately, each seed %d times per program, the order "+
		"switches every round. The time of a seed is its fastest run. Change: median of gobees-fast against the "+
		"median of %s, negative is faster.\n\n", below.Branch, below.Commit, len(below.Runs), days, below.Rounds,
		below.Branch)
	fmt.Fprintf(b, "| | %s median (min-max) | gobees-fast median (min-max) | Change |\n| --- | ---: | ---: | ---: |\n",
		below.Branch)
	fmt.Fprintf(b, "| Whole run | %s | %s | %s |\n", spread(old), spread(cur), change(median(old), median(cur)))
	fmt.Fprintf(b, "\n| Seed | %s | %s colonies founded | gobees-fast | gobees-fast colonies founded | Change |\n"+
		"| ---: | ---: | ---: | ---: | ---: | ---: |\n", below.Branch, below.Branch)
	for i, o := range below.Runs {
		c := below.Current[i]
		fmt.Fprintf(b, "| %g | %s | %g | %s | %g | %s |\n", o.Seed, format(o.Whole()), o.ColoniesFounded,
			format(c.Whole()), c.ColoniesFounded, change(o.Whole(), c.Whole()))
	}
}

// change shows the relative change from old to cur in percent, with sign.
func change(old, cur time.Duration) string {
	return fmt.Sprintf("%+.1f %%", 100*(float64(cur)-float64(old))/float64(old))
}

// spread shows the median and the range of sorted durations.
func spread(d []time.Duration) string {
	return fmt.Sprintf("%s (%s-%s)", format(median(d)), format(d[0]), format(d[len(d)-1]))
}

// durations returns the durations of the runs, sorted.
func durations(runs []Run, get func(Run) time.Duration) []time.Duration {
	d := make([]time.Duration, len(runs))
	for i, run := range runs {
		d[i] = get(run)
	}
	slices.Sort(d)
	return d
}

// median of sorted durations.
func median(d []time.Duration) time.Duration {
	if len(d)%2 == 1 {
		return d[len(d)/2]
	}
	return (d[len(d)/2-1] + d[len(d)/2]) / 2
}

// format shows durations below one second in ms, longer ones in s.
func format(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0f ms", float64(d)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.1f s", d.Seconds())
}

// CPUModel returns the CPU model from /proc/cpuinfo, or "" where that file does not exist (not Linux).
func CPUModel() string {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if name, value, ok := strings.Cut(scanner.Text(), ":"); ok && strings.TrimSpace(name) == "model name" {
			return strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}
