package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

const (
	policyDefault = "settings/default.policies.json"
	policyBody    = `{"policies":["sha256:1111111111111111111111111111111111111111111111111111111111111111"]}`
)

func TestTheWorkspacePolicyDefaultIsAbsentUntilWrittenAndThenReadWithItsEtag(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if _, err := h.bridge.ReadFile(policyDefault); !errors.Is(err, domain.ErrAbsent) {
		t.Fatalf("an absent default: %v, want ErrAbsent", err)
	}
	written, err := h.bridge.WriteFile(policyDefault, policyBody, "")
	if err != nil || written.Etag != domain.Version([]byte(policyBody)) {
		t.Fatalf("WriteFile = %+v, %v", written, err)
	}
	read, err := h.bridge.ReadFile(policyDefault)
	if err != nil || read.Content != policyBody || read.Etag != written.Etag {
		t.Fatalf("ReadFile = %+v, %v", read, err)
	}
	info, err := os.Stat(filepath.Join(h.dir, "settings", "default.policies.json"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("the default is owner-only: %v, %v", info, err)
	}
}

func TestAStaleOrMissingEtagNeverReplacesTheWorkspacePolicyDefault(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	first, _ := h.bridge.WriteFile(policyDefault, policyBody, "")
	for _, etag := range []string{"", "stale"} {
		if _, err := h.bridge.WriteFile(policyDefault, `{"policies":[]}`, etag); !errors.Is(err, domain.ErrConflict) {
			t.Errorf("write with etag %q: %v, want ErrConflict", etag, err)
		}
	}
	if _, err := h.bridge.WriteFile(policyDefault, `{"policies":[]}`, first.Etag); err != nil {
		t.Fatalf("a write from the current version: %v", err)
	}
}

func TestThePolicyDefaultIsCappedLikeEveryWorkspaceFile(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	big := strings.Repeat("x", domain.MaxFileBytes+1)
	if _, err := h.bridge.WriteFile(policyDefault, big, ""); !errors.Is(err, domain.ErrTooLarge) {
		t.Fatalf("an oversized default: %v, want ErrTooLarge", err)
	}
	if _, err := os.Stat(filepath.Join(h.dir, "settings", "default.policies.json")); !os.IsNotExist(err) {
		t.Fatalf("a refused write left a file: %v", err)
	}
}

func TestNothingElseInSettingsAndNoPolicyFileInDefinitionFoldersIsReadOrWritten(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rels := []string{"settings/notes.json", "settings/other.policies.json", "kinds/default.policies.json", "settings/../kinds/x.kind.json"}
	for _, folder := range append(domain.DefinitionFolders(), domain.ComponentFolders()...) {
		rels = append(rels, folder+"/default.policies.json")
	}
	for _, rel := range rels {
		if _, err := h.bridge.WriteFile(rel, "{}", ""); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("WriteFile(%q) = %v, want ErrBadPath", rel, err)
		}
		if _, err := h.bridge.ReadFile(rel); !errors.Is(err, domain.ErrBadPath) {
			t.Errorf("ReadFile(%q) = %v, want ErrBadPath", rel, err)
		}
	}
}

func TestThePolicyDefaultIsNotListedInTheWorkspace(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.write(t, policyDefault, policyBody)
	listing, err := h.bridge.Workspace()
	if err != nil {
		t.Fatal(err)
	}
	if len(listing.Domains)+len(listing.Compositions)+len(listing.Scenarios)+len(listing.Layouts)+len(listing.Kinds)+len(listing.Components)+len(listing.Runs) != 0 {
		t.Fatalf("the default appears in the listing: %+v", listing)
	}
}
