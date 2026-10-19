package app

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run/internal/durable"
)

var (
	ErrFieldTrace          = rules.ErrFieldTrace
	ErrFieldTraceMissing   = errors.New("run: no field trace in the run folder")
	ErrFieldTraceDigest    = errors.New("run: the field trace differs from the digest run.json recorded")
	ErrFieldTraceUncovered = errors.New("run: run.json records no field_trace_digest for the field trace in its folder")
	ErrFieldTraceFull      = errors.New("run: the field trace reached its size limit")

	ErrInterruptLogMissing   = errors.New("run: no interrupt log in the run folder")
	ErrInterruptLogDigest    = errors.New("run: the interrupt log differs from the digest run.json recorded")
	ErrInterruptLogUncovered = errors.New("run: run.json records no interrupt_log_digest for the interrupt log in its folder")
)

type FieldTrace struct {
	dir    string
	name   string
	noun   string
	file   *durable.FieldTraceFile
	next   uint64
	size   int
	digest string
}

func OpenFieldTrace(dir string) (*FieldTrace, error) {
	return openLog(dir, model.FileFieldTrace, durable.NounFieldTrace)
}

func OpenInterruptLog(dir string) (*FieldTrace, error) {
	return openLog(dir, model.FileInterruptLog, durable.NounInterruptLog)
}

func openLog(dir, name, noun string) (*FieldTrace, error) {
	file, err := durable.OpenEvidenceLog(dir, name, noun)
	if err != nil {
		return nil, err
	}
	return &FieldTrace{dir: dir, name: name, noun: noun, file: file}, nil
}

func (t *FieldTrace) Add(record map[string]any) error {
	line, err := rules.EncodeFieldTraceLine(t.next, record)
	if err != nil {
		return err
	}
	if t.size+len(line) > rules.MaxFieldTraceBytes {
		return fmt.Errorf("%w: %d bytes", ErrFieldTraceFull, rules.MaxFieldTraceBytes)
	}
	if err := t.file.Append(line); err != nil {
		return err
	}
	t.next, t.size = t.next+1, t.size+len(line)
	return nil
}

func (t *FieldTrace) Flush() error {
	return t.file.Flush()
}

func (t *FieldTrace) Close() error {
	if err := t.file.Close(); err != nil {
		return err
	}
	raw, _, err := durable.ReadEvidenceLog(t.dir, t.name, t.noun, rules.MaxFieldTraceBytes)
	if err != nil {
		return err
	}
	t.digest = rules.FieldTraceDigest(raw)
	return nil
}

func (t *FieldTrace) Digest() string {
	return t.digest
}

func ReadFieldTrace(dir string) ([]map[string]any, string, error) {
	raw, present, err := durable.ReadEvidenceLog(dir, model.FileFieldTrace, durable.NounFieldTrace, rules.MaxFieldTraceBytes)
	if err != nil {
		return nil, "", err
	}
	if !present {
		return nil, "", fmt.Errorf("%w: %s", ErrFieldTraceMissing, dir)
	}
	digest := rules.FieldTraceDigest(raw)
	if err := checkRecordedFieldTrace(dir, digest); err != nil {
		return nil, "", err
	}
	records, err := rules.ParseFieldTrace(raw)
	return records, digest, err
}

func checkRecordedFieldTrace(dir, digest string) error {
	return checkRecordedDigest(dir, digest, func(art *model.RunArtifact) string { return art.FieldTraceDigest },
		ErrFieldTraceUncovered, ErrFieldTraceDigest)
}

func ReadInterruptLog(dir string) ([]map[string]any, string, error) {
	raw, err := interruptLogBytes(dir)
	if err != nil {
		return nil, "", err
	}
	digest := rules.FieldTraceDigest(raw)
	if err := checkRecordedInterruptLog(dir, digest); err != nil {
		return nil, "", err
	}
	records, err := rules.ParseFieldTrace(raw)
	if err != nil {
		return nil, "", fmt.Errorf("run: interrupt log: %w", err)
	}
	return records, digest, nil
}

func interruptLogBytes(dir string) ([]byte, error) {
	raw, present, err := durable.ReadEvidenceLog(dir, model.FileInterruptLog, durable.NounInterruptLog, rules.MaxFieldTraceBytes)
	if err == nil && !present {
		return nil, fmt.Errorf("%w: %s", ErrInterruptLogMissing, dir)
	}
	return raw, err
}

func checkRecordedInterruptLog(dir, digest string) error {
	return checkRecordedDigest(dir, digest, func(art *model.RunArtifact) string { return art.InterruptLogDigest },
		ErrInterruptLogUncovered, ErrInterruptLogDigest)
}

func checkRecordedDigest(dir, digest string, recorded func(*model.RunArtifact) string, uncovered, differs error) error {
	art, err := LoadArtifactForVerification(filepath.Join(dir, model.FileRunArtifact))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case recorded(art) == "":
		return fmt.Errorf("%w: %s", uncovered, dir)
	case recorded(art) != digest:
		return fmt.Errorf("%w: recorded %s, the file digests to %s", differs, recorded(art), digest)
	}
	return nil
}

func RejectInterruptLog(dir string) error {
	return durable.RenameEvidenceLog(dir, model.FileInterruptLog, model.FileInterruptLogRejected)
}

func (r *Run) SetFieldTrace(digest string) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.fieldTraceDigest = digest
}

func (r *Run) SetInterruptLog(digest string) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.interruptLogDigest = digest
}
