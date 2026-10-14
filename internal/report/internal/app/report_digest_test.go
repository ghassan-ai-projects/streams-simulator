package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func editRunArtifact(t *testing.T, dir string, edit func(doc map[string]any)) {
	t.Helper()
	path := filepath.Join(dir, model.FileRunArtifact)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	edit(doc)
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestReportOfARunWithAnUnusableEvidenceDigestIsAFailingVerificationNotAnError(t *testing.T) {
	t.Parallel()
	schema, err := jsonschema.CompileJSON(schemas.RunReport())
	if err != nil {
		t.Fatal(err)
	}
	edits := map[string]func(doc map[string]any){
		"absent":    func(doc map[string]any) { delete(doc, "ledger_digest") },
		"empty":     func(doc map[string]any) { doc["ledger_digest"] = "" },
		"malformed": func(doc map[string]any) { doc["ledger_digest"] = "sha256:00" },
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := copyFolder(t, testsupport.ReportRun())
			editRunArtifact(t, dir, edit)
			built, err := Build(t.Context(), dir)
			if err != nil {
				t.Fatalf("a run.json that verify accepts must not be rejected by report: %v", err)
			}
			checks := built.Verification.Checks
			ledger := checks[len(checks)-2]
			if built.Verification.Status != domain.FilesDiffer || ledger.Name != "ledger" || ledger.Status != domain.CheckNoDigest || ledger.Recorded != "" {
				t.Fatalf("verification = %+v, want files_differ with a ledger check digest_missing", built.Verification)
			}
			var doc any
			if err := model.DecodeBytes(render(t, dir, JSONFormat), &doc); err != nil {
				t.Fatal(err)
			}
			if errs := schema.Validate(doc); len(errs) > 0 {
				t.Fatalf("report does not match run-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
			}
		})
	}
}
