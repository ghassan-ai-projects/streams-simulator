package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/uds"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type RelayOptions struct {
	GatewaySocket string
	HostSocket    string
	Out           string
	Ready         func()
}

const (
	relayTick = 250 * time.Millisecond
	relayWait = 30 * time.Second
)

func Relay(ctx context.Context, options RelayOptions) (domain.RelayStats, error) {
	ctx, stop := uds.UntilInterrupted(ctx)
	defer stop()
	ticks, stopTicks := uds.Ticks(relayTick)
	defer stopTicks()
	return serveRelay(ctx, options, ticks)
}

func serveRelay(ctx context.Context, options RelayOptions, ticks <-chan time.Time) (domain.RelayStats, error) {
	loop, err := openRelay(ctx, options, ticks)
	if err != nil {
		return domain.RelayStats{}, fmt.Errorf("plantwire relay: %w", err)
	}
	err = loop.serve(ctx)
	return loop.relay.Stats(), err
}

func openRelay(ctx context.Context, options RelayOptions, ticks <-chan time.Time) (*relayLoop, error) {
	if err := checkRelayOptions(options); err != nil {
		return nil, err
	}
	trace, err := openTrace(options.Out)
	if err != nil {
		return nil, err
	}
	gateways, err := listenGateways(ctx, options, trace)
	if err != nil {
		return nil, err
	}
	return &relayLoop{relay: domain.NewRelay(relayTick.Milliseconds()), options: options, trace: trace, gateways: gateways, ticks: ticks}, nil
}

func listenGateways(ctx context.Context, options RelayOptions, trace *run.WireTrace) (*uds.Gateways, error) {
	gateways, err := uds.ListenGateways(ctx, options.GatewaySocket, relayWait)
	if err != nil {
		return nil, errors.Join(err, closeTrace(trace))
	}
	if options.Ready != nil {
		options.Ready()
	}
	return gateways, nil
}

func checkRelayOptions(o RelayOptions) error {
	if o.GatewaySocket == "" || o.HostSocket == "" || o.GatewaySocket == o.HostSocket {
		return fmt.Errorf("%w: a relay needs two different sockets, the gateway's and the host's", ErrBadOptions)
	}
	return nil
}

func openTrace(dir string) (*run.WireTrace, error) {
	if dir == "" {
		return nil, nil
	}
	trace, err := run.NewWireTrace(dir)
	if err != nil {
		return nil, fmt.Errorf("plantwire relay: %w", err)
	}
	return trace, nil
}

func closeTrace(trace *run.WireTrace) error {
	if trace == nil {
		return nil
	}
	if err := trace.Close(); err != nil {
		return fmt.Errorf("plantwire relay: close trace: %w", err)
	}
	return nil
}
