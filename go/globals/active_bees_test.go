package globals

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"
)

// netLogoNext is NetLogo's NextActiveBeeREP on a list sorted by time: the first bee goes back at the position of
// NetLogo's binary search (with its rounding), the new first bee is returned. The reference for ActiveBees.Next.
func netLogoNext(list []ecs.Entity, time map[ecs.Entity]float64) ecs.Entity {
	round := func(x float64) float64 { return math.Floor(x + 0.5) }
	activeBee := list[0]
	activeBeeTime := time[activeBee]
	minPosition := 0.0
	maxPosition := float64(len(list) - 1)
	currentPosition := round(maxPosition / 2)
	for maxPosition-minPosition > 1 {
		if time[list[int(currentPosition)]] > activeBeeTime {
			maxPosition = currentPosition
			currentPosition = round((currentPosition + minPosition) / 2)
		} else {
			minPosition = currentPosition
			currentPosition = round((maxPosition + currentPosition) / 2)
		}
	}
	beginningList := list[:int(currentPosition)]
	endList := list[int(currentPosition):]
	if len(list) > 1 {
		beginningList = beginningList[1:]
	} else {
		endList = endList[1:]
	}
	if len(endList) == 1 && activeBeeTime > time[endList[0]] {
		copy(list, list[1:])
		list[len(list)-1] = activeBee
	} else {
		copy(list, beginningList)
		list[len(beginningList)] = activeBee
	}
	return list[0]
}

func TestActiveBeesWorkedExamples(t *testing.T) {
	// the first bee A gets the time of the first value, the others are sorted
	for _, tc := range []struct {
		name  string
		times []float64
		want  string
	}{
		{"A 1810", []float64{1810, 900, 1200, 2000}, "BCAD"},
		{"A 2500", []float64{2500, 900, 1200, 2000}, "BCDA"},
		{"A 2000, tie with the last bee: before it", []float64{2000, 900, 1200, 2000}, "BCAD"},
		{"A 1200, tie with a bee in the middle: after it", []float64{1200, 900, 1200, 2000}, "BCAD"},
		{"three ties", []float64{10, 10, 10}, "BAC"},
		{"A smallest", []float64{5, 10, 20}, "ABC"},
		{"two bees, tie", []float64{20, 20}, "AB"},
		{"two bees", []float64{30, 20}, "BA"},
		{"one bee", []float64{30}, "A"},
	} {
		w := ecs.NewWorld()
		names := map[ecs.Entity]string{}
		time := map[ecs.Entity]float64{}
		list := []ecs.Entity{}
		for i, x := range tc.times {
			e := w.NewEntity()
			names[e] = string(rune('A' + i))
			time[e] = x
			list = append(list, e)
		}
		// A is first with time 0, then she acts and gets her time
		time[list[0]] = 0
		q := ActiveBees{}
		q.Set(list, func(e ecs.Entity) float64 { return time[e] })
		time[list[0]] = tc.times[0]
		first := q.Next(func(e ecs.Entity) float64 { return time[e] })
		got := ""
		for _, e := range q.Entities() {
			got += names[e]
		}
		if got != tc.want || first != q.Entities()[0] {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestActiveBeesEqualsNetLogo(t *testing.T) {
	// Random days of the activity loop with many ties: the queue and NetLogo's list give the same order after every
	// step. Steps: the first bee acts (her time grows by 0 to 4, or becomes the time of the last bee) and goes back
	// (Next), or a random bee dies (Remove).
	for seed := uint64(1); seed <= 200; seed++ {
		rng := rand.New(rand.NewPCG(seed, 7))
		w := ecs.NewWorld()
		time := map[ecs.Entity]float64{}
		n := 1 + rng.IntN(40)
		list := make([]ecs.Entity, n)
		for i := range list {
			list[i] = w.NewEntity()
			time[list[i]] = float64(rng.IntN(6))
		}
		slices.SortStableFunc(list, func(a, b ecs.Entity) int {
			switch {
			case time[a] < time[b]:
				return -1
			case time[a] > time[b]:
				return 1
			}
			return 0
		})
		q := ActiveBees{}
		get := func(e ecs.Entity) float64 { return time[e] }
		q.Set(list, get)
		ref := slices.Clone(list)

		for step := range 400 {
			if len(ref) == 0 {
				break
			}
			if rng.IntN(10) == 0 {
				e := ref[rng.IntN(len(ref))]
				ref = slices.DeleteFunc(ref, func(x ecs.Entity) bool { return x == e })
				q.Remove(e)
			} else {
				head := ref[0]
				switch rng.IntN(4) {
				case 0:
					time[head] = time[ref[len(ref)-1]] // a tie with the last bee
				default:
					time[head] += float64(rng.IntN(5))
				}
				a, b := netLogoNext(ref, time), q.Next(get)
				if a != b {
					t.Fatalf("seed %d, step %d: next bee %v, NetLogo %v", seed, step, b, a)
				}
			}
			if got := q.Entities(); !slices.Equal(got, ref) {
				t.Fatalf("seed %d, step %d: queue %v, NetLogo %v", seed, step, got, ref)
			}
		}
	}
}
