package domain

import "testing"

func TestAConsumingStateFollowsTheSumOverEveryKindConnectedIntoIt(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "N-1.in_load"), link("B-1.out_load", "N-1.in_load"), link("S-1.out_heat", "N-1.in_heat")},
		entryOf("acme/source@1", "source", "S-1"), entryOf("acme/node@1", "node", "N-1"), entryOf("acme/node@1", "node_b", "B-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	follow := member(t, named(t, member(t, doc, "kinds", 1, "dynamics"), "dynamics", "target", "inflow"), "follow")
	want := []string{"node__in_load__source/out", "node__in_load__node_b/outflow", "node__in_heat__source/out"}
	sources := member(t, follow, "sources").([]any)
	if len(sources) != len(want) || member(t, follow, "aggregate") != "sum" {
		t.Fatalf("follow = %v, want the sum over %v", follow, want)
	}
	for i, source := range sources {
		if got := member(t, source, "link").(string) + "/" + member(t, source, "state").(string); got != want[i] {
			t.Errorf("source %d = %s, want %s", i, got, want[i])
		}
	}
}

func TestOneLinkTypeCarriesOneProvidedState(t *testing.T) {
	t.Parallel()
	source := mutate(t, sourceJSON, func(doc map[string]any) {
		kind := member(t, doc, "kind").(map[string]any)
		kind["state"] = append(member(t, kind, "state").([]any), map[string]any{"name": "spare", "unit": "kW", "initial": 5, "min": 0})
		kind["dynamics"] = append(member(t, kind, "dynamics").([]any), map[string]any{"target": "spare", "tier": "F0", "f0": map[string]any{"baseline": 5}})
		doc["ports"] = append(member(t, doc, "ports").([]any), map[string]any{"name": "out_t", "type": "thermal_load", "direction": "provides", "state": "spare"})
	})
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/source.component.json", source), file("equipment/node.component.json", nodeJSON))
	raw := connected(t, []any{link("S-1.out_load", "N-1.in_load"), link("S-1.out_t", "N-1.in_load")},
		entryOf("acme/source@1", "source", "S-1"), entryOf("acme/node@1", "node", "N-1"))
	_, err := compose(t, loadLibrary(t, files.directory("lib")), raw)
	wantIssue(t, err, "composition-v0.1", "/connections/1", "one provided state per link type")
}
