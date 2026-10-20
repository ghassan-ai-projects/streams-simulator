package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

var (
	ErrTwinNoJournal     = errors.New("plantwire: the run folder holds no journal.jsonl, where the commands the host sent are recorded")
	ErrTwinJournalDigest = errors.New("plantwire: the journal differs from the digest run.json recorded")

	ErrTwinInterruptLogRejected = errors.New("plantwire: the run folder holds an interrupt log the session set aside as unreplayable, so the commands the node refused or stopped are unknown")
	ErrTwinNoInterruptLog       = errors.New("plantwire: run.json records an interrupt log digest but the run folder holds no interrupts.jsonl")
)

type TwinReportOptions struct {
	Run, DomainsDir, Domain string
	KindsDirs, Tolerances   []string
	SpanFraction            float64
}

type twinFolder struct {
	trace      []map[string]any
	journal    []byte
	interrupts []map[string]any
	evidence   domain.TwinEvidence
}

func TwinReport(ctx context.Context, options TwinReportOptions) (domain.TwinReport, error) {
	declared, parsed := domain.ParseTolerances(options.Tolerances)
	if err := errors.Join(ctx.Err(), options.valid(), parsed); err != nil {
		return domain.TwinReport{}, fmt.Errorf("plantwire twin-report: %w", err)
	}
	report, err := buildTwinReport(options, declared)
	if err != nil {
		return domain.TwinReport{}, fmt.Errorf("plantwire twin-report: %w", err)
	}
	return report, validateTwinReport(report)
}

func (o TwinReportOptions) valid() error {
	if o.Run == "" || o.DomainsDir == "" || o.Domain == "" || len(o.KindsDirs) == 0 {
		return fmt.Errorf("%w: a twin report needs a run folder, a domains directory, a domain and a kinds directory", ErrBadOptions)
	}
	return nil
}

func buildTwinReport(options TwinReportOptions, declared map[string]float64) (domain.TwinReport, error) {
	folder, err := readTwinFolder(options.Run)
	if err != nil {
		return domain.TwinReport{}, err
	}
	input, twinModel, err := twinInput(options, declared, folder)
	if err != nil {
		return domain.TwinReport{}, err
	}
	twin, err := newTwinWorld(twinModel, input.Trace.StartNS, input.Trace.Devices)
	if err != nil {
		return domain.TwinReport{}, err
	}
	return domain.EvaluateTwin(input, twin)
}

type twinModel struct {
	spec  *simdomain.Compiled
	faces domain.Interfaces
	kinds domain.KindFiles
}

func twinInput(options TwinReportOptions, declared map[string]float64, folder twinFolder) (domain.TwinInput, twinModel, error) {
	board := domain.ReadBoardTrace(folder.trace)
	if len(board.Devices) == 0 {
		return domain.TwinInput{}, twinModel{}, domain.ErrTwinNoDevices
	}
	twin, info, err := twinDomain(options, board)
	if err != nil {
		return domain.TwinInput{}, twinModel{}, err
	}
	commands, err := twinCommands(folder, board)
	return domain.TwinInput{Trace: board, Commands: commands, Kinds: twin.kinds, Declared: declared, SpanFraction: options.SpanFraction,
		Run: folder.evidence, Twin: info}, twin, err
}

func twinDomain(options TwinReportOptions, board domain.BoardTrace) (twinModel, domain.TwinInfo, error) {
	spec, err := loadTwinDomain(options)
	if err != nil {
		return twinModel{}, domain.TwinInfo{}, err
	}
	twin, checked, err := twinKinds(spec, board, options.KindsDirs)
	info := domain.TwinInfo{Domain: spec.Spec.ID, DomainDigest: spec.Digest, Kinds: twinKindDigests(checked), StartNS: board.StartNS}
	return twin, info, err
}

func loadTwinDomain(options TwinReportOptions) (*simdomain.Compiled, error) {
	catalog, err := simdomain.NewCatalogFromDir(options.DomainsDir)
	if err != nil {
		return nil, fmt.Errorf("read the domains directory: %w", err)
	}
	spec, err := catalog.Describe(options.Domain)
	if err != nil {
		return nil, fmt.Errorf("describe the twin domain: %w", err)
	}
	return spec, nil
}

