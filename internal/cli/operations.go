package cli

import "github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/app"

func Main(args []string, build Build) int {
	return app.Main(args, build.Version, build.Commit)
}
