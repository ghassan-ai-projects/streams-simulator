package world

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

// Advance processes every event scheduled at or before to, then sets the
// clock to to, and reports how many native events were emitted and how many
// effect kicks started. Moving the clock backwards is refused.
func (w *World) Advance(to int64) (emitted, effectsApplied int, err error) {
	return w.world.Advance(to)
}

// Clock is the current world time.
func (w *World) Clock() int64 {
	return w.world.Clock()
}

// NextEventNS is the time of the next scheduled event, 0 when none.
func (w *World) NextEventNS() int64 {
	return w.world.NextEventNS()
}

// SetEmitter installs the event sink. Events flow world -> emitter.
func (w *World) SetEmitter(emitter func(model.SimEvent)) {
	w.world.SetEmitter(emitter)
}

// EmittedCount is the total number of native events produced.
func (w *World) EmittedCount() int64 {
	return w.world.EmittedCount()
}

// InvokeEffector actuates an effector. commandID is the idempotency key;
// repeating a call with the same command_id inside the declared window
// returns the original result and applies no second effect.
func (w *World) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*InvokeResult, error) {
	return w.world.InvokeEffector(effector, entityID, commandID, args, atNS)
}

// EffectorCalls returns the effector call log (the authority when scoring
// actions).
func (w *World) EffectorCalls() []EffectorCall {
	return w.world.EffectorCalls()
}

// SetFailureMode overrides the failure-mode selection for subsequent
// invocations ("" restores the declared distribution).
func (w *World) SetFailureMode(mode string) {
	w.world.SetFailureMode(mode)
}

// PendingKicks is the number of scheduled-but-unapplied effect kicks.
func (w *World) PendingKicks() int {
	return w.world.PendingKicks()
}

// InjectFault applies a declared fault to an entity at onsetNS.
func (w *World) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	return w.world.InjectFault(entityID, faultID, onsetNS, params)
}

// ClearFault ends an injected fault instance at atNS.
func (w *World) ClearFault(faultID string, atNS int64) error {
	return w.world.ClearFault(faultID, atNS)
}

// ListFaults lists the active fault instances in injection order.
func (w *World) ListFaults() []FaultInfo {
	return w.world.ListFaults()
}

// ActiveFaultsCount is the number of fault instances still in force.
func (w *World) ActiveFaultsCount() int {
	return w.world.ActiveFaultsCount()
}

// Reading is the noise-free value of a channel at time t (director only).
func (w *World) Reading(entityID, channelName string, t int64) float64 {
	return w.world.Reading(entityID, channelName, t)
}

// StateValue exposes a hidden state to the director only (solver and truth).
func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.world.StateValue(entity, state, t)
}

// EntityIDs returns the live entity ids in creation order.
func (w *World) EntityIDs() []string {
	return w.world.EntityIDs()
}

// InitialEntityIDs returns the entities created at world construction, in
// order (before any churn or entity.add).
func (w *World) InitialEntityIDs() []string {
	return w.world.InitialEntityIDs()
}

// AddEntity creates an entity at the given time.
func (w *World) AddEntity(id string, atNS int64) error {
	return w.world.AddEntity(id, atNS)
}

// Retire removes an entity and its scheduled emissions.
func (w *World) Retire(entityID, reason string, atNS int64) {
	w.world.Retire(entityID, reason, atNS)
}

// Entity returns the entity by id, or nil.
func (w *World) Entity(id string) *Entity {
	return w.world.Entity(id)
}
