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

Every command except `help`, `mcp` and `canvas` (which owns its stdout and prints no document) prints exactly **one JSON document** on stdout, for
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
binary (`twin_member_required`, `twin_unit_required`, `twin_instrument_mismatch`, `twin_name_pattern`,
`twin_range`, `kind_range_empty`, `domain_rule`, `implements_drift`, `json` for a document that is not
JSON). `member` names the member a rule is about (`resolution`, `writable`, `safe_stop`, `unit`, `eu_range`, `data_type`, or the
missing or extra property of `required` and `additionalProperties`). `message` stays plain English; a
surface never parses it. `result` carries the command's report for a failed check (for example the replay
report of a failed `verify`). `mcp` and `canvas` own their stdout (the protocol; the server) and report failures on
stderr only, a failing `canvas` exiting 1 or 2 with nothing on stdout; `device serve` prints no success document while it serves.

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

- `--kinds-dir <dir>` — a directory of kind files (`*.kind.json`), repeatable. A domain kind that declares `implements: "<kind>"` must equal that kind file's metrics and commands as a host sees them in `NBIRTH.kinds`; a mismatch, an `implements` naming no loaded kind file, or no `--kinds-dir` at all fails with `domain_invalid`, an issue at `/kinds/<i>/implements` naming the first differing member. A domain without `implements` is unaffected and reads no kind file. A kind that also declares `board_type: "<type>"` is the raw model of that board type and implements its kind file through the board mapping `<type>.board.json` (plant wire `BOARD`) found in the same directories by the `board_type` it declares: the kind's channels and effectors are projected through the mapping (names, words, scaled ranges and the arguments each raw command needs are checked against the raw kind and the kind file), and the projected interface must equal the kind file, so `implements_checked` carries the kind file's own digest. A mapping that is not loaded is `implements_missing`; one that disagrees with the raw kind or the kind file is `implements_drift`, with the first differing member in the message. A node of such a domain announces the kind file, translates each command at its boundary into the raw command and arguments the twin runs (the raw command is what the run's command log records), and reports each raw reading as the kind's metric; a raw reading the mapping does not name is not sent. A domain that implements reads only the kind files of the kinds it implements, found by the kind a file declares (a file that does not parse or read is known by its file name, so its failure is reported when its name is an implemented kind); a malformed kind file of another kind is never read as a problem, so one unfinished device does not block the others (it still fails `kind validate`). The check compares the safe-stop flag of each command too: a domain flags a safe stop with `safe_stop: true` on its effector, and the kind file must flag the same commands. A kind file must be a regular file of at most 1 MiB (a symbolic link or a named pipe is refused), one JSON document, with no member named twice and no number beyond the range of a double. `domain validate` and `run` print `implements_checked`, a list of `{kind, kind_digest}` for the kind files the domain was checked against; it is empty when no kind of the domain implements one, so an empty list means there was nothing to check. `domain compose`, `run`, `plantwire node` and `mcp` take the same option and run the same check; `domain lint` takes none, because it judges a nominal day of the domain's behaviour, not its interface (`--kinds-dir` there is a usage error). A recorded run (`replay`, `verify`, `scenario from-run`) does not run it either: it reproduces a recorded run and sends nothing to a host.

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

- `--suggest` — prints, and writes nothing, a valid choices document with the binary's defaults for the kind files, as `{choices}`, so a surface needs no defaults of its own: `id` and `title` from the kind names of the given files only (`fan-twin` and `Twin of fan`; for a board too long for an id, the first kind and a count, `board-part-number-0-and-7-more-twin`), `version` `0.1.0`, one `instance` per kind (`<kind>-01`, underscores as hyphens), per reading the library instrument that fits its unit and range (`none` when none does, for a boolean and for a reading without a unit or range), and per command that is not a safe stop `direct` when the command's numeric arguments are named for the writable readings of the kind and no other command claims them, else `none`. It takes no `--choices` or `--out`. The suggestion always passes the schema and, for kind files that meet the twin requirements, `kind twin --choices`; a kind file that does not is refused with every violation at once, as `kind validate --twin` reports them.

- `--choices <file>` — the choices document, `twin-choices-v0.1` ([schema](../../contracts/twin-choices-v0.1.schema.json)): `id`, `version`, `title`, optional `description`, `site {name, area?}` and `board {name}`, and `kinds[]` in the order the domain declares them. Each entry is `{kind, instance, metrics, commands}`: `kind` names a kind file, `instance` is the one instance the twin has, `metrics` has one `{instrument, initial?}` for every metric of the kind file, `commands` one `{response, sets?, drives?}` for every command that is not a safe stop.
- Kind files come from the arguments, from `--kinds-dir` (every `*.kind.json` there, all of which must parse), or both; only the kinds the choices name are used. Give the files as arguments to keep an unrelated kind file out of the twin altogether; `--suggest` names its default `id` and `title` after the files it is given.
- An instrument whose unit differs from the metric's is used through a unit conversion: the twin keeps the unit of its kind file, and the instrument's noise, drift and fault magnitudes are converted into it.
- `--out <file>` — write the domain there; without it the spec is printed inside the output document (`domain`).
- `--components-dir <dir>` — a customer library directory, repeatable (see `domain compose`).

