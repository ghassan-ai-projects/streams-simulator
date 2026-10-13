package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func workspace(t *testing.T) (*Workspace, string) {
	t.Helper()
	dir := t.TempDir()
	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = w.Close() })
	if err := w.Ensure([]string{"plants", "runs/w-1"}); err != nil {
		t.Fatal(err)
	}
	return w, dir
}

func TestAWriteNamesTheVersionItStartedFromAndIsAtomic(t *testing.T) {
	t.Parallel()
	w, dir := workspace(t)
	first, err := w.WriteIfUnchanged("plants/a.json", []byte("one"), "", domain.Version, 100)
	if err != nil || first != domain.Version([]byte("one")) {
		t.Fatalf("a new file: %q, %v", first, err)
	}
	current, err := w.WriteIfUnchanged("plants/a.json", []byte("two"), "", domain.Version, 100)
	if !errors.Is(err, domain.ErrConflict) || current != first {
		t.Fatalf("a write from a stale version: %q, %v; want a conflict naming %q", current, err, first)
	}
	if _, err := w.WriteIfUnchanged("plants/a.json", []byte("two"), first, domain.Version, 100); err != nil {
		t.Fatal(err)
	}
	if got, err := w.Read("plants/a.json", 100); err != nil || string(got) != "two" {
		t.Fatalf("after the write: %q, %v", got, err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "plants"))
	if len(entries) != 1 {
		t.Fatalf("plants holds %d entries, want the file alone (no temporary file left)", len(entries))
	}
}

func TestReadsStayInsideTheWorkspaceAndWithinTheirLimit(t *testing.T) {
	t.Parallel()
	w, dir := workspace(t)
	outside := filepath.Join(t.TempDir(), "secret.json")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "plants", "link.json")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"plants/link.json", "../secret.json"} {
		if _, err := w.Read(rel, 100); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("Read(%q) = %v, want a bad path for leaving the workspace", rel, err)
		}
	}
	if _, err := w.Read("plants/none.json", 100); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("an absent file: %v", err)
	}
	if _, err := w.WriteIfUnchanged("plants/big.json", []byte("0123456789"), "", domain.Version, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Read("plants/big.json", 5); !errors.Is(err, domain.ErrTooLarge) {
		t.Fatalf("a file over the limit: %v", err)
	}
}

func TestListingsTailsAndFingerprintsFollowTheFolder(t *testing.T) {
	t.Parallel()
	w, _ := workspace(t)
	before := w.Fingerprint([]string{"plants"})
	if names, err := w.Names("plants"); err != nil || len(names) != 0 {
		t.Fatalf("Names = %v, %v", names, err)
	}
	if names, err := w.Names("missing"); err != nil || names != nil {
		t.Fatalf("a missing folder lists as empty: %v, %v", names, err)
	}
	if _, err := w.WriteIfUnchanged("runs/w-1/journal.jsonl", []byte("line one\nline two\n"), "", domain.Version, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteIfUnchanged("plants/b.json", []byte("x"), "", domain.Version, 100); err != nil {
		t.Fatal(err)
	}
	if tail, err := w.ReadFrom("runs/w-1/journal.jsonl", 9, 100); err != nil || string(tail) != "line two\n" {
		t.Fatalf("ReadFrom = %q, %v", tail, err)
	}
	if _, err := w.ReadFrom("runs/w-2/journal.jsonl", 0, 10); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("tail of an absent journal: %v", err)
	}
	if w.Fingerprint([]string{"plants"}) == before || !w.Exists("plants/b.json") || w.Exists("plants/c.json") {
		t.Fatal("the fingerprint follows the folder and Exists the files")
	}
	if absolute, err := Absolute("rel"); err != nil || !filepath.IsAbs(absolute) || w.Name() == "" {
		t.Fatalf("Absolute = %q, %v", absolute, err)
	}
}

func TestOpeningAMissingWorkspaceFails(t *testing.T) {
	t.Parallel()
	if _, err := Open(filepath.Join(t.TempDir(), "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("a missing workspace opened: %v", err)
	}
}

func TestAWriteLargerThanTheReadLimitIsRefused(t *testing.T) {
	t.Parallel()
	w, _ := workspace(t)
	if _, err := w.WriteIfUnchanged("plants/big.json", []byte("0123456789"), "", domain.Version, 5); !errors.Is(err, domain.ErrTooLarge) {
		t.Fatalf("a write over the limit: %v", err)
	}
	if w.Exists("plants/big.json") {
		t.Fatal("the refused write left a file")
	}
}

func TestTheTailIsTheLastBytesOfTheFile(t *testing.T) {
	t.Parallel()
	w, _ := workspace(t)
	if _, err := w.WriteIfUnchanged("runs/w-1/journal.jsonl", []byte("line one\nline two\n"), "", domain.Version, 100); err != nil {
		t.Fatal(err)
	}
	for limit, want := range map[int64]string{9: "line two\n", 100: "line one\nline two\n"} {
		if got, err := w.Tail("runs/w-1/journal.jsonl", limit); err != nil || string(got) != want {
			t.Errorf("Tail(%d) = %q, %v; want %q", limit, got, err, want)
		}
	}
	if _, err := w.Tail("runs/w-2/journal.jsonl", 9); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("the tail of an absent journal: %v", err)
	}
}

func TestFailuresAreNamedWithTheOperationThatFailed(t *testing.T) {
	t.Parallel()
	w, dir := workspace(t)
	if err := os.WriteFile(filepath.Join(dir, "plants", "a.json"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plants", ".b.json.tmp-"+strconv.Itoa(os.Getpid())), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cases := map[string]error{
		"canvasbridge: open plants/a.json/x":       second(w.Read("plants/a.json/x", 10)),
		"canvasbridge: read plants":                second(w.Read("plants", 10)),
		"canvasbridge: read runs":                  second(w.ReadFrom("runs", 0, 10)),
		"canvasbridge: read runs/w-1":              second(w.Tail("runs/w-1", 10)),
		"canvasbridge: list plants/a.json":         second(w.Names("plants/a.json")),
		"canvasbridge: write plants/.b.json.tmp-":  second(w.WriteIfUnchanged("plants/b.json", []byte("x"), "", domain.Version, 10)),
		"canvasbridge: create plants/a.json/inner": w.Ensure([]string{"plants/a.json/inner"}),
	}
	for want, err := range cases {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want one naming %q", err, want)
		}
	}
}

func second[T any](_ T, err error) error { return err }
