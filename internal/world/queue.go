package world

// The world advances by popping a priority queue keyed by
// (event_time_ns, tiebreak_seq), with tiebreak_seq a monotonic counter
// assigned at scheduling time. The tiebreak makes the processing order a
// pure function of schedule order, independent of insertion order.

import "container/heap"

// Compile-time proof that priorityQueue satisfies heap.Interface.
var _ heap.Interface = (*priorityQueue)(nil)

// eventKind identifies what a scheduled event does.
type eventKind int

const (
	kindEmission    eventKind = iota // a channel emission
	kindEffectStart                  // an effect kick becomes active (dead time elapsed)
	kindBirth                        // autonomous entity birth
	kindDeath                        // autonomous entity retirement
)

// item is one scheduled event.
type item struct {
	timeNS   int64
	tiebreak uint64
	kind     eventKind
	entity   string
	channel  string
	// payload carries per-kind data: effect start -> *kick, birth -> nil.
	payload any
	index   int // index in the heap, maintained by heap.Interface
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

// Less orders by (timeNS, tiebreak): the tiebreak sequence assigned at
// scheduling time decides simultaneity.
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].timeNS != pq[j].timeNS {
		return pq[i].timeNS < pq[j].timeNS
	}
	return pq[i].tiebreak < pq[j].tiebreak
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	it := x.(*item)
	it.index = n
	*pq = append(*pq, it)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	it := old[n-1]
	old[n-1] = nil
	it.index = -1
	*pq = old[:n-1]
	return it
}
