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

type Build struct{ Version, Commit string }

type session struct {
	build  Build
	stderr io.Writer
	now    func() time.Time
	getenv func(string) string
}

type handler func(s *session, args []string) (any, error)

type command struct {
	run        handler
	ownsStdout bool
}

var commands = map[string]command{
	"catalog":     {run: cmdCatalog},
	"domain":      {run: cmdDomain},
	"component":   {run: cmdComponent},
	"adapter":     {run: cmdAdapter},
	"run":         {run: cmdRun},
	"replay":      {run: cmdReplay},
	"verify":      {run: cmdVerify},
	"mcp":         {run: cmdMCP, ownsStdout: true},
	"canvas":      {run: cmdCanvas, ownsStdout: true},
	"refconsumer": {run: cmdRefconsumer},
	"suite":       {run: cmdSuite},
	"score":       {run: cmdScore},
	"report":      {run: cmdReport},
	"manifest":    {run: cmdManifest},
	"device":      {run: cmdDevice},
	"scenario":    {run: cmdScenario},
	"plantwire":   {run: cmdPlantWire},
}

func Main(args []string, version, commit string) int {
	return Run(args, Build{Version: version, Commit: commit}, process.Standard())
}

func Run(args []string, build Build, env process.Env) int {
	s := &session{build: build, stderr: env.Stderr, now: env.Now, getenv: env.Getenv}
	if len(args) < 2 {
		s.usage()
		return s.fail(env.Stdout, "", usageExit{code: 2, message: "a command is required"})
	}
	if isHelpCommand(args[1]) {
		s.usage()
		return 0
	}
	return s.dispatch(env.Stdout, args[1], args[2:])
}

func (s *session) dispatch(stdout io.Writer, name string, args []string) int {
	registered, known := commands[name]
	if !known {
		return s.unknownCommand(stdout, name)
	}
	if registered.ownsStdout {
		return s.execute(stdout, io.Discard, name, registered.run, args)
	}
	return s.execute(stdout, stdout, name, registered.run, args)
}

func isHelpCommand(command string) bool {
	return command == "help" || command == "-h" || command == "--help"
}

func (s *session) unknownCommand(stdout io.Writer, command string) int {
	_, _ = fmt.Fprintf(s.stderr, "streamsim: unknown command %q\n\n", command)
	s.usage()
	return s.fail(stdout, "", usageExit{code: 2, message: fmt.Sprintf("unknown command %q", command)})
}

func (s *session) execute(stdout, documents io.Writer, name string, run handler, args []string) int {
	result, err := run(s, args)
	if err != nil {
		return s.fail(documents, name, err)
	}
	if result == nil {
		return 0
	}
	if err := printJSON(stdout, result); err != nil {
		return s.fail(documents, name, err)
	}
	return 0
}

func (s *session) fail(documents io.Writer, name string, err error) int {
	if exit, ok := asUsageExit(err); ok && exit.code == 0 {
		return 0
	}
	if _, flagReported := asUsageExit(err); !flagReported {
		_, _ = fmt.Fprintf(s.stderr, "streamsim: %v\n", err)
	}
	_ = printJSON(documents, describeFailure(name, err))
	return exitStatusOf(err)
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
  domain format <path>                  Rewrite a domain spec in canonical formatting
  domain compose <composition> [--out]  Compose a plant from the component library into a domain
  component list|show <ref>|validate    The component library: equipment, instruments, vocabulary
  component examples|example <id>       Ready-to-run example plants shipped with the library
                                        (built in; --components-dir adds customer
                                        directories, repeatable; STREAMSIM_COMPONENTS_DIRS)
  adapter list|verify [path]            Installed adapters; verify one against its golden
  run [--scenario <file>]               One-shot run from flags or a scenario file
  scenario from-run <run.json> --out f  Export the scenario that reproduces a run
  replay <run.json>                     Reproduce a run from its artifact
  verify <run.json>                     Verify a run artifact reproduces
  mcp --role director                   Serve the director role over stdio
                                        (--operator-addr binds the operator role
                                        over streamable HTTP; sim.world.create
                                        returns the endpoint with the token)
  canvas --workspace <dir> --ui-dir <d> Serve the visual canvas on a loopback address
                                        (prints the address and a pairing code)
  refconsumer --trace <file>            Detect episodes in a trace; write verdict.json
                                        (--mcp <url> --token <t> --run <id> closes
                                        the loop over the operator endpoint)
  suite --domain --profile              Generate an audited graded suite
  score --run --label                   Offline scorecard from artifacts
  report <run-dir>                      Summarize a run folder (--format json|markdown,
                                        --out <file>; markdown requires --out)
  manifest                              Write release-manifest.json (author +
                                        reviewer identity, optional ed25519)
  device serve --socket <path>          Serve the device emulator over a UDS
                                        (state/receipt link; --world enables plant wiring)
  help
`
