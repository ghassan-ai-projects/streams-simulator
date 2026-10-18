package app

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/process"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func invokeWithInput(t *testing.T, input string, args ...string) outcome {
	t.Helper()
	var stdout, stderr bytes.Buffer
	env := process.Env{Stdin: strings.NewReader(input), Stdout: &stdout, Stderr: &stderr, Now: func() time.Time { return fixedNow }}
	code := Run(append([]string{"streamsim"}, args...), Build{Version: "v-test", Commit: "c-test"}, env)
	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func pumpKindText(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

const noArgs = `{"type":"object","properties":{},"required":[],"additionalProperties":false}`

const incompleteKind = `{"kind":"probe","metrics":[
 {"name":"temperature","data_type":"number","unit":"Cel","eu_range":{"min":0,"max":50},"writable":false},
 {"name":"humidity","data_type":"number","unit":"%","writable":false},
 {"name":"level","data_type":"number","resolution":1,"writable":false}],
 "commands":[{"name":"halt","args_schema":` + noArgs + `,"risk_class":"low"}]}`

var incompleteIssues = []any{
	map[string]any{"pointer": "/metrics/0/resolution", "message": `needs "resolution": the interface a twin reproduces always states it`, "keyword": "twin_member_required", "member": "resolution"},
	map[string]any{"pointer": "/metrics/1/resolution", "message": `needs "resolution": the interface a twin reproduces always states it`, "keyword": "twin_member_required", "member": "resolution"},
	map[string]any{"pointer": "/metrics/2/unit", "message": `metric "level" is a number or integer: it states its unit`, "keyword": "twin_unit_required", "member": "unit"},
	map[string]any{"pointer": "/commands/0/safe_stop", "message": `needs "safe_stop": the interface a twin reproduces always states it`, "keyword": "twin_member_required", "member": "safe_stop"},
}

func TestKindValidateStdinChecksOneDocumentLikeAFileAndNamesItStdin(t *testing.T) {
	t.Parallel()
	doc := invokeWithInput(t, pumpKindText(t), "kind", "validate", "--stdin").mustSucceed(t).document(t, "kind_validate")
	kinds, _ := doc["kinds"].([]any)
	pump, _ := kinds[0].(map[string]any)
	file := invoke(t, "kind", "validate", testsupport.PumpKindFile()).mustSucceed(t).document(t, "kind_validate")
	fileKinds, _ := file["kinds"].([]any)
	fromFile, _ := fileKinds[0].(map[string]any)
	if len(kinds) != 1 || pump["path"] != "<stdin>" || pump["kind_digest"] != fromFile["kind_digest"] || pump["metrics"] != float64(2) {
		t.Fatalf("kind validate --stdin = %v, want the pump kind file read from <stdin> with the digest of the file", doc)
	}
}

func TestKindValidateStdinNeedsExactlyOneSource(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{"stdin and a file", []string{"kind", "validate", "--stdin", testsupport.PumpKindFile()}},
		{"neither", []string{"kind", "validate", "--twin"}},
		{"an unknown flag", []string{"kind", "validate", "--stdin", "--lenient"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invokeWithInput(t, pumpKindText(t), tc.args...)
			if body := got.failure(t); got.code != 2 || body["code"] != "usage_invalid" {
				t.Fatalf("exit %d with %v, want a usage refusal", got.code, body)
			}
		})
	}
}

func TestKindValidateStdinIsStrictAndBounded(t *testing.T) {
	t.Parallel()
	pump := pumpKindText(t)
	cases := []struct {
		name, input, message string
	}{
		{"not JSON", "{", ""},
		{"empty", "", ""},
		{"two documents", pump + "{}", ""},
		{"a member named twice", `{"kind": "k", "kind": "j", "metrics": [], "commands": []}`, ""},
		{"a document over 1 MiB", pump + strings.Repeat(" ", 1<<20), "<stdin> holds more than 1048576 bytes"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invokeWithInput(t, tc.input, "kind", "validate", "--stdin")
			body := got.failure(t)
			if got.code != 1 || body["code"] != "kind_failed" || !strings.Contains(body["message"].(string), tc.message) {
				t.Fatalf("exit %d with %v, want kind_failed mentioning %q", got.code, body, tc.message)
			}
		})
	}
}

func TestKindValidateTwinReportsEveryMissingMemberAtOnceFromStdinAndFromAFile(t *testing.T) {
	t.Parallel()
	fromStdin := invokeWithInput(t, incompleteKind, "kind", "validate", "--stdin", "--twin")
	path := writeFile(t, "probe.kind.json", incompleteKind)
	fromFile := invoke(t, "kind", "validate", "--twin", path)
	for name, got := range map[string]outcome{"stdin": fromStdin, "file": fromFile} {
		body := got.failure(t)
		if got.code != 1 || body["code"] != "kind_failed" || !reflect.DeepEqual(body["issues"], incompleteIssues) {
			t.Fatalf("%s: exit %d with %v, want kind_failed with the four issues of a twin at once", name, got.code, body)
		}
	}
	if body := fromStdin.failure(t); body["source"] != "<stdin>" || fromFile.failure(t)["source"] != path {
		t.Fatalf("sources = %v and %v, want <stdin> and the file", body["source"], fromFile.failure(t)["source"])
	}
}

func TestKindValidateWithoutTwinAcceptsADocumentOnlyATwinWouldRefuse(t *testing.T) {
	t.Parallel()
	invokeWithInput(t, incompleteKind, "kind", "validate", "--stdin").mustSucceed(t).document(t, "kind_validate")
}

func TestKindValidateTwinAcceptsTheThermalKindFilesAndTheirSuggestedChoicesScaffoldATwin(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(testsupport.ThermalBoardKindsDir(), "*.kind.json"))
	if err != nil || len(paths) != 3 {
		t.Fatalf("thermal kind files = %v, %v", paths, err)
	}
	invoke(t, append([]string{"kind", "validate", "--twin"}, paths...)...).mustSucceed(t).document(t, "kind_validate")
	suggested := invoke(t, append([]string{"kind", "twin", "--suggest"}, paths...)...).mustSucceed(t).document(t, "kind_suggest")
	choices := writeFile(t, "suggested.twin.json", string(mustJSON(t, suggested["choices"])))
	twin := invoke(t, "kind", "twin", "--choices", choices, "--kinds-dir", testsupport.ThermalBoardKindsDir()).mustSucceed(t).document(t, "kind_twin")
	if implements, _ := twin["implements"].([]any); len(implements) != 3 {
		t.Fatalf("kind twin of the suggested choices = %v, want the three kinds implemented", twin)
	}
}
