# Run report

- **Status**: implemented in phase P4 of the simulation canvas goal
  (`governance/CANVAS_PLAN.md`); contract
  [`run-report-v0.1`](../../contracts/run-report-v0.1.schema.json); module `internal/report`
  (a core with a `files` edge); command `streamsim report`; evidence
  `docs/simulation-canvas/evidence/P4.md`.
- **Why**: the evidence of a run is spread over seven files. A person, a coding agent or the
  canvas report view needs one document that says what the run was, whether its files still agree
  with what the run recorded, what was done to it and how it was scored, without re-deriving any
  of that from the raw files.

## 1. A pure function of the run folder

`report(run folder) -> bytes`. Same folder, same bytes. The report reads no wall clock, no
environment, no catalog and no network; every list has a defined order; every map is written in
sorted key order (struct fields in the order of the contract). It never replays through a
directory of domains: scoring uses the domain and adapter documents the artifact embeds.

Input files (all inside the run folder, none required except that `run.json` or `journal.jsonl`
must exist):

| File | Holds | Truth-bearing |
|---|---|---|
| `run.json` | Run artifact: identity, inputs, command log, digests, counts | Yes (command log, seed) |
| `trace.jsonl` | What the consumer was delivered | No |
| `journal.jsonl` | Observer stream | No |
| `sealed/journal.jsonl` | Director stream: commands, effector outcomes | Yes |
| `ledger.jsonl` | Delivery ledger | Yes |
| `world_state_history.jsonl` | True state | Yes (not read by the report) |
| `verdict.json` | What the consumer concluded | No |
| `label.json` | Sealed ground truth | Yes |

A file that is absent is absent from the report; a file that exists and does not parse is an
error that names it. A run folder is never modified.

## 2. Who may read a report

A report carries truth: the seed, injected faults, perturbations, the ledger's drop reasons and
the real outcome of every effector call. It is a director-side document, like `score`.

Truth is included only when the run has ended, that is when `run.json` exists (it is written at
`End`), or when the run is an Explore run (unblinded from creation). For an Evaluate run without
`run.json` (a run still open, or one that crashed) the report neither reads nor uses
`ledger.jsonl` or `sealed/`: it holds the identity from the observer journal and the `observed`
count only. `run.ended` is `false`, and there is no `truth` object, seed, ledger count, drop
reason or verification. A live director world writes its ledger while `run.json` is absent, so
the rule is on `run.json`, not on the ledger. The report of an open Evaluate run therefore cannot
be used to read ahead. For an open Explore run the seed comes from the director `run.opened`
record.

A file belongs to the run of the journals or it is not used. `run.json` must carry the `run_id`,
`sim_version`, `domain` and `adapter` of the observer `run.opened`; `sealed/journal.jsonl` must carry
the `run_id` of that run (and the seed of `run.json`). A folder holding a file of another run reports
`truth_unavailable` with the reason (it names the fields that differ, never their values), shows no
`truth` object or seed, is not scored, and takes `run.json` (or the sealed journal and the ledger)
out of the report as if absent.

A run that crashed mid-write may end a journal or the ledger on a line without its newline. The report
drops that last line, lists the file in `torn_tail` and builds the rest; a damaged line anywhere
else still fails the report. The ledger rows of such a line are not counted.

## 3. Contents

Contract: [`run-report-v0.1`](../../contracts/run-report-v0.1.schema.json).

