package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	catalog "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/grading/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func loadReports(dir, label string) (*model.Verdict, []*model.GroundTruthRecord, error) {
	var verdict model.Verdict
	if err := loadDocument(filepath.Join(dir, model.FileVerdict), model.ValidateVerdict, &verdict); err != nil {
		return nil, nil, err
	}
	labels, err := loadLabels(dir, label)
	if err != nil {
		return nil, nil, err
	}
	return &verdict, labels, nil
}

func loadLabels(dir, label string) ([]*model.GroundTruthRecord, error) {
	raw, path, err := readLabels(dir, label)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return domain.DecodeLabels(raw, path)
}

func readLabels(dir, label string) ([]byte, string, error) {
	if label == "" {
		return readBesideRun(dir)
	}
	raw, err := files.Read(label)
	return raw, label, err
}

func readBesideRun(dir string) ([]byte, string, error) {
	single := filepath.Join(dir, model.FileLabel)
	raw, err := files.Read(single)
	if !errors.Is(err, fs.ErrNotExist) {
		return raw, single, err
	}
	several := filepath.Join(dir, model.FileLabels)
	if raw, again := files.Read(several); again == nil {
		return raw, several, nil
	}
	return nil, single, err
}

func loadDocument(path string, validate func([]byte) error, dst any) error {
	raw, err := files.Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return domain.DecodeDocument(raw, path, validate, dst)
}

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
	return withReplayedRecords(domain.ArtifactEvidence(art, spec, replayed.Calls), replayed), nil
}

func withReplayedRecords(evidence score.Evidence, replayed *run.ReplayEvidence) score.Evidence {
	evidence.Ledger, evidence.History = replayed.Ledger, replayed.History
	return evidence
}

func replayInputs(options Options, art *model.RunArtifact) (*catalog.Compiled, *model.Adapter, error) {
	spec, adap, err := run.ResolveInputs(art, run.Directories{Domains: options.DomainsDir, Adapters: options.AdaptersDir})
	if err != nil {
		return nil, nil, fmt.Errorf("%w", err)
	}
	return spec, adap, nil
}
