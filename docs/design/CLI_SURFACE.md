# CLI surface — output contract

- **Status**: accepted design for phase P1 of the simulation canvas goal
  (`governance/CANVAS_PLAN.md`); contract
  [`cli-output-v0.1`](../../contracts/cli-output-v0.1.schema.json).
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
    "message": "domain: domains/x.domain.json fails domain-spec-v0.2 validation: …",
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
| `issues` | Validation failures | One entry per violation; `pointer` (RFC 6901) is present when the location is known. Contract-schema violations always carry one, a "missing required property" or "additional property" issue pointing at the member itself; domain cross-reference issues carry one since domain-spec v0.2 (phase P5a). A document that is not JSON, adapter cross-reference checks and run-flag issues have no location and omit it |
| `result` | `replay_mismatch`, `adapter_verify_failed`, `domain_lint` | The command's result document describing the failure (e.g. the replay report with `first_divergence`) |

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
| `replay_mismatch` | 1 | `replay`, `verify` or `score` cannot reproduce the artifact: the trace digest differs (the replay report is the `result`), or the artifact's own recorded digests do not hold: the embedded domain or adapter spec does not hash to its recorded digest, the recorded specs replay to a different world digest, or `trace.jsonl`, `ledger.jsonl` or `world_state_history.jsonl` beside the artifact differs from the digest `run.json` recorded (the replay report is then the `result`, with `matches: false` and a `detail` naming the reason, so a verdict is told from a failure to run). An artifact that records no specs and meets a changed domain on disk, a missing `run.json` and unreadable files stay errors with no `result` (`ErrDigestMismatch` still names the first) |
| `domain_lint` | 1 | `domain lint` found at least one error-severity finding in the nominal day; the lint report is the `result` |
| `adapter_verify_failed` | 1 | `adapter verify` found a schema or golden divergence, or the adapter declares nothing to check |
| `io_failed` | 1 | Reading, writing or binding a path failed for a reason other than absence |
| `<command>_failed` | 1 | Any other failure of that command: `catalog_failed`, `domain_failed`, `adapter_failed`, `run_failed`, `replay_failed`, `verify_failed`, `suite_failed`, `score_failed`, `manifest_failed`, `refconsumer_failed`, `device_failed`, `scenario_failed`, `report_failed`, `plantwire_failed` |

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
| `domain lint <path> [--hours H] [--seed S]` | `{domain, digest, seed, hours, samples, findings, errors, warnings}`; a nominal-day plausibility check ([DOMAIN_SPEC_V0_2](DOMAIN_SPEC_V0_2.md) §27); exit 1 with `domain_lint` when a finding is an error |
| `domain format <path>` | `{formatted, changed, digest}`; rewrites the file in canonical formatting ([DOMAIN_SPEC_V0_2](DOMAIN_SPEC_V0_2.md) §6) |
| `adapter list` | `{adapters: [{id, version, encoding, title}]}` |
| `adapter verify <path>` | `{adapter, schema_checked, schema_ok, golden_checked, golden_match, records}` |
| `run` | `{run_id, trace_digest, reproducible, emitted, ledger}` |
| `replay <run.json>` | `{matches, version_match, got, want, first_divergence, detail}` (exit 0 whatever the outcome); replays the artifact's recorded domain and adapter specs, and adds `domain_changed: true` and `current_domain_digest` when the domains directory now holds a different digest for the run's domain (information, not failure); the keys are omitted when the current domain cannot be read (no catalog in the directory, or the run's domain is no longer in it), so an absent `domain_changed` means unknown, not unchanged |
| `verify <run.json>` | The replay document on a match; `replay_mismatch` with the document as `result` otherwise (a broken recorded digest is `matches: false` with `got` empty, `want` the recorded trace digest and a `detail` such as "the recorded domain spec does not match its digest"; a changed `trace.jsonl` is "trace.jsonl differs from the digest run.json recorded") |
| `suite` | `{suite, scenarios, trivial_excluded, terminal_state}` |
| `scenario from-run <run.json> --out <file>` | `{scenario_written, commands, omitted, refused}`; the scenario is checked against the domain and adapter specs recorded in `run.json`, not the files on disk |
| `score` | The scorecard (`scorecard-bundle`), see `internal/score`; the artifact is replayed from its recorded domain and adapter specs, whatever the directories now hold |
| `report <run-dir>` | The report (`run-report-v0.1`, [design](RUN_REPORT.md)); with `--out <file>`, `{report_written, format}` and the file holds the report in the chosen format. `--format markdown` requires `--out`: stdout stays one JSON document, so a Markdown rendering is only ever a file |
| `refconsumer` | `{verdict_written, detections}` |
| `manifest` | `{manifest_written, signed, sim}` |
| `plantwire node` | `{run_id, out}`: the run the node served; its artifact and journal (with `wire` and `pacing` records) are in `out` |
| `plantwire host` | The reference host's report: `{refused?, quarantined, rebirths, commands: {id: {receipts, result?, verified?}}}` |

The exact shapes are the `$defs` of `cli-output-v0.1`; tests validate every command's real
output against them.

## 5. Deliberate changes from the previous behaviour

Recorded in the behaviour pin and CHANGELOG:

1. Failures now print an error document on stdout (previously nothing, or for `verify` the bare
   replay document).
2. Argument errors that are not flag-parse errors (missing verb, unknown verb, missing
   required argument, malformed argument value) exit **2** with `usage_invalid` (previously 1).
3. stderr text is unchanged.
