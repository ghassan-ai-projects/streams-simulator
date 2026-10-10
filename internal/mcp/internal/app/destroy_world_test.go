package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A run whose End failed is finished all the same. Destroying its world must
// leave the world destroyable: the retry may not fail with "already finished".
func TestDestroyWorldAfterAFailedEndCanBeRetried(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	// Replace the world's artifact directory with a file: End cannot write.
	dir := filepath.Join(d.OutDir, worldID)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := d.DestroyWorld(worldID)
	var toolErr *ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != CodeDomainInvalid {
		t.Fatalf("err = %v, want a %s tool error for the failed End", err, CodeDomainInvalid)
	}
	if _, err := d.DestroyWorld(worldID); err != nil {
		t.Fatalf("the retry must destroy the world: %v", err)
	}
	if d.World(worldID) != nil {
		t.Fatal("the world is still registered")
	}
}
