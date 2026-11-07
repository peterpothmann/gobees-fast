package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gobees/obs"
	"gobees/params"

	"github.com/alecthomas/kong"
)

// defaultOutputFile is the daily output file of -o, relative to the working directory (go/ with go run .).
const defaultOutputFile = "daily_output.csv"

// cli holds the command-line flags.
type cli struct {
	Config      []string `short:"c" type:"path" sep:"none" placeholder:"PATH" help:"Run file (YAML, see package input). Its landscape files are relative to its folder. Repeat -c for several run files: their runs share the threads, each writes its own output. Default: the example farm, which is built in. tools/netlogo2gobees converts NetLogo input files into run files."`
	Seed        *int64   `short:"s" placeholder:"SEED" help:"Random seed of the first run, 1 or more. Default: run.seed of the run file."`
	Days        *int     `placeholder:"DAYS" help:"Number of days (ticks) to run. 0 means no limit: the run only ends with stopExtinct when the population is extinct. Default: run.days of the run file, else ${default_days}."`
	Output      bool     `short:"o" help:"Write the daily output (the daily totals and further values, see obs.DailyOutputExtended) to ${default_output_file}, one row per tick and run, tab separated, with the columns run (from 1), seed, the varied keys and t in front, and the run info next to it (${default_run_info}, a run file that repeats the run). With several run files, the name of each run file goes in front of the extension."`
	OutputFile  string   `short:"f" type:"path" placeholder:"PATH" help:"Write the daily output to PATH instead (implies -o). The run info is PATH without extension plus _run_info.yaml."`
	Columns     []string `placeholder:"A,B,C" help:"Columns of the daily output, comma separated, in this order (needs -o or -f). The column ticks is always first. Default: output.columns of the run file, else all columns. --list-columns shows the names."`
	ListColumns bool     `help:"Print the columns of the daily output with a description and exit."`
	Runs        *int     `placeholder:"N" help:"Number of runs (of each combination of the variations), 1 or more. The run i has the seed SEED + i - 1. Default: run.runs of the run file."`
	Threads     *int     `placeholder:"T" help:"Number of runs at the same time, 1 or more (on a cluster: the CPUs of the job). It does not change the output. Default: run.threads of the run files, which must then be equal."`
	UI          bool     `name:"ui" help:"Show the run in a window while it runs: map, daily plots, colonies, flower patches. The run, its output and run info are the same as without the window. One run only (one run file, runs 1, no variations). Closing the window ends the run on that day. On Linux only in a build with the tag ui: go run -tags ui . --ui"`

	// for the window, not shown in --help
	UIExplain    bool   `name:"ui-explain" hidden:"" help:"Open only the window that explains the model (started by the button of --ui)."`
	UIScreenshot string `name:"ui-screenshot" type:"path" hidden:"" placeholder:"PATH" help:"With --ui: run to the end at once, save a screenshot of the window to PATH (PNG) and close it."`
}

// outputFile returns the daily output file of the flags, "" without output.
func (c *cli) outputFile() string {
	if c.OutputFile == "" && c.Output {
		return defaultOutputFile
	}
	return c.OutputFile
}

// listColumns writes the columns of the daily output with their description to w.
func listColumns(w io.Writer) {
	columns := obs.AllColumns()
	width := 0
	for _, name := range columns {
		width = max(width, len(name))
	}
	for _, name := range columns {
		fmt.Fprintf(w, "%-*s  %s\n", width, name, obs.ColumnDoc(name))
	}
}

// newParser returns the parser of the command-line flags into c.
func newParser(c *cli) (*kong.Kong, error) {
	defaults := params.Default()
	return kong.New(c,
		kong.Name("gobees"),
		kong.Description("Go port of the NetLogo model BEE-STEWARD. Without flags, it runs the example farm."),
		kong.UsageOnError(),
		kong.Vars{
			"default_output_file": defaultOutputFile,
			"default_run_info":    runInfoPath(defaultOutputFile),
			"default_days":        strconv.Itoa(defaults.Termination.MaxTicks),
		},
	)
}

