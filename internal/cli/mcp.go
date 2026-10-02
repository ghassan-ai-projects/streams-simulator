package cli

import (
	"context"
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"net"
	"net/http"
	"os"
	"time"
)

func cmdMCP(args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	role := fs.String("role", "", "director")
	operatorAddr := fs.String("operator-addr", "", "serve the operator role over streamable HTTP at this listen address (e.g. 127.0.0.1:0); sim.world.create returns the endpoint with the token")
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	outDir := fs.String("out", "runs", "run artifact output directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	switch *role {
	case "director":
		cat, err := loadCatalog(*domainsDir)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		adapters, err := loadAdapters(*adaptersDir)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		d := mcp.NewDirector(context.Background(), cat, adapters, *outDir)
		srv := mcp.NewDirectorServer(d)
		var httpServer *http.Server
		if *operatorAddr != "" {
			httpServer, err = serveOperatorEndpoint(d, *operatorAddr)
			if err != nil {
				return err
			}
		}
		err = srv.Run(context.Background(), &mcpsdk.StdioTransport{})
		if httpServer != nil {
			_ = httpServer.Close()
		}
		if err != nil {
			return fmt.Errorf("mcp director: %w", err)
		}
		return nil
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
	go func() {
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "operator endpoint failed: %v\n", err)
		}
	}()
	return httpServer, nil
}
