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

func TestConnectionsWriteALinkTypePerConsumingPortAndProvidingKind(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in"), link("S-2.out_load", "K-1.load_in")},
		entryOf("acme/source@1", "source", "S-1", "S-2"), entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	types := member(t, doc, "link_types").([]any)
	if len(types) != 1 || member(t, types[0], "name") != "kiln__load_in__source" || member(t, types[0], "from") != "source" || member(t, types[0], "to") != "kiln" {
		t.Fatalf("link types = %v, want the one kiln__load_in__source", types)
	}
	links := member(t, doc, "links").([]any)
	if len(links) != 2 || member(t, links[1], "from") != "S-2" || member(t, links[1], "to") != "K-1" || member(t, links[1], "type") != "kiln__load_in__source" {
		t.Fatalf("links = %v, want one link per connection in order", links)
	}
}

func TestConnectionsMakeTheConsumingStateFollowTheSumOfTheProvidingStates(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "K-1.load_in")}, entryOf("acme/source@1", "source", "S-1"), entryOf("acme/kiln@1", "kiln", "K-1"))
	doc := composeOK(t, plantLibrary(t), raw)
	dynamics := member(t, doc, "kinds", 1, "dynamics")
	follow := member(t, named(t, dynamics, "dynamics", "target", "load"), "follow").(map[string]any)
	if sources := follow["sources"].([]any); len(sources) != 1 || member(t, sources[0], "link") != "kiln__load_in__source" || member(t, sources[0], "state") != "out" || follow["aggregate"] != "sum" {
		t.Fatalf("follow = %v, want sum of out over kiln__load_in__source", follow)
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

func TestConnectionsRefuseKindsThatReadEachOtherInACycleWithoutASampledRead(t *testing.T) {
	t.Parallel()
	a, b := entryOf("acme/node@1", "node_a", "A-1"), entryOf("acme/node@1", "node_b", "B-1")
	cases := []struct {
		name        string
		connections []any
		entries     []map[string]any
		want        string
	}{
		{"a kind reading itself", []any{link("A-1.out_load", "A-1.in_load")}, []map[string]any{a}, "[node_a node_a]"},
		{"two kinds reading each other", []any{link("A-1.out_load", "B-1.in_load"), link("B-1.out_load", "A-1.in_load")}, []map[string]any{a, b}, "[node_a node_b node_a]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := compose(t, plantLibrary(t), connected(t, tc.connections, tc.entries...))
			invalid := asViolation(t, err, "domain-spec-v0.2")
			for _, part := range []string{"in a cycle", tc.want, "test.composition.json"} {
				if !strings.Contains(err.Error(), part) {
					t.Fatalf("error = %v (%v), want it to name %q", err, invalid.Issues, part)
				}
			}
		})
	}
}

func sampledNodeLibrary(t *testing.T) *Library {
	t.Helper()
	node := mutate(t, nodeJSON, func(doc map[string]any) { member(t, doc, "ports", 0).(map[string]any)["sample_s"] = 60 })
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/source.component.json", sourceJSON), file("equipment/node.component.json", node))
	return loadLibrary(t, files.directory("lib"))
}

func TestAComponentLoopComposesWhenOneConsumingPortIsSampled(t *testing.T) {
	t.Parallel()
	a, b := entryOf("acme/node@1", "node_a", "A-1"), entryOf("acme/node@1", "node_b", "B-1")
	raw := connected(t, []any{link("A-1.out_load", "B-1.in_load"), link("B-1.out_load", "A-1.in_load")}, a, b)
	doc := composeOK(t, sampledNodeLibrary(t), raw)
	for kind := 0; kind < 2; kind++ {
		follow := member(t, named(t, member(t, doc, "kinds", kind, "dynamics"), "dynamics", "target", "inflow"), "follow").(map[string]any)
		if follow["sample_s"] == nil || number(t, follow["sample_s"]) != 60 {
			t.Fatalf("kind %d follow = %v, want sample_s 60 copied from the consuming port", kind, follow)
		}
	}
}

