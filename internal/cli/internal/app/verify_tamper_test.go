package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func appendToFile(t *testing.T, path, tail string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	place(t, path, append(raw, tail...))
}

func editEmbeddedSpec(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), `"initial": 6.5,`, `"initial": 6.4,`, 1)
	if edited == string(raw) {
		t.Fatal("run.json embeds no initial 6.5 to edit")
	}
	place(t, artifact, []byte(edited))
}

func TestVerifyReportsATamperedRunAsAVerdictThatNamesTheReason(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		tamper func(t *testing.T, artifact string)
		detail string
	}{
		{"embedded spec edited", editEmbeddedSpec, "the recorded domain spec does not match its digest"},
		{"trace file with an appended newline", func(t *testing.T, artifact string) {
			appendToFile(t, filepath.Join(filepath.Dir(artifact), "trace.jsonl"), "\n")
		}, "trace.jsonl differs from the digest run.json recorded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			domainsDir, artifact := runInEditableDomains(t)
			tc.tamper(t, artifact)
			got := invoke(t, "verify", artifact, "--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir())
			failure := got.failure(t)
			result, _ := failure["result"].(map[string]any)
			detail, _ := result["detail"].(string)
			if got.code != 1 || failure["code"] != "replay_mismatch" || result["matches"] != false || !strings.Contains(detail, tc.detail) {
				t.Fatalf("exit %d, %v; want replay_mismatch whose result is matches:false with detail %q", got.code, failure, tc.detail)
			}
			replayed := invoke(t, "replay", artifact, "--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir()).mustSucceed(t).document(t, "replay")
			if replayed["matches"] != false || replayed["detail"] != detail {
				t.Fatalf("replay document %v; want the same verdict as verify", replayed)
			}
		})
	}
}

func TestVerifyOfAMissingRunStaysAFailureWithoutAResult(t *testing.T) {
	t.Parallel()
	got := invoke(t, "verify", filepath.Join(t.TempDir(), "run.json"))
	failure := got.failure(t)
	if _, hasResult := failure["result"]; got.code != 1 || hasResult || failure["code"] != "file_not_found" {
		t.Fatalf("exit %d, %v; want file_not_found without a result", got.code, failure)
	}
}
