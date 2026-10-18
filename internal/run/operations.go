package run

import (
	"context"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/app"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func FolderUsed(dir string) bool {
	return layer.FolderUsed(dir)
}

func LoadArtifact(path string) (*model.RunArtifact, error) {
	return layer.LoadArtifact(path)
}

func LoadArtifactForVerification(path string) (*model.RunArtifact, error) {
	return layer.LoadArtifactForVerification(path)
}

func Reproduced(res *ReplayResult) bool {
	return layer.Reproduced(res)
}

func ReplayArtifact(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter, sinkTarget string) (*ReplayResult, error) {
	return layer.ReplayArtifact(ctx, art, spec, adapterSpec, sinkTarget)
}

func ReplayArtifactEvidence(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adapterSpec *model.Adapter) (*ReplayEvidence, error) {
	return layer.ReplayArtifactEvidence(ctx, art, spec, adapterSpec)
}

func (r *Run) Advance(ctx context.Context, toNS int64, awaitConsumer bool) (int, error) {
	return r.run.Advance(ctx, toNS, awaitConsumer)
}

func (r *Run) InjectFault(entityID, faultID string, onsetNS int64, params map[string]any) (string, error) {
	return r.run.InjectFault(entityID, faultID, onsetNS, params)
}

func (r *Run) ClearFault(faultID string, atNS int64) error {
	return r.run.ClearFault(faultID, atNS)
}

func (r *Run) ApplyPerturb(name string, params map[string]any, fromNS, untilNS int64) (string, error) {
	return r.run.ApplyPerturb(name, params, fromNS, untilNS)
}

func (r *Run) ClearPerturb(id string) error {
	return r.run.ClearPerturb(id)
}

func (r *Run) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*world.InvokeResult, error) {
	return r.run.InvokeEffector(effector, entityID, commandID, args, atNS)
}

func (r *Run) InvokeEffectorNow(effector, entityID, commandID string, args map[string]any) (*world.InvokeResult, error) {
	return r.run.InvokeEffectorNow(effector, entityID, commandID, args)
}

func (r *Run) RetireEntity(entityID, reason string, atNS int64) error {
	return r.run.RetireEntity(entityID, reason, atNS)
}

func (r *Run) EnvInject(target, fault string, params map[string]any, atNS int64) (string, error) {
	return r.run.EnvInject(target, fault, params, atNS)
}

func (r *Run) ReportQuiesced(throughNS int64) {
	r.run.ReportQuiesced(throughNS)
}

func (r *Run) SubmitVerdict(v *model.Verdict) error {
	return r.run.SubmitVerdict(v)
}

func (r *Run) Verdict() *model.Verdict {
	return r.run.Verdict()
}

func (r *Run) Ledger() []model.LedgerRecord {
	return r.run.Ledger()
}

func (r *Run) History() []model.StateSnapshot {
	return r.run.History()
}

func (r *Run) AppliedPerturbations() []string {
	return r.run.AppliedPerturbations()
}

func (r *Run) Domain() *domain.Compiled {
	return r.run.Domain()
}

func (r *Run) Digest() string {
	return r.run.Digest()
}

func (r *Run) Reproducible() bool {
	return r.run.Reproducible()
}

func (r *Run) Unblind() {
	r.run.Unblind()
}

func (r *Run) UnblindedStamp() bool {
	return r.run.UnblindedStamp()
}

func (r *Run) Finished() bool {
	return r.run.Finished()
}

func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	return r.run.End(outDir)
}

func (r *Run) Abandon(reason string) {
	r.run.Abandon(reason)
}

func (r *Run) Trace() []byte {
	return r.run.Trace()
}

func (r *Run) SetEvidenceRecorder(fn func(model.SimEvent)) {
	r.run.SetEvidenceRecorder(fn)
}

func (r *Run) JournalWire(direction string, record map[string]any) {
	r.run.JournalWire(direction, record)
}

func (r *Run) JournalPacing(throughSeq uint64, waitedMS int64, missed bool) {
	r.run.JournalPacing(throughSeq, waitedMS, missed)
}

func (r *Run) SetDeliveryObserver(fn func(model.SimEvent)) {
	r.run.SetDeliveryObserver(fn)
}

func (r *Run) SetQuiesceParkedHook(h func()) {
	r.run.SetQuiesceParkedHook(h)
}

func (r *Run) SetFailureMode(mode world.FailureMode) {
	r.run.SetFailureMode(mode)
}

func (r *Run) Status() WorldStatus {
	return r.run.Status()
}

func (r *Run) Faults() []world.FaultInfo {
	return r.run.Faults()
}

func (r *Run) ExecuteScenario(ctx context.Context, scenario *model.Scenario) error {
	return r.run.ExecuteScenario(ctx, scenario)
}

func LoadInputs(domainsDir, adaptersDir, domainID, adapterID string) (*domain.Compiled, *model.Adapter, error) {
	return layer.LoadInputs(domainsDir, adaptersDir, domainID, adapterID)
}

func ResolveInputs(art *model.RunArtifact, dirs Directories) (*domain.Compiled, *model.Adapter, error) {
	return layer.ResolveInputs(art, dirs)
}

func ExportScenarioFile(ctx context.Context, artifactPath string, dirs Directories) (ScenarioExport, error) {
	return layer.ExportScenarioFile(ctx, artifactPath, dirs)
}

func ReplayFile(ctx context.Context, artifactPath string, dirs Directories) (*ReplayResult, error) {
	return layer.ReplayFile(ctx, artifactPath, dirs)
}

func RunScenario(ctx context.Context, scenario *model.Scenario, options ScenarioRunOptions) (RunSummary, error) {
	return layer.RunScenario(ctx, scenario, options)
}

func CheckDomain(check model.DomainCheck, spec *model.DomainSpec) ([]model.ImplementsChecked, error) {
	return layer.CheckDomain(check, spec)
}
