package files

import (
	"errors"
	"os"
	"path/filepath"
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
	if _, err := w.Read("plants/link.json", 100); err == nil {
		t.Fatal("a symlink out of the workspace was followed")
	}
	if _, err := w.Read("../secret.json", 100); err == nil {
		t.Fatal("a path out of the workspace was read")
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
	if _, err := Open(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("a missing workspace opened")
	}
}
