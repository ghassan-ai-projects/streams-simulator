package cli

import "github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/app"

// Main runs the command named by args (args[0] is the program name) in the
// current process and returns its exit status.
func Main(args []string, build Build) int {
	return app.Main(args, build.Version, build.Commit)
}
