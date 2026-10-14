package app

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func cmdVerify(s *session, args []string) (any, error) {
	res, err := replayFile(s, "verify", args)
	if err != nil {
		return nil, err
	}
	if res.Matches {
		return replayResult(res), nil
	}
	message := fmt.Sprintf("verify: replay does not reproduce the artifact: %v", res.Detail)
	return nil, &outcomeError{code: codeReplayMismatch, message: message, result: replayResult(res)}
}

func cmdReplay(s *session, args []string) (any, error) {
	res, err := replayFile(s, "replay", args)
	if err != nil {
		return nil, err
	}
	return replayResult(res), nil
}

func replayFile(s *session, command string, args []string) (*run.ReplayResult, error) {
	fs, domainsDir, adaptersDir, err := parseReplayOptions(command, args, s.stderr)
	if err != nil {
		return nil, err
	}
	res, err := run.ReplayFile(context.Background(), fs.Arg(0), run.Directories{Domains: domainsDir, Adapters: adaptersDir})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return res, nil
}

func parseReplayOptions(command string, args []string, stderr io.Writer) (*flag.FlagSet, string, string, error) {
	fs := newFlagSet(command, stderr)
	var domains, adapters string
	registerInputDirs(fs, &domains, &adapters)
	if err := parseFlags(fs, args); err != nil {
		return nil, "", "", err
	}
	if fs.NArg() < 1 {
		return nil, "", "", usagef("%s requires a run artifact path", command)
	}
	if err := refuseExtraArguments(fs, 1); err != nil {
		return nil, "", "", err
	}
	return fs, domains, adapters, nil
}

func replayResult(res *run.ReplayResult) map[string]any {
	doc := map[string]any{"matches": res.Matches, "version_match": res.VersionMatch,
		"got": res.GotDigest, "want": res.WantDigest, "first_divergence": res.FirstDivergence, "detail": res.Detail, "reason": res.Reason}
	if res.File != "" {
		doc["file"] = res.File
	}
	if res.DomainChanged {
		doc["domain_changed"], doc["current_domain_digest"] = true, res.CurrentDomainDigest
	}
	return doc
}
