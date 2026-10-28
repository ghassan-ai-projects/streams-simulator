package app

import (
	"testing"
)

func TestMissedBarriersBetweenMetOnesNeverLoseTheHost(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) { h.silent = syncs%3 != 0 }}
	if _, err := session(t, host, 9); err != nil {
		t.Fatalf("two misses then a met barrier, three times, must not lose the host: %v", err)
	}
	if got := countOf(host.types(), "NSYNC"); got != 9 {
		t.Fatalf("%d barriers, want all 9", got)
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
