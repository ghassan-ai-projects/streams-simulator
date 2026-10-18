package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pumpDomainImplementing(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	doc["kinds"].([]any)[0].(map[string]any)["implements"] = "pump"
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return writeFile(t, "pump-board.domain.json", string(out))
}

func pumpKindsDirectory(t *testing.T, edit func(raw string) string) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.PumpKindFile())
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(writeFile(t, "pump.kind.json", edit(string(raw))))
}

func asWritten(raw string) string { return raw }

func TestKindValidateReportsEachKindFileWithItsDigest(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "kind", "validate", testsupport.PumpKindFile()).mustSucceed(t).document(t, "kind_validate")
	kinds, _ := doc["kinds"].([]any)
	pump, _ := kinds[0].(map[string]any)
	if len(kinds) != 1 || pump["kind"] != "pump" || pump["path"] != testsupport.PumpKindFile() || pump["metrics"] != float64(2) || pump["commands"] != float64(3) {
		t.Fatalf("kind validate = %v, want the pump kind file with two metrics and three commands", doc)
	}
}

func TestKindValidateRefusesWhatIsNotAKindFile(t *testing.T) {
	t.Parallel()
	thin := writeFile(t, "thin.kind.json", `{"kind": "thin"}`)
	assertFailureCodes(t, []failureCase{
		{"no verb", []string{"kind"}, "usage_invalid", 2},
		{"an unknown verb", []string{"kind", "check", thin}, "usage_invalid", 2},
		{"no file", []string{"kind", "validate"}, "usage_invalid", 2},
		{"a kind file without commands", []string{"kind", "validate", thin}, "kind_failed", 1},
		{"a file that does not exist", []string{"kind", "validate", filepath.Join(t.TempDir(), "absent.kind.json")}, "file_not_found", 1},
	})
}

// failureCase is one invocation expected to fail with an exit status and a failure code.
type failureCase struct {
	name   string
	args   []string
	code   string
	status int
}

func assertFailureCodes(t *testing.T, cases []failureCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if body := got.failure(t); got.code != tc.status || body["code"] != tc.code {
				t.Fatalf("exit %d with %v, want exit %d and code %s", got.code, body, tc.status, tc.code)
			}
		})
	}
}

func TestDomainValidateChecksAKindAgainstTheKindFileItImplements(t *testing.T) {
	t.Parallel()
	domainFile := pumpDomainImplementing(t)
	drifted := func(raw string) string { return strings.Replace(raw, `"m3/h"`, `"l/s"`, 1) }
	cases := []struct {
		name string
		dirs []string
		want string
	}{
		{"a kind file that matches", []string{pumpKindsDirectory(t, asWritten)}, ""},
		{"a unit that drifted", []string{pumpKindsDirectory(t, drifted)}, "/metrics/0/unit"},
		{"no kinds directory", nil, "no kind file of that name was loaded"},
		{"a directory without the kind file", []string{t.TempDir()}, "no kind file of that name was loaded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := []string{"domain", "validate", domainFile}
			for _, dir := range tc.dirs {
				args = append(args, "--kinds-dir", dir)
			}
			got := invoke(t, args...)
			if tc.want == "" {
				got.mustSucceed(t).document(t, "domain_validate")
				return
			}
			assertImplementsRefusal(t, got, domainFile, tc.want)
		})
	}
}

