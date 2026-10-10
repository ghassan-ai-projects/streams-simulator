package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func cmdScore(s *session, args []string) (any, error) {
	artifact, label, err := parseScoreOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	return scoreArtifact(artifact, label)
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

func parseScoreOptions(args []string, stderr io.Writer) (string, string, error) {
	fs := newFlagSet("score", stderr)
	artifact := fs.String("run", "", "run artifact (run.json)")
	label := fs.String("label", "", "ground truth label (JSON)")
	if err := parseFlags(fs, args); err != nil {
		return "", "", err
	}
	if *artifact == "" {
		return "", "", fmt.Errorf("score requires --run")
	}
	return *artifact, *label, nil
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

func artifactPerturbations(path string) []string {
	var art model.RunArtifact
	if err := loadJSON(path, &art); err == nil {
		return art.AppliedPerturbations
	}
	return nil
}

func scoreArtifact(artifact, label string) (any, error) {
	dir := filepath.Dir(artifact)
	verdict, gt, err := loadScoreReports(dir, label)
	if err != nil {
		return nil, err
	}
	ledger, err := loadLedger(filepath.Join(dir, "ledger.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return score.Offline(verdict, gt, ledger, nil, artifactPerturbations(artifact)), nil
}
