package domain

import (
	"testing"
)

func TestACommandIsReadFromItsRecordAndCheckedForFreshness(t *testing.T) {
	t.Parallel()
	in, ok := Decode([]byte(`{"type":"DCMD","protocol_version":1,"session_id":"s-1","seq":0,"lane":"control","command_id":"c-1",
		"idempotency_key":"sha256:80f43ea83bf83014412e6861b2b1a43a18df67abca91d82fb745f110442dfbac","device":"P1","command":"start",
		"args":{},"expected_boot_id":"b","policy_digest":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"deadline_ns":200,"not_before_ns":50,"authority_epoch":3,"fallback_on_lease_loss":"safe_off"}`))
	if !ok || in.Invalid != nil {
		t.Fatalf("decode: %v", in.Invalid)
	}
	c := CommandOf(in.Message)
	if c.ID != "c-1" || c.Device != "P1" || c.Boot != "b" || c.DeadlineNS != 200 || c.NotBeforeNS != 50 || c.Epoch != 3 || c.Content() == "" {
		t.Fatalf("command %+v", c)
	}
	for now, want := range map[int64]string{100: "", 201: "expired", 200 - MaxDeadlineAheadNS - 1: "out_of_range"} {
		if got := Freshness(c, now); got != want {
			t.Errorf("Freshness at %d = %q, want %q", now, got, want)
		}
	}
}

func TestTheLedgerHoldsSettlesAndTellsOtherContentApart(t *testing.T) {
	t.Parallel()
	ledger := Ledger{}
	c := Command{ID: "c-1", Key: "k", Device: "P1", Name: "start", Args: map[string]any{}, Boot: "b"}
	if entry, conflict := ledger.Recall(c); entry != nil || conflict {
		t.Fatal("an empty ledger recalled a command")
	}
	ledger.Hold(c)
	if entry, _ := ledger.Recall(c); entry == nil || !entry.Pending {
		t.Fatal("a held command is not pending")
	}
	if waiting := ledger.Settle(c, Answer{Status: "accepted", Outcome: "succeeded"}); len(waiting) != 1 {
		t.Fatalf("waiting %v", waiting)
	}
	if entry, _ := ledger.Recall(c); entry.Pending || entry.Answer.Outcome != "succeeded" {
		t.Fatalf("settled entry %+v", entry)
	}
	other := c
	other.Args = map[string]any{"pct": 1.0}
	if _, conflict := ledger.Recall(other); !conflict {
		t.Fatal("other content under the same key is not a conflict")
	}
	if waiting := ledger.Release(c); len(waiting) != 1 || ledger["k"] != nil {
		t.Fatal("release")
	}
}

func TestWireFaultsLeaveRecordsWithoutAFaultAlone(t *testing.T) {
	t.Parallel()
	records := []map[string]any{{"seq": uint64(0)}, {"seq": uint64(1)}}
	if got := ApplyWireFaults(records, nil); len(got) != 2 {
		t.Fatalf("no faults changed %v", got)
	}
	if got := ApplyWireFaults(records, []WireFault{{Seq: 1, Kind: FaultSwap}}); len(got) != 2 || Seq(got[1]["seq"]) != 1 {
		t.Fatalf("a swap of the last record = %v", got)
	}
	if Seq(3.0) != 3 || Seq(-1.0) != 0 || Seq("x") != 0 {
		t.Fatal("number reading")
	}
}
