package app

import (
	"context"
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (b *Bridge) callTool(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	toolArgs, err := action.ToolArguments(args)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	result, err := b.call(ctx, action.Tool, toolArgs)
	if err != nil {
		return nil, err
	}
	b.track(action.Name, args, result)
	return action.Withhold(result)
}

func (b *Bridge) call(ctx context.Context, tool string, args map[string]any) (json.RawMessage, error) {
	director, err := b.currentDirector(ctx)
	if err != nil {
		return nil, err
	}
	result, err := director.Call(ctx, tool, args)
	switch {
	case err != nil:
		return nil, err
	case result.Failure != "":
		return nil, domain.ToolFailure(result.Failure)
	}
	return result.Structured, nil
}

func (b *Bridge) currentDirector(ctx context.Context) (Director, error) {
	catalog := b.edges.Workspace.Fingerprint([]string{"domains"})
	if b.director != nil && (catalog == b.catalog || len(b.worlds) > 0) {
		return b.director, nil
	}
	b.stopDirector()
	director, err := b.edges.StartDirector(ctx)
	if err != nil {
		return nil, err
	}
	b.director, b.catalog = director, catalog
	return director, nil
}

func (b *Bridge) stopDirector() {
	if b.director != nil {
		_ = b.director.Close()
		b.director = nil
	}
}

func (b *Bridge) track(action string, args map[string]any, structured json.RawMessage) {
	switch action {
	case "world.create":
		var created struct {
			WorldID string `json:"world_id"`
		}
		if json.Unmarshal(structured, &created) == nil && created.WorldID != "" {
			b.worlds[created.WorldID] = true
		}
	case "world.destroy":
		if id, ok := args["world_id"].(string); ok {
			delete(b.worlds, id)
		}
	}
}
