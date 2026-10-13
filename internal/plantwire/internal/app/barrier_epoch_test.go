package app

import (
	"errors"
	"testing"
)

func TestAMetBarrierResetsTheMissedCount(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) { h.silent = syncs%3 != 0 }}
	if _, err := session(t, host, 9); err != nil {
		t.Fatalf("two misses then a met barrier, three times, must not lose the host: %v", err)
	}
	if got := countOf(host.types(), "NSYNC"); got != 9 {
		t.Fatalf("%d barriers, want all 9", got)
	}
}

func TestMissesSeparatedByAMetBarrierNeverAddUpToHostLoss(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) { h.silent = syncs != 3 }}
	if _, err := session(t, host, 6); !errors.Is(err, ErrHostLost) {
		t.Fatalf("err = %v, want host loss after three consecutive misses following the met barrier", err)
	}
	if got := countOf(host.types(), "NSYNC"); got != 6 {
		t.Fatalf("%d barriers, want 2 misses + 1 met + 3 misses = 6", got)
	}
}

func TestACommandFromAHigherEpochThanTheSessionsIsRefused(t *testing.T) {
	t.Parallel()
	now := sliceEnd(1)
	cases := map[string]int64{"c-lower": 0, "c-same": 1, "c-higher": 2}
	_, answers := commandSession(t, 2, func(h *scriptedHost, syncs int) {
		if syncs != 1 {
			return
		}
		for id, epoch := range cases {
			h.reply("DCMD", set(commandFields(id, "meter-1", "shed_load", map[string]any{"meter_id": id}, now), "authority_epoch", epoch))
		}
	}, nil)
	want := map[string]string{"c-lower": "DRECEIPT:refusedstale_epoch", "c-higher": "DRECEIPT:refusedstale_epoch", "c-same": "DRECEIPT:accepted"}
	for id, answer := range want {
		if got := answers[id]; len(got) == 0 || got[0] != answer {
			t.Errorf("%s: answers %v, want first %s", id, got, answer)
		}
	}
}
