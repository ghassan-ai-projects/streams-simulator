package app

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

func cmdSuite(s *session, args []string) (any, error) {
	if len(args) > 0 && args[0] == "campaign" {
		return cmdSuiteCampaign(s, args[1:])
	}
	options, err := parseSuiteOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	published, err := publishSuite(options)
	if err != nil {
		return nil, err
	}
	return map[string]any{"suite": published.Path, "scenarios": published.Scenarios,
		"trivial_excluded": published.TrivialExcluded, "terminal_state": published.TerminalState}, nil
}

func publishSuite(options suiteOptions) (suite.Published, error) {
	published, err := suite.Publish(suite.PublishOptions{DomainsDir: options.domainsDir, DomainID: options.domainID,
		Profile: options.profile, Out: options.out, N: options.n, Seed: options.seed})
	var unknown *suite.UnknownProfileError
	if errors.As(err, &unknown) {
		return suite.Published{}, usagef("%s", unknown.Error())
	}
	if err != nil {
		return suite.Published{}, fmt.Errorf("%w", err)
	}
	return published, nil
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
	if err := parseOnlyFlags(fs, args); err != nil {
		return suiteOptions{}, err
	}
	return admitSuiteOptions(options)
}

func registerSuiteFlags(fs *flag.FlagSet, options *suiteOptions) {
	registerDomainFlags(fs, &options.domainsDir, &options.domainID)
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
