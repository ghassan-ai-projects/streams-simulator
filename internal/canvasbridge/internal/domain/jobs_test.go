package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestThePerturbationCatalogIsAReadOnlyDirectorAction(t *testing.T) {
	t.Parallel()
	action, err := Lookup("catalog.perturbations")
	if err != nil || action.Kind != KindMCP || action.Tool != "sim.catalog.perturbations" || action.Mutating {
		t.Fatalf("catalog.perturbations = %+v, %v", action, err)
	}
	if got, err := action.ToolArguments(map[string]any{}); err != nil || len(got) != 0 {
		t.Fatalf("ToolArguments = %v, %v", got, err)
	}
	if _, err := action.ToolArguments(map[string]any{"world_id": "w-1"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("an argument the catalog does not take: %v", err)
	}
}

func TestJobActionsMapToTheDirectorsJobTools(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, tool string
		mutating   bool
	}{
		{"job.start", "sim.job.start", true},
		{"job.status", "sim.job.status", false},
		{"job.stop", "sim.job.stop", true},
		{"job.list", "sim.job.list", false},
		{"job.pause", "sim.job.pause", true},
		{"job.resume", "sim.job.resume", true},
		{"job.step", "sim.job.step", true},
		{"job.rate", "sim.job.rate", true},
		{"job.inject_fault", "sim.job.inject_fault", true},
		{"job.invoke_effector", "sim.job.invoke_effector", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			action, err := Lookup(tc.name)
			if err != nil || action.Kind != KindMCP || action.Tool != tc.tool || action.Mutating != tc.mutating {
				t.Fatalf("%s = %+v, %v", tc.name, action, err)
			}
		})
	}
}

func TestJobStartCarriesItsNestedPlanAndNothingElse(t *testing.T) {
	t.Parallel()
	start, _ := Lookup("job.start")
	plan := map[string]any{
		"name": "Loss and recovery", "purpose": "Does it recover?",
		"domain": "d", "seed": 3.0, "explore": true, "duration_s": 60.0, "rate": 0.0,
		"faults":        []any{map[string]any{"entity_id": "CT-1", "fault": "f", "at_s": 10.0, "until_s": 30.0}},
		"perturbations": []any{map[string]any{"kind": "drop", "params": map[string]any{"rate": 0.3}, "at_s": 5.0, "until_s": 40.0}},
		"calls":         []any{map[string]any{"entity_id": "CT-1", "effector": "e", "args": map[string]any{"speed": 0.8}, "at_s": 5.0}},
	}
	if got, err := start.ToolArguments(plan); err != nil || !reflect.DeepEqual(got, plan) {
		t.Fatalf("ToolArguments = %v, %v", got, err)
	}
	for _, key := range []string{"token", "world_id", "sink", "sink_target", "adapter", "unblind"} {
		if _, err := start.ToolArguments(map[string]any{"domain": "d", key: "x"}); !errors.Is(err, ErrBadArguments) {
			t.Errorf("job.start took %q: %v", key, err)
		}
	}
}

func TestJobStatusStopAndListTakeOnlyTheWorld(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"job.status", "job.stop"} {
		action, _ := Lookup(name)
		if _, err := action.ToolArguments(map[string]any{"world_id": "w-1"}); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if _, err := action.ToolArguments(map[string]any{"world_id": "w-1", "token": "t"}); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s took a token: %v", name, err)
		}
	}
	list, _ := Lookup("job.list")
	if _, err := list.ToolArguments(map[string]any{"world_id": "w-1"}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("job.list took an argument: %v", err)
	}
}

func TestJobResultsNeverCarryTheOperatorCapability(t *testing.T) {
	t.Parallel()
	start, _ := Lookup("job.start")
	got, err := start.Withhold([]byte(`{"world_id":"w-1","state":"running","token":"t-secret","operator_endpoint":"http://127.0.0.1:1"}`))
	if err != nil || strings.Contains(string(got), "t-secret") || strings.Contains(string(got), "operator_endpoint") || !strings.Contains(string(got), `"world_id":"w-1"`) {
		t.Fatalf("Withhold = %s, %v", got, err)
	}
}

