package audit

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/audit/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

// Panel audits one injection: it runs the world twice (clean control and
// faulted), samples the delivered channel readings on a grid, and fits each
// trivial detector with hindsight on the labeled series.
type Panel struct {
	panel *layer.Panel
}

// NewPanel builds the audit panel. sampleNS <= 0 selects the default sample
// grid; a nil spec is refused with ErrNoSpec.
func NewPanel(spec *domain.Compiled, seed uint64, sampleNS int64) (*Panel, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Panel{panel: layer.NewPanel(spec, seed, sampleNS)}, nil
}
