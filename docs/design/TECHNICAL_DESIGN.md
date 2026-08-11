# Streams Simulator — Technical Design

Status: design, ready for review
Date: 2026-08-11
Codename: `streamsim`

## 1. Outcome

A standalone, deterministic, closed-loop world simulator: it generates realistic event
streams from data-defined domains, corrupts their delivery in specified ways, carries
sealed ground truth, accepts commands back through declared effectors, and scores whatever
consumed it.

**It is an independent product.** It has one customer today and knows nothing about that
customer. No consumer's name, schema, field, or behaviour appears in the binary. A
consumer is a file — an [output adapter](../contracts/output-adapter-v0.1.schema.json) —
plus, optionally, an MCP client that invokes effectors and submits a verdict.

It is also a **test instrument**, and that word constrains everything. An instrument less
trustworthy than the system it measures is worse than no instrument, because it produces
confident wrong answers. So its determinism guarantees are strong, its ground truth is
generative rather than annotated, and its own correctness is proven against analytic
oracles before it judges anything.

## 2. Product boundary

### Included

- data-defined domains loaded through one injection path, with no per-domain code;
- a discrete-event world core with a virtual clock and no wall-clock dependence;
- a seeded PRNG tree giving byte-reproducible runs;
- a native event format and declarative output adapters, so consumers are data;
- three sinks: `inproc`, `file`, `http-push`;
- world faults, delivery perturbations and environment faults as three separate,
  independently seeded injection surfaces;
- effectors with ack latency, typed failure modes, interlock predicates and modelled
  physical effect — the closed loop;
- generative ground truth with three onset timestamps and a trivial-baseline audit;
- one MCP server with `director` and `operator` roles;
- a command log making an improvised session exactly reproducible;
- a neutral consumer-verdict contract and scorecards over it;
- a reference consumer, so the product is demonstrable with nothing else installed.

### Explicitly excluded

- a physics engine, a solver, or any claim of physical validity beyond the declared
  fidelity tier;
- real actuation of anything, ever;
- any knowledge of a consumer's internals — its storage, its logs, its vocabulary;
- consumer-specific code, including plugins, which are consumer-specific code renamed;
- evidence over MCP, for the reasons in
  [TRANSPORT_ANALYSIS §3](../research/TRANSPORT_ANALYSIS.md);
- FMI/FMU co-simulation — the seam is specified in §9.4, nothing is built;
- binary adapter encodings (protobuf, Sparkplug B);
- broker sinks until a consumer needs one;
- a web UI, multi-tenancy, distributed execution;
- clinical, safety-certification or regulatory validity of any kind;
- any simulation of a real, named, identifiable organization, site or person.

An excluded item requires a new decision. It does not enter as "small infrastructure."

## 3. Determinism architecture

Everything rests on one property:

> **A run is a pure function of `(sim_version, domain_digest, adapter_digest, seed,
> command_log, sink)`.** Wall-clock time, map iteration order, goroutine scheduling, MCP
> call timing and the consumer's read pattern must all be incapable of changing an emitted
> byte.

### 3.1 The discrete-event core

The world advances by popping a priority queue keyed by `(event_time_ns, tiebreak_seq)`,
with `tiebreak_seq` a monotonic counter assigned at scheduling time. `time.Now()` appears
in exactly two places: the `http-push + wall` sink and the log timestamp. Both are outside
the world model, and a `simdet` build tag makes them unavailable so the deterministic test
suite cannot link them by accident.

### 3.2 The PRNG tree

A single global generator is the classic determinism bug: output depends on the order
callers happen to draw, so adding one entity shifts every other entity's noise. Instead:

```text
substream(name) = splitmix64( seed XOR fnv1a64(name) )
name = "<world_id>/<entity_id>/<channel>/<purpose>"
```

Substreams are created lazily, cached by name, never shared. `purpose` separates `noise`,
`delay`, `dropout`, `fault_shape` and `perturb`, so enabling duplicate injection does not
change sensor noise.

**Test**: generate; add an unrelated entity; regenerate. Every original entity's events
byte-identical.

