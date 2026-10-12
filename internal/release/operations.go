package release

import "github.com/ghassan-ai-projects/streams-simulator/internal/release/internal/app"

// Publish digests every domain and adapter file, builds the manifest, signs
// its canonical body when a key is given and writes the manifest document.
func Publish(options Options) (Result, error) {
	return app.Publish(options)
}
