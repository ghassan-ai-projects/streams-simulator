package suite

import layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"

// Generate builds a suite with the generate-audit-regenerate loop. The same
// configuration always yields the same suite bytes.
func Generate(cfg Config) (*Suite, error) {
	if cfg.Domain == nil {
		return nil, ErrNoDomain
	}
	return layer.Generate(cfg)
}
