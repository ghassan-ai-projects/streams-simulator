package app

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

type runOptions struct {
	domainsDir, adaptersDir, domainID, adapterID                    string
	sinkName, sinkTarget, outDir, faults, perts, effectors, profile string
	seed                                                            uint64
	durationS                                                       float64
	startTime                                                       int64
}

func parseRunOptions(args []string, stderr io.Writer) (runOptions, error) {
	var options runOptions
	fs := newFlagSet("run", stderr)
	registerRunInputs(fs, &options)
	registerRunExecution(fs, &options)
	registerRunScript(fs, &options)
	if err := parseFlags(fs, args); err != nil {
		return runOptions{}, err
	}
	return admitRunOptions(options)
}

func loadRunConfig(options runOptions) (run.Config, error) {
	spec, adap, err := loadSimulatorInputs(options.domainsDir, options.adaptersDir, options.domainID, options.adapterID)
	if err != nil {
		return run.Config{}, err
	}
	return run.Config{Domain: spec, Adapter: adap, Seed: options.seed, SinkName: options.sinkName,
		SinkTarget: options.sinkTarget, TimeMode: model.TimeStepped, StartTimeNS: options.startTime,
		StartTimeSet: true, ScenarioProfile: options.profile}, nil
}

func registerRunInputs(fs *flag.FlagSet, options *runOptions) {
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.adapterID, "adapter", "native-jsonl", "adapter id")
	fs.Uint64Var(&options.seed, "seed", 1, "world seed")
}

func registerRunExecution(fs *flag.FlagSet, options *runOptions) {
	fs.StringVar(&options.sinkName, "sink", "inproc", "sink: inproc|file|http-push")
	fs.StringVar(&options.sinkTarget, "sink-target", "", "sink target (file path)")
	fs.StringVar(&options.outDir, "out", "", "output directory for artifacts")
	fs.Float64Var(&options.durationS, "duration", 6*3600, "run duration (seconds)")
	fs.Int64Var(&options.startTime, "start-time", model.DefaultStartTimeNS, "world start (ns epoch)")
}

func registerRunScript(fs *flag.FlagSet, options *runOptions) {
	fs.StringVar(&options.faults, "fault", "", "repeatable: entity=fault@offset_s")
	fs.StringVar(&options.perts, "perturb", "", "repeatable: name@from_s[@until_s]")
	fs.StringVar(&options.effectors, "effector", "", "repeatable: effector@entity@offset_s")
	fs.StringVar(&options.profile, "profile", "", "scenario profile")
}

func admitRunOptions(options runOptions) (runOptions, error) {
	if options.domainID == "" {
		return runOptions{}, fmt.Errorf("run requires --domain")
	}
	if options.sinkName == model.SinkFile && options.sinkTarget == "" {
		if options.outDir == "" {
			return runOptions{}, fmt.Errorf("--sink=file requires --sink-target or --out")
		}
		options.sinkTarget = filepath.Join(options.outDir, "trace.jsonl")
	}
	return options, nil
}
