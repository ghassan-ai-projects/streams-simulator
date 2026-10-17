package app

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twinArguments(t *testing.T, extra ...string) []string {
	t.Helper()
	return append([]string{"kind", "twin", "--choices", testsupport.ThermalBoardChoices(), "--kinds-dir", testsupport.ThermalBoardKindsDir()}, extra...)
}

func TestKindTwinWritesADomainThatImplementsTheKindFilesAndNamesThem(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "twin.domain.json")
	doc := invoke(t, twinArguments(t, "--out", out)...).mustSucceed(t).document(t, "kind_twin")
	implements, _ := doc["implements"].([]any)
	if doc["composed"] != out || doc["domain"] != nil || len(implements) != 3 || implements[0] != "thermal_sensor" || implements[2] != "fan" {
		t.Fatalf("kind twin = %v, want the file written and the three kinds named", doc)
	}
	invoke(t, "domain", "validate", out, "--kinds-dir", testsupport.ThermalBoardKindsDir()).mustSucceed(t).document(t, "domain_validate")
	invoke(t, "domain", "lint", out).mustSucceed(t).document(t, "domain_lint")
}

func TestKindTwinPrintsTheDomainWhenNoFileIsAskedFor(t *testing.T) {
	t.Parallel()
	doc := invoke(t, twinArguments(t)...).mustSucceed(t).document(t, "kind_twin")
	spec, _ := doc["domain"].(map[string]any)
	if doc["composed"] != nil || spec["id"] != "arduino-thermal-twin" {
		t.Fatalf("kind twin = %v, want the twin inside the document and no file", doc)
	}
}

func TestKindTwinTakesKindFilesAsArguments(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(testsupport.ThermalBoardKindsDir(), "*.kind.json"))
	if err != nil || len(paths) != 3 {
		t.Fatalf("kind files %v: %v", paths, err)
	}
	args := slices.Concat([]string{"kind", "twin", "--choices", testsupport.ThermalBoardChoices()}, paths)
	invoke(t, args...).mustSucceed(t).document(t, "kind_twin")
}

func TestKindTwinRefusesWhatItCannotScaffold(t *testing.T) {
	t.Parallel()
	absent := filepath.Join(t.TempDir(), "absent.twin.json")
	cases := []struct {
		name    string
		args    []string
		code    string
		status  int
		pointer string
	}{
		{"no choices", []string{"kind", "twin", "--kinds-dir", testsupport.ThermalBoardKindsDir()}, "usage_invalid", 2, ""},
		{"a choices file that does not exist", []string{"kind", "twin", "--choices", absent}, "file_not_found", 1, ""},
		{"a kind with no kind file", []string{"kind", "twin", "--choices", testsupport.ThermalBoardChoices()}, "kind_failed", 1, "/kinds/0/kind"},
		{"a kinds folder that does not exist", twinArguments(t, "--kinds-dir", filepath.Join(t.TempDir(), "absent")), "file_not_found", 1, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			body := got.failure(t)
			if got.code != tc.status || body["code"] != tc.code {
				t.Fatalf("exit %d with %v, want exit %d and code %s", got.code, body, tc.status, tc.code)
			}
			if tc.pointer == "" {
				return
			}
			issues, _ := body["issues"].([]any)
			issue, _ := issues[0].(map[string]any)
			if issue["pointer"] != tc.pointer || body["source"] != testsupport.ThermalBoardChoices() {
				t.Fatalf("failure = %v, want an issue at %s from the choices file", body, tc.pointer)
			}
		})
	}
}

func TestKindTwinRefusesChoicesThatBreakTheirSchemaAtTheMemberThatIsWrong(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.ThermalBoardChoices())
	if err != nil {
		t.Fatal(err)
	}
	broken := writeFile(t, "broken.twin.json", strings.Replace(string(raw), `"first_order"`, `"ramp"`, 1))
	got := invoke(t, "kind", "twin", "--choices", broken, "--kinds-dir", testsupport.ThermalBoardKindsDir())
	body := got.failure(t)
	issues, _ := body["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if got.code != 1 || body["code"] != "kind_failed" || issue["pointer"] != "/kinds/2/commands/set_pwm_lease/response" {
		t.Fatalf("exit %d with %v, want kind_failed at /kinds/2/commands/set_pwm_lease/response", got.code, body)
	}
}
