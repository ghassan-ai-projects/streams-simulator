package release

import "github.com/ghassan-ai-projects/streams-simulator/internal/release/internal/app"

func Publish(options Options) (Result, error) {
	return app.Publish(options)
}
