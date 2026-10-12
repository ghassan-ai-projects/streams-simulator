package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func referenceHost(t *testing.T, path string) <-chan plantwire.HostReport {
	t.Helper()
	ready, reports := make(chan struct{}), make(chan plantwire.HostReport, 1)
	options := plantwire.HostOptions{Environment: "simulated", ProtocolMax: 1, Epoch: 1, Ready: func() { close(ready) }}
	go func() {
		report, _ := plantwire.Host(t.Context(), path, 0, options)
		reports <- report
	}()
	<-ready
	return reports
}

func TestAPlantWireNodeRunsAgainstAHostAndWritesItsRunFolder(t *testing.T) {
	t.Parallel()
	path := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	reports := referenceHost(t, path)
	out := t.TempDir()
	doc := invoke(t, "plantwire", "node", "--host-socket", path, "--domain", "pump-board", "--domains-dir", filepath.Dir(testsupport.PumpBoardDomain()),
		"--adapters-dir", testsupport.AdaptersDir(), "--out", out, "--duration", "180").mustSucceed(t).document(t, "plantwire_node")
	if doc["out"] != out || (<-reports).Refused != "" {
		t.Fatalf("document %v", doc)
	}
	if _, err := os.Stat(filepath.Join(out, "journal.jsonl")); err != nil {
		t.Fatal(err)
	}
}

func TestAPlantWireHostReportsItsSession(t *testing.T) {
	t.Parallel()
	path := filepath.Join(filepath.Dir(testsupport.SocketPath(t)), "h.sock")
	script := writeFile(t, "script.json", `[{"at_ns": 1767225660000000000, "device": "P1", "command": "start", "args": {}}]`)
	inventory := writeFile(t, "inventory.json", `{"P1": "pump"}`)
	hosted := make(chan outcome, 1)
	go func() {
		hosted <- invoke(t, "plantwire", "host", "--socket", path, "--script", script, "--inventory", inventory)
	}()
	waitForSocket(t, path)
	options := plantwire.Options{Environment: "simulated", ProtocolMax: 1, SessionID: "s-1", SliceNS: 60e9, UntilNS: 1767225600000000000 + 180e9}
	if err := plantwire.Serve(t.Context(), pumpBoardRun(t), path, 5*time.Second, options); err != nil {
		t.Fatal(err)
	}
	doc := (<-hosted).mustSucceed(t).document(t, "plantwire_host")
	if quarantined := doc["quarantined"].([]any); len(quarantined) != 1 || quarantined[0] != "P2" {
		t.Fatalf("report %v", doc)
	}
}

func TestPlantWireUsageIsChecked(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"plantwire"}, {"plantwire", "fly"}, {"plantwire", "node"}, {"plantwire", "host"},
		{"plantwire", "host", "--socket", "/tmp/x.sock", "--script", writeFile(t, "bad.json", `{"oops": true}`)}} {
		if code := invoke(t, args...).code; code != 2 {
			t.Errorf("%v exited %d, want 2", args, code)
		}
	}
}

func waitForSocket(t *testing.T, path string) {
	t.Helper()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		select {
		case <-t.Context().Done():
			t.Fatal("the host never listened")
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func pumpBoardRun(t *testing.T) *run.Run {
	t.Helper()
	spec, adapterSpec, err := run.LoadInputs(filepath.Dir(testsupport.PumpBoardDomain()), testsupport.AdaptersDir(), "pump-board", "native-jsonl")
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: 1, SinkName: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
