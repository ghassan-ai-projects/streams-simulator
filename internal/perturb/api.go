package perturb

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/perturb/internal/domain"
)

// Delivered is one delivered record: the (possibly modified) event plus its
// ledger metadata. Drop yields Delivered=false; duplicate yields extra
// records with the same seq.
type Delivered = layer.Delivered

// ErrNoSpec is returned by New when no domain spec is given: the layer needs
// the contract (declared enums, ranges, units) to corrupt a stream against.
var ErrNoSpec = errors.New("perturb: a domain spec is required")
