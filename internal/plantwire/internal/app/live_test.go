package app

import "testing"

func assertLiveRecording(t *testing.T, out string) {
	t.Helper()
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
