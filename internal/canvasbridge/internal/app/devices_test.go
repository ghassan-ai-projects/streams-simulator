package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const (
	fanKind      = "kinds/fan.kind.json"
	boardFile    = "components/equipment/board.component.json"
	dhtFile      = "components/instruments/dht.instrument.json"
	fanKindBytes = `{"kind":"fan","metrics":[],"commands":[]}`

	commandsIssue = `{"pointer":"/commands","message":"missing required property \"commands\""}`
)

func TestAnEmptyWorkspaceListsNoKindsAndNoComponentsAsEmptyArrays(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	listing, err := h.bridge.Workspace()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(listing)
	if err != nil || !strings.Contains(string(raw), `"kinds":[]`) || !strings.Contains(string(raw), `"components":[]`) {
		t.Fatalf("listing = %s, %v", raw, err)
	}
}

func TestTheWorkspaceListsKindsAndComponentsWithTheirTopLevelNames(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, "kinds/broken.kind.json", "{")
	h.write(t, "kinds/notes.txt", "x")
	h.write(t, "kinds/.hidden.kind.json", `{"kind":"hidden"}`)
	h.write(t, boardFile, `{"id":"lab/board","title":"Board","class":"static/board","kind":{"state":[]}}`)
	h.write(t, dhtFile, `{"id":"lab/dht","title":"DHT sensor"}`)
	h.write(t, "components/vocabulary.json", "{}")
	h.write(t, "components/equipment/notes.txt", "x")
	listing, err := h.bridge.Workspace()
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []domain.KindEntry{{Path: "kinds/broken.kind.json"}, {Path: fanKind, Kind: "fan"}}
	wantComponents := []domain.ComponentEntry{
		{Path: boardFile, ID: "lab/board", Title: "Board", Class: domain.ClassEquipment},
		{Path: dhtFile, ID: "lab/dht", Title: "DHT sensor", Class: domain.ClassInstrument},
	}
	if !reflect.DeepEqual(listing.Kinds, wantKinds) || !reflect.DeepEqual(listing.Components, wantComponents) {
		t.Fatalf("kinds = %+v\ncomponents = %+v", listing.Kinds, listing.Components)
	}
	raw, _ := json.Marshal(listing.Components[0])
	if want := `{"path":"` + boardFile + `","id":"lab/board","title":"Board","class":"equipment"}`; string(raw) != want {
		t.Fatalf("component entry = %s, want %s", raw, want)
	}
}

func TestKindAndComponentFilesAreReadAndWrittenWithTheEtagRules(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, rel := range []string{fanKind, boardFile, dhtFile} {
		written, err := h.bridge.WriteFile(rel, fanKindBytes, "")
		if err != nil || written.Etag != domain.Version([]byte(fanKindBytes)) {
			t.Fatalf("WriteFile(%s) = %+v, %v", rel, written, err)
		}
		if read, err := h.bridge.ReadFile(rel); err != nil || read.Content != fanKindBytes || read.Etag != written.Etag {
			t.Fatalf("ReadFile(%s) = %+v, %v", rel, read, err)
		}
		if _, err := h.bridge.WriteFile(rel, "{}", "stale"); !errors.Is(err, domain.ErrConflict) {
			t.Errorf("a stale write of %s: %v", rel, err)
		}
	}
}

func TestTheDeviceFoldersRefuseEveryPathOutsideTheirFiles(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, "runs/w-1/run.json", "{}")
	for _, rel := range []string{"kinds/fan.json", "kinds/fan.domain.json", "kinds/../runs/w-1/run.json", "../kinds/fan.kind.json",
		"kinds/sub/fan.kind.json", "components/board.component.json", "components/equipment/dht.instrument.json",
		"components/instruments/board.component.json", "components/vocabulary.json", "components/equipment/../../x.component.json"} {
		if _, err := h.bridge.WriteFile(rel, "{}", ""); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("WriteFile(%q) = %v, want ErrBadPath", rel, err)
		}
		if _, err := h.bridge.ReadFile(rel); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("ReadFile(%q) = %v, want ErrBadPath", rel, err)
		}
	}
	if matches, _ := filepath.Glob(filepath.Join(h.dir, "kinds", "*")); len(matches) != 0 {
		t.Fatalf("a refused write left %v", matches)
	}
}

