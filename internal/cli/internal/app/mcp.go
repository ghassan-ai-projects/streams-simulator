package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/serve"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
)

func cmdMCP(s *session, args []string) (any, error) {
	options, err := parseMCPOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	switch options.role {
	case "director":
		return nil, serveDirector(s, options)
	case "operator":
		return nil, fmt.Errorf("the operator role is served from a director process; start `streamsim mcp --role director --operator-addr 127.0.0.1:PORT` and use the sim.world.create token")
	default:
		return nil, fmt.Errorf("mcp requires --role director|operator")
	}
}

type mcpOptions struct{ role, operatorAddr, domainsDir, adaptersDir, outDir string }

func parseMCPOptions(args []string, stderr io.Writer) (mcpOptions, error) {
	var options mcpOptions
	fs := newFlagSet("mcp", stderr)
	fs.StringVar(&options.role, "role", "", "director")
	fs.StringVar(&options.operatorAddr, "operator-addr", "", "serve the operator role over streamable HTTP at this listen address (e.g. 127.0.0.1:0); sim.world.create returns the endpoint with the token")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.outDir, "out", "runs", "run artifact output directory")
	if err := parseFlags(fs, args); err != nil {
		return mcpOptions{}, err
	}
	return options, nil
}

func serveDirector(s *session, options mcpOptions) error {
	d, err := loadDirector(options)
	if err != nil {
		return err
	}
	return serve.Director(s.stderr, d, options.operatorAddr)
}

func loadDirector(options mcpOptions) (*mcp.Director, error) {
	cat, err := loadCatalog(options.domainsDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(options.adaptersDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return mcp.NewDirector(context.Background(), cat, adapters, options.outDir), nil
}
