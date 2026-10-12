# Changelog

This project has not established a formal release stream yet. Until it does, changes are recorded in pull requests and the repository history.

## Unreleased

- **Run journal** (`docs/design/RUN_JOURNAL.md`, `run-journal-v0.1`, new `internal/journal` module):
  two append-only JSON Lines streams written while a run happens. The observer stream
  (`journal.jsonl`) holds what a consumer could know: the run's identity, every delivered
  observation exactly as the sink received it (the adapter's rendered line), clock advances, the acknowledgement each effector call returned,
  consumer reports and the close; never the seed (MCP_SURFACE §4.5). The director stream (`sealed/journal.jsonl`) holds commands,
  delivery rows, world state, effector outcomes, unblinding and the close. Truth-bearing records
  never enter the observer stream (canary test). `run --out <dir>` and a director world write both in
  the run directory; no `--out`, no files. Both streams are pure functions of the run;
  `run.json` records `journal_digest` and `sealed_journal_digest`; `replay` rebuilds both in memory
  and `verify` fails with `replay_mismatch` naming the stream (and, from the files beside the
  artifact, the first differing record). Journals are buffered, flushed at command boundaries and
  synced at the end; a torn last line is ignored by the reader; an unencodable record fails the run.
  Consumer reports are session records: in the file, outside the digest (replay cannot reproduce
  them). A `clock.advance` is journaled with its target only (`await_consumer` is pacing).
  Fuzz target `FuzzJournalRead` in `make fuzz` and `make fuzz-soak`.
- **Explore mode:** `run --explore` and `sim.world.create {explore: true}` stamp the run unblinded
  from creation (`mode: "explore"` and `unblinded: true` in `run.json`, `run.opened.mode`, a
  `run.unblinded` record). **Behaviour change:** `streamsim score` refuses every run stamped
  unblinded (`score_failed`), as the director's `sim.score` already did; it used to print a
  scorecard carrying `unblinded: true`.
- `run.json` gains `mode`, `journal_digest` and `sealed_journal_digest` (behaviour pin: only
  `run.json` lines change, and the journal files are pinned). `sim_version` is unchanged: the trace,
  ledger, world digest and determinism tuple are identical.
- Director composition tests moved to `test/acceptance` (finding R2d-01); `Director.RunOf` is the
  documented test-harness handle on the run behind a world, replacing the `export_test.go` seam.

- **Fix (truth leak):** `sim.effector.invoke` returned the world's internal result to the
  consumer, including the failure `mode`, `effect_applied` and `ack_latency_ms`, which exposed
  `silent_no_effect`. The operator now returns only the acknowledgement fields of MCP_SURFACE
  §4.3 (`model.EffectorAcknowledgement`), and every acknowledging mode returns a byte-identical
  response (`confirmed_no_effect` now promises an effect time like the others). A lost
  acknowledgement reads as `reason: "acknowledgement_timeout"` with no effect time. The reference
  consumer reads the same acknowledgement.

- **Scenario files** (`docs/design/SCENARIO_FILE.md`, `scenario-v0.1`): `run --scenario <file>`
  runs a repeatable scenario; run flags build the same scenario in memory, so there is one
  execution path, and the behaviour pin shows every existing run unchanged. `scenario from-run`
  exports the scenario that reproduces a run (identical command log, trace and world digest).
- **Behaviour change:** a scenario is checked before anything is delivered (`scenario_invalid`
  for a file, located at the field; `usage_invalid` for flags, each issue naming the flag). A
  static check reports every effector, fault, perturbation and profile that does not exist,
  every refused effector argument or fault or perturbation parameter, every inexact time and
  every backwards or over-long advance. A dry run on the same world with emission off then finds
  the first command the run would refuse as it unfolds: entities that do not exist at that point
  (churning domains included), ids already in use, clears of ids never assigned. During a run, an
  effector call the world refuses through an interlock or a reused command id is a recorded
  outcome. Replay still keeps every recorded refusal.
- **Behaviour change (every surface):**
  - A run's clock may not move more than 30 days past its start (`run.MaxRunSpanNS`;
    `--duration` at most 2592000 s).
  - A perturbation window that ends before it starts is refused (`--perturb x@100@50` used to
    succeed and never act).
  - An entity may not be added before the world clock.
  - A world starts between 0 and `world.MaxStartNS` (`--start-time` likewise); flag offsets
    and `--duration` must be finite numbers within the run span.
  - Artifacts recorded with such commands no longer replay.
- **Behaviour change:** `scenario from-run` checks the scenario as `run --scenario` would
  (`--domains-dir`, `--adapters-dir`) and writes nothing for a run that cannot be expressed as a
  runnable scenario. Effector calls the world refused at admission are dropped and counted
  (`refused`); a changed domain or adapter is `replay_mismatch`.
- **Behaviour change:** a run refuses a profile its domain does not declare on every surface,
  replay included (the director answers `domain_invalid`); an artifact recorded with such a
  profile no longer replays. `force_failure_mode` must be an effector failure mode in the
  scenario and run-artifact contracts and when a world is created.
- **Behaviour change:** scenario and exported `clock.advance` commands carry only `to_ns`;
  `await_consumer` is pacing of a live session, and scenario runs record it as `false`.
- **Behaviour change:** flags may follow positional arguments in every command (`replay <run.json>
  --domains-dir d` now honours the flag); `--` ends flag parsing unless it is a flag's value. A
  command or verb given more positional arguments than it takes fails with `usage_invalid`
  instead of ignoring them.
- **CLI output contract** (`docs/design/CLI_SURFACE.md`, `cli-output-v0.1`): every command except
  `help` and `mcp` prints exactly one JSON document on stdout, for failures too. Failures carry a
  stable code, the file at fault, every contract issue with a JSON pointer when known, and for
  `verify`/`adapter verify`/`score` the result that diverged.
- **Behaviour change:** argument errors that are not flag-parse errors (missing or unknown verb,
  missing required argument, malformed `--fault`/`--effector` value) now exit 2 with
  `usage_invalid` instead of 1. A failed `verify` prints the error document with the replay
  report as `result` instead of the bare report.
- Domain, adapter, run-artifact, verdict, label and ledger failures are typed
  (`model.ValidationError`) instead of text; `domain.ErrUnknownDomain` names an unknown id.
- **Behaviour change:** argument values are checked before running: `--sink` must be a known sink,
  `http-push` needs `--sink-target`, `--duration`, `suite --n`, `refconsumer --window` and
  `--threshold` must be positive, `suite` needs `--domain`, and `--profile` must name one of the
  domain's profiles (a `run` with an unknown profile used to succeed and ignore it). All exit 2.
- **Behaviour change:** `score` checks the verdict and label against their contracts, decodes the
  ledger and state history strictly, and loads the run artifact first; `-h` works after a verb
  (`domain validate -h`, `device -h`); flags may precede a verb (`adapter --adapters-dir d verify f`).
- `run.ErrDigestMismatch`, `adapter.Lookup`/`adapter.ErrUnknownAdapter` and
  `model.ValidateGroundTruth` are new typed entry points.
- Gates: `godox` and `nolintlint` in lint; a consumer-knowledge gate with a stale-exception check; a
  pinned `device serve` transcript; failure documents in the behaviour pin (stdout and stderr
  separately); `make behaviour-pin-check`; `make phase-gate`.

- Introduced a curated public documentation set under `documentation/`.
- Documented the distinction between the public set and the engineering archive under `docs/`.
- Added explicit limitations and evidence guidance for the current working-tree inventory.

When a release process is established, each entry should name contract changes, behavior changes, migration notes, and evidence for the exact release commit.
