package director

import (
	"context"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func (d *Director) CreateWorld(req WorldRequest) (map[string]any, error) {
	return d.impl.CreateWorld(req)
}

func (d *Director) DescribeWorld(worldID string) (map[string]any, error) {
	return d.impl.DescribeWorld(worldID)
}

func (d *Director) DestroyWorld(worldID string) (map[string]any, error) {
	return d.impl.DestroyWorld(worldID)
}

func (d *Director) Advance(ctx context.Context, worldID string, toNS int64, await bool) (map[string]any, error) {
	return d.impl.Advance(ctx, worldID, toNS, await)
}

func (d *Director) AdvanceBy(ctx context.Context, worldID string, byNS int64, await bool) (map[string]any, error) {
	return d.impl.AdvanceBy(ctx, worldID, byNS, await)
}

func (d *Director) ClockState(worldID string) (map[string]any, error) {
	return d.impl.ClockState(worldID)
}

func (d *Director) InjectFault(worldID, entityID, fault string, onsetNS int64, params map[string]any) (map[string]any, error) {
	return d.impl.InjectFault(worldID, entityID, fault, onsetNS, params)
}

func (d *Director) ClearFault(worldID, faultID string) (map[string]any, error) {
	return d.impl.ClearFault(worldID, faultID)
}

func (d *Director) ListFaults(worldID string) (map[string]any, error) {
	return d.impl.ListFaults(worldID)
}

func (d *Director) ApplyPerturb(worldID, name string, params map[string]any, fromNS, untilNS int64) (map[string]any, error) {
	return d.impl.ApplyPerturb(worldID, name, params, fromNS, untilNS)
}

func (d *Director) ClearPerturb(worldID, id string) (map[string]any, error) {
	return d.impl.ClearPerturb(worldID, id)
}

func (d *Director) EnvInject(worldID, target, fault string, params map[string]any) (map[string]any, error) {
	return d.impl.EnvInject(worldID, target, fault, params)
}

func (d *Director) RetireEntity(worldID, entityID, reason string) (map[string]any, error) {
	return d.impl.RetireEntity(worldID, entityID, reason)
}

func (d *Director) BeginRun(worldID, label string) (map[string]any, error) {
	return d.impl.BeginRun(worldID, label)
}

func (d *Director) EndRun(worldID string) (map[string]any, error) { return d.impl.EndRun(worldID) }

func (d *Director) SealTruth(runID string, rec *model.GroundTruthRecord) error {
	return d.impl.SealTruth(runID, rec)
}

func (d *Director) RevealTruth(runID string, unblind bool) (map[string]any, error) {
	return d.impl.RevealTruth(runID, unblind)
}

func (d *Director) SealStatus(runID string) (map[string]any, error) { return d.impl.SealStatus(runID) }

func (d *Director) Score(runID string) (map[string]any, error) { return d.impl.Score(runID) }

func (d *Director) VerifyRun(artifactPath string) (map[string]any, error) {
	return d.impl.VerifyRun(artifactPath)
}

func (d *Director) AuditScenario(domainID, entityID, fault string, onsetNS, startNS, durationNS int64) (map[string]any, error) {
	return d.impl.AuditScenario(domainID, entityID, fault, onsetNS, startNS, durationNS)
}

func (d *Director) CatalogList(group string) map[string]any { return d.impl.CatalogList(group) }

func (d *Director) CatalogDescribe(domainID string) (map[string]any, error) {
	return d.impl.CatalogDescribe(domainID)
}

func (d *Director) CatalogCoverage() any { return d.impl.CatalogCoverage() }

func (d *Director) DomainSpec(domainID string) (any, error) { return d.impl.DomainSpec(domainID) }

func (d *Director) ListAdapters() map[string]any { return d.impl.ListAdapters() }

func (d *Director) LoadLibrary(dirs []string) error { return d.impl.LoadLibrary(dirs) }

func (d *Director) ComponentList() (map[string]any, error) { return d.impl.ComponentList() }

func (d *Director) ComponentDescribe(ref string) (any, error) { return d.impl.ComponentDescribe(ref) }

func (d *Director) ComposeDomain(composition map[string]any, install bool) (map[string]any, error) {
	return d.impl.ComposeDomain(composition, install)
}

func (d *Director) SetOperatorEndpoint(endpoint string) { d.impl.SetOperatorEndpoint(endpoint) }

func (d *Director) RequireOperator(token string) error { return d.impl.RequireOperator(token) }

func (d *Director) OperatorNameplate(token string) (*Nameplate, error) {
	return d.impl.OperatorNameplate(token)
}

func (d *Director) OperatorEffectors(token string) ([]EffectorInfo, error) {
	return d.impl.OperatorEffectors(token)
}

func (d *Director) OperatorInvoke(token, effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, error) {
	return d.impl.OperatorInvoke(token, effector, entityID, commandID, args)
}

func (d *Director) OperatorReport(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) (string, error) {
	return d.impl.OperatorReport(token, runID, quiescedThroughNS, verdict)
}

func ToolErrorf(code Code, format string, args ...any) error {
	return rules.ToolErrorf(code, format, args...)
}

func (d *Director) RunOf(worldID string) *run.Run {
	return d.impl.RunOf(worldID)
}
