package testsupport

import (
	"encoding/json"
	"testing"
)

func TestAsKindsLiftsAFlatDomainIntoOneKindAndOneInstanceSet(t *testing.T) {
	t.Parallel()
	flat := `{"id":"pump-test","entities":{"id_template":"p-{n}","count":{"default":2},"churn":{"births_per_hour":1}},` +
		`"state":[{"name":"x"}],"channels":[],"faults":[],"profiles":[]}`
	var doc map[string]any
	if err := json.Unmarshal(AsKinds([]byte(flat)), &doc); err != nil {
		t.Fatal(err)
	}
	kind := doc["kinds"].([]any)[0].(map[string]any)
	instances := doc["instances"].([]any)[0].(map[string]any)
	if kind["name"] != "pump" || instances["kind"] != "pump" || instances["id_template"] != "p-{n}" || instances["churn"] == nil ||
		kind["state"] == nil || doc["entities"] != nil || doc["state"] != nil {
		t.Fatalf("lifted document = %v", doc)
	}
}

func TestAsKindsNamesTheKindAfterEntityTypeWhenGiven(t *testing.T) {
	t.Parallel()
	flat := `{"id":"x","entities":{"id_template":"a-{n}","entity_type":"valve","count":{"default":1}}}`
	var doc map[string]any
	if err := json.Unmarshal(AsKinds([]byte(flat)), &doc); err != nil {
		t.Fatal(err)
	}
	if name := doc["kinds"].([]any)[0].(map[string]any)["name"]; name != "valve" {
		t.Fatalf("kind name = %v, want valve", name)
	}
}
