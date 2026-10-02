package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

const (
	aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"
	nativeAdapter   = "../../adapters/native-jsonl.adapter.json"
)

// newTestDirector builds a director with the aquaculture-pond domain and
// the native-jsonl adapter.
func newTestDirector(t *testing.T) *Director {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	d := NewDirector(context.Background(), cat, map[string]*model.Adapter{"native-jsonl": a}, t.TempDir())
	return d
}

// connect returns a client session to the server over in-memory transports.
func connect(t *testing.T, server *mcp.Server) (*mcp.ClientSession, func()) {
	t.Helper()
	ctx := context.Background()
	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		_ = cs.Close()
		_ = ss.Close()
		_ = ss.Wait()
	}
	t.Cleanup(cleanup)
	return cs, cleanup
}

func callTool(t *testing.T, server *mcp.Server, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	cs, _ := connect(t, server)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return nil, fmt.Errorf("callTool %s: %w", name, err)
	}
	return res, nil
}

func TestRoleSeparation(t *testing.T) {
	d := newTestDirector(t)
	dirServer := NewDirectorServer(d)
	res, err := callTool(t, dirServer, "tools/list", nil)
	_ = res
	_ = err
	// Director tools include truth, fault and perturb surfaces.
	for _, name := range []string{"sim.fault.list", "sim.truth.reveal", "sim.perturb.apply", "sim.score"} {
		if !serverHasTool(t, dirServer, name) {
			t.Fatalf("director server missing %s", name)
		}
	}
	// Operator server advertises exactly the four operator tools.
	d2 := newTestDirector(t)
	world := createWorld(t, d2)
	opServer := NewOperatorServer(d2.Worlds[world].Operator)
	for _, name := range []string{"sim.nameplate.read", "sim.effector.list", "sim.effector.invoke", "sim.consumer.report"} {
		if !serverHasTool(t, opServer, name) {
			t.Fatalf("operator server missing %s", name)
		}
	}
	for _, name := range []string{"sim.truth.reveal", "sim.fault.list", "sim.score", "sim.perturb.apply"} {
		if serverHasTool(t, opServer, name) {
			t.Fatalf("operator server must not expose %s", name)
		}
	}
}

func TestToolSchemasAreClosedAndMachineReadable(t *testing.T) {
	d := newTestDirector(t)
	servers := map[string]*mcp.Server{"director": NewDirectorServer(d)}
	worldID := createWorld(t, d)
	servers["operator"] = NewOperatorServer(d.Worlds[worldID].Operator)

	for role, server := range servers {
		client, _ := connect(t, server)
		res, err := client.ListTools(context.Background(), &mcp.ListToolsParams{})
		if err != nil {
			t.Fatalf("%s tools/list: %v", role, err)
		}
		if len(res.Tools) == 0 {
			t.Fatalf("%s advertised no tools", role)
		}
		for _, tool := range res.Tools {
			raw, err := json.Marshal(tool.InputSchema)
			if err != nil {
				t.Fatalf("%s %s schema marshal: %v", role, tool.Name, err)
			}
			var schema map[string]any
			if err := json.Unmarshal(raw, &schema); err != nil {
				t.Fatalf("%s %s schema decode: %v", role, tool.Name, err)
			}
			if schema["type"] != "object" {
				t.Errorf("%s %s schema type = %v, want object", role, tool.Name, schema["type"])
			}
			if schema["additionalProperties"] != false {
				t.Errorf("%s %s must reject unknown properties", role, tool.Name)
			}
			if _, ok := schema["properties"].(map[string]any); !ok {
				t.Errorf("%s %s schema has no properties object", role, tool.Name)
			}
		}
	}
}

func TestMCPRejectsUnknownAndMissingArguments(t *testing.T) {
	d := newTestDirector(t)
	server := NewDirectorServer(d)

	res, err := callTool(t, server, "sim.catalog.describe", map[string]any{
		"domain": "aquaculture-pond", "unexpected": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("unknown argument must be rejected")
	}
	if len(d.Worlds) != 0 {
		t.Fatal("rejected catalog call must not mutate director state")
	}

	res, err = callTool(t, server, "sim.world.create", map[string]any{
		"domain": "aquaculture-pond", "sink": "file",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("file sink without sink_target must be rejected")
	}

	res, err = callTool(t, server, "sim.world.describe", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("missing required argument must be rejected")
	}
}

func TestClockAdvanceSupportsRelativeTimeAndReportsTotal(t *testing.T) {
	d := newTestDirector(t)
	created, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(42), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	worldID := created["world_id"].(string)
	w := d.World(worldID)
	if got := w.Run.World.Clock(); got != model.DefaultStartTimeNS {
		t.Fatalf("default MCP start time = %d, want %d", got, model.DefaultStartTimeNS)
	}

	res, err := callTool(t, NewDirectorServer(d), "sim.clock.advance", map[string]any{
		"world_id": worldID, "by_ns": float64(60 * 1e9),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("relative advance failed: %+v", res)
	}
	var out map[string]any
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out["emitted_total"] != float64(w.Run.World.EmittedCount()) {
		t.Fatalf("emitted_total = %v, want %d", out["emitted_total"], w.Run.World.EmittedCount())
	}
	if w.Run.World.Clock() != model.DefaultStartTimeNS+60*1e9 {
		t.Fatalf("relative advance clock = %d", w.Run.World.Clock())
	}
}

func serverHasTool(t *testing.T, server *mcp.Server, name string) bool {
	t.Helper()
	cs, _ := connect(t, server)
	res, err := cs.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func createWorld(t *testing.T, d *Director) string {
	t.Helper()
	res, err := d.CreateWorld(map[string]any{
		"domain": "aquaculture-pond", "seed": float64(42), "adapter": "native-jsonl",
		"sink": model.SinkInproc, "time_mode": model.TimeStepped,
		"start_time": float64(model.DefaultStartTimeNS + 4*3600*1e9),
	})
	if err != nil {
		t.Fatal(err)
	}
	return res["world_id"].(string)
}
