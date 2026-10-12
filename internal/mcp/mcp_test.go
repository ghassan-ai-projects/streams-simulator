package mcp_test

import (
	"io"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/director"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/mcp"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestTheOperatorServerExposesTheOperatorTools(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.Example())
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	d := director.New(t.Context(), domain.NewCatalog([]*domain.Compiled{spec}), map[string]*model.Adapter{"native-jsonl": native}, t.TempDir())
	ct, st := mcpsdk.NewInMemoryTransports()
	ss, err := mcp.NewOperatorServerResolver(d).Connect(t.Context(), st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "facade-test"}, nil).Connect(t.Context(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	names := map[string]bool{}
	for tool, err := range cs.Tools(t.Context(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		names[tool.Name] = true
	}
	if !names["sim.nameplate.read"] || names["sim.truth.reveal"] {
		t.Fatalf("operator tools = %v", names)
	}
}

func TestServeDirectorRefusesADomainsDirectoryThatHoldsNoDomain(t *testing.T) {
	t.Parallel()
	err := mcp.ServeDirector(t.Context(), io.Discard, mcp.ServeOptions{DomainsDir: t.TempDir(), AdaptersDir: testsupport.AdaptersDir(), OutDir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "no domains found in") {
		t.Fatalf("err = %v", err)
	}
}
