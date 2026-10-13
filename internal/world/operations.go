package world

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func (w *World) Advance(to int64) (emitted, effectsApplied int, err error) {
	return w.world.Advance(to)
}

func (w *World) Clock() int64 {
	return w.world.Clock()
}

func (w *World) StreamKey() string {
	return w.world.StreamKey()
}

func (w *World) NextEventNS() int64 {
	return w.world.NextEventNS()
}

func (w *World) SetEmitter(emitter func(model.SimEvent)) {
	w.world.SetEmitter(emitter)
}

func (w *World) EmittedCount() int64 {
	return w.world.EmittedCount()
}

func (w *World) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*InvokeResult, error) {
	return w.world.InvokeEffector(effector, entityID, commandID, args, atNS)
}

func (w *World) EffectorCalls() []EffectorCall {
	return w.world.EffectorCalls()
}

func (w *World) SetFailureMode(mode FailureMode) {
	w.world.SetFailureMode(mode)
}

func (w *World) PendingKicks() int {
	return w.world.PendingKicks()
}

func (w *World) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	return w.world.InjectFault(entityID, faultID, onsetNS, params)
}

func (w *World) ClearFault(faultID string, atNS int64) error {
	return w.world.ClearFault(faultID, atNS)
}

func (w *World) ListFaults() []FaultInfo {
	return w.world.ListFaults()
}

func (w *World) ActiveFaultsCount() int {
	return w.world.ActiveFaultsCount()
}

func (w *World) Reading(entityID, channelName string, t int64) float64 {
	return w.world.Reading(entityID, channelName, t)
}

func (w *World) ReadingNoise(entityID, channelName string, t int64) float64 {
	return w.world.ReadingNoise(entityID, channelName, t)
}

func (w *World) ReadingDropped(entityID, channelName string, t int64) bool {
	return w.world.ReadingDropped(entityID, channelName, t)
}

func (w *World) KPITruth() *model.KPITruth {
	return w.world.KPITruth()
}

func (w *World) StateValue(entity, state string, t int64) float64 {
	return w.world.StateValue(entity, state, t)
}

func (w *World) ViolatedBoundaries(entityID string, t int64) []string {
	return w.world.ViolatedBoundaries(entityID, t)
}

func (w *World) EntityIDs() []string {
	return w.world.EntityIDs()
}

func (w *World) PeerIDs(entityID string) []string {
	return w.world.PeerIDs(entityID)
}

func (w *World) InitialEntityIDs() []string {
	return w.world.InitialEntityIDs()
}

func (w *World) AddEntity(id string, atNS int64) error {
	return w.world.AddEntity(id, atNS)
}

func (w *World) Retire(entityID, reason string, atNS int64) {
	w.world.Retire(entityID, reason, atNS)
}

func (w *World) Entity(id string) *Entity {
	return w.world.Entity(id)
}

func (w *World) CheckEffectorArgs(effector string, args map[string]any) error {
	return w.world.CheckEffectorArgs(effector, args)
}

func (w *World) CheckFaultParams(faultID string, params map[string]any) error {
	return w.world.CheckFaultParams(faultID, params)
}

func (w *World) StateNamesOf(entityID string) []string {
	return w.world.StateNamesOf(entityID)
}

func (w *World) EffectorOf(entityID, name string) *model.Effector {
	return w.world.EffectorOf(entityID, name)
}
