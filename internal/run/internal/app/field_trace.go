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
)

type FieldTrace struct {
	dir    string
	file   *durable.FieldTraceFile
	next   uint64
	size   int
	digest string
}

func OpenFieldTrace(dir string) (*FieldTrace, error) {
	file, err := durable.OpenFieldTrace(dir)
	if err != nil {
		return nil, err
	}
	return &FieldTrace{dir: dir, file: file}, nil
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
	raw, _, err := durable.ReadFieldTrace(t.dir, rules.MaxFieldTraceBytes)
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
	raw, present, err := durable.ReadFieldTrace(dir, rules.MaxFieldTraceBytes)
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
	art, err := LoadArtifactForVerification(filepath.Join(dir, model.FileRunArtifact))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case art.FieldTraceDigest == "":
		return fmt.Errorf("%w: %s", ErrFieldTraceUncovered, dir)
	case art.FieldTraceDigest != digest:
		return fmt.Errorf("%w: recorded %s, the file digests to %s", ErrFieldTraceDigest, art.FieldTraceDigest, digest)
	}
	return nil
}

func (r *Run) SetFieldTrace(digest string) {
	r.commandMu.Lock()
	defer r.commandMu.Unlock()
	r.fieldTraceDigest = digest
}
