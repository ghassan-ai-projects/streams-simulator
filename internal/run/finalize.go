package run

import (
	"encoding/json"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
	"os"
	"path/filepath"
)

// End finalizes the run: closes the sink, computes the trace digest, and
// writes the run artifact, ledger, world-state history and (if any) verdict
// into outDir.
func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()

	if r.finished {
		return nil, fmt.Errorf("run: already finished")
	}
	if err := r.finishTrace(); err != nil {
		return nil, err
	}
	r.finished = true
	if outDir != "" {
		if err := r.publishEvidence(outDir); err != nil {
			return nil, err
		}
	}

	art := r.artifact()
	if outDir != "" {
		raw, err := json.MarshalIndent(art, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
		if err := os.WriteFile(filepath.Join(outDir, "run.json"), raw, 0o600); err != nil {
			return nil, fmt.Errorf("End: %w", err)
		}
	}
	if r.runErr != nil {
		return art, r.runErr
	}
	return art, nil
}

func (r *Run) finishTrace() error {
	if r.runErr == nil {
		if lines, err := r.Engine.End(r.traceEndTime()); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		} else if err := writeSinkLines(r.Sink, lines); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		}
	}
	trace, err := r.Sink.Close()
	if err != nil {
		return fmt.Errorf("End: %w", err)
	}
	r.trace = trace
	r.traceDigest = canonical.DigestBytes(trace)
	if err := r.closeDurableLedger(); err != nil {
		return err
	}
	return nil
}

func (r *Run) closeDurableLedger() error {
	if r.ledgerFile == nil {
		return nil
	}
	if err := r.ledgerWriter.Flush(); err != nil {
		_ = r.ledgerFile.Close()
		return fmt.Errorf("End: flush ledger: %w", err)
	}
	if err := r.ledgerFile.Sync(); err != nil {
		_ = r.ledgerFile.Close()
		return fmt.Errorf("End: sync ledger: %w", err)
	}
	if err := r.ledgerFile.Close(); err != nil {
		return fmt.Errorf("End: close ledger: %w", err)
	}
	return nil
}

func (r *Run) publishEvidence(outDir string) error {
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	tracePath := filepath.Join(outDir, "trace.jsonl")
	if r.Config.SinkName == model.SinkFile && r.Config.SinkTarget != "" {
		tracePath = r.Config.SinkTarget
	}
	if err := os.WriteFile(tracePath, r.trace, 0o600); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if r.ledgerFile == nil {
		if err := writeJSONL(filepath.Join(outDir, "ledger.jsonl"), r.ledger); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	if err := writeJSONL(filepath.Join(outDir, "world_state_history.jsonl"), r.history); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if r.verdict != nil {
		raw, _ := json.MarshalIndent(r.verdict, "", "  ")
		if err := os.WriteFile(filepath.Join(outDir, "verdict.json"), raw, 0o600); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func writeSinkLines(dst sink.Sink, lines []string) error {
	for _, line := range lines {
		if err := dst.Write([]byte(line)); err != nil {
			return fmt.Errorf("write sink line: %w", err)
		}
	}
	return nil
}
