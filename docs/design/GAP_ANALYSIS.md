# Gap Analysis

Status: review of the design as committed at `d0fd5b6`
Date: 2026-08-11
Scope: what the design does not say, taking its own goals as given

This pass accepts the design's premises and asks only whether it is *complete enough to
build*. The separate [CRITICAL_REVIEW.md](CRITICAL_REVIEW.md) attacks the premises.

Severity: **G-blocker** means M0 should not start until it is resolved. **G-major** means
it will be discovered painfully mid-build. **G-minor** means it will cost an afternoon.

Twelve gaps. Three blockers.

---

## G-01 · The 25 Situation Models are an unscheduled deliverable — **blocker**

`streamsim domain scaffold-model` deliberately emits only inputs, units and enums, and
deliberately does *not* author thresholds, states, hysteresis, sustained durations,
triggers, cooldowns or Decision schemas — because authoring those is the work under
evaluation, and generating them would be the simulator grading its own homework.

That reasoning is right, and it means **every domain needs a hand-authored Situation
Model that appears nowhere in the plan**. The plan lists domains as one-line bullets
("Domain 21 `host-system-health`") as if the domain spec were the whole job. It is
roughly half of it.

Realistic cost: a Situation Model with defensible bands, a proven `enter ⇒ stay`
property, tuned `sustained_for` and cooldowns, and a Decision schema is 1–3 days per
domain. For the 13 domains in M1–M5 that is **13–39 days against a 30-day plan**.

The plan is understated by something between 50% and 130%, and it is understated on the
one axis that cannot be cut, because a domain without a Situation Model produces no
Situations and therefore no result.

**Resolve before M0** by either: scoping to 3–5 domains and saying so; or splitting
"domain spec" and "situation model" into separate, separately estimated work items with
an explicit author; or accepting scaffolded models for mechanism-only domains and
hand-authoring only for graded ones — which is probably right, and needs the distinction
made explicit in the catalog.

## G-02 · No barrier between `clock.advance` and the system under test — **blocker**

In `stepped` mode `clock.advance` returns once every emission is written to every sink.
That is well-defined for the *simulator*. It says nothing about the **stream and Tamoz**,
which are separate processes that then have to ingest, evaluate, admit a trigger, call a
model, validate a Decision, run the policy pipeline, and dispatch a command.

The harness loop in [MCP_SURFACE §5.1](MCP_SURFACE.md) shows
`clock.advance(24h) → … → effector call → clock.advance(2h)` as if the middle step were
synchronous. It is not. Nothing in the design says how the harness knows the system under
test has quiesced.

Consequences if unresolved:

- the effector call lands at an arbitrary world time, so the effect propagates from the
  wrong instant and the run is not reproducible after all;
- or the harness sleeps, which reintroduces wall-clock dependence into the one place the
  entire determinism architecture was built to exclude;
- or `clock.advance` is called before reasoning completes and the episode is superseded
  by evidence it should have seen — producing a "failure" that is a harness artifact.

**This is the largest hole in the closed loop**, and the closed loop is the design's main
justification for existing.

What is needed: a **quiescence protocol**. The stream must expose "I have processed every
record up to recorded time T and have no pending episodes, timers, or outbox entries," and
`clock.advance` must block on it. That is a real contract, it belongs in
[CONSUMERS.md](CONSUMERS.md) as a consumer-optional capability, and it is the kind of
thing that is straightforward to design now and miserable to retrofit.

## G-03 · `first_observable_time` is single-channel, and the important faults are not — **blocker**

`observability` declares one `primary_channel`, and both onset timestamps are computed as
an SNR crossing against that channel's noise σ.

This is wrong for the faults the catalog is proudest of:

- `do_probe_fouling` — the probe reads a *plausible* value. Its single-channel SNR never
  crosses anything. Detectability comes from divergence against the diurnal expectation
  and against sibling ponds.
- `algae_bloom_crash` — signal is spread across dissolved oxygen, pH and ammonia; the
  earliest detection is joint, and any single channel understates it.
- `sensor_drift` (domain 07) — the definition of the fault is that one channel moves while
  correlated channels do not. The quantity is a *correlation residual*, not a level.
- `background_leak_growth` (domain 15) — visible only in a conservation law across
  entities. No channel shows it at all.

Two consequences, and the second is worse than the first:

1. The timestamp is confidently wrong, so detection latency is measured against a
   fictional reference.
2. Because it is wrong in the **too-early** direction (a single-channel SNR crossing that
   never happens is computed as "never," but a spurious crossing on a noisy channel is
   computed as "early"), correct detections get classified `suspicious` and trigger
   phantom leak investigations. **The control designed to catch cheating will generate
   false accusations, on exactly the domains where reasoning matters most.**

