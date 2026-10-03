package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func cmdScore(args []string) error {
	artifact, label, err := parseScoreOptions(args)
	if err != nil {
		return err
	}
	return scoreArtifact(artifact, label)
}

func loadJSON(path string, dst any) error {
	raw, err := os.ReadFile(path)
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
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open ledger %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return decodeLedger(json.NewDecoder(f), path)
}

func printJSON(v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Println(string(raw))
	return nil
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

func parseScoreOptions(args []string) (string, string, error) {
	fs := flag.NewFlagSet("score", flag.ExitOnError)
	artifact := fs.String("run", "", "run artifact (run.json)")
	label := fs.String("label", "", "ground truth label (JSON)")
	if err := fs.Parse(args); err != nil {
		return "", "", fmt.Errorf("streamsim: %w", err)
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

func scoreArtifact(artifact, label string) error {
	dir := filepath.Dir(artifact)
	verdict, gt, err := loadScoreReports(dir, label)
	if err != nil {
		return err
	}
	ledger, err := loadLedger(filepath.Join(dir, "ledger.jsonl"))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(score.Offline(verdict, gt, ledger, nil, artifactPerturbations(artifact)))
}
