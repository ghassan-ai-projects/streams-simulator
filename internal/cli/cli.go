// Package cli implements the streamsim command-line interface: catalog,
// domain/adapter tooling, scripted runs, replay/verify, the MCP stdio
// servers, suite generation, and offline scoring. The entrypoint in
// cmd/streamsim is a thin dispatch over this package.
package cli

import (
	"fmt"
	"os"
)

// Main dispatches to a subcommand. It is the whole CLI surface.
// Version and Commit are injected by cmd/streamsim from the Makefile
// ldflags; the manifest command records them.
var (
	Version = "dev"
	Commit  = "none"
)

func Main(args []string) int {
	if len(args) < 2 {
		usage()
		return 2
	}
	cmd, rest := args[1], args[2:]
	var err error
	switch cmd {
	case "catalog":
		err = cmdCatalog(rest)
	case "domain":
		err = cmdDomain(rest)
	case "adapter":
		err = cmdAdapter(rest)
	case "run":
		err = cmdRun(rest)
	case "replay":
		err = cmdReplay(rest, false)
	case "verify":
		err = cmdReplay(rest, true)
	case "mcp":
		err = cmdMCP(rest)
	case "refconsumer":
		err = cmdRefconsumer(rest)
	case "suite":
		err = cmdSuite(rest)
	case "score":
		err = cmdScore(rest)
	case "manifest":
		err = cmdManifest(rest)
	case "device":
		err = cmdDevice(rest)
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "streamsim: unknown command %q\n\n", cmd)
		usage()
		return 2
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "streamsim: %v\n", err)
		return 1
	}
	return 0
}

func usage() {
	fmt.Fprintf(os.Stderr, `streamsim — deterministic world simulator for testing stream processors

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
`)
}
