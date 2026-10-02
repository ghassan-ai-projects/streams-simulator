package cli

import (
	"flag"
	"fmt"
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

func parseRunOptions(args []string) (runOptions, error) {
	var options runOptions
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.adapterID, "adapter", "native-jsonl", "adapter id")
	fs.Uint64Var(&options.seed, "seed", 1, "world seed")
	fs.StringVar(&options.sinkName, "sink", "inproc", "sink: inproc|file|http-push")
	fs.StringVar(&options.sinkTarget, "sink-target", "", "sink target (file path)")
	fs.StringVar(&options.outDir, "out", "", "output directory for artifacts")
	fs.Float64Var(&options.durationS, "duration", 6*3600, "run duration (seconds)")
	fs.Int64Var(&options.startTime, "start-time", model.DefaultStartTimeNS, "world start (ns epoch)")
	fs.StringVar(&options.faults, "fault", "", "repeatable: entity=fault@offset_s")
	fs.StringVar(&options.perts, "perturb", "", "repeatable: name@from_s[@until_s]")
	fs.StringVar(&options.effectors, "effector", "", "repeatable: effector@entity@offset_s")
	fs.StringVar(&options.profile, "profile", "", "scenario profile")
	if err := fs.Parse(args); err != nil {
		return runOptions{}, fmt.Errorf("streamsim: %w", err)
	}
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

func loadRunConfig(options runOptions) (run.Config, error) {
	cat, err := loadCatalog(options.domainsDir)
	if err != nil {
		return run.Config{}, fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(options.adaptersDir)
	if err != nil {
		return run.Config{}, fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(options.domainID)
	if err != nil {
		return run.Config{}, fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[options.adapterID]
	if !ok {
		return run.Config{}, fmt.Errorf("unknown adapter %q", options.adapterID)
	}
	cfg := run.Config{
		Domain: spec, Adapter: adap, Seed: options.seed, SinkName: options.sinkName,
		SinkTarget: options.sinkTarget, TimeMode: model.TimeStepped, StartTimeNS: options.startTime,
		StartTimeSet:    true,
		ScenarioProfile: options.profile,
	}
	return cfg, nil
}
