package serve

import (
	"bytes"
	"context"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
)

func TestDirectorServesItsToolsUntilTheSessionEndsThenStopsTheOperatorEndpoint(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	client, server := mcpsdk.NewInMemoryTransports()
	done := make(chan error, 1)
	d := mcp.NewDirector(t.Context(), nil, nil, t.TempDir())
	go func() { done <- directorOn(&stderr, d, "127.0.0.1:0", server) }()

	session, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "serve-test"}, nil).Connect(context.Background(), client, nil)
	if err != nil {
		t.Fatal(err)
	}
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil || len(tools.Tools) == 0 {
		t.Fatalf("director tools = %v (%v)", tools, err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("director session: %v", err)
	}
	if !strings.Contains(stderr.String(), "operator endpoint: http://127.0.0.1:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDirectorRefusesAnOperatorAddressItCannotListenOn(t *testing.T) {
	t.Parallel()
	_, server := mcpsdk.NewInMemoryTransports()
	d := mcp.NewDirector(t.Context(), nil, nil, t.TempDir())
	err := directorOn(&bytes.Buffer{}, d, "not-an-address", server)
	if err == nil || !strings.Contains(err.Error(), "mcp operator listener:") {
		t.Fatalf("err = %v", err)
	}
}
