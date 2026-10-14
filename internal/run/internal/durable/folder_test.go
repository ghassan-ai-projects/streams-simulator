package durable

import (
	"path/filepath"
	"testing"
)

func TestAFolderIsUsedOnceItExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if FolderUsed(filepath.Join(dir, "w-1")) || !FolderUsed(dir) {
		t.Fatal("a missing folder is free and an existing one is used")
	}
}