### 3.3 Go-specific hazards

| Hazard | Rule | Test |
|---|---|---|
| randomized map iteration | never iterate a map to produce output; sort keys or use an ordered slice | `go test -count=20` over the golden digest; a lint forbidding `range` over a map in `world/` and `adapter/` |
| goroutine scheduling | the world core is strictly single-goroutine; concurrency lives only in sinks, behind a bounded channel with ordered writes | race detector plus a single-writer assertion |
| float non-associativity | fixed-order aggregation, no parallel reduction | golden digest in CI |

Emitted numeric values are quantized to the channel's declared `resolution` before
hashing. Every real instrument has a resolution, so this is physically honest as well as
being the thing that removes cross-architecture float divergence as a hazard.

### 3.4 The command log

Every mutation appends a record with a monotonic `seq`, canonical arguments, and the world
clock at which it applied:

```json
{"seq": 7, "at_ns": 1785312000000000000, "op": "fault.inject",
 "args": {"entity_id": "site-a/pond-3", "fault": "aerator_failure",
          "onset_ns": 1785315600000000000, "params": {"severity": 0.8}}}
```

A run is reproduced by replaying the log, **not** by re-issuing MCP calls. That is what
makes an improvised session reproducible, and `streamsim replay run.json` needs no server.
Concurrent calls serialize under the world mutex before `seq` assignment, so the order is
total regardless of arrival concurrency.

### 3.5 The run artifact

One file reproduces any run — see
[run-artifact-v0.1](../contracts/run-artifact-v0.1.schema.json). If a failure cannot be
reduced to one of these, it is not reproducible and the architecture is wrong.

### 3.6 What is *not* deterministic

- `http-push + wall` receipt timing. Stamped `reproducible: false`.
- Anything a consumer does. The simulator makes no claim about it.
- Real elapsed wall time.

Where determinism does not hold the run record says so, and the scorer refuses
hash-comparison metrics for it.

## 4. Architecture

```text
       ┌──────────────── MCP: director role ─────────────────┐
harness│ catalog · world · clock · fault · perturb · env      │
       │ export · run · truth · score                        │
       └──────────────────────┬──────────────────────────────┘
                              │ command log
   domain spec ──────────────►│
   (data)                     ▼
                     ┌──────────────────┐
                     │   world core     │  seeded · discrete-event
                     │  hidden state    │  dynamics F0/F1/F2
                     └───┬──────────┬───┘
       native events     │          │  effects, with time constants
                         ▼          │
                 ┌───────────────┐  │
                 │ perturbation  │  │
                 └───────┬───────┘  │
                         ▼          │
   adapter ─────►┌───────────────┐  │
   (data)        │    adapter    │  │
                 └───────┬───────┘  │
                         ▼          │
                 inproc · file · http-push
                         ▼          │
                 ╔═══════════════╗  │
                 ║   consumer    ║  │
                 ╚═══════╤═══════╝  │
                         │          │
                         └──────────┘
                  MCP: operator role
          effector.invoke · consumer.report

       ┌──────────────────────────────────────────────────────┐
       │ truth (sealed) · delivery ledger · effector log       │  director only
       │ scorer · run artifact                                 │
       └──────────────────────────────────────────────────────┘
```

Six components, one binary:

| Component | Responsibility |
|---|---|
| **world core** | hidden state, dynamics, fault overlay, DES queue, effect application |
| **perturbation layer** | delivery-level corruption of an otherwise perfect emission stream |
| **adapter layer** | projection into a consumer's format; pure data-driven |
| **sinks** | delivery |
| **MCP server** | `director` and `operator` roles, disjoint by construction |
| **truth and scoring** | sealed labels, observability solver, scorecards over submitted verdicts |

Two placements are deliberate.

**Perturbation sits between world and adapter.** The world produces what physically
happened; the perturbation layer produces what the observer got. Ground truth is defined
against the former, so a scenario the consumer never saw is a *transport* miss, not a
reasoning miss. Conflating them is how a simulator starts blaming a consumer for its own
dropped packets. §5.5 makes that distinction enforceable rather than merely available.

