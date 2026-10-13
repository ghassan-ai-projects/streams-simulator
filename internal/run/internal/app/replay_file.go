package app

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type Directories struct{ Domains, Adapters string }

func ReplayFile(ctx context.Context, artifactPath string, dirs Directories) (*ReplayResult, error) {
	art, err := LoadArtifact(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	res, err := replayRecordedInputs(ctx, art, dirs, filepath.Dir(artifactPath))
	if err != nil {
		return nil, err
	}
	noteDomainChange(res, art, dirs.Domains)
	return res, nil
}

func replayRecordedInputs(ctx context.Context, art *model.RunArtifact, dirs Directories, runDir string) (*ReplayResult, error) {
	spec, adap, err := recordedInputs(art, dirs)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return replayWithJournalFiles(ctx, art, spec, adap, runDir)
}

func recordedInputs(art *model.RunArtifact, dirs Directories) (*domain.Compiled, *model.Adapter, error) {
	if len(art.DomainSpec) > 0 && len(art.AdapterSpec) > 0 {
		return nil, nil, nil
	}
	spec, adap, err := LoadInputs(dirs.Domains, dirs.Adapters, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return nil, nil, err
	}
	spec, adap = unrecordedOnly(art, spec, adap)
	return spec, adap, nil
}

func unrecordedOnly(art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter) (*domain.Compiled, *model.Adapter) {
	if len(art.DomainSpec) > 0 {
		spec = nil
	}
	if len(art.AdapterSpec) > 0 {
		adap = nil
	}
	return spec, adap
}

func noteDomainChange(res *ReplayResult, art *model.RunArtifact, domainsDir string) {
	current, err := currentDomain(domainsDir, art.Domain.ID)
	if err != nil || current.Digest == art.Domain.Digest {
		return
	}
	res.DomainChanged, res.CurrentDomainDigest = true, current.Digest
}

func currentDomain(domainsDir, id string) (*domain.Compiled, error) {
	catalog, err := domain.NewCatalogFromDir(domainsDir)
	if err != nil {
		return nil, fmt.Errorf("run: current domains: %w", err)
	}
	current, err := catalog.Describe(id)
	if err != nil {
		return nil, fmt.Errorf("run: current domain: %w", err)
	}
	return current, nil
}

func replayWithJournalFiles(ctx context.Context, art *model.RunArtifact, spec *domain.Compiled, adap *model.Adapter, dir string) (*ReplayResult, error) {
	res, replay, err := replayArtifactRun(ctx, art, spec, adap, "")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := errors.Join(checkJournalFiles(dir, replay, art, res), checkEvidenceFiles(dir, replay, art, res)); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return res, nil
}
