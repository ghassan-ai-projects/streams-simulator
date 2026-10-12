package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twoKindDocument(t *testing.T, change func(doc map[string]any)) []byte {
	t.Helper()
	return documentOf(t, testsupport.TwoKindDomain(), change)
}

func kindAt(doc map[string]any, index int) map[string]any {
	return doc["kinds"].([]any)[index].(map[string]any)
}

func TestSeveralKindsResolveByDomainWideNames(t *testing.T) {
	t.Parallel()
	c, err := Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	if c.FaultKind("pump_wear").Name != "pump" || c.Fault("tank_leak") == nil || c.Effector("boost") == nil || c.Channel("tank.level") == nil {
		t.Fatal("a domain-wide name did not resolve to its declaration")
	}
	if c.FaultCount() != 4 || c.EffectorCount() != 2 || len(c.ChannelNames()) != 2 || len(c.KindChannelNames("pump")) != 1 {
		t.Fatalf("counts: faults %d effectors %d channels %v", c.FaultCount(), c.EffectorCount(), c.ChannelNames())
	}
	if got := c.StateNames("tank"); len(got) != 2 || got[0] != "inflow" || len(c.StateNames("none")) != 0 {
		t.Fatalf("tank states = %v", got)
	}
	if c.InstanceSetFor("pump-3") != 1 || c.InstanceSetFor("tank-1") != 0 || c.InstanceSetFor("valve-1") != -1 {
		t.Fatal("instance sets were not told apart by id template")
	}
}

func TestEachKindIsCheckedAgainstItsOwnDeclarations(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		change  func(doc map[string]any)
		pointer string
	}{
		{"channel name shared by two kinds", func(doc map[string]any) {
			kindAt(doc, 1)["channels"].([]any)[0].(map[string]any)["name"] = "tank.level"
		}, "/kinds/1/channels/0/name"},
		{"fault id shared by two kinds", func(doc map[string]any) {
			kindAt(doc, 1)["faults"].([]any)[0].(map[string]any)["id"] = "tank_leak"
		}, "/kinds/1/faults/0/id"},
		{"effector name shared by two kinds", func(doc map[string]any) {
			kindAt(doc, 1)["effectors"].([]any)[0].(map[string]any)["name"] = "refill"
		}, "/kinds/1/effectors/0/name"},
		{"kind name declared twice", func(doc map[string]any) {
			kindAt(doc, 1)["name"] = "tank"
			doc["instances"].([]any)[1].(map[string]any)["kind"] = "tank"
		}, "/kinds/1/name"},
		{"detector reads a channel of another kind", func(doc map[string]any) {
			kindAt(doc, 0)["faults"].([]any)[0].(map[string]any)["observability"].(map[string]any)["detector"].(map[string]any)["channel"] = "pump.flow"
		}, "/kinds/0/faults/0/observability/detector/channel"},
		{"fault expects an effector of another kind", func(doc map[string]any) {
			kindAt(doc, 0)["faults"].([]any)[0].(map[string]any)["expected_effector"] = "boost"
		}, "/kinds/0/faults/0/expected_effector"},
		{"channel observes a state of another kind", func(doc map[string]any) {
			kindAt(doc, 1)["channels"].([]any)[0].(map[string]any)["observes"] = "level_true"
		}, "/kinds/1/channels/0/observes"},
		{"kind without an instance set", func(doc map[string]any) { doc["instances"] = doc["instances"].([]any)[:1] }, "/kinds/1/name"},
		{"instance set of an undeclared kind", func(doc map[string]any) { doc["instances"].([]any)[1].(map[string]any)["kind"] = "valve" }, "/instances/1/kind"},
		{"second instance set out of its bounds", func(doc map[string]any) {
			doc["instances"].([]any)[1].(map[string]any)["count"] = map[string]any{"default": 9, "max": 8}
		}, "/instances/1/count/default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(twoKindDocument(t, tc.change), "two-kind.json")
			var invalid *model.ValidationError
			if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != tc.pointer {
				t.Fatalf("err = %v, want one issue at %s", err, tc.pointer)
			}
		})
	}
}

func TestExplicitIdsAreCountedPerInstanceSet(t *testing.T) {
	t.Parallel()
	c, err := Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	if err := c.CheckEntityIDs([]string{"tank-1", "pump-1", "pump-2"}); err != nil {
		t.Fatalf("ids of both sets: %v", err)
	}
	tooManyPumps := []string{"pump-1", "pump-2", "pump-3", "pump-4", "pump-5", "pump-6", "pump-7", "pump-8", "pump-9"}
	if err := c.CheckEntityIDs(tooManyPumps); err == nil || !strings.Contains(err.Error(), "above the instance count maximum 8") {
		t.Fatalf("nine pumps: %v, want a refusal naming the pump set's maximum of eight", err)
	}
	if err := c.CheckEntityIDs([]string{"tank-1", "valve-1"}); err == nil || !strings.Contains(err.Error(), `"valve-1" does not fit instance id template`) {
		t.Fatalf("an id that fits no set: %v, want a refusal naming it", err)
	}
}
