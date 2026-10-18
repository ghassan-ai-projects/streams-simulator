package domain

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
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

var ErrNotCanonical = errors.New("plantwire: canonical form of a kind")

func KindDigest(entry map[string]any) (string, error) {
	body := maps.Clone(entry)
	delete(body, "kind_digest")
	digest, err := canonical.Digest(body)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotCanonical, err)
	}
	return digest, nil
}

func WithKindDigests(templates []any) ([]any, error) {
	for _, template := range templates {
		entry := template.(map[string]any)
		digest, err := KindDigest(entry)
		if err != nil {
			return nil, err
		}
		entry["kind_digest"] = digest
	}
	return templates, nil
}

func KindsMisdigested(kinds []any) bool {
	return slices.ContainsFunc(kinds, KindMisdigested)
}

func KindMisdigested(kind any) bool {
	entry, _ := kind.(map[string]any)
	declared, _ := entry["kind_digest"].(string)
	digest, err := KindDigest(entry)
	return err != nil || declared != digest
}
