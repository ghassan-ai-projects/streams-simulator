package files

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func directoryWith(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func selectedNames(sources []domain.KindSource, wanted ...string) []string {
	names := map[string]bool{}
	for _, kind := range wanted {
		names[kind] = true
	}
	picked := []string{}
	for _, source := range domain.SelectKindSources(sources, names) {
		picked = append(picked, filepath.Base(source.Path))
	}
	return picked
}

func TestScanReadsEveryKindFileOfTheDirectoriesAndKeepsTheFailureOfEachUnreadableOne(t *testing.T) {
	t.Parallel()
	dir := directoryWith(t, map[string]string{"a.kind.json": `{"kind": "alpha"}`, "b.kind.json": "{", "notes.txt": "ignored", "c.kind.json.bak": "ignored"})
	if err := syscall.Mkfifo(filepath.Join(dir, "pipe.kind.json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "a.kind.json"), filepath.Join(dir, "link.kind.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "huge.kind.json"), bytes.Repeat([]byte(" "), MaxKindFileBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	sources, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	failed := map[string]error{}
	var names []string
	for _, source := range sources {
		names = append(names, filepath.Base(source.Path))
		failed[filepath.Base(source.Path)] = source.Err
	}
	wantNames := []string{"a.kind.json", "b.kind.json", "huge.kind.json", "link.kind.json", "pipe.kind.json"}
	if !slices.Equal(names, wantNames) {
		t.Fatalf("Scan lists %v, want %v: only *.kind.json, in name order", names, wantNames)
	}
	for _, name := range []string{"a.kind.json", "b.kind.json"} {
		if failed[name] != nil {
			t.Errorf("%s: Err = %v, want the file read (its content is the parser's business)", name, failed[name])
		}
	}
	if !errors.Is(failed["pipe.kind.json"], kernel.ErrNotRegular) {
		t.Errorf("pipe: Err = %v, want kernel.ErrNotRegular", failed["pipe.kind.json"])
	}
	for _, name := range []string{"huge.kind.json", "link.kind.json"} {
		if failed[name] == nil || !strings.Contains(failed[name].Error(), name) {
			t.Errorf("%s: Err = %v, want a read failure naming the file", name, failed[name])
		}
	}
}

func TestScanSelectionFindsAKindByItsDeclaredNameOrByItsFileNameWhenItCannotBeParsed(t *testing.T) {
	t.Parallel()
	dir := directoryWith(t, map[string]string{"renamed.kind.json": `{"kind": "pump"}`, "valve.kind.json": `{"kind": "valve"}`,
		"fan.kind.json": "{", "twice-a.kind.json": `{"kind": "twice"}`, "twice-b.kind.json": `{"kind": "twice"}`})
	sources, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		wanted []string
		want   []string
	}{
		{"a declared kind", []string{"pump"}, []string{"renamed.kind.json"}},
		{"a file that does not parse, by its name", []string{"fan"}, []string{"fan.kind.json"}},
		{"a kind declared by two files", []string{"twice"}, []string{"twice-a.kind.json", "twice-b.kind.json"}},
		{"a kind no file holds", []string{"ghost"}, []string{}},
		{"several kinds", []string{"pump", "valve"}, []string{"renamed.kind.json", "valve.kind.json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := selectedNames(sources, tc.wanted...); !slices.Equal(got, tc.want) {
				t.Fatalf("selected %v, want %v", got, tc.want)
			}
		})
	}
}

func TestScanRefusesADirectoryItCannotListAndListsNothingForNoDirectoryOrAnEmptyOne(t *testing.T) {
	t.Parallel()
	empty := t.TempDir()
	if _, err := Scan([]string{filepath.Join(empty, "absent")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Scan of a missing directory = %v, want fs.ErrNotExist", err)
	}
	for name, dirs := range map[string][]string{"no directory": nil, "an empty directory": {empty}} {
		if sources, err := Scan(dirs); err != nil || len(sources) != 0 {
			t.Errorf("Scan of %s = %v, %v, want nothing", name, sources, err)
		}
	}
}

func TestDirectoriesAndPathsFailOnTheFirstFileTheyCannotRead(t *testing.T) {
	t.Parallel()
	dir := directoryWith(t, map[string]string{"a.kind.json": `{"kind": "alpha"}`})
	if err := syscall.Mkfifo(filepath.Join(dir, "pipe.kind.json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if sources, err := Directories([]string{dir}); !errors.Is(err, kernel.ErrNotRegular) || sources != nil {
		t.Fatalf("Directories = %v, %v, want kernel.ErrNotRegular", sources, err)
	}
	sources, err := Paths([]string{filepath.Join(dir, "a.kind.json")})
	if err != nil || len(sources) != 1 || sources[0].Err != nil || string(sources[0].Raw) != `{"kind": "alpha"}` {
		t.Fatalf("Paths = %+v, %v, want the file read", sources, err)
	}
	if _, err := Paths([]string{filepath.Join(dir, "absent.kind.json")}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Paths of a missing file = %v, want fs.ErrNotExist", err)
	}
}

func TestStreamReadsOneDocumentAndRefusesOneBeyondTheLimit(t *testing.T) {
	t.Parallel()
	source, err := Stream(strings.NewReader(`{"kind": "alpha"}`))
	if err != nil || source.Path != StreamName || string(source.Raw) != `{"kind": "alpha"}` {
		t.Fatalf("Stream = %+v, %v, want the document under the name %s", source, err, StreamName)
	}
	if _, err := Stream(bytes.NewReader(bytes.Repeat([]byte(" "), MaxKindFileBytes+1))); !errors.Is(err, kernel.ErrTooLarge) {
		t.Fatalf("Stream of an oversized document = %v, want kernel.ErrTooLarge", err)
	}
}
