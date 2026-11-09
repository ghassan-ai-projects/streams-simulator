package acceptance

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// droppedChannel is a reading a story drops out: after the fault's injection the binary delivers none of it.
type droppedChannel struct {
	entity, channel, fault string
}

var storyDropouts = map[string]droppedChannel{
	"s5b-a": {"D-3", "conveyor_drive.vibration", "conveyor_drive__vibration__dropout"},
}

func dropoutOnset(t *testing.T, scenarioPath string, dropped droppedChannel) time.Time {
	t.Helper()
	raw, err := os.ReadFile(scenarioPath)
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := model.DecodeScenario(raw)
	if err != nil {
		t.Fatal(err)
	}
	at := slices.IndexFunc(scenario.Commands, func(c model.ScenarioCommand) bool {
		return c.Op == model.OpFaultInject && c.Args["fault"] == dropped.fault && c.Args["entity_id"] == dropped.entity
	})
	if at < 0 {
		t.Fatalf("%s injects no %s on %s", scenarioPath, dropped.fault, dropped.entity)
	}
	onset, err := scenario.Commands[at].Args["onset_ns"].(json.Number).Int64()
	if err != nil {
		t.Fatal(err)
	}
	return time.Unix(0, onset)
}

// deliveredAround counts the dropped channel's delivered readings before and at or after its dropout.
func deliveredAround(t *testing.T, runDir string, dropped droppedChannel, onset time.Time) (before, after int) {
	t.Helper()
	file, err := os.Open(filepath.Join(runDir, model.FileTrace))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	lines := bufio.NewScanner(file)
	for lines.Scan() {
		var reading struct {
			EntityID  string    `json:"entity_id"`
			Channel   string    `json:"channel"`
			EventTime time.Time `json:"event_time"`
		}
		if err := json.Unmarshal(lines.Bytes(), &reading); err != nil {
			t.Fatal(err)
		}
		switch {
		case reading.EntityID != dropped.entity || reading.Channel != dropped.channel:
		case reading.EventTime.Before(onset):
			before++
		default:
			after++
		}
	}
	if err := lines.Err(); err != nil {
		t.Fatal(err)
	}
	return before, after
}

func checkDropoutDelivered(t *testing.T, story, scenarioPath, runDir string) {
	t.Helper()
	dropped, drops := storyDropouts[story]
	if !drops {
		return
	}
	onset := dropoutOnset(t, scenarioPath, dropped)
	before, after := deliveredAround(t, runDir, dropped, onset)
	t.Logf("%s: %s %s delivered %d readings before its dropout at %s and %d after", story, dropped.entity, dropped.channel, before, onset.UTC(), after)
	if before == 0 || after != 0 {
		t.Errorf("%s: %s %s delivered %d readings before its dropout and %d after, want some and none", story, dropped.entity, dropped.channel, before, after)
	}
}
