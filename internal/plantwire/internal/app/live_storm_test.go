package app

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func connection(side, state string) map[string]any {
	return map[string]any{"side": side, "state": state}
}

func TestAConnectLoopWritesSnapshotsPerTickNotPerConnection(t *testing.T) {
	t.Parallel()
	notes := &liveNotes{}
	rec := recording{notes: notes, live: domain.NewLive(true, nil, nil), clock: func() int64 { return 1 }}
	for range 200 {
		rec.liveConnection(connection("gateway", "online"))
		rec.liveConnection(connection("gateway", "offline"))
		rec.liveConnection(connection("gateway", "offline"))
	}
	if len(notes.live) != 0 {
		t.Fatalf("snapshots before the tick = %d, want none", len(notes.live))
	}
	rec.tickLive(250)
	rec.tickLive(250)
	marks := notes.live[0]["marks"].([]domain.LiveMark)
	if len(notes.live) != 1 || len(marks) != 256 {
		t.Fatalf("snapshots %d with %d marks, want one bounded snapshot", len(notes.live), len(marks))
	}
}

func TestClosingTheRelayRecordingWritesAFinalSnapshotWithLinksOffline(t *testing.T) {
	t.Parallel()
	notes := &liveNotes{}
	rec := recording{notes: notes, live: domain.NewLive(true, nil, nil), clock: func() int64 { return 1 }}
	rec.liveConnection(connection("gateway", "online"))
	rec.liveConnection(connection("host", "online"))
	rec.tickLive(250)
	if err := rec.close(); err != nil || len(notes.live) != 2 {
		t.Fatalf("close = %v with %d snapshots", err, len(notes.live))
	}
	last := notes.live[1]
	if last["gateway"] != "offline" || last["host"] != "offline" {
		t.Fatalf("final snapshot links = %v / %v, want offline", last["gateway"], last["host"])
	}
}
