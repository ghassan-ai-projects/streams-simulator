package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// cliOutputSchema compiles cli-output-v0.1 itself (def "") or one of its
// $defs entries as a stand-alone schema, so one command's document is checked
// against its own shape and not merely against "any CLI document".
func cliOutputSchema(t *testing.T, def string) *jsonschema.Schema {
	t.Helper()
	raw, err := os.ReadFile(testsupport.Contract("cli-output-v0.1"))
	if err != nil {
		t.Fatal(err)
	}
	if def != "" {
		raw = entrySchema(t, raw, def)
	}
	schema, err := jsonschema.CompileJSON(raw)
	if err != nil {
		t.Fatalf("compile cli-output %q: %v", def, err)
	}
	return schema
}

func entrySchema(t *testing.T, raw []byte, def string) []byte {
	t.Helper()
	var contract map[string]any
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	entry, err := json.Marshal(map[string]any{"$defs": contract["$defs"], "$ref": "#/$defs/" + def})
	if err != nil {
		t.Fatal(err)
	}
	return entry
}

// document checks that stdout is exactly one JSON document matching the
// contract entry def and returns it.
func (o outcome) document(t *testing.T, def string) map[string]any {
	t.Helper()
	var doc any
	if err := model.DecodeBytes([]byte(o.stdout), &doc); err != nil {
		t.Fatalf("stdout is not exactly one JSON document: %v\n%s", err, o.stdout)
	}
	for _, schema := range []string{def, ""} {
		if errs := cliOutputSchema(t, schema).Validate(doc); len(errs) > 0 {
			t.Fatalf("stdout does not match cli-output %q:\n%s\n%s", schema, jsonschema.FormatErrors(errs), o.stdout)
		}
	}
	return o.json(t)
}

// failure checks that the command failed with a contract-valid error
// document and returns its body.
func (o outcome) failure(t *testing.T) map[string]any {
	t.Helper()
	body, _ := o.document(t, "error")["error"].(map[string]any)
	return body
}

func TestEveryCatalogAndAdapterDocumentMatchesTheContract(t *testing.T) {
	t.Parallel()
	domains, adapters := []string{"--domains-dir", testsupport.DomainsDir()}, []string{"--adapters-dir", testsupport.AdaptersDir()}
	cases := []struct {
		def  string
		args []string
	}{
		{"catalog_list", append([]string{"catalog", "list"}, domains...)},
		{"catalog_describe", append([]string{"catalog", "describe"}, append(domains, pondDomain)...)},
		{"catalog_coverage", append([]string{"catalog", "coverage"}, domains...)},
		{"domain_validate", []string{"domain", "validate", testsupport.Domain(pondDomain)}},
		{"domain_lint", []string{"domain", "lint", testsupport.Domain(pondDomain), "--hours", "2"}},
		{"adapter_list", append([]string{"adapter", "list"}, adapters...)},
		{"adapter_verify", append([]string{"adapter", "verify"}, append(adapters, testsupport.Adapter("native-jsonl"))...)},
	}
	for _, tc := range cases {
		t.Run(tc.def, func(t *testing.T) {
			t.Parallel()
			invoke(t, tc.args...).mustSucceed(t).document(t, tc.def)
		})
	}
}

func TestEveryRunWorkflowDocumentMatchesTheContract(t *testing.T) {
	t.Parallel()
	dir := runPondDocument(t)
	inputs := []string{"--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir()}
	artifact := filepath.Join(dir, "run.json")
	invoke(t, append(append([]string{"replay"}, inputs...), artifact)...).mustSucceed(t).document(t, "replay")
	invoke(t, append(append([]string{"verify"}, inputs...), artifact)...).mustSucceed(t).document(t, "replay")
	verdict := filepath.Join(dir, "verdict.json")
	invoke(t, "refconsumer", "--trace", filepath.Join(dir, "trace.jsonl"), "--out", verdict).mustSucceed(t).document(t, "refconsumer")
	label := writeNominalLabel(t, dir)
	invoke(t, append([]string{"score", "--run", artifact, "--label", label}, inputs...)...).mustSucceed(t).document(t, "score")
	suite := []string{"suite", "--domains-dir", testsupport.DomainsDir(), "--domain", suiteDomain, "--n", "1", "--out", t.TempDir()}
	invoke(t, suite...).mustSucceed(t).document(t, "suite")
	manifest := []string{"manifest", "--out", filepath.Join(t.TempDir(), "m.json"), "--author", "A <a@x>", "--reviewer", "R <r@x>"}
	invoke(t, append(manifest, inputs...)...).mustSucceed(t).document(t, "manifest")
}

// runPondDocument runs the pond briefly, checks the run document against the
// contract and returns the output directory.
func runPondDocument(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", pondDomain,
		"--seed", "7", "--duration", "900", "--sink", "file", "--out", dir).mustSucceed(t).document(t, "run")
	return dir
}

