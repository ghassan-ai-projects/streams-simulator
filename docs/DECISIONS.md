# Decisions — Streams Simulator Implementation

This file records every decision made while implementing the design in
`docs/`, per the implementation plan's S0 gate ("record decisions in
DECISIONS.md, freeze constants, validate the example domain"). Decisions are
ordered by when they were made. A decision marked **deviation** changes the
documented design and must be reviewed against `docs/design/*` before the
next stage.

## S0 decisions

### D-01. World ids are deterministic, not random
The design's world ids (`w-7fa2` in examples) must not come from a wall
clock or OS entropy: they are derived from `fnv1a64(domain_id:seed)`, so a
replayed run reconstructs the same id and the same substream names. A world
created with an explicit id keeps it (the golden-vector test relies on this).

### D-02. Substreams are named, never shared
Every RNG consumer draws from `splitmix64(seed XOR fnv1a64(name))` where
`name = <world_id>/<entity_id>/<channel>/<purpose>`. Substreams are created
lazily and cached per world. Isolation (S1 gate 5) follows from the naming:
adding an entity, channel, fault or perturbation never shifts another
consumer's stream.

### D-03. F1 states integrate independently, with inputs frozen at step boundaries
Each F1 state advances with its own declared dt; the RK4 stages read the
inputs' values at the step start (explicit coupling). This satisfies the
analytic cross-check (S1 gate 1: agreement with the closed-form first-order
lag across the swept parameter space) and keeps the integrator simple. It is
first-order accurate in the couplings, which is acceptable for a test
instrument; the gate is what the design specifies.

*Deviation (2026-10-12, domain-spec v0.2, P5a):* the rule is stated per kind, not per domain. A
kind's dynamics apply to every instance of the kind independently, so the coupling stays inside one
entity and a state never reads another instance's state. Part 1 of v0.2 (kinds and instances) leaves
the integrator and the step-boundary freezing unchanged: the rewritten domains run bit-identically.
The v0.2 vocabulary of P6 (`store`, `flow`, `regime`, `trip`, and the event clock of P6e) supersedes
"each state advances with its own declared dt" by a shared event schedule between which the system
is linear; that revision is recorded here when it lands, with its RNG draw-order table.

### D-04. Faults and effect kicks are recency-ruled drivers of a state
A state's value is `natural + contributions`, but when both a fault and an
effect kick target the same state, the **most recently activated** driver
rules. This is the only semantics under which the domain's expected
responses work: `start_aerator` must recover an `aerator_failure`, and a
fault must break a running aerator. Purely additive offsets would make one
of those impossible.

*Deviation (2026-10-12, domain-spec v0.2, P5a):* faults and effectors belong to the kind, so a fault
or an effect kick drives a state of one instance and the recency rule is evaluated per instance.
Part 1 changes no precedence. The new driver forms of P6 (`regime`, `trip`, latched `lockout`,
`responder` effects) also act on states; where one of them can compete with a fault or a kick, the
precedence is decided in the v0.2 semantics part 2 and recorded here before that phase ships. Until
then a domain that needs it cannot be expressed, so the recency rule stays the only rule.

### D-05. Silent-no-effect shadows are per-state kick lists
The counterfactual is carried as shadow kicks added to the real value for
confirmation channels only; `silent_no_effect` changes nothing physical.

### D-06. The observability solver runs the world twice, noiselessly
The onset timestamps come from running a clean and a faulted world with
noise disabled and diffs the detector quantity against the declared SNR
floor. This reuses the verified integrator instead of deriving closed forms
per detector form, and is exact for the linear F1 family. Scenario setup
(aerator running at night) is applied through `SetupCall`s with effectors
forced to mode ok.

### D-07. The trivial-baseline audit is adversarial by construction
The panel's five detectors are fitted with hindsight on the scenario itself
and its negative-class control. The audit rejects aggressively: at a 24 h
horizon, a ramped probe fouling saturates past the natural range and a
hindsight threshold solves it — the audit says so, and the suite keeps only
the non-trivial draws. A domain that cannot sustain the declared
composition reports an aggregated **terminal state** (G-06) rather than
silently degrading; the generator adapts its negative sampling to hold the
declared band when the domain can sustain it.

### D-08. Adapters render with a closed transform set; conformance is golden-byte
The native event renders through `source/const/concat/template/format_time/
counter/run_meta` only. Every shipped adapter proves itself against the
committed 12-event fixture: every rendered record validates against its
declared output schema, and the full stream byte-matches the committed
golden. Goldens are regenerated deliberately (`REGEN_GOLDEN=1 go test
./internal/adapter/internal/domain/ -run TestRegenerateGoldens`), never automatically.

### D-09. Effector failures are deterministic per (world, effector)
Failure modes and ack latencies come from the effector's own substream;
tests may force a mode via `SetFailureMode`. Idempotency replays the
original result and does **not** add a second effector-log entry: the
command_id is the action, a retry is the same action.

### D-10. Entity churn rates are bounded by usability
The catalog describes k8s-cluster at 4,000 pod births/hour. The shipped
domain uses 120/hour with an 18-minute mean lifetime — the same churn
behaviour at a rate a single machine can stream (the domain's description
documents the deviation). The mechanism is unchanged.

## Deviations from the documented design

### D-11 (deviation). File-based persistence instead of SQLite
The design's tech stack lists SQLite WAL (`modernc.org/sqlite`) for storage.
This implementation persists runs as JSON artifacts (`run.json`, trace,
ledger, world-state history, verdict) because:
- the run artifact *is* the reproduction contract — one file, no engine;
- AGENTS.md forbids top-level dependencies without justification and names
  the MCP Go SDK as the only documented exception; modernc.org/sqlite is a
  very large pure-Go dependency;
