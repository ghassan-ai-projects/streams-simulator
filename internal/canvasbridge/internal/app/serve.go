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

var workspaceFolders = append(domain.DefinitionFolders(), "runs")

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
	return transport.Serve(ctx, listener, boundary{bridge}, cfg.UIDir)
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
	return Settings{WorkspaceDir: dir, WorkspaceAliases: files.Aliases(dir), AdaptersDir: adapters, ComponentDirs: cfg.ComponentDirs, PairingCode: entropy.PairingCode()}, nil
}

func wire(workspace *files.Workspace, executable string, settings Settings) *Bridge {
	bridge := New(Edges{Workspace: workspace, NewSession: entropy.SessionID, NewRunID: entropy.RunID, Tick: process.Ticker,
		StartLasting: startLasting(executable, settings.WorkspaceDir), NewControlSocket: process.NewControlSocket,
		ControlReady: process.ControlReady, SendControl: process.SendControl}, settings)
	bridge.edges.RunCommand = func(ctx context.Context, argv []string, stdin []byte) (process.CommandResult, error) {
		return process.RunWithInput(ctx, executable, settings.WorkspaceDir, argv, stdin)
	}
	bridge.edges.StartDirector = func(ctx context.Context) (Director, error) {
		return process.StartDirector(context.WithoutCancel(ctx), executable, bridge.DirectorArguments())
	}
	return bridge
}

func startLasting(executable, dir string) func(argv []string, stdin []byte) (Lasting, error) {
	return func(argv []string, stdin []byte) (Lasting, error) {
		lasting, err := process.StartLasting(executable, dir, argv, stdin)
		if err != nil {
			return nil, err
		}
		return lasting, nil
	}
}

func openWorkspace(dir string) (*files.Workspace, error) {
	workspace, err := files.Open(dir)
	if err != nil {
		return nil, err
	}
	return workspace, workspace.Ensure(workspaceFolders)
}
