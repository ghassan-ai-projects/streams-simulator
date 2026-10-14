package acceptance

// The canvas bridge through the real binary: pair a browser, edit and compose
// a workspace, drive a world, follow its live journal, see the truth gate hold
// while the run is open and open when it ends, and never hear the operator
// capability token.

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	canvasDomain = "factory-utilities"
	canvasFault  = "cooling_tower__fan_failure"
)

var factoryDomainPath = testsupport.WorkspaceDomain(canvasDomain)

var operatorToken = regexp.MustCompile(`\bt-[A-Za-z0-9_-]{43}\b`)

func copyIntoWorkspace(t *testing.T, source, workspace, rel string) {
	t.Helper()
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(workspace, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // G703: the target is a fixed relative path inside the test's own temporary workspace
	if err := os.WriteFile(target, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTheCanvasBridgeDrivesAWorldFromPairingToTheReport(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs the factory for a simulated hour through the bridge")
	}
	workspace, uiDir := t.TempDir(), t.TempDir()
	copyIntoWorkspace(t, testsupport.FactoryComposition(), workspace, "compositions/factory-utilities.composition.json")
	copyIntoWorkspace(t, testsupport.FactoryDomain(), workspace, factoryDomainPath)
	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>canvas ui</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	bridge := startCanvasBridge(t, workspace, uiDir)
	client := newBridgeClient(t, bridge.url)

	pairBrowser(t, client, bridge)
	composeAndValidate(t, client, workspace)
	world := createWorld(t, client)
	stream := client.follow("/api/runs/"+world+"/events", "")
	driveWorld(t, client, world, stream)
	checkSealedWhileOpen(t, client, world)
	endRun(t, client, world, stream)
	checkEvidence(t, client, world)
	if token := operatorToken.FindString(client.heard()); token != "" || strings.Contains(client.heard(), "operator_endpoint") {
		t.Fatalf("the bridge let the operator capability through: %q", token)
	}
	checkSecretsStayOut(t, client, bridge)
}

func checkSecretsStayOut(t *testing.T, client *bridgeClient, bridge *canvasBridge) {
	t.Helper()
	heard := client.heard()
	if strings.Contains(heard, bridge.code) || strings.Contains(strings.ToUpper(heard), strings.ReplaceAll(bridge.code, "-", "")) {
		t.Error("the pairing code appeared in a response")
	}
	base, _ := url.Parse(bridge.url)
	cookies := client.http.Jar.Cookies(base)
	if len(cookies) != 1 || strings.Contains(heard, cookies[0].Value) || strings.Contains(bridge.stderr.String(), cookies[0].Value) {
		t.Errorf("the session id appeared in a response or a log line (%d cookies)", len(cookies))
	}
}

func pairBrowser(t *testing.T, client *bridgeClient, bridge *canvasBridge) {
	t.Helper()
	if status, document, _ := client.call(http.MethodGet, "/api/workspace", nil); status != http.StatusUnauthorized || errorCode(document) != "unpaired" {
		t.Fatalf("an unpaired workspace listing: %d %v", status, document)
	}
	if status, document, _ := client.call(http.MethodPost, "/api/session", map[string]string{"code": "AAAA-AAAA-AAAA"}); status != http.StatusUnauthorized || errorCode(document) != "wrong_code" {
		t.Fatalf("a wrong code: %d %v", status, document)
	}
	status, document, header := client.call(http.MethodPost, "/api/session", map[string]string{"code": strings.ToLower(bridge.code)})
	cookie := header.Get("Set-Cookie")
	if status != http.StatusOK || document["paired"] != true || !strings.Contains(cookie, "HttpOnly") || !strings.Contains(cookie, "SameSite=Strict") {
		t.Fatalf("pairing: %d %v, cookie %q", status, document, cookie)
	}
	if status, _, _ := client.call(http.MethodPost, "/api/session", map[string]string{"code": bridge.code}); status != http.StatusUnauthorized {
		t.Fatalf("the pairing code paired a second time: %d", status)
	}
	if page := client.page("/"); !strings.Contains(page, "canvas ui") {
		t.Fatalf("the built UI was not served: %q", page)
	}
}

func composeAndValidate(t *testing.T, client *bridgeClient, workspace string) {
	t.Helper()
	_, listing, _ := client.call(http.MethodGet, "/api/workspace", nil)
	domains, _ := listing["domains"].([]any)
	compositions, _ := listing["compositions"].([]any)
	if len(domains) != 1 || len(compositions) != 1 || domains[0].(map[string]any)["id"] != canvasDomain {
		t.Fatalf("workspace listing = %v", listing)
	}
	args := map[string]any{"composition": "compositions/factory-utilities.composition.json", "out": factoryDomainPath}
	status, refused, _ := client.call(http.MethodPost, "/api/actions/domain.compose", map[string]any{"args": args})
	etag, _ := refused["etag"].(string)
	if status != http.StatusConflict || errorCode(refused) != "conflict" || etag == "" {
		t.Fatalf("composing over an existing domain without its version: %d %v", status, refused)
	}
	args["etag"] = etag
	composed := client.action("domain.compose", args)
	if composed["id"] != canvasDomain {
		t.Fatalf("compose = %v", composed)
	}
	if valid := client.action("domain.validate", map[string]any{"path": factoryDomainPath}); valid["valid"] != true {
		t.Fatalf("validate = %v", valid)
	}
	want, _ := os.ReadFile(testsupport.FactoryDomain())
	got, _ := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(factoryDomainPath)))
	if string(got) != string(want) {
		t.Error("composing the factory composition no longer writes the shipped factory domain")
	}
}

