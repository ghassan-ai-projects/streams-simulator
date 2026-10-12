package app

import (
	"context"
	"fmt"
	"path/filepath"

	catalog "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

// loadReports reads the consumer's verdict and the sealed label, each checked
// against its contract. An empty label selects label.json in dir.
func loadReports(dir, label string) (*model.Verdict, *model.GroundTruthRecord, error) {
	if label == "" {
		label = filepath.Join(dir, "label.json")
	}
	var verdict model.Verdict
	if err := loadDocument(filepath.Join(dir, "verdict.json"), model.ValidateVerdict, &verdict); err != nil {
		return nil, nil, err
	}
	var gt model.GroundTruthRecord
	if err := loadDocument(label, model.ValidateGroundTruth, &gt); err != nil {
		return nil, nil, err
	}
	return &verdict, &gt, nil
}

func loadDocument(path string, validate func([]byte) error, dst any) error {
	raw, err := files.Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return domain.DecodeDocument(raw, path, validate, dst)
}

// scoringEvidence assembles everything the scorer needs except the verdict:
// the replayed identity, domain and effector calls, and the published ledger
// and state history.
func scoringEvidence(ctx context.Context, options Options, art *model.RunArtifact, dir string) (score.Evidence, error) {
	evidence, err := replayedEvidence(ctx, options, art)
	if err != nil {
		return score.Evidence{}, err
	}
	if evidence.Ledger, err = loadLedger(filepath.Join(dir, "ledger.jsonl")); err != nil {
		return score.Evidence{}, fmt.Errorf("streamsim: %w", err)
	}
	if evidence.History, err = loadHistory(filepath.Join(dir, "world_state_history.jsonl")); err != nil {
		return score.Evidence{}, fmt.Errorf("streamsim: %w", err)
	}
	return evidence, nil
}

// replayedEvidence replays the artifact and returns the evidence it can
// state on its own: identity, domain, effector calls, perturbations, counts
// and the reproducible/unblinded stamps.
func replayedEvidence(ctx context.Context, options Options, art *model.RunArtifact) (score.Evidence, error) {
	spec, adap, err := replayInputs(options, art)
	if err != nil {
		return score.Evidence{}, fmt.Errorf("%w", err)
	}
	replayed, err := run.ReplayArtifactEvidence(ctx, art, spec, adap)
	if err != nil {
		return score.Evidence{}, fmt.Errorf("streamsim: %w", err)
	}
	if err := domain.RequireReproduced(replayed); err != nil {
		return score.Evidence{}, err
	}
	return domain.ArtifactEvidence(art, spec, replayed.Calls), nil
}

func loadLedger(path string) ([]model.LedgerRecord, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("open ledger %s: %w", path, err)
	}
	return domain.DecodeLedger(raw, path)
}

func loadHistory(path string) ([]model.StateSnapshot, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("open state history %s: %w", path, err)
	}
	return domain.DecodeHistory(raw, path)
}

func replayInputs(options Options, art *model.RunArtifact) (*catalog.Compiled, *model.Adapter, error) {
	var spec *catalog.Compiled
	var adap *model.Adapter
	var err error
	if options.DomainsDir == "" && options.AdaptersDir == "" {
		spec, adap, err = run.EmbeddedInputs(art)
	} else {
		spec, adap, err = run.LoadInputs(options.DomainsDir, options.AdaptersDir, art.Domain.ID, art.Adapter.ID)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%w", err)
	}
	return spec, adap, nil
}
