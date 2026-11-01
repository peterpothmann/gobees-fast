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

// Report is the comparison of the run times.
type Report struct {
	NetLogo NetLogo
	Gobees  []Run // gobees-fast in the test process, same seeds and days as NetLogo, in the same order
	// Programs holds gobees and gobees-fast as command line programs, nil if they did not run.
	Programs *Programs
	Threads  int // runs at the same time, in all models
	Machine  Machine
}

// Programs are the runs of the base commit (gobees) and of the current code (gobees-fast) as command line programs
// (RunPrograms), with the same seeds and days as NetLogo, in the same order.
type Programs struct {
	BaseCommit string
	Base       []Run
	Current    []Run
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
		"The example farm with %d seeds, %d days each, %d runs at the same time in all models. NetLogo 5.3.1 runs "+
		"headless with BehaviorSpace and the GUI switches off. Details: go/benchmark/README.md.\n\n",
		len(n.Runs), n.Days, r.Threads)

	b.WriteString("## NetLogo and gobees-fast\n\n" +
		"Setup is the time of NetLogo's Setup and of the gobees-fast setup, including reading the input files. The " +
		"days are timed separately, without output during the run.\n\n")
	b.WriteString("| | NetLogo median (min-max) | gobees-fast median (min-max) | NetLogo / gobees-fast |\n" +
		"| --- | ---: | ---: | ---: |\n")
	setup := func(run Run) time.Duration { return run.Setup }
	days := func(run Run) time.Duration { return run.Days }
	for _, row := range []struct {
		name string
		get  func(Run) time.Duration
	}{{"Setup", setup}, {fmt.Sprintf("%d days", n.Days), days}} {
		nl, gb := durations(n.Runs, row.get), durations(r.Gobees, row.get)
		fmt.Fprintf(&b, "| %s | %s | %s | %.0f |\n", row.name, spread(nl), spread(gb),
			float64(median(nl))/float64(median(gb)))
	}

	if p := r.Programs; p != nil {
		fmt.Fprintf(&b, "\n## NetLogo, gobees and gobees-fast\n\n"+
			"gobees is the base commit %s, gobees-fast the current code. Both run as command line programs without "+
			"output, one version after the other, and are timed as a whole, from the start of the program to its "+
			"end, including reading the input files. NetLogo's whole run is its Setup and its days, without the "+
			"start of Java.\n\n", p.BaseCommit)
		whole := func(run Run) time.Duration { return run.Whole() }
		nl, base, cur := durations(n.Runs, whole), durations(p.Base, whole), durations(p.Current, whole)
		b.WriteString("| | NetLogo median (min-max) | gobees median (min-max) | gobees-fast median (min-max) | " +
			"NetLogo / gobees-fast | gobees / gobees-fast |\n| --- | ---: | ---: | ---: | ---: | ---: |\n")
		fmt.Fprintf(&b, "| Whole run | %s | %s | %s | %.0f | %.2f |\n", spread(nl), spread(base), spread(cur),
			float64(median(nl))/float64(median(cur)), float64(median(base))/float64(median(cur)))
	}

	b.WriteString("\n## Runs\n\nThe runs of NetLogo and of the Go versions with the same seed are not the same " +
		"run: NetLogo's random numbers cannot be reproduced in Go. The colonies founded show that all did a " +
		"comparable amount of work.\n\n" +
		"| Run | Seed | NetLogo setup | NetLogo days | NetLogo colonies founded | gobees-fast setup | gobees-fast days | " +
		"gobees-fast colonies founded |\n| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for i, nl := range n.Runs {
		gb := r.Gobees[i]
		fmt.Fprintf(&b, "| %d | %g | %s | %s | %g | %s | %s | %g |\n", nl.Number, nl.Seed,
			format(nl.Setup), format(nl.Days), nl.ColoniesFounded, format(gb.Setup), format(gb.Days), gb.ColoniesFounded)
	}
	if p := r.Programs; p != nil {
		b.WriteString("\nAs programs:\n\n" +
			"| Run | Seed | gobees | gobees colonies founded | gobees-fast | gobees-fast colonies founded |\n" +
			"| ---: | ---: | ---: | ---: | ---: | ---: |\n")
		for i, base := range p.Base {
			cur := p.Current[i]
			fmt.Fprintf(&b, "| %d | %g | %s | %g | %s | %g |\n", base.Number, base.Seed, format(base.Program),
				base.ColoniesFounded, format(cur.Program), cur.ColoniesFounded)
		}
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
