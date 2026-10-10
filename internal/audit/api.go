package audit

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/audit/internal/domain"
)

// Verdict is the audit outcome for one scenario.
type Verdict = layer.Verdict

// Perturbation is one delivery perturbation the audit applies to the
// delivered stream, exactly as the scenario declares it.
type Perturbation = layer.Perturbation

// ErrNoSpec is returned by NewPanel when no domain spec is given.
var ErrNoSpec = errors.New("audit: a domain spec is required")

// ErrNoPanel is returned by Audit on a nil or zero Panel.
var ErrNoPanel = errors.New("audit: a panel built by NewPanel is required")