- file artifacts make `replay`/`verify` trivially offline.
If a real store is ever needed, the run layer is the single seam.

### D-12 (deviation). Hand-written JSON Schema validator instead of a schema-engine dependency
The core has a zero-external-dependency requirement. The committed contract
schemas are validated by `internal/jsonschema`, which implements exactly the
draft 2020-12 keywords the six contracts use (including recursive `$defs`).
The embedded schema copies are proven byte-identical to `contracts/` by
a test. The validator is closed and reviewed, not a general engine.

### D-13 (deviation). `simdet` build tag semantics
Under `simdet`, `internal/wall.Now` returns the zero time and the http-push
wall sub-mode cannot link. `make test-simdet` runs the suite with `-tags
simdet`, and the race run of `make test-ci` (the `ci-check` test tier) carries the same
tag; CI gates on it, so no test can depend on a wall clock by accident.

*Erratum (2026-10-10, hardening):* the code has no http-push wall sub-mode.
The seam is now used by the run module's clock edge, so under `simdet` the
`created_at` and `unblinded_at` stamps of run artifacts are the zero time. The
CLI's process edge (`internal/cli/internal/process`) reads the clock directly
because the binary is never built with `simdet`; it is declared in `ioEdges`.

### D-14 (deviation). env.inject is record-only
Environment faults target a consumer's process, which a simulator must not
act on. `sim.env.inject` validates the target is configured, records the
command, and never signals anything. Replay treats it as a record.

*Deviation (2026-10-12, domain-spec v0.2, P5a):* unchanged and narrowed. An environment fault still
targets a consumer's process and stays record-only. What v0.2 adds is separate: system entities
(weather, tariff, plant-wide measurements, P5b) and the plant wire (P7) make some environment
changes world commands that the simulated plant does act on, as effectors and schedules of the
world, and those are not `env.inject`. D-14 does not apply to them, and the two are told apart by
what they target: the consumer's process (record-only) or the simulated plant (a world command).
Part 1 of v0.2 adds neither.

### D-15 (deviation). Domain specs are kinds and instances
Domain-spec v0.2 (P5a, [`DOMAIN_SPEC_V0_2.md`](design/DOMAIN_SPEC_V0_2.md)) replaces the flat v0.1
document with `kinds` (the template) and `instances` (the population). The v0.1 contract is
removed. The domain digest is the digest of the v0.2 document and `sim_version` is `0.3.0`.
Explicit entity ids are checked against the instance sets' templates and count bounds when the run
is created. P5a compiled the document to the v0.1-shaped runtime form with one kind; P5a2 removes
that lowering: the validated document is the runtime form, a domain may declare several kinds and
instance sets, channel, fault and effector names are unique across the domain, and every
reference resolves inside its kind. Single-kind domains are bit-identical to P5a.

### D-16 (decision, deferred until the UI goal is done). A run stores actions, not data
Owner decision, 2026-10-14. A 1-hour run of the 26-hall campus wrote 1.6 GB (each reading written about five times
across trace, ledger, state history and two journals) and took 4–9 GB of memory to run, report and verify. The owner
does not want gigabytes written at all. Because a run is deterministic, its actions reproduce every stream exactly,
so the run folder keeps only `run.json`: the inputs (simulator version, domain, adapter, seed) and every action in
order (clock advances, faults, perturbations, environment changes, consumer effector calls, quiescence), in the
style of event sourcing / CQRS. Trace, delivery ledger, true state, KPIs and the report are projections computed by
re-running the actions in the binary, streamed and never written to disk; the delivery ledger is one of those
projections. The trace, ledger, state-history and journal files, the `replay` and `verify` commands with their
file comparisons, the report's file checks, the bridge's journal streaming and sealed-file gate, and the canvas's
journal-based playback are removed. Acceptance test A5 (playback without the binary) is dropped. Seeking re-runs
from the start; a few minutes for a long campus run is acceptable, so no checkpoints. Not started: the canvas UI goal
([GOAL-UI-PRODUCTION.md](simulation-canvas/GOAL-UI-PRODUCTION.md)) is finished first. Before implementing, a design
note covers the new `run.json` contract, the projection API the canvas uses, how determinism is still evidenced once
`verify` is gone, and the impact on the nine non-negotiables (determinism, delivery ledger, sealed oracle).

## Frozen constants

| Constant | Value | Where |
|---|---|---|
| `sim_version` | `0.3.0` | `internal/model` |
| default world start | `2026-01-01T00:00:00Z` | `internal/model` |
| RFC 8785 canonical JSON | all digests | `internal/canonical` |
| intermittent fault period | 3600 s (or 3600/rate_per_hour) | `internal/world` |
| interlock/effector refusal | terminal, never retried | `internal/world` |
| trivial-baseline cutoff | 0.9 balanced accuracy | `internal/audit` |
| ack idempotency window | 3600 s default | `internal/world` |
| first-observable/unavoidable SNR | 3 / 10 defaults | domain specs |

## Not built (deliberately, per the cut list)

- A continuous `sim.clock.run` tool; continuous mode maps to repeated
  `sim.clock.advance`. Streamable HTTP is implemented for the operator
  endpoint, while the director remains on stdio.
- F3/FMU co-simulation: the seam exists in the schema and cross-checks
  reject F3 specs with `not_implemented`.
- F2 reference models (FAO-56 soil water balance and peers): the integrator
  has no F2 forms, so cross-checks reject F2 specs with `not_implemented`
  rather than silently emitting a static state. Shipped domains use F0/F1,
  which the open-field-irrigation domain documents.
- A broker sink: `not_implemented`.
- A separate production consumer: the CLI reference consumer can use the live
  operator endpoint via `--mcp`, and the closed loop is covered by tests, but
  no external production consumer is shipped.
