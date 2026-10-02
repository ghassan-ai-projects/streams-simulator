package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
	"io"
	"os"
	"path/filepath"
)

func cmdScore(args []string) error {
	fs := flag.NewFlagSet("score", flag.ExitOnError)
	runArtifact := fs.String("run", "", "run artifact (run.json)")
	labelPath := fs.String("label", "", "ground truth label (JSON)")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *runArtifact == "" {
		return fmt.Errorf("score requires --run")
	}
	dir := filepath.Dir(*runArtifact)
	verdictPath := filepath.Join(dir, "verdict.json")
	label := *labelPath
	if label == "" {
		label = filepath.Join(dir, "label.json")
	}
	var verdict model.Verdict
	if err := loadJSON(verdictPath, &verdict); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	var gt model.GroundTruthRecord
	if err := loadJSON(label, &gt); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	ledger, err := loadLedger(filepath.Join(dir, "ledger.jsonl"))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	var calls []world.EffectorCall
	var perturbations []string
	var art model.RunArtifact
	if err := loadJSON(*runArtifact, &art); err == nil {
		perturbations = art.AppliedPerturbations
	}
	sc := score.Offline(&verdict, &gt, ledger, calls, perturbations)
	return printJSON(sc)
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

	dec := json.NewDecoder(f)
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

func printJSON(v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	fmt.Println(string(raw))
	return nil
}
