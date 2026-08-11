# Streams Simulator

Status: Level 1 replayable prototype; not benchmark-ready
Date: 2026-08-11
Codename: `streamsim`

> **Implementation review:** the correctness spine and replayable-prototype bar are green,
> but the benchmark-release gates are not yet satisfied. Do not use current scores as
> benchmark evidence. See the
> [implementation-readiness review](IMPLEMENTATION_READINESS_REVIEW.md) for confirmed gaps,
> acceptance criteria, and the recommended implementation sequence.
>
> The executable completion bar and per-change remediation loop are defined in
> [QUALITY_BAR.md](QUALITY_BAR.md).

A standalone, deterministic, closed-loop world simulator for testing stream processors.

## 1. What it is

It generates realistic event streams from data-defined domains, corrupts their delivery in
specified ways, carries sealed ground truth, accepts commands back through declared
effectors so the world actually changes, and scores whatever consumed it.

**It is an independent product with no knowledge of its consumers.** No consumer's name,
schema, field, or behaviour appears in the binary. A consumer is a file — an
[output adapter](contracts/output-adapter-v0.1.schema.json) — plus, optionally, an MCP
client that invokes effectors and submits a verdict.

It is also a **test instrument**, and that word constrains everything. An instrument less
trustworthy than the system it measures is worse than no instrument, because it produces
confident wrong answers. So its determinism guarantees are strong, its ground truth is
generative rather than annotated, and its correctness is proven against analytic oracles
before it judges anything.

## 2. Three decisions that shape everything

### MCP manages the simulator. It does not transmit the streams.

MCP names, seeds, steps, perturbs, actuates and audits a run. Evidence leaves on a sink.

MCP's server-to-client push is best-effort by specification — the server *may* close a
stream at any time, *must not* re-broadcast a message, and resumability is a *may*. It has
no back-pressure. It makes delivery timing a function of TCP, which destroys
reproducibility. And its consumer is an LLM host, which would put raw producer-controlled
strings one hop from a prompt — the exact property the injection probe exists to test.

A large share of this instrument's value is in *absence*: a producer going quiet, a
heartbeat stopping. A transport where a dropped message is a blessed outcome would make
every absence finding permanently ambiguous. The instrument would be injecting its own
worst failure mode into itself, undeclared.

Meanwhile everything that is *not* an event — create a world, advance six hours, inject a
fault at T+2h, **invoke an effector** — is naturally request/response with an ack and an
idempotency key. That is the textbook shape MCP was built for.

### The simulator emits its own format. Consumers are adapters.

Events are `sim-event-v0.1`, the simulator's native envelope, projected into a consumer's
wire format by a declarative adapter. Adapters are **data**, loaded through one path,
exactly like domain specs.

Emitting the first consumer's format natively would have been fewer moving parts on day
one and wrong three ways: it embeds one consumer's decisions in the core, it corrupts the
domain model upward as specs start being written to satisfy a client, and it makes that
consumer's contract untestable — nothing proves the format is a *projection* rather than
an *assumption*.

### One MCP server, two roles, and no domain knowledge in either.

`director` builds and drives worlds and holds the truth. `operator` is what a consumer
gets: read the nameplate, list effectors, invoke one, submit a verdict. Four tools.

`sim.effector.invoke` takes an effector *name* validated against the loaded domain spec.
The binary contains no effector name, so twenty-five domains and the twenty-sixth all work
through the same tool. An earlier draft generated one MCP tool per effector — domain
knowledge reaching the protocol surface, and a tool catalog that mutates whenever a world
is created.

## 3. Shape

```text
        ┌──────────────── MCP: director role ─────────────────┐
 harness│ catalog · world · clock · fault · perturb · truth    │
        └──────────────────────┬──────────────────────────────┘
                               │ command log
   domain spec ───────────────►│
   (data)                      ▼
                      ┌──────────────────┐
                      │   world core     │  seeded · discrete-event
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
                          └──────────┘
                    MCP: operator role
            effector.invoke · consumer.report
```

Two placements are deliberate. **Perturbation sits between world and adapter** — the world
produces what physically happened, the perturbation layer produces what the observer got —
so a scenario the consumer never saw is scored as a transport miss, not a reasoning miss.
**The adapter sits last**, so every consumer sees the same perturbation and cross-consumer
comparison means something.

