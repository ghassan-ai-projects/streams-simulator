package app

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestWorldConfigPreservesEpochZeroAndSeparateResourceIdentities(t *testing.T) {
	d := newTestDirector(t)
	first, err := d.CreateWorld(map[string]any{"domain": "aquaculture-pond", "start_time": float64(0)})
	if err != nil {
		t.Fatal(err)
	}
	second, err := d.CreateWorld(map[string]any{"domain": "aquaculture-pond"})
	if err != nil {
		t.Fatal(err)
	}
	for _, out := range []map[string]any{first, second} {
		rec := d.World(out["world_id"].(string))
		if rec == nil {
			t.Fatal("missing world")
		}
		t.Cleanup(func() {
			if _, err := rec.Run.End(""); err != nil {
				t.Error(err)
			}
		})
	}
	a := d.World(first["world_id"].(string))
	if a == nil {
		t.Fatal("missing first world")
	}
	b := d.World(second["world_id"].(string))
	if b == nil {
		t.Fatal("missing second world")
	}
	if a.Run.World.Clock() != 0 || b.Run.World.Clock() != model.DefaultStartTimeNS {
		t.Fatal("presence-aware start time changed")
	}
	if a.Run.ID == b.Run.ID || first["world_id"] == second["world_id"] || first["token"] == second["token"] {
		t.Fatal("resource identities must remain distinct")
	}
}
