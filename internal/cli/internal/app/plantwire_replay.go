package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

type plantWireReplayOptions struct {
	run, hostSocket, sessionID string
	capMS                      int64
}

func cmdPlantWireReplayTrace(s *session, args []string) (any, error) {
	o, err := parsePlantWireReplayOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	stats, err := plantwire.ReplayTrace(context.Background(), plantwire.ReplayOptions{Run: o.run, HostSocket: o.hostSocket,
		SessionID: o.sessionID, BarrierCap: time.Duration(o.capMS) * time.Millisecond})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return stats, nil
}

func parsePlantWireReplayOptions(args []string, stderr io.Writer) (plantWireReplayOptions, error) {
	var o plantWireReplayOptions
	fs := newFlagSet("plantwire replay-trace", stderr)
	fs.StringVar(&o.run, "run", "", "run folder holding field-trace.jsonl, checked against its run.json when there is one (required)")
	fs.StringVar(&o.hostSocket, "host-socket", "", "the host's owner-only unix socket (required)")
	fs.StringVar(&o.sessionID, "session-id", "replay-1", "session id")
	fs.Int64Var(&o.capMS, "barrier-cap-ms", 30000, "wall-time cap of each barrier")
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	if o.run == "" || o.hostSocket == "" {
		return o, usagef("plantwire replay-trace requires --run and --host-socket")
	}
	return o, nil
}
