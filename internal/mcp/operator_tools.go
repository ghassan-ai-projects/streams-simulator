package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewOperatorServer builds the operator-role server for one world's view.
// It advertises exactly the four operator tools and nothing else.
func NewOperatorServer(v *OperatorView) *mcp.Server {
	return NewOperatorServerResolver(viewResolver{v: v})
}

// NewOperatorServerResolver builds the operator-role server for a token
// resolver. One endpoint serves every world: each tool call resolves the
// capability token to its owning OperatorView before dispatching.
func NewOperatorServerResolver(r OperatorResolver) *mcp.Server {
	s := mcp.NewServer(implementation, nil)
	resolve := func(args map[string]any) (*OperatorView, error) {
		v, err := r.ResolveOperator(str(args, "token"))
		if err != nil {
			return nil, errTool(CodeCapabilityDenied, "capability token required")
		}
		return v, nil
	}
	addTool(s, toolDef{name: "sim.nameplate.read", description: "The static world nameplate: entities, channels, effectors.", schema: toolSchema("sim.nameplate.read"), handler: func(_ context.Context, args map[string]any) (any, error) {
		v, err := resolve(args)
		if err != nil {
			return nil, err
		}
		return v.ReadNameplate(str(args, "token"))
	}})
	addTool(s, toolDef{name: "sim.effector.list", description: "The declared effectors and their argument schemas.", schema: toolSchema("sim.effector.list"), handler: func(_ context.Context, args map[string]any) (any, error) {
		v, err := resolve(args)
		if err != nil {
			return nil, err
		}
		return v.ListEffectors(str(args, "token"))
	}})
	addTool(s, toolDef{name: "sim.effector.invoke", description: "Invoke an effector by name with a command_id (idempotency key) and capability token.", schema: toolSchema("sim.effector.invoke"), handler: func(_ context.Context, args map[string]any) (any, error) {
		v, err := resolve(args)
		if err != nil {
			return nil, err
		}
		return v.Invoke(str(args, "token"), str(args, "effector"), str(args, "entity_id"), str(args, "command_id"), mapArg(args, "args"), num(args, "at_ns", 0))
	}})
	addTool(s, toolDef{name: "sim.consumer.report", description: "Report quiescence and submit a consumer verdict. Write-only; never returns a score.", schema: toolSchema("sim.consumer.report"), handler: func(_ context.Context, args map[string]any) (any, error) {
		v, err := resolve(args)
		if err != nil {
			return nil, err
		}
		var verdict *model.Verdict
		if vd, ok := args["verdict"].(map[string]any); ok {
			b, err := json.Marshal(vd)
			if err != nil {
				return nil, errTool(CodeInvalidArgs, "%v", err)
			}
			verdict = &model.Verdict{}
			if err := json.Unmarshal(b, verdict); err != nil {
				return nil, errTool(CodeInvalidArgs, "%v", err)
			}
		}
		if err := v.Report(str(args, "token"), str(args, "run_id"), num(args, "quiesced_through_ns", 0), verdict); err != nil {
			return nil, fmt.Errorf("mcp: %w", err)
		}
		return map[string]any{"accepted": true, "world_id": v.WorldID, "simulated": true}, nil
	}})
	return s
}
