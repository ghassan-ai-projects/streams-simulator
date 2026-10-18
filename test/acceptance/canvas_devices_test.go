package acceptance

// The device definitions of a workspace through the real binary: kind files
// and workspace components are listed, written and validated by actions, and a
// domain that implements a kind is checked against the workspace's own kinds
// folder, failing closed when that folder has no such kind.

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var thermalDomain = testsupport.WorkspaceDomain("thermal-board")

func startDeviceBridge(t *testing.T) (*bridgeClient, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("builds the binary and serves a workspace")
	}
	workspace, uiDir := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(uiDir, "index.html"), []byte("<html>canvas ui</html>"), 0o600); err != nil {
		t.Fatal(err)
	}
	bridge := startCanvasBridge(t, workspace, uiDir)
	client := newBridgeClient(t, bridge.url)
	pairBrowser(t, client, bridge)
	return client, workspace
}

func putFile(t *testing.T, client *bridgeClient, rel, content string) {
	t.Helper()
	status, document, _ := client.call(http.MethodPut, "/api/file", map[string]string{"path": rel, "content": content, "etag": ""})
	if status != http.StatusOK || document["path"] != rel {
		t.Fatalf("writing %s: %d %v", rel, status, document)
	}
}

func listed(t *testing.T, client *bridgeClient, key string) []any {
	t.Helper()
	status, listing, _ := client.call(http.MethodGet, "/api/workspace", nil)
	entries, ok := listing[key].([]any)
	if status != http.StatusOK || !ok {
		t.Fatalf("workspace %s: %d %v", key, status, listing)
	}
	return entries
}

func refusedAction(t *testing.T, client *bridgeClient, name string, args map[string]any) map[string]any {
	t.Helper()
	status, document, _ := client.call(http.MethodPost, "/api/actions/"+name, map[string]any{"args": args})
	failure, _ := document["error"].(map[string]any)
	if status != http.StatusUnprocessableEntity || failure == nil {
		t.Fatalf("action %s %v = %d %v, want a refusal with status 422", name, args, status, document)
	}
	return failure
}

func TestTheCanvasBridgeChecksADomainAgainstTheKindsOfItsWorkspace(t *testing.T) {
	t.Parallel()
	client, workspace := startDeviceBridge(t)
	copyIntoWorkspace(t, testsupport.ThermalBoardDomain(), workspace, thermalDomain)
	if kinds, components := listed(t, client, "kinds"), listed(t, client, "components"); len(kinds) != 0 || len(components) != 0 {
		t.Fatalf("a fresh workspace lists kinds %v and components %v", kinds, components)
	}
	validate := map[string]any{"path": thermalDomain}
	failure := refusedAction(t, client, "domain.validate", validate)
	issues, _ := failure["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	message, _ := issue["message"].(string)
	if failure["code"] != "domain_invalid" || issue["pointer"] != "/kinds/0/implements" || !strings.Contains(message, "no kind file of that name was loaded") {
		t.Fatalf("a domain implementing kinds the workspace does not have: %v", failure)
	}
	putKindFiles(t, client)
	if valid := client.action("domain.validate", validate); valid["valid"] != true || valid["id"] != "arduino-thermal" {
		t.Fatalf("validate with the workspace kinds = %v", valid)
	}
}

func putKindFiles(t *testing.T, client *bridgeClient) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(testsupport.ThermalBoardKindsDir(), "*.kind.json"))
	if err != nil || len(paths) != 3 {
		t.Fatalf("thermal board kind files = %v, %v", paths, err)
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		putFile(t, client, "kinds/"+filepath.Base(path), string(content))
	}
	names := []string{}
	for _, entry := range listed(t, client, "kinds") {
		names = append(names, entry.(map[string]any)["kind"].(string))
	}
	if strings.Join(names, ",") != "fan,indicator,thermal_sensor" {
		t.Fatalf("kinds listed = %v", names)
	}
}

