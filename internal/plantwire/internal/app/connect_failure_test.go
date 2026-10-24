package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func recordedRun(t *testing.T, out string) model.RunArtifact {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	var art model.RunArtifact
	if err := json.Unmarshal(raw, &art); err != nil {
		t.Fatal(err)
	}
	return art
}

func closedIncomplete(t *testing.T, out, file string) bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, file))
	if err != nil {
		t.Fatal(err)
	}
	incomplete, closed := journal.ClosedIncomplete(raw)
	if !closed {
		t.Fatalf("%s did not close with a run_closed record", file)
	}
	return incomplete
}

func assertRecordedAs(t *testing.T, out string, complete bool, reason string) {
	t.Helper()
	art := recordedRun(t, out)
	if art.Incomplete == complete || !strings.Contains(art.Error, reason) {
		t.Fatalf("run.json incomplete=%v error=%q, want complete=%v with error containing %q", art.Incomplete, art.Error, complete, reason)
	}
	for _, file := range []string{model.FileObserverJournal, model.FileDirectorJournal} {
		if closedIncomplete(t, out, file) == complete {
			t.Fatalf("%s closed with incomplete=%v, want %v", file, !complete, !complete)
		}
	}
	got, err := report.Build(t.Context(), out)
	if err != nil {
		t.Fatal(err)
	}
	if got.Run.Complete != complete || !strings.Contains(got.Run.IncompleteReason, reason) || (complete && got.Run.IncompleteReason != "") {
		t.Fatalf("report run complete=%v reason=%q, want complete=%v with reason containing %q", got.Run.Complete, got.Run.IncompleteReason, complete, reason)
	}
}

func TestANodeSessionThatCouldNotReachItsHostIsRecordedIncompleteWithTheFailure(t *testing.T) {
	t.Parallel()
	missing := testsupport.SocketPath(t)
	out := filepath.Join(t.TempDir(), "c-1")
	result, err := Connect(t.Context(), connectOptions(t, planText(missing, ""), out))
	if err == nil || !strings.Contains(err.Error(), "plantwire connect") || result.Stopped {
		t.Fatalf("Connect = %+v, %v, want the failure returned and the session not stopped by the operator", result, err)
	}
	assertRecordedAs(t, out, false, err.Error())
}

func TestAMixedSessionThatNeverReachedItsHostIsRecordedIncompleteWithTheReason(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	cancel, done := startConnect(t, connectOptions(t, planText(hostPath, gatewayPath), out), true)
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want the operator's stop", ended)
	}
	assertRecordedAs(t, out, false, "host")
}

func TestASessionTheOperatorStoppedIsRecordedComplete(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, true)
	cancel, done := startConnect(t, connectOptions(t, planText(hostPath, ""), out), false)
	host.waitFor(func(r []map[string]any) bool { return hasRecord(r, "DBIRTH", "device", "fan-01") })
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want the operator's stop", ended)
	}
	assertRecordedAs(t, out, true, "")
}

func TestAMixedRelayThatEndsWithAFailureAbandonsItsSimulatedRun(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "r-1")
	r, _, err := newSimRun(t.Context(), *mixedOptions(t, out).Simulated, out)
	if err != nil {
		t.Fatal(err)
	}
	node := &simNode{run: r}
	if err := node.end(out, "", "", errors.New("gateway write failed")); err != nil {
		t.Fatal(err)
	}
	assertRecordedAs(t, out, false, "plantwire relay: gateway write failed")
}