## 4. What is new here

### The loop closes, and that enables one test nothing else can run

An effector call changes hidden state; the effect propagates with a physical time
constant; the consequence appears in ordinary evidence. The simulator is therefore an
**independent observer**, which a consumer grading its own actions structurally cannot be.

> **`silent_no_effect`** — the effector acknowledges success, the world does not change,
> and *nothing confirms it but the absent outcome*. A consumer that trusts acknowledgements
> reports success. If it also learns from outcomes, it has just taught itself something
> untrue.

The confirmed variant, where a second channel contradicts the ack, is a deterministic
two-signal comparison and proves nothing. Both are injectable and scored separately.

### The benchmark is designed not to flatter its author

Wu & Keogh showed the standard anomaly-detection benchmarks share four flaws severe enough
to make much apparent progress illusory. Each has a countermeasure here; the sharpest is
the **trivial-baseline audit** — a scenario is run against one-line detectors *fitted with
hindsight on the scenario itself*, and if one wins, the scenario cannot enter the graded
suite. It will shrink the suite dramatically on first run, and the correct response is
relief.

### Exploration becomes a regression test

Every MCP mutation appends to a command log, so an improvised session collapses into one
run artifact that `streamsim replay` reproduces exactly with no server running.

Worth being precise: that is a property of the **command log**, not of MCP. A CLI writing
the same log would earn it too. MCP's contribution is that the improvisation can be
conversational.

## 5. The 25 domains

A coverage matrix over twelve axes of stream physics, not a list of industries. Each
domain declares its property vector and one required sentence naming the runtime property
it exists to break. Two domains with the same vector means one is decoration.

| Group | Domains |
|---|---|
| Agriculture, food, environment | greenhouse-climate · open-field-irrigation · livestock-herd-health · aquaculture-pond · grain-storage · cold-chain-transit |
| Manufacturing and process | rotating-machinery · cnc-tool-wear · discrete-line-oee · injection-molding-spc · bioreactor-batch |
| Health, energy, utilities | patient-vitals-rpm · solar-pv-plant · bess-thermal · water-distribution · hvac-building |
| Facilities, mobility, logistics | datacenter-power-cooling · ev-fleet-telematics · rail-trackside · wind-turbine |
| Digital systems | host-system-health · k8s-cluster · service-slo · cicd-pipeline-health |
| Adversarial | auth-and-edr |

A domain is **data**. All 25 load through one schema with no per-domain code. If any needs
a code branch in the binary, the simulator is wrong and the domain found the bug.

**Six get built**, per [CRITICAL_REVIEW C-02](design/CRITICAL_REVIEW.md); the rest stay
specified. The catalog's value — eight design questions about the first consumer, four
probable defects — was collected by *writing* it, and building the other 19 costs 20–40
days while spreading a fixed scenario budget thinner.

## 6. Non-negotiables

Nine controls separate a test instrument from a trace generator. The cut list may remove
anything else; it may never remove these.

1. **Determinism.** A run is a pure function of `(sim_version, domain_digest,
   adapter_digest, seed, command_log, sink)`. One artifact reproduces any failure.
2. **The analytic cross-check.** A first-order lag has a closed form; implement the
   integrator twice and assert agreement. A real correctness oracle, not a consistency
   check.
3. **The reference consumer.** The product must be complete on its own.
4. **The delivery ledger.** Without it, transport misses and reasoning misses are
   indistinguishable and the perturbation layer's placement is an assertion.
5. **The quiescence barrier.** Without it a closed-loop run is not reproducible.
6. **A sealed oracle.** Compile-time role separation plus a differential
   prefix-indistinguishability test, because checklists miss timing and error-string side
   channels.
7. **The injection probe.** Adversarial text in admitted free-text channels must change no
   consumer conclusion, byte for byte against a benign control.
8. **The `silent_no_effect` test.** Rate zero.
9. **The trivial-baseline audit.** No scenario a one-line detector solves enters the
   graded suite.

## 7. Documents