func TestTheCanvasBridgeCreatesANewKindFileWithAHyphenatedNameAndStillRefusesBadOnes(t *testing.T) {
	t.Parallel()
	client, workspace := startDeviceBridge(t)
	const rel = "kinds/e2e-fan.kind.json"
	content := `{"kind":"e2e-fan","metrics":[],"commands":[]}`
	putFile(t, client, rel, content)
	status, read, _ := client.call(http.MethodGet, "/api/file?path="+rel, nil)
	if written, _ := os.ReadFile(filepath.Join(workspace, "kinds", "e2e-fan.kind.json")); status != http.StatusOK || read["content"] != content || string(written) != content {
		t.Fatalf("reading the new kind file: %d %v, on disk %q", status, read, written)
	}
	if valid := client.action("kind.validate", map[string]any{"path": rel}); valid["valid"] != true {
		t.Fatalf("validating the new kind file = %v", valid)
	}
	for _, bad := range []string{"kinds/e2e-fan.json", "kinds/../e2e-fan.kind.json", "../kinds/e2e-fan.kind.json", "kinds/sub/e2e-fan.kind.json"} {
		status, document, _ := client.call(http.MethodPut, "/api/file", map[string]string{"path": bad, "content": content, "etag": ""})
		if status != http.StatusBadRequest || errorCode(document) != "bad_path" {
			t.Errorf("writing %s: %d %v", bad, status, document)
		}
	}
}

func TestTheCanvasBridgeValidatesKindFilesAndFailsClosedOnABadOneOnlyForTheDomainThatImplementsIt(t *testing.T) {
	t.Parallel()
	client, workspace := startDeviceBridge(t)
	putKindFiles(t, client)
	good := client.action("kind.validate", map[string]any{"path": "kinds/fan.kind.json"})
	kinds, _ := good["kinds"].([]any)
	fan, _ := kinds[0].(map[string]any)
	digest, _ := fan["kind_digest"].(string)
	path, _ := fan["path"].(string)
	if good["valid"] != true || len(kinds) != 1 || fan["kind"] != "fan" || digest == "" || fan["metrics"] != float64(1) ||
		fan["commands"] != float64(2) || !strings.HasSuffix(path, "kinds/fan.kind.json") || strings.Contains(path, workspace) {
		t.Fatalf("a good kind file = %v", good)
	}
	putFile(t, client, "kinds/thin.kind.json", `{"kind": "thin"}`)
	failure := refusedAction(t, client, "kind.validate", map[string]any{"path": "kinds/thin.kind.json"})
	message, _ := failure["message"].(string)
	if failure["code"] != "kind_failed" || !strings.Contains(message, "kinds/thin.kind.json") || strings.Contains(message, workspace) {
		t.Fatalf("a bad kind file = %v", failure)
	}
	wantIssues := []any{map[string]any{"pointer": "/commands", "message": `missing required property "commands"`, "keyword": "required", "member": "commands"},
		map[string]any{"pointer": "/metrics", "message": `missing required property "metrics"`, "keyword": "required", "member": "metrics"}}
	if !jsonEqual(failure["issues"], wantIssues) {
		t.Fatalf("the issues of a bad kind file = %v, want %v", failure["issues"], wantIssues)
	}
	copyIntoWorkspace(t, testsupport.ThermalBoardDomain(), workspace, thermalDomain)
	if valid := client.action("domain.validate", map[string]any{"path": thermalDomain}); valid["valid"] != true {
		t.Fatalf("a broken kind file the domain does not implement must not stop domain.validate, got %v", valid)
	}
}

