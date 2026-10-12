# CLI surface — output contract

- **Status**: accepted design for phase P1 of the simulation canvas goal
  (`docs/simulation-canvas/PLAN.md`); contract
  [`cli-output-v0.1`](../contracts/cli-output-v0.1.schema.json).
- **Why**: every canvas action has a CLI or MCP equivalent, and a coding agent or the canvas
  bridge must act on results and failures without parsing prose.

## 1. Streams and exit status

| Stream | Content |
|---|---|
| stdout | Exactly **one JSON document** per invocation of a non-exempt command, for success and for failure, followed by a newline. Indented with two spaces. |
| stderr | Human diagnostics only: the `flag` package's messages, usage text, and one `streamsim: <message>` line on failure. Never parsed by tools. |

| Exit | Meaning |
|---|---|
| 0 | Success; stdout holds the command's result document |
| 1 | Failure; stdout holds an error document |
| 2 | Usage error (bad flag, missing or unknown verb or argument, invalid argument value); stdout holds an error document with code `usage_invalid` |

A success document never has a top-level `error` key; an error document has only that key.
Tools may branch on the exit status or on the presence of `error`.

### Exempt invocations

| Invocation | Behaviour |
|---|---|
| `streamsim help`, `-h`, `--help`, `<command> -h` | Usage on stderr, nothing on stdout, exit 0 |
| `streamsim mcp …` | stdout carries the MCP protocol, so the CLI never writes a document there. Failures go to stderr; exit 1 (or 2 for usage) |
| `streamsim device serve …` while serving | Runs until stopped and writes no success document. A failure before or during serving writes an error document and exits 1 |

## 2. Error document

```json
{
  "error": {
    "code": "domain_invalid",
    "message": "domain: domains/x.domain.json fails domain-spec-v0.1 validation: …",
    "source": "domains/x.domain.json",
    "issues": [ { "pointer": "/channels/2/unit", "message": "missing required property \"unit\"" } ],
    "result": { }
  }
}
```

| Field | Presence | Meaning |
|---|---|---|
| `code` | Always | Stable code from §3 |
| `message` | Always | Human-readable cause on one line, without repeated `streamsim:` prefixes; for a contract violation the issues carry the details |
| `source` | When a file is at fault | The path as given on the command line |
| `issues` | Validation failures | One entry per violation; `pointer` (RFC 6901) is present when the location is known. Contract-schema violations always carry one; until domain-spec v0.2 (phase P5a) cross-reference checks in domains may omit it |
| `result` | `replay_mismatch`, `adapter_verify_failed` | The command's result document describing the failure (e.g. the replay report with `first_divergence`) |

## 3. Error codes

Codes shared with the MCP taxonomy (`MCP_SURFACE.md` §7) mean the same thing on both surfaces.

| Code | Exit | Raised when |
|---|---|---|
| `usage_invalid` | 2 | A flag fails to parse; `report --format markdown` without `--out`; a required verb, argument or flag is missing; a verb or subcommand is unknown; a command or verb is given more positional arguments than it takes; an argument value is malformed or out of range (`--sink`, `--duration`, `--n`, `--window`, `--threshold`); the scenario the `run` flags describe fails its checks (each issue names the flag occurrence and carries no pointer); a flag a `--scenario` file defines is given beside it |
| `file_not_found` | 1 | A path the command reads, writes into, or binds does not exist (an input file, an output's directory, a socket's directory, a file an adapter refers to) |
| `domain_not_found` | 1 | A domain id is not in the catalog |
| `adapter_not_found` | 1 | An adapter id is not in the adapters directory. The MCP director reports the same lookup as `adapter_invalid`; the CLI separates an unknown id from an invalid document |
| `domain_invalid` | 1 | A domain document fails `domain-spec` or its cross-reference checks |
| `adapter_invalid` | 1 | An adapter document fails `output-adapter` |
| `scenario_invalid` | 1 | A scenario file fails `scenario-v0.1`, holds a value a scenario cannot hold, or fails a check against the domain and world (`SCENARIO_FILE.md` §3); `scenario from-run` would write a scenario that fails `scenario-v0.1` or that check (`source` is then the artifact) |
| `artifact_invalid` | 1 | A run artifact, consumer verdict or ground-truth label is not valid JSON or fails its contract; a delivery ledger or world-state history line does not decode as its record type (unknown fields refused) |
| `replay_mismatch` | 1 | `replay`, `verify` or `score` cannot reproduce the artifact: the trace digest differs (the replay report is the `result`), or the artifact's domain, adapter or world digest no longer matches the inputs (no `result`: nothing was replayed) |
| `adapter_verify_failed` | 1 | `adapter verify` found a schema or golden divergence, or the adapter declares nothing to check |
| `io_failed` | 1 | Reading, writing or binding a path failed for a reason other than absence |
| `<command>_failed` | 1 | Any other failure of that command: `catalog_failed`, `domain_failed`, `adapter_failed`, `run_failed`, `replay_failed`, `verify_failed`, `suite_failed`, `score_failed`, `manifest_failed`, `refconsumer_failed`, `device_failed`, `scenario_failed`, `report_failed` |

**Classification precedence** (first match wins): usage, replay or verify mismatch, contract
violation (by contract: domain → `domain_invalid`, adapter → `adapter_invalid`, scenario →
`scenario_invalid`, others → `artifact_invalid`), unknown domain or adapter id, missing file, other file-system failure,
the command's own `_failed` code. Classification uses typed errors and sentinels from the
modules (`errors.As`/`errors.Is`), never message text.

## 4. Success documents

| Command | Document |
|---|---|
| `catalog list` | `{domains: [{id, title, version, axes, stresses}]}` |
| `catalog describe <id>` | `{digest, spec}` |
| `catalog coverage` | `{by_axis, thin}` |
| `domain validate <path>` | `{valid: true, id, digest, channels, faults, effectors}` |
| `adapter list` | `{adapters: [{id, version, encoding, title}]}` |
| `adapter verify <path>` | `{adapter, schema_checked, schema_ok, golden_checked, golden_match, records}` |
| `run` | `{run_id, trace_digest, reproducible, emitted, ledger}` |
| `replay <run.json>` | `{matches, version_match, got, want, first_divergence, detail}` (exit 0 whatever the outcome) |
| `verify <run.json>` | The replay document on a match; `replay_mismatch` with the document as `result` otherwise |
| `suite` | `{suite, scenarios, trivial_excluded, terminal_state}` |
| `scenario from-run <run.json> --out <file>` | `{scenario_written, commands, omitted, refused}` |
| `score` | The scorecard (`scorecard-bundle`), see `internal/score` |
| `report <run-dir>` | The report (`run-report-v0.1`, [design](RUN_REPORT.md)); with `--out <file>`, `{report_written, format}` and the file holds the report in the chosen format. `--format markdown` requires `--out`: stdout stays one JSON document, so a Markdown rendering is only ever a file |
| `refconsumer` | `{verdict_written, detections}` |
| `manifest` | `{manifest_written, signed, sim}` |

The exact shapes are the `$defs` of `cli-output-v0.1`; tests validate every command's real
output against them.

## 5. Deliberate changes from the previous behaviour

Recorded in the behaviour pin and CHANGELOG:

1. Failures now print an error document on stdout (previously nothing, or for `verify` the bare
   replay document).
2. Argument errors that are not flag-parse errors (missing verb, unknown verb, missing
   required argument, malformed argument value) exit **2** with `usage_invalid` (previously 1).
3. stderr text is unchanged.
