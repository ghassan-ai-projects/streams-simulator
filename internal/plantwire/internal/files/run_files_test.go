package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
