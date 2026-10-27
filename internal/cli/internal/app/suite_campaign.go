package app

import (
	"errors"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

type campaignOptions struct {
	domainsDir, spec, out string
}

func cmdSuiteCampaign(s *session, args []string) (any, error) {
	options, err := parseCampaignOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	published, err := suite.PublishCampaign(suite.CampaignOptions{DomainsDir: options.domainsDir, SpecPath: options.spec, Out: options.out})
	var invalid *suite.CampaignSpecError
	if errors.As(err, &invalid) {
		return nil, usagef("%s", invalid.Error())
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"campaign": published.Path, "runs": published.Runs}, nil
}

func parseCampaignOptions(args []string, stderr io.Writer) (campaignOptions, error) {
	var options campaignOptions
	fs := newFlagSet("suite campaign", stderr)
	registerDomainsDir(fs, &options.domainsDir, "domains")
	fs.StringVar(&options.spec, "spec", "", "campaign spec file")
	fs.StringVar(&options.out, "out", "campaign", "output directory")
	if err := parseOnlyFlags(fs, args); err != nil {
		return campaignOptions{}, err
	}
	if options.spec == "" {
		return campaignOptions{}, usagef("suite campaign requires --spec")
	}
	return options, nil
}
