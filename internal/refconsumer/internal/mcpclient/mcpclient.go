// Package mcpclient is the network edge of the refconsumer module: the
// reference consumer's out-of-process operator client over the MCP operator
// endpoint, the documented deployable process boundary.
package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
)

type MCPOperator struct {
	session *mcpsdk.ClientSession
	token   string
	runID   string
}

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

func (o *MCPOperator) Close() error {
	if err := o.session.Close(); err != nil {
		return fmt.Errorf("refconsumer: close session: %w", err)
	}
	return nil
}

func errorText(res *mcpsdk.CallToolResult) string {
	for _, c := range res.Content {
		if t, ok := c.(*mcpsdk.TextContent); ok && t.Text != "" {
			return t.Text
		}
	}
	return ""
}

func (o *MCPOperator) call(name string, args map[string]any, out any) error {
	res, err := o.session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return fmt.Errorf("refconsumer: %s: %w", name, err)
	}
	if err := admitOperatorResult(name, res); err != nil {
		return err
	}
	return decodeOperatorResult(name, res.StructuredContent, out)
}

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

func (o *MCPOperator) Nameplate() (*domain.Nameplate, error) {
	var np nameplateShape
	if err := o.call("sim.nameplate.read", map[string]any{"token": o.token}, &np); err != nil {
		return nil, err
	}
	out := &domain.Nameplate{WorldID: np.WorldID}
	np.appendEntities(out)
	np.appendChannels(out)
	np.appendEffectors(out)
	return out, nil
}

func (o *MCPOperator) ListEffectors() ([]domain.EffectorInfo, error) {
	var raw []struct {
		Name       string         `json:"name"`
		ArgsSchema map[string]any `json:"args_schema"`
	}
	if err := o.call("sim.effector.list", map[string]any{"token": o.token}, &raw); err != nil {
		return nil, err
	}
	out := make([]domain.EffectorInfo, 0, len(raw))
	for _, e := range raw {
		out = append(out, domain.EffectorInfo{Name: e.Name, Schema: e.ArgsSchema})
	}
	return out, nil
}

func (o *MCPOperator) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*model.EffectorAcknowledgement, error) {
	var res model.EffectorAcknowledgement
	if err := o.call("sim.effector.invoke", map[string]any{
		"token": o.token, "effector": effector, "entity_id": entityID,
		"command_id": commandID, "args": args,
	}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (o *MCPOperator) SubmitVerdict(v *model.Verdict) error {
	return o.call("sim.consumer.report", map[string]any{
		"token": o.token, "run_id": o.runID, "verdict": v,
	}, nil)
}

func (o *MCPOperator) ReportQuiesced(throughNS int64) error {
	return o.call("sim.consumer.report", map[string]any{
		"token": o.token, "run_id": o.runID, "quiesced_through_ns": throughNS,
	}, nil)
}

func admitOperatorResult(name string, res *mcpsdk.CallToolResult) error {
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
	return nil
}

func decodeOperatorResult(name string, content, out any) error {
	raw, err := json.Marshal(content)
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

func (np *nameplateShape) appendEntities(out *domain.Nameplate) {
	for _, entity := range np.Entities {
		out.Entities = append(out.Entities, domain.EntityInfo{ID: entity.ID, Type: entity.Type})
	}
}

func (np *nameplateShape) appendChannels(out *domain.Nameplate) {
	for _, channel := range np.Channels {
		out.Channels = append(out.Channels, domain.ChannelInfo{Name: channel.Name, Unit: channel.Unit,
			RangeMin: channel.RangeMin, RangeMax: channel.RangeMax, Resolution: channel.Resolution})
	}
}

func (np *nameplateShape) appendEffectors(out *domain.Nameplate) {
	for _, effector := range np.Effectors {
		out.Effectors = append(out.Effectors, domain.EffectorInfo{Name: effector.Name, Schema: effector.ArgsSchema})
	}
}
