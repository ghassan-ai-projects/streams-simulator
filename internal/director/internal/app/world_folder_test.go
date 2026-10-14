package app

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestANewWorldNeverReusesTheFolderOfAnEarlierRun(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	for _, stale := range []string{"w-1", "w-2"} {
		if err := os.MkdirAll(filepath.Join(d.OutDir, stale), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d.OutDir, stale, "run.json"), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	worldID := createWorld(t, d)
	if worldID != "w-3" {
		t.Fatalf("the world took %q, want the first id whose folder is free (w-3)", worldID)
	}
	if _, err := os.Stat(filepath.Join(d.OutDir, "w-3", "run.json")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the new run folder already holds a finished run's artifact: %v", err)
	}
}
