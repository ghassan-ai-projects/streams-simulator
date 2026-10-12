package domain

import "testing"

func TestASystemKindHasExactlyOneNamedInstance(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"a second instance set", func(d map[string]any) {
			put(d, append(d["instances"].([]any), map[string]any{"kind": "header", "names": []any{"air_header_2"}}), "instances")
		}, "/kinds/2/system"},
		{"two names", func(d map[string]any) { put(d, []any{"air_header", "air_header_2"}, "instances", 4, "names") }, "/instances/4/names"},
		{"an id template", func(d map[string]any) {
			put(d, "header-{n}", "instances", 4, "id_template")
			put(d, map[string]any{"default": 1}, "instances", 4, "count")
		}, "/instances/4/names"},
		{"a peer residual fault", func(d map[string]any) {
			put(d, "peer_residual", "kinds", 2, "faults", 0, "observability", "detector", "form")
		}, "/kinds/2/faults/0/observability/detector/form"},
	})
}

func TestLinksAreTypedAndTheirEndsExist(t *testing.T) {
	t.Parallel()
	link := func(kind, from, to string) map[string]any {
		return map[string]any{"type": kind, "from": from, "to": to}
	}
	runViolations(t, []violationCase{
		{"an undeclared link type", func(d map[string]any) { put(d, "cools", "links", 0, "type") }, "/links/0/type"},
		{"a source that is no instance", func(d map[string]any) { put(d, "tank-x", "links", 0, "from") }, "/links/0/from"},
		{"a source of the wrong kind", func(d map[string]any) { put(d, "pump-1", "links", 0, "from") }, "/links/0/from"},
		{"a target of the wrong kind", func(d map[string]any) { put(d, "tank-2", "links", 0, "to") }, "/links/0/to"},
		{"a link stated twice", func(d map[string]any) {
			put(d, append(d["links"].([]any), link("feeds", "tank-1", "pump-1")), "links")
		}, "/links/3"},
		{"a link type declared twice", func(d map[string]any) { put(d, "feeds", "link_types", 1, "name") }, "/link_types/1/name"},
		{"a link type between undeclared kinds", func(d map[string]any) { put(d, "valve", "link_types", 0, "from") }, "/link_types/0/from"},
		{"a link type to an undeclared kind", func(d map[string]any) { put(d, "valve", "link_types", 0, "to") }, "/link_types/0/to"},
	})
}

func TestBoardsHostInstancesOnAtMostOneBoard(t *testing.T) {
	t.Parallel()
	runViolations(t, []violationCase{
		{"a member that is no instance", func(d map[string]any) { put(d, "T-gone", "boards", 0, "members", 0) }, "/boards/0/members/0"},
		{"an instance on two boards", func(d map[string]any) { put(d, "T-main", "boards", 1, "members", 0) }, "/boards/1/members/0"},
		{"an instance twice on one board", func(d map[string]any) { put(d, "tank-1", "boards", 1, "members", 1) }, "/boards/1/members/1"},
		{"a board declared twice", func(d map[string]any) { put(d, "board-01", "boards", 1, "name") }, "/boards/1/name"},
	})
}

func TestTheStructuredDomainKeepsItsLinksAndBoards(t *testing.T) {
	t.Parallel()
	c := loadStructured(t)
	if len(c.Spec.LinkTypes) != 2 || len(c.Spec.Links) != 3 || len(c.Spec.Boards) != 2 || !c.Spec.Kinds[2].System {
		t.Fatalf("structure = %+v", c.Spec)
	}
}
