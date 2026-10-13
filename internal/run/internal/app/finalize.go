package app

import (
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
	"github.com/ghassan-ai-projects/streams-simulator/internal/sink"
)

func (r *Run) End(outDir string) (*model.RunArtifact, error) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.endLocked(outDir)
}

func (r *Run) Finished() bool {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	return r.finished
}

func (r *Run) endLocked(outDir string) (*model.RunArtifact, error) {
	if r.finished {
		return nil, fmt.Errorf("run: already finished")
	}
	return r.finishRun(outDir)
}

func (r *Run) finishRun(outDir string) (*model.RunArtifact, error) {

	traceErr := r.finishTrace()
	r.closeJournal()
	r.finished = true
	if traceErr != nil {
		return nil, traceErr
	}
	return r.publishArtifact(outDir)
}

func (r *Run) publishArtifact(outDir string) (*model.RunArtifact, error) {
	art := r.artifact()
	if err := r.publishRequestedEvidence(outDir); err != nil {
		return art, err
	}
	if err := durable.WriteRunArtifact(outDir, art); err != nil {
		return art, err
	}
	return art, r.runErr
}

func (r *Run) publishRequestedEvidence(outDir string) error {
	if outDir == "" {
		return nil
	}
	return durable.PublishEvidence(outDir, durable.Evidence{
		TracePath:   r.tracePath(outDir),
		Trace:       r.trace,
		Ledger:      r.ledger,
		WriteLedger: r.durableLedger == nil,
		History:     r.history,
		Verdict:     r.verdict,
	})
}

func (r *Run) tracePath(outDir string) string {
	if r.Config.SinkName == model.SinkFile && r.Config.SinkTarget != "" {
		return r.Config.SinkTarget
	}
	return filepath.Join(outDir, model.FileTrace)
}

func (r *Run) finishTrace() error {
	r.finishPostamble()
	trace, err := r.Sink.Close()

	ledgerErr := r.closeDurableLedger()
	if err != nil {
		return fmt.Errorf("End: %w", err)
	}
	r.trace = trace
	r.traceDigest = canonical.DigestBytes(trace)
	return ledgerErr
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
	if r.durableLedger == nil {
		return nil
	}
	return r.durableLedger.Finish()
}

func writeSinkLines(dst sink.Sink, lines []string) error {
	for _, line := range lines {
		if err := dst.Write([]byte(line)); err != nil {
			return fmt.Errorf("write sink line: %w", err)
		}
	}
	return nil
}
