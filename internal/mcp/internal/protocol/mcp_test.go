package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// newTestDirector builds a director with the aquaculture-pond domain and
// the native-jsonl adapter.
func newTestDirector(t *testing.T) *director.Director {
	t.Helper()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	d := director.New(context.Background(), cat, map[string]*model.Adapter{"native-jsonl": a}, t.TempDir())
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
	t.Parallel()
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
	createWorld(t, d2)
	opServer := NewOperatorServerResolver(d2)
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
	t.Parallel()
	d := newTestDirector(t)
	servers := map[string]*mcp.Server{"director": NewDirectorServer(d)}
	createWorld(t, d)
	servers["operator"] = NewOperatorServerResolver(d)

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
	t.Parallel()
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
	if _, err := d.DescribeWorld("w-1"); err == nil || !strings.HasPrefix(err.Error(), "world_not_found: ") {
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
	t.Parallel()
	d := newTestDirector(t)
	created, err := d.CreateWorld(director.WorldRequest{
		Domain: "aquaculture-pond", Seed: 42, Adapter: "native-jsonl",
		Sink: model.SinkInproc, TimeMode: model.TimeStepped,
	})
	if err != nil {
		t.Fatal(err)
	}
	worldID := created["world_id"].(string)
	if got, want := created["clock"], model.FormatTime(model.DefaultStartTimeNS); got != want {
		t.Fatalf("default MCP start time = %v, want %v", got, want)
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
	described, err := d.DescribeWorld(worldID)
	if err != nil {
		t.Fatal(err)
	}
	if out["emitted_total"] != float64(described["emitted"].(int64)) {
		t.Fatalf("emitted_total = %v, want %v", out["emitted_total"], described["emitted"])
	}
	if want := model.FormatTime(model.DefaultStartTimeNS + 60*1e9); out["clock"] != want || described["clock"] != want {
		t.Fatalf("relative advance clock = %v / %v, want %v", out["clock"], described["clock"], want)
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

func createWorld(t *testing.T, d *director.Director) string {
	t.Helper()
	return newWorld(t, d).id
}

// createdWorld is what sim.world.create hands its caller: the identities a
// test needs to drive the world through its tools.
type createdWorld struct{ id, runID, token, entity string }

func newWorld(t *testing.T, d *director.Director) createdWorld {
	t.Helper()
	start := model.DefaultStartTimeNS + 4*3600*1e9
	res, err := d.CreateWorld(director.WorldRequest{
		Domain: "aquaculture-pond", Seed: 42, Adapter: "native-jsonl",
		Sink: model.SinkInproc, TimeMode: model.TimeStepped, StartTimeNS: &start,
	})
	if err != nil {
		t.Fatal(err)
	}
	return createdWorld{
		id: res["world_id"].(string), runID: res["run_id"].(string), token: res["token"].(string),
		entity: res["entity_ids"].([]string)[0],
	}
}

// A client that reads the server's identity sees the simulator it is
// measuring with, not a constant that outlives a version bump.
func TestServerAdvertisesTheSimulatorVersion(t *testing.T) {
	t.Parallel()
	cs, _ := connect(t, NewDirectorServer(newTestDirector(t)))
	if got := cs.InitializeResult().ServerInfo.Version; got != model.SimVersion {
		t.Fatalf("server version = %q, want %q", got, model.SimVersion)
	}
}
