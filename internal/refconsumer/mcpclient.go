package refconsumer

// MCPOperator is the reference consumer's out-of-process operator client:
// it actuates and reports through the director process's operator endpoint,
// the documented deployable process boundary. Everything it does goes over
// the operator tools; it never sees a director tool, the ledger, or truth.

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// MCPOperator is a VerdictSink, QuiescenceReporter and EffectorInvoker over
// the MCP operator surface.
type MCPOperator struct {
	session *mcpsdk.ClientSession
	token   string
	runID   string
}

// NewMCPOperator connects to an operator endpoint with a capability token.
func NewMCPOperator(endpoint, token, runID string) (*MCPOperator, error) {
	if endpoint == "" || token == "" || runID == "" {
		return nil, fmt.Errorf("refconsumer: --mcp, --token and --run are required together")
	}
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "streamsim-refconsumer", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{Endpoint: endpoint}, nil)
	if err != nil {
		return nil, fmt.Errorf("refconsumer: connect %s: %w", endpoint, err)
	}
	return &MCPOperator{session: session, token: token, runID: runID}, nil
}

// Close ends the session.
func (o *MCPOperator) Close() error {
	if err := o.session.Close(); err != nil {
		return fmt.Errorf("refconsumer: close session: %w", err)
	}
	return nil
}

// errorText extracts the first text content of a refused tool result. The
// simulator's operator errors carry the stable code and reason in the
// content; discarding it would hide capability and effector refusals.
func errorText(res *mcpsdk.CallToolResult) string {
	for _, c := range res.Content {
		if t, ok := c.(*mcpsdk.TextContent); ok && t.Text != "" {
			return t.Text
		}
	}
	return ""
}

// call invokes one operator tool and decodes the structured result.
func (o *MCPOperator) call(name string, args map[string]any, out any) error {
	res, err := o.session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return fmt.Errorf("refconsumer: %s: %w", name, err)
	}
	if res.IsError {
		reason := errorText(res)
		if reason != "" {
			return fmt.Errorf("refconsumer: %s refused by the simulator: %s", name, reason)
		}
		return fmt.Errorf("refconsumer: %s refused by the simulator", name)
	}
	if res.StructuredContent == nil {
		return fmt.Errorf("refconsumer: %s returned no structured content", name)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		return fmt.Errorf("refconsumer: %s: %w", name, err)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("refconsumer: %s: %w", name, err)
		}
	}
	return nil
}

// nameplateShape mirrors the simulator's nameplate over the wire.
type nameplateShape struct {
	WorldID       string `json:"world_id"`
	Domain        string `json:"domain"`
	DomainVersion string `json:"domain_version"`
	Entities      []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		BornAtNS int64  `json:"born_at_ns"`
	} `json:"entities"`
	Channels []struct {
		Name       string  `json:"name"`
		Unit       string  `json:"unit,omitempty"`
		RangeMin   float64 `json:"range_min,omitempty"`
		RangeMax   float64 `json:"range_max,omitempty"`
		Resolution float64 `json:"resolution"`
	} `json:"channels"`
	Effectors []struct {
		Name       string         `json:"name"`
		ArgsSchema map[string]any `json:"args_schema"`
	} `json:"effectors"`
}

// Nameplate reads the world's static nameplate over the operator surface.
func (o *MCPOperator) Nameplate() (*Nameplate, error) {
	var np nameplateShape
	if err := o.call("sim.nameplate.read", map[string]any{"token": o.token}, &np); err != nil {
		return nil, err
	}
	out := &Nameplate{WorldID: np.WorldID}
	for _, e := range np.Entities {
		out.Entities = append(out.Entities, EntityInfo{ID: e.ID, Type: e.Type})
	}
	for _, c := range np.Channels {
		out.Channels = append(out.Channels, ChannelInfo{
			Name: c.Name, Unit: c.Unit, RangeMin: c.RangeMin, RangeMax: c.RangeMax, Resolution: c.Resolution,
		})
	}
	for _, e := range np.Effectors {
		out.Effectors = append(out.Effectors, EffectorInfo{Name: e.Name, Schema: e.ArgsSchema})
	}
	return out, nil
}

// ListEffectors reads the declared effectors over sim.effector.list.
func (o *MCPOperator) ListEffectors() ([]EffectorInfo, error) {
	var raw []struct {
		Name       string         `json:"name"`
		ArgsSchema map[string]any `json:"args_schema"`
	}
	if err := o.call("sim.effector.list", map[string]any{"token": o.token}, &raw); err != nil {
		return nil, err
	}
	out := make([]EffectorInfo, 0, len(raw))
	for _, e := range raw {
		out = append(out, EffectorInfo{Name: e.Name, Schema: e.ArgsSchema})
	}
	return out, nil
}

// InvokeEffector actuates through sim.effector.invoke.
func (o *MCPOperator) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	var res world.InvokeResult
	if err := o.call("sim.effector.invoke", map[string]any{
		"token": o.token, "effector": effector, "entity_id": entityID,
		"command_id": commandID, "args": args, "at_ns": atNS,
	}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// SubmitVerdict submits the final report through sim.consumer.report.
func (o *MCPOperator) SubmitVerdict(v *model.Verdict) error {
	return o.call("sim.consumer.report", map[string]any{
		"token": o.token, "run_id": o.runID, "verdict": v,
	}, nil)
}

// ReportQuiesced asserts the consumer's watermark through
// sim.consumer.report.
func (o *MCPOperator) ReportQuiesced(throughNS int64) error {
	return o.call("sim.consumer.report", map[string]any{
		"token": o.token, "run_id": o.runID, "quiesced_through_ns": throughNS,
	}, nil)
}