**Adapter sits after perturbation.** Perturbations are physical facts about delivery —
a duplicate is a duplicate in every format. Putting the adapter last means every consumer
sees the same perturbation, which is what makes cross-consumer comparison meaningful.

## 5. Injection surfaces

Three, independently seeded, never conflated.

### 5.1 World faults — "the plant is broken"

Applied inside the world core; they change hidden state; they are what a consumer is
supposed to diagnose. Each declares affected state, an onset shape (`step`, `ramp`,
`exponential`, `intermittent`, `stochastic`), a severity parameter, and an
**observability model** — the mapping from severity to detectability, which is what makes
the onset timestamps computable rather than guessed.

### 5.2 Delivery perturbations — "the observer is unreliable"

Applied in the perturbation layer. They change nothing physical. They test a consumer's
ingest, time handling and deduplication.

| Perturbation | Parameters | What it probes |
|---|---|---|
| `duplicate_burst` | rate, window | duplicate handling |
| `id_reuse` | rate | same identity, different payload |
| `reorder` | max displacement | out-of-order tolerance |
| `delay_tail` | distribution | heavy-tailed lateness |
| `gross_backfill` | gap, burst size | store-and-forward reconnect |
| `drop` | rate | silent loss — the case absence detection must catch |
| `clock_skew` | offset, sign | `observed_time < event_time` |
| `non_monotonic` | — | receipt order violations |
| `out_of_enum` | rate | undeclared string values |
| `out_of_range` | rate, magnitude | values outside the nameplate range |
| `unit_mismatch` | — | wrong unit on a declared channel |
| `oversize` | bytes | payload limits |
| `malformed` | — | broken framing; one bad record must not poison a file |
| `nan_inf` | — | non-finite numerics |
| `storm` | multiplier, duration | rate limits and budgets |
| `producer_flap` | period | birth/death bursts |
| `time_encoding` | offsets, `Z`, `+00:00` | identical normalization |
| `precision_edge` | ns-level | sub-microsecond retention |
| `injection_probe` | payload set | §5.4 |

### 5.3 Environment faults — "the consumer is broken"

Applied to a *configured target* — a process id, a URL, a container name supplied at world
creation. The simulator has no idea what the target is.

`pause` (SIGSTOP — the slow-node case, nastier than a crash), `kill`, `endpoint_down`,
`partition`, `clock_jump`.

`pause` deserves emphasis: a paused-then-resumed consumer is exactly the condition that
produces a stale worker writing output after a timeout retry. Any consumer with a fencing
or lease mechanism has probably never had an adversary for it.

### 5.4 The injection probe

Adversarial text placed in channels the domain spec marks `attacker_controlled: true` —
operator notes, command lines, log messages. Free text that is *supposed* to reach a
reasoner is the real hole; values already rejected as out-of-contract are not.

Families: direct instruction, fake authority, tool coercion, exfiltration lure, structural
(JSON/XML that may break out of a serialization), encoding (confusables, RTL overrides,
zero-width joiners), and volume.

**Pass criterion**: the presence of a probe payload changes no consumer conclusion,
compared to the same run with a benign string of equal length. A differential test with an
exact expected result, not a human reading outputs and judging.

### 5.5 The delivery ledger

Every native event is recorded with what the world emitted **and** what the sink actually
delivered: `delivered` and `delivery_reason` (`ok`, `dropped_by_perturbation`,
`duplicated`, `delayed`, `mangled`, `sink_error`).

Without this the world/perturbation separation is an assertion rather than a mechanism —
the scorer cannot compute what the consumer actually had, so it cannot tell a transport
miss from a reasoning miss. The ledger is director-side, always.

## 6. Channels, dynamics, emission

### 6.1 Channel model

Declared per [domain-spec-v0.1](../contracts/domain-spec-v0.1.schema.json): name, value
type, unit, mandatory `resolution`, cadence (with a mandatory deadband when
report-by-exception), fidelity tier, the hidden state observed, noise, drift, absence
semantics, availability, link delay, and `attacker_controlled`.

