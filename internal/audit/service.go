package audit

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/audit/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

type Panel struct {
	panel *layer.Panel
}

func NewPanel(spec *domain.Compiled, seed uint64, sampleNS int64) (*Panel, error) {
	if spec == nil {
		return nil, ErrNoSpec
	}
	return &Panel{panel: layer.NewPanel(spec, seed, sampleNS)}, nil
}
