package testsupport

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestACampusRepeatsEveryHallInstanceAndConnectionAndSharesTheNamedOnes(t *testing.T) {
	t.Parallel()
	composition := []byte(`{"id": "plant", "title": "Plant", "equipment": [{"names": ["WX-1"]}, {"names": ["P-1", "P-2"]}],
		"connections": [{"from": "WX-1.out", "to": "P-1.in"}, {"from": "P-1.out", "to": "P-2.in"}]}`)
	raw, err := Campus(composition, 2, "WX-1")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		ID          string                     `json:"id"`
		Equipment   []struct{ Names []string } `json:"equipment"`
		Connections []map[string]string        `json:"connections"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID != "plant-campus-2" || !slices.Equal(doc.Equipment[0].Names, []string{"WX-1"}) ||
		!slices.Equal(doc.Equipment[1].Names, []string{"h01/P-1", "h01/P-2", "h02/P-1", "h02/P-2"}) {
		t.Fatalf("campus = %s", raw)
	}
	want := []map[string]string{{"from": "WX-1.out", "to": "h01/P-1.in"}, {"from": "WX-1.out", "to": "h02/P-1.in"},
		{"from": "h01/P-1.out", "to": "h01/P-2.in"}, {"from": "h02/P-1.out", "to": "h02/P-2.in"}}
	if len(doc.Connections) != len(want) {
		t.Fatalf("connections = %v, want %v", doc.Connections, want)
	}
	for i := range want {
		if doc.Connections[i]["from"] != want[i]["from"] || doc.Connections[i]["to"] != want[i]["to"] {
			t.Errorf("connection %d = %v, want %v", i, doc.Connections[i], want[i])
		}
	}
}

func TestACampusRefusesAnInvalidComposition(t *testing.T) {
	t.Parallel()
	if _, err := Campus([]byte("{"), 2); err == nil {
		t.Fatal("an invalid composition was accepted")
	}
}

func TestEveryFixturePathIsUnderTheRepository(t *testing.T) {
	t.Parallel()
	root := RepositoryRoot()
	for _, path := range []string{StructuredDomain(), TransferDomain(), UtilityDomain(), TwoKindDomain(), Contract("domain-spec-v0.2"),
		DeviceCapabilities(), DeviceWorldBindings(), DeviceConformance("valid", "x"), PumpBoardDomain(), CoupledDomain(), LoopDomain(),
		ComponentsDir(), PilotComposition(), FactoryComposition()} {
		if len(path) <= len(root) || path[:len(root)] != root {
			t.Errorf("%s is not under %s", path, root)
		}
	}
}