**The observation function** maps a hidden state to a reading:

```text
reading = gain·observes + offset + coef·bias_state + noise·σ_scale
```

`gain` and `offset` are the ordinary calibration of the instrument. The `observation_bias`
term is what makes a sensor able to *lie*: a hidden state — a fouling variable, a
calibration drift — corrupts the reading without touching the physical quantity, and
`σ_scale` interpolates the noise toward `noise_scale_at_full` as the bias state approaches
1. A fouled oxygen probe reads high (positive `coef`) **and** stable (`σ_scale` collapsing
toward 0.1), which is exactly why it is lethal: it looks more trustworthy as it becomes
more wrong.

Without this hook a fault that moves only a fouling state would change no emitted byte, and
the sensor-pathology profile — the case a threshold detector cannot get right, and the one
reason to prefer a reasoner — would be a silent no-op. It is not optional decoration; it is
the mechanism that gives the hardest label a physical signature.

### 6.2 Dynamics tiers

- **F0** — `baseline + trend + seasonality + noise + drift`. No state. Enough for domains
  with no physics.
- **F1** — the seven closed forms below, integrated with fixed-step RK4 at a declared `dt`
  that is part of the digest. **The default** for physical domains.
- **F2** — a named reference model from the standard library (FAO-56 Penman-Monteith,
  IEC power curve, Arrhenius spoilage, psychrometrics), each separately unit-tested
  against published vectors.
- **F3** — an FMU. Not built; §9.4.

Deliberately not an expression language. An arbitrary evaluator is a nondeterminism
surface and a security surface, and no domain in the catalog needs one. The same reasoning
governs adapter transforms.

**The F1 forms are part of the contract.** The domain spec declares a form name; the
equation it denotes is fixed here, and a name without an equation is not a specification.
Each form drives its target state `x` from a weighted input
`u = Σ cᵢ·sᵢ` over its declared `inputs` (a bare input name is `c = 1`). After every step
`x` is clamped to the form's `clamp` interval.

| Form | Equation | Required params |
|---|---|---|
| `first_order_lag` | `dx/dt = (g·u − x) / τ` | `time_constant_s` (τ), `gain` (g) |
| `rc_network` | two cascaded lags: `dm/dt = (g·u − m)/τ₁`, `dx/dt = (m − x)/τ₂` | `time_constant_s`, `time_constant_2_s` |
| `dead_time` | `x(t) = g·u(t − δ)`, then a lag of τ if τ is given | `time_constant_s`; `dead_time_s` (δ) |
| `saturation` | algebraic: `x = clamp(g·u, clamp.min, clamp.max)` | `clamp` |
| `hysteresis` | relay: `x → 1` when `u ≥ threshold_high`, `x → 0` when `u ≤ threshold_low`, holds between | `threshold_low`, `threshold_high` |
| `integrator` | `dx/dt = g·u` | `gain` |
| `threshold_integrator` | `dx/dt = g·max(0, threshold − u)` for `direction: below`; `g·max(0, u − threshold)` for `above` | `threshold`, `direction` |

`threshold_integrator` is why the aquaculture example's mortality only accrues below the
lethal oxygen level rather than integrating a healthy pond toward death — a distinction a
plain `integrator` cannot express, and one the whole deadline mechanic depends on.

`dead_time` is realized with a per-state ring buffer sized to `⌈δ / dt⌉`, so the delay is
exact in whole steps; `dt` is chosen per channel to make δ a whole multiple. The dead time
is modelled **once**, on whichever state carries it — an effector that declares its own
`dead_time_s` must not also drive a `dead_time` state, or the delay applies twice.

### 6.3 Emission

Compute true state → apply the channel's observation function → add noise and drift from
the channel's substream → quantize to `resolution` → produce a native event with
`event_time` from the world clock and `observed_time = event_time + link_delay(seed)`.
The perturbation layer then has its way with it, and the adapter renders it.

Birth bursts are modelled explicitly: on `producer_flap` recovery a producer republishes
every channel at one `observed_time`, with event times spread across the outage, and marks
them `birth: true`. That single behaviour generates a silence gap, a late-data storm, and
a same-instant tie at once, and it is completely realistic.