func main() {
	var c cli
	parser, err := newParser(&c)
	if err != nil {
		log.Fatal(err)
	}
	_, err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	if err := c.run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run runs the model for the flags and writes a summary to w: every combination of the variations of every run file,
// runs times each. With an output file, it writes the daily output of each run file and, before the runs, its run
// info.
func (c *cli) run(w io.Writer) error {
	if c.ListColumns {
		listColumns(w)
		return nil
	}
	if c.UIExplain {
		return explainWindow()
	}
	if len(c.Columns) > 0 && c.outputFile() == "" {
		return errors.New("--columns needs an output: add -o or -f")
	}
	if c.UIScreenshot != "" && !c.UI {
		return errors.New("--ui-screenshot needs --ui")
	}
	started := time.Now()
	configs := c.Config
	if len(configs) == 0 {
		configs = []string{""} // the example farm
	}
	var experiments []*experiment
	for _, config := range configs {
		e, err := c.newExperiment(config)
		if err != nil {
			return err
		}
		experiments = append(experiments, e)
	}
	threads, err := c.threadCount(experiments)
	if err != nil {
		return err
	}
	if err := outputFiles(c.outputFile(), experiments); err != nil {
		return err
	}
	var watch watchFunc
	if c.UI {
		if err := checkWatch(experiments); err != nil {
			return err
		}
		if watch, err = watcher(experiments[0].title(), c.UIScreenshot); err != nil {
			return err
		}
	}

	rs := runSet{threads: threads, watch: watch}
	writers := make([]*outputWriter, len(experiments))
	defer func() {
		for _, o := range writers {
			if o != nil {
				o.close()
			}
		}
	}()
	for i, e := range experiments {
		rs.jobs = append(rs.jobs, e.jobs(i, len(experiments) > 1)...)
		if e.output == "" {
			continue
		}
		// before the runs, so that a run that fails can be repeated
		info := newRunInfo(e.files, e.config, &e.params[0], e.runs, threads, e.output, e.columns, started)
		if err := info.write(runInfoPath(e.output)); err != nil {
			return err
		}
		if writers[i], err = newOutputWriter(e.output, e.header()); err != nil {
			return err
		}
	}

	err = rs.run(func(i int, r runResult) error {
		j := rs.jobs[i]
		if o := writers[j.exp]; o != nil {
			if err := o.write(r.rows); err != nil {
				return err
			}
		}
		r.rows = nil
		r.label = j.label
		experiments[j.exp].results = append(experiments[j.exp].results, r)
		return nil
	})
	for i, o := range writers {
		if o != nil {
			if cerr := o.close(); err == nil {
				err = cerr
			}
			writers[i] = nil
		}
	}
	if err != nil {
		return err
	}

	for _, e := range experiments {
		e.summary(w, len(experiments) > 1, min(threads, len(rs.jobs)))
	}
	return nil
}

// checkWatch reports an error unless the experiments are one run: the window (--ui) shows one run.
func checkWatch(experiments []*experiment) error {
	e := experiments[0]
	if len(experiments) > 1 || len(e.combinations) > 1 || e.runs > 1 {
		return fmt.Errorf("--ui shows one run: one run file (-c), runs 1 (--runs 1), no variations; here %d run "+
			"files, %d combinations, %d runs", len(experiments), len(e.combinations), e.runs)
	}
	return nil
}

// title names the run file of the experiment in the window.
func (e *experiment) title() string {
	if e.config == "" {
		return "example farm"
	}
	return filepath.Base(e.config)
}

// summary writes the summary of the experiment to w. One run keeps the lines that the benchmark reads.
func (e *experiment) summary(w io.Writer, several bool, threads int) {
	if several {
		fmt.Fprintf(w, "%s:\n", e.config)
	}
	fmt.Fprintf(w, "species: %v\n", e.results[0].species)
	seed := e.params[0].Seed
	if len(e.results) == 1 {
		r := e.results[0]
		fmt.Fprintf(w, "seed: %d\n", seed)
		fmt.Fprintf(w, "colonies founded: %v, bees died: %v\n", r.founded, r.died)
		fmt.Fprintf(w, "finished after %d days\n", r.days)
	} else {
		fmt.Fprintf(w, "runs: %d, seeds: %d to %d, threads: %d", e.runs, seed, seed+int64(e.runs)-1, threads)
		if len(e.combinations) > 1 {
			fmt.Fprintf(w, ", combinations: %d", len(e.combinations))
		}
		fmt.Fprintln(w)
		for _, r := range e.results {
			fmt.Fprintf(w, "%s: colonies founded: %v, bees died: %v, finished after %d days\n", r.label, r.founded,
				r.died, r.days)
		}
	}
	if e.output != "" {
		fmt.Fprintf(w, "daily output written to %s, run info to %s\n", e.output, runInfoPath(e.output))
	}
}