func assertImplementsRefusal(t *testing.T, got outcome, source, want string) {
	t.Helper()
	body := got.failure(t)
	issues, _ := body["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	message, _ := issue["message"].(string)
	if got.code != 1 || body["code"] != "domain_invalid" || body["source"] != source || issue["pointer"] != "/kinds/0/implements" || !strings.Contains(message, want) {
		t.Fatalf("exit %d with %v, want domain_invalid from %s at /kinds/0/implements mentioning %q", got.code, body, source, want)
	}
}

func TestDomainValidateRefusesTheKindsOptionOutsideValidate(t *testing.T) {
	t.Parallel()
	got := invoke(t, "domain", "lint", pumpDomainImplementing(t), "--kinds-dir", t.TempDir())
	body := got.failure(t)
	if got.code != 2 || body["code"] != "usage_invalid" || body["message"] != "domain lint: --kinds-dir applies to validate only" {
		t.Fatalf("exit %d with %v, want a usage refusal of --kinds-dir on lint", got.code, body)
	}
}

func TestRunChecksTheKindFilesOfItsDomainBeforeAnyRun(t *testing.T) {
	t.Parallel()
	domains := filepath.Dir(pumpDomainImplementing(t))
	base := []string{"run", "--domains-dir", domains, "--adapters-dir", testsupport.AdaptersDir(), "--domain", "pump-board", "--duration", "600"}
	assertImplementsRefusal(t, invoke(t, base...), "pump-board", "no kind file of that name was loaded")
	invoke(t, append(base, "--kinds-dir", pumpKindsDirectory(t, asWritten))...).mustSucceed(t).document(t, "run")
}

func TestDomainValidateAndRunSayWhichKindFilesTheyCheckedAndAnEmptyListWhenThereWereNone(t *testing.T) {
	t.Parallel()
	domainFile, kinds := pumpDomainImplementing(t), pumpKindsDirectory(t, asWritten)
	checkedOf := func(doc map[string]any) []any {
		list, _ := doc["implements_checked"].([]any)
		return list
	}
	validated := invoke(t, "domain", "validate", domainFile, "--kinds-dir", kinds).mustSucceed(t).document(t, "domain_validate")
	entry, _ := checkedOf(validated)[0].(map[string]any)
	if len(checkedOf(validated)) != 1 || entry["kind"] != "pump" || !strings.HasPrefix(entry["kind_digest"].(string), "sha256:") {
		t.Fatalf("domain validate = %v, want the pump kind file reported with its digest", validated)
	}
	ran := invoke(t, "run", "--domains-dir", filepath.Dir(domainFile), "--adapters-dir", testsupport.AdaptersDir(), "--domain", "pump-board",
		"--duration", "600", "--kinds-dir", kinds).mustSucceed(t).document(t, "run")
	if reported := checkedOf(ran); len(reported) != 1 || !equalDocuments(reported[0], entry) {
		t.Fatalf("run = %v, want the same kind file reported as domain validate: %v", ran, entry)
	}
	plain := invoke(t, "domain", "validate", testsupport.Domain("aquaculture-pond")).mustSucceed(t).document(t, "domain_validate")
	if list, present := plain["implements_checked"].([]any); !present || len(list) != 0 {
		t.Fatalf("domain validate of a domain that implements nothing = %v, want implements_checked as an empty list", plain)
	}
}

func equalDocuments(a, b any) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}

func TestAMalformedKindFileBlocksOnlyTheDomainsThatImplementAndTheKindCommand(t *testing.T) {
	t.Parallel()
	broken := filepath.Dir(writeFile(t, "broken.kind.json", `{"kind": "thin"}`))
	pond := testsupport.Domain("aquaculture-pond")
	invoke(t, "domain", "validate", pond, "--kinds-dir", broken).mustSucceed(t).document(t, "domain_validate")
	invoke(t, "run", "--domains-dir", testsupport.DomainsDir(), "--adapters-dir", testsupport.AdaptersDir(), "--domain", "aquaculture-pond",
		"--duration", "600", "--kinds-dir", broken).mustSucceed(t).document(t, "run")
	if got := invoke(t, "domain", "validate", pumpDomainImplementing(t), "--kinds-dir", broken); got.code != 1 || got.failure(t)["code"] != "kind_failed" {
		t.Fatalf("a domain that implements a kind file, beside a malformed one: exit %d, want kind_failed", got.code)
	}
	if got := invoke(t, "kind", "validate", filepath.Join(broken, "broken.kind.json")); got.code != 1 || got.failure(t)["code"] != "kind_failed" {
		t.Fatalf("kind validate of the malformed file: exit %d, want kind_failed", got.code)
	}
}

func TestDomainComposeTakesTheKindsOption(t *testing.T) {
	t.Parallel()
	invoke(t, "domain", "compose", pilotComposition(), "--kinds-dir", t.TempDir()).mustSucceed(t).document(t, "domain_compose")
}
