// Package cli implements the streamsim command-line interface: catalog,
// domain/adapter tooling, scripted runs, replay/verify, the MCP stdio
// servers, suite generation, and offline scoring. The entrypoint in
// cmd/streamsim is a thin dispatch over this package.
package cli

import (
	"fmt"
	"os"
)

// Version and Commit are injected by the entrypoint from Makefile ldflags.
// The manifest command records them.
var (
	Version = "dev"
	Commit  = "none"
)

// Main dispatches the command and returns its process exit status.
func Main(args []string) int {
	if len(args) < 2 {
		usage()
		return 2
	}
	if isHelpCommand(args[1]) {
		usage()
		return 0
	}
	handler, known := commandHandlers[args[1]]
	if !known {
		return unknownCommand(args[1])
	}
	return executeCommand(handler, args[2:])
}

var commandHandlers = map[string]func([]string) error{
	"catalog":     cmdCatalog,
	"domain":      cmdDomain,
	"adapter":     cmdAdapter,
	"run":         cmdRun,
	"replay":      func(args []string) error { return cmdReplay(args, false) },
	"verify":      func(args []string) error { return cmdReplay(args, true) },
	"mcp":         cmdMCP,
	"refconsumer": cmdRefconsumer,
	"suite":       cmdSuite,
	"score":       cmdScore,
	"manifest":    cmdManifest,
	"device":      cmdDevice,
}

func isHelpCommand(command string) bool {
	return command == "help" || command == "-h" || command == "--help"
}

func unknownCommand(command string) int {
	fmt.Fprintf(os.Stderr, "streamsim: unknown command %q\n\n", command)
	usage()
	return 2
}

func executeCommand(handler func([]string) error, args []string) int {
	if err := handler(args); err != nil {
		fmt.Fprintf(os.Stderr, "streamsim: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Fprint(os.Stderr, usageText)
}

const usageText = `streamsim — deterministic world simulator for testing stream processors

Usage: streamsim <command> [flags]

Commands:
  catalog list|describe|coverage        Installed domains and axis coverage
  domain validate <path>                Validate a domain spec against the contract
  adapter list|verify [path]            Installed adapters; verify one against its golden
  run                                   One-shot scripted run
  replay <run.json>                     Reproduce a run from its artifact
  verify <run.json>                     Verify a run artifact reproduces
  mcp --role director                   Serve the director role over stdio
                                        (--operator-addr binds the operator role
                                        over streamable HTTP; sim.world.create
                                        returns the endpoint with the token)
  refconsumer --trace <file>            Detect episodes in a trace; write verdict.json
                                        (--mcp <url> --token <t> --run <id> closes
                                        the loop over the operator endpoint)
  suite --domain --profile              Generate an audited graded suite
  score --run --label                   Offline scorecard from artifacts
  manifest                              Write release-manifest.json (author +
                                        reviewer identity, optional ed25519)
  device serve --socket <path>          Serve the device emulator over a UDS
                                        (state/receipt link; --world enables plant wiring)
  help
`
