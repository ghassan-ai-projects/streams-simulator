# Gaps between today and world-class

- **Date**: 2026-10-11
- **Frame**: a world-class test instrument has five properties. Each section names the gap,
  the evidence, and the cost of leaving it.

| Pillar | Question a buyer asks | Today |
| --- | --- | --- |
| 1. Trust | "Is the instrument right, everywhere?" | Strong; two verification gaps |
| 2. Proof | "Did this run show the product beats alternatives?" | Missing |
| 3. Coverage | "Does it span the worlds I operate in?" | Eight domains, 14 thin cells |
| 4. Reach | "Can my team use it this week?" | CLI + MCP; no authoring kit, release or CI recipe |
| 5. Operability | "Can I pin it, cite it, and reproduce it in a year?" | Contracts v0.1, no release stream |

## 1. Trust

### 1.1 The register and the docs disagree with the code

The code is ahead of its documentation (see [ASSESSMENT.md](ASSESSMENT.md)). An instrument
whose own records are stale invites doubt about the parts that are right. Cost: low to fix,
high to ignore.

### 1.2 Cross-architecture determinism is unproven

A run is promised to be a pure function of its inputs. The dynamics use floating-point
integration (RK4). Go on arm64 may fuse multiply-add, which can change low-order bits
relative to amd64. If a digest includes any float formatted at full precision, the same seed
could produce different traces on a developer's Apple-silicon laptop and a Linux amd64 CI
runner. This is a **risk to verify, not a known defect**: the repository has a `simdet`
build tag and canonical number rules, which may already neutralise it. Verify with a CI
matrix (darwin/arm64, linux/amd64, linux/arm64) comparing `scripts/behaviour-pin` output.
If it diverges, the fix is explicit rounding or `float64()` casts at digest-bearing seams,
recorded as a design decision because it moves digests.

### 1.3 Oracle independence is one-deep

The analytic cross-check implements the integrator twice. Good. Strengthen with:

- property tests that generate random domain specs within schema bounds and assert
  replay-equality, ledger conservation (`emitted = delivered + dropped + in-flight`),
  and solver agreement;
- fuzzing of the adapter expression language, the JSON Schema compiler and canonical JSON
  (three panics were found by reading; fuzzing finds the rest);
- mutation testing of the scorer and solver: mutate a comparison, expect a test to fail.

### 1.4 Domain data quality has one owner and no gate

D-52 shows a domain can declare an unobservable fault and still pass. A **domain linter**
should fail a domain where a fault's detector channel is not affected by the fault's
`affects` states, where an effector's `confirmation_channels` are not influenced by its
effect, or where an expected effector cannot reach the faulted state. This is data
validation, so it stays consumer-agnostic and belongs in `domain validate`.

## 2. Proof

This is the largest gap. See [DESIGN-outcome-economics.md](DESIGN-outcome-economics.md) and
[DESIGN-baseline-ladder.md](DESIGN-baseline-ladder.md) for the proposals.

| Missing | Today | Consequence |
| --- | --- | --- |
| Outcome in money, downtime, safety | `consequence` is an axis tag; `counterfactual.if_no_action` is a 200-character label | Results cannot be summed, compared or priced |
| Regret against a bound | Loop metrics are booleans and one time-to-resolution | "Good" has no ceiling; every score is relative to nothing |
| Unsafe-action count | `silent_no_effect_calls`, `unnecessary_action` | A harmful action in a hazardous state is not distinguished from a wasted one |
| Strong baseline | Reference consumer is a threshold detector; audit removes *too easy* scenarios | Nothing removes scenarios where a good rules engine already ties the product |
| Held-out evaluation | `min_scenarios: 100`, seeds are public | A consumer can be tuned to the suite; the result is attackable |
| Statistical protocol | Per-run scorecards | No paired comparison, confidence interval, or win rate across seeds |
| Cost accounting | `docs/design` §4 defines cognition rate in operator units | Not a scorer field; model calls and dollars per correct outcome are unmeasured |

## 3. Coverage

From `streamsim catalog coverage` across eight domains:

| Axis | Observed | Not exercised |
| --- | --- | --- |
| `time_reference` | all eight are `wall` | event-time vs processing-time semantics, the defining problem of stream processing |
| `rate` | low 2, medium 5, torrent 1 | `sparse`, `high` |
| `cardinality` | small 2, fleet 5, churning 1 | `singleton`, `large` |
| `cadence` | periodic 7, event-driven 7, exception 1 | `batch`, `human_driven` |
| `value_shape` | scalar 7, enum 7, counter 5, boolean 1, text 1 | `distribution` |
| `fidelity` | F0 3, F1 6 | F2, F3 |
| `consequence` | cost 4, downtime 3, regulatory 1, safety 1 | safety and regulatory are one domain each |
| `actuation` | advisory 4, fast 2, interlocked 1, slow 1 | `observe_only` |

Fourteen cells are "thin" (one domain). The product-proof review's frontier list (sparse
rate, large cardinality, distributions, adversarial-adaptive faults) matches these holes.
Coverage should grow **only** where a hole is needed by a proof scenario
([ROADMAP.md](ROADMAP.md) phase 4), not to raise a count.

Structural capability gaps behind several holes:

- **Entity coupling.** `correlation: cascading` is declared by two domains, but a fault on
  one entity cannot propagate through declared relations to others. Ground truth has no
  `root_entity`. This blocks the root-cause scenario.
- **Effector-triggered observation.** An effector cannot cause a channel to emit a reading of
  hidden state after a delay. This blocks investigation and probe-cost scenarios.
- **Adaptive adversary.** Injection corpora are fixed; an attacker cannot react to the
  consumer's actions.

## 4. Reach

| Gap | Why it limits adoption |
| --- | --- |
| Two adapters (`agentic-stream`, `native-jsonl`) | Teams on OpenTelemetry, CloudEvents, or Prometheus must write adapters before first use |
| No domain-authoring kit | Domains are the unit of value; authoring is JSON by hand against a 450-line schema |
| No CI recipe or GitHub Action | The natural use is a regression gate on a consumer's pull requests |
| No container image or install path beyond `go build` | Friction for non-Go teams |
| No scorecard diff | "Did my change make it worse?" needs a comparator across two runs of one suite |
| No hosted example | A skeptical reader cannot see a scorecard without installing |

## 5. Operability and governance

- **No release stream.** Version 0.2.0 exists in code; there is no tag, signed artifact,
  SBOM or checksum, and `CHANGELOG.md` says so.
- **Contracts are `v0.1`.** Domain, adapter, event, verdict, ground-truth and run-artifact
  schemas are pre-1.0, so external consumers have no stability promise.
- **Digest-bearing changes need a policy.** Hardening moved digests deliberately. A world-class
  instrument states what changes require a new `sim_version`, what migration is offered, and
  how an old artifact is replayed.
- **Release evidence is historical.** The manifest pins an older commit.
- **Process cost is outrunning output** in the idea portfolio: about 3,000 lines of
  evaluation prose and no admitted artifact. Judge the next round by runs, not documents.
