package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

type plantWireRelayOptions struct {
	gatewaySocket, hostSocket, out string
}

func cmdPlantWireRelay(s *session, args []string) (any, error) {
	options, err := parsePlantWireRelayOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	stats, err := plantwire.Relay(context.Background(), plantwire.RelayOptions{GatewaySocket: options.gatewaySocket,
		HostSocket: options.hostSocket, Out: options.out})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return stats, nil
}

func parsePlantWireRelayOptions(args []string, stderr io.Writer) (plantWireRelayOptions, error) {
	var o plantWireRelayOptions
	fs := newFlagSet("plantwire relay", stderr)
	fs.StringVar(&o.gatewaySocket, "gateway-socket", "", "owner-only unix socket the relay creates and the gateway dials (required)")
	fs.StringVar(&o.hostSocket, "host-socket", "", "the host's owner-only unix socket the relay dials (required)")
	fs.StringVar(&o.out, "out", "", "folder for the journal of every relayed record")
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	if o.gatewaySocket == "" || o.hostSocket == "" {
		return o, usagef("plantwire relay requires --gateway-socket and --host-socket")
	}
	return o, nil
}