## 7. Sinks and adapters

Full analysis in [TRANSPORT_ANALYSIS.md](../research/TRANSPORT_ANALYSIS.md).

| Sink | Determinism | Ceiling | Built in |
|---|---|---|---|
| `inproc` | total | — | S1 |
| `file` (default) | total, byte-reproducible | offline | S1 |
| `http-push` + stepped clock | total | to be measured | S3 |
| `http-push` + wall clock | none | to be measured | S3 |
| `broker` | at-least-once | high | deferred |

Adapters are chosen independently of sinks and are pure data. `streamsim adapter verify`
checks a rendered fixture against the consumer's published schema and a committed golden
file, so an adapter is proven correct with the consumer absent.

**Sink equivalence** is a CI gate: the same `(domain, seed, command_log, adapter)` produces
byte-identical output under `inproc`, `file`, and `http-push + stepped`. `inproc` versus
`file` isolates serialization; `file` versus `http-push` isolates delivery and the
POST-withholding realization of the link-delay model, which is where a divergence is
actually likely.

## 8. The closed loop

### 8.1 Direction and authority

Commands flow **into** the simulator from a consumer, over the operator role. The
simulator's requirement is minimal and consumer-agnostic: a valid capability token and a
`command_id`. It expresses one opinion and only one — *an effector is actuated by a
deliberate, identified dispatch, not by a passing thought* — and leaves every question of
who inside the consumer was allowed to decide that entirely to the consumer.

### 8.2 Effector contract

Declared in the domain spec: argument schema, ack latency distribution, failure modes,
interlock predicate, effect (state deltas with a time constant and dead time), idempotency
window.

| Mode | Behaviour | What it tests |
|---|---|---|
| `ok` | ack, effect applies | baseline |
| `slow` | ack after a long delay | timeouts, fencing |
| `ack_lost` | effect applies, ack never returns | the ambiguous outcome — retry must not double-apply |
| `reject` | typed refusal, no effect | error taxonomy |
| `partial` | effect applies at reduced magnitude | reconciliation against reality |
| `confirmed_no_effect` | ack succeeds, world unchanged, **a confirming channel exists** | a mechanism test: two signals disagree |
| `silent_no_effect` | ack succeeds, world unchanged, **nothing confirms it but the absent outcome** | the real test |

The last two were one mode until review. Splitting them matters: with a confirming channel
present, catching a no-effect ack is a deterministic two-signal comparison and proves
nothing about outcome reconciliation. `silent_no_effect` is the version where the only
evidence is that the expected consequence did not arrive — which forces the reconciliation
window, the outcome definition and the counterfactual expectation to all be right.

An effector that declares either no-effect mode **must** declare `confirmation_channels`
(the schema enforces it): the channels that independently report whether the effect
physically happened, such as an aerator's current draw. The two modes differ only in what
those channels say:

- under `confirmed_no_effect` they keep reporting the **truth** — no current, no rotation —
  so a consumer that reads them catches the lie by direct comparison;
- under `silent_no_effect` the simulator runs a **shadow state** in which the effect *did*
  apply, and the confirmation channels report *that* — the motor draws its normal current
  while the shaft is sheared. Every emitted byte is internally consistent with success, and
  only the physical outcome the effect was supposed to produce — oxygen recovering — never
  arrives.

The shadow state is why `silent_no_effect` is a fair test rather than a trick: the evidence
does not contain a tell. A consumer reporting success under it has recorded a false
outcome, and if it learns from outcomes it has taught itself something untrue. That is the
most valuable single measurement the instrument produces.

### 8.3 Interlocks

Some domains declare an independent interlock: a predicate over hidden state that can
refuse an effector call, and an autonomous action it may take unasked. Required
behaviours, asserted:

1. Refusal is a terminal outcome, not a retryable error.
2. No alternate route attempts the same physical effect within the same episode.
3. An autonomous interlock action appears in the evidence stream as an ordinary event —
   the consumer gets no privileged knowledge of it.

