package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	specs "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twinBytes(t *testing.T, choices string) []byte {
	t.Helper()
	composed, err := ScaffoldTwin(baseLibrary(t), kindDocuments(t, zoneKindJSON, fanKindJSON), []byte(choices), "fan-zone.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return composed.Bytes
}

func fileSide(raw []byte, label string) DiffInput {
	return DiffInput{Raw: raw, Label: label}
}

func diffOf(t *testing.T, a, b DiffInput) *Diff {
	t.Helper()
	got, err := DescribeDiff(a, b)
	if err != nil {
		t.Fatalf("DescribeDiff = %v, want a document", err)
	}
	return got
}

func editedDomain(t *testing.T, raw []byte, change func(doc map[string]any)) []byte {
	t.Helper()
	return []byte(mutate(t, string(raw), change))
}

func (d *Diff) kind(t *testing.T, name string) KindDiff {
	t.Helper()
	for _, kind := range d.Kinds {
		if kind.Kind == name {
			return kind
		}
	}
	t.Fatalf("no kind %q in the diff of %v", name, d.kindNames())
	return KindDiff{}
}

func (d *Diff) kindNames() []string {
	names := []string{}
	for _, kind := range d.Kinds {
		names = append(names, kind.Kind+":"+kind.Status)
	}
	return names
}

func (k KindDiff) change(t *testing.T, code string) Change {
	t.Helper()
	for _, change := range k.Changes {
		if change.Code == code {
			return change
		}
	}
	t.Fatalf("kind %s has no %s change: %v", k.Kind, code, codesOf(k.Changes))
	return Change{}
}

func codesOf(changes []Change) []string {
	codes := make([]string, len(changes))
	for i, change := range changes {
		codes[i] = change.Code
	}
	return codes
}

func gainChoices(gain string) string {
	return strings.Replace(fanZoneChoicesJSON, `"gain": -0.02`, `"gain": `+gain, 1)
}

func runSide(t *testing.T, domainRaw []byte, label string) DiffInput {
	t.Helper()
	run := readRunTemplate(t)
	compiled := compiledDomain(t, domainRaw)
	run["domain"] = map[string]any{"id": compiled.Spec.ID, "version": compiled.Spec.Version, "digest": compiled.Digest}
	run["domain_spec"] = json.RawMessage(domainRaw)
	raw, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	return DiffInput{Raw: raw, Label: label, Run: true}
}

func readRunTemplate(t *testing.T) map[string]any {
	t.Helper()
	raw := readTestFile(t, testsupport.ReportRun()+"/run.json")
	var run map[string]any
	if err := json.Unmarshal(raw, &run); err != nil {
		t.Fatal(err)
	}
	return run
}

func compiledDomain(t *testing.T, raw []byte) *specs.Compiled {
	t.Helper()
	compiled, err := specs.Parse(raw, "test.domain.json")
	if err != nil {
		t.Fatal(err)
	}
	return compiled
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func addKindCopy(doc map[string]any, from, name, instance string) {
	kinds, _ := doc["kinds"].([]any)
	for _, raw := range kinds {
		if kind, _ := raw.(map[string]any); kind["name"] == from {
			var copied map[string]any
			_ = json.Unmarshal(mustJSONValue(kind), &copied)
			copied["name"] = name
			delete(copied, "implements")
			renameDeclared(copied, "channels", "effectors", "faults")
			doc["kinds"] = append(kinds, copied)
			break
		}
	}
	instances, _ := doc["instances"].([]any)
	doc["instances"] = append(instances, map[string]any{"kind": name, "names": []any{instance}})
}

func mustJSONValue(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

func rename(item map[string]any, key string) {
	if text, ok := item[key].(string); ok {
		item[key] = text + "_b"
	}
}

func renameDeclared(kind map[string]any, sections ...string) {
	for _, section := range sections {
		items, _ := kind[section].([]any)
		renamed := make([]any, len(items))
		for i, raw := range items {
			item := map[string]any{}
			for key, value := range raw.(map[string]any) {
				item[key] = value
			}
			rename(item, "name")
			rename(item, "id")
			if observation, ok := item["observation"].(map[string]any); ok {
				rename(observation, "channel")
				rename(item["observability"].(map[string]any)["detector"].(map[string]any), "channel")
			}
			renamed[i] = item
		}
		kind[section] = renamed
	}
}

func effectorOf(doc map[string]any, kind, command string) map[string]any {
	kinds, _ := doc["kinds"].([]any)
	for _, raw := range kinds {
		if entry, _ := raw.(map[string]any); entry["name"] == kind {
			effectors, _ := entry["effectors"].([]any)
			for _, one := range effectors {
				if effector, _ := one.(map[string]any); effector["name"] == command {
					return effector
				}
			}
		}
	}
	return nil
}

func kindOf(doc map[string]any, name string) map[string]any {
	kinds, _ := doc["kinds"].([]any)
	for _, raw := range kinds {
		if kind, _ := raw.(map[string]any); kind["name"] == name {
			return kind
		}
	}
	return nil
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

const digestSample = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
