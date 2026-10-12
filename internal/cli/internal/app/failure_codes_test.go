package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryFailureClassHasItsStableCode(t *testing.T) {
	t.Parallel()
	inputs := []string{"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}
	missing := filepath.Join(t.TempDir(), "absent.json")
	corrupt := writeFile(t, "run.json", "{")
	cases := []struct {
		name, code string
		args       []string
	}{
		{"missing artifact", "file_not_found", []string{"replay", missing}},
		{"unknown domain id", "domain_not_found", []string{"catalog", "describe", "--domains-dir", testsupport.DomainsDir(), "no-such"}},
		{"unknown adapter id", "adapter_not_found", append([]string{"run", "--domain", pondDomain, "--adapter", "no-such", "--duration", "1"}, inputs...)},
		{"invalid adapter", "adapter_invalid", []string{"adapter", "verify", writeFile(t, "a.adapter.json", `{"id":"x"}`)}},
		{"corrupt artifact", "artifact_invalid", append([]string{"replay"}, append(inputs, corrupt)...)},
		{"unreadable trace", "io_failed", []string{"refconsumer", "--trace", t.TempDir()}},
		{"nothing to verify", "adapter_verify_failed", []string{"adapter", "verify", adapterWithoutConformance(t)}},
		{"empty catalog", "catalog_failed", []string{"catalog", "list", "--domains-dir", t.TempDir()}},
		{"refused effector", "run_failed", append([]string{"run", "--domain", pondDomain, "--duration", "60", "--effector", "no_such@no-entity@1"}, inputs...)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != 1 || !strings.HasPrefix(got.stderr, "streamsim: ") {
				t.Fatalf("exit %d stderr %q; want exit 1 with the reason on stderr", got.code, got.stderr)
			}
			if failure := got.failure(t); failure["code"] != tc.code {
				t.Fatalf("failure = %v, want code %s", failure, tc.code)
			}
		})
	}
}

func TestAFailedFileOperationNamesItsPath(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent.json")
	failure := invoke(t, "replay", missing).failure(t)
	if failure["source"] != missing || strings.HasPrefix(failure["message"].(string), "streamsim: ") {
		t.Fatalf("failure = %v, want source %s and a message without the streamsim prefix", failure, missing)
	}
}

func TestACorruptScoreLabelIsAnInvalidArtifactNamingTheLabel(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	writeReferenceVerdict(t, dir)
	label := writeFile(t, "label.json", "not json")
	got := invoke(t, "score", "--run", filepath.Join(dir, "run.json"), "--label", label,
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir())
	if failure := got.failure(t); failure["code"] != "artifact_invalid" || failure["source"] != label {
		t.Fatalf("failure = %v, want artifact_invalid naming %s", failure, label)
	}
}

// adapterWithoutConformance is the native adapter with its conformance
// section removed, so verification has nothing to compare.
func adapterWithoutConformance(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	start := strings.Index(text, `"conformance"`)
	if start < 0 {
		t.Fatal("native adapter declares no conformance section")
	}
	stripped := strings.TrimRight(text[:start], " \n\t,") + "\n}\n"
	return writeFile(t, "plain.adapter.json", stripped)
}

func TestScoreRefusesAnArtifactWhoseReplayDoesNotReproduce(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	writeReferenceVerdict(t, dir)
	tamperTraceDigest(t, filepath.Join(dir, "run.json"))
	got := invoke(t, "score", "--run", filepath.Join(dir, "run.json"), "--label", writeNominalLabel(t, dir),
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir())
	failure := got.failure(t)
	result, _ := failure["result"].(map[string]any)
	if got.code != 1 || failure["code"] != "replay_mismatch" || result["matches"] != false {
		t.Fatalf("exit %d failure %v; want replay_mismatch carrying the replay report", got.code, failure)
	}
}

// tamperTraceDigest replaces the artifact's expected trace digest so a replay
// can no longer reproduce it.
func tamperTraceDigest(t *testing.T, artifact string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	doc["expected_trace_digest"] = "sha256:" + strings.Repeat("0", 64)
	tampered, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
}

// writeReferenceVerdict runs the reference consumer over the run's trace and
// writes its contract-valid verdict as dir/verdict.json.
func writeReferenceVerdict(t *testing.T, dir string) {
	t.Helper()
	invoke(t, "refconsumer", "--trace", filepath.Join(dir, "trace.jsonl"), "--out", filepath.Join(dir, "verdict.json")).mustSucceed(t)
}

func TestScoreChecksItsInputsAgainstTheirContracts(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	if err := os.WriteFile(filepath.Join(dir, "verdict.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got := invoke(t, "score", "--run", filepath.Join(dir, "run.json"), "--label", writeNominalLabel(t, dir),
		"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir())
	failure := got.failure(t)
	if failure["code"] != "artifact_invalid" || failure["source"] != filepath.Join(dir, "verdict.json") {
		t.Fatalf("failure = %v, want artifact_invalid naming the verdict", failure)
	}
}

func TestScoreNamesAMissingRunArtifactFirst(t *testing.T) {
	t.Parallel()
	artifact := filepath.Join(t.TempDir(), "run.json")
	failure := invoke(t, "score", "--run", artifact, "--label", filepath.Join(t.TempDir(), "label.json")).failure(t)
	if failure["code"] != "file_not_found" || failure["source"] != artifact {
		t.Fatalf("failure = %v, want file_not_found naming %s", failure, artifact)
	}
}
