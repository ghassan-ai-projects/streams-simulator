// Package app is the use-case layer of the suite module: it generates the
// suite a domain and profile describe and publishes it as a file.
package app

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/files"
)

// PublishOptions say which suite to publish: the domain (looked up by id in
// DomainsDir), its profile, the target scenario count and seed, and the
// directory the suite file is written to.
type PublishOptions struct {
	DomainsDir, DomainID, Profile, Out string
	N                                  int
	Seed                               uint64
}

// Published is the suite file that was written and what the suite holds.
type Published struct {
	Path            string
	Scenarios       int
	TrivialExcluded int
	TerminalState   string
}

// Publish generates the audited suite and writes it as
// <out>/<domain>-<profile>-suite.json. A profile the domain does not declare
// is refused with a *rules.UnknownProfileError before anything is generated.
func Publish(options PublishOptions) (Published, error) {
	spec, err := loadSpec(options)
	if err != nil {
		return Published{}, err
	}
	generated, err := rules.Generate(rules.Config{Domain: spec, Profile: options.Profile, N: options.N,
		Seed: options.Seed, SampleNS: rules.DefaultSampleNS})
	if err != nil {
		return Published{}, fmt.Errorf("streamsim: %w", err)
	}
	return write(generated, options)
}

// loadSpec is the domain with the requested id among those in the directory,
// checked to declare the profile.
func loadSpec(options PublishOptions) (*domain.Compiled, error) {
	cat, err := domain.NewCatalogFromDir(options.DomainsDir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	spec, err := cat.Describe(options.DomainID)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if !spec.HasProfile(options.Profile) {
		return nil, &rules.UnknownProfileError{Domain: options.DomainID, Profile: options.Profile}
	}
	return spec, nil
}

func write(generated *rules.Suite, options PublishOptions) (Published, error) {
	raw, _ := json.MarshalIndent(generated, "", "  ")
	path, err := files.Write(options.Out, rules.FileName(options.DomainID, options.Profile), raw)
	if err != nil {
		return Published{}, fmt.Errorf("streamsim: %w", err)
	}
	return Published{Path: path, Scenarios: len(generated.Scenarios),
		TrivialExcluded: len(generated.TrivialExcluded), TerminalState: generated.TerminalState}, nil
}
