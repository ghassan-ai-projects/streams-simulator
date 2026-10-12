# Changelog

This project has not established a formal release stream yet. Until it does, changes are recorded in pull requests and the repository history.

## Unreleased

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
