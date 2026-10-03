package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

func cmdSuite(args []string) error {
	options, err := parseSuiteOptions(args)
	if err != nil {
		return err
	}
	spec, err := loadDomainSpec(options.domainsDir, options.domainID)
	if err != nil {
		return err
	}
	s, err := suite.Generate(suite.Config{Domain: spec, Profile: options.profile, N: options.n, Seed: options.seed, SampleNS: 120 * 1e9})
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return publishSuite(s, options.out, spec.Spec.ID, options.profile)
}

type suiteOptions struct {
	domainsDir, domainID, profile, out string
	n                                  int
	seed                               uint64
}

func parseSuiteOptions(args []string) (suiteOptions, error) {
	var options suiteOptions
	fs := flag.NewFlagSet("suite", flag.ExitOnError)
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory")
	fs.StringVar(&options.domainID, "domain", "", "domain id")
	fs.StringVar(&options.profile, "profile", "nominal", "profile name")
	fs.IntVar(&options.n, "n", 100, "target scenario count")
	fs.Uint64Var(&options.seed, "seed", 1, "suite seed")
	fs.StringVar(&options.out, "out", "suites", "output directory")
	if err := fs.Parse(args); err != nil {
		return suiteOptions{}, fmt.Errorf("streamsim: %w", err)
	}
	return options, nil
}

func publishSuite(s *suite.Suite, out, domain, profile string) error {
	if err := os.MkdirAll(out, 0o700); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	raw, _ := json.MarshalIndent(s, "", "  ")
	path := filepath.Join(out, fmt.Sprintf("%s-%s-suite.json", domain, profile))
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"suite": path, "scenarios": len(s.Scenarios),
		"trivial_excluded": len(s.TrivialExcluded), "terminal_state": s.TerminalState})
}
