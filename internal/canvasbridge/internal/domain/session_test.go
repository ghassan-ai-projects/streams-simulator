package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestTheSessionActionsRunPlantWireConnectWithThePlanOnStandardInput(t *testing.T) {
	t.Parallel()
	check, _ := Lookup("session.check")
	start, _ := Lookup("session.start")
	got, err := check.Argv(map[string]any{"content": "{}"})
	if err != nil || !reflect.DeepEqual(got, literals([]string{"plantwire", "connect", "--check", "--stdin"})) || check.Lasting != "" || check.Mutating {
		t.Fatalf("check argv = %+v, %v, want plantwire connect --check --stdin, read-only and not lasting", got, err)
	}
	stdin, err := start.Stdin(map[string]any{"content": `{"schema_version":"0.1"}`})
	if err != nil || string(stdin) != `{"schema_version":"0.1"}` || start.Lasting != LastingStart || !start.Mutating || !start.Catalog || !start.Kinds {
		t.Fatalf("start = %+v, stdin %q, %v, want a lasting, mutating command given the catalog and kinds", start, stdin, err)
	}
	if _, err := start.Stdin(map[string]any{}); !errors.Is(err, ErrBadArguments) {
		t.Fatalf("a start without a plan = %v, want ErrBadArguments", err)
	}
	for _, name := range []string{"session.status", "session.stop"} {
		action, _ := Lookup(name)
		if _, err := action.Argv(map[string]any{"run": "c-1"}); !errors.Is(err, ErrBadArguments) || action.Lasting == "" {
			t.Fatalf("%s with an argument = %v, want ErrBadArguments from a lasting action that takes none", name, err)
		}
	}
}

func TestASessionThatExitsCleanlyEndedAndOneThatFailsSaysWhy(t *testing.T) {
	t.Parallel()
	view := SessionView{Run: "c-1", Domain: PlanDomain([]byte(`{"domain":"arduino-thermal"}`)), Plan: []byte(`{"domain":"arduino-thermal"}`)}
	ended := view.Outcome([]byte(`{"out":"runs/c-1","stopped":true}`), 0, "")
	if ended.State != SessionEnded || ended.Run != "c-1" || ended.Domain != "arduino-thermal" || string(ended.Result) != `{"out":"runs/c-1","stopped":true}` ||
		ended.Failure != nil {
		t.Fatalf("a clean exit = %+v, want ended with the command's document, the domain and the plan", ended)
	}
	refused := view.Outcome([]byte(`{"error":{"code":"plantwire_failed","message":"no host"}}`), 1, "")
	if refused.State != SessionFailed || refused.Failure.Code != "plantwire_failed" || refused.Result != nil {
		t.Fatalf("a refused session = %+v, want failed with the command's error", refused)
	}
	crashed := view.Outcome(nil, -1, "panic: boom\n")
	if crashed.State != SessionFailed || crashed.Failure.Code != "command_failed" || crashed.Failure.Message != "panic: boom" {
		t.Fatalf("a crashed session = %+v, want failed with its last words", crashed)
	}
	huge := view.Outcome([]byte(`"`+strings.Repeat("x", MaxSessionResult)+`"`), 0, "")
	if huge.State != SessionEnded || huge.Result != nil {
		t.Fatalf("an oversized result = state %s with %d bytes, want ended without the result", huge.State, len(huge.Result))
	}
	if unlaunched := view.Unlaunched(errors.New("no such file")); unlaunched.State != SessionFailed || unlaunched.Failure.Message != "no such file" {
		t.Fatalf("Unlaunched = %+v, want failed with the launch error", unlaunched)
	}
	if PlanDomain([]byte(`not json`)) != "" || SessionPrecheck().Name != "session.check" {
		t.Fatal("a check that is not a document names no domain, and the precheck is session.check")
	}
}

func TestASessionRunFolderIsARunOfTheWorkspace(t *testing.T) {
	t.Parallel()
	if part, err := SessionFolder("c-0a1b2c3d"); err != nil || part.Path != "runs/c-0a1b2c3d" {
		t.Fatalf("SessionFolder = %+v, %v", part, err)
	}
	if _, err := SessionFolder(".."); !errors.Is(err, ErrBadPath) {
		t.Fatalf("a climbing run id = %v, want ErrBadPath", err)
	}
	if failure := SessionStillRunning("c-1"); failure.Code != "session_running" {
		t.Fatalf("SessionStillRunning = %+v, want session_running", failure)
	}
	if entry := (RunEntry{WorldID: "c-1"}).AsSession(true); !entry.Connected || !entry.Running {
		t.Fatalf("AsSession = %+v, want a running connected session", entry)
	}
}
