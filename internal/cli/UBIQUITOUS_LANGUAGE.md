# Ubiquitous language — cli

The command-line module turns argv into one of the simulator's use cases and
prints one JSON document. It owns no simulator rules and no I/O beyond the
process streams and the files its flags name: every command calls the facade of
the module that owns the work. `test/architecture/cli_wiring_test.go` holds the
line: a file, an import or a construct beyond wiring fails the build.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Command | One top-level word that selects a use case: catalog, domain, adapter, run, replay, verify, mcp, refconsumer, suite, score, manifest, device | `command`, `commands`, `handler` | `streamsim <command>` |
| Result | The single JSON document a successful command prints on stdout; commands that serve print nothing | `handler` return value | stdout |
| Failure document | The single JSON document a failed command prints on stdout: `{"error": {code, message, source?, issues?, result?}}` (`docs/design/CLI_SURFACE.md`) | `failureDocument`, `describeFailure` | `cli-output-v0.1` `$defs/error` |
| Failure code | The stable classification of a failure, e.g. `usage_invalid`, `domain_invalid`, `replay_mismatch`, `<command>_failed` | `failureCode`, `classifyFailure` | `error.code` |
| Usage failure | An invocation the command cannot run as written: a flag that does not parse, a missing or unknown verb, argument or value, or an argument beyond those the command or verb takes | `usageExit`, `usageError`, `usagef`, `parseOnlyFlags`, `refuseExtraArguments`, `verbArity` | `usage_invalid`, exit 2 |
| Outcome failure | A command that ran but whose check failed; it carries the command's result document | `outcomeError` | `replay_mismatch`, `adapter_verify_failed`, `error.result` |
| Exit status | 0 success or help, 1 failure, 2 usage failure | `Run`, `exitStatusOf` | process exit code |
| Scripted run | A run described by flags: faults, perturbations and effectors as offsets from the start time; the flags build a scenario | `scenarioFromFlags` | `--fault`, `--perturb`, `--effector` |
| Flag origin | The flag occurrence behind a command of a scripted run; a refusal names it instead of pointing into a document the user never wrote | `scripted`, `flagOrigin`, `flagScenarioFailure` | `--effector "…": …` |
| Scenario-defined flag | A run flag a scenario file defines, refused beside `--scenario`: exactly the flags `registerScenarioFlags` registers | `registerScenarioFlags`, `scenarioDefinedFlags` | `--domain`, `--seed`, … |
| Scenario file | A run defined by a `scenario-v0.1` file instead of flags. Not a domain's profile, and not a suite's generated scenario | `loadScenarioFile` | `run --scenario` |
| Kind twin | A domain scaffolded from kind files and a choices document by the `component` module, checked against the same kind files by `plantwire` before it is printed or written | `cmdKindTwin`, `component.Twin` | `kind twin --choices` |
| Kind validate stdin | One kind document read from standard input (`--stdin`), at most 1 MiB and strict like a file, named `<stdin>`; `--twin` adds every twin requirement to the schema's issues | `cmdKindValidate`, `validateKindStream` | `kind validate --stdin [--twin]` |
| Kind twin suggestion | The binary's default choices for kind files, printed as `{choices}` and written nowhere | `suggestChoices`, `component.Suggest` | `kind twin --suggest` |
| Scenario export | The scenario that reproduces a recorded run | `cmdScenario`, `exportScenario` | `scenario from-run` |
| Build | The version and commit the binary was linked with | `Build` | `sim` in the manifest |
| Release manifest | The signed record of simulator, domain, adapter, suite and consumer digests with author and reviewer identity; written by the `release` module | `cmdManifest`, `release.Publish` | `release-manifest.json` |
| Director / operator endpoint | The stdio director session and the optional HTTP operator listener of `mcp`; served by the `mcp` module | `mcp.ServeDirector` | `--operator-addr` |
| Device serve | The device emulator on a unix socket until interrupted; composed by `deviceworld`, served by `device` | `deviceworld.OpenDevice`, `device.Serve` | `device serve --socket` |
| Wiring | What the command layer may hold: flag parsing, scenario from flags, the command registry, JSON output and failure documents; each file declares the facades it may import and a line cap | `cliAppFiles`, `cliEdges` (gate) | – |
| Stdout owner | A command whose stdout is a protocol, so no document is printed there | `command.ownsStdout` | `mcp` |
| Environment | What a command may reach of its process: stdout, stderr, wall clock | `process.Env` | — |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `cli.Version`, `cli.Commit` (exported variables) | `cli.Build` passed to `cli.Main` | the facade exports no mutable state |
| `flag.ExitOnError` | `usageExit` | commands must be runnable in-process under test |
| `printJSON` inside commands | the dispatcher prints the returned result | one output site |
| `serve` edge, manifest signing, score workflow, catalog and adapter directory loading, device-world wiring, reference-consumer orchestration | the `mcp`, `device`, `release`, `grading`, `domain`/`adapter`, `deviceworld` and `refconsumer` modules | the CLI is wiring; the module that owns the work owns its code |
