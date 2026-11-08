package app

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func awaitEnd(t *testing.T, done <-chan connected) connected {
	t.Helper()
	select {
	case ended := <-done:
		return ended
	case <-time.After(time.Minute):
		t.Fatal("the session did not end itself at its horizon")
		return connected{}
	}
}

func TestASessionWithAHorizonEndsItselfAtTheLastSliceWithinItAndReplays(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	plan := strings.TrimSuffix(planText(hostPath, ""), "}") + `,"horizon_s":930` + heatLoad
	options := connectOptions(t, plan, out)
	cancel, done := startConnect(t, options, false)
	defer cancel()
	ended := awaitEnd(t, done)
	if ended.err != nil || ended.result.Stopped || ended.result.Out != out {
		t.Fatalf("Connect = %+v, want a session that ended itself, not stopped", ended)
	}
	times := syncTimes(host.snapshot())
	if len(times) != 15 || slices.Max(times) != model.DefaultStartTimeNS+900e9 {
		t.Fatalf("barriers at %v, want fifteen one-minute slices up to the last slice end within 930 s", times)
	}
	if !hasRecord(host.snapshot(), "NDEATH", "", "") {
		t.Fatal("the node did not send its NDEATH at the horizon")
	}
	if !fileHolds(t, out, "label.json", "zone_heat_load") {
		t.Fatal("a session that ended itself did not write the label of its one disturbance")
	}
	replaysWhole(t, out, options.DomainsDir)
}

func TestAPacedSessionCountsItsHorizonFromTheWorldStart(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	options := connectOptions(t, strings.TrimSuffix(planText(hostPath, ""), "}")+`,"pace":"wall","horizon_s":0.5}`, out)
	options.SliceNS = int64(100 * time.Millisecond)
	cancel, done := startConnect(t, options, false)
	defer cancel()
	if ended := awaitEnd(t, done); ended.err != nil || ended.result.Stopped {
		t.Fatalf("Connect = %+v, want a paced session that ended itself", ended)
	}
	start := recordedStartNS(t, out)
	if times := syncTimes(host.snapshot()); len(times) == 0 || slices.Max(times) != start+int64(500*time.Millisecond) {
		t.Fatalf("barriers at %v, want the last one half a second after the recorded world start %d", times, start)
	}
}

func TestACheckRefusesAHorizonThatCutsADisturbanceShort(t *testing.T) {
	t.Parallel()
	plan := strings.TrimSuffix(planText("/tmp/h.sock", ""), "}") + `,"horizon_s":700` + heatLoad
	_, err := CheckConnect(connectOptions(t, plan, ""))
	invalid, ok := model.AsValidationError(err)
	if !errors.Is(err, ErrPlanRefused) || !ok || invalid.Issues[0].Pointer != "/disturbances/0/duration_s" ||
		invalid.Issues[0].Keyword != domain.KeywordClearAfterHorizon {
		t.Fatalf("a heat load cleared at 720 s under a 700 s horizon: %v, want clear_after_horizon", err)
	}
}
