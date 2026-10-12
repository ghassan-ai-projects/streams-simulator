package domain

import (
	"errors"
	"testing"
)

func TestControlRecordsTransmitFirstAndSequencesHaveNoGaps(t *testing.T) {
	t.Parallel()
	out := NewTransmitter("s-1")
	out.Queue(DeviceData("P1", "b", false, []map[string]any{Sample(1, 1.0, 1)}))
	out.Queue(Sync(0, 1))
	out.Queue(DeviceData("P1", "b", false, []map[string]any{Sample(1, 2.0, 2)}))
	records := out.Drain()
	if records[0]["type"] != "NSYNC" || records[0]["seq"] != uint64(0) || records[2]["seq"] != uint64(2) || out.LastSeq() != 2 {
		t.Fatalf("transmitted %v", records)
	}
}

func TestTheReceiverTellsInOrderDuplicateAndGap(t *testing.T) {
	t.Parallel()
	var in Receiver
	if in.Accept(0) != InOrder || in.Accept(0) != Duplicate || in.Accept(2) != Gap || in.Accept(3) != InOrder || in.Accept(1) != Duplicate {
		t.Fatal("receiver arrivals")
	}
}

func TestNegotiationTakesTheHostsVersionOrRefuses(t *testing.T) {
	t.Parallel()
	if version, err := Negotiate(map[string]any{"online": true, "protocol_version": 1, "protocol_min": 1, "protocol_max": 1}, 2); err != nil || version != 1 {
		t.Fatalf("negotiate = %d, %v", version, err)
	}
	for _, state := range []map[string]any{{"online": false, "refusal": "environment_mismatch"}, {"online": true, "protocol_version": 3, "protocol_min": 1, "protocol_max": 3}} {
		if _, err := Negotiate(state, 2); !errors.Is(err, ErrRefused) {
			t.Fatalf("negotiate %v = %v, want a refusal", state, err)
		}
	}
}
