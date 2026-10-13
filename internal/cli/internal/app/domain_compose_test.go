package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestDomainComposeWritesTheDomainTheExampleHolds(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "pilot.domain.json")
	doc := invoke(t, "domain", "compose", pilotComposition(), "--components-dir", builtInComponents(), "--out", out).
		mustSucceed(t).document(t, "domain_compose")
	if doc["composed"] != out || doc["id"] != "chiller-pilot" || doc["domain"] != nil {
		t.Fatalf("compose = %v, want the file named and no inline domain", doc)
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "examples", "chiller-pilot.domain.json"))
	if err != nil || !bytes.Equal(written, committed) {
		t.Fatalf("the written domain differs from the committed pilot example (%v)", err)
	}
	invoke(t, "domain", "validate", out).mustSucceed(t).document(t, "domain_validate")
}

func TestDomainComposeWithoutOutPrintsTheDomainInTheDocument(t *testing.T) {
	t.Parallel()
	doc := invoke(t, "domain", "compose", "--components-dir", builtInComponents(), pilotComposition()).
		mustSucceed(t).document(t, "domain_compose")
	spec, _ := doc["domain"].(map[string]any)
	if _, written := doc["composed"]; written || spec["id"] != "chiller-pilot" {
		t.Fatalf("compose = %v, want the domain inline and no file", doc)
	}
}

func TestDomainComposeReadsTheLibraryDirectoriesInOrder(t *testing.T) {
	t.Parallel()
	custom := customerLibrary(t)
	raw, err := os.ReadFile(pilotComposition())
	if err != nil {
		t.Fatal(err)
	}
	swapped := strings.Replace(string(raw), `"use": "streamsim/ahu@1"`, `"use": "acme/ahu@1"`, 1)
	path := writeFile(t, "custom.composition.json", swapped)
	environment := map[string]string{"STREAMSIM_COMPONENTS_DIRS": custom}
	doc := invokeWithEnvironment(t, environment, "domain", "compose", path, "--components-dir", builtInComponents()).
		mustSucceed(t).document(t, "domain_compose")
	if doc["id"] != "chiller-pilot" {
		t.Fatalf("compose = %v", doc)
	}
}

func TestDomainComposeReportsACompositionThatBreaksALibraryRule(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(pilotComposition())
	if err != nil {
		t.Fatal(err)
	}
	path := writeFile(t, "bad.composition.json", strings.Replace(string(raw), `"preset": "large"`, `"preset": "huge"`, 1))
	failure := invoke(t, "domain", "compose", path, "--components-dir", builtInComponents()).failure(t)
	issues, _ := failure["issues"].([]any)
	issue, _ := issues[0].(map[string]any)
	if failure["code"] != "composition_invalid" || failure["source"] != path || issue["pointer"] != "/equipment/0/preset" {
		t.Fatalf("failure = %v, want composition_invalid at /equipment/0/preset of %s", failure, path)
	}
}

func TestDomainComposeReportsADomainThatFailsItsOwnContract(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(pilotComposition())
	if err != nil {
		t.Fatal(err)
	}
	path := writeFile(t, "weights.composition.json", strings.Replace(string(raw), "chiller_large__chw_supply_temp__offset", "no_such_fault", 1))
	failure := invoke(t, "domain", "compose", path, "--components-dir", builtInComponents()).failure(t)
	if failure["code"] != "domain_invalid" {
		t.Fatalf("failure = %v, want domain_invalid", failure)
	}
}

func TestDomainComposeRefusesWhatItCannotRead(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		code string
		exit int
	}{
		{"no composition", []string{"domain", "compose", "--components-dir", builtInComponents()}, "usage_invalid", 2},
		{"two compositions", []string{"domain", "compose", "a.json", "b.json"}, "usage_invalid", 2},
		{"an unknown flag", []string{"domain", "compose", "--frobnicate", "a.json"}, "usage_invalid", 2},
		{"a composition that does not exist", []string{"domain", "compose", "absent.composition.json", "--components-dir", builtInComponents()}, "file_not_found", 1},
		{"a library that does not exist", []string{"domain", "compose", pilotComposition(), "--components-dir", "no-such-dir"}, "file_not_found", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tc.args...)
			if got.code != tc.exit || got.failure(t)["code"] != tc.code {
				t.Fatalf("exit %d, want %d with code %s: %+v", got.code, tc.exit, tc.code, got)
			}
		})
	}
}