func TestJobsOfReadsOneJobOrAList(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, result string
		want         []JobView
	}{
		{"one job", `{"world_id":"w-1","state":"running","calls":[]}`, []JobView{{WorldID: "w-1", State: "running"}}},
		{"a named job", `{"world_id":"w-1","state":"running","name":"Loss","domain":"pond","explore":true,"faults":[]}`,
			[]JobView{{WorldID: "w-1", State: "running", Name: "Loss", Domain: "pond", Explore: true}}},
		{"a list", `{"jobs":[{"world_id":"w-1","state":"completed"},{"world_id":"w-2","state":"stopping"}]}`,
			[]JobView{{WorldID: "w-1", State: "completed"}, {WorldID: "w-2", State: "stopping"}}},
		{"an empty list", `{"jobs":[]}`, []JobView{}},
		{"a job without a state", `{"world_id":"w-1"}`, []JobView{}},
		{"not an object", `[1]`, nil},
		{"not JSON", `nope`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := JobsOf([]byte(tc.result))
			if len(got) != len(tc.want) || (len(got) > 0 && !reflect.DeepEqual(got, tc.want)) {
				t.Fatalf("JobsOf(%s) = %v, want %v", tc.result, got, tc.want)
			}
		})
	}
}

func TestOnlyAFinishedJobIsNotActive(t *testing.T) {
	t.Parallel()
	for state, active := range map[string]bool{"running": true, "stopping": true, "completed": false, "stopped": false, "failed": false, "paused": true} {
		if got := (JobView{WorldID: "w-1", State: state}).Active(); got != active {
			t.Errorf("%s active = %v, want %v", state, got, active)
		}
	}
}

func TestJobControlsTakeTheirClosedKeyListsAndNothingElse(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]any{
		"job.pause":           {"world_id": "w-1"},
		"job.resume":          {"world_id": "w-1"},
		"job.step":            {"world_id": "w-1"},
		"job.rate":            {"world_id": "w-1", "rate": 2.5},
		"job.inject_fault":    {"world_id": "w-1", "entity_id": "e", "fault": "f"},
		"job.invoke_effector": {"world_id": "w-1", "entity_id": "e", "effector": "x", "args": map[string]any{"level": 1.0}},
	}
	for name, args := range cases {
		action, err := Lookup(name)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := action.ToolArguments(args); err != nil || !reflect.DeepEqual(got, map[string]any(args)) {
			t.Errorf("%s: ToolArguments = %v, %v", name, got, err)
		}
		for _, key := range []string{"token", "command_id", "onset_ns", "params"} {
			bad := map[string]any{key: "x"}
			if _, err := action.ToolArguments(bad); !errors.Is(err, ErrBadArguments) {
				t.Errorf("%s took %q: %v", name, key, err)
			}
		}
	}
}

func TestARunEntryOfARunningJobTakesItsNamePlantAndModeFromTheJob(t *testing.T) {
	t.Parallel()
	job := JobView{WorldID: "w-1", State: "running", Name: "Loss", Domain: "pond", Explore: true}
	entry := RunEntry{WorldID: "w-1", Running: true}.WithJob(job)
	if entry.Name != "Loss" || entry.Plant != "pond" || !entry.Explore || entry.WorldID != "w-1" || !entry.Running {
		t.Fatalf("entry = %+v, want the job's name, plant and mode", entry)
	}
	unknown := RunEntry{WorldID: "w-2", Name: "kept", Plant: "kept", Explore: true}.WithJob(JobView{})
	if unknown.Name != "kept" || unknown.Plant != "kept" || !unknown.Explore {
		t.Fatalf("entry = %+v, want a world that is no job left as it was", unknown)
	}
	blank := RunEntry{WorldID: "w-1", Name: "kept", Plant: "kept"}.WithJob(JobView{WorldID: "w-1", State: "running"})
	if blank.Name != "kept" || blank.Plant != "kept" || blank.Explore {
		t.Fatalf("entry = %+v, want empty job fields to invent nothing but the mode", blank)
	}
}
