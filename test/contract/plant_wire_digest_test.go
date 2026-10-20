package contract

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func TestFixturesCarryTheRealContractAndCapabilityDigests(t *testing.T) {
	t.Parallel()
	pinned, err := os.ReadFile(filepath.Join(plantWireDir(), "DIGEST"))
	if err != nil {
		t.Fatal(err)
	}
	birth := readJSON(t, filepath.Join(plantWireDir(), "fixtures", "valid", "NBIRTH.json"))
	if member(birth, "contract_digest") != strings.TrimSpace(string(pinned)) {
		t.Errorf("the NBIRTH fixture announces %v, the contract is %s", member(birth, "contract_digest"), pinned)
	}
	if got := capabilityDigest(t, member(birth, "kinds").([]any)); member(birth, "capability_digest") != got {
		t.Errorf("the NBIRTH fixture announces capability %v, its kinds give %s", member(birth, "capability_digest"), got)
	}
}

func TestEveryCommandFixtureCarriesTheIdempotencyKeyOfItsIntent(t *testing.T) {
	t.Parallel()
	for _, file := range fixtures(t, "exchanges", "*.jsonl") {
		if strings.HasPrefix(filepath.Base(file), "duplicate-") {
			continue
		}
		_, steps := exchangeSteps(t, file)
		for _, step := range steps {
			if record := member(step, "record"); member(record, "type") == "DCMD" && member(record, "idempotency_key") != idempotencyKey(t, record) {
				t.Errorf("%s: %v does not hash its command id and content", filepath.Base(file), member(record, "command_id"))
			}
		}
	}
}

func capabilityDigest(t *testing.T, kinds []any) string {
	t.Helper()
	var entries []any
	for _, kind := range kinds {
		commands := slices.Clone(member(kind, "commands").([]any))
		slices.SortFunc(commands, func(a, b any) int { return strings.Compare(member(a, "name").(string), member(b, "name").(string)) })
		entries = append(entries, map[string]any{"kind": member(kind, "kind"), "commands": commands})
	}
	slices.SortFunc(entries, func(a, b any) int { return strings.Compare(member(a, "kind").(string), member(b, "kind").(string)) })
	digest, err := canonical.Digest(entries)
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func idempotencyKey(t *testing.T, record any) string {
	t.Helper()
	content := map[string]any{"device": member(record, "device"), "command": member(record, "command"), "args": member(record, "args"), "expected_boot_id": member(record, "expected_boot_id")}
	digest, err := canonical.Digest(content)
	if err != nil {
		t.Fatal(err)
	}
	key, err := canonical.Digest(map[string]any{"command_id": member(record, "command_id"), "content": digest})
	if err != nil {
		t.Fatal(err)
	}
	return key
}
