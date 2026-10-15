package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func (b *Bridge) callTool(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	toolArgs, err := action.ToolArguments(args)
	if err != nil {
		return nil, err
	}
	b.directorMu.Lock()
	defer b.directorMu.Unlock()
	result, err := b.call(ctx, action.Tool, toolArgs)
	if err != nil {
		return nil, err
	}
	b.track(action.Name, args, result)
	return action.Withhold(result)
}

func (b *Bridge) call(ctx context.Context, tool string, args map[string]any) (json.RawMessage, error) {
	result, err := b.callOnce(ctx, tool, args)
	if errors.Is(err, process.ErrGone) {
		b.dropDirector()
		return b.callOnce(ctx, tool, args)
	}
	return result, err
}

func (b *Bridge) callOnce(ctx context.Context, tool string, args map[string]any) (json.RawMessage, error) {
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
	if b.director != nil && (catalog == b.catalog || b.runLive()) {
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

func (b *Bridge) dropDirector() {
	b.stopDirector()
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	clear(b.running)
}

func (b *Bridge) runLive() bool {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return len(b.running) > 0
}

func (b *Bridge) runIsRunning(runID string) bool {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	return b.running[runID]
}

func (b *Bridge) track(action string, args map[string]any, structured json.RawMessage) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	id, _ := args["world_id"].(string)
	switch action {
	case "world.create":
		b.trackCreated(structured)
	case "run.end":
		delete(b.running, id)
	case "world.destroy":
		delete(b.running, id)
	}
}

func (b *Bridge) trackCreated(structured json.RawMessage) {
	var created struct {
		WorldID string `json:"world_id"`
	}
	if json.Unmarshal(structured, &created) == nil && created.WorldID != "" {
		b.running[created.WorldID] = true
	}
}
