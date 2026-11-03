package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pacedSlices = 8

func recordedStartNS(t *testing.T, out string) int64 {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		WorldConfig struct {
			StartTimeNS int64 `json:"start_time_ns"`
		} `json:"world_config"`
	}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	return artifact.WorldConfig.StartTimeNS
}

func syncTimes(received []map[string]any) []int64 {
	var times []int64
	for _, record := range received {
		if record["type"] == "NSYNC" {
			at, _ := strconv.ParseInt(fmt.Sprint(record["world_time_ns"]), 10, 64)
			times = append(times, at)
		}
	}
	return times
}

func TestAPacedSessionRunsInWallTimeFromItsRecordedStartAndReplays(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	options := connectOptions(t, strings.TrimSuffix(planText(hostPath, ""), "}")+`,"pace":"wall"}`, out)
	options.SliceNS = int64(100 * time.Millisecond)
	before := time.Now()
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return len(syncTimes(r)) >= pacedSlices })
	elapsed := time.Since(before)
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want a stopped paced session", ended)
	}
	start := recordedStartNS(t, out)
	if start < before.UnixNano()-int64(time.Millisecond) || start > before.Add(elapsed).UnixNano() {
		t.Fatalf("run.json start %d, want the wall time the session started (between %d and %d)", start, before.UnixNano(), before.Add(elapsed).UnixNano())
	}
	if minimum := pacedSlices*time.Duration(options.SliceNS) - time.Millisecond; elapsed < minimum || elapsed > minimum+30*time.Second {
		t.Fatalf("%d slices of %v took %v of wall time", pacedSlices, time.Duration(options.SliceNS), elapsed)
	}
	if times := syncTimes(host.snapshot()); times[0] != start+options.SliceNS {
		t.Errorf("the first barrier closes world time %d, want start + one slice %d", times[0], start+options.SliceNS)
	}
	if !hasClockMode(host.snapshot(), "wall") {
		t.Error("the paced node's NBIRTH does not announce clock_mode wall")
	}
	replaysWhole(t, out, options.DomainsDir)
}

func hasClockMode(received []map[string]any, mode string) bool {
	for _, record := range received {
		if metrics, _ := record["node_metrics"].(map[string]any); record["type"] == "NBIRTH" && metrics["clock_mode"] == mode {
			return true
		}
	}
	return false
}

func TestAnUnpacedSessionKeepsTheFixedStartAndAVirtualClock(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	cancel, done := startConnect(t, connectOptions(t, planText(hostPath, ""), out), false)
	host.waitFor(func(r []map[string]any) bool { return len(syncTimes(r)) >= 2 })
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatal(ended.err)
	}
	if start := recordedStartNS(t, out); start != model.DefaultStartTimeNS || !hasClockMode(host.snapshot(), "virtual") {
		t.Fatalf("an unpaced session starts at %d with clock modes %v, want %d and virtual", start, host.snapshot()[0]["node_metrics"], model.DefaultStartTimeNS)
	}
}

func TestAnUnpacedSessionStartsItsWorldAtThePlansOffset(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-2")
	host := serveWireHost(t, hostPath, false)
	plan := strings.TrimSuffix(planText(hostPath, ""), "}") + `,"start_s":86400}`
	options := connectOptions(t, plan, out)
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return len(syncTimes(r)) >= 2 })
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatal(ended.err)
	}
	want := model.DefaultStartTimeNS + 86400*int64(time.Second)
	if start, times := recordedStartNS(t, out), syncTimes(host.snapshot()); start != want || times[0] != want+options.SliceNS {
		t.Fatalf("the session starts at %d and first syncs %d, want %d and one slice later", start, times[0], want)
	}
}

func TestAPacedMixedSessionAnnouncesTheWallClockAndRecordsItsStart(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-3")
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	host := serveWireHost(t, hostPath, false)
	plan := strings.TrimSuffix(planText(hostPath, gatewayPath, "zone-01", "uno-07", "zone-01"), "}") + `,"pace":"wall"}`
	options := connectOptions(t, plan, out)
	options.SliceNS = int64(50 * time.Millisecond)
	before := time.Now()
	cancel, done := startConnect(t, options, true)
	gateway := dialGateway(t, gatewayPath, benchGateway(t, "zone-01"))
	if state := gateway.handshake(); !state.Online {
		t.Fatalf("STATE = %v, want the gateway admitted against the plan", state.Record)
	}
	host.waitFor(func(r []map[string]any) bool { return len(syncTimes(r)) >= 3 })
	cancel()
	if ended := <-done; ended.err != nil || !ended.result.Stopped {
		t.Fatalf("Connect = %+v, want a stopped paced mixed session", ended)
	}
	if start := recordedStartNS(t, out); start < before.UnixNano()-int64(time.Millisecond) || start > time.Now().UnixNano() {
		t.Fatalf("run.json start %d, want the wall time the relay opened (after %d)", start, before.UnixNano())
	}
	if !hasClockMode(host.snapshot(), "wall") {
		t.Error("the merged NBIRTH of a paced mixed session does not announce clock_mode wall")
	}
	replaysWhole(t, out, options.DomainsDir)
}
