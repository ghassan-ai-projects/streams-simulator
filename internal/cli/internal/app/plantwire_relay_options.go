package app

import (
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

type plantWireRelayOptions struct {
	gatewaySocket string
	node          plantWireNodeOptions
}

func parsePlantWireRelayOptions(args []string, stderr io.Writer) (plantWireRelayOptions, error) {
	var o plantWireRelayOptions
	fs := newFlagSet("plantwire relay", stderr)
	fs.StringVar(&o.gatewaySocket, "gateway-socket", "", "owner-only unix socket the relay creates and the gateway dials (required)")
	registerPlantWireNodeFlags(fs, &o.node)
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	return o, o.check()
}

func (o plantWireRelayOptions) check() error {
	switch {
	case o.gatewaySocket == "" || o.node.hostSocket == "":
		return usagef("plantwire relay requires --gateway-socket and --host-socket")
	case o.node.domain != "" && o.node.out == "":
		return usagef("a mixed relay (--domain) requires --out, the run folder of its artifact and field trace")
	}
	return nil
}

func (o plantWireRelayOptions) relay() plantwire.RelayOptions {
	options := plantwire.RelayOptions{GatewaySocket: o.gatewaySocket, HostSocket: o.node.hostSocket, Out: o.node.out}
	if o.node.domain != "" {
		options.Simulated = &plantwire.SimulatedOptions{DomainsDir: o.node.domainsDir, AdaptersDir: o.node.adaptersDir,
			Domain: o.node.domain, Adapter: o.node.adapter, Seed: o.node.seed,
			BarrierCap: time.Duration(o.node.capMS) * time.Millisecond, Node: o.node.node()}
	}
	return options
}
