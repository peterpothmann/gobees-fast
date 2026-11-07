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
	species string  // the species and their initial queens, for the summary
	founded float64 // colonies founded
	died    float64 // bees died
	days    int     // days run
	rows    []byte  // the rows of the daily output, nil without output
	err     error
}

// runSet is a number of runs of the same model. The run i (counted from 0) has the seed p.Seed + i.
type runSet struct {
	p       params.Params
	runs    int
	threads int
	// columns of the daily output. nil: no output. Every row starts with the columns run (counted from 1) and seed.
	columns []string
}

// errSkipped is the result of a run that did not start because another run failed.
var errSkipped = errors.New("not run")

// one runs the run i and returns its result. A panic of the model is returned as an error with its stack.
func (rs runSet) one(i int) (res runResult) {
	p := rs.p // the input is shared by all runs, the model does not change it
	p.Seed += int64(i)
	defer func() {
		if r := recover(); r != nil {
			res.err = fmt.Errorf("run %d (seed %d): panic: %v\n%s", i+1, p.Seed, r, debug.Stack())
		}
	}()

	a := schedule.Default(p, nil)
	// after all model systems: colonies founded and bees died for the summary (the benchmark reads it)
	counters := obs.Counters{}
	a.AddSystem(&counters)
	var rows *csvRows
	if rs.columns != nil {
		// after all model systems, so that each row holds the state at the end of the day.
		// The comparisons with the reference commit (go/compare) read this output.
		rows = &csvRows{observer: &obs.DailyOutputExtended{Columns: rs.columns},
			prefix: fmt.Sprintf("%d\t%d\t", i+1, p.Seed)}
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

// run runs all runs with at most rs.threads at the same time. It calls done in the calling goroutine with the result
// of every run in the order of the runs, so that the output does not depend on the number of threads. The first run
// that fails (or the first error of done) ends the call: runs that have not started are skipped, running ones finish.
func (rs runSet) run(done func(i int, r runResult) error) error {
	results := make([]chan runResult, rs.runs)
	for i := range results {
		results[i] = make(chan runResult, 1)
	}
	var next atomic.Int64
	var stop atomic.Bool
	var wg sync.WaitGroup
	for range min(rs.threads, rs.runs) {
		wg.Go(func() {
			for {
				i := int(next.Add(1)) - 1
				if i >= rs.runs {
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

	for i := range rs.runs {
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

// newOutputWriter creates the file path (and missing folders) and writes the header: run, seed, t, then the columns.
func newOutputWriter(path string, columns []string) (*outputWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	o := &outputWriter{file: f, writer: bufio.NewWriter(f)}
	header := append([]string{"run", "seed", "t"}, columns...)
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
