package app

import "testing"

func TestConnectedSessionClassificationAndPlanSurviveBridgeMemoryLoss(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/recorded/journal.jsonl",
		"{\"type\":\"run.opened\",\"mode\":\"evaluate\",\"domain\":{\"id\":\"arduino-thermal\"}}\n"+
			"{\"type\":\"session\",\"mode\":\"connected\",\"plan\":"+sessionCheck+"}\n"+
			"{\"type\":\"run.closed\",\"at_ns\":20}\n")
	listing, err := h.bridge.Workspace()
	if err != nil || len(listing.Runs) != 1 {
		t.Fatalf("recorded session listing = %+v, %v", listing, err)
	}
	entry := listing.Runs[0]
	if !entry.Connected || entry.Running || string(entry.Plan) != sessionCheck || entry.Plant != "arduino-thermal" {
		t.Fatalf("persisted session = %+v", entry)
	}
}
