# Implementation Plan

Status: v3 — supersedes the plans committed at `d0fd5b6` and `ffa2803`
Date: 2026-08-11

Rewritten because the product changed shape. v1 assumed evidence might travel over MCP;
v2 assumed the simulator would be built interleaved with its first consumer. Both are
gone: **the simulator is a standalone product**, a consumer is one adapter file, and
nothing in the build plan waits on another codebase.

Old milestone names map in §8 so the two reviews stay readable.

## 0. What being standalone changes about the plan

v2's central move was a "walking skeleton of both products together," on the reasoning
that the simulator had nothing to test against because no consumer exists yet.

That reasoning was right about the problem and wrong about the fix. The fix is not to
build someone else's product; it is for the simulator to **carry its own consumer**.

`streamsim refconsumer` — a few hundred lines that read an adapter's output, threshold a
channel, invoke an effector, and submit a verdict — makes the whole chain demonstrable and
CI-gated with nothing else installed. It also happens to be the thing every future consumer
copies, so it is not scaffolding; it ships.

Three consequences:

- **No stage in this plan depends on another product.** The dependency section is empty.
- **Validation gets stronger, not weaker.** v2's stop/go was "reproduce another system's
  generator byte-for-byte," which was unachievable because that generator does not exist.
  The replacement — an analytic cross-check against a closed-form solution — is a genuine
  correctness oracle rather than a consistency check, and it needs nothing external.
- **Two stages disappear.** v2's S2 (build a stream skeleton) and its share of S3 and S6
  come out. That is roughly 23 days, of which the reference consumer costs back 3.

## 1. Owner decisions required before S0

Three. None can be resolved by more design work, and each changes the shape of what
follows. About a day of thinking; they save weeks.

### D-1 · How many domains get built?

The recommendation from [CRITICAL_REVIEW C-02](CRITICAL_REVIEW.md) is **six built, 25
specified**:

| Domain | Why this one |
|---|---|
| `rotating-machinery` | the reference; the most legible to anyone from condition monitoring |
| `host-system-health` | cheapest domain with counters, resets and saturation; mostly F0 |
| `aquaculture-pond` | the closed loop, with a hard physical deadline |
| `open-field-irrigation` | four orders of magnitude of cadence in one entity |
| `k8s-cluster` | unbounded entity churn |
| `discrete-line-oee` | no sensors; free text; the injection probe's native home |

Four exist to stress a property no cheaper domain stresses. That is the selection rule; a
seventh needs the same justification. Each additional domain is a spec plus scenario
tuning, ~2.5 days, and it *reduces* statistical power by spreading a fixed scenario budget
thinner.

### D-2 · Who is the customer of the instrument?

From [CRITICAL_REVIEW C-08](CRITICAL_REVIEW.md). "Us, testing our own runtime" needs
S1–S4 and stops. "A public conformance benchmark for stream processors" needs S5–S6 and
the blinding. "A demo" wants neither and wants a UI. The plan builds the intersection
first so this can be decided late — but it must be decided **before S5**.

Being standalone makes the middle option materially more plausible than it was, which is
worth weighing rather than assuming.

### D-3 · Can scenario authoring be separated from consumer configuration?

From [CRITICAL_REVIEW C-01](CRITICAL_REVIEW.md), the finding that compromises any graded
comparison. If the same person writes a domain's fault model and then writes the consumer
configuration that is supposed to summarize evidence from it, the consumer is being fed a
summary hand-fitted to the process under test.

Decoupling helps here in a way it did not before: the simulator now has no idea what a
consumer's configuration looks like, so the separation is structural on one side already.
The remaining risk is a single author holding both. Fixes, in descending strength:

1. **Two people.** One authors the world, another authors the consumer configuration from
   the nameplate and a nominal trace only.
2. **Time-separated authoring with a committed hash** of the consumer configuration before
   the fault list is read. Auditable, weaker.
3. **Narrow the claim** to "does a hand-authored summary help on a process we modelled,"
   which is a real and smaller result.

If this is a solo project, option 1 is unavailable and pretending otherwise is the worst
outcome.

## 2. Sequencing principle

1. **The instrument must be trustworthy before it judges.** Determinism and anti-leak
   controls land before the first graded scenario.
2. **Nothing waits on an external product.** Every gate below is achievable with only this
   repository checked out.
3. **The reference consumer arrives early**, because a producer with no consumer is
   unverifiable, and because it is the only honest test that the output is usable by
   something other than its author.

## 3. Stages

### S0 · Decisions and pre-build fixes — 1 day

The three decisions in §1, plus the fixes the reviews ranked highest:

- freeze and hash the tunable constants — SNR thresholds, negative-class fraction,
  trivial-audit cutoff, suite composition (C-07);
- split `succeeds_but_no_effect` into `confirmed_no_effect` and `silent_no_effect` — done
  in the contract, needs the scorer to treat them separately (C-05);
- change `observability` from a single `primary_channel` to a detector expression —
  `single_channel_snr`, `channel_divergence`, `peer_residual`, `conservation_residual` —
  with noise-propagation rules (G-03);
