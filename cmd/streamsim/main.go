// Command streamsim is the Streams Simulator CLI. See internal/cli for the
// subcommand implementations.
package main

import (
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli"
)

// Version and Commit are set via -ldflags in the Makefile.
var (
	Version = "dev"
	Commit  = "none"
)

func main() {
	cli.Version = Version
	cli.Commit = Commit
	os.Exit(cli.Main(os.Args))
}
