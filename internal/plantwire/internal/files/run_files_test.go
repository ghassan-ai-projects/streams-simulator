package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
)

func TestARunFileIsReadWhenPresent(t *testing.T) {
	t.Parallel()
	dir := directoryWith(t, map[string]string{"journal.jsonl": "{}\n"})
	raw, present, err := ReadRunFile(dir, "journal.jsonl")
	if err != nil || !present || string(raw) != "{}\n" {
		t.Fatalf("ReadRunFile = %q, %v, %v, want the file's bytes", raw, present, err)
	}
}

func TestAnAbsentRunFileIsNotAnError(t *testing.T) {
	t.Parallel()
	raw, present, err := ReadRunFile(t.TempDir(), "journal.jsonl")
	if err != nil || present || raw != nil {
		t.Fatalf("ReadRunFile = %q, %v, %v, want nothing present and no error", raw, present, err)
	}
}

func TestARunFileThatCannotBeReadIsNamedInTheError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "journal.jsonl"), 0o700); err != nil {
		t.Fatal(err)
	}
	raw, present, err := ReadRunFile(dir, "journal.jsonl")
	if err == nil || present || raw != nil || !strings.Contains(err.Error(), "plantwire: read journal.jsonl") {
		t.Fatalf("ReadRunFile = %q, %v, %v, want a read error naming journal.jsonl", raw, present, err)
	}
}

func TestARunFileThatIsASymbolicLinkIsRefusedSoAFolderCannotPointTheReaderAtADevice(t *testing.T) {
	t.Parallel()
	dir, target := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "real.jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "real.jsonl"), filepath.Join(dir, "journal.jsonl")); err != nil {
		t.Fatal(err)
	}
	if _, present, err := ReadRunFile(dir, "journal.jsonl"); !errors.Is(err, kernel.ErrNotRegular) || present {
		t.Fatalf("ReadRunFile = %v, %v, want kernel.ErrNotRegular for a link", err, present)
	}
}

func TestARunFileOverItsLimitIsRefusedBeforeItIsHeldInMemory(t *testing.T) {
	t.Parallel()
	dir := directoryWith(t, map[string]string{"journal.jsonl": "0123456789"})
	if _, present, err := ReadRunFileUpTo(dir, "journal.jsonl", 4); !errors.Is(err, kernel.ErrTooLarge) || present {
		t.Fatalf("ReadRunFileUpTo = %v, %v, want kernel.ErrTooLarge", err, present)
	}
}
