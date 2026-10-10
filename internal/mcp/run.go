package mcp

import (
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// BeginRun opens a run after the director has sealed its ground-truth record.
// Truth is intentionally a separate operation because the complete label is
// generated from the selected scenario and fault, not from the human label
// carried by run.begin.
func (d *Director) BeginRun(worldID, label string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	if w.Started {
		return nil, errTool(CodeDomainInvalid, "a run is already open for %q", worldID)
	}
	if err := d.requireSealedRun(w.Run.ID); err != nil {
		return nil, err
	}
	w.Started = true
	return map[string]any{"run_id": w.Run.ID, "truth_sealed": true}, nil
}

// SealTruth installs the director-only ground-truth record before a run is
// opened. The record is copied and cannot be mutated through this pointer.
func (d *Director) SealTruth(runID string, rec *model.GroundTruthRecord) error {
	w := d.worldByRun(runID)
	if w == nil {
		return errTool(CodeWorldNotFound, "unknown run %q", runID)
	}
	if w.Started || w.RunEnded {
		return errTool(CodeTruthSealed, "truth must be sealed before run.begin")
	}
	if err := d.Truth.Seal(runID, rec); err != nil {
		return errTool(CodeTruthSealed, "%v", err)
	}
	return nil
}

// EndRun finalizes the run and writes the artifact.
func (d *Director) EndRun(worldID string) (map[string]any, error) {
	w := d.World(worldID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown world %q", worldID)
	}
	return d.finalizeRun(w, worldID)
}

// RevealTruth returns the sealed label; unblind permits revealing on an
// open run, permanently stamping it.
func (d *Director) RevealTruth(runID string, unblind bool) (map[string]any, error) {
	rec, err := d.Truth.Reveal(runID, unblind)
	if err != nil {
		return nil, errTool(CodeTruthSealed, "%v", err)
	}
	if unblind {
		if w := d.worldByRun(runID); w != nil {
			w.Run.Unblind()
		}
	}
	return map[string]any{"ground_truth": rec}, nil
}

// SealStatus reports sealing state.
func (d *Director) SealStatus(runID string) (map[string]any, error) {
	sealed, unblinded, err := d.Truth.SealStatus(runID)
	if err != nil {
		return nil, errTool(CodeWorldNotFound, "%v", err)
	}
	return map[string]any{"sealed": sealed, "unblinded": unblinded}, nil
}

// Score computes the scorecard for a run's submitted verdict.
func (d *Director) Score(runID string) (map[string]any, error) {
	w := d.worldByRun(runID)
	if w == nil {
		return nil, errTool(CodeWorldNotFound, "unknown run %q", runID)
	}
	if w.Run.UnblindedStamp() {
		return nil, errTool(CodeRunUnblinded, "scoring refused; the run is stamped unblinded")
	}
	return d.scoreClosedRun(w, runID)
}

// VerifyRun replays a run artifact and compares digests.
func (d *Director) VerifyRun(artifactPath string) (map[string]any, error) {
	art, err := run.LoadArtifact(artifactPath)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	spec, err := d.Catalog.Describe(art.Domain.ID)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	adap, ok := d.Adapters[art.Adapter.ID]
	if !ok {
		return nil, errTool(CodeAdapterInvalid, "unknown adapter %q", art.Adapter.ID)
	}
	return d.verifyArtifact(art, spec, adap)
}

// AuditScenario runs the trivial-baseline audit on one injection.
func (d *Director) AuditScenario(domainID, entityID, fault string, onsetNS, startNS int64, durationNS int64) (map[string]any, error) {
	spec, panel, err := d.auditPanel(domainID)
	if err != nil {
		return nil, err
	}
	v, err := panel.Audit(entityID, fault, onsetNS, startNS, auditEntityIDs(spec), durationNS, nil, nil)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{"trivial": v.Trivial, "scores": v.Scores, "best": v.Best}, nil
}

func (d *Director) auditPanel(domainID string) (*domain.Compiled, *audit.Panel, error) {
	spec, err := d.Catalog.Describe(domainID)
	if err != nil {
		return nil, nil, errTool(CodeDomainInvalid, "%v", err)
	}
	panel, err := audit.NewPanel(spec, 1, 60*1e9)
	if err != nil {
		return nil, nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return spec, panel, nil
}

func (d *Director) worldByRun(runID string) *WorldRecord {
	d.mu.Lock()
	defer d.mu.Unlock()
	worldID, ok := d.byRun[runID]
	if !ok {
		return nil
	}
	return d.Worlds[worldID]
}

func (d *Director) requireSealedRun(runID string) error {
	sealed, unblinded, err := d.Truth.SealStatus(runID)
	if err != nil || !sealed {
		return errTool(CodeTruthSealed, "ground truth must be sealed before run.begin")
	}
	if unblinded {
		return errTool(CodeRunUnblinded, "run is already stamped unblinded")
	}
	return nil
}

func (d *Director) finalizeRun(w *WorldRecord, worldID string) (map[string]any, error) {
	dir := filepath.Join(d.OutDir, worldID)
	art, err := w.Run.End(dir)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	w.RunEnded = true
	return map[string]any{
		"run_artifact_path": filepath.Join(dir, "run.json"),
		"trace_digest":      art.ExpectedTraceDigest,
		"reproducible":      art.Reproducible,
	}, nil
}

func (d *Director) scoreClosedRun(w *WorldRecord, runID string) (map[string]any, error) {
	gt, err := d.Truth.Reveal(runID, false)
	if err != nil {
		return nil, errTool(CodeTruthSealed, "%v", err)
	}
	sc, err := score.Score(scoringEvidence(w.Run), gt)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	if !w.Run.Reproducible() {
		sc.Detail = "non-reproducible run (wall-clock delivery or consumer not quiesced); hash-equality metrics refused"
	}
	return map[string]any{"scorecard": sc}, nil
}

func (d *Director) verifyArtifact(art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter) (map[string]any, error) {
	res, err := run.ReplayArtifact(d.ctx, art, spec, adap, "")
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	return map[string]any{
		"matches": res.Matches, "version_match": res.VersionMatch,
		"first_divergence": res.FirstDivergence, "detail": res.Detail,
	}, nil
}

func auditEntityIDs(spec *domain.Compiled) []string {
	var ids []string
	n := spec.Spec.Entities.Count.Default
	// Entity ids come from the domain's own template — the binary contains
	// no domain literal.
	tmpl := spec.Spec.Entities.IDTemplate
	for i := 1; i <= n; i++ {
		ids = append(ids, world.RenderID(tmpl, i))
	}
	return ids
}

// scoringEvidence packs a closed run into the scorer's input.
func scoringEvidence(r *run.Run) score.Evidence {
	return score.Evidence{
		RunID: r.ID, Domain: r.Domain(), Verdict: r.Verdict(), Ledger: r.Ledger(),
		Calls: r.World.EffectorCalls(), Perturbations: r.AppliedPerturbations(),
		Emitted: r.World.EmittedCount(), History: r.History(),
		Reproducible: r.Reproducible(), Unblinded: r.UnblindedStamp(),
	}
}
