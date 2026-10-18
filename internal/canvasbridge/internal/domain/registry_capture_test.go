package domain

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestKindCaptureIsReadOnlyAndComparesWithTheWorkspaceKinds(t *testing.T) {
	t.Parallel()
	capture, err := Lookup("kind.capture")
	if err != nil || capture.Mutating || capture.Etag || !capture.Kinds || capture.Library || capture.Catalog || capture.Input != "content" {
		t.Fatalf("kind.capture = %+v, %v, want a read-only action that takes the kinds folder for the comparison", capture, err)
	}
}

func TestKindCaptureTakesItsContentOnStandardInputOrTheObserverJournalOfARun(t *testing.T) {
	t.Parallel()
	capture, _ := Lookup("kind.capture")
	cases := []struct {
		name  string
		args  map[string]any
		argv  []ArgvPart
		stdin []byte
	}{
		{"content", map[string]any{"content": "{}"}, []ArgvPart{{Literal: "kind"}, {Literal: "capture"}, {Literal: "--stdin"}}, []byte("{}")},
		{"empty content", map[string]any{"content": ""}, []ArgvPart{{Literal: "kind"}, {Literal: "capture"}, {Literal: "--stdin"}}, []byte{}},
		{"a run", map[string]any{"run": "w-1"}, []ArgvPart{{Literal: "kind"}, {Literal: "capture"}, {Path: "runs/w-1/journal.jsonl"}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			argv, err := capture.Argv(tc.args)
			stdin, stdinErr := capture.Stdin(tc.args)
			if err != nil || stdinErr != nil || !reflect.DeepEqual(argv, tc.argv) || !reflect.DeepEqual(stdin, tc.stdin) {
				t.Fatalf("Argv = %+v, %v and Stdin = %q, %v, want %+v and %q", argv, err, stdin, stdinErr, tc.argv, tc.stdin)
			}
		})
	}
}

func TestKindCaptureNeedsExactlyOneSourceAndNoOtherArgument(t *testing.T) {
	t.Parallel()
	capture, _ := Lookup("kind.capture")
	cases := map[string]map[string]any{
		"neither":                nil,
		"both":                   {"content": "{}", "run": "w-1"},
		"a path":                 {"content": "{}", "path": "runs/w-1/journal.jsonl"},
		"the sealed journal":     {"path": "runs/w-1/sealed/journal.jsonl"},
		"an out file":            {"run": "w-1", "out": inDomains("x.domain.json")},
		"an etag":                {"run": "w-1", "etag": ""},
		"the twin switch":        {"content": "{}", "twin": true},
		"a flag as a run":        {"run": "--kinds-dir"},
		"a run that climbs":      {"run": ".."},
		"a run with a folder":    {"run": "w-1/sealed"},
		"a run that is a path":   {"run": "../../etc/passwd"},
		"a run with a separator": {"run": "w-1\\sealed"},
		"an empty run":           {"run": ""},
		"a run that is a number": {"run": 7.0},
	}
	for name, args := range cases {
		if _, err := capture.Argv(args); !errors.Is(err, ErrBadArguments) {
			t.Errorf("%s: Argv = %v, want ErrBadArguments", name, err)
		}
	}
}

func TestKindCaptureStandardInputIsHeldToTheLimitAndOnlyTextIsTaken(t *testing.T) {
	t.Parallel()
	capture, _ := Lookup("kind.capture")
	if _, err := capture.Stdin(map[string]any{"content": strings.Repeat("a", MaxInputBytes+1)}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Stdin of content over the limit = %v, want ErrTooLarge", err)
	}
	for _, content := range []any{1.0, []any{"{}"}, map[string]any{"kind": "k"}, nil} {
		if _, err := capture.Stdin(map[string]any{"content": content}); !errors.Is(err, ErrBadArguments) {
			t.Errorf("Stdin of %v = %v, want ErrBadArguments", content, err)
		}
	}
}

func TestNoRunIdNamesTheSealedJournalOfAKindCapture(t *testing.T) {
	t.Parallel()
	capture, _ := Lookup("kind.capture")
	for _, run := range []string{"w-1", "run.2026", "a_b-c", strings.Repeat("a", 128)} {
		argv, err := capture.Argv(map[string]any{"run": run})
		if err != nil || len(argv) != 3 || argv[2].Path != "runs/"+run+"/journal.jsonl" || strings.Contains(argv[2].Path, "sealed") {
			t.Errorf("Argv of run %q = %+v, %v, want only the observer journal of that run", run, argv, err)
		}
	}
}
