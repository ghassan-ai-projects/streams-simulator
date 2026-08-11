package adapter

// Golden regeneration: run `REGEN_GOLDEN=1 go test ./internal/adapter/ -run
// TestRegenerateGoldens -v` after an intentional adapter or fixture change.
// The goldens are committed artifacts; this test exists so the regeneration
// step is recorded and repeatable, not so it runs by default.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestRegenerateGoldens(t *testing.T) {
	if os.Getenv("REGEN_GOLDEN") == "" {
		t.Skip("set REGEN_GOLDEN=1 to regenerate the committed golden files")
	}
	fx, err := FixtureEvents()
	if err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{
		"run_id": "verify", "sim_version": "0.1.0",
		"domain_id": "fixture", "domain_version": "0.0.0",
		"world_start_time": fx[0].EventTime, "seed": float64(0),
	}
	root := filepath.Join("..", "..", "adapters")
	for _, name := range []string{"native-jsonl", "agentic-stream"} {
		a, err := Load(filepath.Join(root, name+".adapter.json"))
		if err != nil {
			t.Fatal(err)
		}
		e, err := NewEngine(a, meta)
		if err != nil {
			t.Fatal(err)
		}
		end, _ := model.ParseTime(fx[len(fx)-1].ObservedTime)
		out, err := e.RenderRun(fx, end)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "golden", name+".jsonl"), out, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote golden %s (%d bytes)", name, len(out))
	}
}
