//nolint:staticcheck // ST1005: "End" names the public Run.End operation these messages report on; the text is part of the error contract
package durable

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Evidence is what a finished run publishes beside its artifact.
type Evidence struct {
	// TracePath is where the delivered trace is written; empty selects
	// trace.jsonl under the output directory.
	TracePath string
	Trace     []byte
	// Ledger is written as ledger.jsonl unless the ledger was already kept
	// durably (WriteLedger false).
	Ledger      []model.LedgerRecord
	WriteLedger bool
	History     []model.StateSnapshot
	Verdict     *model.Verdict
}

// PublishEvidence writes the run's evidence files under outDir: trace,
// ledger, world-state history and, when one was submitted, the verdict.
func PublishEvidence(outDir string, e Evidence) error {
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if err := publishTrace(outDir, e); err != nil {
		return err
	}
	if err := publishLedger(outDir, e); err != nil {
		return err
	}
	if err := writeJSONL(filepath.Join(outDir, "world_state_history.jsonl"), e.History); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return publishVerdict(outDir, e.Verdict)
}

// WriteRunArtifact writes run.json under outDir ("" writes nothing).
func WriteRunArtifact(outDir string, art *model.RunArtifact) error {
	if outDir == "" {
		return nil
	}
	raw, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "run.json"), raw, 0o600); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return nil
}

// ReadArtifact reads the raw bytes of a run artifact file.
func ReadArtifact(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("run: read artifact %s: %w", path, err)
	}
	return raw, nil
}

func publishTrace(outDir string, e Evidence) error {
	tracePath := e.TracePath
	if tracePath == "" {
		tracePath = filepath.Join(outDir, "trace.jsonl")
	}
	if err := os.WriteFile(tracePath, e.Trace, 0o600); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return nil
}

func publishLedger(outDir string, e Evidence) error {
	if e.WriteLedger {
		if err := writeJSONL(filepath.Join(outDir, "ledger.jsonl"), e.Ledger); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func publishVerdict(outDir string, verdict *model.Verdict) error {
	if verdict != nil {
		raw, _ := json.MarshalIndent(verdict, "", "  ")
		if err := os.WriteFile(filepath.Join(outDir, "verdict.json"), raw, 0o600); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func writeJSONL[T any](path string, records []T) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = file.Close() }()
	for _, record := range records {
		raw, _ := json.Marshal(record)
		if _, err := file.Write(append(raw, '\n')); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}
