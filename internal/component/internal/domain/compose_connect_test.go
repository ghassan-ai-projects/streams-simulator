package domain

import (
	"strings"
	"testing"
)

const sourceJSON = `{
  "schema_version": "0.1", "id": "acme/source", "version": 1, "title": "Source", "class": "process/source/heat",
  "kind": {
    "state": [{"name": "out", "unit": "kW", "initial": 100, "min": 0}, {"name": "t", "unit": "degC", "initial": 20}],
    "dynamics": [{"target": "out", "tier": "F0", "f0": {"baseline": 100, "drift_per_hour": 0, "seasonality": []}},
      {"target": "t", "tier": "F0", "f0": {"baseline": 20, "drift_per_hour": 0, "seasonality": []}}]
  },
  "measurement_points": [{"name": "t", "state": "t", "unit_family": "temperature", "range": {"min": 0, "max": 100}, "tag": "TT", "default_instrument": "acme/probe@1"}],
  "ports": [{"name": "out_load", "type": "thermal_load", "direction": "provides", "state": "out"},
    {"name": "out_heat", "type": "heat_rejection", "direction": "provides", "state": "out"}]
}`

const nodeJSON = `{
  "schema_version": "0.1", "id": "acme/node", "version": 1, "title": "Node", "class": "process/node/heat",
  "kind": {
    "state": [{"name": "inflow", "unit": "kW", "initial": 10, "min": 0}, {"name": "outflow", "unit": "kW", "initial": 10, "min": 0},
      {"name": "bare", "unit": "kW", "initial": 0, "min": 0}, {"name": "t", "unit": "degC", "initial": 20}],
    "dynamics": [{"target": "inflow", "tier": "F0", "f0": {"baseline": 10, "drift_per_hour": 0, "seasonality": []}},
      {"target": "outflow", "tier": "F0", "f0": {"baseline": 10, "drift_per_hour": 0, "seasonality": []}},
      {"target": "t", "tier": "F0", "f0": {"baseline": 20, "drift_per_hour": 0, "seasonality": []}}]
  },
  "measurement_points": [{"name": "t", "state": "t", "unit_family": "temperature", "range": {"min": 0, "max": 100}, "tag": "TT", "default_instrument": "acme/probe@1"}],
  "ports": [{"name": "in_load", "type": "thermal_load", "direction": "consumes", "state": "inflow"},
    {"name": "in_heat", "type": "heat_rejection", "direction": "consumes", "state": "inflow"},
    {"name": "in_bare", "type": "thermal_load", "direction": "consumes", "state": "bare"},
    {"name": "out_load", "type": "thermal_load", "direction": "provides", "state": "outflow"}]
}`

func plantLibrary(t *testing.T) *Library {
	t.Helper()
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/source.component.json", sourceJSON), file("equipment/node.component.json", nodeJSON))
	return loadLibrary(t, files.directory("lib"))
}

func entryOf(use, as string, names ...string) map[string]any {
	list := make([]any, len(names))
	for i, name := range names {
		list[i] = name
	}
	return map[string]any{"use": use, "as": as, "names": list}
}

func connected(t *testing.T, connections []any, entries ...map[string]any) string {
	t.Helper()
	return mutate(t, withEquipment(t, entries...), func(doc map[string]any) { doc["connections"] = connections })
}

func link(from, to string) map[string]any { return map[string]any{"from": from, "to": to} }

func TestConnectionsWriteALinkTypePerPortTypeAndKindPair(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in"), link("S-2.out_load", "K-1.load_in")},
		entryOf("acme/source@1", "source", "S-1", "S-2"), entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	types := member(t, doc, "link_types").([]any)
	if len(types) != 1 || member(t, types[0], "name") != "thermal_load__source__kiln" || member(t, types[0], "from") != "source" || member(t, types[0], "to") != "kiln" {
		t.Fatalf("link types = %v, want the one thermal_load__source__kiln", types)
	}
	links := member(t, doc, "links").([]any)
	if len(links) != 2 || member(t, links[1], "from") != "S-2" || member(t, links[1], "to") != "K-1" || member(t, links[1], "type") != "thermal_load__source__kiln" {
		t.Fatalf("links = %v, want one link per connection in order", links)
	}
}

func TestConnectionsMakeTheConsumingStateFollowTheSumOfTheProvidingStates(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in")}, entryOf("acme/source@1", "source", "S-1"), entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	dynamics := member(t, doc, "kinds", 1, "dynamics")
	follow := member(t, named(t, dynamics, "dynamics", "target", "load"), "follow").(map[string]any)
	if follow["link"] != "thermal_load__source__kiln" || follow["state"] != "out" || follow["aggregate"] != "sum" {
		t.Fatalf("follow = %v, want sum of out over thermal_load__source__kiln", follow)
	}
	if _, present := named(t, member(t, doc, "kinds", 0, "dynamics"), "dynamics", "target", "out")["follow"]; present {
		t.Error("the providing kind follows something")
	}
}

