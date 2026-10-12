package acceptance

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	directormod "github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

// The composition tests drive a director in process through the MCP server
// layer, and reach the run beneath a world through Director.RunOf, the one
// documented test-harness handle: the closed-loop and covert-channel tests
// park and observe the run, which no surface may do. They live here, not in
// the director module, because they compose director and mcp.

// newCompositionDirector builds a director with the aquaculture-pond domain and the
// native-jsonl adapter.
func newCompositionDirector(t *testing.T) *directormod.Director {
	t.Helper()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	cat := domain.NewCatalog([]*domain.Compiled{spec})
	return directormod.New(t.Context(), cat, map[string]*model.Adapter{"native-jsonl": native}, t.TempDir())
}

// connect returns a client session to the server over in-memory transports.
func connect(t *testing.T, server *mcpsdk.Server) *mcpsdk.ClientSession {
	t.Helper()
	ctx := context.Background()
	ct, st := mcpsdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// startOperatorEndpoint binds the operator role over streamable HTTP, the
// same wiring the CLI's --operator-addr uses.
func startOperatorEndpoint(t *testing.T, d *directormod.Director) string {
	t.Helper()
	opServer := mcp.NewOperatorServerResolver(d)
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
