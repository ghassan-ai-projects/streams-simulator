package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
)

const connectUsage = "usage: streamsim plantwire connect (--plan <file> | --stdin) [--check] [--out <dir>] [--domains-dir d] [--adapters-dir a] [--kinds-dir k]"

type plantWireConnectOptions struct {
	plan, out, domainsDir, adaptersDir string
	stdin, check                       bool
	sliceS                             float64
	capMS                              int64
	kinds, policies                    listFlag
}

func cmdPlantWireConnect(s *session, args []string) (any, error) {
	o, err := parsePlantWireConnectOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	options, err := o.connect(s)
	if err != nil {
		return nil, err
	}
	if o.check {
		return connectDocument(plantwire.CheckConnect(options))
	}
	return connectDocument(plantwire.Connect(context.Background(), options))
}

func connectDocument(value any, err error) (any, error) {
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return value, nil
}

func parsePlantWireConnectOptions(args []string, stderr io.Writer) (plantWireConnectOptions, error) {
	var o plantWireConnectOptions
	fs := newFlagSet("plantwire connect", stderr)
	registerConnectFlags(fs, &o)
	if err := parseOnlyFlags(fs, args); err != nil {
		return o, err
	}
	return o, o.valid()
}

func registerConnectFlags(fs *flag.FlagSet, o *plantWireConnectOptions) {
	fs.StringVar(&o.plan, "plan", "", "session plan file (session-plan-v0.1)")
	fs.BoolVar(&o.stdin, "stdin", false, "read the session plan from standard input instead of a file")
	fs.BoolVar(&o.check, "check", false, "check the plan against its domain and kind files, start nothing")
	registerRunFolder(fs, &o.out, "")
	registerDomainsDir(fs, &o.domainsDir, "domains")
	registerAdaptersDir(fs, &o.adaptersDir, "adapters")
	registerKindsDirs(fs, &o.kinds)
	fs.Var(&o.policies, "policy", "allowed policy digest of commands to simulated devices; repeatable")
	fs.Float64Var(&o.sliceS, "slice-s", 1, "world seconds per slice")
	fs.Int64Var(&o.capMS, "barrier-cap-ms", 30000, "wall-time cap of each barrier")
}

func (o plantWireConnectOptions) valid() error {
	if (o.plan == "") == !o.stdin || (!o.check && o.out == "") {
		return usagef(connectUsage)
	}
	return nil
}

func (o plantWireConnectOptions) connect(s *session) (plantwire.ConnectOptions, error) {
	plan, source := s.input(), "stdin"
	if o.plan != "" {
		raw, err := files.Read(o.plan)
		if err != nil {
			return plantwire.ConnectOptions{}, fmt.Errorf("streamsim: %w", err)
		}
		plan, source = strings.NewReader(string(raw)), o.plan
	}
	return plantwire.ConnectOptions{Plan: plan, Source: source, DomainsDir: o.domainsDir, AdaptersDir: o.adaptersDir, KindsDirs: o.kinds.items(),
		Policies: o.policies.items(), Out: o.out, SliceNS: int64(o.sliceS * 1e9), BarrierCap: time.Duration(o.capMS) * time.Millisecond}, nil
}