Needs: `observability` must accept a **detector expression** rather than a channel —
minimally `single_channel_snr`, `channel_divergence(a, b)`, `peer_residual(channel)`, and
`conservation_residual(inputs, outputs)` — each with a defined noise-propagation rule so
the SNR is computable. That is a schema change and a solver change, and it is cheaper now
than after 13 domains are authored against the wrong contract.

---

## G-04 · No scorecard contract — **major**

Three contracts exist (domain spec, run artifact, ground truth). `sim.score` returns "a
scorecard," which is defined nowhere. [GROUND_TRUTH_AND_SCORING §3](GROUND_TRUTH_AND_SCORING.md)
lists three metric classes and about thirty metrics in prose, with no schema, no units, no
aggregation rule, and no statement of which are per-scenario and which are per-suite.

This matters more than a missing schema usually does, because the scorecard is the
**only artifact anyone outside the project will ever read**. It is the product of the
instrument. It should be the most carefully specified contract in the repository and it
is the least.

Also missing: any report format. There is no CLI output spec, no HTML or Markdown report,
no worked example of what a result looks like. §7's reporting rules ("every table carries
the negative-class fraction…") describe tables that no code is specified to produce.

## G-05 · No delivery ledger, so the transport/reasoning distinction is unenforceable — **major**

[TECHNICAL_DESIGN §4](TECHNICAL_DESIGN.md) makes a good argument for placing the
perturbation layer between world and sink: the world produces what happened, the
perturbation layer produces what the observer got, and a scenario the runtime never saw
should be scored as a transport miss rather than a reasoning miss.

Nothing implements that. The `emissions` table records what the world emitted. There is no
record of what the sink actually **delivered**, so the scorer cannot compute the
difference, and the distinction the architecture was arranged to make available is not
actually available.

Needs: `emissions` gains `delivered`, `delivery_reason` (`ok`, `dropped_by_chaos`,
`suppressed_late`, `sink_error`) and the sequence position at which it was written; the
scorer joins on it; and every miss in the report is attributed.

## G-06 · The trivial-baseline audit has no failure path — **major**

The audit gates scenarios into the graded suite. The suite declares a 100-scenario
minimum. Nothing connects them. If the audit rejects 80% of generated scenarios — which is
plausible, and for `host-system-health` is close to certain — the generator has no defined
behaviour: it does not regenerate, it does not report a shortfall, and the suite silently
falls below its own declared floor.

Needs: a generate-audit-regenerate loop with a bounded attempt count, a per-label quota,
and an explicit terminal state. "This domain cannot produce non-trivial scenarios at the
declared prevalence" must be a **reportable finding** — it is a genuinely interesting
result about the domain — rather than an empty directory.

## G-07 · Committed traces will not fit in git — **major**

[GROUND_TRUTH_AND_SCORING §6](GROUND_TRUTH_AND_SCORING.md) requires suites "committed as
JSONL plus `labels.jsonl` plus the seed" and, two lines later, "regenerable byte-for-byte
from the seed."

Those are in tension, and the arithmetic settles it. `k8s-cluster` at torrent rate for one
hour is millions of events; 100 such scenarios is tens of gigabytes. Even `rotating-
machinery` at 100 scenarios × 72 hours is substantial. Twenty-five domains of this is not
a repository.

Resolution is obvious once stated: **commit seeds, labels, audit verdicts and digests;
regenerate traces**. That is what the byte-reproducibility guarantee is *for*. But it
must be said, because the current text says the opposite, and because it makes the
digest-verification gate load-bearing rather than decorative.

## G-08 · No cost model for the graded suite — **major**

The full sweep implied by the design is 25 domains × 100 scenarios × 3 arms × 3
repetitions ≈ 22,500 model calls, plus whatever multi-turn reasoning each episode
involves. There is no budget, no per-domain cost estimate, no sampling strategy, and no
statement of which domains get graded suites versus mechanism-only treatment.

`sim.score` reports "cost per resolved scenario" as a loop metric, so the design knows
cost exists. It has just never added it up.

The likely correct answer — graded suites for 2–3 domains, mechanism-only for the rest —
is also the answer to G-01, and the two should be decided together.

## G-09 · Domain specs have no versioning or migration story — **major**

The domain digest is part of the run artifact and part of `verify`. Domain specs will
change constantly during early development. Every change invalidates every committed run
artifact for that domain, and nothing in the design detects, reports, or migrates that.

Within a month the repository accumulates fixtures that fail `verify` for reasons nobody
can reconstruct, and the team learns to ignore `verify` failures — which destroys the
gate.

Needs: `verify` must distinguish "digest mismatch because the domain changed" (report the
version delta; offer to re-baseline) from "digest mismatch on an unchanged domain" (a real
defect, fail loudly). Plus a deprecation policy for artifacts whose domain version no
longer exists.

---

## G-10 · `sim.score` couples to the stream's SQLite schema — **minor**

`sim.score {run_id, stream_db}` reads the stream's database directly, which couples the
simulator to a private storage schema that the stream is free to change. The stream
already exposes inspection APIs and `explain`. Read through those, or declare the storage
coupling and version it.

