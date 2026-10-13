package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/entropy"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/transport"
)

type Config struct {
	Workspace     string
	UIDir         string
	Addr          string
	AdaptersDir   string
	ComponentDirs []string
	Stderr        io.Writer
}

var workspaceFolders = []string{"domains", "compositions", "scenarios", "layouts", "runs"}

func Serve(ctx context.Context, cfg Config) error {
	ctx, stop := process.UntilInterrupted(ctx)
	defer stop()
	bridge, err := open(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = bridge.Close() }()
	listener, err := transport.Listen(ctx, cfg.Addr)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cfg.Stderr, "canvas: http://%s\ncanvas: pairing code %s\n", listener.Addr(), domain.FormatCode(bridge.settings.PairingCode))
	return transport.Serve(ctx, listener, bridge, cfg.UIDir)
}

func open(cfg Config) (*Bridge, error) {
	workspace, err := openWorkspace(cfg.Workspace)
	if err != nil {
		return nil, err
	}
	settings, err := settingsOf(cfg)
	if err != nil {
		return nil, err
	}
	executable, err := process.Self()
	if err != nil {
		return nil, err
	}
	return wire(workspace, executable, settings), nil
}

func settingsOf(cfg Config) (Settings, error) {
	dir, err := files.Absolute(cfg.Workspace)
	if err != nil {
		return Settings{}, err
	}
	adapters, err := files.Absolute(cfg.AdaptersDir)
	if err != nil {
		return Settings{}, err
	}
	code, err := entropy.PairingCode()
	return Settings{WorkspaceDir: dir, AdaptersDir: adapters, ComponentDirs: cfg.ComponentDirs, PairingCode: code}, err
}

func wire(workspace *files.Workspace, executable string, settings Settings) *Bridge {
	bridge := New(Edges{Workspace: workspace, NewSession: entropy.SessionID, Tick: process.Ticker}, settings)
	bridge.edges.RunCommand = func(ctx context.Context, argv []string) (process.CommandResult, error) {
		return process.Run(ctx, executable, settings.WorkspaceDir, argv)
	}
	bridge.edges.StartDirector = func(ctx context.Context) (Director, error) {
		return process.StartDirector(context.WithoutCancel(ctx), executable, bridge.DirectorArguments())
	}
	return bridge
}

func openWorkspace(dir string) (*files.Workspace, error) {
	workspace, err := files.Open(dir)
	if err != nil {
		return nil, err
	}
	return workspace, workspace.Ensure(workspaceFolders)
}
