package protocol

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
)

func TestDirectorServesItsToolsUntilTheSessionEndsThenStopsTheOperatorEndpoint(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	client, server := mcpsdk.NewInMemoryTransports()
	done := make(chan error, 1)
	d := director.New(t.Context(), nil, nil, t.TempDir())
	go func() { done <- serveDirectorOn(t.Context(), &stderr, d, "127.0.0.1:0", server) }()

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
	endpoint, ok := strings.CutPrefix(strings.TrimSpace(stderr.String()), "operator endpoint: http://")
	if !ok {
		t.Fatalf("stderr = %q", stderr.String())
	}
	conn, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", endpoint)
	if err == nil {
		_ = conn.Close()
		t.Fatalf("operator endpoint %s still accepts connections after the session ended", endpoint)
	}
}

func TestDirectorRefusesAnOperatorAddressItCannotListenOn(t *testing.T) {
	t.Parallel()
	_, server := mcpsdk.NewInMemoryTransports()
	d := director.New(t.Context(), nil, nil, t.TempDir())
	err := serveDirectorOn(t.Context(), &bytes.Buffer{}, d, "not-an-address", server)
	if err == nil || !strings.Contains(err.Error(), "mcp operator listener:") {
		t.Fatalf("err = %v", err)
	}
}

func TestServeDirectorRefusesInputsItCannotLoad(t *testing.T) {
	t.Parallel()
	err := ServeDirector(t.Context(), &bytes.Buffer{}, ServeOptions{DomainsDir: t.TempDir(), AdaptersDir: t.TempDir(), OutDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "no domains found in") {
		t.Fatalf("an empty domains directory: err = %v", err)
	}
}
