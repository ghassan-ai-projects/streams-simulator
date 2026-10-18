package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

type liveNotes struct{ live []map[string]any }

func (n *liveNotes) JournalOrigin(map[string]any)  {}
func (n *liveNotes) JournalBinding(map[string]any) {}
func (n *liveNotes) JournalLive(fields map[string]any) {
	n.live = append(n.live, fields)
}

func hostCommandLine(id string, admitted bool) domain.Traced {
	return domain.Traced{Direction: domain.TraceHost, Admitted: admitted,
		Record: map[string]any{"type": "DCMD", "command_id": id, "device": "P1", "command": "start", "expected_boot_id": "b-1"}}
}

func TestRelayRecordingJournalsOneLiveSnapshotPerTickAndAFinalOneOnClose(t *testing.T) {
	t.Parallel()
	notes := &liveNotes{}
	rec := recording{notes: notes, live: domain.NewLive(true, nil, nil), clock: func() int64 { return 7 }}
	for i := range 3 {
		if err := rec.note(hostCommandLine(fmt.Sprintf("c-%d", i), true)); err != nil {
			t.Fatalf("note: %v", err)
		}
	}
	if len(notes.live) != 0 {
		t.Fatalf("live snapshots per record = %d, want none before the tick", len(notes.live))
	}
	rec.tickLive(250)
	if len(notes.live) != 1 || len(notes.live[0]["commands"].([]*domain.LiveCommand)) != 3 {
		t.Fatalf("snapshots after one tick = %v, want one holding three commands", notes.live)
	}
	if err := rec.note(hostCommandLine("refused-by-relay", false)); err != nil {
		t.Fatalf("note: %v", err)
	}
	rec.tickLive(250)
	if len(notes.live) != 1 {
		t.Fatalf("a quiet tick or an unadmitted host record journalled a snapshot: %d", len(notes.live))
	}
	if err := rec.note(hostCommandLine("c-9", true)); err != nil {
		t.Fatalf("note: %v", err)
	}
	if err := rec.close(); err != nil || len(notes.live) != 2 {
		t.Fatalf("close = %v with %d snapshots, want the pending change journalled", err, len(notes.live))
	}
}

func TestTheRecordedSessionPlanWritesEmptyListsAsArraysForTheJournalSchema(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(sessionNote(ConnectCheck{Domain: "d", Environment: domain.EnvironmentSimulated}))
	if err != nil || !strings.Contains(string(raw), `"instances":[]`) || !strings.Contains(string(raw), `"implements_checked":[]`) {
		t.Fatalf("session note = %s, %v; want empty arrays, not null", raw, err)
	}
}
