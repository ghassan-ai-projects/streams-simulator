package suite

import (
	"errors"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
)

// ErrNoDomain is returned by Generate when the configuration has no domain.
var ErrNoDomain = errors.New("suite: a domain is required")

// Perturbation is one delivery perturbation in a scenario: the same record
// the audit applies, so a scenario is audited exactly as declared.
type Perturbation = audit.Perturbation

// Scenario is one executable scenario recipe.
type Scenario = layer.Scenario

// TrivialCase is a scenario the audit rejected, kept as a mechanism
// regression fixture.
type TrivialCase = layer.TrivialCase

// Suite is a generated graded suite.
type Suite = layer.Suite

// Config pins suite generation.
type Config = layer.Config

// PublishOptions say which suite to publish: the domain (looked up by id in
// DomainsDir), its profile, the target scenario count and seed, and the
// directory the suite file is written to.
type PublishOptions = app.PublishOptions

// Published is the suite file that was written and what the suite holds.
type Published = app.Published

// UnknownProfileError refuses a suite for a profile its domain does not
// declare.
type UnknownProfileError = layer.UnknownProfileError
