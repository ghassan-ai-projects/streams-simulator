package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestEveryActionNamesItsToolOrCommandOnce(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, action := range Actions() {
		if seen[action.Name] || (action.Kind == KindMCP) == (action.Tool == "") || (action.Kind == KindCLI) == (len(action.Command) == 0) {
			t.Errorf("action %+v: a unique name and exactly one of tool or command", action)
		}
		seen[action.Name] = true
	}
	if _, err := Lookup("fault.list"); !errors.Is(err, ErrUnknownAction) {
		t.Fatal("the director's fault list reveals truth in Evaluate runs and has no action")
	}
}

func TestEffectorActionsTakeTheWorldAndNeverTheCapabilityToken(t *testing.T) {
	t.Parallel()
	list, _ := Lookup("effector.list")
	invoke, _ := Lookup("effector.invoke")
	if list.Tool != "sim.world.effectors" || list.Mutating || invoke.Tool != "sim.world.invoke" || !invoke.Mutating {
		t.Fatalf("effector actions = %+v, %+v", list, invoke)
	}
	args := map[string]any{"world_id": "w-1", "entity_id": "P-1", "effector": "pump__start", "args": map[string]any{"pump_id": "P-1"}, "command_id": "c-1"}
	if got, err := invoke.ToolArguments(args); err != nil || !reflect.DeepEqual(got, args) {
		t.Fatalf("ToolArguments = %v, %v", got, err)
	}
	if _, err := invoke.ToolArguments(map[string]any{"world_id": "w-1", "token": "t-secret"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a token passed through: %v", err)
	}
}

func TestToolArgumentsAreClosedAndNeverUnblind(t *testing.T) {
	t.Parallel()
	reveal, err := Lookup("truth.reveal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reveal.ToolArguments(map[string]any{"run_id": "r", "unblind": true}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("unblind passed through: %v", err)
	}
	create, _ := Lookup("world.create")
	if _, err := create.ToolArguments(map[string]any{"domain": "d", "sink": "file", "sink_target": "/etc/x"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a sink target outside the workspace passed through: %v", err)
	}
	args, err := create.ToolArguments(map[string]any{"domain": "d", "seed": 7.0, "explore": true})
	if err != nil || !reflect.DeepEqual(args, map[string]any{"domain": "d", "seed": 7.0, "explore": true}) {
		t.Fatalf("ToolArguments = %v, %v", args, err)
	}
}

func TestCommandArgumentsBecomeSeparateArgvEntriesOfWorkspacePaths(t *testing.T) {
	t.Parallel()
	compose, _ := Lookup("domain.compose")
	got, err := compose.Argv(map[string]any{"composition": "compositions/f.composition.json", "out": inDomains("f.domain.json")})
	want := []ArgvPart{{Literal: "domain"}, {Literal: "compose"}, {Path: "compositions/f.composition.json"}, {Literal: "--out"}, {Path: inDomains("f.domain.json")}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v", got, err)
	}
	lint, _ := Lookup("domain.lint")
	if got, err := lint.Argv(map[string]any{"path": inDomains("f.domain.json"), "hours": 2.0}); err != nil || got[len(got)-1].Literal != "2" {
		t.Fatalf("lint argv = %+v, %v", got, err)
	}
	verify, _ := Lookup("run.verify")
	if got, err := verify.Argv(map[string]any{"run": "w-1"}); err != nil || got[1].Path != "runs/w-1/run.json" {
		t.Fatalf("verify argv = %+v, %v", got, err)
	}
	report, _ := Lookup("run.report")
	if got, err := report.Argv(map[string]any{"run": "w-1"}); err != nil || got[1].Path != "runs/w-1" {
		t.Fatalf("report argv = %+v, %v", got, err)
	}
}

func TestCommandArgumentsRefuseInjectionTraversalAndUnknownKeys(t *testing.T) {
	t.Parallel()
	compose, _ := Lookup("domain.compose")
	show, _ := Lookup("component.show")
	lint, _ := Lookup("domain.lint")
	verify, _ := Lookup("run.verify")
	cases := map[string]func() error{
		"a flag as a component ref": func() error { _, err := show.Argv(map[string]any{"ref": "--components-dir=/"}); return err },
		"a missing argument": func() error {
			_, err := compose.Argv(map[string]any{"composition": "compositions/f.composition.json"})
			return err
		},
		"a path of the wrong class": func() error {
			_, err := compose.Argv(map[string]any{"composition": inDomains("f.domain.json"), "out": inDomains("f.domain.json")})
			return err
		},
		"a path outside the workspace": func() error {
			_, err := compose.Argv(map[string]any{"composition": "../x.composition.json", "out": inDomains("f.domain.json")})
			return err
		},
		"a run id that climbs": func() error { _, err := verify.Argv(map[string]any{"run": ".."}); return err },
		"a number given as text": func() error {
			_, err := lint.Argv(map[string]any{"path": inDomains("f.domain.json"), "hours": "2; rm"})
			return err
		},
		"an unknown key": func() error {
			_, err := lint.Argv(map[string]any{"path": inDomains("f.domain.json"), "seed": 1.0})
			return err
		},
		"a path given as a number":         func() error { _, err := lint.Argv(map[string]any{"path": 3.0}); return err },
		"a component ref that is a number": func() error { _, err := show.Argv(map[string]any{"ref": 3.0}); return err },
	}
	for name, run := range cases {
		if err := run(); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: %v, want ErrBadArguments", name, err)
		}
	}
}

func TestTheOperatorTokenNeverLeavesTheBridge(t *testing.T) {
	t.Parallel()
	create, _ := Lookup("world.create")
	got, err := create.Withhold([]byte(`{"world_id":"w-1","token":"t-secret","operator_endpoint":"http://127.0.0.1:1","clock":"x"}`))
	if err != nil || string(got) != `{"clock":"x","world_id":"w-1"}` {
		t.Fatalf("Withhold = %s, %v", got, err)
	}
	if _, err := create.Withhold([]byte(`not json`)); err == nil || !strings.Contains(err.Error(), "world.create result") {
		t.Fatalf("a malformed result: %v", err)
	}
	advance, _ := Lookup("clock.advance")
	if got, err := advance.Withhold([]byte(`{"clock":"x"}`)); err != nil || string(got) != `{"clock":"x"}` {
		t.Fatalf("an action without withheld keys changed its result: %s, %v", got, err)
	}
}

func TestANewPlantPassesItsEmptySwitchOnlyWhenOn(t *testing.T) {
	t.Parallel()
	create, _ := Lookup("component.new")
	got, err := create.Argv(map[string]any{"id": "lab", "title": "Lab", "from": "compositions/f.composition.json", "empty": true})
	want := []ArgvPart{{Literal: "component"}, {Literal: "new"}, {Literal: "lab"}, {Literal: "--title"}, {Literal: "Lab"},
		{Literal: "--from"}, {Path: "compositions/f.composition.json"}, {Literal: "--empty"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v", got, err)
	}
	if got, err := create.Argv(map[string]any{"id": "lab", "title": "Lab", "example": "chiller-pilot", "empty": false}); err != nil || len(got) != 7 {
		t.Fatalf("empty off = %+v, %v", got, err)
	}
	if _, err := create.Argv(map[string]any{"id": "lab", "title": "Lab", "empty": "yes"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a switch that is not a boolean = %v", err)
	}
}

func TestRunTwinReportIsACommandOnTheRunThatDoesNotMutate(t *testing.T) {
	t.Parallel()
	action, err := Lookup("run.twin_report")
	if err != nil || action.Kind != KindCLI || !reflect.DeepEqual(action.Command, []string{"plantwire", "twin-report"}) || action.Mutating || action.Etag || action.Input != "" {
		t.Fatalf("run.twin_report = %+v, %v, want a read-only CLI action on plantwire twin-report", action, err)
	}
	if !action.Domains || action.Catalog || !action.Kinds || !action.KindsHome {
		t.Fatalf("run.twin_report flags = %+v, want the workspace domains and kinds folders and no adapters", action)
	}
	got, err := action.Argv(map[string]any{"run": "r1", "domain": "arduino-thermal", "tolerance": "zone-01.temperature=2", "span_fraction": 0.1})
	want := []ArgvPart{{Literal: "plantwire"}, {Literal: "twin-report"}, {Path: "runs/r1"}, {Literal: "--domain"}, {Literal: "arduino-thermal"},
		{Literal: "--tolerance"}, {Literal: "zone-01.temperature=2"}, {Literal: "--span-fraction"}, {Literal: "0.1"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("Argv = %+v, %v, want %+v", got, err, want)
	}
	minimal, err := action.Argv(map[string]any{"run": "r1", "domain": "arduino-thermal"})
	if err != nil || len(minimal) != 5 {
		t.Fatalf("minimal Argv = %+v, %v, want the command, the run and the domain only", minimal, err)
	}
}

func TestRunTwinReportRefusesWhatItCannotRender(t *testing.T) {
	t.Parallel()
	action, _ := Lookup("run.twin_report")
	cases := map[string]map[string]any{
		"no run":                  {"domain": "arduino-thermal"},
		"no domain":               {"run": "r1"},
		"a flag as the domain":    {"run": "r1", "domain": "--kinds-dir=/"},
		"a flag as the tolerance": {"run": "r1", "domain": "d", "tolerance": "-x"},
		"an empty tolerance":      {"run": "r1", "domain": "d", "tolerance": ""},
		"a span given as text":    {"run": "r1", "domain": "d", "span_fraction": "0.1"},
		"a run id that climbs":    {"run": "../x", "domain": "d"},
		"an argument it lacks":    {"run": "r1", "domain": "d", "kinds": "kinds"},
		"the adapters directory":  {"run": "r1", "domain": "d", "adapters_dir": "/"},
	}
	for name, args := range cases {
		if _, err := action.Argv(args); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: %v, want ErrBadArguments", name, err)
		}
	}
}
