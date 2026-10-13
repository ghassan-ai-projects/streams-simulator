package app

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

func (b *Bridge) RunAction(ctx context.Context, name string, args map[string]any) (json.RawMessage, error) {
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
	argv := append(b.absolute(parts), b.flags(action)...)
	result, err := b.edges.RunCommand(ctx, argv)
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
		for _, dir := range b.settings.ComponentDirs {
			flags = append(flags, "--components-dir", dir)
		}
	}
	return flags
}

func (b *Bridge) DirectorArguments() []string {
	args := []string{"mcp", "--role", "director", "--domains-dir", filepath.Join(b.settings.WorkspaceDir, "domains"),
		"--adapters-dir", b.settings.AdaptersDir, "--out", filepath.Join(b.settings.WorkspaceDir, "runs")}
	for _, dir := range b.settings.ComponentDirs {
		args = append(args, "--components-dir", dir)
	}
	return args
}
