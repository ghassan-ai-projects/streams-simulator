package app

import (
	"os"
	"path/filepath"
	"testing"

	wire "github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func semanticSourcesOf(t *testing.T) []wire.KindSource {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(semanticKindsDir(), "*.kind.json"))
	sources := make([]wire.KindSource, len(paths))
	for i, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources[i] = wire.KindSource{Path: path, Raw: raw}
	}
	return sources
}

func birthsOfRecords(sent []map[string]any) []map[string]any {
	var births []map[string]any
	for _, record := range sent {
		if record["type"] == "NBIRTH" {
			births = append(births, record)
		}
	}
	return births
}

func containsWord(record map[string]any, word string) bool {
	if record["type"] != "DDATA" || record["device"] != "led-01" {
		return false
	}
	samples, _ := record["metrics"].([]any)
	for _, sample := range samples {
		if fields, _ := sample.(map[string]any); fields["value"] == word {
			return true
		}
	}
	return false
}
