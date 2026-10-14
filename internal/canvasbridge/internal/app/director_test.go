package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

func TestTheDirectorRestartsForANewCatalogOnlyWhileNoWorldIsOpen(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.bridge.RunAction(ctx, "catalog.list", nil); err != nil {
		t.Fatal(err)
	}
	h.directors[0].results["sim.world.create"] = process.ToolResult{Structured: []byte(`{"world_id":"w-1","token":"t-secret"}`)}
	created, err := h.bridge.RunAction(ctx, "world.create", map[string]any{"domain": "plant"})
	if err != nil || strings.Contains(string(created), "t-secret") {
		t.Fatalf("world.create = %s, %v; the operator token must be withheld", created, err)
	}
	h.write(t, plant, "{}")
	if _, err := h.bridge.RunAction(ctx, "catalog.list", nil); err != nil || len(h.directors) != 1 {
		t.Fatalf("the director restarted while a world was open (%d directors), %v", len(h.directors), err)
	}
	if _, err := h.bridge.RunAction(ctx, "world.destroy", map[string]any{"world_id": "w-1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.bridge.RunAction(ctx, "catalog.list", nil); err != nil || len(h.directors) != 2 || !h.directors[0].closed {
		t.Fatalf("a changed catalog with no open world restarts the director: %d directors, %v", len(h.directors), err)
	}
	if err := h.bridge.Close(); err != nil || !h.directors[1].closed {
		t.Fatalf("Close: %v", err)
	}
}

func TestToolFailuresAndClosedArgumentsReachTheCaller(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.bridge.RunAction(ctx, "truth.reveal", map[string]any{"run_id": "r", "unblind": true}); !errors.Is(err, domain.ErrBadArguments) || len(h.directors) != 0 {
		t.Fatalf("unblind reached the director: %v", err)
	}
	if _, err := h.bridge.RunAction(ctx, "clock.state", map[string]any{"world_id": "w-9"}); err != nil {
		t.Fatal(err)
	}
	h.directors[0].results["sim.clock.state"] = process.ToolResult{Failure: "world_not_found: no world w-9"}
	var failure *domain.Failure
	if _, err := h.bridge.RunAction(ctx, "clock.state", map[string]any{"world_id": "w-9"}); !errors.As(err, &failure) || failure.Code != "world_not_found" {
		t.Fatalf("a tool failure: %v", err)
	}
	if !slices.Equal(h.directors[0].calls, []string{"sim.clock.state", "sim.clock.state"}) || len(h.bridge.Actions()) == 0 {
		t.Fatalf("calls = %v", h.directors[0].calls)
	}
	if args := h.bridge.DirectorArguments(); !slices.Contains(args, "--components-dir") || args[0] != "mcp" {
		t.Fatalf("director arguments = %v", args)
	}
}
