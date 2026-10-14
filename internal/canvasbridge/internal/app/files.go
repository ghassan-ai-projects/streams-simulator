package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"

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
	return b.summary(runID).State(b.runIsRunning(runID))
}

func writablePath(rel string) (domain.WorkspacePath, error) {
	path, err := domain.ParsePath(rel)
	if err != nil {
		return domain.WorkspacePath{}, err
	}
	return path, path.Writable()
}
