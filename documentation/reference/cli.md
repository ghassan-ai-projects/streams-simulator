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
`domain_invalid`, `adapter_invalid`, `component_invalid`, `composition_invalid`,
`artifact_invalid`, `replay_mismatch`, `domain_lint`,
`adapter_verify_failed`, `io_failed`, and `<command>_failed` for any other failure of that
command. `issues` lists every contract violation, with an RFC 6901 `pointer` when the location
is known; each issue is `{pointer?, message, keyword?, member?}`. `keyword` says why, for a
program: the JSON Schema keyword that failed (`type`, `enum`, `pattern`, `required`,
`additionalProperties`, `minimum`, `exclusiveMinimum`, `maxLength`, ...) or a stable rule code of the
binary (`twin_member_required`, `twin_unit_required`, `twin_conversion_needed`, `twin_name_pattern`,
`twin_range`, `kind_range_empty`, `domain_rule`, `implements_drift`, `json` for a document that is not
JSON). `member` names the member a rule is about (`resolution`, `writable`, `safe_stop`, `unit`, or the
missing or extra property of `required` and `additionalProperties`). `message` stays plain English; a
surface never parses it. `result` carries the command's report for a failed check (for example the replay
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
streamsim domain validate <path> [--kinds-dir <dir>]...
```

- `--kinds-dir <dir>` — a directory of kind files (`*.kind.json`), repeatable. A domain kind that declares `implements: "<kind>"` must equal that kind file's metrics and commands as a host sees them in `NBIRTH.kinds`; a mismatch, an `implements` naming no loaded kind file, or no `--kinds-dir` at all fails with `domain_invalid`, an issue at `/kinds/<i>/implements` naming the first differing member. A domain without `implements` is unaffected and reads no kind file, so a malformed kind file in the directory blocks only a domain that implements one (and `kind validate`). The check compares the safe-stop flag of each command too: a domain flags a safe stop with `safe_stop: true` on its effector, and the kind file must flag the same commands. A kind file must be a regular file of at most 1 MiB (a symbolic link or a named pipe is refused), one JSON document, with no member named twice and no number beyond the range of a double. `domain validate` and `run` print `implements_checked`, a list of `{kind, kind_digest}` for the kind files the domain was checked against; it is empty when no kind of the domain implements one, so an empty list means there was nothing to check. `domain compose`, `run`, `plantwire node` and `mcp` take the same option and run the same check; `domain lint` takes none, because it judges a nominal day of the domain's behaviour, not its interface (`--kinds-dir` there is a usage error). A recorded run (`replay`, `verify`, `scenario from-run`) does not run it either: it reproduces a recorded run and sends nothing to a host.

### `kind validate`

Checks one or more kind files against the plant wire kind schema and prints `{valid, kinds}`, where each kind is `{path, kind, kind_digest, metrics, commands}`. A kind file is one `NBIRTH.kinds` entry without its `kind_digest`; its digest is the one a host sees. A kind file that carries a `kind_digest` member is refused (issue at `/kind_digest`), even a correct one, because a stale digest would be accepted and ignored; a metric's `eu_range` and `instrument_range` need `min` below `max` (issue at `/metrics/<i>/eu_range`). The kind file is the contract: `implements` proves a domain equal to it, and only an announced `kind_digest` equal to the file's proves a device matches it.

```text
streamsim kind validate [--twin] (<file.kind.json>... | --stdin)
```

A file that is not JSON, breaks the schema or defines a kind another file defines fails with `kind_failed`. A file that is not JSON or breaks the schema also carries `source` (the file) and `issues`, as `domain validate` does: one `{pointer, message, keyword, member?}` per schema violation, sorted by pointer, with `pointer` an RFC 6901 pointer into the kind file (for example `/metrics/0/data_type`, `/commands` for a missing member), `keyword` the schema keyword that failed (`required`, `enum`, `pattern`, ...) and `member` the missing or extra property; a file that is not JSON has an issue with `message` and `keyword: "json"` only. A kind file that breaks the schema in a `--kinds-dir` fails `domain validate`, `domain compose` and `run` the same way.

- `--stdin` — checks one kind document read from standard input instead of files: at most 1 MiB, decoded as strictly as a file (one JSON document, no member named twice, no number beyond a double), the same issues as a file with `source` `<stdin>`. It reads and writes no file; the output is `{valid, kinds}` with one kind whose `path` is `<stdin>`. Giving files beside `--stdin`, or neither, is a usage error.
- `--twin` — also reports everything a twin of the kind needs that the document lacks, **all at once**, each as an issue at the member in the kind file with `keyword` `twin_member_required` (`resolution` and `writable` on a reading; `name`, `args_schema`, `risk_class` and `safe_stop` on a command), `twin_member_unsupported`, `twin_name_pattern`, `twin_type_unsupported` (an enum or string reading), `twin_unit_required` (a number or integer reading without a unit), `twin_unit_unexpected` (a boolean with a unit), `twin_range` (an `eu_range` with `min` not below `max`), `twin_risk_class` or `twin_not_object`, and `member` the member concerned. Schema violations are reported first and alone: fix them, then the twin requirements follow in one list. A kind file that passes `--twin` produces a twin with `kind twin` without refusal for the kind file (the choices can still be refused). `kind twin` reports the same rules for a kind file it is given, all of them, at `/kinds/<i>/kind` of the choices.

### `kind twin`

Scaffolds a runnable twin of kind files: a domain whose kinds `implements` them. The binary does the work, so a surface needs no simulation knowledge of its own. Per metric, the choices name the library instrument that measures it (or `none` for a plain channel); per command that is not a safe stop, one of three fixed responses. Everything else comes from the kind files.

```text
streamsim kind twin [<file.kind.json>...] (--choices <choices.json> [--out <file>] | --suggest) [--kinds-dir <dir>]... [--components-dir <dir>]...
```

- `--suggest` — prints, and writes nothing, a valid choices document with the binary's defaults for the kind files, as `{choices}`, so a surface needs no defaults of its own: `id` from the kind names, `version` `0.1.0`, a `title`, one `instance` per kind (`<kind>-01`, underscores as hyphens), per reading the library instrument that fits its unit and range (`none` when none does, for a boolean and for a reading without a unit or range), and per command that is not a safe stop `direct` when the command's numeric arguments are named for the writable readings of the kind and no other command claims them, else `none`. It takes no `--choices` or `--out`. The suggestion always passes the schema and, for kind files that meet the twin requirements, `kind twin --choices`; a kind file that does not is refused with every violation at once, as `kind validate --twin` reports them.

- `--choices <file>` — the choices document, `twin-choices-v0.1` ([schema](../../contracts/twin-choices-v0.1.schema.json)): `id`, `version`, `title`, optional `description`, `site {name, area?}` and `board {name}`, and `kinds[]` in the order the domain declares them. Each entry is `{kind, instance, metrics, commands}`: `kind` names a kind file, `instance` is the one instance the twin has, `metrics` has one `{instrument, initial?}` for every metric of the kind file, `commands` one `{response, sets?, drives?}` for every command that is not a safe stop.
- Kind files come from the arguments, from `--kinds-dir` (every `*.kind.json` there), or both; only the kinds the choices name are used.
- `--out <file>` — write the domain there; without it the spec is printed inside the output document (`domain`).
- `--components-dir <dir>` — a customer library directory, repeatable (see `domain compose`).

Responses:

| `response` | Takes | Result |
| --- | --- | --- |
| `direct` | `sets`: argument name to the writable metric of the same kind it sets | The command assigns the argument to the metric's state |
| `first_order` | `sets`, and `drives[]` of `{from, kind, metric, gain, time_constant_s}` | As `direct`, and the read-only metric `metric` of `kind` relaxes with `time_constant_s` toward its `initial` plus `gain` times the value of `from` (a first-order lag) |
| `none` | nothing | The command is acknowledged and changes nothing |

A safe-stop command lowers every metric the kind's commands set to the low end of its range. A command whose arguments carry `lease_ms` (or `lease_s`) holds its effect for that lease and falls back to the low end of the range. An instrument keeps its noise, cadence, faults and tag; the kind file keeps the metric's unit, range and resolution, so the instrument must read in the metric's unit (an alias of the same quantity, such as `Cel` for `degC`, is accepted) and cover its range.

It prints `{id, digest, channels, faults, effectors, implements}` and `composed` (the file written) or `domain`. The twin is checked against the same kind files before it is printed, so it always passes `domain validate --kinds-dir`. Choices that break their schema, name something the kind files do not hold, leave a metric or command undecided, or ask a response that does not fit fail with `kind_failed`, each issue at its pointer in the choices (`/kinds/2/commands/set_pwm_lease/drives/0/gain`) with a `keyword` (`twin_kind_missing`, `twin_name_collision`, `twin_unknown_choice`, `twin_choice_missing`, `twin_instrument_mismatch`, `twin_response_mismatch`, `twin_initial_range`, or `twin_conversion_needed` with `member` `unit` for an instrument whose unit needs a conversion); a kind file the twin cannot reproduce (a string metric, a member beyond the interface, a missing `resolution`) fails the same way at `/kinds/<i>/kind`, one issue per violation, each with the twin `keyword` and `member`.

### `domain lint`

Simulates a nominal day of one domain (noiseless, default population, no faults, no effector calls) and judges it with five rules: a non-finite state or reading, a channel reading outside its declared range, an integrated state pinned at its bound, an integrated state that has not settled by the end of the run, and a conservation balance that does not close. It prints `{domain, digest, seed, hours, samples, findings, errors, warnings}`; each finding carries `severity`, `code`, `entity`, `channel`, `state` or `fault`, `at_s` and a `detail`. Any `error` finding exits 1 with a `domain_lint` error document whose `result` is the same report; warnings alone exit 0. The output is deterministic for a domain, seed and hours.

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

### `domain compose`

Composes a plant from the component library into an ordinary domain ([components](../../docs/design/COMPONENTS_V0_1.md)). The result is validated with the domain contract and written in canonical layout, so the same composition and library files always give the same bytes.

```text
streamsim domain compose <composition.json> [--out <file>] [--components-dir <dir>]... [--kinds-dir <dir>]...
```

- `--out <file>` — write the domain there; without it the composed spec is printed inside the output document (`domain`).
- `--components-dir <dir>` — a customer library directory, repeatable. The built-in library is embedded in the binary and always loaded first, so no directory is needed and the working directory does not matter; it is the only library that may hold `streamsim/` ids. `STREAMSIM_COMPONENTS_DIRS` (a path list, `:` on Unix) is appended after the flags; empty entries (a trailing `:`) are ignored.

It prints `{id, digest, channels, faults, effectors}` as `domain validate` does, plus `composed` (the file written). A composition that breaks a library rule fails with `composition_invalid`, each issue at its pointer in the composition (`/equipment/0/preset`, `/connections/1`); a composed domain that fails its own contract fails with `domain_invalid`.

### `component`

Lists, shows and validates the component library: equipment, instruments and vocabulary.

```text
streamsim component [list|show <namespace/name[@major]>|validate|examples|example <id>] [--components-dir <dir>]...
```

- `component list` — every entry as `{id, version, kind, title, class, tags, digest}` (no file path), sorted by id and version. For an instrument `class` is its unit family and `tags` its ISA-5.1 letters, and the entry also carries its `unit` and, when it states one, its measuring `range {min, max}`.
- `component list --unit <u>` — only the instruments that measure in `u` or in a unit convertible to it (the unit conversion `kind twin` uses), the exact unit first and then by id and version; each entry carries `match` (`exact` or `convertible`) beside its `unit` and `range`. A unit no family knows lists nothing. `--unit` belongs to `list` only.
- `component show <ref>` — the entry and the file's document; `ns/name` without a major selects the highest major.
- `component examples` — the ready-to-run example plants shipped with the library, as `{id, title, description}`.
- `component example <id>` — one example plant's composition, ready to copy into a workspace; an id that is not shipped fails with `component_failed`.
- `component new <id> --title <title> (--from <composition.json> | --example <id>) [--empty]` — a new plant's composition as `{composition}`: a full copy of the base with the new id and title, or with `--empty` only its frame (site, areas, axes, stresses, ground truth and profiles without fault weights; no equipment, connections or description). The id and title are checked against `composition-v0.1`; a bad id fails with `composition_invalid` and says what an id must be. Nothing is written.
- `component validate` — loads and checks every file of the directories and prints `{valid, directories, equipment, instruments}`, where `directories` are the customer directories (the built-in library is always checked and counted). A file that breaks its schema or a library rule fails with `component_invalid`, naming the file and listing every issue with its pointer.

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
                         [--policy <digest>]... [--kinds-dir <dir>]...
```

