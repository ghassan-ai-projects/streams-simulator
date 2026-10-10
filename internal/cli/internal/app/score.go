package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// scoreOptions are the inputs of the offline score command.
type scoreOptions struct{ artifact, label, domainsDir, adaptersDir string }

func cmdScore(s *session, args []string) (any, error) {
	options, err := parseScoreOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	return scoreArtifact(options)
}

func loadJSON(path string, dst any) error {
	raw, err := files.Read(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

// loadLedger reads the newline-delimited delivery ledger written by Run.End.
// A ledger is JSONL, not one JSON array; decoding it as a single JSON value
// would silently discard delivery evidence from the offline score path.
func loadLedger(path string) ([]model.LedgerRecord, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("open ledger %s: %w", path, err)
	}
	return decodeLedger(json.NewDecoder(bytes.NewReader(raw)), path)
}

func decodeLedger(dec *json.Decoder, path string) ([]model.LedgerRecord, error) {
	var ledger []model.LedgerRecord
	for {
		var record model.LedgerRecord
		err := dec.Decode(&record)
		if err == io.EOF {
			return ledger, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decode ledger %s: %w", path, err)
		}
		ledger = append(ledger, record)
	}
}

func parseScoreOptions(args []string, stderr io.Writer) (scoreOptions, error) {
	var options scoreOptions
	fs := newFlagSet("score", stderr)
	fs.StringVar(&options.artifact, "run", "", "run artifact (run.json)")
	fs.StringVar(&options.label, "label", "", "ground truth label (JSON)")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory (the artifact is replayed to recover the effector calls)")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	if err := parseFlags(fs, args); err != nil {
		return scoreOptions{}, err
	}
	if options.artifact == "" {
		return scoreOptions{}, fmt.Errorf("score requires --run")
	}
	return options, nil
}

func loadScoreReports(dir, label string) (*model.Verdict, *model.GroundTruthRecord, error) {
	if label == "" {
		label = filepath.Join(dir, "label.json")
	}
	var verdict model.Verdict
	if err := loadJSON(filepath.Join(dir, "verdict.json"), &verdict); err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	var gt model.GroundTruthRecord
	if err := loadJSON(label, &gt); err != nil {
		return nil, nil, fmt.Errorf("streamsim: %w", err)
	}
	return &verdict, &gt, nil
}

// scoreArtifact grades a finished run from its artifacts with the same scorer
// the director uses online, so the two scorecards agree. Delivery and state
// evidence come from the files the run published; the effector calls come
// from replaying the artifact, because the command log records the
// invocations but not whether each took effect. A replay that does not
// reproduce the recorded trace is refused rather than graded.
func scoreArtifact(options scoreOptions) (any, error) {
	dir := filepath.Dir(options.artifact)
	verdict, gt, err := loadScoreReports(dir, options.label)
	if err != nil {
		return nil, err
	}
	evidence, err := scoringEvidence(options, dir)
	if err != nil {
		return nil, err
	}
	evidence.Verdict = verdict
	return gradeRun(evidence, gt)
}

func gradeRun(evidence score.Evidence, gt *model.GroundTruthRecord) (any, error) {
	card, err := score.Score(evidence, gt)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return card, nil
}

// scoringEvidence assembles everything the scorer needs except the verdict:
// the replayed identity, domain and effector calls, and the published ledger
// and state history.
func scoringEvidence(options scoreOptions, dir string) (score.Evidence, error) {
	evidence, err := replayedEvidence(options)
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
func replayedEvidence(options scoreOptions) (score.Evidence, error) {
	art, err := run.LoadArtifact(options.artifact)
	if err != nil {
		return score.Evidence{}, fmt.Errorf("streamsim: %w", err)
	}
	spec, adap, err := loadSimulatorInputs(options.domainsDir, options.adaptersDir, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return score.Evidence{}, err
	}
	replayed, err := replayReproducing(art, spec, adap)
	if err != nil {
		return score.Evidence{}, err
	}
	return artifactEvidence(art, spec, replayed.Calls), nil
}

func artifactEvidence(art *model.RunArtifact, spec *domain.Compiled, calls []world.EffectorCall) score.Evidence {
	return score.Evidence{
		RunID: art.RunID, Domain: spec, Calls: calls, Perturbations: art.AppliedPerturbations,
		Emitted: art.Counts.Emitted, Reproducible: art.Reproducible, Unblinded: art.Unblinded,
	}
}

func replayReproducing(art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter) (*run.ReplayEvidence, error) {
	replayed, err := run.ReplayArtifactEvidence(context.Background(), art, spec, adap)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if !replayed.Result.Matches {
		return nil, fmt.Errorf("score: replay does not reproduce the artifact, so its effector calls cannot be graded: %s", replayed.Result.Detail)
	}
	return replayed, nil
}

// loadHistory reads the director-only per-emission state history a run
// publishes beside its artifact.
func loadHistory(path string) ([]model.StateSnapshot, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("open state history %s: %w", path, err)
	}
	history, err := decodeHistory(json.NewDecoder(bytes.NewReader(raw)))
	if err != nil {
		return nil, fmt.Errorf("decode state history %s: %w", path, err)
	}
	return history, nil
}

func decodeHistory(dec *json.Decoder) ([]model.StateSnapshot, error) {
	var history []model.StateSnapshot
	for {
		var snapshot model.StateSnapshot
		err := dec.Decode(&snapshot)
		if errors.Is(err, io.EOF) {
			return history, nil
		}
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		history = append(history, snapshot)
	}
}
