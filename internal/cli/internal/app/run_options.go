package app

import (
	"flag"
	"io"
	"path/filepath"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type runOptions struct {
	domainsDir, adaptersDir, domainID, adapterID string
	sinkName, sinkTarget, outDir, profile        string
	scenarioPath                                 string
	explore                                      bool
	faults, perts, effectors                     listFlag
	seed                                         uint64
	durationS                                    float64
	startTime                                    int64
	explicit                                     []string // flags given on the command line
}

func parseRunOptions(args []string, stderr io.Writer) (runOptions, error) {
	var options runOptions
	fs := newFlagSet("run", stderr)
	registerScenarioFlags(fs, &options)
	registerExecutionFlags(fs, &options)
	if err := parseOnlyFlags(fs, args); err != nil {
		return runOptions{}, err
	}
	fs.Visit(func(set *flag.Flag) { options.explicit = append(options.explicit, set.Name) })
	return admitRunOptions(options)
}

// registerScenarioFlags registers what a scenario defines: the inputs, the
// world settings and the scripted commands. A scenario file defines exactly
// these, so they are refused beside --scenario (scenarioDefinedFlags).
func registerScenarioFlags(fs *flag.FlagSet, options *runOptions) {
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.adapterID, "adapter", "native-jsonl", "adapter id")
	fs.Uint64Var(&options.seed, "seed", 1, "world seed")
	fs.StringVar(&options.profile, "profile", "", "scenario profile")
	fs.Int64Var(&options.startTime, "start-time", model.DefaultStartTimeNS, "world start (ns epoch)")
	fs.Float64Var(&options.durationS, "duration", 6*3600, "run duration (seconds)")
	fs.Var(&options.faults, "fault", "repeatable: entity=fault@offset_s")
	fs.Var(&options.perts, "perturb", "repeatable: name@from_s[@until_s]")
	fs.Var(&options.effectors, "effector", `repeatable: effector@entity@offset_s[@{"arg":value}]`)
}

// registerExecutionFlags registers how a scenario is executed: where its
// inputs are looked up, the scenario file itself and where the run goes.
func registerExecutionFlags(fs *flag.FlagSet, options *runOptions) {
	registerInputDirs(fs, &options.domainsDir, &options.adaptersDir)
	fs.StringVar(&options.scenarioPath, "scenario", "", "scenario file (scenario-v0.1) defining the run")
	fs.StringVar(&options.sinkName, "sink", "inproc", "sink: inproc|file|http-push")
	fs.StringVar(&options.sinkTarget, "sink-target", "", "sink target (file path)")
	registerRunFolder(fs, &options.outDir, "")
	fs.BoolVar(&options.explore, "explore", false, "explore mode: stamp the run unblinded from creation; it is never scored")
}

// registerDomainFlags registers where the domains live and which one to use.
func registerDomainFlags(fs *flag.FlagSet, dir, id *string) {
	registerDomainsDir(fs, dir, "domains")
	fs.StringVar(id, "domain", "", "domain id")
}

func admitRunOptions(options runOptions) (runOptions, error) {
	if options.scenarioPath != "" {
		return admitScenarioRun(options)
	}
	if options.domainID == "" {
		return runOptions{}, usagef("run requires --domain or --scenario")
	}
	if !(options.durationS > 0 && options.durationS*1e9 <= float64(run.MaxRunSpanNS)) {
		return runOptions{}, usagef("run: --duration must be positive and at most %d seconds (30 days), got %v", run.MaxRunSpanNS/1e9, options.durationS)
	}
	if options.startTime < 0 || options.startTime > run.MaxStartNS {
		return runOptions{}, usagef("run: --start-time must be in [0, %d], got %d", int64(run.MaxStartNS), options.startTime)
	}
	return admitRunSink(options)
}

// admitScenarioRun refuses the flags a scenario file defines.
func admitScenarioRun(options runOptions) (runOptions, error) {
	defined := scenarioDefinedFlags()
	var conflicting []string
	for _, name := range options.explicit {
		if defined[name] {
			conflicting = append(conflicting, "--"+name)
		}
	}
	if len(conflicting) > 0 {
		return runOptions{}, usagef("run: --scenario defines the run; remove %s", strings.Join(conflicting, ", "))
	}
	return admitRunSink(options)
}

// scenarioDefinedFlags names the flags a scenario file defines: exactly the
// ones registerScenarioFlags registers, so a new one cannot be missed
// (docs/design/SCENARIO_FILE.md §3).
func scenarioDefinedFlags() map[string]bool {
	fs := flag.NewFlagSet("scenario", flag.ContinueOnError)
	registerScenarioFlags(fs, &runOptions{})
	defined := map[string]bool{}
	fs.VisitAll(func(registered *flag.Flag) { defined[registered.Name] = true })
	return defined
}

// admitRunSink checks the sink choice and derives the file target from --out.
func admitRunSink(options runOptions) (runOptions, error) {
	switch {
	case options.sinkName != model.SinkInproc && options.sinkName != model.SinkFile && options.sinkName != model.SinkHTTPPush:
		return runOptions{}, usagef("run: --sink must be inproc, file or http-push, got %q", options.sinkName)
	case options.sinkName == model.SinkHTTPPush && options.sinkTarget == "":
		return runOptions{}, usagef("--sink=http-push requires --sink-target")
	case options.sinkName != model.SinkFile || options.sinkTarget != "":
		return options, nil
	case options.outDir == "":
		return runOptions{}, usagef("--sink=file requires --sink-target or --out")
	}
	options.sinkTarget = filepath.Join(options.outDir, "trace.jsonl")
	return options, nil
}
