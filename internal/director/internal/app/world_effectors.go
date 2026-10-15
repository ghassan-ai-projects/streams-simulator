package app

import (
	"strconv"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (d *Director) WorldEffectors(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	effectors, err := w.Operator.ListEffectors(w.Token)
	if err != nil {
		return nil, err
	}
	return map[string]any{"world_id": worldID, "effectors": effectors, "simulated": true}, nil
}

func (d *Director) InvokeWorldEffector(worldID, effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, rules.ToolErrorf(rules.CodeWorldNotFound, "unknown world %q", worldID)
	}
	if commandID == "" {
		commandID = d.nextCommandID(w)
	}
	return w.Operator.Invoke(w.Token, effector, entityID, commandID, args)
}

func (d *Director) nextCommandID(w *WorldRecord) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	w.DirectorCommands++
	return "director-" + strconv.Itoa(w.DirectorCommands)
}
