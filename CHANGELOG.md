# Changelog

This project has not established a formal release stream yet. Until it does, changes are recorded in pull requests and the repository history.

## Unreleased

- **Schedule, regime and peak** (`docs/design/DOMAIN_SPEC_V0_2.md` section 16; phase P6b, review item S9). `f0.schedule` adds
  a clock-aligned piecewise profile (step or linear) to an F0 state; `f1.form: "regime"` is a lag whose gain and time constant
  switch when a selector state crosses a threshold at a step boundary; `f1.form: "peak"` is the billing-period demand peak: the
  largest epoch-aligned block mean of one input, with a ratchet on the previous period's peak and seeded `initial` and
  `previous_peak`. New fixture `utility-plant.domain.json`. Shipped domains and every pinned output are unchanged.

- **Stores and flows** (`docs/design/DOMAIN_SPEC_V0_2.md` sections 13 to 15; phase P6a, findings R0-11, R2-05,
  R5b-01..03). `f1.form: "store"` with a `capacity` accumulates its inputs and flows; `kinds[].flows` move
  `conductance × Π scale × (from − to)` between stores of one entity, or drain a store without `to`. Every store of an
  entity is integrated jointly on one shared grid; every store of a kind shares one `dt_ms`. **Integrator fix:** every
  integrated state of an entity now advances together, earliest step first, and an input is read at its exact step
  boundary, so no value depends on the order states are read (before, a lag fed by a lag read ahead could see a later
  value, and an emitting world could differ from the silent oracle world). A test proves the domain schema holds no free-text expression. Narrowing refuses a target named twice and
  argument bounds that leave no value; a `peer_residual` fault added by an override may not watch a channel only the
  override adds; a profile's entity count sizes the first templated instance set; idempotency compares argument numbers
  by value. New fixture `transfer-plant.domain.json`. Shipped domains and every pinned output are unchanged.

- **Named instances, overrides, links, system entities and boards** (`docs/design/DOMAIN_SPEC_V0_2.md` sections 8 to 12;
  phase P5b, finding R5a-02). An instance set may list `names` instead of or beside an `id_template`; an explicit entity id
  is accepted when it is a name or fits a template. `override` on an instance set adds states, dynamics (for added states),
  channels, faults and effectors, or narrows a state's bounds, a channel's range or an effector argument's bounds; removing
  or retyping is refused with a located issue, and the world gives each entity the effective kind of its set. `kinds[].system`
  marks a plant-wide kind with exactly one named instance. `link_types` and `links` declare typed topology between
  instances; `boards` group the instances hosted on one device and own the `boot_id` (structure only, no run-time behaviour).
  The domain-spec contract moves its definitions of state, dynamics, channel, fault and effector into `$defs` so a kind and
  an override share them. Shipped domains and every pinned output are unchanged. New world methods `StateNamesOf` and
  `EffectorOf`; new fixture `structured-plant.domain.json`.

- **Multi-kind world** (`docs/design/DOMAIN_SPEC_V0_2.md` §2.3 and §3; phase P5a2, finding R5a-01). A domain may declare
  several **kinds** and several **instance sets**, each set following a kind with its own `id_template`, `count` and
  `churn`; the `maxItems: 1` limits on `kinds` and `instances` are lifted. Channel, fault and effector names are unique
  across the domain, state names are unique within a kind, and every reference (observes, dynamics, affects, effects,
  detector channels, expected effector, confirmation channels) resolves inside its own kind; a kind with no instance
  set is refused. The validated document is now the runtime form (`model.DomainSpec` holds `Kinds` and `Instances`; the
  P5a lowering is removed). Each world entity carries its kind, which selects its state, dynamics, channels, faults and
  effectors: injecting a fault or invoking an effector on an entity of another kind is refused, a peer-residual
  detector pools only the entity's kind, a suite scenario draws its entity from the fault's kind and the audit uses
  that kind's channels and control fault. Single-kind domains are bit-identical (runtime-equivalence digests,
  trace, ledger and history in the behaviour pin). **Behaviour changes:** `catalog describe` prints the v0.2 document
  (`kinds`, `instances`) instead of the flat runtime shape; an explicit entity-id list is matched to an instance set
  and counted per set; `world.RenderID` is replaced by `world.PopulationIDs`.

- **Domain spec v0.2, kinds and instances** (`docs/design/DOMAIN_SPEC_V0_2.md`, `domain-spec-v0.2`, replacing
  `domain-spec-v0.1`; phase P5a). A domain declares one **kind** (the entity template: state, dynamics, channels,
  faults, effectors; its name is the `entity_type`) and one **instance set** (`id_template`, `count`, `churn`) that
  follows it. The document compiles to the v0.1-shaped runtime form, so the world, perturbation, adapter, truth,
  score and run code are unchanged; the eight shipped domains are rewritten and each compiles to the runtime world
  its v0.1 form produced (the recorded trace, ledger, world-state history, replay and suite outputs are
  byte-identical). **Behaviour changes:** `domain_digest` is the digest of the v0.2 document (every shipped domain's
  digest, every run artifact's `domain_spec` and `world_digest` change once); `sim_version` is `0.3.0`; run artifacts
  recorded before this change no longer replay (their `domain_spec` is v0.1); `catalog describe` prints the runtime
  form and the new digest.
- `streamsim domain format <path>` rewrites a valid domain file in the canonical layout (stable member order,
  line-fitting objects) and prints `{formatted, changed, digest}`; idempotent, digest unchanged, an invalid
  document is refused and left untouched. `cli-output-v0.1` gains `domain_format`.
- **Validation issues are located** (findings R1-01, R0-10, S15): a "missing required property" or "additional
  property" issue from any contract points at the member, not its parent, and every domain cross-reference issue
  carries a JSON pointer into the v0.2 document (for example `/kinds/0/channels/1/observes`). Domain cross-checks
  also reject a `count.default` outside `count.min`..`count.max`.
- **Explicit entity ids are checked** (finding R2-02): a scenario's `world.entity_ids` and the director's world
  creation must fit the domain's instance id template and count bounds; a violation refuses the world (a scenario
  violation at `/world`).

- **Run report** (`docs/design/RUN_REPORT.md`, `run-report-v0.1`, new `internal/report` module):
  `streamsim report <run-dir> [--format json|markdown] [--out <file>]` summarises a run folder as one
  deterministic document, a pure function of its files: identity and inputs, recorded digests,
  `verification` (do the files still agree with the recorded digests; not a replay), delivery counts
  with drop reasons from the ledger, faults, perturbations, environment faults and effector calls
  with the world's real outcomes from the command log, mode (`evaluate` or `explore`), whether the
  run is complete and ended, the consumer from `verdict.json` (its `config_digest` only when the
  verdict recorded one) and the scorecard when `verdict.json` and `label.json` are present. The
  truth sections appear only for a run that has ended or an Explore run; the report of an open
  Evaluate run reads neither the ledger nor `sealed/`, so it cannot be used to read ahead. JSON is the command's one stdout document; Markdown
  requires `--out` and stdout then names the file (`{report_written, format}`), keeping stdout one
  JSON document. `cli-output-v0.1` gains `report` and `report_written` and the `report_failed` code.
- `grading.Options` with no domains or adapters directory now grades from the domain and adapter the
  artifact embeds (`run.EmbeddedInputs`); `streamsim score` still passes its directories.

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
