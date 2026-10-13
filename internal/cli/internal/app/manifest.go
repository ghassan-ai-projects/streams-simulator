package app

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/release"
)

func cmdManifest(s *session, args []string) (any, error) {
	options, err := parseManifestOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	result, err := release.Publish(release.Options{Out: options.out, DomainsDir: options.domainsDir,
		AdaptersDir: options.adaptersDir, Author: options.author, Reviewer: options.reviewer, KeyFile: options.keyFile,
		Version: s.build.Version, Commit: s.build.Commit, Now: s.now})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return map[string]any{"manifest_written": result.Path, "signed": result.Signed, "sim": result.Sim}, nil
}
