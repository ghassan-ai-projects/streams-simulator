# Ground Truth and Scoring

Status: design
Date: 2026-08-11

A simulator that generates its own benchmark will produce a flattering one unless it is
designed not to. This document is the set of countermeasures.

The threat is documented: Wu & Keogh found that the standard anomaly-detection
benchmarks — Yahoo, NAB, NASA — share four flaws (triviality, unrealistic anomaly
density, mislabeled ground truth, run-to-failure bias) severe enough that "much of the
apparent progress in recent years may be illusionary." We will reproduce all four by
default if we do nothing. See [PRIOR_ART.md §5](../research/PRIOR_ART.md).

## 1. What ground truth is here

Generative, not annotated. The label is what we injected, so there is no annotator to
disagree with. That removes one whole class of error and creates a subtler one, addressed
in §2.

A ground-truth record:

```json
{
  "scenario_id": "aquaculture-pond/0042",
  "domain": "aquaculture-pond", "seed": 42,
  "entity_id": "site-a/pond-3",
  "label": "do_probe_fouling",
  "expected_episode": true,
  "injection_time_ns":         1785315600000000000,
  "first_observable_time_ns":  1785319140000000000,
  "unavoidable_time_ns":       1785326400000000000,
  "observability": {"detector_form": "peer_residual", "channels": ["pond.dissolved_oxygen"], "effective_sigma": 0.11, "first_observable_snr": 3.0, "unavoidable_snr": 10.0},
  "deadline_ns":               1785333600000000000,
  "expected_effector": "start_aerator",
  "counterfactual": {"if_no_action": "stock_loss", "if_action_by_deadline": "recovery"},
  "trivial_baseline_verdict": "non_trivial",
  "perturbations": ["duplicate_burst@0.02", "reorder_within_allowance@30s"]
}
```

## 2. The four flaws, and the four countermeasures

### 2.1 Triviality → the trivial-baseline audit

**The flaw**: a benchmark where a one-line detector wins measures nothing about
reasoning.

**The countermeasure**: every candidate scenario is run against a panel of trivial
detectors before it may enter the graded suite.

| Detector | Definition |
|---|---|
| `fixed_threshold` | best single threshold on any channel, chosen with hindsight |
| `zscore` | \|x − μ\| > kσ over a trailing window, best k |
| `first_difference` | largest single-step change |
| `moving_median_residual` | residual from a trailing median, best window |
| `channel_silence` | any channel quiet for > T |

Each is fitted **with hindsight on the scenario itself** — deliberately unfair to the
scenario. If a detector tuned on the answer still cannot separate the fault from the
`transient_none` controls at ≥ 0.9 balanced accuracy, the scenario is `non_trivial`.

`streamsim scenario audit` runs this and refuses to admit a `trivial` scenario to the
graded suite. Trivial scenarios stay in the repository as mechanism regression fixtures —
they are useful for testing admission and replay — but they never count as evidence that
reasoning helped.

This makes the suite adversarial toward its own authors. That is the point. It is also
the single check most likely to shrink the suite dramatically on first run, and the
correct response to that is relief, not tuning the threshold down.

### 2.2 Density → a large, declared negative class

**The flaw**: benchmarks where most of the data is anomalous reward trigger-happy
detectors and hide false-positive cost.

**The countermeasure**: **35–45% no-fault scenarios**, declared per suite in
`suite.json` and asserted by the generator. For reference, a fault taxonomy with one
negative label out of seven — a common shape — is about 14% negatives, which is far too
few. Negatives are not empty traces; they are the
hard cases — a cloud passing over a PV plant, a legitimate 4-minute door opening, a
nightly backup, morning warm-up, a planned changeover, an animal in heat. Every domain in
the catalog declares at least one, and most declare a `transient_none` family with
several distinct shapes.

Precision is reported beside recall, always, with the negative-class fraction printed on
the same line. A recall number without its negative fraction is not a result.

