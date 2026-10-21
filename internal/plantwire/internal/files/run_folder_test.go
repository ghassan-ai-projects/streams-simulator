package files

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTheRunFolderListingNamesTopAndSealedFilesAndNothingForAMissingFolder(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sealed"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"run.json", filepath.Join("sealed", "journal.jsonl")} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	names, err := RunFolderFiles(dir)
	if err != nil || !slices.Contains(names, "run.json") || !slices.Contains(names, "sealed/journal.jsonl") {
		t.Fatalf("RunFolderFiles = %v, %v, want run.json and sealed/journal.jsonl", names, err)
	}
	if names, err := RunFolderFiles(filepath.Join(dir, "absent")); err != nil || len(names) != 0 {
		t.Fatalf("a missing folder = %v, %v, want nothing and no error", names, err)
	}
}

func TestARunFolderThatCannotBeListedIsAnError(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "plain")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RunFolderFiles(file); err == nil || !strings.Contains(err.Error(), "list the run folder") {
		t.Fatalf("a file as the run folder = %v, want a listing error", err)
	}
}