- commit **seeds, labels, audit verdicts and digests**; regenerate traces (G-07).

**Gate**: decisions recorded in `DECISIONS.md`; domain-spec schema updated; the example
domain re-validates.

### S1 · Spine — 9 days

- Go module, SQLite WAL, RFC 8785 canonical JSON.
- Domain spec loader, validator, compiler to typed internal form.
- DES core: priority queue, virtual clock, `stepped` mode, single-goroutine world.
- PRNG substream tree; the `simdet` build tag; the map-iteration lint.
- Channel model; F0 generators; F1 forms with fixed-step RK4; resolution quantization.
- Native `sim-event-v0.1` emission; `inproc` and `file` sinks.
- **Adapter engine**: the closed transform set, preamble/postamble, id rewrite,
  `adapter verify` with schema and golden comparison.
- Command log; run artifact; `replay`; `verify`.
- **Domain 1: `rotating-machinery`.**
- **Adapter 1**, plus a plain `native-jsonl` adapter as the trivial case.

**Gates** — the replacement for v2's deleted stop/go, and better because nothing external
is involved:

1. **Analytic cross-check.** A first-order lag has a closed-form solution. Implement it
   twice — RK4 and analytic — and assert agreement to within the channel resolution across
   a swept parameter space. This is a real correctness oracle for the integrator; a
   byte-equivalence gate never was.
2. **Hand-computed golden vector.** A 12-event trace where a human verified every value,
   timestamp and digest, committed with the working.
3. **Adapter conformance**: both adapters render the fixture, validate against their
   declared schemas, and byte-match their goldens.
4. Three runs byte-identical; `verify` reproduces from the artifact.
5. Adding an unrelated entity, channel or fault perturbs no pre-existing output.

Gate 1 is the one to insist on. It is the difference between "self-consistent" and
"right."

### S2 · The reference consumer and the first loop — 6 days

Earlier than anything consumer-facing was in v2, because a producer with no consumer
cannot be checked.

- MCP server, `director` role over stdio: catalog, world, clock, fault, export, run.
- MCP `operator` role: `nameplate.read`, `effector.list`, `effector.invoke`,
  `consumer.report`.
- `OperatorView` with no truth, fault-registry, perturbation-log or hidden-state
  reference — the compile-time separation.
- Effector model: argument validation against the spec, ack latency, idempotency by
  `command_id`, effect application with time constants.
- **`streamsim refconsumer`**: read, threshold, invoke, report.
- Verdict storage.

**Gates**: the reference consumer closes the loop on `rotating-machinery` — evidence →
detection → effector → world effect → observable consequence → verdict; a repeated
`command_id` applies exactly one effect; an improvised MCP session collapses to a run
artifact that `replay` reproduces with no server running.

**At the end of S2 the product works end to end.** Everything after is widening.

### S3 · Perturbations, delivery ledger, http-push — 7 days

- The full perturbation catalogue.
- **Delivery ledger** (G-05): `emissions` gains `delivered` and `delivery_reason`, so a
  transport miss is distinguishable from a reasoning miss. Without it the world/
  perturbation separation is an assertion rather than a mechanism.
- Injection probe, all seven payload families, with the benign-control differential.
- `http-push` sink, stepped and wall sub-modes, with POST withholding for link delay.
- **Measure** the `http-push` ceiling (G-12) rather than asserting it.
- Environment faults: `pause`, `kill`.
- **Domain 2: `host-system-health`.** **Domain 3: `open-field-irrigation`.**

**Gates**: sink equivalence across `inproc`, `file`, `http-push + stepped`; every
perturbation is visible in the ledger and invisible in the delivered event; the injection
probe changes no reference-consumer conclusion; `pause` produces the stale-worker
condition.

### S4 · The full loop — 8 days

- Seven effector failure modes, including both no-effect variants.
- Interlock predicates and autonomous interlock action.
- **Quiescence protocol**: `consumer.report{quiesced_through_ns}` and
  `clock.advance{await_consumer}` (G-02). Without it a closed-loop run is not
  reproducible.
- Truth store; seal and unblind stamping; prefix-indistinguishability harness.
- Streamable HTTP transport and capability tokens.
- Seeded fuzz mode: random perturbation composition, checking invariants.
- **Domain 4: `aquaculture-pond`.** **Domain 5: `discrete-line-oee`.**

**Gates**:

1. **`silent_no_effect` is caught** — the variant where only the absent outcome betrays
   it. A consumer reporting false success is detected; the metric reads zero for one that
   does not. *The highest-value single test in the plan.*
2. Interlock refusal is terminal: zero retries, zero alternate routes.
3. Prefix indistinguishability passes for every built domain.
4. `await_consumer` makes a closed-loop run byte-reproducible across three executions.

### S5 · Ground truth, scoring, reporting — 9 days

Start only once D-2 is decided.

- Observability solver for all four detector expressions; the three onset timestamps.
- Trivial-baseline audit panel with the **generate-audit-regenerate loop**, a bounded
  attempt count, and a reportable terminal state when a domain cannot produce non-trivial
  scenarios (G-06) — itself a finding.
