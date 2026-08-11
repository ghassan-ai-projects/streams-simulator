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

### D-04. Faults and effect kicks are recency-ruled drivers of a state
A state's value is `natural + contributions`, but when both a fault and an
effect kick target the same state, the **most recently activated** driver
rules. This is the only semantics under which the domain's expected
responses work: `start_aerator` must recover an `aerator_failure`, and a
fault must break a running aerator. Purely additive offsets would make one
of those impossible.

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
./internal/adapter/ -run TestRegenerateGoldens`), never automatically.

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
The embedded schema copies are proven byte-identical to `docs/contracts/` by
a test. The validator is closed and reviewed, not a general engine.

### D-13 (deviation). `simdet` build tag semantics
Under `simdet`, `internal/wall.Now` returns the zero time and the http-push
wall sub-mode cannot link. `make test-simdet` runs the suite with `-tags
simdet`; CI gates on it, so no test can depend on a wall clock by accident.

### D-14 (deviation). env.inject is record-only
Environment faults target a consumer's process, which a simulator must not
act on. `sim.env.inject` validates the target is configured, records the
command, and never signals anything. Replay treats it as a record.

## Frozen constants

| Constant | Value | Where |
|---|---|---|
| `sim_version` | `0.1.0` | `internal/model` |
| default world start | `2026-01-01T00:00:00Z` | `internal/model` |
| RFC 8785 canonical JSON | all digests | `internal/canonical` |
| intermittent fault period | 3600 s (or 3600/rate_per_hour) | `internal/world` |
| interlock/effector refusal | terminal, never retried | `internal/world` |
| trivial-baseline cutoff | 0.9 balanced accuracy | `internal/audit` |
| ack idempotency window | 3600 s default | `internal/world` |
| first-observable/unavoidable SNR | 3 / 10 defaults | domain specs |

## Not built (deliberately, per the cut list)

- Streamable HTTP transport for MCP (stdio is the default and the cut-list
  item); `sim.clock.run` (continuous mode) maps to repeated `advance`.
- F3/FMU co-simulation: the seam exists in the schema and cross-checks
  reject F3 specs with `not_implemented`.
- F2 reference models (FAO-56 soil water balance and peers): the integrator
  has no F2 forms, so cross-checks reject F2 specs with `not_implemented`
  rather than silently emitting a static state. Shipped domains use F0/F1,
  which the open-field-irrigation domain documents.
- A broker sink: `not_implemented`.
- The reference consumer over a live MCP endpoint: the CLI consumer is
  observe-only over a trace; the in-process closed loop is exercised by
  tests, and a full stdio session is a small follow-up.
