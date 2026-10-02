package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

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
