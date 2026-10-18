package app

import (
	"flag"
	"io"
)

type plantWireNodeOptions struct {
	hostSocket, domain, domainsDir, adapter, adaptersDir, out, sessionID string
	seed                                                                 uint64
	sliceS, durationS                                                    float64
	capMS                                                                int64
	policies, kinds                                                      listFlag
}

func parsePlantWireNodeOptions(args []string, stderr io.Writer) (plantWireNodeOptions, error) {
	var o plantWireNodeOptions
	fs := newFlagSet("plantwire node", stderr)
	registerPlantWireNodeFlags(fs, &o)
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	if o.hostSocket == "" || o.domain == "" || o.out == "" {
		return o, usagef("plantwire node requires --host-socket, --domain and --out")
	}
	return o, nil
}

func registerPlantWireNodeFlags(fs *flag.FlagSet, o *plantWireNodeOptions) {
	fs.StringVar(&o.hostSocket, "host-socket", "", "the host's owner-only unix socket (required)")
	fs.StringVar(&o.domain, "domain", "", "domain id (required)")
	registerDomainsDir(fs, &o.domainsDir, "domains")
	fs.StringVar(&o.adapter, "adapter", "native-jsonl", "adapter id")
	registerAdaptersDir(fs, &o.adaptersDir, "adapters")
	registerRunFolder(fs, &o.out, "")
	fs.StringVar(&o.sessionID, "session-id", "node-1", "session id")
	fs.Uint64Var(&o.seed, "seed", 1, "world seed")
	fs.Float64Var(&o.sliceS, "slice-s", 60, "world seconds per slice")
	fs.Float64Var(&o.durationS, "duration", 3600, "world seconds to run")
	fs.Int64Var(&o.capMS, "barrier-cap-ms", 30000, "wall-time cap of each barrier")
	fs.Var(&o.policies, "policy", "allowed policy digest; repeatable")
	registerKindsDirs(fs, &o.kinds)
}

type plantWireHostOptions struct {
	socket, script, inventory, environment, policy string
	epoch, delayMS                                 int64
}

func parsePlantWireHostOptions(args []string, stderr io.Writer) (plantWireHostOptions, error) {
	var o plantWireHostOptions
	fs := newFlagSet("plantwire host", stderr)
	registerPlantWireHostFlags(fs, &o)
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	if o.socket == "" {
		return o, usagef("plantwire host requires --socket")
	}
	return o, nil
}

func registerPlantWireHostFlags(fs *flag.FlagSet, o *plantWireHostOptions) {
	fs.StringVar(&o.socket, "socket", "", "unix socket to listen on, created owner-only (required)")
	fs.StringVar(&o.script, "script", "", "JSON array of scripted commands")
	fs.StringVar(&o.inventory, "inventory", "", "JSON object of commissioned devices to kinds")
	fs.StringVar(&o.environment, "environment", "simulated", "environment the host accepts: simulated, hardware or mixed (a mixed host accepts all three)")
	fs.StringVar(&o.policy, "policy", "", "policy digest the host signs commands with")
	fs.Int64Var(&o.epoch, "epoch", 1, "authority epoch")
	fs.Int64Var(&o.delayMS, "delay-ms", 0, "wall delay before each barrier answer")
}
