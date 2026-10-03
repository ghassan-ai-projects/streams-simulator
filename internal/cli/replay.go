package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func cmdReplay(args []string, verifyOnly bool) error {
	fs, domainsDir, adaptersDir, err := parseReplayOptions(args)
	if err != nil {
		return err
	}
	art, err := run.LoadArtifact(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return executeReplay(art, domainsDir, adaptersDir)
}

func parseReplayOptions(args []string) (*flag.FlagSet, string, string, error) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	domains := fs.String("domains-dir", "domains", "domain specs directory")
	adapters := fs.String("adapters-dir", "adapters", "adapters directory")
	if err := fs.Parse(args); err != nil {
		return nil, "", "", fmt.Errorf("streamsim: %w", err)
	}
	if fs.NArg() < 1 {
		return nil, "", "", fmt.Errorf("replay requires a run artifact path")
	}
	return fs, *domains, *adapters, nil
}

func printReplayResult(res *run.ReplayResult) error {
	return printJSON(map[string]any{"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest, "first_divergence": res.FirstDivergence, "detail": res.Detail})
}

func executeReplay(art *model.RunArtifact, domainsDir, adaptersDir string) error {
	spec, adap, err := loadSimulatorInputs(domainsDir, adaptersDir, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return err
	}
	res, err := run.ReplayArtifact(context.Background(), art, spec, adap, "")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printReplayResult(res)
}
