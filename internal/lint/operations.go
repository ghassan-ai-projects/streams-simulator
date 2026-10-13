package lint

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/lint/internal/domain"
)

func Check(spec *domain.Compiled, options Options) (*Report, error) {
	return layer.Check(spec, options)
}
