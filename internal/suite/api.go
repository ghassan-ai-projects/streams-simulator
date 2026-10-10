package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
)

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