func TestTheCanvasBridgeListsAndValidatesWorkspaceComponents(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	const instrument = "components/instruments/lab-run-status.instrument.json"
	putFile(t, client, instrument, labRunStatus(t))
	entries := listed(t, client, "components")
	want := map[string]any{"path": instrument, "id": "lab/run-status", "title": "Run-status contact (digital input)", "class": "instrument"}
	if len(entries) != 1 || !jsonEqual(entries[0], want) {
		t.Fatalf("components = %v, want %v", entries, want)
	}
	validated := client.action("component.validate", nil)
	if directories, _ := validated["directories"].([]any); validated["valid"] != true || len(directories) != 1 || directories[0] != "components" {
		t.Fatalf("component.validate = %v", validated)
	}
	if shown, _ := json.Marshal(client.action("component.list", nil)); !strings.Contains(string(shown), "lab/run-status") {
		t.Fatalf("component.list does not list the workspace instrument: %s", shown)
	}
	putFile(t, client, "components/instruments/bad.instrument.json", `{"id": "lab/bad"}`)
	failure := refusedAction(t, client, "component.validate", nil)
	if message, _ := failure["message"].(string); failure["code"] != "component_invalid" || !strings.Contains(message, "components/instruments/bad.instrument.json") {
		t.Fatalf("a bad instrument = %v", failure)
	}
}

func labRunStatus(t *testing.T) string {
	t.Helper()
	source := filepath.Join(testsupport.RepositoryRoot(), "internal", "component", "internal", "files", "builtin", "instruments", "run-status.instrument.json")
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Replace(string(content), "streamsim/run-status", "lab/run-status", 1)
}

func jsonEqual(got, want any) bool {
	a, errA := json.Marshal(got)
	b, errB := json.Marshal(want)
	return errA == nil && errB == nil && string(a) == string(b)
}

func TestTheDeviceFoldersRefuseTraversalSymlinksAndWrongNames(t *testing.T) {
	t.Parallel()
	client, workspace := startDeviceBridge(t)
	outside := filepath.Join(t.TempDir(), "secret.kind.json")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workspace, "kinds"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(workspace, "kinds", "leak.kind.json")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"kinds/leak.kind.json", "kinds/fan.json", "kinds/../" + testsupport.WorkspaceDomain("x"), "../kinds/fan.kind.json",
		"components/fan.component.json", "components/equipment/fan.instrument.json", "components/equipment/../../x.component.json"} {
		status, document, _ := client.call(http.MethodGet, "/api/file?path="+rel, nil)
		if status != http.StatusBadRequest || errorCode(document) != "bad_path" {
			t.Errorf("reading %s: %d %v", rel, status, document)
		}
		status, document, _ = client.call(http.MethodPut, "/api/file", map[string]string{"path": rel, "content": "{}", "etag": ""})
		if status != http.StatusBadRequest || errorCode(document) != "bad_path" {
			t.Errorf("writing %s: %d %v", rel, status, document)
		}
	}
	if got, _ := os.ReadFile(outside); string(got) != "secret" {
		t.Fatalf("the file outside the workspace was changed: %q", got)
	}
}

func TestTheWorkspaceStreamTellsWhenAKindOrAComponentFileChanges(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	stream := client.follow("/api/events", "")
	for _, rel := range []string{"kinds/watch.kind.json", "components/equipment/watch.component.json"} {
		awaitWorkspaceEvent(t, client, stream, rel)
	}
}

func awaitWorkspaceEvent(t *testing.T, client *bridgeClient, stream *eventStream, rel string) {
	t.Helper()
	etag := ""
	for attempt := range 5 {
		body := map[string]string{"path": rel, "content": strings.Repeat(" ", attempt) + "{}", "etag": etag}
		status, document, _ := client.call(http.MethodPut, "/api/file", body)
		if status != http.StatusOK {
			t.Fatalf("writing %s: %d %v", rel, status, document)
		}
		etag, _ = document["etag"].(string)
		select {
		case event := <-stream.events:
			if event.name != "workspace" {
				t.Fatalf("an unexpected event %+v after writing %s", event, rel)
			}
			return
		case <-time.After(3 * time.Second):
		}
	}
	t.Fatalf("no workspace event after writing %s", rel)
}
