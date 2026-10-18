package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

func cmdPlantWireRelay(s *session, args []string) (any, error) {
	options, err := parsePlantWireRelayOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	stats, err := plantwire.Relay(context.Background(), options.relay())
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return stats, nil
}
