package director

import (
	"context"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// CreateWorld creates a world from a typed request (sim.world.create).
func (d *Director) CreateWorld(req WorldRequest) (map[string]any, error) {
	return d.impl.CreateWorld(req)
}

// DescribeWorld reports a world's config, digest, clock and emitted count.
func (d *Director) DescribeWorld(worldID string) (map[string]any, error) {
	return d.impl.DescribeWorld(worldID)
}

// DestroyWorld finalizes and removes a world.
func (d *Director) DestroyWorld(worldID string) (map[string]any, error) {
	return d.impl.DestroyWorld(worldID)
}

// Advance moves the clock to an absolute time; await blocks on consumer quiescence.
func (d *Director) Advance(ctx context.Context, worldID string, toNS int64, await bool) (map[string]any, error) {
	return d.impl.Advance(ctx, worldID, toNS, await)
}

// AdvanceBy moves the clock forward by a relative delta; await blocks on consumer quiescence.
func (d *Director) AdvanceBy(ctx context.Context, worldID string, byNS int64, await bool) (map[string]any, error) {
	return d.impl.AdvanceBy(ctx, worldID, byNS, await)
}

// ClockState reports the clock, the next scheduled event and the pending effects.
func (d *Director) ClockState(worldID string) (map[string]any, error) {
	return d.impl.ClockState(worldID)
}

// InjectFault applies a world fault to an entity.
func (d *Director) InjectFault(worldID, entityID, fault string, onsetNS int64, params map[string]any) (map[string]any, error) {
	return d.impl.InjectFault(worldID, entityID, fault, onsetNS, params)
}

// ClearFault clears a fault by id.
func (d *Director) ClearFault(worldID, faultID string) (map[string]any, error) {
	return d.impl.ClearFault(worldID, faultID)
}

// ListFaults lists the active faults with ids and onsets (director only).
func (d *Director) ListFaults(worldID string) (map[string]any, error) {
	return d.impl.ListFaults(worldID)
}

// ApplyPerturb activates a delivery perturbation.
func (d *Director) ApplyPerturb(worldID, name string, params map[string]any, fromNS, untilNS int64) (map[string]any, error) {
	return d.impl.ApplyPerturb(worldID, name, params, fromNS, untilNS)
}

// ClearPerturb deactivates a perturbation by id.
func (d *Director) ClearPerturb(worldID, id string) (map[string]any, error) {
	return d.impl.ClearPerturb(worldID, id)
}

// EnvInject records an environment fault against a configured target.
func (d *Director) EnvInject(worldID, target, fault string, params map[string]any) (map[string]any, error) {
	return d.impl.EnvInject(worldID, target, fault, params)
}

// RetireEntity retires an entity at the world's current clock.
func (d *Director) RetireEntity(worldID, entityID, reason string) (map[string]any, error) {
	return d.impl.RetireEntity(worldID, entityID, reason)
}

// BeginRun opens a run once its ground truth is sealed.
func (d *Director) BeginRun(worldID, label string) (map[string]any, error) {
	return d.impl.BeginRun(worldID, label)
}

// EndRun finalizes the run and writes its artifact.
func (d *Director) EndRun(worldID string) (map[string]any, error) { return d.impl.EndRun(worldID) }

// SealTruth installs the director-only ground-truth record before a run is opened.
func (d *Director) SealTruth(runID string, rec *model.GroundTruthRecord) error {
	return d.impl.SealTruth(runID, rec)
}

// RevealTruth returns the sealed label; unblind permits revealing on an open run and stamps it.
func (d *Director) RevealTruth(runID string, unblind bool) (map[string]any, error) {
	return d.impl.RevealTruth(runID, unblind)
}

// SealStatus reports whether a run's truth is sealed and whether it was unblinded.
func (d *Director) SealStatus(runID string) (map[string]any, error) { return d.impl.SealStatus(runID) }

// Score computes the scorecard for a closed run's submitted verdict.
func (d *Director) Score(runID string) (map[string]any, error) { return d.impl.Score(runID) }

// VerifyRun replays a run artifact and compares its digests.
func (d *Director) VerifyRun(artifactPath string) (map[string]any, error) {
	return d.impl.VerifyRun(artifactPath)
}

// AuditScenario runs the trivial-baseline audit on one injection.
func (d *Director) AuditScenario(domainID, entityID, fault string, onsetNS, startNS, durationNS int64) (map[string]any, error) {
	return d.impl.AuditScenario(domainID, entityID, fault, onsetNS, startNS, durationNS)
}

// CatalogList lists the installed domains, optionally one group.
func (d *Director) CatalogList(group string) map[string]any { return d.impl.CatalogList(group) }

// CatalogDescribe describes one domain: its spec and digest.
func (d *Director) CatalogDescribe(domainID string) (map[string]any, error) {
	return d.impl.CatalogDescribe(domainID)
}

// CatalogCoverage reports the axis-coverage matrix.
func (d *Director) CatalogCoverage() any { return d.impl.CatalogCoverage() }

// DomainSpec returns the declared spec of one domain.
func (d *Director) DomainSpec(domainID string) (any, error) { return d.impl.DomainSpec(domainID) }

// ListAdapters lists the installed output adapters in id order.
func (d *Director) ListAdapters() map[string]any { return d.impl.ListAdapters() }

// SetOperatorEndpoint records the operator HTTP endpoint this process serves; world creation returns it.
func (d *Director) SetOperatorEndpoint(endpoint string) { d.impl.SetOperatorEndpoint(endpoint) }

// RequireOperator refuses a capability token that names no world.
func (d *Director) RequireOperator(token string) error { return d.impl.RequireOperator(token) }

// OperatorNameplate returns the static nameplate of the token's world.
func (d *Director) OperatorNameplate(token string) (*Nameplate, error) {
	return d.impl.OperatorNameplate(token)
}

// OperatorEffectors returns the effectors the token's world declares.
func (d *Director) OperatorEffectors(token string) ([]EffectorInfo, error) {
	return d.impl.OperatorEffectors(token)
}

// OperatorInvoke actuates an effector of the token's world; the consumer gets the acknowledgement only.
func (d *Director) OperatorInvoke(token, effector, entityID, commandID string, args map[string]any) (*model.EffectorAcknowledgement, error) {
	return d.impl.OperatorInvoke(token, effector, entityID, commandID, args)
}

// OperatorReport records the consumer's quiescence and verdict; returns the world id and never a score.
func (d *Director) OperatorReport(token, runID string, quiescedThroughNS int64, verdict *model.Verdict) (string, error) {
	return d.impl.OperatorReport(token, runID, quiescedThroughNS, verdict)
}

// ToolErrorf builds a coded refusal for a surface's own checks.
func ToolErrorf(code Code, format string, args ...any) error {
	return rules.ToolErrorf(code, format, args...)
}

// RunOf is the test harness's handle on the run behind a world, nil for an
// unknown world. Composition tests (test/acceptance) use it to observe and
// park the run beneath the tools, as run.SetEvidenceRecorder and
// run.SetQuiesceParkedHook are used; no surface may. The MCP surface cannot
// reach it either: it may not import the run module (TestMCPHoldsOnlyProtocol),
// so it cannot name the type RunOf returns.
func (d *Director) RunOf(worldID string) *run.Run {
	return d.impl.RunOf(worldID)
}
