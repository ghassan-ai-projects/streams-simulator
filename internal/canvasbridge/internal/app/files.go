package app

import (
	"bytes"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (b *Bridge) ReadFile(rel string) (domain.FileDocument, error) {
	path, err := domain.ParsePath(rel)
	if err != nil {
		return domain.FileDocument{}, err
	}
	if err := domain.ReadableRunFile(path, b.runState(path.RunID)); err != nil {
		return domain.FileDocument{}, err
	}
	content, err := b.edges.Workspace.Read(path.Rel, domain.MaxFileBytes)
	if err != nil {
		return domain.FileDocument{}, err
	}
	return domain.FileDocument{Path: path.Rel, Content: string(content), Etag: domain.Version(content)}, nil
}

func (b *Bridge) WriteFile(rel, content, etag string) (domain.FileDocument, error) {
	path, err := writablePath(rel)
	if err != nil {
		return domain.FileDocument{}, err
	}
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	version, err := b.edges.Workspace.WriteIfUnchanged(path.Rel, []byte(content), etag, domain.Version, domain.MaxFileBytes)
	if err != nil {
		return domain.FileDocument{Path: path.Rel, Etag: version}, err
	}
	return domain.FileDocument{Path: path.Rel, Content: content, Etag: version}, nil
}

func (b *Bridge) runState(runID string) domain.RunState {
	if runID == "" {
		return domain.RunState{}
	}
	return domain.RunState{Ended: b.closedRun(runID) && !b.runIsRunning(runID), Explore: b.exploreRun(runID)}
}

func (b *Bridge) closedRun(runID string) bool {
	tail, err := b.edges.Workspace.Tail("runs/"+runID+"/journal.jsonl", closingLimit)
	return err == nil && domain.ClosedJournal(tail)
}

func (b *Bridge) exploreRun(runID string) bool {
	head, err := b.edges.Workspace.ReadFrom("runs/"+runID+"/journal.jsonl", 0, firstLineLimit)
	if err != nil {
		return false
	}
	first, _, _ := bytes.Cut(head, []byte("\n"))
	return domain.ExploreRun(first)
}

func writablePath(rel string) (domain.WorkspacePath, error) {
	path, err := domain.ParsePath(rel)
	if err != nil {
		return domain.WorkspacePath{}, err
	}
	return path, path.Writable()
}