### 2.3 Mislabeled truth → three timestamps, not one

**The flaw**: usually mislabeled annotations. Ours is different and sharper: the label is
correct, but *when* it became true is not the same as when it became knowable.

A bearing ramp at 0.01 mm/s per hour is physically present at injection and completely
invisible in the noise for eleven hours. Scoring detection latency against injection time
punishes a perfect detector for not being clairvoyant, and — worse — *rewards a leak*,
because the only way to score well is to know something the evidence does not contain.

**The countermeasure**: three computed timestamps.

| Timestamp | Definition | Computed how |
|---|---|---|
| `injection_time` | the world changed | the command log |
| `first_observable_time` | the earliest instant at which the fault's contribution exceeds the **detector's** noise floor at the declared SNR (default 3.0) | the fault's declared detector form applied to the affected states, against the detector's effective σ; solved analytically for `step`/`ramp`, numerically otherwise |
| `unavoidable_time` | the instant by which the deviation is so large that any competent detector must catch it (default SNR 10.0) | same computation, higher threshold |

The detector is an **expression, not a channel** (G-03). A fouled probe reading a plausible
value has no single-channel signature; its fault becomes observable only as a `peer_residual`
against sibling entities, and a leak only as a `conservation_residual` across a balance. The
"effective σ" is the noise floor *after that form's propagation rule* — a `channel_divergence`
combines two channels' σ in quadrature, a `peer_residual` pools the sibling variance. Scoring
against a single channel's raw σ would put `first_observable_time` at a fictional instant and,
worse, in the too-early direction, which turns correct detections into false `suspicious`
flags on exactly the domains where reasoning matters most.

Scoring consequences:

- Detection **before** `first_observable_time` is `suspicious` — flagged, not credited.
  Repeated suspicious detections trigger a leak investigation (§5), because the only
  ways to beat the noise floor are luck and cheating.
- Detection **between** `first_observable_time` and `unavoidable_time` is the graded
  interval. This is where the arms actually differ and where all the signal in the
  experiment lives.
- Detection **after** `unavoidable_time` is a miss, however confident the eventual
  diagnosis.
- Detection latency is reported as `t_detect − first_observable_time`, never as
  `t_detect − injection_time`.

This is, in my judgement, the most consequential detail in the scoring design. Without
it, every latency number in every report is wrong by an amount that varies per scenario,
and the variation correlates with fault type — which produces confident, systematically
biased comparisons between arms.

### 2.4 Run-to-failure bias → randomized onset, including before t₀

**The flaw**: benchmarks where the fault is always at the end, so "predict the last
segment" scores well.

**The countermeasure**: onset position is sampled uniformly across the trace, subject to
leaving room for `unavoidable_time` to fall inside it. Additionally, **10–15% of
scenarios begin with the asset already degraded** — the fault predates t₀ and there is no
healthy baseline to compare against.

That second case is worth more than it costs. It is realistic (you deploy monitoring onto
an existing plant, not a new one), and it exercises a consumer's cold-start handling: the
case where there is no healthy baseline to compare against, which many designs leave
under-specified and few benchmarks test.

### 2.5 The fifth flaw, ours: the oracle leak

Covered in [MCP_SURFACE.md §4](MCP_SURFACE.md). It belongs in this list because it is the
one that would fool us hardest: a leaked answer produces *excellent* scores, and nothing
in an ordinary results table looks wrong.

The `suspicious` classification in §2.3 is the tripwire. A system that scores well and
detects before the noise floor is not good; it is cheating, and the leak is a bug in our
code, not misconduct by the model.

## 3. Three metric classes

Never mixed. Only the first can gate CI.

### 3.1 Mechanism metrics — deterministic, no model required

Computed from the simulator's own records and the submitted verdict alone. Pass/fail with exact expected values.

Split in two, because the simulator owns one half outright and can only observe the
other through a submitted verdict.