| Section | Content | Source |
|---|---|---|
| `schema_version` | `0.1` | constant |
| `run` | `run_id`, `mode` (`evaluate` or `explore`), `complete`, `incomplete_reason`, `ended`, `unblinded`, `sim_version`, `domain`, `adapter` (`id`, `version`, `digest`), `seed` (only with `truth`), `sink`, `time_mode`, `start_time_ns`, `end_time_ns`, `entity_count` | `run.json`; the observer `run.opened` and last `clock.advanced` when `run.json` is absent |
| `digests` | `world`, `trace`, `journal`, `sealed_journal` as the run recorded them; a digest the run did not record is absent | `run.json` |
| `verification` | `status` and one check per recorded digest | §4 |
| `counts` | `emitted` (the run's own count), `observed` (observer observations), `ledger_rows`, `delivered`, `dropped`, `delivered_by_reason`, `dropped_by_reason` | `run.json`, observer journal, `ledger.jsonl` |
| `truth` | `faults` (one interval per fault, from the director stream's `fault.truth` record), `perturbations`, `environment`, `effector_calls` (from the command log in command order) | §5 |
| `consumer` | `name`, `version`, `arm`, `config_digest` | `verdict.json` |
| `scoring` | `status` (`scored`, `not_scored`, `failed`), `reason`, `scorecard` | §6 |

**Complete.** `run.complete` is true when `run.json` exists and does not say `incomplete`. A run
without `run.json` ended abnormally or is still open; the report says so in
`incomplete_reason`. A run whose artifact says `incomplete` repeats the artifact's `error`.

**Consumer-side digests.** `consumer.config_digest` appears only when the consumer's verdict
recorded one (consumer-verdict-v0.1: opaque to the simulator, recorded so results stay
comparable). The report never computes, guesses or defaults a consumer-side digest; a verdict
without one yields a `consumer` section without the key.

**Counts.** `delivered` and `dropped` are ledger rows by `delivered`; the `*_by_reason` maps
count rows by the ledger's `delivery_reason`. A `duplicated` row is a delivery, so `ledger_rows`
can exceed `emitted`. `observed` counts the observer stream's `observation` records that are not adapter framing
(`framing` absent): the delivered lines, not the preamble and postamble.

## 4. Verification

`verification` answers one question: do the files in this folder still agree with the digests the
run recorded about itself? It is **not a replay**: `streamsim verify` re-runs the world and is the
proof; the report's check is cheap and needs no inputs.

| Check | Computed from | Recorded in `run.json` as |
|---|---|---|
| `trace` | SHA-256 of `trace.jsonl` | `expected_trace_digest` |
| `journal` | digest of the replayable lines of `journal.jsonl` | `journal_digest` |
| `sealed_journal` | digest of the replayable lines of `sealed/journal.jsonl` | `sealed_journal_digest` |
| `journal_file` | SHA-256 of `journal.jsonl`, session records included | `journal_file_digest` |
| `sealed_journal_file` | SHA-256 of `sealed/journal.jsonl`, session records included | `sealed_journal_file_digest` |

Each check has `status` `match`, `mismatch` or `missing_file` (a digest was recorded and the file
is absent), plus `recorded` and `computed` digests. A digest the run did not record is not
checked and has no entry. `verification.status` is `files_differ` when any check is not `match`,
`files_match` when there is at least one check and all match, and `not_checked` when there are no
checks (no `run.json`, or an artifact that recorded no digests).

## 5. Faults, perturbations, environment and effector calls

`perturbations`, `environment` and `effector_calls` come from the command log (`run.json`; for an Explore
run with no `run.json`, from the `command` records of the director stream), in command order, each entry
carrying the command's `seq` and `at_ns`. `faults` is not a command list: the world owns when a fault is active,
and a clear command alone does not say it (the clear carries no entity and an effective `at_ns` that may be
later than its command time; a pulse fault and an effector responder end faults with no command at all).

| List | Commands | Entry |
|---|---|---|
| `faults` | the director stream's last `fault.truth` record | an interval, below |
| `perturbations` | `perturb.apply`, `perturb.clear` | `action` (`apply`, `clear`), `perturbation`, `perturb_id`, `from_ns`, `until_ns`, `params` as present |
| `environment` | `env.inject` | `target`, `fault`, `params` |
| `effector_calls` | `effector.invoke` | `command_id`, `effector`, `entity_id`, `at_ns`, `args`, and `outcome` |

A `faults` entry is the interval the world gave the fault, in injection order: `fault_id` (the id a clear
command names), `entity_id`, `fault` (the kind's fault id), `injected_at_ns` (the world clock when the inject
command ran), `onset_ns`, `cleared_ns` (the instant the world ends the fault, `null` when it is still active at
the end of the run, including a clear scheduled after the end) and `cleared_by` (`command`: a `fault.clear` took
effect at its `at_ns`; `pulse_end`: the pulse ran its `duration_s`; `responder`: an effector responder with
`clear_fault` cleared it at its arrival; `null` while active). The earliest end wins: a clear command ahead of a
pulse's end shortens the pulse, one at or after it leaves `pulse_end`. The entries exist only in the truth
section: the director stream is sealed until the run closes in evaluate mode, so the intervals are withheld
exactly as the rest of the truth. An open Explore run shows the schedule as of the last fault command (an end
the world has scheduled is shown even when the clock has not reached it).

`outcome` is the director stream's `effector.outcome` record for the call (matched by
`command_id`, first unused): `accepted`, `mode` (the failure mode the world chose, for example
`ok` or `silent_no_effect`), `effect_applied` and `ack_latency_ms`, or `refusal` when the world
declined the call. It is absent when the director stream is absent. The outcome is the truth the
operator never sees: every acknowledging mode tells the consumer the same thing, and the report
shows what the world actually did. The values are the world's record verbatim: for
`silent_no_effect` the effect is applied to shadow state only and `effect_applied` still reads
`true`, so read `mode` first.

`kpis` is the director stream's `kpi.truth` record verbatim (domain-spec v0.2 §29), absent when the stream has none.
The Markdown report shows it as three tables: KPIs (mean, minimum, maximum, samples), alarms (priority, activations,
chattering, active and longest active seconds, active at the end) and the alarm system.

## 6. Scoring

`scoring` is present only when `verdict.json` and `label.json` both exist and `run.json` exists.

| Status | When | Fields |
|---|---|---|
| `scored` | `grading.Grade` succeeded | `scorecard`: the scorecard exactly as `streamsim score` prints it |
| `not_scored` | The run is stamped unblinded (an Explore run, or truth revealed on an open run) | `reason` |
| `failed` | Grading refused: the replay does not reproduce the trace, or a document is invalid or missing | `reason`: the refusal, with the run folder's path removed so the bytes do not depend on where the folder lives |

Grading replays the artifact from the domain and adapter documents the artifact embeds
(`grading.Options` with no directories), so the report stays a function of the folder.

## 7. Markdown rendering

`--format markdown` renders the same report document; it adds no information and drops none that
a person needs. Rules, so that the output is byte-stable:

1. UTF-8, `\n` line ends, one trailing newline, no trailing spaces, no HTML.
2. Title `# Run report: <run_id>` (`unknown` when the run has no id), then one status line
   `mode: <mode> | complete: yes/no | ended: yes/no | verification: <status>`.
3. Sections in this order, each `##`: Run, Digests, Verification, Deliveries, Faults,
   Perturbations, Environment, Effector calls, Consumer, Scoring. A section whose source is
   absent is omitted. A truth section of a run with truth and nothing to list reads `None.`; a run
   without truth gets one `## Truth` section holding the line
   `Truth withheld: the run has not ended (no run.json).` in place of the four truth sections, or
   `Truth unavailable: <reason>` when a file belongs to another run.
4. Key-value content is a two-column table `| Field | Value |`; lists are tables with a fixed
   column order; maps (`by_reason`) are tables with sorted keys.
5. Times are RFC 3339 UTC with the shortest fractional part (`model.FormatTime`), followed by
   `+<seconds>s` from the run start where the run start is known. Durations and counts are plain
   integers.
6. Cell text has `\`, `|`, `*`, `_`, `` ` ``, `[`, `]`, `<`, `>` and `&` escaped with a backslash and
   line breaks replaced by a space; the title is escaped the same way. Digests and JSON values
   (`params`, `args`) are written in a code span, in which only `|` is escaped; JSON is compact with
   sorted keys. A `## Torn tail` section lists the files in `torn_tail`.
7. The scorecard is shown as a table of its scalar metrics grouped by section
   (`instrument`, `consumer`, `loop`, `judgment`), keys sorted; the full scorecard is in the JSON
   report.

## 8. Command

```
streamsim report <run-dir> [--format json|markdown] [--out <file>]
```

- `json` (the default) without `--out` prints the report as the command's one stdout document
  (`run-report-v0.1`).
- `--out <file>` writes the chosen format to the file and prints one JSON document,
  `{"report_written": <file>, "format": <format>}`.
- `markdown` requires `--out`. The CLI prints exactly one JSON document on stdout for every
  command (CLI_SURFACE §1); a Markdown rendering is a file artifact, so it is never printed
  beside, instead of, or inside that document.
- A folder with neither `run.json` nor `journal.jsonl` fails with `file_not_found`; a file that
  does not parse fails with `artifact_invalid` (`run.json`, `verdict.json`) or `report_failed`.

## 9. Module

`internal/report` is a core with a file edge, layer 7: facade, `internal/domain` (the report
vocabulary, assembly from parsed inputs, verification, JSON and Markdown rendering; no I/O),
`internal/app` (the use case: read the folder, grade, assemble) and `internal/files` (the folder
reads). It composes `run` (artifact loading), `journal` (reading and digesting streams) and
`grading` (scoring) and is called only by `cli`.

## 10. Proof

- Every report produced in a test validates against `run-report-v0.1`; the embedded copy in
  `internal/schemas` is byte-identical to `contracts`.
- Golden JSON and Markdown for a pinned run folder (`internal/report/testdata/run`) with faults, a
  perturbation, effector calls, a verdict and a duplicated delivery.
- Determinism: building the same folder twice gives identical bytes.
- An Explore run reports `mode: explore`, `unblinded: true` and `not_scored`; a folder cut before
  `run.json` reports `complete: false`, `ended: false` and no truth (canary: with `run.json` removed and
  `ledger.jsonl` and `sealed/` left in place, no fault, perturbation, seed, ledger count or drop
  reason appears in the JSON or Markdown of an Evaluate run).
- A tampered `trace.jsonl` reports `files_differ`.
- A `run.json` or sealed journal of another run reports `truth_unavailable` and leaks no truth; a
  torn last line is dropped and named in `torn_tail`, a damaged middle line fails.