Responses:

| `response` | Takes | Result |
| --- | --- | --- |
| `direct` | `sets`: argument name to the writable metric of the same kind it sets | The command assigns the argument to the metric's state |
| `first_order` | `sets`, and `drives[]` of `{from, kind, metric, gain, time_constant_s}` | As `direct`, and the read-only metric `metric` of `kind` relaxes with `time_constant_s` toward its `initial` plus `gain` times the value of `from` (a first-order lag) |
| `none` | nothing | The command is acknowledged and changes nothing |

A safe-stop command lowers every metric the kind's commands set to the low end of its range. A command whose arguments carry `lease_ms` (or `lease_s`) holds its effect for that lease and falls back to the low end of the range. An instrument keeps its noise, cadence, faults and tag; the kind file keeps the metric's unit, range and resolution, so the instrument must read in the metric's unit (an alias of the same quantity, such as `Cel` for `degC`, is accepted) and cover its range.

It prints `{id, digest, channels, faults, effectors, implements}` and `composed` (the file written) or `domain`. The twin is checked against the same kind files before it is printed, so it always passes `domain validate --kinds-dir`. Choices that break their schema, name something the kind files do not hold, leave a metric or command undecided, or ask a response that does not fit fail with `kind_failed`, each issue at its pointer in the choices (`/kinds/2/commands/set_pwm_lease/drives/0/gain`) with a `keyword` (`twin_kind_missing`, `twin_name_collision`, `twin_unknown_choice`, `twin_choice_missing`, `twin_instrument_mismatch`, `twin_response_mismatch`, `twin_initial_range`) whose `member` names the part of the metric at fault: `unit` for an instrument of another quantity, `eu_range` for a metric without a range or one the instrument's range does not cover (compared after conversion into the instrument's unit), `data_type` for a boolean metric; a kind file the twin cannot reproduce (a string metric, a member beyond the interface, a missing `resolution`) fails the same way at `/kinds/<i>/kind`, one issue per violation, each with the twin `keyword` and `member`.

### `kind capture`

Drafts kind files from the `NBIRTH` a node announced, for a person to review: a birth is a claim, so a captured kind becomes the model only after review, and the command writes nothing. The source is one `NBIRTH` record (a file holding it, as one line or indented), or a JSON Lines file in which the **first** `NBIRTH` is taken: a recorded wire stream, or a run's observer journal (`journal.jsonl`), whose `wire` records wrap the record as `{"stream": "observer", "type": "wire", "direction": "node", "record": {...}}`.

```bash
streamsim kind capture (<file> | --stdin) [--kinds-dir <dir>]...
```

- `<file>` — a regular file, read as a stream line by line and never held whole: a line is at most 1 MiB (plus its envelope), and the scan stops after 32 MiB, so a journal of any size works when its birth is early. A symbolic link or a named pipe is refused. A relative path is walked component by component and none may be a symbolic link (the canvas bridge passes `runs/<id>/journal.jsonl` relative to the workspace); for an absolute path only the last component is checked, because the system's own links above it cannot be told from a planted one. `--stdin` reads at most 1 MiB instead, as `kind validate --stdin` does, with `source` `<stdin>`. A file or `--stdin`, exactly one.
- `--kinds-dir <dir>` — repeatable; each captured kind is compared with the kind file of the same name (a file is found by the kind it declares; a name that differs only in case is a collision). Only the files of the captured kinds are **parsed**; a kind file that cannot be read degrades to an issue of its own kind, never to a failure of the capture. A directory that does not exist holds no kind file (every kind is `new`, aimed at a file of it) and is not created. Without the option there is no comparison and no `target`.

Only the observer stream is ever taken. A record of the director stream (`sealed_stream`) or of any other stream, or a wrapped record without a stream (`stream_not_observer`), is refused; a bare `NBIRTH` with no envelope is allowed. Records are read strictly: member names are matched exactly (`Stream` is not `stream`), a member named twice makes the line unreadable, and the stream is checked before a record is accepted.

The birth must be sound: it validates against the plant wire `NBIRTH` schema, `protocol_min` does not exceed `protocol_max`, the contract version and digest are the ones this binary speaks, the `kind_digest` of every kind is recomputed and equals the one announced (issue at `/kinds/<i>/kind_digest`, `keyword` `kind_digest_mismatch`) and no kind is announced twice, not even under another case (`kind_duplicate`, which the reference host refuses as `inventory_mismatch` too). Any violation fails with `kind_failed` and the issues in the shape `domain validate` uses; a source with no readable `NBIRTH` fails with the issue `no_birth`. A first birth that breaks the contract is refused, not skipped for a later one. Member names and values echoed in a message or pointer are cut at 128 characters, and a pointer escapes `~` and `/` (RFC 6901).

