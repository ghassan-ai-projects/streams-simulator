package plantwire

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestARelayStopsWithItsContextAndReportsWhatItCarried(t *testing.T) {
	t.Parallel()
	hostPath := testsupport.SocketPath(t)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	stats, err := Relay(ctx, RelayOptions{GatewaySocket: filepath.Join(filepath.Dir(hostPath), "g.sock"), HostSocket: hostPath, Ready: cancel})
	if err != nil || stats != (RelayStats{}) {
		t.Fatalf("Relay = %+v, %v, want an empty report and no error from a relay that served nothing", stats, err)
	}
}

func TestAReplayOfAFolderWithoutAFieldTraceIsRefusedAsMissing(t *testing.T) {
	t.Parallel()
	options := ReplayOptions{Run: t.TempDir(), HostSocket: testsupport.SocketPath(t), SessionID: "replay-1", BarrierCap: time.Second}
	if _, err := ReplayTrace(t.Context(), options); !errors.Is(err, run.ErrFieldTraceMissing) {
		t.Fatalf("ReplayTrace = %v, want %v", err, run.ErrFieldTraceMissing)
	}
}
