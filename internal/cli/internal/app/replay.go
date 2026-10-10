package app

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// cmdVerify replays the artifact like cmdReplay but fails the command when
// the replay does not reproduce the recorded trace, so scripts can gate on
// the exit status. The result document is still printed.
func cmdVerify(s *session, args []string) (any, error) {
	result, err := cmdReplay(s, args)
	if err != nil {
		return nil, err
	}
	report, _ := result.(map[string]any)
	if report["matches"] == true {
		return result, nil
	}
	return result, fmt.Errorf("verify: replay does not reproduce the artifact: %v", report["detail"])
}

// cmdReplay reproduces the run and reports whether the trace digest matches;
// it exits 0 whatever the outcome (see cmdVerify).
func cmdReplay(s *session, args []string) (any, error) {
	fs, domainsDir, adaptersDir, err := parseReplayOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	art, err := run.LoadArtifact(fs.Arg(0))
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return executeReplay(art, domainsDir, adaptersDir)
}

func parseReplayOptions(args []string, stderr io.Writer) (*flag.FlagSet, string, string, error) {
	fs := newFlagSet("replay", stderr)
	domains := fs.String("domains-dir", "domains", "domain specs directory")
	adapters := fs.String("adapters-dir", "adapters", "adapters directory")
	if err := parseFlags(fs, args); err != nil {
		return nil, "", "", err
	}
	if fs.NArg() < 1 {
		return nil, "", "", fmt.Errorf("replay requires a run artifact path")
	}
	return fs, *domains, *adapters, nil
}

func replayResult(res *run.ReplayResult) map[string]any {
	return map[string]any{"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest, "first_divergence": res.FirstDivergence, "detail": res.Detail}
}

func executeReplay(art *model.RunArtifact, domainsDir, adaptersDir string) (any, error) {
	spec, adap, err := loadSimulatorInputs(domainsDir, adaptersDir, art.Domain.ID, art.Adapter.ID)
	if err != nil {
		return nil, err
	}
	res, err := run.ReplayArtifact(context.Background(), art, spec, adap, "")
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return replayResult(res), nil
}
