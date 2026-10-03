package cli

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
)

func cmdMCP(args []string) error {
	options, err := parseMCPOptions(args)
	if err != nil {
		return err
	}
	switch options.role {
	case "director":
		return serveDirector(options)
	case "operator":
		return fmt.Errorf("the operator role is served from a director process; start `streamsim mcp --role director --operator-addr 127.0.0.1:PORT` and use the sim.world.create token")
	default:
		return fmt.Errorf("mcp requires --role director|operator")
	}
}

// serveOperatorEndpoint binds the operator role over streamable HTTP so a
// consumer process can connect with the capability token from
// sim.world.create. sim.world.create responses include the endpoint; the
// returned server is closed when the director stdio session ends.
func serveOperatorEndpoint(d *mcp.Director, addr string) (*http.Server, error) {
	opServer := mcp.NewOperatorServerResolver(d)
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return opServer }, nil)
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("mcp operator listener: %w", err)
	}
	endpoint := "http://" + ln.Addr().String()
	d.SetOperatorEndpoint(endpoint)
	fmt.Fprintf(os.Stderr, "operator endpoint: %s\n", endpoint)
	httpServer := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go runOperatorHTTP(httpServer, ln)
	return httpServer, nil
}

func runOperatorHTTP(server *http.Server, listener net.Listener) {
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "operator endpoint failed: %v\n", err)
	}
}

type mcpOptions struct{ role, operatorAddr, domainsDir, adaptersDir, outDir string }

func parseMCPOptions(args []string) (mcpOptions, error) {
	var options mcpOptions
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	fs.StringVar(&options.role, "role", "", "director")
	fs.StringVar(&options.operatorAddr, "operator-addr", "", "serve the operator role over streamable HTTP at this listen address (e.g. 127.0.0.1:0); sim.world.create returns the endpoint with the token")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.outDir, "out", "runs", "run artifact output directory")
	if err := fs.Parse(args); err != nil {
		return mcpOptions{}, fmt.Errorf("streamsim: %w", err)
	}
	return options, nil
}

func serveDirector(options mcpOptions) error {
	d, err := loadDirector(options)
	if err != nil {
		return err
	}
	srv := mcp.NewDirectorServer(d)
	var httpServer *http.Server
	if options.operatorAddr != "" {
		httpServer, err = serveOperatorEndpoint(d, options.operatorAddr)
		if err != nil {
			return err
		}
	}
	return runDirectorSession(srv, httpServer)
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

func runDirectorSession(server *mcpsdk.Server, operator *http.Server) error {
	err := server.Run(context.Background(), &mcpsdk.StdioTransport{})
	if operator != nil {
		_ = operator.Close()
	}
	if err != nil {
		return fmt.Errorf("mcp director: %w", err)
	}
	return nil
}