func twinKinds(spec *simdomain.Compiled, board domain.BoardTrace, dirs []string) (twinModel, []model.ImplementsChecked, error) {
	kinds, err := readKindFiles(nil, dirs)
	if err != nil {
		return twinModel{}, nil, err
	}
	if err := domain.TwinImplements(spec.Spec, board.Devices); err != nil {
		return twinModel{}, nil, err
	}
	boards, err := readBoards(spec.Spec, dirs)
	if err != nil {
		return twinModel{}, nil, err
	}
	checked, faces, err := domain.CheckInterfaces(spec.Spec, kinds, boards)
	return twinModel{spec: spec, faces: faces, kinds: kinds}, checked, err
}

func twinKindDigests(checked []model.ImplementsChecked) []domain.TwinKind {
	digests := make([]domain.TwinKind, len(checked))
	for i, entry := range checked {
		digests[i] = domain.TwinKind{Kind: entry.Kind, KindDigest: entry.KindDigest}
	}
	return digests
}

func twinCommands(folder twinFolder, board domain.BoardTrace) ([]domain.TwinCommand, error) {
	sent, err := domain.ReadHostCommands(folder.journal)
	if err != nil {
		return nil, err
	}
	sent = append(sent, domain.NodeStops(folder.interrupts)...)
	return domain.ClassifyCommands(sent, board, domain.InterruptBlocks(folder.interrupts)), nil
}

func validateTwinReport(report domain.TwinReport) error {
	raw, err := json.Marshal(report)
	if err == nil {
		err = model.ValidateTwinReport(raw)
	}
	if err != nil {
		return fmt.Errorf("plantwire twin-report: the report breaks its contract: %w", err)
	}
	return nil
}

func readTwinFolder(dir string) (twinFolder, error) {
	art, err := twinArtifact(dir)
	if err != nil {
		return twinFolder{}, err
	}
	folder, err := readTwinFiles(dir, art)
	folder.evidence.EvidenceBound = art != nil && art.JournalFileDigest != ""
	return folder, err
}

func readTwinFiles(dir string, art *model.RunArtifact) (twinFolder, error) {
	trace, traceDigest, err := run.ReadFieldTrace(dir)
	if err != nil {
		return twinFolder{}, fmt.Errorf("read the field trace: %w", err)
	}
	journal, journalDigest, err := readTwinJournal(dir, art)
	if err != nil {
		return twinFolder{}, err
	}
	interrupts, interruptDigest, err := readTwinInterrupts(dir, art)
	evidence := domain.TwinEvidence{FieldTraceDigest: traceDigest, JournalDigest: journalDigest, InterruptLogDigest: interruptDigest}
	return twinFolder{trace: trace, journal: journal, interrupts: interrupts, evidence: evidence}, err
}

func twinArtifact(dir string) (*model.RunArtifact, error) {
	art, err := run.LoadArtifactForVerification(filepath.Join(dir, model.FileRunArtifact))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("read the run artifact: %w", err)
	}
	return art, nil
}

func readTwinJournal(dir string, art *model.RunArtifact) ([]byte, string, error) {
	raw, present, err := files.ReadRunFile(dir, model.FileObserverJournal)
	if err != nil {
		return nil, "", err
	}
	if !present {
		return nil, "", fmt.Errorf("%w: %s", ErrTwinNoJournal, dir)
	}
	digest := canonical.DigestBytes(raw)
	if art != nil && art.JournalFileDigest != "" && art.JournalFileDigest != digest {
		return nil, "", fmt.Errorf("%w: recorded %s, the file digests to %s", ErrTwinJournalDigest, art.JournalFileDigest, digest)
	}
	return raw, digest, nil
}

func refuseRejectedLog(dir string) error {
	present, err := files.RunFilePresent(dir, model.FileInterruptLogRejected)
	if err == nil && present {
		return fmt.Errorf("%w: %s", ErrTwinInterruptLogRejected, model.FileInterruptLogRejected)
	}
	return err
}

func readTwinInterrupts(dir string, art *model.RunArtifact) ([]map[string]any, string, error) {
	if err := refuseRejectedLog(dir); err != nil {
		return nil, "", err
	}
	records, digest, err := verifiedInterruptLog(dir)
	if !errors.Is(err, run.ErrInterruptLogMissing) {
		return records, digest, err
	}
	if art != nil && art.InterruptLogDigest != "" {
		return nil, "", fmt.Errorf("%w: run.json records %s", ErrTwinNoInterruptLog, art.InterruptLogDigest)
	}
	return nil, "", nil
}
