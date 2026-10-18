package app

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func (h *wireHost) writeLines(copies int, record domain.Record, seq *uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tx.Queue(record)
	raw, err := json.Marshal(h.tx.Drain()[0])
	if err != nil {
		h.t.Error(err)
		return
	}
	if seq != nil {
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			h.t.Error(err)
			return
		}
		fields["seq"] = *seq
		raw, _ = json.Marshal(fields)
	}
	for range copies {
		if _, err := h.conn.Write(append(raw, '\n')); err != nil {
			h.t.Error(err)
		}
	}
}

func offlineState() domain.Record {
	return domain.Record{Type: "STATE", Lane: domain.LaneControl, Fields: map[string]any{"host_id": "wire-host", "online": false, "heartbeat": 1,
		"heartbeat_interval_ms": 1000, "authority_epoch": 7, "protocol_min": 1, "protocol_max": 1}}
}

type replayCapture struct {
	stats ReplayStats
	host  *wireHost
	err   error
}

func replayAgainst(t *testing.T, folder string, hostile func(*wireHost)) replayCapture {
	t.Helper()
	hostPath := testsupport.SocketPath(t)
	host := serveWireHost(t, hostPath, true)
	done := make(chan replayCapture, 1)
	go func() {
		stats, err := ReplayTrace(t.Context(), replayOptions(folder, hostPath))
		done <- replayCapture{stats: stats, host: host, err: err}
	}()
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NSYNC", "", nil) })
	hostile(host)
	host.release()
	return <-done
}

func receiptsOf(host *wireHost) map[string]map[string]any {
	receipts := map[string]map[string]any{}
	for _, receipt := range ofType(host.snapshot(), "DRECEIPT") {
		receipts[receipt["command_id"].(string)] = receipt
	}
	return receipts
}

func TestAReplayAnswersEveryKindOfHostCommandWithARefusalOrNoneAtAll(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	cases := map[string]struct {
		send    func(*wireHost)
		refused int
		reasons map[string]string
	}{
		"a command the host's schema rejects": {func(h *wireHost) {
			h.write(domain.Record{Type: "DCMD", Lane: domain.LaneControl, Fields: map[string]any{"command_id": "bad-1"}})
		}, 1, map[string]string{"bad-1": "malformed"}},
		"a command whose id is not wire-safe": {func(h *wireHost) {
			h.write(commandRecord("not a wire id", "fan-01", "boot-1", 1))
		}, 0, map[string]string{}},
		"a command the host sent twice": {func(h *wireHost) {
			h.writeLines(2, commandRecord("dup", "fan-01", "boot-1", 1), nil)
		}, 1, map[string]string{"dup": "not_ready"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := replayAgainst(t, folder, tc.send)
			if got.err != nil {
				t.Fatal(got.err)
			}
			got.host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NDEATH", "", nil) })
			receipts := receiptsOf(got.host)
			if got.stats.Refused != tc.refused || len(receipts) != len(tc.reasons) {
				t.Fatalf("stats %+v receipts %v, want %d refusals", got.stats, receipts, tc.refused)
			}
			for id, reason := range tc.reasons {
				if receipts[id]["reason"] != reason || receipts[id]["status"] != "refused" {
					t.Errorf("receipt %q = %v, want a refusal for %s", id, receipts[id], reason)
				}
			}
		})
	}
}

func TestAReplayEndsWithTheHostThatSkippedARecord(t *testing.T) {
	t.Parallel()
	folder := recordedHardwareRun(t, 22.5)
	seq := uint64(9)
	got := replayAgainst(t, folder, func(h *wireHost) { h.writeLines(1, commandRecord("late", "fan-01", "boot-1", 1), &seq) })
	if !errors.Is(got.err, ErrHostGap) {
		t.Fatalf("ReplayTrace = %v, want %v", got.err, ErrHostGap)
	}
}

func TestAReplayStopsPlayingWhenTheHostGoesOffline(t *testing.T) {
	t.Parallel()
	folder := recordedTwoBarrierRun(t)
	got := replayAgainst(t, folder, func(h *wireHost) { h.write(offlineState()) })
	if got.err != nil {
		t.Fatal(got.err)
	}
	got.host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "NDEATH", "", nil) })
	if data := ofType(got.host.snapshot(), "DDATA"); got.stats.Barriers != 1 || len(data) != 1 {
		t.Fatalf("stats %+v with %d DDATA, want the second chunk never sent to an offline host", got.stats, len(data))
	}
}

func TestAReplayWithTwoBarriersPlaysBothChunksInOrder(t *testing.T) {
	t.Parallel()
	stats, played := playTrace(t, recordedTwoBarrierRun(t))
	if data := ofType(played, "DDATA"); stats.Barriers != 2 || len(data) != 2 {
		t.Fatalf("stats %+v with %d DDATA, want both chunks and a barrier after each", stats, len(data))
	}
}

func recordedTwoBarrierRun(t *testing.T) string {
	t.Helper()
	hostPath, out := testsupport.SocketPath(t), t.TempDir()
	started := startHost(t, hostPath, relayHostOptions("hardware"))
	relay := startRelay(t, hostPath, out, nil)
	gateway := dialGateway(t, relay.gateway, relayGateway(t, "s-1"))
	gateway.handshake()
	gateway.send(gateway.node.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(1, 21, "", 1767225600123456789)}))...)
	gateway.sync()
	gateway.send(gateway.node.Emit(domain.DeviceData("zone-01", "boot-1", false, []map[string]any{domain.Sample(2, 22, "", 1767225660123456789)}))...)
	gateway.sync()
	relay.stop()
	started.stop()
	return out
}
