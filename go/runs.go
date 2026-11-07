package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"gobees/globals"
	"gobees/obs"
	"gobees/params"
	"gobees/schedule"

	"github.com/mlange-42/ark-tools/observer"
	"github.com/mlange-42/ark/ecs"
)

// csvRows is a system that formats one row of the observer per tick, like reporter.CSV of ark-tools: the column t
// (the number of the tick, counted from 0), then the values, tab separated, each value as the shortest decimal that
// reads back to the same float64. The rows are kept in memory, each run has its own csvRows. prefix is written in
// front of every row.
type csvRows struct {
	observer observer.Row
	prefix   string
	rows     []byte
	step     int
}

func (s *csvRows) Initialize(w *ecs.World) { s.observer.Initialize(w) }

func (s *csvRows) Update(w *ecs.World) {
	s.observer.Update(w)
	s.rows = append(s.rows, s.prefix...)
	s.rows = strconv.AppendInt(s.rows, int64(s.step), 10)
	for _, v := range s.observer.Values(w) {
		s.rows = append(s.rows, '\t')
		s.rows = strconv.AppendFloat(s.rows, v, 'f', -1, 64)
	}
	s.rows = append(s.rows, '\n')
	s.step++
}

func (s *csvRows) Finalize(w *ecs.World) {}

// runResult is the result of one run.
type runResult struct {
	label   string  // the run in the summary (job.label)
	species string  // the species and their initial queens, for the summary
	founded float64 // colonies founded
	died    float64 // bees died
	days    int     // days run
	rows    []byte  // the rows of the daily output, nil without output
	err     error
}

// job is one run of the model.
type job struct {
	p       params.Params // the parameters, with the seed of the run
	columns []string      // the columns of the daily output, nil: no output
	prefix  string        // in front of each row: run, seed and the values of the variations, each with a tab after it
	label   string        // the run in messages, e.g. "run 2 (seed 6, parameters.badgers=2)"
	exp     int           // the index of its experiment
}

// runSet is a list of runs, of one or more experiments.
type runSet struct {
	jobs    []job
	threads int
}

// errSkipped is the result of a run that did not start because another run failed.
var errSkipped = errors.New("not run")

// one runs the job i and returns its result. A panic of the model is returned as an error with its stack.
func (rs runSet) one(i int) (res runResult) {
	j := rs.jobs[i]
	p := j.p // the input is shared by all runs, the model does not change it
	defer func() {
		if r := recover(); r != nil {
			res.err = fmt.Errorf("%s: panic: %v\n%s", j.label, r, debug.Stack())
		}
	}()

	a := schedule.Default(p, nil)
	// after all model systems: colonies founded and bees died for the summary (the benchmark reads it)
	counters := obs.Counters{}
	a.AddSystem(&counters)
	var rows *csvRows
	if j.columns != nil {
		// after all model systems, so that each row holds the state at the end of the day.
		// The comparisons with the reference commit (go/compare) read this output.
		rows = &csvRows{observer: &obs.DailyOutputExtended{Columns: j.columns}, prefix: j.prefix}
		a.AddSystem(rows)
	}
	a.Run()

	res.species = fmt.Sprint(ecs.GetResource[params.Derived](a.World).InitialQueens)
	res.founded = counters.TotalColoniesEverProduced
	res.died = counters.TotalBeesEverDied
	res.days = int(ecs.GetResource[globals.Ticks](a.World).Ticks)
	if rows != nil {
		res.rows = rows.rows
	}
	return res
}

// run runs all jobs with at most rs.threads at the same time, the jobs of all experiments in one pool. It calls done
// in the calling goroutine with the result of every job in the order of the jobs, so that the output does not depend
// on the number of threads. The first run that fails (or the first error of done) ends the call: runs that have not
// started are skipped, running ones finish.
func (rs runSet) run(done func(i int, r runResult) error) error {
	n := len(rs.jobs)
	results := make([]chan runResult, n)
	for i := range results {
		results[i] = make(chan runResult, 1)
	}
	var next atomic.Int64
	var stop atomic.Bool
	var wg sync.WaitGroup
	for range min(rs.threads, n) {
		wg.Go(func() {
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				if stop.Load() {
					results[i] <- runResult{err: errSkipped}
					continue
				}
				results[i] <- rs.one(i)
			}
		})
	}
	defer wg.Wait()

	for i := range n {
		r := <-results[i]
		err := r.err
		if err == nil {
			err = done(i, r)
		}
		if err != nil {
			stop.Store(true)
			return err
		}
	}
	return nil
}

// outputWriter writes the daily output of all runs into one file.
type outputWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// newOutputWriter creates the file path (and missing folders) and writes the header.
func newOutputWriter(path string, header []string) (*outputWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	o := &outputWriter{file: f, writer: bufio.NewWriter(f)}
	if _, err := o.writer.WriteString(strings.Join(header, "\t") + "\n"); err != nil {
		f.Close()
		return nil, err
	}
	return o, nil
}

// write appends rows.
func (o *outputWriter) write(rows []byte) error {
	_, err := o.writer.Write(rows)
	return err
}

// close flushes and closes the file.
func (o *outputWriter) close() error {
	err := o.writer.Flush()
	if cerr := o.file.Close(); err == nil {
		err = cerr
	}
	return err
}
