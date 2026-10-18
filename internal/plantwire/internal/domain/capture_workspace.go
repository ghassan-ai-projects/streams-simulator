package domain

type WorkspaceKind struct {
	Name   string
	Path   string
	Entry  map[string]any
	Digest string
	Err    error
}

func LoadWorkspaceKinds(sources []KindSource) []WorkspaceKind {
	kinds := make([]WorkspaceKind, len(sources))
	for i, source := range sources {
		kinds[i] = loadWorkspaceKind(source)
	}
	return kinds
}

func loadWorkspaceKind(source KindSource) WorkspaceKind {
	name := kindNameOfSource(source)
	files, summaries, err := ParseKindFiles([]KindSource{source})
	if err != nil {
		return WorkspaceKind{Name: name, Path: source.Path, Err: err}
	}
	return WorkspaceKind{Name: name, Path: source.Path, Entry: files[summaries[0].Kind], Digest: summaries[0].KindDigest}
}