### 8.4 Outcome

Effects propagate with a time constant; lubrication reduces vibration over hours, a valve
changes soil moisture over tens of minutes. The resulting evidence is ordinary evidence.
Whatever the consumer concludes about its own action, it declares in the verdict's
`outcome_believed`, and the simulator compares that against what the world actually did.

That comparison is the whole point of closing the loop, and the simulator is the only
party positioned to make it: it is an **independent observer**, which a consumer grading
its own actions structurally cannot be.

## 9. Interfaces

### 9.1 CLI

```text
streamsim catalog list | describe <domain> | coverage
streamsim domain validate <spec>
streamsim adapter list | verify <adapter>

streamsim run --domain <id> --seed <n> --adapter <id> --sink file --out trace.jsonl
streamsim replay <run.json>
streamsim verify <run.json>

streamsim scenario audit <suite>
streamsim suite generate --domain <id> --n 100 --seed <n>
streamsim score --run <run.json> --verdict verdict.json

streamsim mcp --role director | --role operator
streamsim serve --addr :7801

streamsim refconsumer --trace trace.jsonl --mcp <addr>
```

`refconsumer` is the reference consumer (§9.5).

Note what is absent: no command reads a consumer's database, and no flag names one.

### 9.2 MCP

[MCP_SURFACE.md](MCP_SURFACE.md).

### 9.3 Storage

SQLite WAL. Tables: `runs`, `command_log`, `worlds`, `world_state_history`, `emissions`
(carrying the delivery ledger), `effector_calls`, `verdicts`, `ground_truth` (sealed),
`scorecards`.

`world_state_history` records hidden state at every emission. It is what makes post-hoc
analysis possible — what was *actually* happening when the consumer said what it said —
and it is director-only, always.

### 9.4 The FMI seam, unbuilt

A channel may declare `fidelity: "F3"` with an `fmu` reference. The interface the world
core would use — `instantiate`, `setTime`, `setReal`, `doStep`, `getReal`, with the world
clock as co-simulation master and the domain's `dt` as the communication step — is
specified here and returns `not_implemented`. Cost today: this paragraph.

### 9.5 The reference consumer

`streamsim refconsumer` reads an adapter's output, applies a trivial threshold detector,
invokes effectors over MCP, and submits a verdict. A few hundred lines.

It earns its place three times over: it proves the output is consumable by something other
than its author, it is what a new consumer copies, and it lets the closed loop be
demonstrated and CI-gated with **nothing else installed** — no other product needs to
exist for this one to be finished.

## 10. Security and isolation

### 10.1 Role separation

One server, two roles, compile-time separated views. Release-blocking invariant:
everything reachable from the operator role must be computable from the delivered evidence
alone. Enforced by the differential test in
[MCP_SURFACE §5.2](MCP_SURFACE.md), not by a checklist.

### 10.2 Simulated actuation must be unmistakable

Robotics MCP servers advertise as a feature that "the agent doesn't know whether it's
controlling a simulated or a real robot." Here that is a hazard. Every operator response
carries `"simulated": true` and a `world_id`; every effector call is logged with its
world; and the simulator has no code path that reaches anything outside its own process
except a configured `http-push` endpoint and a configured `env.inject` target.

### 10.3 Everything the simulator emits is untrusted

By design (§5.4). The simulator's own MCP tool descriptions are static, reviewed and
content-addressed, so probe content is data inside a result and never instruction text in
a description.

### 10.4 Egress

No outbound connections other than the configured sink endpoint. No telemetry, no model
calls, no run-time package fetching. A domain spec and an adapter are data and cannot
cause a network call.

## 11. Failure behaviour

