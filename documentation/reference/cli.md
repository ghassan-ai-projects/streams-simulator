# Command-line interface (CLI) reference

> Status: Implemented reference. Authority: `internal/cli/internal/app/cli.go` and [`cli-output-v0.1`](../../contracts/cli-output-v0.1.schema.json). Verified by: `streamsim help`, targeted command smoke tests, and contract tests over every command's output. Last verified: 2026-10-12.

The executable is `streamsim`. During development, run it without installing a binary:

```bash
go run ./cmd/streamsim <command> [flags]
```

Paths are interpreted by the current process. Flags may come before, between or after
positional arguments; everything after `--` is positional (a `--` given as a flag's value is
that value). A command given more positional arguments than it takes fails with
`usage_invalid` naming the first unexpected one.

## Output contract

Every command except `help` and `mcp` prints exactly **one JSON document** on stdout, for
success and for failure ([design](../../docs/design/CLI_SURFACE.md), [schema](../../contracts/cli-output-v0.1.schema.json)). stderr carries human diagnostics only.

| Exit | stdout |
| --- | --- |
| 0 | The command's result document (help prints nothing) |
| 1 | `{"error": {"code", "message", "source"?, "issues"?, "result"?}}` |
| 2 | The same error document with code `usage_invalid` |

Error codes: `usage_invalid`, `file_not_found`, `domain_not_found`, `adapter_not_found`,
`domain_invalid`, `adapter_invalid`, `artifact_invalid`, `replay_mismatch`, `domain_lint`,
`adapter_verify_failed`, `io_failed`, and `<command>_failed` for any other failure of that
command. `issues` lists every contract violation, with an RFC 6901 `pointer` when the location
is known; `result` carries the command's report for a failed check (for example the replay
report of a failed `verify`). `mcp` owns its stdout for the protocol and reports failures on
stderr only; `device serve` prints no success document while it serves.

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

### `domain lint`

Simulates a nominal day of one domain (noiseless, default population, no faults, no effector calls) and judges it with five rules: a non-finite state or reading, a channel reading outside its declared range, a state pinned at its bound, an integrated state that has not settled by the end of the run, and a conservation balance that does not close. It prints `{domain, digest, seed, hours, samples, findings, errors, warnings}`; each finding carries `severity`, `code`, `entity`, `channel`, `state` or `fault`, `at_s` and a `detail`. Any `error` finding exits 1 with a `domain_lint` error document whose `result` is the same report; warnings alone exit 0. The output is deterministic for a domain, seed and hours.

```bash
streamsim domain lint <path> [--hours 24] [--seed 1]
```

- `--hours <n>` — simulated hours of the nominal day, above 0 and at most 720; defaults to `24`.
- `--seed <n>` — world seed; defaults to `1`.

### `domain format`

Rewrites one valid domain file in the canonical layout and prints `{formatted, changed, digest}`. It is idempotent and never changes the digest; a document that does not validate is refused and left untouched.

```text
streamsim domain format <path>
```

### `adapter`

Lists adapters or verifies one against its schema and golden fixture.

```text
streamsim adapter [list|verify] [flags]
```

Flags:

- `--adapters-dir <path>` — adapter directory; defaults to `adapters`.
- `adapter verify <path>` — renders the conformance fixture and checks it against the adapter's declared output schema and golden file. It fails when a declared check fails, and when the adapter declares neither (nothing would be compared); the result reports `schema_checked` and `golden_checked`.

### `device serve`

Serves the data-defined device emulator over a Unix domain socket for a typed
gateway-link integration test. The emulator emits an initial `state` record,
accepts command records, and emits receipts; query-state control returns fresh
state. With `--world`, accepted commands are routed through the data-defined
world effector plant and lease expiry invokes its safe-stop path. It is not a
physical serial gateway.

```text
streamsim device serve --socket <path> --capabilities <path> [--world <bindings.json>]
```

Flags:

- `--socket` — required Unix domain socket path. A regular file at this path is
  never removed.
- `--capabilities` — required device capability catalog JSON path.
- `--world` — optional deviceworld binding catalog JSON path. When present, the
  default world is `domains/cold-chain-transit.domain.json` and accepted device
  commands invoke its declared effectors.
- `--world-domain` — world domain JSON path; used only with `--world`.
- `--world-entity` — world entity for the bindings; defaults to the first world
  entity.
- `--device-id` — device identity; defaults to `dev-01`.
- `--boot-id` — initial boot identity; defaults to `boot-A`.
- `--fault` — repeatable deterministic fault in the form
  `name[@accepted-command-ordinal]`. Supported names are `ack_lost`, `stuck`,
  `reboot`, `duplicate`, `stale`, `expired`, and `disconnect`. The default
  ordinal is one. Fault application is logged to stderr; duplicate delivery
  remains idempotent and disconnect closes only the current UDS connection.

### `plantwire node` and `plantwire host`

