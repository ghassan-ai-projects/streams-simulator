package app

import (
	"context"
	"slices"
	"testing"
)

func TestDomainDiffPassesARunFolderAndAWorkspaceDomainToTheBinaryAndNothingElse(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command.Stdout = []byte(`{"schema_version":"0.1"}`)
	args := map[string]any{"a": "w-1", "b": inDomains("fan.domain.json")}
	if _, err := h.bridge.RunAction(context.Background(), "domain.diff", args); err != nil {
		t.Fatal(err)
	}
	want := []string{"domain", "diff", h.dir + "/runs/w-1", h.dir + "/" + inDomains("fan.domain.json")}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
}
