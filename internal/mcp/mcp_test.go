package mcp_test

import (
	"context"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func newDirector(t *testing.T) *mcp.Director {
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
	adapters := map[string]*model.Adapter{"native-jsonl": native}
	return mcp.NewDirector(t.Context(), cat, adapters, t.TempDir())
}

func toolNames(t *testing.T, server *mcpsdk.Server) map[string]bool {
	t.Helper()
	ctx := context.Background()
	ct, st := mcpsdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "facade-test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	names := map[string]bool{}
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		names[tool.Name] = true
	}
	return names
}

func TestRolesExposeDisjointToolSurfaces(t *testing.T) {
	t.Parallel()
	d := newDirector(t)
	director := toolNames(t, mcp.NewDirectorServer(d))
	operator := toolNames(t, mcp.NewOperatorServerResolver(d))
	if !director["sim.truth.reveal"] || !operator["sim.nameplate.read"] {
		t.Fatalf("each role must expose its own tools: director=%v operator=%v", director, operator)
	}
	for name := range operator {
		if director[name] && name != "sim.consumer.report" {
			t.Fatalf("operator tool %q leaked onto the director surface", name)
		}
	}
	for _, truth := range []string{"sim.truth.reveal", "sim.score", "sim.perturb.apply"} {
		if operator[truth] {
			t.Fatalf("operator role must never expose %q", truth)
		}
	}
}

func TestSetOperatorEndpointIsAdvertisedByWorldCreate(t *testing.T) {
	t.Parallel()
	d := newDirector(t)
	d.SetOperatorEndpoint("http://127.0.0.1:9/operator")
	ctx := context.Background()
	ct, st := mcpsdk.NewInMemoryTransports()
	ss, err := mcp.NewDirectorServer(d).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "facade-test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	res, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{Name: "sim.world.create", Arguments: map[string]any{
		"domain": "aquaculture-pond", "adapter": "native-jsonl", "seed": 1,
	}})
	if err != nil || res.IsError {
		t.Fatalf("world create failed: err=%v result=%+v", err, res)
	}
	out, _ := res.StructuredContent.(map[string]any)
	if out["operator_endpoint"] != "http://127.0.0.1:9/operator" {
		t.Fatalf("world create must advertise the recorded endpoint: %v", out)
	}
}
