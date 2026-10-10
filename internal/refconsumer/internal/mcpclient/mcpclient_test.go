package mcpclient

import (
	"context"
	"errors"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// fakeOperator serves the four operator tools the client uses, recording the
// arguments of each call.
type fakeOperator struct {
	calls   map[string]map[string]any
	refuse  string
	refusal string
}

func (f *fakeOperator) handler(name string, result any) func(context.Context, *mcpsdk.CallToolRequest, map[string]any) (*mcpsdk.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcpsdk.CallToolRequest, args map[string]any) (*mcpsdk.CallToolResult, any, error) {
		f.calls[name] = args
		if f.refuse == name {
			return nil, nil, errors.New(f.refusal)
		}
		return nil, result, nil
	}
}

func operatorOver(t *testing.T, f *fakeOperator) *MCPOperator {
	t.Helper()
	f.calls = map[string]map[string]any{}
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "fake-operator"}, nil)
	object := map[string]any{"type": "object"}
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.nameplate.read", InputSchema: object}, f.handler("sim.nameplate.read", map[string]any{
		"world_id": "w-1", "entities": []any{map[string]any{"id": "e-1", "type": "pond"}},
		"channels":  []any{map[string]any{"name": "c", "unit": "u", "range_min": 0.0, "range_max": 9.0, "resolution": 0.1}},
		"effectors": []any{map[string]any{"name": "act", "args_schema": map[string]any{"type": "object"}}},
	}))
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.effector.list", InputSchema: object}, f.handler("sim.effector.list", []any{map[string]any{"name": "act", "args_schema": map[string]any{"type": "object"}}}))
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.effector.invoke", InputSchema: object}, f.handler("sim.effector.invoke", map[string]any{"accepted": true, "command_id": "c-1", "mode": "ok"}))
	mcpsdk.AddTool(server, &mcpsdk.Tool{Name: "sim.consumer.report", InputSchema: object}, f.handler("sim.consumer.report", map[string]any{"ok": true}))
	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test"}, nil)
	session, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	return &MCPOperator{session: session, token: "tok", runID: "r-1"}
}

func TestNewMCPOperatorRequiresEndpointTokenAndRunTogether(t *testing.T) {
	t.Parallel()
	for _, args := range [][3]string{{"", "t", "r"}, {"http://x", "", "r"}, {"http://x", "t", ""}} {
		if _, err := NewMCPOperator(args[0], args[1], args[2]); err == nil || !strings.Contains(err.Error(), "required together") {
			t.Fatalf("%v: err = %v", args, err)
		}
	}
}

func TestNameplateAndEffectorListAreDecodedFromTheOperatorSurface(t *testing.T) {
	t.Parallel()
	f := &fakeOperator{}
	o := operatorOver(t, f)
	np, err := o.Nameplate()
	if err != nil {
		t.Fatal(err)
	}
	if np.WorldID != "w-1" || len(np.Entities) != 1 || np.Entities[0].Type != "pond" ||
		len(np.Channels) != 1 || np.Channels[0].RangeMax != 9 || len(np.Effectors) != 1 {
		t.Fatalf("nameplate = %+v", np)
	}
	if f.calls["sim.nameplate.read"]["token"] != "tok" {
		t.Fatalf("the capability token must travel with every call: %v", f.calls)
	}
	list, err := o.ListEffectors()
	if err != nil || len(list) != 1 || list[0].Name != "act" {
		t.Fatalf("effectors = %+v (%v)", list, err)
	}
}

func TestInvokeReportAndQuiescenceTravelWithTokenAndRun(t *testing.T) {
	t.Parallel()
	f := &fakeOperator{}
	o := operatorOver(t, f)
	result, err := o.InvokeEffector("act", "e-1", "c-1", map[string]any{"k": "v"}, 42)
	if err != nil || !result.Accepted || result.CommandID != "c-1" {
		t.Fatalf("invoke = %+v (%v)", result, err)
	}
	if got := f.calls["sim.effector.invoke"]; got["effector"] != "act" || got["entity_id"] != "e-1" || got["token"] != "tok" {
		t.Fatalf("invoke arguments = %v", got)
	}
	if err := o.SubmitVerdict(&model.Verdict{}); err != nil {
		t.Fatal(err)
	}
	if got := f.calls["sim.consumer.report"]; got["run_id"] != "r-1" || got["token"] != "tok" {
		t.Fatalf("report arguments = %v", got)
	}
	if err := o.ReportQuiesced(7); err != nil {
		t.Fatal(err)
	}
	if got := f.calls["sim.consumer.report"]; got["quiesced_through_ns"] == nil {
		t.Fatalf("quiescence arguments = %v", got)
	}
}

func TestARefusalCarriesTheSimulatorsReason(t *testing.T) {
	t.Parallel()
	o := operatorOver(t, &fakeOperator{refuse: "sim.effector.invoke", refusal: "capability_denied: no token"})
	_, err := o.InvokeEffector("act", "e-1", "c-1", nil, 1)
	if err == nil || !strings.Contains(err.Error(), "sim.effector.invoke") {
		t.Fatalf("err = %v", err)
	}
	if err := o.Close(); err != nil {
		t.Fatal(err)
	}
}
