package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
)

// End finalizes the run: closes the sink, computes the trace digest, and
// writes the run artifact, ledger, world-state history and (if any) verdict
// into outDir.
func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.endLocked(outDir)
}

func (r *Run) endLocked(outDir string) (*model.RunArtifact, error) {
	if r.finished {
		return nil, fmt.Errorf("run: already finished")
	}
	return r.finishRun(outDir)
}

func (r *Run) finishRun(outDir string) (*model.RunArtifact, error) {
	if err := r.finishTrace(); err != nil {
		return nil, err
	}
	r.finished = true
	if err := r.publishRequestedEvidence(outDir); err != nil {
		return nil, err
	}
	art := r.artifact()
	if err := writeRunArtifact(outDir, art); err != nil {
		return nil, err
	}
	return art, r.runErr
}

func (r *Run) publishRequestedEvidence(outDir string) error {
	if outDir == "" {
		return nil
	}
	return r.publishEvidence(outDir)
}

func writeRunArtifact(outDir string, art *model.RunArtifact) error {
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

func (r *Run) finishTrace() error {
	r.finishPostamble()
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

func (r *Run) finishPostamble() {
	if r.runErr == nil {
		if lines, err := r.Engine.End(r.traceEndTime()); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		} else if err := writeSinkLines(r.Sink, lines); err != nil {
			r.fail(fmt.Errorf("End: adapter postamble: %w", err))
		}
	}
}

func (r *Run) closeDurableLedger() error {
	if r.ledgerFile == nil {
		return nil
	}
	if err := r.flushDurableLedger(); err != nil {
		_ = r.ledgerFile.Close()
		return err
	}
	if err := r.ledgerFile.Close(); err != nil {
		return fmt.Errorf("End: close ledger: %w", err)
	}
	return nil
}

func (r *Run) flushDurableLedger() error {
	if err := r.ledgerWriter.Flush(); err != nil {
		return fmt.Errorf("End: flush ledger: %w", err)
	}
	if err := r.ledgerFile.Sync(); err != nil {
		return fmt.Errorf("End: sync ledger: %w", err)
	}
	return nil
}

func (r *Run) publishEvidence(outDir string) error {
	if err := os.MkdirAll(outDir, 0o700); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if err := r.publishTrace(outDir); err != nil {
		return err
	}
	if err := r.publishLedger(outDir); err != nil {
		return err
	}
	if err := writeJSONL(filepath.Join(outDir, "world_state_history.jsonl"), r.history); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return r.publishVerdict(outDir)
}

func (r *Run) publishTrace(outDir string) error {
	tracePath := filepath.Join(outDir, "trace.jsonl")
	if r.Config.SinkName == model.SinkFile && r.Config.SinkTarget != "" {
		tracePath = r.Config.SinkTarget
	}
	if err := os.WriteFile(tracePath, r.trace, 0o600); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return nil
}

func (r *Run) publishLedger(outDir string) error {
	if r.ledgerFile == nil {
		if err := writeJSONL(filepath.Join(outDir, "ledger.jsonl"), r.ledger); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func (r *Run) publishVerdict(outDir string) error {
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
