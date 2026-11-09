package contract

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const feederPolicy = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func limpFeedScript() []plantwire.Scripted {
	return []plantwire.Scripted{{ID: "limp", AtNS: model.DefaultStartTimeNS + 10e9, Device: "F-1", Command: "set_feed",
		Args: map[string]any{"feed_setpoint": 18}, Verify: &plantwire.Verify{Metric: "feed_setpoint", Equals: 18, Tolerance: 0.1, WithinNS: 6e9}}}
}

func feederSession(t *testing.T) (plantwire.HostReport, string) {
	t.Helper()
	socket, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "session")
	hosted, listening := make(chan error, 1), make(chan struct{})
	var report plantwire.HostReport
	go func() {
		var err error
		report, err = plantwire.Host(t.Context(), socket, 0, plantwire.HostOptions{Environment: "simulated", ProtocolMax: 1, Epoch: 1,
			Policy: feederPolicy, Script: limpFeedScript(), Ready: func() { close(listening) }})
		hosted <- err
	}()
	select {
	case <-listening:
	case err := <-hosted:
		t.Fatalf("the host ended before it listened: %v", err)
	}
	plan := `{"schema_version":"0.1","domain":"belt-conveyor-drives-b","seed":7,"host_socket":"` + socket +
		`","horizon_s":30,"policies":["` + feederPolicy + `"]}`
	result, err := plantwire.Connect(t.Context(), plantwire.ConnectOptions{Plan: strings.NewReader(plan), Source: "plan.json",
		DomainsDir: filepath.Join(testsupport.RepositoryRoot(), "examples"), AdaptersDir: testsupport.AdaptersDir(), Out: out,
		KindsDirs: []string{testsupport.BeltConveyorKindsDir()}, SliceNS: 2e9, BarrierCap: 30 * time.Second})
	if err != nil || result.Stopped || <-hosted != nil {
		t.Fatalf("session = %+v, %v", result, err)
	}
	return report, filepath.Join(out, "journal.jsonl")
}

func TestTheFeederIsBornThroughItsKindFileAndASetFeedIsVerifiedByItsReadBack(t *testing.T) {
	t.Parallel()
	report, journal := feederSession(t)
	status := report.Commands["limp"]
	if len(report.Quarantined) != 0 || status == nil || status.Result != "succeeded" || status.Verified == nil || !*status.Verified {
		t.Fatalf("host report %+v, limp command %+v: want set_feed 18 %% succeeded and read back, nothing quarantined", report, status)
	}
	capture, err := plantwire.CaptureFile(journal, []string{testsupport.BeltConveyorKindsDir()})
	if err != nil || len(capture.Kinds) != 3 {
		t.Fatalf("capture = %+v, %v, want the feeder, the belt and the drive", capture, err)
	}
	for _, kind := range capture.Kinds {
		if kind.Comparison == nil || kind.Comparison.Status != "same" {
			t.Errorf("kind %s is born as %+v, want it equal to its committed kind file", kind.Kind, kind.Comparison)
		}
		if kind.Kind == "bulk_feeder" && !writableSetpoint(kind.Draft) {
			t.Errorf("the feeder is born with %v, want feed and a writable feed_setpoint", kind.Draft["metrics"])
		}
	}
}

func writableSetpoint(draft map[string]any) bool {
	metrics, _ := draft["metrics"].([]any)
	writable := map[string]bool{}
	for _, entry := range metrics {
		metric, _ := entry.(map[string]any)
		name, _ := metric["name"].(string)
		writable[name] = metric["writable"] == true
	}
	return len(writable) == 2 && writable["feed_setpoint"] && !writable["feed"]
}
