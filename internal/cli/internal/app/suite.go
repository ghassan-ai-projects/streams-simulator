package app

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

func cmdSuite(s *session, args []string) (any, error) {
	options, err := parseSuiteOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	spec, err := loadDomainSpec(options.domainsDir, options.domainID)
	if err != nil {
		return nil, err
	}
	generated, err := suite.Generate(suite.Config{Domain: spec, Profile: options.profile, N: options.n, Seed: options.seed, SampleNS: 120 * 1e9})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return publishSuite(generated, options.out, spec.Spec.ID, options.profile)
}

type suiteOptions struct {
	domainsDir, domainID, profile, out string
	n                                  int
	seed                               uint64
}

func parseSuiteOptions(args []string, stderr io.Writer) (suiteOptions, error) {
	var options suiteOptions
	fs := newFlagSet("suite", stderr)
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.profile, "profile", "nominal", "profile name")
	fs.IntVar(&options.n, "n", 100, "target scenario count")
	fs.Uint64Var(&options.seed, "seed", 1, "suite seed")
	fs.StringVar(&options.out, "out", "suites", "output directory")
	if err := parseFlags(fs, args); err != nil {
		return suiteOptions{}, err
	}
	return options, nil
}

func publishSuite(s *suite.Suite, out, domain, profile string) (any, error) {
	if err := files.EnsureDir(out); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	raw, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(out, fmt.Sprintf("%s-%s-suite.json", domain, profile))
	if err := files.Write(path, raw); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"suite": path, "scenarios": len(s.Scenarios),
		"trivial_excluded": len(s.TrivialExcluded), "terminal_state": s.TerminalState}, nil
}
