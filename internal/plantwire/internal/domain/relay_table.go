package domain

import "sort"

const syncWindow = 4096

type seqPair struct{ gateway, up uint64 }

type seqTable struct{ pairs []seqPair }

func (t *seqTable) add(gateway, up uint64) {
	t.pairs = append(t.pairs, seqPair{gateway: gateway, up: up})
	if len(t.pairs) > syncWindow {
		t.pairs = t.pairs[len(t.pairs)-syncWindow:]
	}
}

func (t *seqTable) upFor(gateway uint64) (uint64, bool) {
	at := sort.Search(len(t.pairs), func(i int) bool { return t.pairs[i].gateway > gateway })
	if at == 0 {
		return 0, false
	}
	return t.pairs[at-1].up, true
}

const pendingSyncs = 4096

type syncLedger struct{ ups []uint64 }

func (s *syncLedger) add(up uint64) {
	s.ups = append(s.ups, up)
	if len(s.ups) > pendingSyncs {
		s.ups = s.ups[len(s.ups)-pendingSyncs:]
	}
}

func (s *syncLedger) answer(through uint64) bool {
	at := sort.Search(len(s.ups), func(i int) bool { return s.ups[i] > through })
	s.ups = s.ups[at:]
	return at > 0
}

func (t *seqTable) gatewayFor(up uint64) (uint64, bool) {
	at := sort.Search(len(t.pairs), func(i int) bool { return t.pairs[i].up > up })
	if at == 0 {
		return 0, false
	}
	return t.pairs[at-1].gateway, true
}
