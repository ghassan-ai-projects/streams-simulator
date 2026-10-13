package app

import (
	"context"
	"fmt"
	"io"

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
		return nil, usagef("the operator role is served from a director process; start `streamsim mcp --role director --operator-addr 127.0.0.1:PORT` and use the sim.world.create token")
	default:
		return nil, usagef("mcp requires --role director|operator")
	}
}

type mcpOptions struct {
	role, operatorAddr, domainsDir, adaptersDir, outDir string
	componentsDirs                                      listFlag
}

func parseMCPOptions(args []string, stderr io.Writer) (mcpOptions, error) {
	var options mcpOptions
	fs := newFlagSet("mcp", stderr)
	fs.StringVar(&options.role, "role", "", "director")
	fs.StringVar(&options.operatorAddr, "operator-addr", "", "serve the operator role over streamable HTTP at this listen address (e.g. 127.0.0.1:0); sim.world.create returns the endpoint with the token")
	registerInputDirs(fs, &options.domainsDir, &options.adaptersDir)
	registerRunFolder(fs, &options.outDir, "runs")
	registerComponentDirs(fs, &options.componentsDirs)
	if err := parseOnlyFlags(fs, args); err != nil {
		return mcpOptions{}, err
	}
	return options, nil
}

func serveDirector(s *session, options mcpOptions) error {
	err := mcp.ServeDirector(context.Background(), s.stderr, mcp.ServeOptions{DomainsDir: options.domainsDir,
		AdaptersDir: options.adaptersDir, OutDir: options.outDir, OperatorAddr: options.operatorAddr,
		ComponentsDirs: s.declaredComponentDirs(options.componentsDirs)})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
