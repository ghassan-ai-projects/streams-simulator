package run

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// LoadArtifact reads and validates a run artifact from path.
func LoadArtifact(path string) (*model.RunArtifact, error) {
	return layer.LoadArtifact(path)
}

// ReplayArtifact reproduces a run from its artifact with no server running
// and reports whether the trace digest matched.
func ReplayArtifact(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, error) {
	return layer.ReplayArtifact(ctx, art, spec, adapterSpec, sinkTarget)
}

// ReplayArtifactEvidence replays an artifact like ReplayArtifact and also
// returns the effector calls the replayed world made.
func ReplayArtifactEvidence(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (*ReplayEvidence, error) {
	return layer.ReplayArtifactEvidence(ctx, art, spec, adapterSpec)
}

// Advance moves the world to toNS, delivering what it emits through the
// perturbation layer, adapter and sink; with awaitConsumer it then waits for
// the consumer to quiesce.
func (r *Run) Advance(ctx context.Context, toNS int64, awaitConsumer bool) (int, error) {
	return r.run.Advance(ctx, toNS, awaitConsumer)
}

// InjectFault applies a declared fault and records the command.
func (r *Run) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	return r.run.InjectFault(entityID, faultID, onsetNS, params)
}

// ClearFault ends an injected fault and records the command.
func (r *Run) ClearFault(faultID string, atNS int64) error {
	return r.run.ClearFault(faultID, atNS)
}

// ApplyPerturb activates a delivery perturbation and records the command.
func (r *Run) ApplyPerturb(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	return r.run.ApplyPerturb(name, params, fromNS, untilNS)
}

// ClearPerturb deactivates a perturbation and records the command.
func (r *Run) ClearPerturb(id string) error {
	return r.run.ClearPerturb(id)
}

// InvokeEffector actuates a world effector and records the command.
func (r *Run) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	return r.run.InvokeEffector(effector, entityID, commandID, args, atNS)
}

// RetireEntity retires an entity and records the command.
func (r *Run) RetireEntity(entityID, reason string, atNS int64) error {
	return r.run.RetireEntity(entityID, reason, atNS)
}

// EnvInject injects an environment fault where the run allows it.
func (r *Run) EnvInject(target, fault string, params map[string]any, atNS int64) (string, error) {
	return r.run.EnvInject(target, fault, params, atNS)
}

// ReportQuiesced tells the run the consumer has processed everything through
// throughNS.
func (r *Run) ReportQuiesced(throughNS int64) {
	r.run.ReportQuiesced(throughNS)
}

// SubmitVerdict records the consumer's verdict for the run.
func (r *Run) SubmitVerdict(v *model.Verdict) error {
	return r.run.SubmitVerdict(v)
}

// Verdict returns a copy of the submitted verdict, nil when none.
func (r *Run) Verdict() *model.Verdict {
	return r.run.Verdict()
}

// Ledger returns a copy of the delivery ledger.
func (r *Run) Ledger() []model.LedgerRecord {
	return r.run.Ledger()
}

// History returns the director-only hidden-state history.
func (r *Run) History() []model.StateSnapshot {
	return r.run.History()
}

// AppliedPerturbations lists the perturbation applications of the run.
func (r *Run) AppliedPerturbations() []string {
	return r.run.AppliedPerturbations()
}

// Domain returns the compiled domain the run was built from.
func (r *Run) Domain() *domain.Compiled {
	return r.run.Domain()
}

// Digest returns the world digest of the run's determinism tuple.
func (r *Run) Digest() string {
	return r.run.Digest()
}

// Reproducible reports whether the run is byte-reproducible by replay.
func (r *Run) Reproducible() bool {
	return r.run.Reproducible()
}

// Unblind stamps the run as unblinded, excluding it from every scorecard.
func (r *Run) Unblind() {
	r.run.Unblind()
}

// UnblindedStamp reports whether the run was unblinded.
func (r *Run) UnblindedStamp() bool {
	return r.run.UnblindedStamp()
}

// End closes the run, publishes its evidence under outDir when set, and
// returns the run artifact.
func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	return r.run.End(outDir)
}

// Trace returns a copy of the delivered trace bytes (test harness hook).
func (r *Run) Trace() []byte {
	return r.run.Trace()
}

// SetEvidenceRecorder installs a per-event delivery hook (test harness hook).
func (r *Run) SetEvidenceRecorder(fn func(model.SimEvent)) {
	r.run.SetEvidenceRecorder(fn)
}

// SetQuiesceParkedHook installs a hook fired when a quiescence wait blocks
// (test harness hook).
func (r *Run) SetQuiesceParkedHook(h func()) {
	r.run.SetQuiesceParkedHook(h)
}

// SetFailureMode overrides the effector failure-mode selection (test hook).
func (r *Run) SetFailureMode(mode string) {
	r.run.SetFailureMode(mode)
}