Speaks the plant wire ([contract](../../contracts/plant-wire/SPEC.md)). `plantwire node` runs a domain as a node: it dials the
host's owner-only Unix socket, sends births, data in slices closed by an `NSYNC`/`HSYNC` barrier, executes the host's commands
through the run's effector path, and writes the run's artifact and journal (with `wire` and `pacing` records) to `--out`.
`plantwire host` is the reference host: it creates an owner-only socket, accepts one node, checks the environment, version,
contract and that every commissioned device is in the node's birth, quarantines devices its inventory does not list, requests a rebirth on a gap, issues scripted commands
when the world reaches their time, verifies them against later observations, and prints its report.

```text
streamsim plantwire host --socket <path> [--script <commands.json>] [--inventory <devices.json>] [--environment simulated]
                         [--policy <digest>] [--epoch <n>] [--delay-ms <n>]
streamsim plantwire node --host-socket <path> --domain <id> --out <dir> [--domains-dir domains] [--adapter <id>] [--adapters-dir adapters]
                         [--session-id <id>] [--seed <n>] [--slice-s 60] [--duration 3600] [--barrier-cap-ms 30000]
                         [--policy <digest>]... [--safe-stop <command>]...
```

The socket's directory must be owner-only (it is created `0700` when missing); a stale socket left by a crash is replaced,
a live host or a regular file at the path is not. Both commands stop on interrupt or `SIGTERM`, and the node still writes its
run folder. The script is a JSON array of
`{id?, at_ns, device, command, args, verify?: {metric, equals, tolerance, within_ns}}`, issued in `at_ns` order; `id` (a
unique wire-safe name, default `ref-<n>` by position) becomes the command's `command_id`; a command to a
device that is quarantined or never born is not sent and is reported `denied`. A command is verified only by a sample that
arrives after its result. The inventory is a JSON object of device to kind; without one, every device is accepted. A node
must send its birth within 10 s; after a refusal the host closes the connection. The node accepts commands signed by a `--policy` it lists; `--safe-stop` names commands that
skip freshness, epoch and policy checks. `--barrier-cap-ms` (default 30000) bounds the wall time the node waits for each
`HSYNC`; `--delay-ms` makes the reference host slow.

### `run`

Runs one world and writes an artifact when `--out` is supplied. The run is defined either by a
[scenario file](../../docs/design/SCENARIO_FILE.md) (`--scenario`) or by the flags below; both take
the same path: the flags build a scenario in memory. Before anything is delivered, the scenario is
checked against the domain and world: every effector, fault, perturbation and profile named must
exist; effector arguments and fault and perturbation parameters must be accepted; perturbation
windows must not end before they start; times must be plain integers; the clock may not move more
than 30 days past the start. The scenario is then dry-run on the same world with emission off,
which finds the first command the run would refuse as it unfolds (an entity that does not exist
at that point, an id already in use, a clear naming an id never assigned). A scenario file that fails is `scenario_invalid`, with every issue
located in the file. Flags that fail are `usage_invalid`, with every issue naming the flag
occurrence behind it. During the run, an effector call the world refuses through an interlock or
a reused command id is a recorded outcome.
A scenario's `description` is for readers; a scenario without a `clock.advance` emits nothing.

```text
streamsim run [flags]
streamsim run --scenario <file> [--domains-dir d] [--adapters-dir d] [--sink s] [--sink-target t] [--out dir] [--explore]
```

