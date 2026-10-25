package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twinWithGain(t *testing.T, gain string) string {
	t.Helper()
	choices, err := os.ReadFile(testsupport.TwinThermalChoices())
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(choices), `"gain": -0.02`, `"gain": `+gain, 1)
	dir := t.TempDir()
	choicesPath := writeIn(t, dir, "thermal-board.twin.json", edited)
	out := filepath.Join(dir, "twin.domain.json")
	invoke(t, "kind", "twin", "--choices", choicesPath, "--kinds-dir", testsupport.TwinThermalKindsDir(), "--out", out).mustSucceed(t)
	return out
}

func writeIn(t *testing.T, dir, name, content string) string {
	t.Helper()
	testsupport.WriteFileIn(t, dir, name, []byte(content))
	return filepath.Join(dir, name)
}

func TestDomainDiffNamesTheChangedGainTheSilentEditAndIsTheSameEveryTime(t *testing.T) {
	t.Parallel()
	before, after := twinWithGain(t, "-0.02"), twinWithGain(t, "-0.03")
	first := invoke(t, "domain", "diff", before, after).mustSucceed(t)
	doc := first.document(t, "domain_diff")
	identity, _ := doc["identity"].(map[string]any)
	kinds, _ := doc["kinds"].([]any)
	kind, _ := kinds[0].(map[string]any)
	changes, _ := kind["changes"].([]any)
	change, _ := changes[0].(map[string]any)
	if identity["same_id_and_version_different_digest"] != true || len(kinds) != 1 || change["code"] != "gain_changed" || change["before"] != -0.02 || change["after"] != -0.03 {
		t.Fatalf("diff = %v, want a gain change under an unchanged version", doc)
	}
	if second := invoke(t, "domain", "diff", before, after).mustSucceed(t); second.stdout != first.stdout {
		t.Fatal("two runs printed different documents")
	}
}

func TestDomainDiffOfEqualDomainsAndOfARunFolderHasNoChange(t *testing.T) {
	t.Parallel()
	pond := testsupport.Domain("aquaculture-pond")
	doc := invoke(t, "domain", "diff", testsupport.ReportRun(), pond).mustSucceed(t).document(t, "domain_diff")
	summary, _ := doc["summary"].(map[string]any)
	sourceA, _ := doc["a"].(map[string]any)["source"].(map[string]any)
	if summary["changes"] != 0.0 || sourceA["kind"] != "run" {
		t.Fatalf("diff = %v, want the run's recorded domain equal to the file", doc)
	}
}

func TestDomainDiffPrintsMarkdownAsTheTextOfOneJSONDocument(t *testing.T) {
	t.Parallel()
	before, after := twinWithGain(t, "-0.02"), twinWithGain(t, "-0.03")
	doc := invoke(t, "domain", "diff", before, after, "--format", "markdown").mustSucceed(t).document(t, "report_markdown")
	if text, _ := doc["markdown"].(string); !strings.Contains(text, "`gain_changed`") || !strings.Contains(text, "# Domain diff") {
		t.Fatalf("markdown = %v, want the change as text", doc)
	}
}

func TestDomainDiffRefusesWhatItCannotCompare(t *testing.T) {
	t.Parallel()
	pond := testsupport.Domain("aquaculture-pond")
	empty := t.TempDir()
	for name, tc := range map[string]struct {
		args []string
		code string
	}{
		"a missing side":              {[]string{"domain", "diff", pond, filepath.Join(empty, "none.domain.json")}, "file_not_found"},
		"a folder without run.json":   {[]string{"domain", "diff", empty, pond}, "file_not_found"},
		"a domain that breaks schema": {[]string{"domain", "diff", writeFile(t, "bad.domain.json", `{"id":"X"}`), pond}, "domain_invalid"},
		"a run.json without a spec":   {[]string{"domain", "diff", runWithoutSpec(t), pond}, "artifact_invalid"},
		"one side only":               {[]string{"domain", "diff", pond}, "usage_invalid"},
		"three sides":                 {[]string{"domain", "diff", pond, pond, pond}, "usage_invalid"},
		"an unknown format":           {[]string{"domain", "diff", pond, pond, "--format", "html"}, "usage_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if failure := invoke(t, tc.args...).failure(t); failure["code"] != tc.code {
				t.Fatalf("failure = %v, want code %s", failure, tc.code)
			}
		})
	}
}

func runWithoutSpec(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.ReportRun(), "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(raw), `"domain_spec"`)
	if start < 0 {
		t.Fatal("the report run has no domain_spec")
	}
	dir := t.TempDir()
	writeIn(t, dir, "run.json", strings.Replace(string(raw), `"domain_spec"`, `"domain_spec_removed"`, 1))
	return dir
}
