package domain

import (
	"bytes"
	"cmp"
	"encoding/json"
	"io/fs"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/contract"
)

func ContractDigest() string {
	var body bytes.Buffer
	for _, name := range schemaNames() {
		raw, _ := fs.ReadFile(contract.Schemas, "schemas/"+name)
		var schema any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		_ = decoder.Decode(&schema)
		canonicalJSON, _ := canonical.Marshal(schema)
		body.WriteString(name + "\n")
		body.Write(canonicalJSON)
		body.WriteString("\n")
	}
	return canonical.DigestBytes(body.Bytes())
}

func schemaNames() []string {
	entries, _ := fs.ReadDir(contract.Schemas, "schemas")
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return names
}

func CapabilityDigest(templates []any) string {
	entries := make([]any, 0, len(templates))
	for _, template := range templates {
		kind := template.(map[string]any)
		commands := slices.Clone(kind["commands"].([]any))
		slices.SortFunc(commands, func(a, b any) int { return cmp.Compare(nameOf(a), nameOf(b)) })
		entries = append(entries, map[string]any{"kind": kind["kind"], "commands": commands})
	}
	slices.SortFunc(entries, func(a, b any) int {
		return cmp.Compare(a.(map[string]any)["kind"].(string), b.(map[string]any)["kind"].(string))
	})
	digest, _ := canonical.Digest(entries)
	return digest
}

func nameOf(command any) string {
	name, _ := command.(map[string]any)["name"].(string)
	return name
}

func KindDigest(entry map[string]any) string {
	body := maps.Clone(entry)
	delete(body, "kind_digest")
	digest, _ := canonical.Digest(body)
	return digest
}

func WithKindDigests(templates []any) []any {
	for _, template := range templates {
		entry := template.(map[string]any)
		entry["kind_digest"] = KindDigest(entry)
	}
	return templates
}

func KindsMisdigested(kinds []any) bool {
	return slices.ContainsFunc(kinds, func(kind any) bool {
		entry, _ := kind.(map[string]any)
		declared, _ := entry["kind_digest"].(string)
		return declared != KindDigest(entry)
	})
}
