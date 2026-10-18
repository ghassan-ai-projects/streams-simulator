package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertLiveRecording(t *testing.T, out string) {
	t.Helper()
	assertSessionRecordFollowsOpening(t, out)
	metadata := journalOfType(t, out, "session")
	live := journalOfType(t, out, "live")
	if len(metadata) != 1 || metadata[0]["mode"] != "connected" {
		t.Fatalf("persisted connected mode = %v", metadata)
	}
	plan, ok := metadata[0]["plan"].(map[string]any)
	if !ok || plan["domain"] != "arduino-thermal" || len(plan["instances"].([]any)) != 3 {
		t.Fatalf("persisted checked plan = %v", plan)
	}
	if len(live) == 0 {
		t.Fatal("no live projection recorded")
	}
	devices := 0
	for _, snapshot := range live {
		devices = max(devices, len(snapshot["devices"].([]any)))
	}
	if devices != 3 {
		t.Fatalf("live devices = %d, want full plan", devices)
	}
}

func assertSessionRecordFollowsOpening(t *testing.T, out string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, "journal.jsonl"))
	lines := strings.Split(string(raw), "\n")
	if err != nil || len(lines) < 2 || !strings.Contains(lines[0], `"type":"run.opened"`) || !strings.Contains(lines[1], `"type":"session"`) {
		t.Fatalf("observer journal starts %q, %v; want run.opened then the session record", lines[:min(2, len(lines))], err)
	}
}