func TestConnectionsResolveTemplateInstancesAndSplitAtTheLastDot(t *testing.T) {
	t.Parallel()
	source := map[string]any{"use": "acme/source@1", "as": "source", "id_template": "line.{n}", "count": map[string]any{"default": 2}}
	raw := connected(t, []any{link("line.2.out_load", "K-1.load_in")}, source, entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	if got := member(t, doc, "links", 0, "from"); got != "line.2" {
		t.Fatalf("link from = %v, want line.2", got)
	}
}

func TestConnectionsRefuseWhatDoesNotJoin(t *testing.T) {
	t.Parallel()
	source, kiln, node := entryOf("acme/source@1", "source", "S-1"), entryOf("acme/kiln@1", "kiln", "K-1"), entryOf("acme/node@1", "node", "N-1")
	cases := []struct {
		name, pointer, message string
		connections            []any
		entries                []map[string]any
	}{
		{"unknown instance", "/connections/0/from", `"S-9" names no instance`, []any{link("S-9.out_load", "K-1.load_in")}, []map[string]any{source, kiln}},
		{"unknown port", "/connections/0/to", `has no port "load_out"`, []any{link("S-1.out_load", "K-1.load_out")}, []map[string]any{source, kiln}},
		{"a consuming port at the from end", "/connections/0/from", "must be a port that provides", []any{link("K-1.load_in", "K-1.load_in")}, []map[string]any{kiln}},
		{"a providing port at the to end", "/connections/0/to", "must be a port that consumes", []any{link("S-1.out_load", "S-1.out_heat")}, []map[string]any{source}},
		{"a type mismatch", "/connections/0", `of type "heat_rejection", port K-1.load_in of type "thermal_load"`, []any{link("K-1.heat_out", "K-1.load_in")}, []map[string]any{kiln}},
		{"two link types into one consuming state", "/connections/1", "one link type and one provided state per consuming state",
			[]any{link("S-1.out_load", "N-1.in_load"), link("S-1.out_heat", "N-1.in_heat")}, []map[string]any{source, node}},
		{"a consuming state without dynamics", "/connections", `kind node has no dynamics for state "bare"`, []any{link("S-1.out_load", "N-1.in_bare")}, []map[string]any{source, node}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, plantLibrary(t), connected(t, tc.connections, tc.entries...))
			wantIssue(t, err, "composition-v0.1", tc.pointer, tc.message)
		})
	}
}

func TestConnectionsRefuseKindsThatReadEachOtherInACycle(t *testing.T) {
	t.Parallel()
	a, b := entryOf("acme/node@1", "node_a", "A-1"), entryOf("acme/node@1", "node_b", "B-1")
	cases := []struct {
		name        string
		connections []any
		entries     []map[string]any
		want        string
	}{
		{"a kind reading itself", []any{link("A-1.out_load", "A-1.in_load")}, []map[string]any{a}, "node_a -> node_a"},
		{"two kinds reading each other", []any{link("A-1.out_load", "B-1.in_load"), link("B-1.out_load", "A-1.in_load")}, []map[string]any{a, b}, "node_a -> node_b -> node_a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, plantLibrary(t), connected(t, tc.connections, tc.entries...))
			wantIssue(t, err, "composition-v0.1", "/connections", "in a cycle: "+tc.want)
		})
	}
}

func TestConnectionsAcceptAChainWithoutACycle(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "N-1.in_load"), link("N-1.out_load", "K-1.load_in")},
		entryOf("acme/source@1", "source", "S-1"), entryOf("acme/node@1", "node", "N-1"), entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	if got := len(member(t, doc, "link_types").([]any)); got != 2 {
		t.Fatalf("link types = %d, want 2", got)
	}
}

func TestConnectionsThatRepeatALinkAreADomainViolation(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in"), link("S-1.out_load", "K-1.load_in")},
		entryOf("acme/source@1", "source", "S-1"), entryOf("acme/kiln@1", "kiln", "K-1"))
	_, err := compose(t, plantLibrary(t), raw)
	invalid := asViolation(t, err, "domain-spec-v0.2")
	if !strings.Contains(invalid.Error(), "link") {
		t.Fatalf("error = %v, want the domain's complaint about the repeated link", err)
	}
}
