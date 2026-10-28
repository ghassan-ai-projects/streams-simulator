package app

import (
	"errors"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAHostThatStopsHeartbeatingIsLostAtTheBarrierLongBeforeTheCap(t *testing.T) {
	t.Parallel()
	path := testsupport.SocketPath(t)
	host := serveWireHost(t, path, true)
	served := make(chan error, 1)
	r := utilityRun(t)
	go func() { served <- Serve(t.Context(), r, path, time.Minute, sliceOptions(3)) }()
	host.waitFor(func(received []map[string]any) bool { return hasRecord(received, "NSYNC", "", nil) })
	started := time.Now()
	host.beat(1, 100)
	if err := <-served; !errors.Is(err, ErrHostLost) || !errors.Is(err, uds.ErrHostSilent) {
		t.Fatalf("Serve = %v, want host loss by silence", err)
	}
	if lost := time.Since(started); lost > 10*time.Second {
		t.Fatalf("the silent host was lost after %v, want three heartbeat intervals, not the barrier cap", lost)
	}
}
