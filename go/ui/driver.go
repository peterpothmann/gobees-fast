//go:build ui || windows || (darwin && cgo)

package ui

import (
	"fmt"
	"time"

	"gobees/sys"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
)

// command tells the model how far to run.
type command int

const (
	toDayEnd command = iota // to the end of the day
	toPart                  // to the end of the next part of Morning, Day or Evening (or the end of the day)
	toTurn                  // to the end of the next bee turn (or the end of the day)
)

// stopKind is where the model stopped.
type stopKind int

const (
	stopDayEnd stopKind = iota // the day is over, also the last day of the run
	stopPart                   // after a part
	stopTurn                   // after a bee turn
)

// stop is where the model stopped and waits for the next command.
type stop struct {
	kind         stopKind
	system, part string     // the part done last (stopPart) or running (stopTurn: Day, activities)
	partIndex    int        // parts of the day done, 1 to the number of parts
	turn         int        // bee turns of the day so far
	bee          ecs.Entity // stopTurn: the bee whose turn it was, maybe dead
	finished     bool       // stopDayEnd: the run is over, Finalize has run
	err          error      // a panic of the model: the run is over
}

// partTime is the time one part took on one day, without the time the model waited for the window.
type partTime struct {
	name string // system.part, e.g. "Morning.flowerRefill"
	took time.Duration
}

// driver runs the model in a goroutine of its own, so that the run can stop in the middle of a day: after a part
// (sys.Hooks.AfterPart) and after a bee turn (sys.Hooks.AfterTurn). The window sends a command and waits for the
// stop. While the model waits for the next command, the window may read the world: every command and every stop
// goes through an unbuffered channel, which orders the reads of the window after the writes of the model.
type driver struct {
	app   *app.App
	cmds  chan command
	stops chan stop

	// the model goroutine only, read by the window after a stop
	target          command
	partIndex, turn int
	mark            time.Duration // the model works since mark (now)
	work            time.Duration // work of the current part so far
	today           []partTime    // the parts done today
	lastDay         []partTime    // the parts of the last full day
	total           map[string]time.Duration
	order           []string // the parts in schedule order, from the first day
	fullDays        int      // days in total

	inDay bool // the window only: the model stopped in the middle of a day
}

// newDriver adds the hooks to a, which must not be initialized yet.
func newDriver(a *app.App) *driver {
	d := &driver{app: a, cmds: make(chan command), stops: make(chan stop), total: map[string]time.Duration{}}
	ecs.AddResource(a.World, &sys.Hooks{AfterPart: d.afterPart, AfterTurn: d.afterTurn})
	return d
}

// start starts the model goroutine. a must be initialized.
func (d *driver) start() {
	go d.loop()
}

// run sends a command and returns where the model stopped. Window only.
func (d *driver) run(cmd command) stop {
	d.cmds <- cmd
	s := <-d.stops
	d.inDay = s.kind != stopDayEnd
	return s
}

// close ends the model goroutine. The model must have stopped at the end of a day.
func (d *driver) close() {
	close(d.cmds)
}

// loop runs one day per command that arrives at the end of a day. The commands in the middle of a day arrive in
// pause.
func (d *driver) loop() {
	for cmd := range d.cmds {
		d.target = cmd
		d.partIndex, d.turn, d.today = 0, 0, nil
		d.work, d.mark = 0, now()
		s := stop{kind: stopDayEnd}
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.err, s.finished = fmt.Errorf("%v", r), true
				}
			}()
			if !d.app.Update() {
				d.app.Finalize()
				s.finished = true
			}
		}()
		d.lastDay = d.today
		for _, p := range d.today {
			d.total[p.name] += p.took
		}
		if d.order == nil {
			for _, p := range d.today {
				d.order = append(d.order, p.name)
			}
		}
		d.fullDays++
		d.stops <- s
	}
}

func (d *driver) afterPart(system, part string) {
	d.work += now() - d.mark
	d.partIndex++
	d.today = append(d.today, partTime{system + "." + part, d.work})
	d.work = 0
	if d.target == toPart {
		d.pause(stop{kind: stopPart, system: system, part: part, partIndex: d.partIndex, turn: d.turn})
	}
	d.mark = now()
}

func (d *driver) afterTurn(bee ecs.Entity) {
	d.work += now() - d.mark
	d.turn++
	if d.target == toTurn {
		d.pause(stop{kind: stopTurn, system: "Day", part: "activities", partIndex: d.partIndex, turn: d.turn,
			bee: bee})
	}
	d.mark = now()
}

// pause reports the stop and waits for the next command.
func (d *driver) pause(s stop) {
	d.stops <- s
	d.target = <-d.cmds
}
