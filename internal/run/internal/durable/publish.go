//nolint:staticcheck // ST1005: "End" names the public Run.End operation these messages report on; the text is part of the error contract
package durable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Evidence struct {
	TracePath string
	Trace     []byte

	Ledger      []model.LedgerRecord
	WriteLedger bool
	History     []model.StateSnapshot
	Verdict     *model.Verdict
}

func PublishEvidence(outDir string, e Evidence) error {
	if err := kernel.PrivateDir(outDir); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if err := publishTrace(outDir, e); err != nil {
		return err
	}
	if err := publishLedger(outDir, e); err != nil {
		return err
	}
	if err := writeJSONL(filepath.Join(outDir, model.FileHistory), e.History); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return publishVerdict(outDir, e.Verdict)
}

func WriteRunArtifact(outDir string, art *model.RunArtifact) error {
	if outDir == "" {
		return nil
	}
	raw, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		return fmt.Errorf("End: %w", err)
	}
	if err := kernel.WriteAtomic(filepath.Join(outDir, model.FileRunArtifact), raw); err != nil {
		return fmt.Errorf("End: %w", err)
	}
	return nil
}

func ReadArtifact(path string) ([]byte, error) {
	raw, err := kernel.Read(path)
	if err != nil {
		return nil, fmt.Errorf("run: read artifact %s: %w", path, err)
	}
	return raw, nil
}

func publishTrace(outDir string, e Evidence) error {
	paths := []string{filepath.Join(outDir, model.FileTrace)}
	if e.TracePath != "" && e.TracePath != paths[0] {
		paths = append(paths, e.TracePath)
	}
	for _, path := range paths {
		if err := kernel.WriteAtomic(path, e.Trace); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func publishLedger(outDir string, e Evidence) error {
	if e.WriteLedger {
		if err := writeJSONL(filepath.Join(outDir, model.FileLedger), e.Ledger); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func publishVerdict(outDir string, verdict *model.Verdict) error {
	if verdict != nil {
		raw, _ := json.MarshalIndent(verdict, "", "  ")
		if err := kernel.WriteAtomic(filepath.Join(outDir, model.FileVerdict), raw); err != nil {
			return fmt.Errorf("End: %w", err)
		}
	}
	return nil
}

func JSONLines[T any](records []T) []byte {
	var lines bytes.Buffer
	for _, record := range records {
		raw, _ := json.Marshal(record)
		lines.Write(append(raw, '\n'))
	}
	return lines.Bytes()
}

func writeJSONL[T any](path string, records []T) error {
	if err := kernel.WriteAtomic(path, JSONLines(records)); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func ReadJournal(dir, name string) (raw []byte, present bool, err error) {
	raw, present, err = kernel.ReadIfPresent(filepath.Join(dir, name))
	if err != nil {
		return nil, false, fmt.Errorf("run: read journal %s: %w", name, err)
	}
	return raw, present, nil
}
