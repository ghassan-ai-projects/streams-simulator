package acceptance

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestMainForwardsTheBuildIdentity(t *testing.T) {
	t.Parallel()
	root := testsupport.RepositoryRoot()
	bin := buildBinary(t, root, "-X main.Version=v-sub -X main.Commit=c-sub")
	manifest := filepath.Join(t.TempDir(), "manifest.json")
	// #nosec G204 -- the test runs the freshly built binary with fixed args
	cmd := exec.CommandContext(t.Context(), bin, "manifest", "--out", manifest,
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(),
		"--author", "A", "--reviewer", "R")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if !strings.Contains(string(out), `"sim": "v-sub@c-sub"`) {
		t.Fatalf("manifest output lacks the linked build identity:\n%s", out)
	}
}