**Instrument metrics** — the simulator grading itself. No consumer involved.

| Metric | Gate |
|---|---|
| output reproducibility (3 runs) | byte-identical |
| `verify` from the run artifact | reproduces |
| sink equivalence (`inproc` / `file` / `http-push+stepped`) | byte-identical |
| adapter conformance | schema valid and golden byte-match, every adapter |
| perturbation fidelity | every injected perturbation appears in the delivery ledger and nowhere in the delivered event |
| effector idempotency | repeated `command_id` applies exactly one effect |
| isolation | prefix indistinguishability passes for every domain |

**Consumer metrics** — computed from a submitted
[verdict](../contracts/consumer-verdict-v0.1.schema.json) against sealed truth and the
delivery ledger. Stated in neutral terms, so any stream processor can be measured.

| Metric | Gate |
|---|---|
| duplicate handling | every injected duplicate reported `duplicate`, not `accepted` |
| identity conflict | every `id_reuse` reported `conflict` |
| lateness classification | matches the injected delay distribution |
| dropped-event detection | every injected `drop` yields a detection of absence, or an explicit admission gap |
| clock-skew rejection | every skewed record reported `rejected` or `malformed` |
| evidence grounding | every `evidence_refs` seq was actually delivered to this consumer |
| action fidelity | every claimed action matches an effector-log entry, and vice versa |
| injection-probe neutrality | every verdict field byte-identical to the benign-control run |
| interlock handling | refusal produces no retry and no alternate route to the same effect |

The probe row is a differential test with an exact expected result, which is why an
invariant that usually gets reviewed by a human reading outputs can be gated by CI.

The grounding and action-fidelity rows are only possible because the simulator holds the
delivery ledger and the effector log. A consumer citing evidence it never received, or
claiming an action it never took, is caught by arithmetic rather than by trust.

### 3.2 Judgment metrics — need a model, reported with intervals

| Metric | Definition |
|---|---|
| label accuracy | over the graded, non-trivial suite |
| detection latency | `t_detect − first_observable_time` |
| suspicious-detection rate | before `first_observable_time`; should be ~0 |
| false-positive rate | over the 35–45% negative class |
| sensor-vs-process discrimination | correct attribution on `sensor_pathology` profiles |
| evidence grounding | every claim in the diagnosis traceable to admitted evidence |
| calibration | stated confidence vs. observed accuracy |

Never gated in CI. Reported with confidence intervals clustered by scenario, over the
three repetitions the existing evaluation design already mandates.

### 3.3 Loop metrics — need actuation; new capability

Only measurable because the simulator closes the loop.

| Metric | Definition |
|---|---|
| time to resolution | from `first_observable_time` to the world state recovering |
| deadline adherence | fraction of `aquaculture-pond`-class scenarios resolved before stock loss |
| unnecessary action rate | effector calls on negative-class scenarios |
| action appropriateness | the effector matched `expected_effector` |
| **false-success rate** | success reported under `silent_no_effect` (the shadow-state variant with no confirming tell) |
| compensating-action correctness | a follow-up action correctly compensated an effect a later correction invalidated |
| approval latency | round trip on a consumer's human-approval path, on `auth-and-edr` and `water-distribution` |
| cost per resolved scenario | reasoning spend ÷ correct resolutions |

**False-success rate must be zero.** A non-zero value means the integrated system records
a learned record saying an action worked when the world says otherwise — a learning loop
teaching itself something false. This is the single most important number the simulator
produces, and no other instrument can produce it.

## 4. Cognition rate, in operator units

A rate expressed as a fraction of events — "reasoning invoked on at most 0.5% of
events," a natural gate for a consumer to set itself — scales with event rate, so it means
something entirely different in `k8s-cluster` (torrent) than in `grain-storage` (one event
per sensor per hour). The two numbers are not comparable, which makes them useless for
comparing consumers or domains.

