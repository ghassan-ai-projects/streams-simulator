package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

var plantWireCommands = map[string]func(*session, []string) (any, error){"node": cmdPlantWireNode, "host": cmdPlantWireHost,
	"relay": cmdPlantWireRelay, "replay-trace": cmdPlantWireReplayTrace, "connect": cmdPlantWireConnect, "twin-report": cmdPlantWireTwinReport}

func cmdPlantWire(s *session, args []string) (any, error) {
	if len(args) == 0 {
		return nil, usagef("plantwire requires a subcommand: node, host, relay, replay-trace, connect or twin-report")
	}
	if command, known := plantWireCommands[args[0]]; known {
		return command(s, args[1:])
	}
	return nil, usagef("plantwire: unknown subcommand %q (want: node, host, relay, replay-trace, connect, twin-report)", args[0])
}

func cmdPlantWireNode(s *session, args []string) (any, error) {
	options, err := parsePlantWireNodeOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	r, err := nodeRun(options)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return serveNode(r, options)
}

func nodeRun(options plantWireNodeOptions) (*run.Run, error) {
	spec, adapterSpec, err := run.LoadInputs(options.domainsDir, options.adaptersDir, options.domain, options.adapter)
	if err != nil {
		return nil, fmt.Errorf("plantwire node: %w", err)
	}
	r, err := run.New(context.Background(), run.Config{Domain: spec, Adapter: adapterSpec, Seed: options.seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS, JournalDir: options.out})
	if err != nil {
		return nil, fmt.Errorf("plantwire node: %w", err)
	}
	return r, nil
}

func serveNode(r *run.Run, options plantWireNodeOptions) (any, error) {
	served := plantwire.Serve(context.Background(), r, options.hostSocket, time.Duration(options.capMS)*time.Millisecond, options.node())
	_, ended := r.End(options.out)
	if err := errors.Join(served, ended); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"run_id": r.ID, "out": options.out}, nil
}

func (o plantWireNodeOptions) node() plantwire.Options {
	start := model.DefaultStartTimeNS
	return plantwire.Options{Environment: "simulated", ProtocolMax: 1, SessionID: o.sessionID,
		SliceNS: int64(o.sliceS * 1e9), UntilNS: start + int64(o.durationS*1e9), Policies: o.policies.items(), KindsDirs: o.kinds.items()}
}

func cmdPlantWireHost(s *session, args []string) (any, error) {
	options, err := parsePlantWireHostOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	host, err := hostOptions(options)
	if err != nil {
		return nil, err
	}
	return serveHost(options, host)
}

func serveHost(options plantWireHostOptions, host plantwire.HostOptions) (any, error) {
	report, err := plantwire.Host(context.Background(), options.socket, time.Duration(options.delayMS)*time.Millisecond, host)
	if errors.Is(err, plantwire.ErrHostPolicy) {
		return nil, usagef("%v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return report, nil
}

func hostOptions(options plantWireHostOptions) (plantwire.HostOptions, error) {
	base := plantwire.HostOptions{Environment: options.environment, ProtocolMax: 1, Epoch: options.epoch, Policy: options.policy}
	host, err := plantwire.LoadHost(base, options.script, options.inventory)
	if errors.Is(err, plantwire.ErrHostInput) {
		return host, usagef("%v", err)
	}
	if err != nil {
		return host, fmt.Errorf("streamsim: %w", err)
	}
	return host, nil
}
