package transport

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

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
	probe, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- endpoint.server.Serve(probe) }()
	select {
	case err := <-served:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("operator endpoint %s: %v", announced, err)
		}
	case <-time.After(5 * time.Second):
		_ = probe.Close()
		t.Fatalf("operator endpoint %s still serves after the session ended", announced)
	}
}

func TestAnEndpointThatIsNotBoundClosesWithoutEffect(t *testing.T) {
	t.Parallel()
	var none *Endpoint
	none.Close()
	(&Endpoint{}).Close()
}

func TestCloseRefusesConnectionsOnReturnEvenWhenItRunsBeforeTheServerStarted(t *testing.T) {
	t.Parallel()
	for range 200 {
		var announced string
		endpoint, err := BindOperator(t.Context(), &bytes.Buffer{}, newServer(), "127.0.0.1:0", func(address string) { announced = address })
		if err != nil {
			t.Fatal(err)
		}
		endpoint.Close()
		endpoint.Close()
		conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", strings.TrimPrefix(announced, "http://"))
		if err == nil {
			_ = conn.Close()
			t.Fatalf("operator endpoint %s accepted a connection after Close returned", announced)
		}
	}
}