The simulator therefore reports the operator-facing form as well, borrowed from EEMUA 191
and ISA-18.2:

| Metric | Target | Source |
|---|---|---|
| episodes per entity per operator-hour | < 6 | EEMUA 191 steady-state guidance |
| flood check | never > 10 episodes in any 10-minute window across the monitored fleet | ISA-18.2 alarm-flood definition |
| suppression effectiveness | 45–90% reduction from hysteresis + `sustained_for` vs. the same model with both disabled | measured reductions in the alarm-management literature |

The third is a *self-check on the simulator*, not on the runtime. If turning off
hysteresis and sustained duration does not roughly triple the episode count, either the
noise model is unrealistically clean or the thresholds are nowhere near the operating
point — and in both cases the whole suite is measuring an easier problem than it claims.
This is a cheap and unusually direct test of whether the simulated world is hard enough.

The flood check is exercised deliberately by the `correlated_cascade` profile every
domain must offer, which is also the scenario that should trip a consumer's aggregate
cost ceiling and kill switch — the kind of mechanism that rarely has a test.

## 5. Leak detection

Three independent controls; all three run in CI.

1. **Prefix indistinguishability** ([MCP_SURFACE §4.2](MCP_SURFACE.md)) — the differential
   test over operator-role responses on an identical evidence prefix.
2. **Suspicious-detection monitoring** (§2.3) — detection before the noise floor is
   statistically near-impossible. A rate above ~1% is treated as a leak until explained.
3. **Scrambled-label control** — for a sample of runs, score against a randomly permuted
   label set. Accuracy must collapse to chance. If it does not, something in the scoring
   path is reading truth that the arm should not have had, and the defect is in the
   harness rather than the model.

Control 3 catches leaks in the scorer itself, which the other two miss.

## 6. Suite composition

Per domain, per the existing evaluation design's floor of 70 scenarios and ≥ 10 per
label, tightened:

| Property | Value |
|---|---|
| minimum scenarios | 100 |
| minimum per fault label | 10 |
| negative-class fraction | 35–45% |
| trivial scenarios in the graded set | 0 |
| pre-degraded starts | 10–15% |
| perturbation coverage | every transport perturbation in ≥ 5 scenarios |
| `correlated_cascade` scenarios | ≥ 3 |
| `sensor_pathology` scenarios | ≥ 10 |
| regenerable byte-for-byte from the seed | required |

Committed as JSONL plus `labels.jsonl` plus the seed, exactly as the existing design
requires — and additionally with the audit verdicts, so a reviewer can see which
scenarios were excluded as trivial and why.

## 7. Reporting rules

1. **A simulator score is never reported as a field result.** The existing evaluation
   design already says this; it applies with more force here, because the closed loop
   makes the output look operational. "The agent resolved 94% of oxygen crashes" is a
   statement about a differential equation, not about a fish farm.
2. Every table carries: the negative-class fraction, the trivial-exclusion count, the
   suspicious-detection rate, and whether the run was hash-reproducible.
3. Unblinded runs are excluded and the exclusion count is printed.
4. Detection latency is always relative to `first_observable_time`, and the report says so
   on the same page.
5. When a scenario's outcome depends on a perturbation, say which — a miss caused by a
   `drop` is a transport result, not a reasoning result, and the perturbation layer's
   separation from the world core (TECHNICAL_DESIGN §4) is what makes the distinction
   available.

## 8. What a negative result looks like

Worth stating in advance, before anyone is invested.

If the graded suite shows that a structured Situation does **not** beat a raw evidence
window as model context, the correct response is to report it and to reconsider the
product, not to add scenarios until the number moves. The existing V0.1 design already
commits to this — "a negative answer delivered on time is a successful V0.1."

The simulator makes that commitment testable in a way the current design cannot, because
the trivial-baseline audit removes the easiest way to manufacture a positive result:
filling the suite with scenarios that any detector solves and calling the aggregate a
win.