func TestASymlinkOutOfTheWorkspaceIsNeverFollowedInTheDeviceFolders(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.kind.json")
	escape := filepath.Join(outside, "escape.component.json")
	for _, file := range []string{secret, escape} {
		if err := os.WriteFile(file, []byte("secret"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	h.write(t, "kinds/seed.kind.json", "{}")
	h.write(t, "components/seed.txt", "x")
	if err := os.Symlink(secret, filepath.Join(h.dir, "kinds", "leak.kind.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(h.dir, "components", "equipment")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"kinds/leak.kind.json", "components/equipment/escape.component.json"} {
		if _, err := h.bridge.ReadFile(rel); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("ReadFile(%q) = %v, want ErrBadPath", rel, err)
		}
	}
	if _, err := h.bridge.WriteFile("kinds/leak.kind.json", "{}", ""); !errors.Is(err, domain.ErrBadPath) {
		t.Fatalf("writing through a symlink: %v", err)
	}
	if got, _ := os.ReadFile(secret); string(got) != "secret" {
		t.Fatalf("the file outside the workspace was changed: %q", got)
	}
}

func TestKindValidateRunsTheKindCommandOnAWorkspaceFileAndRelaysItsRefusal(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	result, err := h.bridge.RunAction(context.Background(), "kind.validate", map[string]any{"path": fanKind})
	if err != nil || string(result) != `{"valid":true}` || !slices.Equal(h.argv[0], []string{"kind", "validate", h.dir + "/" + fanKind}) {
		t.Fatalf("a good kind: %s, %v, argv %v", result, err, h.argv)
	}
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"kind_failed","message":"` + h.dir +
		`/kinds/thin.kind.json: kind file fails its schema","source":"` + h.dir + `/kinds/thin.kind.json","issues":[` + commandsIssue + `]}}`)}
	var failure *domain.Failure
	_, err = h.bridge.RunAction(context.Background(), "kind.validate", map[string]any{"path": "kinds/thin.kind.json"})
	if !errors.As(err, &failure) || failure.Code != "kind_failed" || failure.Message != "kinds/thin.kind.json: kind file fails its schema" ||
		string(failure.Issues) != "["+commandsIssue+"]" {
		t.Fatalf("a bad kind: %v", err)
	}
	if _, err := h.bridge.RunAction(context.Background(), "kind.validate", map[string]any{"path": inDomains("x.domain.json")}); !errors.Is(err, domain.ErrBadArguments) {
		t.Fatalf("a domain file given to kind.validate: %v", err)
	}
}

func TestWorkspaceKindsAndComponentsReachTheCommandsOnlyWhenTheirFoldersExist(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	run := func(name string, args map[string]any) []string {
		t.Helper()
		if _, err := h.bridge.RunAction(context.Background(), name, args); err != nil {
			t.Fatal(err)
		}
		return h.argv[len(h.argv)-1]
	}
	validate := map[string]any{"path": plant}
	if got := run("domain.validate", validate); !slices.Equal(got, []string{"domain", "validate", h.dir + "/" + plant}) {
		t.Fatalf("without a kinds folder: %v", got)
	}
	if got := run("component.validate", nil); !slices.Equal(got, []string{"component", "validate", "--components-dir", "/lib"}) {
		t.Fatalf("without a components folder: %v", got)
	}
	h.write(t, fanKind, fanKindBytes)
	h.write(t, boardFile, "{}")
	kindsDir, componentsDir := filepath.Join(h.dir, "kinds"), filepath.Join(h.dir, "components")
	if got := run("domain.validate", validate); !slices.Equal(got, []string{"domain", "validate", h.dir + "/" + plant, "--kinds-dir", kindsDir}) {
		t.Fatalf("with a kinds folder: %v", got)
	}
	if got := run("component.validate", nil); !slices.Equal(got, []string{"component", "validate", "--components-dir", "/lib", "--components-dir", componentsDir}) {
		t.Fatalf("with a components folder: %v", got)
	}
	compose := map[string]any{"composition": "compositions/f.composition.json", "out": inDomains("f.domain.json")}
	want := []string{"domain", "compose", h.dir + "/compositions/f.composition.json", "--out", h.dir + "/" + inDomains("f.domain.json"),
		"--components-dir", "/lib", "--components-dir", componentsDir, "--kinds-dir", kindsDir}
	if got := run("domain.compose", compose); !slices.Equal(got, want) {
		t.Fatalf("compose = %v, want %v", got, want)
	}
	if got := run("kind.validate", map[string]any{"path": fanKind}); slices.Contains(got, "--kinds-dir") {
		t.Fatalf("kind.validate must not take the folder: %v", got)
	}
}

func TestWatchingTellsWhenAKindOrAComponentFileChanges(t *testing.T) {
	t.Parallel()
	for _, rel := range []string{fanKind, boardFile, dhtFile} {
		h := newHarness(t)
		ctx, cancel := context.WithCancel(context.Background())
		changes := make(chan struct{}, 4)
		done := make(chan error, 1)
		go func() {
			done <- h.bridge.Watch(ctx, func() error {
				changes <- struct{}{}
				return nil
			})
		}()
		h.ticks <- time.Time{}
		h.replace(t, rel, fanKindBytes)
		h.ticks <- time.Time{}
		<-changes
		h.ticks <- time.Time{}
		cancel()
		if err := <-done; err != nil || len(changes) != 0 {
			t.Fatalf("Watch after writing %s: %v, %d extra changes", rel, err, len(changes))
		}
	}
}
