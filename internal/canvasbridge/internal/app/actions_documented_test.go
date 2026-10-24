package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func TestEveryToolActionNamesAToolTheDirectorDocuments(t *testing.T) {
	t.Parallel()
	reference, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "documentation", "reference", "mcp.md"))
	if err != nil {
		t.Fatal(err)
	}
	h := newHarness(t)
	for _, action := range h.bridge.Actions() {
		if action.Kind != domain.KindMCP {
			continue
		}
		t.Run(action.Name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(string(reference), "| `"+action.Tool+"` |") {
				t.Fatalf("%s maps to %s, which documentation/reference/mcp.md does not list", action.Name, action.Tool)
			}
		})
	}
}
