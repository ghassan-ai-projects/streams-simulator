// Package app is the use-case layer of the report module: it reads a run folder, grades it and assembles the report.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func Build(ctx context.Context, dir string) (*domain.Report, error) {
	folder, err := load(ctx, dir)
	if err != nil {
		return nil, err
	}
	scoring, err := scoreOf(ctx, dir, folder)
	if err != nil {
		return nil, err
	}
	return domain.Assemble(folder, scoring)
}

var ErrUnknownFormat = errors.New("report: unknown format")

const (
	JSONFormat = "json"
	Markdown   = "markdown"
)

func Render(report *domain.Report, format string) ([]byte, error) {
	switch format {
	case JSONFormat:
		return domain.JSON(report)
	case Markdown:
		return domain.Markdown(report), nil
	}
	return nil, fmt.Errorf("%w %q", ErrUnknownFormat, format)
}

func load(ctx context.Context, dir string) (domain.Folder, error) {
	folder, err := loadConsumerSide(ctx, dir)
	if err != nil {
		return domain.Folder{}, err
	}
	if folder, err = loadSealedSide(ctx, dir, domain.Reconcile(folder)); err != nil {
		return domain.Folder{}, err
	}
	folder.Verdict, err = loadVerdict(dir)
	return folder, err
}

func loadSealedSide(ctx context.Context, dir string, folder domain.Folder) (domain.Folder, error) {
	if !domain.TruthAllowed(folder) {
		return folder, nil
	}
	if err := readInto(ctx, dir, directorSlots, &folder); err != nil {
		return domain.Folder{}, err
	}
	return domain.ReconcileSealed(folder), nil
}

func loadConsumerSide(ctx context.Context, dir string) (domain.Folder, error) {
	folder, err := loadDocuments(ctx, dir, consumerSlots)
	if err != nil {
		return domain.Folder{}, err
	}
	if folder.Artifact, err = loadArtifact(dir); err != nil {
		return domain.Folder{}, err
	}
	if folder.Artifact == nil && !folder.Observer.Present {
		return domain.Folder{}, fmt.Errorf("report: %s holds neither run.json nor journal.jsonl: %w", dir, fs.ErrNotExist)
	}
	return folder, nil
}

type slot struct {
	name   string
	kind   domain.Kind
	target func(*domain.Folder) *domain.Document
}

var consumerSlots = []slot{
	{model.FileTrace, domain.DigestOnly, func(f *domain.Folder) *domain.Document { return &f.Trace }},
	{"journal.jsonl", domain.ObserverJournal, func(f *domain.Folder) *domain.Document { return &f.Observer }},
	{model.FileFieldTrace, domain.DigestOnly, func(f *domain.Folder) *domain.Document { return &f.FieldTrace }},
	{model.FileInterruptLog, domain.DigestOnly, func(f *domain.Folder) *domain.Document { return &f.InterruptLog }},
}

var directorSlots = []slot{
	{"sealed/journal.jsonl", domain.DirectorJournal, func(f *domain.Folder) *domain.Document { return &f.Director }},
	{model.FileLedger, domain.LedgerRows, func(f *domain.Folder) *domain.Document { return &f.Ledger }},
	{model.FileHistory, domain.DigestOnly, func(f *domain.Folder) *domain.Document { return &f.History }},
}

func loadDocuments(ctx context.Context, dir string, slots []slot) (domain.Folder, error) {
	var folder domain.Folder
	err := readInto(ctx, dir, slots, &folder)
	return folder, err
}

func readInto(ctx context.Context, dir string, slots []slot, folder *domain.Folder) error {
	for _, s := range slots {
		doc, err := files.Scan(ctx, filepath.Join(dir, s.name), s.kind)
		if err != nil {
			return err
		}
		*s.target(folder) = doc
	}
	return nil
}

func loadArtifact(dir string) (*model.RunArtifact, error) {
	artifact, err := run.LoadArtifactForVerification(filepath.Join(dir, model.FileRunArtifact))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("report: %w", err)
	}
	return artifact, nil
}

func loadVerdict(dir string) (*model.Verdict, error) {
	path := filepath.Join(dir, model.FileVerdict)
	doc, err := files.Read(path)
	if err != nil || !doc.Present {
		return nil, err
	}
	var verdict model.Verdict
	if err := model.ValidateVerdict(doc.Raw); err != nil {
		return nil, fmt.Errorf("report: %w", model.WithSource(err, path))
	}
	if err := model.DecodeBytes(doc.Raw, &verdict); err != nil {
		return nil, fmt.Errorf("report: %s: %w", path, err)
	}
	return &verdict, nil
}
