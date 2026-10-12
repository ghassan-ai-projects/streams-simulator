package protocol

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func toolNames(t *testing.T, server *mcp.Server) map[string]bool {
	t.Helper()
	cs, _ := connect(t, server)
	names := map[string]bool{}
	for tool, err := range cs.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names[tool.Name] = true
	}
	return names
}

func TestRolesExposeDisjointToolSurfaces(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	directorTools := toolNames(t, NewDirectorServer(d))
	operator := toolNames(t, NewOperatorServerResolver(d))
	if !directorTools["sim.truth.reveal"] || !operator["sim.nameplate.read"] {
		t.Fatalf("each role must expose its own tools: director=%v operator=%v", directorTools, operator)
	}
	for name := range operator {
		if directorTools[name] && name != "sim.consumer.report" {
			t.Fatalf("operator tool %q leaked onto the director surface", name)
		}
	}
}

func TestWorldCreateAdvertisesTheRecordedOperatorEndpoint(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	d.SetOperatorEndpoint("http://127.0.0.1:9/operator")
	cs, _ := connect(t, NewDirectorServer(d))
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "sim.world.create", Arguments: map[string]any{
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
