package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
)

// Generate builds a suite with the generate-audit-regenerate loop. The same
// configuration always yields the same suite bytes.
func Generate(cfg Config) (*Suite, error) {
	if cfg.Domain == nil {
		return nil, ErrNoDomain
	}
	return layer.Generate(cfg)
}

// Publish generates the audited suite a domain and profile describe and writes
// it as <out>/<domain>-<profile>-suite.json. A profile the domain does not
// declare is refused with a *UnknownProfileError before anything is generated.
func Publish(options PublishOptions) (Published, error) {
	return app.Publish(options)
}
