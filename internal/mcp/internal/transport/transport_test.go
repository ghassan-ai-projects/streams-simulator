package transport

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func newServer() *mcpsdk.Server {
	return mcpsdk.NewServer(&mcpsdk.Implementation{Name: "transport-test"}, nil)
}

func TestBindOperatorAnnouncesItsAddressAndRefusesABadOne(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	var announced string
	endpoint, err := BindOperator(t.Context(), &stderr, newServer(), "127.0.0.1:0", func(address string) { announced = address })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(endpoint.Close)
	if !strings.HasPrefix(announced, "http://127.0.0.1:") || stderr.String() != "operator endpoint: "+announced+"\n" {
		t.Fatalf("announced %q, stderr %q", announced, stderr.String())
	}
	_, err = BindOperator(t.Context(), &stderr, newServer(), "not-an-address", func(string) { t.Fatal("a refused address must not be announced") })
	if err == nil || !strings.Contains(err.Error(), "mcp operator listener:") {
		t.Fatalf("a bad listen address must be refused: %v", err)
	}
}

func TestServeRunsTheDirectorUntilTheSessionEndsThenStopsTheOperatorEndpoint(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	var announced string
	endpoint, err := BindOperator(t.Context(), &stderr, newServer(), "127.0.0.1:0", func(address string) { announced = address })
	if err != nil {
		t.Fatal(err)
	}
	client, server := mcpsdk.NewInMemoryTransports()
	done := make(chan error, 1)
	go func() { done <- Serve(t.Context(), newServer(), endpoint, server) }()
	session, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "serve-test"}, nil).Connect(context.Background(), client, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("director session: %v", err)
	}
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", strings.TrimPrefix(announced, "http://"))
	if err == nil {
		_ = conn.Close()
		t.Fatalf("operator endpoint %s still accepts connections after the session ended", announced)
	}
}

func TestAnEndpointThatIsNotBoundClosesWithoutEffect(t *testing.T) {
	t.Parallel()
	var none *Endpoint
	none.Close()
	(&Endpoint{}).Close()
}
