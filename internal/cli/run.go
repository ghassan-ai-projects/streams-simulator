package cli

import (
	"context"
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"path/filepath"
	"strings"
)

func cmdRun(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	domainID := fs.String("domain", "", "domain id")
	adapterID := fs.String("adapter", "native-jsonl", "adapter id")
	seed := fs.Uint64("seed", 1, "world seed")
	sinkName := fs.String("sink", "inproc", "sink: inproc|file|http-push")
	sinkTarget := fs.String("sink-target", "", "sink target (file path)")
	outDir := fs.String("out", "", "output directory for artifacts")
	durationS := fs.Float64("duration", 6*3600, "run duration (seconds)")
	startTime := fs.Int64("start-time", model.DefaultStartTimeNS, "world start (ns epoch)")
	faults := fs.String("fault", "", "repeatable: entity=fault@offset_s")
	perts := fs.String("perturb", "", "repeatable: name@from_s[@until_s]")
	effectors := fs.String("effector", "", "repeatable: effector@entity@offset_s")
	profile := fs.String("profile", "", "scenario profile")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *domainID == "" {
		return fmt.Errorf("run requires --domain")
	}
	if *sinkName == model.SinkFile && *sinkTarget == "" {
		if *outDir == "" {
			return fmt.Errorf("--sink=file requires --sink-target or --out")
		}
		*sinkTarget = filepath.Join(*outDir, "trace.jsonl")
	}
	cat, err := loadCatalog(*domainsDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(*adaptersDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(*domainID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[*adapterID]
	if !ok {
		return fmt.Errorf("unknown adapter %q", *adapterID)
	}
	cfg := run.Config{
		Domain: spec, Adapter: adap, Seed: *seed, SinkName: *sinkName,
		SinkTarget: *sinkTarget, TimeMode: model.TimeStepped, StartTimeNS: *startTime,
		StartTimeSet:    true,
		ScenarioProfile: *profile,
	}
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	start := *startTime
	// Setup first, then faults/perturbations at their offsets, then run.
	if *faults != "" {
		for _, f := range splitCSV(*faults) {
			entity, fault, offsetS, err := parseTriple(f, "=", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			if _, err := r.InjectFault(entity, fault, start+int64(offsetS*1e9), nil); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	if *perts != "" {
		for _, p := range splitCSV(*perts) {
			parts := strings.SplitN(p, "@", 3)
			name := parts[0]
			fromS := 0.0
			untilS := 0.0
			if len(parts) > 1 {
				fromS = parseF(parts[1])
			}
			if len(parts) > 2 {
				untilS = parseF(parts[2])
			}
			from := start + int64(fromS*1e9)
			until := int64(0)
			if untilS > 0 {
				until = start + int64(untilS*1e9)
			}
			if _, err := r.ApplyPerturb(name, nil, from, until); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	if *effectors != "" {
		for _, e := range splitCSV(*effectors) {
			eff, entity, offsetS, err := parseTriple(e, "@", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			cmdID := fmt.Sprintf("cli-%d", timeNanos())
			if _, err := r.InvokeEffector(eff, entity, cmdID, map[string]any{}, start+int64(offsetS*1e9)); err != nil {
				return fmt.Errorf("effector %s: %w", eff, err)
			}
		}
	}
	if _, err := r.Advance(context.Background(), start+int64(*durationS*1e9), false); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	art, err := r.End(*outDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"run_id": art.RunID, "trace_digest": art.ExpectedTraceDigest,
		"reproducible": art.Reproducible, "emitted": art.Counts.Emitted,
		"ledger": len(r.Ledger()),
	})
}

func cmdReplay(args []string, verifyOnly bool) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	domainsDir := fs.String("domains-dir", "domains", "domain specs directory")
	adaptersDir := fs.String("adapters-dir", "adapters", "adapters directory")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("replay requires a run artifact path")
	}
	art, err := run.LoadArtifact(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	cat, err := loadCatalog(*domainsDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adapters, err := loadAdapters(*adaptersDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(art.Domain.ID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	adap, ok := adapters[art.Adapter.ID]
	if !ok {
		return fmt.Errorf("unknown adapter %q", art.Adapter.ID)
	}
	res, err := run.ReplayArtifact(context.Background(), art, spec, adap, "")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest,
		"first_divergence": res.FirstDivergence, "detail": res.Detail,
	})
}
