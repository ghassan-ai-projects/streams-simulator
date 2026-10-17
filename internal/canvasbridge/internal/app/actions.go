package app

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (b *Bridge) RunAction(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
	result, err := b.dispatch(ctx, name, args)
	if err != nil {
		return nil, redact(err, b.settings.WorkspaceDir)
	}
	return json.RawMessage(domain.Scrub(string(result), b.settings.WorkspaceDir)), nil
}

func (b *Bridge) dispatch(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
	action, err := domain.Lookup(name)
	if err != nil {
		return nil, err
	}
	if action.Kind == domain.KindCLI {
		return b.runCommand(ctx, action, args)
	}
	return b.callTool(ctx, action, args)
}

func (b *Bridge) runCommand(ctx context.Context, action domain.Action, args map[string]any) (json.RawMessage, error) {
	parts, err := action.Argv(args)
	if err != nil {
		return nil, err
	}
	guard, guarded, err := action.Guarded(args)
	if err != nil {
		return nil, err
	}
	if guarded {
		return b.executeGuarded(ctx, action, parts, guard)
	}
	return b.execute(ctx, action, parts)
}

func (b *Bridge) executeGuarded(ctx context.Context, action domain.Action, parts []domain.ArgvPart, guard domain.Guard) (json.RawMessage, error) {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	if err := b.checkVersion(guard); err != nil {
		return nil, err
	}
	return b.execute(ctx, action, parts)
}

func (b *Bridge) checkVersion(guard domain.Guard) error {
	current, err := b.edges.Workspace.Version(guard.Path, domain.Version, domain.MaxFileBytes)
	if err != nil {
		return err
	}
	if current != guard.Etag {
		return &domain.ConflictError{Path: guard.Path, Etag: current}
	}
	return nil
}

func (b *Bridge) execute(ctx context.Context, action domain.Action, parts []domain.ArgvPart) (json.RawMessage, error) {
	result, err := b.edges.RunCommand(ctx, append(b.absolute(parts), b.flags(action)...))
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, domain.CommandFailure(result.Stdout, result.ExitCode, string(result.Stderr))
	}
	return decode(result.Stdout)
}

func (b *Bridge) absolute(parts []domain.ArgvPart) []string {
	argv := make([]string, 0, len(parts))
	for _, part := range parts {
		if part.Path != "" {
			argv = append(argv, filepath.Join(b.settings.WorkspaceDir, filepath.FromSlash(part.Path)))
			continue
		}
		argv = append(argv, part.Literal)
	}
	return argv
}

func (b *Bridge) flags(action domain.Action) []string {
	var flags []string
	if action.Catalog {
		flags = append(flags, "--domains-dir", filepath.Join(b.settings.WorkspaceDir, "domains"), "--adapters-dir", b.settings.AdaptersDir)
	}
	if action.Library {
		flags = append(flags, b.libraryFlags()...)
	}
	if action.Kinds {
		flags = append(flags, b.workspaceDirFlag("--kinds-dir", "kinds")...)
	}
	return flags
}

func (b *Bridge) libraryFlags() []string {
	var flags []string
	for _, dir := range b.settings.ComponentDirs {
		flags = append(flags, "--components-dir", dir)
	}
	return append(flags, b.workspaceDirFlag("--components-dir", "components")...)
}

func (b *Bridge) workspaceDirFlag(flag, folder string) []string {
	if !b.edges.Workspace.Exists(folder) {
		return nil
	}
	return []string{flag, filepath.Join(b.settings.WorkspaceDir, folder)}
}

func (b *Bridge) DirectorArguments() []string {
	args := []string{"mcp", "--role", "director", "--domains-dir", filepath.Join(b.settings.WorkspaceDir, "domains"),
		"--adapters-dir", b.settings.AdaptersDir, "--out", filepath.Join(b.settings.WorkspaceDir, "runs")}
	for _, dir := range b.settings.ComponentDirs {
		args = append(args, "--components-dir", dir)
	}
	return append(args, b.workspaceDirFlag("--kinds-dir", "kinds")...)
}