## G-11 · `correlated_cascade` is mandatory for domains where it is fiction — **minor**

Every domain must offer `nominal`, `correlated_cascade` and `sensor_pathology`. For
`cicd-pipeline-health` there is no physical coupling and no root cause that trips many
entities at once; for `grain-storage` a cascade is a stretch. Mandating it produces
invented scenarios, which is worse than an honest absence.

Make it conditional on the domain declaring `cascading` or `spatial` correlation, and
require an explicit `not_applicable` with a reason otherwise.

## G-12 · `http-live` throughput is asserted, not measured — **minor**

"~2–5k eps on loopback" appears three times as a ceiling and was never measured. One POST
per event with no batching (the stream deliberately excludes batching) against 3,000
entities in `livestock-herd-health` is a lot of sequential HTTP, and the performance
targets in §12 cover *generation* only, not delivery.

Cheap fix: measure it in M0 with a stub, before three later milestones depend on the
number.

---

## Summary

| # | Gap | Severity | Costs most if found in |
|---|---|---|---|
| G-01 | 25 Situation Models unscheduled | blocker | M5 |
| G-02 | No quiescence barrier for the closed loop | blocker | M4 |
| G-03 | Single-channel observability model | blocker | M5 |
| G-04 | No scorecard contract or report format | major | M5 |
| G-05 | No delivery ledger | major | M5 |
| G-06 | Trivial-audit has no failure path | major | M5 |
| G-07 | Committed traces exceed git | major | M2 |
| G-08 | No cost model | major | M5 |
| G-09 | No domain-spec versioning | major | M2 |
| G-10 | Scorer couples to stream storage | minor | M5 |
| G-11 | Mandatory cascade profile | minor | M1 |
| G-12 | Unmeasured `http-live` ceiling | minor | M3 |

Six of the twelve first bite in M5, which is the last milestone and the one carrying the
evaluation. That clustering is itself the finding: **the design is thorough about the
mechanism and thin about the measurement**, which is the same imbalance the V0.1 critique
levelled at V0 — spending the schedule on instrument and arriving at the experiment with
nothing left.

The three blockers are cheap now and expensive later, and none requires a decision about
scope. Fix them before M0.

---

## Disposition after the decoupling revision

This review was written against `d0fd5b6`, before the simulator was made standalone. Three
findings changed status; the rest stand and are folded into stage S0 of the plan.

| # | Status |
|---|---|
| **G-01** 25 Situation Models unscheduled | **Reduced, not resolved.** The simulator no longer owns consumer configuration at all — that cost moved to whoever configures a consumer. What remains on this side is six domain specs plus scenario tuning. The underlying error, a deliverable that is half of each domain's work appearing as a one-line bullet, is now costed in the plan's own effort table. |
| **G-02** No quiescence barrier | **Open, designed.** Now `consumer.report{quiesced_through_ns}` plus `clock.advance{await_consumer}`, and it is a stated consumer-optional capability rather than a change demanded of one product. Built in S4. |
| **G-03** Single-channel observability | **Open, scheduled in S0.** Unaffected by the decoupling. |
| **G-04** No scorecard contract | **Open, scheduled in S5.** Partially addressed: `consumer-verdict-v0.1` now defines the scorer's *input* precisely, which was previously as vague as its output. |
| **G-05** No delivery ledger | **Open, scheduled in S3**, and load-bearing for the consumer metrics in the scoring design. |
| **G-06** Trivial-audit failure path | **Open, scheduled in S5.** |
| **G-07** Committed traces exceed git | **Open, scheduled in S0.** |
| **G-08** No cost model | **Reduced.** Six built domains rather than 25 cuts the graded sweep by roughly three quarters. Still needs a number before S5. |
| **G-09** No domain-spec versioning | **Open**, and now slightly larger: adapters are versioned artifacts too, and an adapter digest is part of the determinism tuple. |
| **G-10** Scorer couples to a consumer's storage | **RESOLVED.** The scorer reads a submitted `consumer-verdict`, sealed truth, the delivery ledger and the effector log. It has no access to any consumer's storage and no code path that could acquire one. |
| **G-11** Mandatory cascade profile | **Open, minor.** |
| **G-12** Unmeasured push throughput | **Open, scheduled in S3**, and the number was removed from the performance targets rather than left as an assertion. |

One gap the decoupling *created*, worth recording here rather than pretending the revision
was free:

**G-13 · Adapter correctness is a new failure surface.** A wrong adapter silently produces
plausible but incorrect input for a consumer, and every result downstream is then a
measurement of the adapter. Mitigated by `adapter verify` — schema validation plus a
byte-compared golden fixture, both required in CI — but the mitigation is only as good as
the vendored consumer schema it checks against, and a consumer that changes its contract
without telling us breaks it silently. Adapters need the same version-pinning discipline
as domain specs (G-09).
