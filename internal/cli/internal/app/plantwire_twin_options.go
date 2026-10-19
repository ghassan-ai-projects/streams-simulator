package app

import "io"

const twinReportUsage = "usage: streamsim plantwire twin-report <run-dir> --domain <id> --kinds-dir <dir> [--domains-dir d] [--tolerance selector.metric=value] [--span-fraction f]"

type plantWireTwinOptions struct {
	run, domain, domainsDir string
	spanFraction            float64
	kinds, tolerances       listFlag
}

func parsePlantWireTwinOptions(args []string, stderr io.Writer) (plantWireTwinOptions, error) {
	var o plantWireTwinOptions
	fs := newFlagSet("plantwire twin-report", stderr)
	fs.StringVar(&o.domain, "domain", "", "id of the twin: the domain that implements the board's kinds (required)")
	registerDomainsDir(fs, &o.domainsDir, "domains")
	registerKindsDirs(fs, &o.kinds)
	fs.Var(&o.tolerances, "tolerance", "absolute tolerance of a metric, device.metric=value or kind.metric=value; repeatable or comma separated")
	fs.Float64Var(&o.spanFraction, "span-fraction", 0, "default tolerance as a fraction of a metric's eu_range span, in (0, 1]; 0 keeps the binary's default")
	if err := parseFlags(fs, args); err != nil {
		return o, err
	}
	return withRunDir(o, fs.Args())
}

func withRunDir(o plantWireTwinOptions, positional []string) (plantWireTwinOptions, error) {
	if len(positional) != 1 || o.domain == "" || len(o.kinds.items()) == 0 {
		return o, usagef(twinReportUsage)
	}
	o.run = positional[0]
	return o, nil
}
