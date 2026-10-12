package app

import (
	"bytes"
	"context"
	"testing"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func liveScript(h *scriptedHost, syncs int) {
	switch syncs {
	case 1:
		h.reply("DCMD", onBoard(commandFields("c-1", "P1", "set_speed", map[string]any{"pct": 80, "lease_ms": 90000}, sliceEnd(1))))
		h.reply("DQUERY", map[string]any{"device": "P2"})
	case 2:
		h.reply("NCMD", map[string]any{"command": "rebirth"})
		h.reply("DCMD", onBoard(commandFields("c-2", "P2", "safe_stop", map[string]any{}, sliceEnd(2))))
	}
}

func pumpSession(t *testing.T, host *scriptedHost) *run.Run {
	t.Helper()
	r := runOf(t, testsupport.PumpBoardDomain())
	options := sliceOptions(5)
	options.Policies, options.SafeStop = []string{policy}, []string{"safe_stop"}
	if err := NewSession(r, host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestReplayingTheHostRecordsReproducesTheWireAndTheWorld(t *testing.T) {
	t.Parallel()
	live := &scriptedHost{t: t, afterSync: liveScript}
	liveRun := pumpSession(t, live)
	recorded := live.bySync
	replayed := &scriptedHost{t: t, afterSync: func(h *scriptedHost, syncs int) {
		for _, record := range recorded[syncs] {
			h.reply(record["type"].(string), fieldsOf(record))
		}
	}}
	pumpSession(t, replayed)
	if !bytes.Equal(live.transcript, replayed.transcript) || countOf(live.types(), "DRESULT") != 2 {
		t.Fatalf("the replayed wire differs (%d and %d bytes; records %v)", len(live.transcript), len(replayed.transcript), live.types())
	}
	artifact, err := liveRun.End("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := run.ReplayArtifact(context.Background(), artifact, nil, nil, "")
	if err != nil || !result.Matches {
		t.Fatalf("the run with wire commands does not replay: %+v, %v", result, err)
	}
}

func TestWireFaultsActOnSequencedRecords(t *testing.T) {
	t.Parallel()
	host := &scriptedHost{t: t, lenient: true}
	options := sliceOptions(2)
	options.WireFaults = []wire.WireFault{{Seq: 3, Kind: wire.FaultDrop}, {Seq: 4, Kind: wire.FaultDuplicate}, {Seq: 5, Kind: wire.FaultSwap}, {Seq: 8, Kind: wire.FaultGarble}}
	if err := NewSession(runOf(t, testsupport.PumpBoardDomain()), host, options).Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	var seqs []uint64
	for _, record := range host.sent[:8] {
		seqs = append(seqs, wire.Seq(record["seq"]))
	}
	want := []uint64{0, 1, 2, 4, 4, 6, 5, 7}
	for i := range want {
		if seqs[i] != want[i] {
			t.Fatalf("sequence on the wire %v, want %v", seqs, want)
		}
	}
	if garbled := host.sent[8]; len(garbled) != 2 {
		t.Fatalf("record 8 = %v, want a garbled record", garbled)
	}
}

func onBoard(fields map[string]any) map[string]any {
	return withKey(set(fields, "expected_boot_id", "board-01-boot-1"))
}
