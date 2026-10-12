package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal/internal/domain"
)

func TestOpenWritesBothStreamsAndFlushesAtABoundary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	j, err := Open(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	j.ClockAdvanced(10, 10)
	j.Command(10, commandAt(10))
	if err := j.Flush(); err != nil {
		t.Fatal(err)
	}
	// Flushed at a boundary, before the run ends: a reader sees both lines.
	for _, name := range []string{domain.ObserverFile, domain.DirectorFile} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || strings.Count(string(raw), "\n") != 1 {
			t.Fatalf("%s after flush = %q (%v)", name, raw, err)
		}
	}
	digests, err := j.Close(20, false, "")
	if err != nil {
		t.Fatal(err)
	}
	observer, _ := os.ReadFile(filepath.Join(dir, domain.ObserverFile))
	director, _ := os.ReadFile(filepath.Join(dir, domain.DirectorFile))
	if domain.Digest(observer) != digests.Observer || domain.Digest(director) != digests.Director {
		t.Fatal("the digests must be those of the files")
	}
}

func TestOpenNamesAJournalDirectoryItCannotCreate(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(filepath.Join(blocker, "run"), false); err == nil || !strings.Contains(err.Error(), "journal: directory") {
		t.Fatalf("Open under a file = %v", err)
	}
}

func TestOpenReleasesTheObserverFileWhenTheDirectorFileCannotOpen(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sealed"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir, false); err == nil || !strings.Contains(err.Error(), "journal: directory") {
		t.Fatalf("Open with sealed/ blocked = %v", err)
	}
}

func TestOpenRefusesAJournalFileItCannotCreate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, domain.ObserverFile), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(dir, false); err == nil || !strings.Contains(err.Error(), "journal: open") {
		t.Fatalf("Open over a directory = %v", err)
	}
}

func TestClosingAJournalWhoseFilesWereClosedUnderItReportsTheFailure(t *testing.T) {
	t.Parallel()
	s, err := openSink(filepath.Join(t.TempDir(), "j.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Write([]byte("x\n")); err != nil {
		t.Fatalf("a buffered write cannot fail yet: %v", err)
	}
	if err := s.Close(); err == nil || !strings.Contains(err.Error(), "close ") {
		t.Fatalf("Close over a closed file = %v", err)
	}
}