| Document | Contents |
|---|---|
| [research/TRANSPORT_ANALYSIS.md](research/TRANSPORT_ANALYSIS.md) | Why MCP carries control and actuation but not evidence; why the simulator emits its own format; sinks, receipt timestamps, role separation |
| [research/PRIOR_ART.md](research/PRIOR_ART.md) | Deterministic simulation testing, FMI 3.0, Sparkplug B, ISO 13374, ISA-18.2 / EEMUA 191, the anomaly-benchmark critique — and the twenty patterns adopted |
| [design/DOMAIN_CATALOG.md](design/DOMAIN_CATALOG.md) | Twelve property axes, 25 domains, coverage proof, build order, eight design questions, what was cut |
| [design/TECHNICAL_DESIGN.md](design/TECHNICAL_DESIGN.md) | Determinism, three injection surfaces, channels and dynamics tiers, sinks and adapters, the closed loop, security, success criteria |
| [design/MCP_SURFACE.md](design/MCP_SURFACE.md) | One server, two roles, four operator tools, separation enforcement, error taxonomy |
| [design/CONSUMERS.md](design/CONSUMERS.md) | The integration contract; the reference consumer; the first consumer as a worked example |
| [design/GROUND_TRUTH_AND_SCORING.md](design/GROUND_TRUTH_AND_SCORING.md) | Four benchmark flaws and their countermeasures, three onset timestamps, instrument vs consumer metrics, leak detection |
| [design/IMPLEMENTATION_PLAN.md](design/IMPLEMENTATION_PLAN.md) | Seven stages, ~58 days, gates, cut list, stop/go, **zero external dependencies** |
| [design/GAP_ANALYSIS.md](design/GAP_ANALYSIS.md) | Twelve gaps found by taking the design's own goals as given |
| [design/CRITICAL_REVIEW.md](design/CRITICAL_REVIEW.md) | Eight findings attacking the premises |
| [contracts/](contracts/) | `sim-event`, `output-adapter`, `consumer-verdict`, `domain-spec`, `run-artifact`, `ground-truth` |
| [adapters/](adapters/) | One adapter per consumer. Deleting one changes no simulator behaviour. |
| [examples/](examples/) | `aquaculture-pond`, the closed-loop showcase domain |

## 8. Technology

Go 1.26, one binary, SQLite WAL, RFC 8785 canonical JSON, the official MCP Go SDK.

Go inside the determinism boundary. The decisive argument is not shared code with any
consumer — the simulator digests only its own artifacts — but that **a discrete-event core
cannot be vectorized**, so Python's usual escape hatch fights the architecture head-on.
Add one language for one engineer, and a single static binary for a 24-hour soak.

**The statistics and reporting layer may be Python**, and probably should be: bootstrap
confidence intervals, balanced accuracy, and the report itself are real statistics work,
and they are post-hoc, offline, and outside the determinism boundary. `streamsim score`
emits `scorecard.json`; an analysis script consumes it.

No physics engine, no broker, no ORM, no expression language, no plugin system — a plugin
is consumer-specific code renamed.

## 9. Review status

Two adversarial passes are recorded in the repository. Their three blockers are folded
into stage S0 of the plan; two of the twelve gaps were resolved outright by the decoupling
(see the disposition notes at the end of each review).

The finding that still needs an owner decision:

> **D-3 · Any graded comparison is compromised if one person writes both a domain's fault
> model and the consumer configuration meant to summarize evidence from it.** The consumer
> then receives a summary hand-fitted to the process under test. Decoupling made the
> separation structural on the simulator's side; the remaining risk is a single author
> holding both halves. If this is a solo project the strong fix is unavailable, and
> pretending otherwise is the worst option.

## 10. Start here

[TRANSPORT_ANALYSIS.md](research/TRANSPORT_ANALYSIS.md) — the two decisions in §1
determine everything downstream. Then
[DOMAIN_CATALOG.md §5](design/DOMAIN_CATALOG.md) for the eight design questions, which are
the reason to build this rather than a reason to admire it. Then both reviews, then the
plan.

Begin with S0: three decisions and half a day of fixes. Commit to S0–S2, sixteen days, and
re-decide with a working product — at the end of S2 the loop closes against the reference
consumer, which is the whole thing in miniature.

Do not begin with a broker, a web UI, a physics engine, or a second consumer.
