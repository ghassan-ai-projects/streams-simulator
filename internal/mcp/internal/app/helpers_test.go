package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// newTestDirector builds a director with the aquaculture-pond domain and the
// native-jsonl adapter.
func newTestDirector(t *testing.T) *Director {
	t.Helper()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	return NewDirector(context.Background(), cat, map[string]*model.Adapter{"native-jsonl": native}, t.TempDir())
}

// createWorld creates a seeded aquaculture-pond world and returns its id.
func createWorld(t *testing.T, d *Director) string {
	t.Helper()
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(42), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res["world_id"].(string)
	return id
}
