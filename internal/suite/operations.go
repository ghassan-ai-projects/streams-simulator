package suite

import "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"

// Publish generates the audited suite a domain and profile describe and writes
// it as <out>/<domain>-<profile>-suite.json. A profile the domain does not
// declare is refused with a *UnknownProfileError before anything is generated.
func Publish(options PublishOptions) (Published, error) {
	return app.Publish(options)
}
