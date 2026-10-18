package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestPlantWireReplayTraceUsageIsChecked(t *testing.T) {
	t.Parallel()
	requireUsageRefusals(t, [][]string{{"plantwire", "replay-trace"}, {"plantwire", "replay-trace", "--run", "/tmp/r"},
		{"plantwire", "replay-trace", "--host-socket", "/tmp/h.sock"},
		{"plantwire", "replay-trace", "--run", "/tmp/r", "--host-socket", "/tmp/h.sock", "extra"},
		{"plantwire", "replay-trace", "--run", "/tmp/r", "--host-socket", "/tmp/h.sock", "--gateway-socket", "/tmp/g.sock"}})
}

func TestAMixedRelayNeedsARunFolderForItsArtifactAndFieldTrace(t *testing.T) {
	t.Parallel()
	got := invoke(t, "plantwire", "relay", "--gateway-socket", "/tmp/g.sock", "--host-socket", "/tmp/h.sock", "--domain", "pump-board")
	body := got.failure(t)
	if message, _ := body["message"].(string); got.code != 2 || body["code"] != "usage_invalid" || !strings.Contains(message, "--out") {
		t.Fatalf("exit %d with %v, want a usage refusal naming --out", got.code, body)
	}
}

func TestAReplayAndAMixedRelayReportMatchTheCliOutputContract(t *testing.T) {
	t.Parallel()
	digest := "sha256:" + strings.Repeat("a", 64)
	for def, stats := range map[string]any{
		"plantwire_replay": plantwire.ReplayStats{Records: 9, Skipped: 1, Barriers: 2, Refused: 1, FieldTrace: digest, HostSessionID: "replay-1"},
		"plantwire_relay":  plantwire.RelayStats{GatewaySessions: 1, HostSessions: 1, Up: 30, Down: 2, ToSimulated: 12, FromSimulated: 20, UnknownDevice: 1},
	} {
		raw, err := json.Marshal(stats)
		if err != nil {
			t.Fatal(err)
		}
		var doc any
		if err := model.DecodeBytes(raw, &doc); err != nil {
			t.Fatal(err)
		}
		for _, schema := range []string{def, ""} {
			if errs := cliOutputSchema(t, schema).Validate(doc); len(errs) > 0 {
				t.Errorf("%s does not match cli-output %q:\n%s\n%s", def, schema, jsonschema.FormatErrors(errs), raw)
			}
		}
	}
}

func traceOfNode(t *testing.T, out string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, "journal.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	folder := t.TempDir()
	trace, err := run.OpenFieldTrace(folder)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte("\n")) {
		var entry struct {
			Type      string         `json:"type"`
			Direction string         `json:"direction"`
			Record    map[string]any `json:"record"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			t.Fatal(err)
		}
		if entry.Type == "wire" && entry.Direction == "node" {
			if err := trace.Add(entry.Record); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	return folder
}

func TestAPlantWireReplayTracePlaysARecordedTraceToAHost(t *testing.T) {
	t.Parallel()
	recorded := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	reports := referenceHost(t, recorded)
	out := t.TempDir()
	invoke(t, "plantwire", "node", "--host-socket", recorded, "--domain", "pump-board", "--domains-dir", filepath.Dir(testsupport.PumpBoardDomain()),
		"--adapters-dir", testsupport.AdaptersDir(), "--out", out, "--duration", "180").mustSucceed(t)
	if (<-reports).Refused != "" {
		t.Fatal("the host refused the node whose journal is to be replayed")
	}
	folder := traceOfNode(t, out)
	played := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	replayReports := referenceHost(t, played)
	doc := invoke(t, "plantwire", "replay-trace", "--run", folder, "--host-socket", played).mustSucceed(t).document(t, "plantwire_replay")
	if report := <-replayReports; report.Refused != "" || len(report.Quarantined) != 0 {
		t.Fatalf("host report %+v, want the replayed node accepted with nothing quarantined", report)
	}
	if doc["records"].(float64) < 3 || doc["commands_refused"].(float64) != 0 || doc["session_id"] != "replay-1" {
		t.Fatalf("document %v, want the recorded births and data played in the replay's own session", doc)
	}
}