It prints `{environment, boards, bd_seq, births_seen, taken, later_births?, issues?, kinds}`:

```json
{
  "environment": "hardware",
  "boards": [{"board": "mega-01", "environment": "hardware"}],
  "bd_seq": 0,
  "births_seen": 2,
  "taken": {"line": 12, "bd_seq": 0},
  "later_births": [{"line": 840, "bd_seq": 1, "same_content": false}],
  "issues": [{"message": "the birth at line 840 (bd_seq 1) announces other kinds than the one taken at line 12", "keyword": "rebirth_differs"}],
  "kinds": [
    {
      "kind": "fan",
      "kind_digest": "sha256:...",
      "draft": {"kind": "fan", "metrics": [], "commands": []},
      "claims": ["/commands/0/risk_class", "/commands/0/safe_stop", "/commands/0/args_schema/properties/duty/maximum"],
      "issues": [{"pointer": "/metrics/0/eu_range", "message": "min 5 is not below max 5", "keyword": "kind_range_empty", "member": "eu_range"}],
      "comparison": {"status": "differs", "target": "kinds/fan.kind.json", "file": "kinds/fan.kind.json", "file_digest": "sha256:...", "differs_at": "/metrics/0/unit"}
    }
  ]
}
```

- `environment` is what the node **claims**: `simulated`, `hardware` or `mixed`, like everything else in a birth; the binary does not verify it. `boards` lists every board with the environment it claims (`simulated` or `hardware`), which is the origin of the devices born on it. The two must agree (a `simulated` node has only simulated boards, a `hardware` node only hardware boards, a `mixed` node at least one of each); a birth that contradicts itself fails with `kind_failed` and an issue of keyword `environment_mismatch` at the member that breaks the rule.
- `births_seen`, `taken`, `later_births` — provenance. The first birth found is the one captured (a rebirth later in a journal does not replace it); `taken` says where it stood (`line` counts lines of the source from 1), `later_births` lists up to 64 later births with whether their kinds equal the taken one's. The capture-level `issues` hold `rebirth_differs` (a later birth announces other kinds), `birth_line_unreadable` (a line that names a birth but cannot be read, with its line number; it is skipped, never silently), `capability_digest_mismatch` (the board's `capability_digest` does not hold; the reference host does not check it, so this is a capture-only check and the kinds are captured all the same) and `scan_truncated` (`births_seen` is a lower bound). No option takes the last birth.
- `draft` — the `NBIRTH.kinds` entry without `kind_digest`: the content of a kind file, which must not carry that member. `kind_digest` is the digest the node announced, equal to what `kind validate` reports for the file once saved.
- `claims` — pointers into the draft of the members that came from the board and that nothing verified: `risk_class` and `safe_stop` of each command and every bound of an `args_schema` (`minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `multipleOf`, `minLength`, `maxLength`, `minItems`, `maxItems`, `minProperties`, `maxProperties`). A reviewer reads these before saving.
- `issues` (of a kind) — present only when the draft still breaks the kind file rules (for example a range with `min` not below `max`), with the `keyword` and `member` `kind validate` would report, or when the comparison found `kind_name_collision` or `kind_file_unreadable`.
- `comparison` — `status` is `new` (no kind file of that name), `same` (equal `kind_digest`), `differs` or `unreadable` (the file of that name cannot be read, or two files define the kind). `target` is where the draft would be saved, derived by one rule: an existing file's path for `same`, `differs` and `unreadable`; for `new`, `<first --kinds-dir>/<stem>.kind.json` where the stem is the kind name with `:` made `_` (cut so the file name fits a workspace path segment); on a collision, the colliding file. `file` and `file_digest` name the existing file, and `differs_at` is the pointer of the first member where it and the draft differ, found by the comparison `implements` uses. A kind whose name differs from a kind file's only in case (`Pump` and `pump.kind.json`) is `differs` at `/kind` with `collides_with` the file and a `kind_name_collision` issue, never `new`. Names are matched with their case folded; the schema's name pattern is ASCII, so a Unicode look-alike or a name with a space is refused as a birth and needs no normalisation.

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
- `component list --unit <u> [--min <n> --max <n>]` — only the instruments that measure in `u` or in a unit convertible to it (the unit conversion `kind twin` uses), ordered fitting instruments first, then the exact unit before a convertible one, then by id and version; each entry carries `match` (`exact` or `convertible`) beside its `unit` and `range`, and `fit: {ok, code?, member?}`: whether a twin can read a metric of that unit through it. `--min` and `--max` are the metric's `eu_range` in `u` (both or neither, `min` below `max`, else a usage failure); the instrument's own range is compared with it after conversion, with the comparison `kind twin` uses, so `ok: true` is exactly what `kind twin` accepts. `code` is `unit_incompatible` (`member` `unit`), `range_unstated` (`eu_range`: no `--min`/`--max` given, and a twin needs the metric's range), `range_narrower` (`eu_range`) or `reading_not_number` (`data_type`). A unit no family knows lists nothing. `--unit`, `--min` and `--max` belong to `list` only.
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

### `canvas`

Serves the visual canvas and its bridge on a loopback address until interrupted. The canvas is a client of the binary, equal to the command line: it computes nothing the binary owns, and the bridge maps each canvas action to one binary command ([physical-device actions](../architecture/physical-devices.md#bridge-actions)). The built canvas is not embedded; build it first (`npm ci --ignore-scripts && npm run build` in `canvas/`, which writes `canvas/dist`).

```text
streamsim canvas --workspace <dir> --ui-dir <dir> [--addr 127.0.0.1:0] [--adapters-dir adapters] [--components-dir <dir>]...
```

- `--workspace` — required. The folder the bridge serves: `domains/`, `compositions/`, `scenarios/`, `layouts/`, `kinds/`, `twins/`, `components/`, `settings/` and `runs/`. Keep it outside the repository when it holds customer data.
- `--ui-dir` — required. The built canvas, for example `canvas/dist`.
- `--addr` — the listen address; a non-loopback address is refused. The default `127.0.0.1:0` picks a free port.
- `--adapters-dir`, `--components-dir` — as for the other commands.

The address and a one-time pairing code are printed on **stderr**; stdout carries no document. A browser pairs once with the code and then holds a session cookie. A connected session started from the canvas runs `plantwire connect` as a child process, one at a time, and the bridge stops it when it closes.

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
when the world reaches their time, verifies them against later observations, and prints its report. While a session is live it
sends `STATE` with an increasing `heartbeat` every `heartbeat_interval_ms` (1000) of wall time, so a hardware node does not
declare it lost.

```text
streamsim plantwire host --socket <path> [--script <commands.json>] [--inventory <devices.json>] [--environment simulated|hardware|mixed]
                         [--policy <digest>] [--epoch <n>] [--delay-ms <n>]
streamsim plantwire node --host-socket <path> --domain <id> --out <dir> [--domains-dir domains] [--adapter <id>] [--adapters-dir adapters]
                         [--session-id <id>] [--seed <n>] [--slice-s 60] [--duration 3600] [--barrier-cap-ms 30000]
                         [--policy <digest>]... [--kinds-dir <dir>]...
```

`--environment` (default `simulated`) is what the host accepts, one of `simulated`, `hardware` or `mixed`; any other value fails before the host listens.
A host set to `simulated` or `hardware` accepts only a node of exactly that environment; a `mixed` host accepts all three. Every board of a
birth names its environment, and a node whose environment contradicts its boards (`simulated` with a hardware board, `hardware` with a
simulated board, `mixed` without one of each) is refused even by a `mixed` host. The refusal is `environment_mismatch`, and the report
says which rule broke in `reason`. `plantwire node` has no environment option: a simulated world announces `simulated` for itself and every one of its boards,
and cannot be made to announce `hardware` or `mixed`.

The socket's directory must be owner-only (it is created `0700` when missing); a stale socket left by a crash is replaced,
a live host or a regular file at the path is not. Both commands stop on interrupt or `SIGTERM`, and the node still writes its
run folder. The script is a JSON array of
`{id?, at_ns, device, command, args, verify?: {metric, equals, tolerance, within_ns}}`, issued in `at_ns` order; `id` (a
unique wire-safe name, default `ref-<n>` by position) becomes the command's `command_id`; a command to a
device that is quarantined or never born is not sent and is reported `denied`. A command is verified only by a sample that
arrives after its result. The host sends each verification decision once in its next `HSYNC` as `command_evidence`
(`observed` with the value, or `not_observed` when the window passed), only for a good, live sample of the same boot. The inventory is a JSON object of device to kind; without one, every device is accepted. A node
must send its birth within 10 s; after a refusal the host closes the connection. The host carries its `--policy` digest (`sha256:` and 64 lower-case hexadecimal digits) in every command it issues; it signs nothing, and a host with a `--script` and no well-formed `--policy` is a usage error (the reserved all-zero stop digest is refused too). The node accepts a command only when the command's digest is on the node's `--policy` list (at most 8 distinct digests; a malformed, repeated or reserved digest is refused as bad options before the node connects). The commands that skip freshness, epoch and policy checks are the
safe stops the domain flags (`safe_stop: true` on an effector); a domain that implements a kind file is checked against it
(`--kinds-dir`, as `domain validate`) before the node connects, so the safe stops and the `kind_digest` it announces are the kind file's. `--barrier-cap-ms` (default 30000) bounds the wall time the node waits for each
`HSYNC`; `--delay-ms` makes the reference host slow. **Give the reference host a `--delay-ms` (for example 250) for any session longer than a few seconds:** with none, the world free-runs and writes about 75 MB of run folder per second (a 6 second run wrote about 450 MB), and `plantwire twin-report` then refuses the run because its `journal.jsonl` is over 64 MiB. A `--script` command's `at_ns` is **world** time, which starts at 1767225600000000000 (2026-01-01T00:00:00Z) by default, not wall time; its `id` is the command's `command_id`; `--script` needs a well-formed `--policy`, else the command is a usage error (exit 2, `usage_invalid`). The host prints its report and exits 0 when its node ends, including a node stopped while the host is waiting out a `--delay-ms`, and when it is interrupted (`SIGINT`, `SIGTERM`) before any node arrived, then with an empty report.

### `plantwire relay`

The simulator between an edge gateway and the real host: the relay is the host of one gateway and the one
node the real host sees, alone (hardware only) or with a simulated world beside the gateway's boards (a mixed session). It creates the gateway's owner-only socket, accepts one gateway at a time, and once a gateway has been
admitted dials the host's owner-only socket and presents the gateway's birth as its own. Records go up and commands go down
with their content unchanged (an `HSYNC` loses `command_evidence`, which the relay consumes); the relay owns only the envelope. It stops on interrupt or `SIGTERM` and prints what it carried.

```text
streamsim plantwire relay --gateway-socket <path> --host-socket <path> [--out <dir>]
                          [--domain <id> [--adapter <id>] [--seed N] [--kinds-dir <dir>] [--domains-dir <dir>] [--adapters-dir <dir>]
                           [--slice-s N] [--duration N] [--barrier-cap-ms N] [--policy <digest>] [--session-id <id>]]
```

- `--gateway-socket` — the socket the relay creates and the gateway dials (required). Its directory must be owner-only (it is
  created `0700` when missing); a stale socket is replaced, a live one or a regular file at the path is refused.
- `--host-socket` — the real host's owner-only socket (required). A socket that group or others can reach is refused and ends
  the relay; a host that is not there yet is retried every 250 ms.
- `--out` — a folder for the journal of every record on both links (`journal.jsonl`, observer `wire` session records with the
  directions `gateway`, `node`, `host` and `relay`; a carried record is two lines, as received and as sent) and
  `field-trace.jsonl`, every record the gateway sent that decoded and belonged to an admitted session, in the order the relay
  received it (a refused birth, an invalid record or a record of another session is journalled but not in the field trace; see
  [artifacts](artifacts.md)). When the field trace cannot take a record (it is full, or a line is too long) the relay stops
  recording, disconnects the gateway, refuses the next one, keeps the host session and reports `trace_incomplete`; a run
  folder then records no `field_trace_digest`, so `replay-trace` refuses its trace. A hardware-only relay writes no `run.json`:
  its field trace has no digest to cover it.
  Hardware has no truth, so the director stream holds only its closing record. Without it nothing is written.
- `--domain` — makes the relay **mixed** (and then `--out` is required): the simulated world of that domain runs in the same
  node, its boards announced `simulated` beside the gateway's `hardware` boards in one `NBIRTH` whose `environment` is `mixed`
  (the host must accept `mixed`). The options after it are those of `plantwire node` (`--seed`, `--kinds-dir` for the
  `implements` check, `--slice-s`, `--duration`, `--barrier-cap-ms`, `--policy`, `--session-id`). The gateway is refused
  `inventory_mismatch`, with the rule in `reason`, when it announces a device or a board the world also has or a kind of the
  same name with another `kind_digest` (the same kind with the same digest on both sides is fine: a twin and a board of one
  model). A command is routed by the board that announced its device: a hardware device's goes to the gateway untouched, a
  simulated device's to the world, and one for a device on no board is refused `unknown_device`; nothing the world does has a
  path to the gateway link. The gateway's records are never perturbed; the world keeps its own faults. `--out` then holds a run
  folder: `run.json` (with `field_trace_digest`), the world's trace, ledger and history, both journals and the field trace.
  The world is advanced slice by slice in virtual time as `plantwire node` does, as fast as the host answers each barrier; it is
  not paced to the wall clock (deferred D-71); a session plan with `pace: "wall"` run by `plantwire connect` is.

The gateway is checked as the reference host checks a node that is set to `hardware`: a gateway that announces `simulated` or
`mixed`, speaks another contract, announces a `kind_digest` that does not hold or a kind twice is refused with the refusal in a
`STATE`. The host's `STATE` (heartbeat, `authority_epoch`) reaches the gateway unchanged, a host that is lost is reported to the
gateway as offline with its session kept, a gateway that is lost has each of its devices reported `DDEATH link_down` to the host,
a gateway that stays connected but sends no admitted record for 30 s while its host is present is treated as lost the same way
(the watchdog is suspended while the relay has no host and starts afresh when a host's first `STATE` reaches the gateway), and a host that
refuses the session ends the relay with `plantwire_failed`. The relay never originates or answers a command and
has no fault, perturbation or script option; `run`, `replay`, `verify`, `scenario` and `plantwire node` have no gateway option.
The document is `{gateway_sessions, host_sessions, up, down, dropped, refused?, reason?, host_refused?}` (and, for a mixed relay,
`to_simulated?`, `from_simulated?`, `unknown_device?`, `simulated_lost?`), plus `forged?` and `trace_incomplete?` for any relay,
where `dropped` counts the records received and not carried (invalid, oversize, repeated, of another session, of a type that
does not travel that way, naming a device of the other origin, or with no host to take them). `forged` counts, among those, the
answers (`DRECEIPT`, `DRESULT`) of a side to a command routed to the other side in a mixed relay, and the device births on a
board other than the one that announced the device or of a kind its `NBIRTH` never announced. A command for a device of a gateway
that is not connected is refused `not_ready`; a host's rebirth request in a mixed relay reaches the gateway and the world. When
the simulated node ends before the relay (a mailbox overflow is a failure), its devices are reported offline to the host at once
and `simulated_lost` is 1.

### `plantwire connect`

A connected session: the plant of a session plan (`session-plan-v0.1`) linked to a plant wire host and, optionally, a
gateway, until it is interrupted. The plant does not change; the plan names the instances that are real boards behind the
gateway, and every other instance is simulated.

```text
streamsim plantwire connect (--plan <file> | --stdin) [--check] [--out <dir>] [--domains-dir <dir>] [--adapters-dir <dir>]
                            [--kinds-dir <dir>]... [--slice-s 1] [--barrier-cap-ms 30000]
                            [--control-socket <path>]
```

```json
{"schema_version": "0.1", "domain": "arduino-thermal", "seed": 1, "host_socket": "/abs/host.sock",
 "gateway_socket": "/abs/gateway.sock", "real": [{"instance": "zone-01", "board": "uno-07", "device": "zone-01"}],
 "policies": ["sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"]}
```

- `policies` (optional, 1 to 8 digests) — the policy digests the session's simulated devices accept in a command's
  `policy_digest`. The plan is the only source of the list for a connected session; `plantwire connect` has no `--policy`
  option. Absent, none is accepted and every command to a simulated device except a safe stop is refused
  `policy_not_allowed`. A digest is `sha256:` and 64 lower-case hexadecimal digits; a repeat (`policy_listed_twice`, at the
  second occurrence) and the reserved all-zero digest of the relay's own stops (`policy_reserved`) are refused. Real boards
  are checked by the gateway's own list. The list is recorded in the `session` record of the journal, not in the run's
  identity.

- `disturbances` (optional, at most 64) — faults the session injects into its simulated world, each
  `{at_s, instance, fault, duration_s?, severity?}`: the fault starts `at_s` seconds after the world starts (paced or not) and,
  with `duration_s`, is cleared that much later; `severity` defaults to the domain's magnitude. Every disturbance is injected
  when the session starts, through the same path as `fault.inject` and `fault.clear`, so the run's command log and its sealed
  journal hold it and `replay` and `verify` reproduce it without the plan. The schedule is never sent to the host, never echoed
  by `--check` and never written to the observer journal: the host sees only its effect on the readings. A plan whose
  disturbance names an unknown instance (`unknown_instance`), a fault the instance's kind does not declare (`unknown_fault`),
  an instance bound to a real board (`disturbance_on_real_instance`) or a duration for a pulse fault, whose length the domain
  fixes (`duration_with_pulse_fault`), is refused with the pointer `/disturbances/N/...`; a negative severity or an `at_s` or
  `duration_s` beyond 30 days (the longest run) breaks the plan's schema.

- `pace` (optional, only `"wall"`) — paces the session to the wall clock. The world starts at the wall time the session
  starts (floored to the millisecond; for a mixed session, when the relay opens) instead of 2026-01-01, and that start is
  recorded in `run.json` (`world_config.start_time_ns`), so `replay` and `verify` reproduce the run exactly as an unpaced one.
  Each slice is released when the wall clock reaches the world time it ends at, on an absolute schedule (no drift adds up), so
  a 10 minute session takes 10 minutes and world time is UTC; a slice that is already late is released at once, so after a
  stall the world catches up at the host's speed. The barrier stays a cap (`--barrier-cap-ms`), three missed barriers still
  end the world, and an operator control waits at most one slice. The `NBIRTH` announces `node_metrics.clock_mode` `wall`
  (the merged birth of a mixed session too). A paced session longer than 30 days ends with the run span limit. Absent, the
  world starts at 2026-01-01 and slices follow one another as fast as the host answers, as before.

- `--check` — reads the plan against its domain and kind files and starts nothing. The document is
  `{domain, environment, instances: [{instance, origin, kind, kind_digest?, board?, device?}], implements_checked, policies}`
  (`policies` is the list in effect, sorted, `[]` when the plan names none). `environment` is `simulated` without a gateway,
  `hardware` when every instance is bound to a real device (the world then announces no board), and `mixed` otherwise; the
  relay's merged `NBIRTH` claims the environment of the boards it actually announces, so a strict `hardware` host accepts an
  all-real session. Each
  instance is `real` (with the board and device the gateway must announce and the digest of the kind file it must announce
  them with) or `simulated`. A plan that does not hold is `plantwire_failed` with one issue per broken rule, each pointing
  into the plan: `unknown_instance`, `no_kind_file` (only an instance whose kind implements a kind file can be a real board),
  `instance_bound_twice`, `device_bound_twice`, `gateway_required`, `same_socket` (compared after cleaning the paths),
  `socket_not_absolute`, `socket_parent_component` (a `..` component), `socket_control_character`,
  `policy_listed_twice` and `policy_reserved`; a list that is empty, longer than 8 or holds a malformed digest breaks the
  plan's schema (`minItems`, `maxItems`, `pattern`). The sockets are used
  as cleaned paths.
- Without `--check`, `--out` is required, and **it must be a new folder for every session**: a folder that already holds any run file (`run.json`, a journal, a trace, an interrupt log, a verdict or a label) is refused with `plantwire_failed` before anything is written, so an earlier session is never overwritten or mixed with a new one. A folder that does not exist yet, or holds no run file, is accepted. Name each attempt (`m5-session-1`, `m5-session-2`) and keep the failed ones; the canvas picks a unique `c-<8 hex>` name. The host's socket must exist: without it the command fails `file_not_found` after it has created the run folder (`journal.jsonl`, `ledger.jsonl`, `run.json`) and any missing parent directories (mode `0700`), so a retry needs a new folder. With no `gateway_socket` the world is a plant wire node to the host; with one it is
  the mixed relay (`plantwire relay --domain`) and the gateway is checked against the plan at its handshake: the bound board
  must be announced (`board_missing`), hold the device (`device_missing`), and announce the instance's kind
  (`kind_missing`) with the kind file's `kind_digest` (`kind_mismatch`), or the gateway is refused `inventory_mismatch`. A
  real instance is not announced by the world (it still simulates it, as the twin); its device is the board's.
- `--control-socket <path>` — creates an owner-only unix socket (owner-only directory; a live socket or a regular file at the
  path is refused) on which the session takes the operator's controls at the node, until it ends: **hold** (every command of the
  host in scope is refused `local_control`), **approve each command** (a command in scope is answered `queued` and waits until the
  operator approves it, denies it, or its deadline passes: `expired`) and **emergency stop** (the safe stop of every device in
  scope, then a hold latched until the operator releases it), each for all devices, a board or a device. The node only refuses or
  stops: it never originates an ordinary command, and a device's declared safe-stop command is never held or queued when it carries no arguments (a safe stop with arguments is held like any command, D-100). One request per line
  (`node-control-v0.1`: `{"op": "hold" | "release" | "approve_mode" | "approve" | "deny" | "estop" | "state", "scope":
  {"kind": "all" | "board" | "device", "id"}, "enabled", "command_id"}`), one answer per line (`{"result": "applied" | "refused",
  "reason"?, "op", "command_id"?, "stopped"?, "skipped"?, "state": {"holds", "approve", "pending", "devices"}}`); a request the
  node cannot take within 10 s is answered `refused` with the reason `node_busy`. `approve` and `deny` may carry the
  `content_digest` of the waiting command (shown in the state's `pending` entry with its `args`) and are refused `unknown_command`
  when it differs; a command id that is already waiting is refused `duplicate`; an emergency stop lists the devices it could not
  order in `unreachable`. A control line is at most 4096 bytes and a connection idle for 60 s is closed. Controls are applied at a barrier (a node of
  the world) or on a 250 ms tick (a relay), never in between. The run then holds `interrupts.jsonl` ([the interrupt
  log](artifacts.md#the-interrupt-log)). Without the flag nothing changes. A `release` must name the same scope as the `hold` or the stop it clears ("on exactly that scope"): a hold on a device is not cleared by releasing `all`. The `state` reply lists in `devices` only the devices that have a non-default mode, so it is `[]` on a quiet session. At most 64 scopes are set, at most 64 commands wait for approval (each with at most 8 repeats), and a node has at most 4096 devices. Example lines: `{"op":"release","scope":{"kind":"device","id":"fan-01"}}`, `{"op":"approve","command_id":"<id>"}`, `{"op":"deny","command_id":"<id>"}`. A client needs to talk to a Unix socket: `nc -U -w 2 <socket>` works with the macOS and OpenBSD netcat (GNU netcat has no `-U`); `socat - UNIX-CONNECT:<socket>` works everywhere.
- Every device born to the host is journalled with its origin (observer `origin` records: instance, device, board,
  environment, boot id) and each gateway handshake of a plan that binds an instance with its outcome (`binding` records:
  matched or refused, with the issues).
- The run records that it is a connected session: one observer `session` record right after `run.opened` holds the mode
  `connected` and the checked plan (the `--check` document), so a reopened or relisted run keeps its classification and its
  simulated instances. While the session runs, the binary journals `live` records, the plant wire projection of what the host
  saw: host and gateway link state, each device's health, board, boot id and reboot count, each reading with its value and a
  state (`good`, `stale` after three scan intervals without a sample, `disconnected` when its link or device is lost,
  `invalid` outside its declared type or `eu_range`, `untrusted` for any other quality), and the latest 128 commands with
  separate stages: sent, the board's receipt (`pending`, `accepted`, `refused` with its reason), the board's own report
  (`pending`, `succeeded`, `failed`, `unknown`) and the effect (`unknown` unless the host's `command_evidence` in `HSYNC`
  observed the expected value for the same command, device and boot after a successful report). At most one snapshot is
  written per barrier (no gateway) or per 250 ms relay tick, and one at every link change; `marks` lists what changed since
  the previous snapshot.
- An interrupt (`SIGINT`, `SIGTERM`) is the operator's Stop: the run ends and `--out` holds an ordinary run folder that
  `replay` and `verify` accept. The document is `{out, stopped, relay?}`. The canvas runs this command for its Connected
  session and stops it the same way.

### `plantwire replay-trace`

Plays the field trace of a run folder to a host, the way a gateway would, from the file.

```text
streamsim plantwire replay-trace --run <dir> --host-socket <path> [--session-id replay-1] [--barrier-cap-ms 30000]
```

A replay is not a board: it announces the node and **every board `simulated`**, with the recorded device ids, kinds, kind
digests and values, so a host set to `hardware` refuses it (`environment_mismatch`) and no consumer can take it for hardware.
It has no gateway option and no path to one; a command the host sends is refused (`not_ready` for a recorded device,
`unknown_device` otherwise) and its result is never reported. The recorded barriers are not replayed: the replay raises its own
after each stretch of records and waits for the host's `HSYNC`, so the stream the host sees is a pure function of the trace and
the host's answers (two replays of one trace are byte-identical). Records the relay would have dropped (invalid, repeated, before
the first `NBIRTH`) and the recorded receipts, results and deaths are skipped and counted. The folder's `run.json`, when there is
one, must record the digest of the trace (`field_trace_digest`); a trace that differs from it is refused before anything is
played. The document is `{records, skipped, barriers, commands_refused, field_trace_digest, session_id}`.

### `plantwire twin-report`

Compares the twin with the board: where the domain that implements a board's kinds and the readings the board reported part
ways.

```text
streamsim plantwire twin-report <run-dir> --domain <id> --kinds-dir <dir> [--domains-dir domains] [--tolerance device.metric=value] [--span-fraction 0.05]
```

The run folder must hold a field trace (a hardware or mixed session recorded by `plantwire connect` or `relay`) and the observer
`journal.jsonl`, where the commands the host sent are recorded. The command replays the twin in a fresh, noise-free world
driven by the commands the board **accepted** (the gateway's `DRECEIPT accepted`, no reported failure, and no interrupt refused
or held them; a command refused or never forwarded is listed with its reason and never applied), and compares the twin's
prediction at every recorded board reading with that reading, per device and per metric. A reading is judged only when its
quality is `good`, it is not historical and numeric, the twin has the device and the metric, a tolerance exists, and it is not
earlier than the first applied command; any other reading is `not_compared` with its reason, never agreement or divergence.
The tolerance is declared with `--tolerance` (the selector is a device id, tried first, or a kind name; an unknown selector or
a negative value is refused) or, failing that, `span_fraction` times the span of the metric's `eu_range` in its kind file
(0.05 unless `--span-fraction` says otherwise); with neither the metric is `cannot_judge`. The document is
[`twin-report-v0.1`](../../contracts/twin-report-v0.1.schema.json): the verdict (`agrees`, `diverges`, `cannot_judge`), a
summary, the commands with their status, per device and metric the tolerance, a summary, the reasons for readings not compared
and the series of the twin's predictions (at most 2000 points, truncation explicit), and each named divergence with its
interval, its worst point and the commands in force there. The report is a pure function of the run folder, the domain, the
kind files and the options, and exits 0 whatever the verdict. It measures the twin, not a consumer. A safe stop the node itself
ordered on an operator's emergency stop is applied to the twin like a host command and listed with `origin: node`; when the
board never answered it, or the twin cannot apply it, the verdict is `cannot_judge` (`node_stop_unconfirmed`,
`node_stop_not_applied`) rather than a divergence. The interrupt log is replayed before it is used; `run.evidence_bound`
says whether a `run.json` covers the files. Limits: the command reads `journal.jsonl` through a 64 MiB cap and refuses a larger one (`journal.jsonl is larger than 67108864 bytes`; a field trace and an interrupt log are capped at 64 MiB too, which is what a reference host with no `--delay-ms` produces within seconds), and refuses a field trace whose readings span more than 7 days.

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