| Condition | Behaviour |
|---|---|
| domain spec or adapter fails validation | refuse; report path and line |
| adapter golden comparison fails | refuse; report the first divergent record |
| emission rate exceeds the sink's cap | `profile_rate_exceeded`, naming `file` |
| a sink write fails | abort the run, mark it `incomplete`; **never** silently drop, since a dropped record is indistinguishable from a modelled dropout |
| a configured sink endpoint is unreachable | back off, retain, fail the run after the deadline |
| `await_consumer` times out | `consumer_not_quiesced`; the run is marked non-reproducible |
| `truth.reveal` on an open run without `unblind` | refuse |
| `truth.reveal` with `unblind` | permit, stamp permanently, exclude from scorecards |
| the world clock is asked to move backwards | refuse |
| an effector call without `command_id` or token | refuse |
| digest mismatch on `verify` | fail loudly with the first divergent record index |

The sink-failure row is the one that matters. A test instrument that quietly drops output
under load produces exactly the class of result that is worse than no result.

## 12. Performance targets

| Target | Value |
|---|---|
| generation, `file`, F0 | ≥ 500k events/s single core |
| generation, `file`, F1 | ≥ 50k events/s single core |
| adapter rendering overhead | < 15% over raw native emission |
| memory, 1k entities × 8 channels | < 500 MB |
| a 1-hour torrent-domain trace | generates in < 60 s |
| a 100-scenario suite for one domain | < 5 min |
| golden verification of the committed suite | < 2 min in CI |

`http-push` throughput is deliberately absent: it is measured in S3, not asserted.

## 13. Technology

| Area | Choice |
|---|---|
| Language | Go 1.26 |
| Binary | one, `streamsim` |
| Persistence | SQLite WAL, `modernc.org/sqlite`, no cgo |
| Canonical JSON | RFC 8785, for the simulator's own digests |
| Schema validation | `santhosh-tekuri/jsonschema/v6`, remote resolution off |
| MCP | official Go SDK, stdio and Streamable HTTP |
| Domain specs and adapters | strict JSON, validated, compiled to typed internal form |
| Reference models | hand-written Go with published test vectors |
| Statistics and reporting | **may be Python** — see below |

Go inside the determinism boundary. The decisive argument is that a discrete-event core
cannot be vectorized, so Python's usual escape hatch fights the architecture head-on;
add one language for one engineer and a single static binary for a 24-hour soak.

The **statistics and reporting layer may be Python and probably should be**. Bootstrap
confidence intervals clustered by scenario, balanced accuracy for the trivial-baseline
panel, and the report itself are genuine statistics work — and they are post-hoc, offline,
and outside the determinism boundary. The seam: `streamsim score` emits `scorecard.json`;
an analysis script consumes it. Nothing it does can feed back into a world.

F2 reference models stay in Go: they are in the deterministic hot path, and moving them
out would put an IPC boundary inside the world core.

## 14. Definition of success

**Instrument correctness** — all must hold before the simulator judges anything:

1. Three runs of every committed scenario produce byte-identical output.
2. `verify` reproduces every committed run artifact from its seed.
3. Adding an entity, channel or fault perturbs no pre-existing entity's output.
4. The RK4 integrator agrees with the closed-form solution of a first-order lag, to within
   the channel resolution, across a swept parameter space.
5. Every F2 model matches its published test vectors.
6. A hand-computed 12-event golden vector matches, with the working committed.
7. Sink equivalence holds across `inproc`, `file`, `http-push + stepped`.
8. Every adapter passes schema and golden verification.
9. Prefix indistinguishability holds for every domain.
10. The graded suite contains no scenario a one-line detector solves.

Note that none of these requires another product to exist. That is deliberate, and it is
what "independent" means in practice.

**Instrument usefulness** — the reason to build it:

11. The closed loop runs end to end against the reference consumer: evidence → detection →
    effector → world effect → outcome → verdict → scorecard.
12. `silent_no_effect` is caught — a consumer that reports false success is detected, and
    the metric reads zero for one that does not.
13. The injection probe changes no consumer conclusion on any domain.
14. A `correlated_cascade` scenario produces the alarm-flood condition it is designed to.
15. At least three of the design questions in [DOMAIN_CATALOG §5](DOMAIN_CATALOG.md) are
    answered with evidence rather than opinion, by whichever consumer runs first.

Success is not "the simulator ran." It is that the simulator found things. A build that
produces beautiful traces and discovers nothing was a waste of the schedule.
