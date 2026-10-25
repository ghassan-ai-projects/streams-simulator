package acceptance

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const twinDomainID = "arduino-thermal-twin"

func twinJobPlan() map[string]any {
	return map[string]any{"domain": twinDomainID, "seed": 3, "duration_s": 120, "rate": 0}
}

func changeTwinGain(t *testing.T, client *bridgeClient, workspace string) {
	t.Helper()
	path := filepath.Join(workspace, filepath.FromSlash(lineageTwinPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := bytes.Replace(raw, []byte(`"gain": -0.02`), []byte(`"gain": -0.03`), 1)
	if bytes.Equal(edited, raw) {
		t.Fatal("the twin choices have no gain to edit")
	}
	testsupport.WriteFileIn(t, filepath.Dir(path), filepath.Base(path), edited)
	status, current, _ := client.call(http.MethodGet, "/api/file?path="+twinDomainPath, nil)
	etag, _ := current["etag"].(string)
	if status != http.StatusOK || etag == "" {
		t.Fatalf("reading the twin domain: %d %v", status, current)
	}
	client.action("kind.twin", map[string]any{"kinds": []any{"thermal_sensor", "indicator", "fan"}, "choices": lineageTwinPath, "out": twinDomainPath, "etag": etag})
}

func firstChange(t *testing.T, diff map[string]any) (kind, change map[string]any) {
	t.Helper()
	kinds, _ := diff["kinds"].([]any)
	if len(kinds) != 1 {
		t.Fatalf("diff kinds = %v, want only the driven kind", kinds)
	}
	kind, _ = kinds[0].(map[string]any)
	changes, _ := kind["changes"].([]any)
	if len(changes) != 1 {
		t.Fatalf("diff changes = %v, want the one coupling coefficient", changes)
	}
	change, _ = changes[0].(map[string]any)
	return kind, change
}

func TestAChangedThermalGainIsDiffedAgainstTheRunMadeBeforeItAndAgainstTheVariantRun(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs two jobs through the bridge")
	}
	client, _, workspace := pairedJobBridge(t)
	addThermalTwin(t, client, workspace)
	original := completedJob(t, client, twinJobPlan())
	changeTwinGain(t, client, workspace)
	variant := completedJob(t, client, twinJobPlan())
	diff := client.action("domain.diff", map[string]any{"a": original, "b": twinDomainPath})
	kind, change := firstChange(t, diff)
	if nested(diff, "a", "source", "kind") != "run" || nested(diff, "b", "source", "kind") != "file" || nested(diff, "identity", "same_id_and_version_different_digest") != true {
		t.Fatalf("diff = %v, want a recorded run against the edited file under an unchanged id and version", diff["identity"])
	}
	if kind["kind"] != "thermal_sensor" || change["code"] != "gain_changed" || change["before"] != -0.02 || change["after"] != -0.03 || nested(kind, "instances", "after", "count") != 1.0 {
		t.Fatalf("kind = %v, want the thermal sensor zone with the gain from -0.02 to -0.03", kind)
	}
	if nested(diff, "a", "source", "label") != "runs/"+original || nested(diff, "b", "source", "label") != twinDomainPath {
		t.Fatalf("labels = %v %v, want workspace-relative paths", nested(diff, "a", "source"), nested(diff, "b", "source"))
	}
	checkVariantRunsAndLineage(t, client, original, variant, diff)
}

func checkVariantRunsAndLineage(t *testing.T, client *bridgeClient, original, variant string, diff map[string]any) {
	t.Helper()
	runs := client.action("domain.diff", map[string]any{"a": original, "b": variant})
	if _, change := firstChange(t, runs); change["code"] != "gain_changed" || nested(runs, "a", "digest") != nested(diff, "a", "digest") || nested(runs, "b", "digest") != nested(diff, "b", "digest") {
		t.Fatalf("run against run = %v, want the same change between the digests the recordings carry", runs["summary"])
	}
	if again := client.action("domain.diff", map[string]any{"a": original, "b": variant}); encodedDocument(t, again) != encodedDocument(t, runs) {
		t.Fatal("two diffs of the same recordings differ")
	}
	same := client.action("domain.diff", map[string]any{"a": variant, "b": twinDomainPath})
	if nested(same, "summary", "changes") != 0.0 || nested(same, "identity", "same_digest") != true {
		t.Fatalf("variant against the current file = %v, want no change", same["summary"])
	}
	plant := lineagePlant(t, client.action("domain.lineage", map[string]any{"domain": twinDomainID}), twinDomainID)
	recorded, _ := plant["runs"].([]any)
	if len(recorded) != 2 {
		t.Fatalf("lineage runs = %v, want both recordings related to the plant", recorded)
	}
}

func TestDomainDiffRefusesARunThatIsNotThereAndAFileOfAnotherClass(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and talks to the bridge")
	}
	client, _, _ := pairedJobBridge(t)
	for name, tc := range map[string]struct {
		args map[string]any
		code string
	}{
		"a run that does not exist":   {map[string]any{"a": "no-such-run", "b": factoryDomainPath}, "file_not_found"},
		"a file that is not a domain": {map[string]any{"a": lineageTwinPath, "b": factoryDomainPath}, "bad_arguments"},
		"a missing side":              {map[string]any{"a": factoryDomainPath}, "bad_arguments"},
	} {
		status, document, _ := client.call(http.MethodPost, "/api/actions/domain.diff", map[string]any{"args": tc.args})
		if status == http.StatusOK || errorCode(document) != tc.code {
			t.Errorf("%s: %d %v, want %s", name, status, document, tc.code)
		}
	}
}
