package app

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
)

func TestDomainLineageReadsEveryWorkspaceFolderThatExistsAndNothingElse(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, fanChoices, "{}")
	h.command.Stdout = []byte(`{"schema_version":"0.1","plants":[],"device_kinds":[],"unreadable":[]}`)
	run := func(args map[string]any) []string {
		t.Helper()
		if _, err := h.bridge.RunAction(context.Background(), "domain.lineage", args); err != nil {
			t.Fatal(err)
		}
		return h.argv[len(h.argv)-1]
	}
	want := []string{"domain", "lineage", "--domains-dir", h.domainsDir(), "--kinds-dir", filepath.Join(h.dir, "kinds"),
		"--twins-dir", filepath.Join(h.dir, "twins"), "--compositions-dir", filepath.Join(h.dir, "compositions"), "--runs-dir", filepath.Join(h.dir, "runs"),
		"--components-dir", "/lib"}
	if got := run(nil); !slices.Equal(got, want) {
		t.Fatalf("argv = %v, want %v", got, want)
	}
	if got := run(map[string]any{"domain": "factory-utilities"}); !slices.Equal(got, append(slices.Clone(want[:2]), append([]string{"--domain", "factory-utilities"}, want[2:]...)...)) {
		t.Fatalf("argv with a domain = %v", got)
	}
}

func TestDomainAssumptionsNamesAWorkspaceDomainAndItsTwinChoices(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, fanKind, fanKindBytes)
	h.write(t, fanChoices, "{}")
	args := map[string]any{"path": inDomains("fan.domain.json"), "twin": fanChoices}
	if _, err := h.bridge.RunAction(context.Background(), "domain.assumptions", args); err != nil {
		t.Fatal(err)
	}
	want := []string{"domain", "assumptions", h.dir + "/" + inDomains("fan.domain.json"), "--twin", h.dir + "/" + fanChoices, "--components-dir", "/lib", "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
}

func TestDomainLineageSkipsTheFoldersTheWorkspaceDoesNotHave(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.RunAction(context.Background(), "domain.lineage", nil); err != nil {
		t.Fatal(err)
	}
	want := []string{"domain", "lineage", "--domains-dir", h.domainsDir(), "--compositions-dir", filepath.Join(h.dir, "compositions"),
		"--runs-dir", filepath.Join(h.dir, "runs"), "--components-dir", "/lib"}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv = %v, want %v", h.argv[0], want)
	}
}