The socket's directory must be owner-only (it is created `0700` when missing); a stale socket left by a crash is replaced,
a live host or a regular file at the path is not. Both commands stop on interrupt or `SIGTERM`, and the node still writes its
run folder. The script is a JSON array of
`{id?, at_ns, device, command, args, verify?: {metric, equals, tolerance, within_ns}}`, issued in `at_ns` order; `id` (a
unique wire-safe name, default `ref-<n>` by position) becomes the command's `command_id`; a command to a
device that is quarantined or never born is not sent and is reported `denied`. A command is verified only by a sample that
arrives after its result. The inventory is a JSON object of device to kind; without one, every device is accepted. A node
must send its birth within 10 s; after a refusal the host closes the connection. The node accepts commands signed by a `--policy` it lists. The commands that skip freshness, epoch and policy checks are the
safe stops the domain flags (`safe_stop: true` on an effector); a domain that implements a kind file is checked against it
(`--kinds-dir`, as `domain validate`) before the node connects, so the safe stops and the `kind_digest` it announces are the kind file's. `--barrier-cap-ms` (default 30000) bounds the wall time the node waits for each
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
streamsim run --scenario <file> [--domains-dir d] [--adapters-dir d] [--kinds-dir d]... [--sink s] [--sink-target t] [--out dir] [--explore]
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
| `--sink-target` | empty | File path or HTTP-push URL. Required for `http-push`; for `file`, `--out` can supply the trace path when this flag is omitted. With a target outside `--out`, the trace is also written to `trace.jsonl` in `--out`, so the run folder is complete for `verify` and `report`. |
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

