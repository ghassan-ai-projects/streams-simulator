package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

func generationDefaults(cfg Config) Config {
	if cfg.N <= 0 {
		cfg.N = 100
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3 * cfg.N
	}
	return cfg
}

func prepareGeneration(cfg Config, prof *model.Profile) *suiteGeneration {
	rng := randutil.NewSplitMix64(cfg.Seed ^ randutil.Fnv1a64(cfg.Domain.Spec.ID+"/"+cfg.Profile))
	fraction := cfg.Domain.Spec.GroundTruth.NegativeClassFraction
	if fraction <= 0 {
		fraction = 0.4
	}
	solver := truth.NewSolver(cfg.Domain, cfg.Seed, cfg.SampleNS, 72*3600*1e9)
	panel := audit.NewPanel(cfg.Domain, cfg.Seed^0x5eed, cfg.SampleNS)
	suite := &Suite{DomainID: cfg.Domain.Spec.ID, Profile: cfg.Profile, Seed: cfg.Seed}
	return &suiteGeneration{cfg: cfg, prof: prof, rng: rng, solver: solver, panel: panel, suite: suite,
		perturbCount: map[string]int{}, negativeFrac: fraction}
}

func profileDuration(prof *model.Profile) int64 {
	duration := int64(prof.DurationS * 1e9)
	if duration <= 0 {
		duration = 72 * 3600 * 1e9
	}
	return duration
}
