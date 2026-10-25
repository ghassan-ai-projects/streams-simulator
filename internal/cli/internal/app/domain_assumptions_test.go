package app

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func thermalTwinDomain(t *testing.T) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "twin.domain.json")
	invoke(t, twinArguments(t, "--out", out)...).mustSucceed(t)
	return out
}

func TestDomainAssumptionsOfATwinNameEachChoiceAndWhatTheTwinDoesNotModel(t *testing.T) {
	t.Parallel()
	args := []string{"domain", "assumptions", thermalTwinDomain(t), "--twin", testsupport.TwinThermalChoices(), "--kinds-dir", testsupport.TwinThermalKindsDir()}
	doc := invoke(t, args...).mustSucceed(t).document(t, "domain_assumptions")
	origin, _ := doc["origin"].(map[string]any)
	notModeled, _ := doc["not_modeled"].([]any)
	if origin["kind"] != "twin" || len(notModeled) < 5 {
		t.Fatalf("assumptions = %v, want a twin origin and the limits of the model", doc)
	}
	again := invoke(t, args...).mustSucceed(t)
	if first := invoke(t, args...).mustSucceed(t); first.stdout != again.stdout {
		t.Fatal("two runs printed different documents")
	}
}

func TestDomainAssumptionsOfADomainFileNeedNoTwin(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "domain", "assumptions", filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json")).
		mustSucceed(t).document(t, "domain_assumptions")
	if origin, _ := doc["origin"].(map[string]any); origin["kind"] != "composition" {
		t.Fatalf("assumptions = %v, want the composition origin", doc["origin"])
	}
}

func TestDomainAssumptionsRefuseWhatTheyCannotRead(t *testing.T) {
	t.Parallel()
	domain := thermalTwinDomain(t)
	broken := writeFile(t, "broken.twin.json", `{"id": 1}`)
	for name, tc := range map[string]struct {
		args []string
		code string
	}{
		"a missing domain file":           {[]string{"domain", "assumptions", filepath.Join(t.TempDir(), "none.domain.json")}, "file_not_found"},
		"a domain that breaks its schema": {[]string{"domain", "assumptions", writeFile(t, "bad.domain.json", `{"id":"X"}`)}, "domain_invalid"},
		"twin choices that break theirs":  {[]string{"domain", "assumptions", domain, "--twin", broken, "--kinds-dir", testsupport.TwinThermalKindsDir()}, "kind_failed"},
		"no domain file at all":           {[]string{"domain", "assumptions"}, "usage_invalid"},
		"a twin without kind files":       {[]string{"domain", "assumptions", domain, "--twin", testsupport.TwinThermalChoices()}, "kind_failed"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if failure := invoke(t, tc.args...).failure(t); failure["code"] != tc.code {
				t.Fatalf("failure = %v, want code %s", failure, tc.code)
			}
		})
	}
}

func TestDomainAssumptionsIgnoreKindFilesWhenNoTwinIsGiven(t *testing.T) {
	t.Parallel()
	domain := thermalTwinDomain(t)
	plain := invoke(t, "domain", "assumptions", domain).mustSucceed(t)
	withKinds := invoke(t, "domain", "assumptions", domain, "--kinds-dir", testsupport.TwinThermalKindsDir()).mustSucceed(t)
	if plain.stdout != withKinds.stdout {
		t.Fatal("kind files without a twin changed the document")
	}
}

func TestDomainAssumptionsReadsTheCustomerLibraryItIsGiven(t *testing.T) {
	t.Parallel()
	pilot := filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json")
	invoke(t, "domain", "assumptions", pilot, "--components-dir", customerLibrary(t)).mustSucceed(t).document(t, "domain_assumptions")
}
