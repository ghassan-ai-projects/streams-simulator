package app

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	fanChoices = "twins/fan.twin.json"
	twinResult = `{"id":"fan-twin","implements":["fan"]}`
)

func TestKindTwinRunsTheKindCommandOnAWorkspaceChoicesFileAndWritesADomain(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, fanChoices, "{}")
	h.command.Stdout = []byte(twinResult)
	args := map[string]any{"kinds": []any{"fan"}, "choices": fanChoices, "out": inDomains("fan.domain.json")}
	result, err := h.bridge.RunAction(context.Background(), "kind.twin", args)
	want := []string{"kind", "twin", h.dir + "/" + fanKind, "--choices", h.dir + "/" + fanChoices, "--out", h.dir + "/" + inDomains("fan.domain.json"),
		"--components-dir", "/lib"}
	if err != nil || string(result) != twinResult || !slices.Equal(h.argv[0], want) {
		t.Fatalf("kind.twin = %s, %v with argv %v, want %v", result, err, h.argv, want)
	}
}

func TestKindTwinReadsOnlyTheKindsItIsGivenNeverTheWholeKindsFolder(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, "kinds/pump_monitor.kind.json", `{"kind":"pump_monitor"}`)
	args := map[string]any{"kinds": []any{"fan"}, "choices": fanChoices, "out": inDomains("fan.domain.json")}
	if _, err := h.bridge.RunAction(context.Background(), "kind.twin", args); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(h.argv[0], "--kinds-dir") || slices.Contains(h.argv[0], h.dir+"/kinds/pump_monitor.kind.json") {
		t.Fatalf("argv = %v, want the fan kind file only: another kind of the workspace is not part of this twin", h.argv[0])
	}
}

func TestKindTwinAndKindSuggestResolveKindNamesToWorkspaceKindFilesInOrderWithoutRepeats(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, "kinds/thermal_sensor.kind.json", `{"kind":"thermal_sensor"}`)
	named := []any{"thermal_sensor", "fan", "thermal_sensor"}
	for action, extra := range map[string]map[string]any{"kind.suggest": {}, "kind.twin": {"choices": fanChoices, "out": inDomains("fan.domain.json")}} {
		args := map[string]any{"kinds": named}
		maps.Copy(args, extra)
		if _, err := h.bridge.RunAction(context.Background(), action, args); err != nil {
			t.Fatalf("%s: %v", action, err)
		}
	}
	for _, argv := range h.argv {
		files := slices.DeleteFunc(slices.Clone(argv), func(word string) bool { return !strings.HasSuffix(word, ".kind.json") })
		if want := []string{h.dir + "/kinds/thermal_sensor.kind.json", h.dir + "/" + fanKind}; !slices.Equal(files, want) {
			t.Errorf("argv %v names kind files %v, want %v", argv, files, want)
		}
	}
}

func TestKindTwinAndKindSuggestRefuseKindsThatAreNotTheNameOfAKindFileOfTheWorkspace(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		kinds any
		want  error
	}{
		{"a folder and suffix", []any{fanKind}, domain.ErrBadArguments},
		{"a path out of the workspace", []any{"../fan"}, domain.ErrBadArguments},
		{"no kind", []any{}, domain.ErrBadArguments},
		{"no list", "fan", domain.ErrBadArguments},
		{"a name that is not text", []any{7}, domain.ErrBadArguments},
		{"a kind file the workspace does not hold", []any{"fan", "valve_monitor"}, domain.ErrAbsent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.write(t, fanKind, fanKindBytes)
			for _, action := range []string{"kind.suggest", "kind.twin"} {
				args := map[string]any{"kinds": tc.kinds, "choices": fanChoices, "out": inDomains("fan.domain.json")}
				if action == "kind.suggest" {
					args = map[string]any{"kinds": tc.kinds}
				}
				if _, err := h.bridge.RunAction(context.Background(), action, args); !errors.Is(err, tc.want) {
					t.Errorf("%s with kinds %v = %v, want %v", action, tc.kinds, err, tc.want)
				}
			}
			if len(h.argv) != 0 {
				t.Fatalf("argv = %v, want no command run for a refused kind", h.argv)
			}
		})
	}
}

func TestKindTwinRefusesToReplaceADomainWhoseVersionTheCallerDoesNotHold(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, inDomains("fan.domain.json"), `{"id":"hand-edited"}`)
	_, err := h.bridge.RunAction(context.Background(), "kind.twin", map[string]any{"kinds": []any{"fan"}, "choices": fanChoices, "out": inDomains("fan.domain.json"), "etag": ""})
	var conflict *domain.ConflictError
	if !errors.As(err, &conflict) || conflict.Path != inDomains("fan.domain.json") || len(h.argv) != 0 {
		t.Fatalf("kind.twin over an existing domain = %v with argv %v, want a version conflict before the command runs", err, h.argv)
	}
}

func TestTheDirectorIsStartedWithTheWorkspaceKindsFolderOnceItExists(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if args := h.bridge.DirectorArguments(); slices.Contains(args, "--kinds-dir") {
		t.Fatalf("director arguments = %v, want no kinds directory before the workspace has one", args)
	}
	h.write(t, fanKind, fanKindBytes)
	args := h.bridge.DirectorArguments()
	index := slices.Index(args, "--kinds-dir")
	if index < 0 || args[index+1] != filepath.Join(h.dir, "kinds") {
		t.Fatalf("director arguments = %v, want the workspace kinds folder", args)
	}
}

func TestKindTwinNamingAKindFileTheWorkspaceLacksIsRefusedBeforeTheVersionIsLookedAt(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, inDomains("fan.domain.json"), `{"id":"hand-edited"}`)
	_, err := h.bridge.RunAction(context.Background(), "kind.twin", map[string]any{"kinds": []any{"fan"}, "choices": fanChoices, "out": inDomains("fan.domain.json"), "etag": ""})
	if !errors.Is(err, domain.ErrAbsent) || len(h.argv) != 0 {
		t.Fatalf("kind.twin of an absent kind = %v with argv %v, want not found before any command runs", err, h.argv)
	}
}
