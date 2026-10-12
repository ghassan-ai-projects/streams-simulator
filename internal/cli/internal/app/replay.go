package app

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// cmdVerify replays the artifact like cmdReplay but fails the command when
// the replay does not reproduce the recorded trace, so scripts can gate on
// the exit status. The result document is still printed.
func cmdVerify(s *session, args []string) (any, error) {
	res, err := replayFile(s, args)
	if err != nil {
		return nil, err
	}
	if res.Matches {
		return replayResult(res), nil
	}
	message := fmt.Sprintf("verify: replay does not reproduce the artifact: %v", res.Detail)
	return nil, &outcomeError{code: codeReplayMismatch, message: message, result: replayResult(res)}
}

// cmdReplay reproduces the run and reports whether the trace digest matches;
// it exits 0 whatever the outcome (see cmdVerify).
func cmdReplay(s *session, args []string) (any, error) {
	res, err := replayFile(s, args)
	if err != nil {
		return nil, err
	}
	return replayResult(res), nil
}

func replayFile(s *session, args []string) (*run.ReplayResult, error) {
	fs, domainsDir, adaptersDir, err := parseReplayOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	res, err := run.ReplayFile(context.Background(), fs.Arg(0), run.Directories{Domains: domainsDir, Adapters: adaptersDir})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return res, nil
}

func parseReplayOptions(args []string, stderr io.Writer) (*flag.FlagSet, string, string, error) {
	fs := newFlagSet("replay", stderr)
	domains := fs.String("domains-dir", "domains", "domain specs directory")
	adapters := fs.String("adapters-dir", "adapters", "adapters directory")
	if err := parseFlags(fs, args); err != nil {
		return nil, "", "", err
	}
	if fs.NArg() < 1 {
		return nil, "", "", usagef("replay requires a run artifact path")
	}
	if err := refuseExtraArguments(fs, 1); err != nil {
		return nil, "", "", err
	}
	return fs, *domains, *adapters, nil
}

func replayResult(res *run.ReplayResult) map[string]any {
	return map[string]any{"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest, "first_divergence": res.FirstDivergence, "detail": res.Detail}
}
