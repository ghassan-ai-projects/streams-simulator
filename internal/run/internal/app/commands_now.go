package app

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (r *Run) InjectFaultNow(entityID, faultID string, params map[string]any) (string, int64, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	atNS := r.World.Clock()
	faultInstance, err := r.injectFault(entityID, faultID, atNS, params)
	return faultInstance, atNS, err
}

func (r *Run) InvokeAcknowledgedNow(effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, int64, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	atNS := r.World.Clock()
	result, err := r.invokeEffector(effector, entityID, commandID, args, atNS)
	ack, err := acknowledgement(r, commandID, result, err, keepWorldRefusals)
	return ack, atNS, err
}
