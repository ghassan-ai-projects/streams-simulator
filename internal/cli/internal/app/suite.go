package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

func cmdSuite(s *session, args []string) (any, error) {
	options, err := parseSuiteOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	spec, err := loadSuiteDomain(options)
	if err != nil {
		return nil, err
	}
	generated, err := suite.Generate(suite.Config{Domain: spec, Profile: options.profile, N: options.n, Seed: options.seed, SampleNS: 120 * 1e9})
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return publishSuite(generated, options.out, spec.Spec.ID, options.profile)
}

// loadSuiteDomain loads the suite's domain and checks it has the profile.
func loadSuiteDomain(options suiteOptions) (*domain.Compiled, error) {
	spec, err := loadDomainSpec(options.domainsDir, options.domainID)
	if err != nil {
		return nil, err
	}
	if !spec.HasProfile(options.profile) {
		return nil, usagef("suite: domain %q has no profile %q", options.domainID, options.profile)
	}
	return spec, nil
}

type suiteOptions struct {
	domainsDir, domainID, profile, out string
	n                                  int
	seed                               uint64
}

func parseSuiteOptions(args []string, stderr io.Writer) (suiteOptions, error) {
	var options suiteOptions
	fs := newFlagSet("suite", stderr)
	registerSuiteFlags(fs, &options)
	if err := parseFlags(fs, args); err != nil {
		return suiteOptions{}, err
	}
	return admitSuiteOptions(options)
}

func registerSuiteFlags(fs *flag.FlagSet, options *suiteOptions) {
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.profile, "profile", "nominal", "profile name")
	fs.IntVar(&options.n, "n", 100, "target scenario count")
	fs.Uint64Var(&options.seed, "seed", 1, "suite seed")
	fs.StringVar(&options.out, "out", "suites", "output directory")
}

func admitSuiteOptions(options suiteOptions) (suiteOptions, error) {
	if options.domainID == "" {
		return suiteOptions{}, usagef("suite requires --domain")
	}
	if options.n < 1 {
		return suiteOptions{}, usagef("suite: --n must be at least 1, got %d", options.n)
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
