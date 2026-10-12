package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
)

// Perturbation is one delivery perturbation in a scenario: the same record
// the audit applies, so a scenario is audited exactly as declared.
type Perturbation = audit.Perturbation

// PublishOptions say which suite to publish: the domain (looked up by id in
// DomainsDir), its profile, the target scenario count and seed, and the
// directory the suite file is written to.
type PublishOptions = app.PublishOptions

// Published is the suite file that was written and what the suite holds.
type Published = app.Published

// UnknownProfileError refuses a suite for a profile its domain does not
// declare.
type UnknownProfileError = layer.UnknownProfileError
