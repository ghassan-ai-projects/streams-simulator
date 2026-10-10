package protocol

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/app"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewOperatorServer builds the operator-role server for one world's view.
// It advertises exactly the four operator tools and nothing else.
func NewOperatorServer(v *app.OperatorView) *mcp.Server {
	return NewOperatorServerResolver(app.SingleView(v))
}

// NewOperatorServerResolver builds the operator-role server for a token
// resolver. One endpoint serves every world: each tool call resolves the
// capability token to its owning OperatorView before dispatching.
func NewOperatorServerResolver(r app.OperatorResolver) *mcp.Server {
	s := mcp.NewServer(implementation, nil)
	handlers := operatorToolHandlers{resolver: r}
	addTool(s, toolDef{name: "sim.nameplate.read", description: "The static world nameplate: entities, channels, effectors.", schema: toolSchema("sim.nameplate.read"), handler: handlers.handleNameplateRead})
	addTool(s, toolDef{name: "sim.effector.list", description: "The declared effectors and their argument schemas.", schema: toolSchema("sim.effector.list"), handler: handlers.handleEffectorList})
	addTool(s, toolDef{name: "sim.effector.invoke", description: "Invoke an effector by name with a command_id (idempotency key) and capability token.", schema: toolSchema("sim.effector.invoke"), handler: handlers.handleEffectorInvoke})
	addTool(s, toolDef{name: "sim.consumer.report", description: "Report quiescence and submit a consumer verdict. Write-only; never returns a score.", schema: toolSchema("sim.consumer.report"), handler: handlers.handleConsumerReport})
	return s
}

type operatorToolHandlers struct{ resolver app.OperatorResolver }

func (h operatorToolHandlers) resolve(args map[string]any) (*app.OperatorView, error) {
	v, err := h.resolver.ResolveOperator(app.Str(args, "token"))
	if err != nil {
		return nil, app.ToolErrorf(app.CodeCapabilityDenied, "capability token required")
	}
	return v, nil
}
func (h operatorToolHandlers) handleNameplateRead(_ context.Context, args map[string]any) (any, error) {
	v, err := h.resolve(args)
	if err != nil {
		return nil, err
	}
	return v.ReadNameplate(app.Str(args, "token"))
}
func (h operatorToolHandlers) handleEffectorList(_ context.Context, args map[string]any) (any, error) {
	v, err := h.resolve(args)
	if err != nil {
		return nil, err
	}
	return v.ListEffectors(app.Str(args, "token"))
}
func (h operatorToolHandlers) handleEffectorInvoke(_ context.Context, args map[string]any) (any, error) {
	v, err := h.resolve(args)
	if err != nil {
		return nil, err
	}
	return v.Invoke(app.Str(args, "token"), app.Str(args, "effector"), app.Str(args, "entity_id"), app.Str(args, "command_id"), mapArg(args, "args"), app.Num(args, "at_ns", 0))
}
func (h operatorToolHandlers) handleConsumerReport(_ context.Context, args map[string]any) (any, error) {
	v, err := h.resolve(args)
	if err != nil {
		return nil, err
	}
	verdict, err := consumerVerdictArgument(args)
	if err != nil {
		return nil, err
	}
	if err := v.Report(app.Str(args, "token"), app.Str(args, "run_id"), app.Num(args, "quiesced_through_ns", 0), verdict); err != nil {
		return nil, fmt.Errorf("mcp: %w", err)
	}
	return map[string]any{"accepted": true, "world_id": v.WorldID, "simulated": true}, nil
}
