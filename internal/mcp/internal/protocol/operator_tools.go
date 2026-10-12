package protocol

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewOperatorServerResolver builds the operator-role server over a director.
// One endpoint serves every world: each tool call resolves the capability
// token to its owning world before dispatching. It advertises exactly the
// four operator tools and nothing else.
func NewOperatorServerResolver(dir *director.Director) *mcp.Server {
	s := mcp.NewServer(implementation, nil)
	handlers := operatorToolHandlers{dir}
	addTool(s, toolDef{name: "sim.nameplate.read", description: "The static world nameplate: entities, channels, effectors.", schema: toolSchema("sim.nameplate.read"), handler: handlers.handleNameplateRead})
	addTool(s, toolDef{name: "sim.effector.list", description: "The declared effectors and their argument schemas.", schema: toolSchema("sim.effector.list"), handler: handlers.handleEffectorList})
	addTool(s, toolDef{name: "sim.effector.invoke", description: "Invoke an effector by name with a command_id (idempotency key) and capability token.", schema: toolSchema("sim.effector.invoke"), handler: handlers.handleEffectorInvoke})
	addTool(s, toolDef{name: "sim.consumer.report", description: "Report quiescence and submit a consumer verdict. Write-only; never returns a score.", schema: toolSchema("sim.consumer.report"), handler: handlers.handleConsumerReport})
	return s
}

type operatorToolHandlers struct{ *director.Director }

func (h operatorToolHandlers) handleNameplateRead(_ context.Context, args map[string]any) (any, error) {
	return h.OperatorNameplate(stringArg(args, "token"))
}
func (h operatorToolHandlers) handleEffectorList(_ context.Context, args map[string]any) (any, error) {
	return h.OperatorEffectors(stringArg(args, "token"))
}
func (h operatorToolHandlers) handleEffectorInvoke(_ context.Context, args map[string]any) (any, error) {
	return h.OperatorInvoke(stringArg(args, "token"), stringArg(args, "effector"), stringArg(args, "entity_id"), stringArg(args, "command_id"), mapArg(args, "args"), numberArg(args, "at_ns", 0))
}
func (h operatorToolHandlers) handleConsumerReport(_ context.Context, args map[string]any) (any, error) {
	token := stringArg(args, "token")
	if err := h.RequireOperator(token); err != nil {
		return nil, err
	}
	verdict, err := consumerVerdictArgument(args)
	if err != nil {
		return nil, err
	}
	worldID, err := h.OperatorReport(token, stringArg(args, "run_id"), numberArg(args, "quiesced_through_ns", 0), verdict)
	if err != nil {
		return nil, fmt.Errorf("mcp: %w", err)
	}
	return map[string]any{"accepted": true, "world_id": worldID, "simulated": true}, nil
}
