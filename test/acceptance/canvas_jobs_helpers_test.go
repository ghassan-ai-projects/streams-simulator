package acceptance

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	canvasEffector = "cooling_tower__set_fan_speed"
	jobPollEvery   = 25 * time.Millisecond
)

func pairedJobBridge(t *testing.T) (*bridgeClient, *canvasBridge, string) {
	t.Helper()
	workspace, uiDir := t.TempDir(), t.TempDir()
	copyIntoWorkspace(t, testsupport.FactoryDomain(), workspace, factoryDomainPath)
	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>canvas ui</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	bridge := startCanvasBridge(t, workspace, uiDir)
	client := newBridgeClient(t, bridge.url)
	pairBrowser(t, client, bridge)
	return client, bridge, workspace
}

func reloadedBrowser(t *testing.T, client *bridgeClient, bridge *canvasBridge) *bridgeClient {
	t.Helper()
	reloaded := newBridgeClient(t, bridge.url)
	base, err := url.Parse(bridge.url)
	if err != nil {
		t.Fatal(err)
	}
	reloaded.http.Jar.SetCookies(base, client.http.Jar.Cookies(base))
	return reloaded
}

func jobPlan(duration int, rate float64) map[string]any {
	return map[string]any{
		"domain": canvasDomain, "seed": 3, "duration_s": duration, "rate": rate,
		"faults": []any{map[string]any{"entity_id": "CT-1", "fault": canvasFault, "at_s": duration / 4}},
		"calls": []any{map[string]any{"entity_id": "CT-1", "effector": canvasEffector, "at_s": duration / 8,
			"args": map[string]any{"tower_id": "CT-1", "speed": 0.8}}},
	}
}

func startJob(t *testing.T, client *bridgeClient, plan map[string]any) string {
	t.Helper()
	started := client.action("job.start", plan)
	world, _ := started["world_id"].(string)
	if world == "" || started["state"] != "running" || started["token"] != nil || started["operator_endpoint"] != nil {
		t.Fatalf("job.start = %v", started)
	}
	return world
}

func jobState(t *testing.T, client *bridgeClient, world string) map[string]any {
	t.Helper()
	return client.action("job.status", map[string]any{"world_id": world})
}

func pollUntil(t *testing.T, what string, done func() bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), bridgeTimeout)
	defer cancel()
	ticker := time.NewTicker(jobPollEvery)
	defer ticker.Stop()
	for !done() {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for %s", what)
		case <-ticker.C:
		}
	}
}

func waitForJobState(t *testing.T, client *bridgeClient, world, state string) map[string]any {
	t.Helper()
	var status map[string]any
	pollUntil(t, "job "+world+" to be "+state, func() bool {
		status = jobState(t, client, world)
		return status["state"] == state
	})
	return status
}

func waitForListing(t *testing.T, client *bridgeClient, world, what string, want func(entry map[string]any) bool) {
	t.Helper()
	pollUntil(t, what, func() bool { return want(listedRun(t, client, world)) })
}

func reportRun(t *testing.T, client *bridgeClient, world string) map[string]any {
	t.Helper()
	report := client.action("run.report", map[string]any{"run": world})
	run, _ := report["run"].(map[string]any)
	if run == nil {
		t.Fatalf("run.report has no run: %v", report)
	}
	return run
}

func listedRunCount(t *testing.T, client *bridgeClient) int {
	t.Helper()
	status, listing, _ := client.call(http.MethodGet, "/api/workspace", nil)
	runs, _ := listing["runs"].([]any)
	if status != http.StatusOK {
		t.Fatalf("workspace listing: %d %v", status, listing)
	}
	return len(runs)
}
