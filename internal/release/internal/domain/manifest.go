// Package domain holds the release rules: the manifest body, its canonical
// bytes, the signature over their digest and the published document.
package domain

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

const SchemaVersion = "release-manifest-v0.1"

type Identity struct {
	Version, Commit, Author, Reviewer, CreatedAt string
}

func Body(identity Identity, domains, adapters map[string]any) map[string]any {
	return map[string]any{"schema_version": SchemaVersion,
		"sim":     map[string]any{"version": identity.Version, "commit": identity.Commit, "go": runtime.Version()},
		"domains": domains, "adapters": adapters,
		"consumer": map[string]any{"name": "streamsim-refconsumer", "version": "0.1.0"},
		"suite":    map[string]any{"negative_class_fraction": 0.4, "trivial_cutoff": 0.9},
		"author":   identity.Author, "reviewer": identity.Reviewer, "created_at": identity.CreatedAt}
}

func Canonical(body map[string]any) ([]byte, error) {
	raw, err := canonical.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("manifest: canonical: %w", err)
	}
	return raw, nil
}

func Document(body []byte, signature string) ([]byte, error) {
	doc := map[string]any{"manifest": json.RawMessage(body), "signature": signature}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	return raw, nil
}

func FileNames(entries []string) []string {
	names := make([]string, 0, len(entries))
	for _, name := range entries {
		if !strings.HasPrefix(name, ".") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
