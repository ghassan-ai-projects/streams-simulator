# Ubiquitous language — cli

The command-line module turns argv into one of the simulator's use cases and
prints one JSON document. It owns no simulator rules.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Command | One top-level word that selects a use case: catalog, domain, adapter, run, replay, verify, mcp, refconsumer, suite, score, manifest, device | `command`, `commands`, `handler` | `streamsim <command>` |
| Result | The single JSON document a successful command prints on stdout; commands that serve print nothing | `handler` return value | stdout |
| Failure document | The single JSON document a failed command prints on stdout: `{"error": {code, message, source?, issues?, result?}}` (`docs/design/CLI_SURFACE.md`) | `failureDocument`, `describeFailure` | `cli-output-v0.1` `$defs/error` |
| Failure code | The stable classification of a failure, e.g. `usage_invalid`, `domain_invalid`, `replay_mismatch`, `<command>_failed` | `failureCode`, `classifyFailure` | `error.code` |
| Usage failure | An invocation the command cannot run as written: a flag that does not parse, or a missing or unknown verb, argument or value | `usageExit`, `usageError`, `usagef` | `usage_invalid`, exit 2 |
| Outcome failure | A command that ran but whose check failed; it carries the command's result document | `outcomeError` | `replay_mismatch`, `adapter_verify_failed`, `error.result` |
| Exit status | 0 success or help, 1 failure, 2 usage failure | `Run`, `exitStatusOf` | process exit code |
| Scripted run | A one-shot run whose faults, perturbations and effectors are given as offsets from the start time | `applyRunScript` | `--fault`, `--perturb`, `--effector` |
| Build | The version and commit the binary was linked with | `Build` | `sim` in the manifest |
| Release manifest | The signed record of simulator, domain, adapter, suite and consumer digests with author and reviewer identity | `cmdManifest` | `release-manifest.json` |
| Director / operator endpoint | The stdio director session and the optional HTTP operator listener of `mcp` | `serve.Director` | `--operator-addr` |
| Device serve | The device emulator on a unix socket until interrupted | `serve.Device` | `device serve --socket` |
| Stdout owner | A command whose stdout is a protocol, so no document is printed there | `command.ownsStdout` | `mcp` |
| Environment | What a command may reach of its process: stdout, stderr, wall clock | `process.Env` | — |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `cli.Version`, `cli.Commit` (exported variables) | `cli.Build` passed to `cli.Main` | the facade exports no mutable state |
| `flag.ExitOnError` | `usageExit` | commands must be runnable in-process under test |
| `printJSON` inside commands | the dispatcher prints the returned result | one output site |
