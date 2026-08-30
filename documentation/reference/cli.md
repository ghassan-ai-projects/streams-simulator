# Command-line interface (CLI) reference

> Status: Implemented reference. Authority: `internal/cli/cli.go`. Verified by: `streamsim help` and targeted command smoke tests. Last verified: 2026-08-17.

The executable is `streamsim`. During development, run it without installing a binary:

```bash
go run ./cmd/streamsim <command> [flags]
```

Most successful command results are JSON. The help surface is human-readable, and the long-running `mcp` command serves its protocol transports instead of returning a JSON result. Paths are interpreted by the current process.

## Commands

### `catalog`

Lists and describes loaded domains.

```text
streamsim catalog [list|describe|coverage] [flags]
```

Flags:

- `--domains-dir <path>` — domain directory; defaults to `domains`.
- `catalog describe <id>` — returns the spec and digest.

### `domain validate`

Validates one domain file and reports its ID, digest, channel count, fault count, and effector count.

```text
streamsim domain validate <path>
```

### `adapter`

Lists adapters or verifies one against its schema and golden fixture.

```text
streamsim adapter [list|verify] [flags]
```

Flags:

- `--adapters-dir <path>` — adapter directory; defaults to `adapters`.
- `adapter verify <path>` — verifies the selected adapter.

### `device serve`

Serves the data-defined device emulator over a Unix domain socket for a typed
gateway-link integration test. The emulator emits an initial `state` record,
accepts command records, and emits receipts; query-state control returns fresh
state. It is not a physical serial gateway.

```text
streamsim device serve --socket <path> --capabilities <path>
```

Flags:

- `--socket` — required Unix domain socket path. A regular file at this path is
  never removed.
- `--capabilities` — required device capability catalog JSON path.
- `--device-id` — device identity; defaults to `dev-01`.
- `--boot-id` — initial boot identity; defaults to `boot-A`.

### `run`

Runs one scripted world and writes an artifact when `--out` is supplied.

```text
streamsim run [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--domains-dir` | `domains` | Domain directory. |
| `--adapters-dir` | `adapters` | Adapter directory. |
| `--domain` | required | Domain ID. |
| `--adapter` | `native-jsonl` | Adapter ID. |
| `--seed` | `1` | World seed. |
| `--sink` | `inproc` | `inproc`, `file`, or `http-push`. |
| `--sink-target` | empty | File path or HTTP-push URL. Required for `http-push`; for `file`, `--out` can supply the trace path when this flag is omitted. |
| `--out` | empty | Artifact output directory. |
| `--duration` | `21600` | Duration in seconds. |
| `--start-time` | `2026-01-01T00:00:00Z` | World start as epoch nanoseconds. |
| `--fault` | empty | Comma-separated `entity=fault@offset_s` entries. |
| `--perturb` | empty | Comma-separated `name@from_s[@until_s]` entries. The optional end time is a second `@`-separated value. |
| `--effector` | empty | Comma-separated `effector@entity@offset_s` entries. The CLI currently sends empty arguments; use MCP for typed effector arguments. |
| `--profile` | empty | Scenario profile name. |

### `replay` and `verify`

Both load a run artifact and replay its command log. `verify` is the explicit verification spelling used in scripts. The current CLI also loads the matching domain and adapter from `domains/` and `adapters/`; keep those directories available or override them with the flags.

```text
streamsim replay <run.json> [--domains-dir <path>] [--adapters-dir <path>]
streamsim verify <run.json> [--domains-dir <path>] [--adapters-dir <path>]
```

### `mcp`

Runs the director MCP server over stdio. The optional operator endpoint is served by the same director process.

```text
streamsim mcp --role director [flags]
```

Flags:

- `--role director` — the only standalone role accepted by the CLI.
- `--operator-addr <address>` — streamable HTTP listen address for the operator endpoint.
- `--domains-dir <path>` — defaults to `domains`.
- `--adapters-dir <path>` — defaults to `adapters`.
- `--out <path>` — run artifact root; defaults to `runs`.

### `refconsumer`

Processes a native JSONL trace and writes a consumer verdict.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--trace` | required | Native-format trace file. |
| `--out` | `verdict.json` | Verdict output path. |
| `--threshold` | `4` | Detector z-score threshold. |
| `--window` | `30` | Detector window. |
| `--effector` | empty | Effector to invoke on detection. Requires `--mcp`. |
| `--mcp` | empty | Operator endpoint URL. |
| `--token` | empty | Capability token from `sim.world.create`. |
| `--run` | empty | Run ID to report against. |

### `suite`

Generates an audited suite. The working invocation is:

```text
streamsim suite --domain <id> [flags]
```

Flags:

- `--domains-dir` — defaults to `domains`.
- `--domain` — required domain ID.
- `--profile` — defaults to `nominal`.
- `--n` — target scenario count; defaults to `100`.
- `--seed` — suite seed; defaults to `1`.
- `--out` — output directory; defaults to `suites`.

### `score`

Scores a run from artifact files. `--run` is required; `--label` is optional and defaults to `label.json` beside the run artifact. The command reads `verdict.json` beside the artifact.

```text
streamsim score --run <run.json> [--label <label.json>]
```

### `manifest`

Writes `release-manifest.json` for the current commit and discovered inputs. Author and independent reviewer identities are required; the command also supports alternate output/input directories and an optional signing key. This is a release-evidence command; do not use an old checked-in manifest as evidence for a new commit.

Use `make manifest` for release evidence because the Makefile supplies linker metadata for the version and commit. A plain `go run` without equivalent `-ldflags` reports the development identity (`dev@none`).

```text
streamsim manifest --author "Name <author@example>" --reviewer "Reviewer <review@example>" [flags]
```

Flags:

- `--out` — output path; defaults to `release-manifest.json`.
- `--domains-dir` — defaults to `domains`.
- `--adapters-dir` — defaults to `adapters`.
- `--author` — required release-author identity.
- `--reviewer` — required independent-review identity.
- `--key` — optional file containing a 64-hex-character Ed25519 seed.

## Exit behavior

The CLI returns `0` on success, `1` when a command returns an error, and `2` for missing/unknown command usage. Preserve stderr and JSON output together when a command is part of an evaluation.

## Configuration model

There is no separate application configuration file. Configuration is the combination of CLI flags, JSON domain data, JSON adapter data, MCP tool arguments, and the run artifact. Keep those inputs under version control or in the evidence bundle when a run must be reproduced.

## Next reads

- [Quickstart](../getting-started/quickstart.md)
- [MCP reference](mcp.md)
- [Artifact reference](artifacts.md)