Both load a run artifact and replay its command log. `replay` prints the replay report and exits 0 whatever the outcome. `verify` prints the same report when the replay reproduces the recorded trace digest; otherwise it exits 1 with a `replay_mismatch` error document whose `result` is the report (a simulator-version difference alone does not fail it; it is reported in `version_match` and `detail`), so scripts can gate on it. An artifact whose recorded domain or adapter specification no longer hashes to its recorded digest, or whose recorded specifications replay to a different world digest, is also `replay_mismatch`, with a `result` whose `matches` is `false` and whose `detail` names the reason. For a run that ended complete, `verify` requires every evidence file beside the artifact (`trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, `journal.jsonl`, `sealed/journal.jsonl`) and a digest in `run.json` for each (the run artifact schema requires them), and compares each file with its recorded digest and with the replay; a missing file, a missing digest or a file that differs is `matches: false`, and the `detail` names the file. A run started with `--sink file --sink-target <path>` writes its trace to the target and also to `trace.jsonl` in the run folder, so the folder is always complete. The report carries a `reason` (`match`, `recorded_spec`, `world_digest`, `trace_divergence`, `journal_divergence`, `evidence_divergence`, `unblinded_stamp`, `digest_missing`, `file_differs`, `file_missing`, `incomplete` or `incomplete_stamp`) and, for the file reasons, the `file` it names; the document also carries `emitted`, `incomplete` and, when a journal stream differs from the replay, `journal_diverged`. An unreadable or missing `run.json` is an error with no `result`. Replay also rebuilds both journal streams in memory: `verify` fails with `replay_mismatch` when either differs from the digest in `run.json` (`journal_digest`, `sealed_journal_digest`), or from `journal.jsonl` or `sealed/journal.jsonl` beside the artifact, and the `detail` names the stream and the first differing record. A run that ended incomplete (a sink failure, or a consumer that did not quiesce) is never verified: `verify` exits 1 with `reason: incomplete` even when the replay and every file agree, because its close cannot be reproduced; the journals and the replayed ledger and history are not compared with the replay, but every evidence file and digest it does have is compared with each other. The `incomplete` stamp of `run.json` is checked against the `run.closed` record of each journal and against `reproducible` (an incomplete run is never reproducible): a contradiction is `reason: incomplete_stamp`, which is what editing `incomplete` on a closed run produces. An empty or malformed evidence digest counts as a missing one (`digest_missing`). The replay uses the domain and adapter specifications recorded in `run.json`; the `domains/` and `adapters/` directories (or the flags that override them) are read only for an older artifact that records none, and to report `domain_changed`.

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
- `--components-dir <dir>` — a customer component library directory, repeatable (see `domain compose`).
- `--kinds-dir <dir>` — a directory of kind files, repeatable. `sim.world.create` runs the same `implements` check as `domain validate` on the domain it is asked for, and refuses a domain whose kind implements a kind file this director cannot find (`domain_invalid`, an issue at `/kinds/<i>/implements`). A director started without `--kinds-dir` refuses every domain that declares `implements`; a domain without `implements` is unaffected.

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

Scores a run from the files it published, with the scorer the director uses online, so the offline and online scorecards of one run agree. The one input the files cannot carry is a `reveal` with `unblind:true` made after the run ended: the artifact stamps unblinding at its end, so that later reveal affects only the online scorer. `--run` is required; `--label` defaults to `label.json` beside the run artifact. The command reads `verdict.json` beside the artifact and replays the artifact (against `--domains-dir` and `--adapters-dir`); the replay supplies the effector calls, the delivery ledger and the state history, so an edited `ledger.jsonl` or `world_state_history.jsonl` cannot change a grade. `verify` compares both files with the `ledger_digest` and `history_digest` that `run.json` records and with the replay, and `report` checks every evidence file, `world_state_history.jsonl` included, against its digest. A replay that does not reproduce the recorded trace is refused, not graded, and so is a run stamped unblinded (an `--explore` run, or one whose truth was revealed while it was open): it fails with `score_failed`.

```text
streamsim score --run <run.json> [--label <label.json>] [--domains-dir <path>] [--adapters-dir <path>]
```

### `report`

Summarises a run folder (the directory `run --out` publishes) as one document: run identity and inputs, the digests the run recorded, whether the files in the folder still agree with them (`verification`: a check, not a replay; `verify` is the proof), delivery counts and drop reasons, the faults (each as an interval: `fault_id`, `entity_id`, `fault`, `injected_at_ns`, `onset_ns`, `cleared_ns` and `cleared_by`, which is `command`, `pulse_end` or `responder`, both null while the fault is active at the end of the run), the perturbations, environment faults and effector calls from the command log with the world's real outcomes, the consumer from `verdict.json` (its `config_digest` only when the verdict recorded one), and the scorecard when `verdict.json` and `label.json` are both present (scored from the specs the artifact embeds, so no `--domains-dir` is needed). It also says whether the run is `evaluate` or `explore` and whether it is complete. The report is a pure function of the folder: the same folder gives the same bytes. It carries truth, so the truth sections appear only for a run that has ended (`run.json` exists) or an Explore run; a folder cut before `run.json` is reported from its observer journal alone (the ledger and `sealed/` are not read), as incomplete and not ended, with `truth_unavailable: "sealed until the run ends"`. A `run.json` or sealed journal that does not belong to the run of the observer journal is not read: the report says why in `truth_unavailable` and shows no truth. A torn last line of a journal or the ledger (a crashed run) is dropped and named in `torn_tail`. A run whose domain declares KPI states or alarms carries `truth.kpis`: the true mean, minimum and maximum of every KPI and the ISA-18.2 KPIs of every alarm and of the alarm system. `verification` also checks both journal files as written, session records included (`journal_file`, `sealed_journal_file`). It loads `run.json` as `verify` does: a run that ended complete but records no usable digest for an evidence file (absent, empty or malformed) is a failing verification (`files_differ`, a `digest_missing` check for that file), not an error. Design: [RUN_REPORT.md](../../docs/design/RUN_REPORT.md); contract: [`run-report-v0.1`](../../contracts/run-report-v0.1.schema.json).

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