- Suite generator: declared prevalence, randomized onset, pre-degraded starts.
- **Scorecard contract** (G-04) — a fifth schema — plus the Python analysis layer and an
  actual report. The only artifact anyone outside the project will read, and it currently
  does not exist.
- Cognition rate in operator units; flood check; the suppression self-check.
- Scrambled-label control.
- **Domain 6: `k8s-cluster`.**

**Gates**: graded suites for ≥ 2 domains with zero trivial scenarios and 35–45% negatives;
the suppression self-check lands in the 45–90% band; the scrambled-label control collapses
to chance.

The suppression self-check will probably fail first time. It is a test of the *simulator*:
failing means the world is too clean or the thresholds sit far from the operating point.
Finding that here beats finding it in a report.

### S6 · A real consumer — 5 days

The first consumer that is not the reference one. Whichever it is, the work is: write the
adapter, verify it, run the graded suite, publish the scorecard, and report which of the
predicted design questions were confirmed.

**Gate**: a scorecard for a real consumer, with every reporting rule in
[GROUND_TRUTH_AND_SCORING §7](GROUND_TRUTH_AND_SCORING.md) honoured.

## 4. Effort

| Stage | Days | Cumulative |
|---|---|---|
| S0 decisions and fixes | 1 | 1 |
| S1 spine | 9 | 10 |
| S2 reference consumer and first loop | 6 | 16 |
| S3 perturbations and delivery | 7 | 23 |
| S4 full loop | 8 | 31 |
| S5 ground truth and scoring | 9 | 40 |
| S6 a real consumer | 5 | 45 |
| **Five further domains** (2.5 days each) | **13** | **58** |

**~58 days, one engineer. Roughly twelve weeks.**

Down from v2's ~74, and the difference is almost entirely the ~23 days v2 spent building
another product, minus the 3 the reference consumer costs and plus a few for the adapter
engine.

Domains are on their own row because that is the line item v1 omitted entirely and v2
undercounted. Note the reduction from v2's 13 days: the simulator no longer owns consumer
configuration at all, so what remains on this side is domain specs and scenario tuning.
Whoever configures a consumer pays that cost, on their side, under the D-3 ruling.

### The commitment to make

Do not commit to 58 days. Commit to **S0 + S1 + S2 = 16 days**, and re-decide with a
working product.

At the end of S2 you have a deterministic seeded world, one domain, two adapters, a real
trace, a consumer reading it, an effector call changing the world, and a verdict scored
against sealed truth. That is enough to know whether the architecture is pleasant, whether
the estimates are sane, and whether the contracts survived contact.

## 5. Ordered cut list

1. `http-push + wall` sub-mode (keep stepped).
2. Environment faults beyond `pause` and `kill`.
3. The seeded fuzz mode.
4. `discrete-line-oee`, dropping to five domains — but the injection probe then has no
   native home, so this is a real loss.
5. Streamable HTTP for MCP (keep stdio).
6. The `scaled` time mode.
7. S6 entirely, if D-2 says the customer is "us."

**Never cut**: any determinism gate in S1, the analytic cross-check, the reference
consumer, the delivery ledger, the quiescence protocol, the anti-leak controls, the
injection probe, the `silent_no_effect` test, or the trivial-baseline audit. Those nine
are what distinguish a test instrument from a trace generator.

## 6. Stop / go

**After S1** — did the analytic cross-check pass, and does a human agree with the
hand-computed golden vector? If not, the integrator or the emission path is wrong and no
later result means anything.

**After S2** — the real one. Did 16 days hold, and is the thing pleasant to extend? A
large overrun here forecasts a much larger one later, and the right response is to cut
domains, not to compress S5.

**After S4** — did the loop find anything? If every gate passed first try, be suspicious
rather than pleased: check the failure modes were actually injected, the effectors
actually called, and that `silent_no_effect` was not quietly given a confirming channel
again.

## 7. External dependencies

**None.**

Every gate in S0–S5 is achievable with only this repository checked out. S6 needs a real
consumer to exist, and if none does, S6 waits — the instrument is finished and useful
without it.

## 8. Mapping from earlier plans

| v1 | v2 | v3 | Change |
|---|---|---|---|
| M0 spine | S1 | S1 | adapter engine added; gates no longer reference another product |
| M1 determinism | S1 | S1 | byte-equivalence stop/go **deleted**; analytic cross-check replaces it |
| M2 MCP + anti-leak | S4 | S2 + S4 | **one server, two roles**; evidence-over-MCP deleted; operator surface reduced to four generic tools |
| M3 perturbations | S3 | S3 | delivery ledger added |
| M4 closed loop | S4 | S2 + S4 | first loop pulled forward to S2 against the reference consumer |
| M5 ground truth | S5 | S5 | scorecard contract and audit failure path added |
| M6 remaining domains | deleted | deleted | 19 domains stay specified, unbuilt |
| — | S2 build a stream | **deleted** | replaced by the reference consumer |
| — | — | S2 reference consumer | new: the product carries its own consumer |
| — | — | S6 | a real consumer, whichever arrives first |
