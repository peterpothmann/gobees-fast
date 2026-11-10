package globals

import (
	"slices"

	"github.com/mlange-42/ark/ecs"
)

// ActiveBees is the daily activity queue: the active adult workers and queens, ordered by their personal time.
// wakeUp fills it in the morning (Set), the activity loop takes the next bee (Next), a dying bee is removed (Remove).
// NetLogo: ActiveBeesSortedList, a list sorted by personal time, and NextActiveBeeREP.
//
// The queue is a binary heap ordered by (Time, Seq). Seq numbers the insertions: bees with the same time stay in the
// order in which they came into the queue. NetLogo's list has one exception (a tie with the last bee of the list,
// see Next), which the queue leaves out (docs/changes.md).
type ActiveBees struct {
	// Queue is the binary heap: Queue[0] is the first bee, the rest is not sorted. Entities gives the sorted order.
	Queue []QueuedBee

	seq uint64 // the last sequence number given
}

// QueuedBee is a bee in the activity queue with the personal time it had when it was put in.
type QueuedBee struct {
	Bee  ecs.Entity
	Time float64 // [s]
	Seq  uint64
}

// before is true if a comes before b in the queue.
func (a QueuedBee) before(b QueuedBee) bool {
	return a.Time < b.Time || (a.Time == b.Time && a.Seq < b.Seq)
}

// Set fills the queue with the bees of list, which is sorted by the personal time (time). Bees with the same time
// keep their order in list.
func (q *ActiveBees) Set(list []ecs.Entity, time func(ecs.Entity) float64) {
	q.Queue = q.Queue[:0]
	q.seq = 0
	for _, e := range list {
		q.Queue = append(q.Queue, QueuedBee{e, time(e), q.next()})
	}
	// a sorted list is a valid heap
}

// Len returns the number of bees in the queue.
func (q *ActiveBees) Len() int {
	return len(q.Queue)
}

// Next puts the first bee back into the queue with its personal time now (time) and returns the new first bee. The
// queue must not be empty.
//
// The bee goes behind all bees with a time up to hers. NetLogo's binary search does the same, with one exception: if
// her time equals the time of the last bee of the list, she goes before that bee. NetLogo: NextActiveBeeREP.
func (q *ActiveBees) Next(time func(ecs.Entity) float64) ecs.Entity {
	head := &q.Queue[0]
	head.Time, head.Seq = time(head.Bee), q.next()
	q.down(0)
	return q.Queue[0].Bee
}

// Remove removes the bee e from the queue. A bee that is not in the queue is ignored.
func (q *ActiveBees) Remove(e ecs.Entity) {
	i := q.indexOf(e)
	if i < 0 {
		return
	}
	n := len(q.Queue) - 1
	q.Queue[i] = q.Queue[n]
	q.Queue = q.Queue[:n]
	if i < n {
		q.down(i)
		q.up(i)
	}
}

// Entities returns the bees in the order of the queue, e.g. for tests. Not for the activity loop: it sorts a copy.
func (q *ActiveBees) Entities() []ecs.Entity {
	sorted := slices.Clone(q.Queue)
	slices.SortFunc(sorted, func(a, b QueuedBee) int {
		if a.before(b) {
			return -1
		}
		if b.before(a) {
			return 1
		}
		return 0
	})
	result := make([]ecs.Entity, len(sorted))
	for i, b := range sorted {
		result[i] = b.Bee
	}
	return result
}

func (q *ActiveBees) next() uint64 {
	q.seq++
	return q.seq
}

func (q *ActiveBees) indexOf(e ecs.Entity) int {
	for i := range q.Queue {
		if q.Queue[i].Bee == e {
			return i
		}
	}
	return -1
}

// down moves the bee at i down the heap to its place.
func (q *ActiveBees) down(i int) {
	h := q.Queue
	n := len(h)
	b := h[i] // the bee moves down, the smaller child moves up into the hole
	for {
		c := 2*i + 1
		if c >= n {
			break
		}
		if r := c + 1; r < n && h[r].before(h[c]) {
			c = r
		}
		if !h[c].before(b) {
			break
		}
		h[i] = h[c]
		i = c
	}
	h[i] = b
}

// up moves the bee at i up the heap to its place.
func (q *ActiveBees) up(i int) {
	h := q.Queue
	for i > 0 {
		parent := (i - 1) / 2
		if !h[i].before(h[parent]) {
			return
		}
		h[i], h[parent] = h[parent], h[i]
		i = parent
	}
}
