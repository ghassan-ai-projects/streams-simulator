package protocol

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// startOperatorEndpoint binds the operator role over streamable HTTP, the
// same wiring the CLI's --operator-addr uses.
func startOperatorEndpoint(t *testing.T, d *director.Director) string {
	t.Helper()
	opServer := NewOperatorServerResolver(d)
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return opServer }, nil)
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	endpoint := "http://" + ln.Addr().String()
	d.SetOperatorEndpoint(endpoint)
	return endpoint
}

// connectOperatorClient opens a real MCP client session to the operator
// endpoint over HTTP.
func connectOperatorClient(t *testing.T, endpoint string) *mcpsdk.ClientSession {
	t.Helper()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "e2e-test", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{Endpoint: endpoint}, nil)
	if err != nil {
		t.Fatalf("operator client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestOperatorEndpointServesOnlyOperatorTools(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	endpoint := startOperatorEndpoint(t, d)
	worldID := createWorld(t, d)
	_ = worldID
	session := connectOperatorClient(t, endpoint)
	res, err := session.ListTools(context.Background(), &mcpsdk.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"sim.nameplate.read", "sim.effector.list", "sim.effector.invoke", "sim.consumer.report"} {
		if !names[want] {
			t.Fatalf("operator endpoint missing %s", want)
		}
	}
	for _, forbidden := range []string{"sim.world.create", "sim.fault.inject", "sim.truth.reveal", "sim.score", "sim.clock.advance", "sim.perturb.apply"} {
		if names[forbidden] {
			t.Fatalf("operator endpoint must not expose %s", forbidden)
		}
	}
}
