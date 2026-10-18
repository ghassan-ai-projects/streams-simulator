package domain

import (
	"fmt"
	"maps"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const fallbackReply = `{"reason":"node_busy","result":"refused","state":{"approve":[],"devices":[],"holds":[],"pending":[]}}`

func DecodeControl(line []byte) (InterruptEntry, error) {
	if err := model.ValidateNodeControl(line); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrControlMalformed, err)
	}
	var doc map[string]any
	if err := decode(line, &doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrControlMalformed, err)
	}
	return controlEntry(doc), nil
}

func controlEntry(doc map[string]any) InterruptEntry {
	entry := newEntry(EventControl, "op", text(doc, "op"))
	if scope, ok := doc["scope"].(map[string]any); ok {
		entry["scope"] = Scope{Kind: text(scope, "kind"), ID: text(scope, "id")}.Fields()
	}
	if enabled, ok := doc["enabled"].(bool); ok {
		entry["enabled"] = enabled
	}
	if id := text(doc, "command_id"); id != "" {
		entry["command_id"] = id
	}
	return entry
}

func ControlReply(reply map[string]any) []byte {
	wire := maps.Clone(reply)
	delete(wire, "event")
	delete(wire, "clock")
	delete(wire, "node_ns")
	raw, err := canonical.Marshal(wire)
	if err != nil {
		return []byte(fallbackReply)
	}
	return raw
}

func RefusedReply(reason string, view map[string]any) map[string]any {
	return map[string]any{"result": ResultRefused, "reason": reason, "state": view}
}

func StateReply(view map[string]any) map[string]any {
	return map[string]any{"result": ResultApplied, "op": OpState, "state": view}
}
