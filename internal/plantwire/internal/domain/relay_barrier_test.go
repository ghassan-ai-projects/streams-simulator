package domain

import (
	"encoding/json"
	"testing"
)

func barrierSync(throughSeq uint64) Record {
	return Sync(throughSeq, 1767225660000000000)
}

func hostSync(throughSeq uint64) Record {
	return Record{Type: "HSYNC", Lane: LaneControl, Fields: map[string]any{"through_seq": throughSeq}}
}

func barrierRig(t *testing.T) *relayRig {
	t.Helper()
	g := newRelayRig(t)
	g.bringUp()
	g.gwEdited(validData(), func(r map[string]any) { delete(r, "metrics") })
	g.gw(validData())
	return g
}

func TestABarrierTheGatewayRaisesIsRaisedToTheHostInTheRelaysOwnSequence(t *testing.T) {
	t.Parallel()
	g := barrierRig(t)
	moves := g.gw(barrierSync(4))
	if len(moves.Up) != 1 || moves.Up[0]["type"] != "NSYNC" {
		t.Fatalf("up = %v, want the NSYNC alone", typesOf(moves.Up))
	}
	sync := moves.Up[0]
	if sync["through_seq"] != uint64(3) || sync["seq"] != uint64(4) || sync["world_time_ns"] != json.Number("1767225660000000000") {
		t.Fatalf("NSYNC = %v, want through_seq 3 (the relay's seq of gateway record 4), own seq 4 and the world time unchanged", sync)
	}
}

func TestTheHostsAnswerToABarrierReturnsInTheGatewaysSequence(t *testing.T) {
	t.Parallel()
	g := barrierRig(t)
	g.gw(barrierSync(4))
	moves := g.hostSends(hostSync(4))
	if len(moves.Down) != 1 || moves.Down[0]["type"] != "HSYNC" || moves.Down[0]["through_seq"] != uint64(5) {
		t.Fatalf("down = %v, want HSYNC through_seq 5, the seq the gateway gave its NSYNC", moves.Down)
	}
}

func TestALateAnswerToAnEarlierBarrierDoesNotCoverALaterOne(t *testing.T) {
	t.Parallel()
	g := barrierRig(t)
	g.gw(barrierSync(4))
	moves := g.hostSends(hostSync(3))
	if len(moves.Down) != 1 || moves.Down[0]["through_seq"] != uint64(4) {
		t.Fatalf("down = %v, want through_seq 4, below the gateway's NSYNC at 5", moves.Down)
	}
}

func TestAnAnswerThatNoRelayedRecordCoversIsDroppedNotInvented(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.relay.GatewayLost()
	g.relay.GatewayAttached()
	g.gateway = NewTransmitter("gw-2")
	g.gw(hardwareBirth(t, nil))
	if moves := g.hostSends(hostSync(2)); moves.Down != nil || g.relay.Stats().Dropped != 1 {
		t.Fatalf("down = %v dropped %d, want the answer to the earlier session's barrier dropped", moves.Down, g.relay.Stats().Dropped)
	}
	if moves := g.hostSends(hostSync(5)); len(moves.Down) != 1 || moves.Down[0]["through_seq"] != uint64(0) {
		t.Fatalf("down = %v, want the answer to the new session mapped to its first record", moves.Down)
	}
}

func TestABarrierRaisedForRecordsTheNewHostSessionNeverSawIsDropped(t *testing.T) {
	t.Parallel()
	g := newRelayRig(t)
	g.bringUp()
	g.gw(hardwareBirth(t, nil))
	g.relay.HostLost()
	g.relay.HostAttached()
	g.newHostSession("relay-h2")
	g.hostSends(hostState(7, 0))
	if moves := g.gw(barrierSync(2)); moves.Up != nil {
		t.Fatalf("up = %v, want the NSYNC dropped: it covers records the new host session never had", moves.Up)
	}
}

func TestTheSequenceTableFindsTheLastRecordAtOrBeforeAndForgetsTheOldest(t *testing.T) {
	t.Parallel()
	var table seqTable
	for gateway := range uint64(syncWindow + 100) {
		table.add(gateway*2, gateway)
	}
	if up, found := table.upFor(2*syncWindow + 1); !found || up != syncWindow {
		t.Errorf("upFor(%d) = %d, %v, want %d", 2*syncWindow+1, up, found, syncWindow)
	}
	if gateway, found := table.gatewayFor(syncWindow + 50); !found || gateway != 2*(syncWindow+50) {
		t.Errorf("gatewayFor(%d) = %d, %v, want %d", syncWindow+50, gateway, found, 2*(syncWindow+50))
	}
	if _, found := table.upFor(0); found {
		t.Error("upFor(0) found a record the table should have forgotten")
	}
	if _, found := table.gatewayFor(5); found {
		t.Error("gatewayFor(5) found a record the table should have forgotten")
	}
	if len(table.pairs) != syncWindow {
		t.Errorf("table holds %d pairs, want the %d most recent", len(table.pairs), syncWindow)
	}
}
