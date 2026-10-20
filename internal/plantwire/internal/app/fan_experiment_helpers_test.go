package app

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
)

func rehearsalBoard(t *testing.T) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: "bench-1", Board: "mega-01",
		Kinds: []map[string]any{thermalKindFile(t, "thermal_sensor.kind.json"), thermalKindFile(t, "indicator.kind.json"),
			thermalKindFile(t, "fan.kind.json")},
		Devices: []standin.Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "61", Initial: map[string]any{"temperature": 30.0}},
			{Name: "led-01", Kind: "indicator", BootID: "61"}, {Name: "fan-01", Kind: "fan", BootID: "61"}},
		Script: standin.Script{Replies: []standin.Reply{{Command: "show_state", Receipt: "accepted", Result: "succeeded"},
			{Command: "set_cooling", Receipt: "accepted", Result: "succeeded"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func latestReadings(snapshot map[string]any) map[string]any {
	readings := map[string]any{}
	devices, _ := snapshot["devices"].([]any)
	for _, entry := range devices {
		device, _ := entry.(map[string]any)
		listed, _ := device["readings"].([]any)
		for _, item := range listed {
			reading, _ := item.(map[string]any)
			readings[device["device"].(string)+"."+reading["name"].(string)] = reading["value"]
		}
	}
	return readings
}

func awaitLive(t *testing.T, out string, until func(readings map[string]any) bool) {
	t.Helper()
	giveUp := time.After(30 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(out, "journal.jsonl")); err == nil {
			if live := journalOfType(t, out, "live"); len(live) > 0 && until(latestReadings(live[len(live)-1])) {
				return
			}
		}
		select {
		case <-giveUp:
			t.Fatalf("no live projection in %s ever showed the expected readings", out)
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func zoneTemperatures(t *testing.T, out string) (first, warmest float64) {
	t.Helper()
	file, err := os.Open(filepath.Join(out, model.FileHistory))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	scanner, seen := bufio.NewScanner(file), false
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var snapshot model.StateSnapshot
		if json.Unmarshal(scanner.Bytes(), &snapshot) != nil || snapshot.Entity != "zone-01" {
			continue
		}
		value, held := snapshot.States["zone_temperature"]
		if held && !seen {
			first, warmest, seen = value, value, true
		}
		warmest = max(warmest, value)
	}
	return first, warmest
}
