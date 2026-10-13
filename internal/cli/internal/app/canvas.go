package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge"
)

type canvasOptions struct {
	workspace, uiDir, addr, adaptersDir string
	componentsDirs                      listFlag
}

func cmdCanvas(s *session, args []string) (any, error) {
	options, err := parseCanvasOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	err = canvasbridge.Serve(context.Background(), canvasbridge.Config{Workspace: options.workspace, UIDir: options.uiDir,
		Addr: options.addr, AdaptersDir: options.adaptersDir, ComponentDirs: s.componentDirs(options.componentsDirs), Stderr: s.stderr})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return nil, nil
}

func parseCanvasOptions(args []string, stderr io.Writer) (canvasOptions, error) {
	var options canvasOptions
	fs := newFlagSet("canvas", stderr)
	fs.StringVar(&options.workspace, "workspace", "", "workspace folder: domains, compositions, scenarios, layouts and runs (required)")
	fs.StringVar(&options.uiDir, "ui-dir", "", "folder of the built canvas UI (required)")
	fs.StringVar(&options.addr, "addr", "127.0.0.1:0", "loopback listen address")
	registerAdaptersDir(fs, &options.adaptersDir, "adapters")
	registerComponentDirs(fs, &options.componentsDirs)
	if err := parseOnlyFlags(fs, args); err != nil {
		return canvasOptions{}, err
	}
	return options, requireCanvasFolders(options)
}

func requireCanvasFolders(options canvasOptions) error {
	if options.workspace == "" || options.uiDir == "" {
		return usagef("canvas requires --workspace and --ui-dir")
	}
	return nil
}
