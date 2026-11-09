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
// order in which they came into the queue, as in NetLogo's list. Next keeps the one exception of NetLogo (a tie with
// the last bee of the list). The order of the queue is therefore exactly the order of NetLogo's list.
type ActiveBees struct {
	// Queue is the binary heap: Queue[0] is the first bee, the rest is not sorted. Entities gives the sorted order.
	Queue []QueuedBee

	seq  uint64    // the next sequence number
	last QueuedBee // the last bee of NetLogo's list: the largest (Time, Seq). Valid while Queue is not empty.
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
	if len(q.Queue) > 0 {
		q.last = q.Queue[len(q.Queue)-1]
	}
}

// Len returns the number of bees in the queue.
func (q *ActiveBees) Len() int {
	return len(q.Queue)
}

// Next puts the first bee back into the queue with its personal time now (time) and returns the new first bee. The
// queue must not be empty.
//
// NetLogo's binary search puts the bee behind all bees with a time up to hers, with one exception: if her time
// equals the time of the last bee of the list, she goes before that bee. Next does the same. NetLogo:
// NextActiveBeeREP.
func (q *ActiveBees) Next(time func(ecs.Entity) float64) ecs.Entity {
	head := &q.Queue[0]
	t := time(head.Bee)
	if len(q.Queue) == 1 {
		head.Time = t
		q.last = *head
		return head.Bee
	}
	// with two bees or more the first one is not the last one: the keys differ
	switch {
	case t == q.last.Time:
		// before the last bee: she takes its sequence number, the last bee gets a new one and stays last
		i := q.indexOf(q.last.Bee)
		head.Time, head.Seq = t, q.last.Seq
		q.Queue[i].Seq = q.next()
		q.last = q.Queue[i]
		q.down(i) // the key of the last bee grew
	case t > q.last.Time:
		head.Time, head.Seq = t, q.next()
		q.last = *head
	default:
		head.Time, head.Seq = t, q.next()
	}
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
	if e == q.last.Bee && n > 0 {
		q.last = q.Queue[0]
		for _, b := range q.Queue[1:] {
			if q.last.before(b) {
				q.last = b
			}
		}
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
