package domain

import "container/heap"

var _ heap.Interface = (*priorityQueue)(nil)

type eventKind int

const (
	kindEmission eventKind = iota
	kindEffectStart
	kindBirth
	kindDeath
	kindEventCheck
)

type item struct {
	timeNS   int64
	tiebreak uint64
	kind     eventKind
	entity   string
	channel  string

	payload any
	index   int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int { return len(pq) }

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