With `--scenario`, the flags the scenario defines (`--domain`, `--adapter`, `--seed`,
`--profile`, `--start-time`, `--duration`, `--fault`, `--perturb`, `--effector`) are refused.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--domains-dir` | `domains` | Domain directory. |
| `--adapters-dir` | `adapters` | Adapter directory. |
| `--domain` | required | Domain ID. |
| `--adapter` | `native-jsonl` | Adapter ID. |
| `--seed` | `1` | World seed. |
| `--sink` | `inproc` | `inproc`, `file`, or `http-push`. |
| `--sink-target` | empty | File path or HTTP-push URL. Required for `http-push`; for `file`, `--out` can supply the trace path when this flag is omitted. |
| `--out` | empty | Artifact output directory. With `--out`, the run also writes its [run journal](../../docs/design/RUN_JOURNAL.md) while it runs: `journal.jsonl` (what a consumer could know) and `sealed/journal.jsonl` (truth-bearing, director only). Without `--out` there is no journal file. |
| `--explore` | `false` | Explore mode: the run is stamped unblinded from creation (`mode: "explore"`, `unblinded: true` in `run.json`, a `run.unblinded` journal record) and `score` refuses it. Allowed beside `--scenario`. |
| `--duration` | `21600` | Duration in seconds; positive and at most 2592000 (30 days). |
| `--start-time` | `2026-01-01T00:00:00Z` | World start as epoch nanoseconds, between 0 and 4611686018427387903. |
| `--fault` | empty | Repeatable; each occurrence may be comma-separated `entity=fault@offset_s` entries. |
| `--perturb` | empty | Repeatable; each occurrence may be comma-separated `name@from_s[@until_s]` entries. The optional end time is a second `@`-separated value and may not be before the start. |
| `--effector` | empty | Repeatable. `effector@entity@offset_s[@{"arg":value}]`; the optional JSON object carries the effector arguments (required by every shipped effector). Command ids are `cli-0`, `cli-1`, … in order, so identical invocations record identical command logs. Entries without arguments may still be comma-separated within one occurrence; an entry with arguments takes an occurrence of its own. |
| `--profile` | empty | Scenario profile name; it must be one of the domain's profiles. |

### `scenario from-run`

Writes the scenario that reproduces a recorded run. Running it with `run --scenario` gives the
same command log, trace digest and world digest for any run made by the CLI; for a
director-driven run, the same trace and world digests. A clock advance is exported without
`await_consumer`, which is pacing of the live session, not an input to the world.

```text
streamsim scenario from-run <run.json> --out <file> [--domains-dir domains] [--adapters-dir adapters]
```

The scenario is checked against its contract and then as `run --scenario` would check it, against
the domain and adapter in the given directories. The artifact's domain and adapter must
be the ones it recorded (`replay_mismatch` otherwise). Effector calls the world refused at
admission changed nothing; they are dropped and counted in `refused`. A run that still cannot be
expressed as a runnable scenario writes nothing and fails with `scenario_invalid` naming the
artifact; `replay` and `verify` still reproduce it.

Prints `{"scenario_written", "commands", "omitted", "refused"}`; `omitted` counts command-log entries that
are not scenario commands (lifecycle records and the record-only `env.inject` of director runs).

### `replay` and `verify`

Both load a run artifact and replay its command log. `replay` prints the replay report and exits 0 whatever the outcome. `verify` prints the same report when the replay reproduces the recorded trace digest; otherwise it exits 1 with a `replay_mismatch` error document whose `result` is the report (a simulator-version difference alone does not fail it; it is reported in `version_match` and `detail`), so scripts can gate on it. An artifact whose domain, adapter or world digest no longer matches is also `replay_mismatch`. Replay also rebuilds both journal streams in memory: `verify` fails with `replay_mismatch` when either differs from the digest in `run.json` (`journal_digest`, `sealed_journal_digest`), or from `journal.jsonl` or `sealed/journal.jsonl` beside the artifact, and the `detail` names the stream and the first differing record. A run that ended incomplete is not compared (its close cannot be reproduced). The current CLI also loads the matching domain and adapter from `domains/` and `adapters/`; keep those directories available or override them with the flags.

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

Scores a run from the files it published, with the scorer the director uses online, so the offline and online scorecards of one run agree. The one input the files cannot carry is a `reveal` with `unblind:true` made after the run ended: the artifact stamps unblinding at its end, so that later reveal affects only the online scorer. `--run` is required; `--label` defaults to `label.json` beside the run artifact. The command reads `verdict.json` beside the artifact and replays the artifact (against `--domains-dir` and `--adapters-dir`); the replay supplies the effector calls, the delivery ledger and the state history, so an edited `ledger.jsonl` or `world_state_history.jsonl` cannot change a grade. `verify` compares both files with the `ledger_digest` and `history_digest` that `run.json` records and with the replay, and `report` checks `ledger.jsonl` against its digest. A replay that does not reproduce the recorded trace is refused, not graded, and so is a run stamped unblinded (an `--explore` run, or one whose truth was revealed while it was open): it fails with `score_failed`.

```text
streamsim score --run <run.json> [--label <label.json>] [--domains-dir <path>] [--adapters-dir <path>]
```

### `report`

Summarises a run folder (the directory `run --out` publishes) as one document: run identity and inputs, the digests the run recorded, whether the files in the folder still agree with them (`verification`: a check, not a replay; `verify` is the proof), delivery counts and drop reasons, the faults, perturbations, environment faults and effector calls from the command log with the world's real outcomes, the consumer from `verdict.json` (its `config_digest` only when the verdict recorded one), and the scorecard when `verdict.json` and `label.json` are both present (scored from the specs the artifact embeds, so no `--domains-dir` is needed). It also says whether the run is `evaluate` or `explore` and whether it is complete. The report is a pure function of the folder: the same folder gives the same bytes. It carries truth, so the truth sections appear only for a run that has ended (`run.json` exists) or an Explore run; a folder cut before `run.json` is reported from its observer journal alone (the ledger and `sealed/` are not read), as incomplete and not ended. Design: [RUN_REPORT.md](../../docs/design/RUN_REPORT.md); contract: [`run-report-v0.1`](../../contracts/run-report-v0.1.schema.json).

```text
streamsim report <run-dir> [--format json|markdown] [--out <file>]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--format` | `json` | `json` or `markdown`. |
| `--out` | empty | Write the report to this file. Required for `markdown`. |

Without `--out`, the JSON report is the command's one stdout document. With `--out`, the file holds the chosen format and stdout holds `{"report_written": "<file>", "format": "<format>"}`. Markdown is never printed on stdout, so stdout stays one JSON document. A folder with neither `run.json` nor `journal.jsonl` fails with `file_not_found`; a verdict or artifact that breaks its contract fails with `artifact_invalid`.

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