func createWorld(t *testing.T, client *bridgeClient) string {
	t.Helper()
	created := client.action("world.create", map[string]any{"domain": canvasDomain, "seed": 3, "adapter": "native-jsonl", "time_mode": "stepped"})
	world, _ := created["world_id"].(string)
	if world == "" || created["token"] != nil || created["operator_endpoint"] != nil {
		t.Fatalf("world.create = %v", created)
	}
	return world
}

func driveWorld(t *testing.T, client *bridgeClient, world string, stream *eventStream) {
	t.Helper()
	client.action("clock.advance", map[string]any{"world_id": world, "by_ns": 120 * 1e9})
	injected := client.action("fault.inject", map[string]any{"world_id": world, "entity_id": "CT-1", "fault": canvasFault})
	if injected["fault_id"] == nil {
		t.Fatalf("fault.inject = %v", injected)
	}
	checkEffectorCall(t, client, world)
	client.action("clock.advance", map[string]any{"world_id": world, "by_ns": 60 * 1e9})
	observed := 0
	read := stream.until(func(event serverEvent) bool {
		if event.name != "record" {
			t.Fatalf("unexpected event %+v", event)
		}
		if strings.Contains(event.data, canvasFault) {
			t.Fatalf("the live stream carries sealed truth: %s", event.data)
		}
		observed++
		return observed >= 3
	})
	if first := read[0]; first.recordType() != "run.opened" || first.id != "0" {
		t.Fatalf("the stream opens with %+v", first)
	}
}

func checkEffectorCall(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	listed := client.action("effector.list", map[string]any{"world_id": world})
	if !strings.Contains(client.heard(), "cooling_tower__set_fan_speed") || listed["effectors"] == nil {
		t.Fatalf("effector.list = %v", listed)
	}
	ack := client.action("effector.invoke", map[string]any{"world_id": world, "entity_id": "CT-1", "effector": "cooling_tower__set_fan_speed", "args": map[string]any{"tower_id": "CT-1", "speed": 0.8}})
	if ack["accepted"] != true || ack["command_id"] == "" || ack["token"] != nil {
		t.Fatalf("effector.invoke = %v", ack)
	}
}

func checkSealedWhileOpen(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	if status, document, _ := client.call(http.MethodGet, "/api/runs/"+world+"%2Fsealed/events", nil); status != http.StatusBadRequest || errorCode(document) != "bad_path" {
		t.Errorf("following the sealed journal through an encoded run id: %d %v", status, document)
	}
	if status, document, _ := client.call(http.MethodGet, "/api/runs/w-99/events", nil); status != http.StatusNotFound || errorCode(document) != "not_found" {
		t.Errorf("following a run that does not exist: %d %v", status, document)
	}
	for _, name := range []string{"sealed/journal.jsonl", "run.json", "ledger.jsonl"} {
		status, document, _ := client.call(http.MethodGet, "/api/file?path=runs/"+world+"/"+name, nil)
		if status != http.StatusForbidden || errorCode(document) != "sealed" {
			t.Errorf("%s while the run is open: %d %v", name, status, document)
		}
	}
	if status, document, _ := client.call(http.MethodGet, "/api/file?path=runs/"+world+"/journal.jsonl", nil); status != http.StatusOK || document["etag"] == "" {
		t.Errorf("the observer journal while open: %d", status)
	}
	status, document, _ := client.call(http.MethodPut, "/api/file", map[string]string{"path": "runs/" + world + "/run.json", "content": "{}", "etag": ""})
	if status != http.StatusForbidden || errorCode(document) != "read_only" {
		t.Errorf("writing a run file: %d %v", status, document)
	}
}

func endRun(t *testing.T, client *bridgeClient, world string, stream *eventStream) {
	t.Helper()
	client.action("run.end", map[string]any{"world_id": world})
	rest := stream.drain()
	if len(rest) == 0 || rest[len(rest)-1].name != "closed" {
		t.Fatalf("the stream did not end with a closed event: %+v", rest)
	}
	replay := client.follow("/api/runs/"+world+"/events", "2")
	events := replay.drain()
	if len(events) < 2 || events[0].id != "3" || events[len(events)-1].name != "closed" {
		t.Fatalf("resuming after record 2 gave %+v", events)
	}
	last := events[len(events)-2].id
	if again := client.follow("/api/runs/"+world+"/events", last).drain(); len(again) != 1 || again[0].name != "closed" {
		t.Fatalf("resuming after the closing record %s gave %+v", last, again)
	}
	client.action("world.destroy", map[string]any{"world_id": world})
}

func checkEvidence(t *testing.T, client *bridgeClient, world string) {
	t.Helper()
	status, sealed, _ := client.call(http.MethodGet, "/api/file?path=runs/"+world+"/sealed/journal.jsonl", nil)
	if content, _ := sealed["content"].(string); status != http.StatusOK || !strings.Contains(content, canvasFault) {
		t.Fatalf("the sealed journal after the run ended: %d %v", status, sealed)
	}
	if verified := client.action("run.verify", map[string]any{"run": world}); verified["matches"] != true || verified["version_match"] != true {
		t.Fatalf("verify = %v", verified)
	}
	report := client.action("run.report", map[string]any{"run": world})
	truth, _ := report["truth"].(map[string]any)
	faults, _ := truth["faults"].([]any)
	if len(faults) != 1 || faults[0].(map[string]any)["fault"] != canvasFault {
		t.Fatalf("report truth = %v", truth)
	}
	_, listing, _ := client.call(http.MethodGet, "/api/workspace", nil)
	if runs, _ := listing["runs"].([]any); len(runs) != 1 || runs[0].(map[string]any)["ended"] != true {
		t.Fatalf("runs = %v", listing["runs"])
	}
}
