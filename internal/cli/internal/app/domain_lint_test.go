package app

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/lint"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func pondWithTooTightARange(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.Domain(pondDomain))
	if err != nil {
		t.Fatal(err)
	}
	const declared = `"range": {"min": 0.0, "max": 14.0}`
	if !strings.Contains(string(raw), declared) {
		t.Fatalf("the pond no longer declares %s", declared)
	}
	return writeFile(t, "tight.domain.json", strings.Replace(string(raw), declared, `"range": {"min": 0.0, "max": 0.5}`, 1))
}

func TestDomainLintReportsANominalDayAndExitsZeroWithoutErrors(t *testing.T) {
	t.Parallel()
	got := invoke(t, "domain", "lint", testsupport.Domain(pondDomain), "--hours", "2", "--seed", "5").mustSucceed(t)
	doc := got.document(t, "domain_lint")
	if doc["domain"] != pondDomain || doc["hours"] != 2.0 || doc["seed"] != 5.0 || doc["samples"] != 121.0 || doc["errors"] != 0.0 {
		t.Fatalf("lint = %v, want the pond over 2 h with seed 5, 121 samples and no errors", doc)
	}
}

func TestDomainLintDefaultsAreTheLintModulesDefaults(t *testing.T) {
	t.Parallel()
	fs := newFlagSet("domain", io.Discard)
	flags := registerLintFlags(fs)
	if err := parseFlags(fs, nil); err != nil {
		t.Fatal(err)
	}
	if flags.hours != lint.DefaultHours || flags.seed != lint.DefaultSeed {
		t.Fatalf("flag defaults = %v h seed %d, want %v h seed %d", flags.hours, flags.seed, float64(lint.DefaultHours), lint.DefaultSeed)
	}
}

func TestDomainLintFlagsMayFollowOrPrecedeThePath(t *testing.T) {
	t.Parallel()
	path := testsupport.Domain(pondDomain)
	after := invoke(t, "domain", "lint", path, "--hours", "1").mustSucceed(t)
	before := invoke(t, "domain", "lint", "--hours", "1", path).mustSucceed(t)
	if after.stdout != before.stdout {
		t.Fatalf("flag position changed the document:\n%s\n%s", after.stdout, before.stdout)
	}
}

func TestDomainLintIsByteIdenticalForTheSameInputs(t *testing.T) {
	t.Parallel()
	args := []string{"domain", "lint", testsupport.Domain(pondDomain), "--hours", "1", "--seed", "9"}
	if first, second := invoke(t, args...).stdout, invoke(t, args...).stdout; first != second {
		t.Fatalf("two lints differ:\n%s\n%s", first, second)
	}
}

func TestDomainLintWithAnErrorFindingFailsWithTheReportAsResult(t *testing.T) {
	t.Parallel()
	got := invoke(t, "domain", "lint", pondWithTooTightARange(t), "--hours", "1")
	failure := got.failure(t)
	result, _ := failure["result"].(map[string]any)
	if got.code != 1 || failure["code"] != "domain_lint" || result["errors"] == 0.0 {
		t.Fatalf("exit %d failure %v, want exit 1 with domain_lint carrying the report", got.code, failure)
	}
	if errs := cliOutputSchema(t, "domain_lint").Validate(result); len(errs) > 0 {
		t.Fatalf("the failure result is not a lint report: %v", errs)
	}
}

func TestDomainLintOfAnInvalidDomainIsADomainIssue(t *testing.T) {
	t.Parallel()
	bad := writeFile(t, "bad.domain.json", `{"id":"X"}`)
	if failure := invoke(t, "domain", "lint", bad).failure(t); failure["code"] != "domain_invalid" || failure["source"] != bad {
		t.Fatalf("failure = %v, want domain_invalid naming %s", failure, bad)
	}
}
