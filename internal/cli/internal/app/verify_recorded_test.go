package app

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func place(t *testing.T, path string, content []byte) {
	t.Helper()
	// #nosec G703 -- the path is a file in the test's own temporary directory
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}

func runInEditableDomains(t *testing.T) (domainsDir, artifact string) {
	t.Helper()
	domainsDir, out := t.TempDir(), t.TempDir()
	raw, err := os.ReadFile(testsupport.Domain(pondDomain))
	if err != nil {
		t.Fatal(err)
	}
	place(t, filepath.Join(domainsDir, pondDomain+".domain.json"), raw)
	invoke(t, "run", "--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain,
		"--seed", "7", "--duration", "600", "--sink", "file", "--out", out).mustSucceed(t)
	return domainsDir, filepath.Join(out, "run.json")
}

func TestVerifyReplaysTheRecordedDomainAndReportsAnEditedOneAsInformation(t *testing.T) {
	t.Parallel()
	domainsDir, artifact := runInEditableDomains(t)
	inputs := []string{"--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir()}
	unchanged := invoke(t, append([]string{"verify", artifact}, inputs...)...).mustSucceed(t).document(t, "replay")
	if _, present := unchanged["domain_changed"]; present || unchanged["matches"] != true {
		t.Fatalf("unchanged domain: %v", unchanged)
	}
	path := filepath.Join(domainsDir, pondDomain+".domain.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), `"initial": 6.5,`, `"initial": 6.4,`, 1)
	place(t, path, []byte(edited))
	got := invoke(t, append([]string{"verify", artifact}, inputs...)...).mustSucceed(t).document(t, "replay")
	if got["matches"] != true || got["domain_changed"] != true || !strings.HasPrefix(got["current_domain_digest"].(string), "sha256:") {
		t.Fatalf("edited domain: %v", got)
	}
}

func TestVerifyOfATamperedArtifactFailsWithTheReplayAsItsResult(t *testing.T) {
	t.Parallel()
	domainsDir, artifact := runInEditableDomains(t)
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	zeros := []byte("${1}" + strings.Repeat("0", 64))
	tampered := regexp.MustCompile(`("expected_trace_digest": "sha256:)[0-9a-f]{64}`).ReplaceAll(raw, zeros)
	place(t, artifact, tampered)
	got := invoke(t, "verify", artifact, "--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir())
	failure := got.json(t)["error"].(map[string]any)
	result, _ := failure["result"].(map[string]any)
	if got.code != 1 || failure["code"] != "replay_mismatch" || result["matches"] != false {
		t.Fatalf("exit %d, %v; want replay_mismatch carrying matches:false", got.code, failure)
	}
}
