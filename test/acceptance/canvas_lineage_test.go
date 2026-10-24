package acceptance

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

var twinDomainPath = testsupport.WorkspaceDomain("arduino-thermal-twin")

const (
	lineageCompositionPath = "compositions/factory-utilities.composition.json"
	lineageTwinPath        = "twins/thermal-board.twin.json"
)

func addThermalTwin(t *testing.T, client *bridgeClient, workspace string) {
	t.Helper()
	copyIntoWorkspace(t, testsupport.TwinThermalChoices(), workspace, lineageTwinPath)
	for _, stem := range []string{"fan", "indicator", "thermal_sensor"} {
		copyIntoWorkspace(t, filepath.Join(testsupport.TwinThermalKindsDir(), stem+".kind.json"), workspace, "kinds/"+stem+".kind.json")
	}
	client.action("kind.twin", map[string]any{"kinds": []any{"thermal_sensor", "indicator", "fan"}, "choices": lineageTwinPath, "out": twinDomainPath, "etag": ""})
}

func lineagePlant(t *testing.T, lineage map[string]any, id string) map[string]any {
	t.Helper()
	plants, _ := lineage["plants"].([]any)
	for _, raw := range plants {
		if plant, _ := raw.(map[string]any); plant["domain"].(map[string]any)["id"] == id {
			return plant
		}
	}
	t.Fatalf("lineage %v has no plant %s", lineage, id)
	return nil
}

func nested(document map[string]any, keys ...string) any {
	var current any = document
	for _, key := range keys {
		next, _ := current.(map[string]any)
		current = next[key]
	}
	return current
}

func TestTheBridgeSaysWhereEachPlantComesFromAndNoticesWhenItsCompositionNoLongerMatches(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs one job through the bridge")
	}
	client, _, workspace := pairedJobBridge(t)
	copyIntoWorkspace(t, testsupport.FactoryComposition(), workspace, lineageCompositionPath)
	addThermalTwin(t, client, workspace)
	first := client.action("domain.lineage", map[string]any{})
	factory, twin := lineagePlant(t, first, canvasDomain), lineagePlant(t, first, "arduino-thermal-twin")
	if nested(factory, "composition", "status") != "in_sync" || nested(factory, "editable", "reason") != "composition" || nested(factory, "editable", "route", "code") != "edit_composition" {
		t.Fatalf("factory = %v, want its composition in sync and the edit route", factory)
	}
	if nested(twin, "twin", "status") != "in_sync" || nested(twin, "editable", "route", "code") != "open_device_behavior" || len(twin["device_kinds"].([]any)) != 3 {
		t.Fatalf("twin = %v, want the twin choices in sync routed to the device editor with its three kind files", twin)
	}
	kinds := first["device_kinds"].([]any)
	fan, _ := kinds[0].(map[string]any)
	if fan["kind"] != "fan" || fan["has_own_twin"] != false || len(fan["twins_including"].([]any)) != 1 {
		t.Fatalf("fan = %v, want it included in the three-kind twin and without a twin of its own", fan)
	}
	if again := client.action("domain.lineage", map[string]any{}); encodedDocument(t, again) != encodedDocument(t, first) {
		t.Fatal("two lineages of the same workspace differ")
	}
	editComposition(t, workspace)
	after := lineagePlant(t, client.action("domain.lineage", map[string]any{"domain": canvasDomain}), canvasDomain)
	if nested(after, "composition", "status") != "out_of_sync" || nested(after, "editable", "reason") != "composition_out_of_sync" || nested(after, "editable", "route", "code") != "recompose" {
		t.Fatalf("after editing the composition: %v, want it out of sync with the recompose route", after)
	}
}

func editComposition(t *testing.T, workspace string) {
	t.Helper()
	path := filepath.Join(workspace, filepath.FromSlash(lineageCompositionPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := bytes.Replace(raw, []byte(`"version": "0.1.0"`), []byte(`"version": "0.2.0"`), 1)
	if bytes.Equal(edited, raw) {
		t.Fatal("the composition has no version to edit")
	}
	testsupport.WriteFileIn(t, filepath.Dir(path), filepath.Base(path), edited)
}

func TestTheBridgeFlagsARunThatUsedAnOlderVersionOfAPlantAndStatesItsAssumptions(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds the binary and runs one job through the bridge")
	}
	client, _, workspace := pairedJobBridge(t)
	addThermalTwin(t, client, workspace)
	world := completedJob(t, client, jobPlan(120, 0))
	plant := lineagePlant(t, client.action("domain.lineage", map[string]any{"domain": canvasDomain}), canvasDomain)
	runs, _ := plant["runs"].([]any)
	run, _ := runs[0].(map[string]any)
	if len(runs) != 1 || run["id"] != world || run["same_as_current"] != true || run["recorded"] != true {
		t.Fatalf("runs = %v, want the job's run recorded against the current domain", runs)
	}
	changeDomainFile(t, workspace)
	plant = lineagePlant(t, client.action("domain.lineage", map[string]any{"domain": canvasDomain}), canvasDomain)
	run, _ = plant["runs"].([]any)[0].(map[string]any)
	if run["same_as_current"] != false {
		t.Fatalf("run = %v, want it flagged as made with another version of the plant", run)
	}
	checkAssumptionsOverTheBridge(t, client)
}

func changeDomainFile(t *testing.T, workspace string) {
	t.Helper()
	path := filepath.Join(workspace, filepath.FromSlash(factoryDomainPath))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := bytes.Replace(raw, []byte(`"version": "0.1.0"`), []byte(`"version": "0.3.0"`), 1)
	testsupport.WriteFileIn(t, filepath.Dir(path), filepath.Base(path), edited)
}

func checkAssumptionsOverTheBridge(t *testing.T, client *bridgeClient) {
	t.Helper()
	twin := client.action("domain.assumptions", map[string]any{"path": twinDomainPath, "twin": lineageTwinPath})
	if nested(twin, "origin", "kind") != "twin" || nested(twin, "origin", "twin", "status") != "in_sync" || len(twin["not_modeled"].([]any)) < 5 {
		t.Fatalf("twin assumptions = %v, want a twin in sync and the limits of the model", twin["origin"])
	}
	factory := client.action("domain.assumptions", map[string]any{"path": factoryDomainPath})
	if nested(factory, "origin", "kind") != "composition" {
		t.Fatalf("factory assumptions origin = %v, want composition", factory["origin"])
	}
}
