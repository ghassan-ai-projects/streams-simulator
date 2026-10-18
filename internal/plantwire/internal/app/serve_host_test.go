package app

import (
	"context"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAReferenceHostCancelledBeforeANodeReportsNothing(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	report, err := ServeHost(ctx, testsupport.SocketPath(t), 0, host.Options{Environment: "simulated", Ready: cancel})
	if err == nil || report.Refused != "" || len(report.Commands) != 0 {
		t.Fatalf("ServeHost = %+v, %v", report, err)
	}
}