func TestASamplePeriodOnAProvidingPortIsRefused(t *testing.T) {
	t.Parallel()
	node := mutate(t, nodeJSON, func(doc map[string]any) { member(t, doc, "ports", 3).(map[string]any)["sample_s"] = 60 })
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/node.component.json", node))
	_, err := LoadLibrary([]Directory{files.directory("lib")})
	wantIssue(t, err, "component-v0.1", "/ports/3/sample_s", "only a consuming port reads by sample")
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

func twoLoadPortsLibrary(t *testing.T) *Library {
	t.Helper()
	node := mutate(t, nodeJSON, func(doc map[string]any) {
		ports := member(t, doc, "ports").([]any)
		doc["ports"] = append(ports, map[string]any{"name": "in_load_two", "type": "thermal_load", "direction": "consumes", "state": "outflow"})
	})
	files := baseFiles()
	files.equipment = append(files.equipment, file("equipment/source.component.json", sourceJSON), file("equipment/node.component.json", node))
	return loadLibrary(t, files.directory("lib"))
}

func TestTwoPortPairsBetweenTheSameInstancesGetOneLinkTypeEach(t *testing.T) {
	t.Parallel()
	raw := connected(t, []any{link("S-1.out_load", "N-1.in_load"), link("S-1.out_load", "N-1.in_load_two")},
		entryOf("acme/source@1", "source", "S-1"), entryOf("acme/node@1", "node", "N-1"))
	doc := composeOK(t, twoLoadPortsLibrary(t), raw)
	types := member(t, doc, "link_types").([]any)
	if len(types) != 2 || member(t, types[0], "name") != "node__in_load__source" || member(t, types[1], "name") != "node__in_load_two__source" {
		t.Fatalf("link types = %v, want one per consuming port", types)
	}
	links := member(t, doc, "links").([]any)
	if len(links) != 2 || member(t, links[0], "type") == member(t, links[1], "type") {
		t.Fatalf("links = %v, want two links of different types", links)
	}
	dynamics := member(t, doc, "kinds", 1, "dynamics")
	for state, link := range map[string]string{"inflow": "node__in_load__source", "outflow": "node__in_load_two__source"} {
		if got := member(t, named(t, dynamics, "dynamics", "target", state), "follow", "sources", 0, "link"); got != link {
			t.Errorf("%s follows %v, want %s", state, got, link)
		}
	}
}

func TestALinkTypeNameFitsTheLongestKindAndPortNames(t *testing.T) {
	t.Parallel()
	long := func(n int) string { return strings.Repeat("a", n) }
	files := baseFiles()
	node := mutate(t, nodeJSON, func(doc map[string]any) { member(t, doc, "ports", 0).(map[string]any)["name"] = long(20) })
	files.equipment = append(files.equipment, file("equipment/source.component.json", sourceJSON), file("equipment/node.component.json", node))
	raw := connected(t, []any{link("S-1.out_load", "N-1."+long(20))},
		entryOf("acme/source@1", long(19)+"b", "S-1"), entryOf("acme/node@1", long(19)+"c", "N-1"))
	doc := composeOK(t, loadLibrary(t, files.directory("lib")), raw)
	name := member(t, doc, "link_types", 0, "name").(string)
	if len(name) != 64 {
		t.Fatalf("link type %q has %d characters, want the 64 the longest names reach", name, len(name))
	}
}

func templatedSource(template string, count map[string]any) map[string]any {
	return map[string]any{"use": "acme/source@1", "as": "source", "id_template": template, "count": count}
}

func TestConnectionsAcceptOnlyIdsATemplateRenders(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, instance, template string
		count                    map[string]any
		renders                  bool
	}{
		{"the last default index", "line.2", "line.{n}", map[string]any{"default": 2}, true},
		{"an index beyond the default without a maximum", "line.3", "line.{n}", map[string]any{"default": 2}, false},
		{"an index up to the maximum", "line.5", "line.{n}", map[string]any{"default": 2, "max": 5}, true},
		{"an index beyond the maximum", "line.6", "line.{n}", map[string]any{"default": 2, "max": 5}, false},
		{"index zero", "line.0", "line.{n}", map[string]any{"default": 2}, false},
		{"a leading zero the template does not render", "line.02", "line.{n}", map[string]any{"default": 2}, false},
		{"a named placeholder takes the characters of an id", "north.1", "{site}.{n}", map[string]any{"default": 2}, true},
		{"another literal", "row.1", "line.{n}", map[string]any{"default": 2}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			raw := connected(t, []any{link(tc.instance+".out_load", "K-1.load_in")}, templatedSource(tc.template, tc.count), entryOf("acme/kiln@1", "kiln", "K-1"))
			_, err := compose(t, plantLibrary(t), raw)
			if tc.renders && err != nil {
				t.Fatalf("%s: %v", tc.instance, err)
			}
			if !tc.renders {
				wantIssue(t, err, "composition-v0.1", "/connections/0/from", "names no instance of the composition")
			}
		})
	}
}
