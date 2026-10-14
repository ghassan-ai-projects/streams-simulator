package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func dropRecordedDigest(t *testing.T, artifact, field string) {
	t.Helper()
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	delete(doc, field)
	edited, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	place(t, artifact, edited)
}

func TestVerifyDocumentsNameTheReasonAndValidateAgainstTheContract(t *testing.T) {
	t.Parallel()
	inFolder := func(name string) func(artifact string) string {
		return func(artifact string) string { return filepath.Join(filepath.Dir(artifact), name) }
	}
	cases := []struct {
		name         string
		tamper       func(t *testing.T, artifact string)
		reason, file string
	}{
		{"clean run", func(*testing.T, string) {}, "match", ""},
		{"embedded spec edited", editEmbeddedSpec, "recorded_spec", ""},
		{"trace file with an appended newline", func(t *testing.T, artifact string) {
			appendToFile(t, inFolder("trace.jsonl")(artifact), "\n")
		}, "file_differs", "trace.jsonl"},
		{"ledger file edited", func(t *testing.T, artifact string) {
			appendToFile(t, inFolder("ledger.jsonl")(artifact), "\n")
		}, "file_differs", "ledger.jsonl"},
		{"history file deleted", func(t *testing.T, artifact string) {
			if err := os.Remove(inFolder("world_state_history.jsonl")(artifact)); err != nil {
				t.Fatal(err)
			}
		}, "file_missing", "world_state_history.jsonl"},
		{"ledger digest removed", func(t *testing.T, artifact string) {
			dropRecordedDigest(t, artifact, "ledger_digest")
		}, "digest_missing", "ledger.jsonl"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			domainsDir, artifact := runInEditableDomains(t)
			tc.tamper(t, artifact)
			got := invoke(t, "verify", artifact, "--domains-dir", domainsDir, "--adapters-dir", testsupport.AdaptersDir())
			var doc map[string]any
			if tc.reason == "match" {
				doc = got.mustSucceed(t).document(t, "replay")
			} else {
				doc, _ = got.failure(t)["result"].(map[string]any)
				if errs := cliOutputSchema(t, "replay").Validate(doc); len(errs) > 0 {
					t.Fatalf("the mismatch result does not match cli-output replay: %v\n%v", errs, doc)
				}
			}
			file, _ := doc["file"].(string)
			if doc["reason"] != tc.reason || file != tc.file {
				t.Fatalf("document %v; want reason %q and file %q", doc, tc.reason, tc.file)
			}
		})
	}
}

func TestVerifyAndReplayHelpNameTheirOwnCommand(t *testing.T) {
	t.Parallel()
	for _, command := range []string{"verify", "replay"} {
		got := invoke(t, command, "-h")
		if got.code != 0 || !strings.Contains(got.stderr, "Usage of "+command+":") {
			t.Fatalf("%s -h: exit %d stderr %q; want Usage of %s", command, got.code, got.stderr, command)
		}
	}
}