// nominalPondLabel is a contract-valid ground-truth record for a nominal pond
// scenario: negative class, no episode expected.
const nominalPondLabel = `{"scenario_id":"aquaculture-pond/0001","domain":"aquaculture-pond","seed":7,"entity_id":"",` +
	`"label":"nominal","is_negative_class":true,"expected_episode":false,"injection_time_ns":0,` +
	`"first_observable_time_ns":0,"unavoidable_time_ns":0,"observability":{"detector_form":"single_channel_snr",` +
	`"channels":["pond.dissolved_oxygen"],"effective_sigma":0.1,"first_observable_snr":3,"unavoidable_snr":10},` +
	`"trivial_baseline_verdict":"non_trivial"}`

// writeNominalLabel writes nominalPondLabel into dir.
func writeNominalLabel(t *testing.T, dir string) string {
	t.Helper()
	label := filepath.Join(dir, "label.json")
	if err := os.WriteFile(label, []byte(nominalPondLabel), 0o600); err != nil {
		t.Fatal(err)
	}
	return label
}

// writeFile writes content to name under a fresh temporary directory.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	testsupport.WriteFileIn(t, dir, name, []byte(content))
	return filepath.Join(dir, name)
}

func TestAContractViolationListsEveryIssueWithItsPointer(t *testing.T) {
	t.Parallel()
	bad := writeFile(t, "bad.domain.json", `{"id":"X"}`)
	got := invoke(t, "domain", "validate", bad)
	if got.code != 1 {
		t.Fatalf("exit %d, want 1: %+v", got.code, got)
	}
	failure := got.failure(t)
	if failure["code"] != "domain_invalid" || failure["source"] != bad {
		t.Fatalf("failure = %v, want domain_invalid naming %s", failure, bad)
	}
	issues, _ := failure["issues"].([]any)
	pointers := map[string]bool{}
	for _, issue := range issues {
		entry, _ := issue.(map[string]any)
		pointer, located := entry["pointer"].(string)
		if !located {
			t.Fatalf("schema issue without a pointer: %v", entry)
		}
		pointers[pointer] = true
	}
	if !pointers["/id"] || !pointers["/axes"] {
		t.Fatalf("issues = %v, want the /id pattern violation and the root's missing properties", issues)
	}
	if message := failure["message"].(string); strings.Contains(message, "\n") {
		t.Fatalf("message %q must be one line; the issues carry the details", message)
	}
}

func TestACrossReferenceViolationIsADomainIssueAtItsPointer(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.Domain(pondDomain))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	kinds, _ := doc["kinds"].([]any)
	kind, _ := kinds[0].(map[string]any)
	faults, _ := kind["faults"].([]any)
	first, _ := faults[0].(map[string]any)
	first["expected_effector"] = "no_such_effector"
	broken, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	failure := invoke(t, "domain", "validate", writeFile(t, "x.domain.json", string(broken))).failure(t)
	issues, _ := failure["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if failure["code"] != "domain_invalid" || len(issues) != 1 || issue["pointer"] != "/kinds/0/faults/0/expected_effector" || issue["keyword"] != "domain_rule" ||
		!strings.Contains(issue["message"].(string), "no_such_effector") {
		t.Fatalf("failure = %v, want one domain_invalid issue at the expected_effector member naming the effector", failure)
	}
}
