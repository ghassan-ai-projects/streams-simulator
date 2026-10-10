// Package app implements the streamsim command-line use cases: catalog,
// domain/adapter tooling, scripted runs, replay/verify, the MCP servers,
// suite generation, offline scoring, the release manifest and the device
// emulator. It reaches the process only through process.Env and the files and
// serve edges.
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/process"
)

// Build is the identity of the running binary; the manifest records it.
type Build struct{ Version, Commit string }

// session is what one command invocation reaches of its process.
type session struct {
	build  Build
	stderr io.Writer
	now    func() time.Time
}

// handler runs one command and returns the JSON document to print, or nil
// when the command prints nothing.
type handler func(s *session, args []string) (any, error)

var commandHandlers = map[string]handler{
	"catalog":     cmdCatalog,
	"domain":      cmdDomain,
	"adapter":     cmdAdapter,
	"run":         cmdRun,
	"replay":      cmdReplay,
	"verify":      cmdVerify,
	"mcp":         cmdMCP,
	"refconsumer": cmdRefconsumer,
	"suite":       cmdSuite,
	"score":       cmdScore,
	"manifest":    cmdManifest,
	"device":      cmdDevice,
}

// Main runs the command in the current process and returns its exit status.
func Main(args []string, version, commit string) int {
	return Run(args, Build{Version: version, Commit: commit}, process.Standard())
}

// Run dispatches the command and returns its process exit status.
func Run(args []string, build Build, env process.Env) int {
	s := &session{build: build, stderr: env.Stderr, now: env.Now}
	if len(args) < 2 {
		s.usage()
		return 2
	}
	if isHelpCommand(args[1]) {
		s.usage()
		return 0
	}
	return s.dispatch(env.Stdout, args[1], args[2:])
}

func (s *session) dispatch(stdout io.Writer, name string, args []string) int {
	command, known := commandHandlers[name]
	if !known {
		return s.unknownCommand(name)
	}
	return s.execute(stdout, command, args)
}

func isHelpCommand(command string) bool {
	return command == "help" || command == "-h" || command == "--help"
}

func (s *session) unknownCommand(command string) int {
	_, _ = fmt.Fprintf(s.stderr, "streamsim: unknown command %q\n\n", command)
	s.usage()
	return 2
}

// execute runs the handler and prints its result as indented JSON on stdout.
func (s *session) execute(stdout io.Writer, command handler, args []string) int {
	result, err := command(s, args)
	if result != nil {
		if printErr := printJSON(stdout, result); err == nil {
			err = printErr
		}
	}
	if err == nil {
		return 0
	}
	if exit, ok := asUsageExit(err); ok {
		return exit.code
	}
	_, _ = fmt.Fprintf(s.stderr, "streamsim: %v\n", err)
	return 1
}

func printJSON(stdout io.Writer, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	_, _ = fmt.Fprintln(stdout, string(raw))
	return nil
}

func (s *session) usage() {
	_, _ = fmt.Fprint(s.stderr, usageText)
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
