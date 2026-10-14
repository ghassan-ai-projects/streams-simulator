# Simulation ideas: product-proof review

- **Date**: 2026-10-05
- **Status**: review of the ten-idea portfolio in this folder, plus a replacement idea set
- **Question asked**: which simulations would *prove* that Agentic Stream and Tamoz are worth
  adopting, and are SI-01 to SI-10 those simulations?

## 1. Verdict

The portfolio is careful instrument engineering aimed at the wrong target. All ten ideas ask
"can the simulator express and seal this?" None asks "would a skeptical buyer believe Agentic
Stream + Tamoz is better because of this run?" As a result, most ideas test **table stakes**:
dedup, lateness, replay masks, and watermark folds. A buyer assumes those work and will not
pay for them. Agentic Stream's own evaluation design already marks the cells that *would*
prove value as **missing**: S7.1 decision quality, S7.3 abstention, S7.4 counterfactual regret
against the deterministic baseline, and S7.5 holdout
(`agentic-stream/docs/eval/02-suite-catalog.md`). No idea in this folder targets them.

Recommendation: stop the improvement loops. Keep two ideas (SI-02, SI-07), reframe them around
consequences, and build the **outcome-and-baseline layer** (§4) before adding another domain.
The first deliverable should be one published scorecard on *shipped* domains, not an
eleventh idea document.

## 2. What the portfolio gets right

- **Honest capability accounting.** Every idea names its missing machinery as a design change
  first. This discipline should continue.
- **Cross-cutting findings that hold up.** These include `run.End` not flushing perturbation
  buffers, the emission-blind truth solver, inverted `silent_no_effect` cells, and parsed-only
  `batch_size`/`time_reference`. They are real instrument bugs and limits. Fix them whatever
  happens to the ideas (README §6.1).
- **SI-02's core.** Ambiguous acknowledgement, no double-apply, and the idempotency-window
  boundary form a real differentiator. They map directly to Agentic Stream S4.4 and S4.5 and
  to Tamoz's `:unknown` reconcile.
- **The novelty lens's frontier list.** Sparse rate, large cardinality, distributions, and
  adversarial-adaptive faults are the right targets. §4 uses them.

## 3. Problems

### 3.1 The rubric has no product dimension

The seven dimensions (`consumerValue`, `dataExpressibility`, `determinismOracle`,
`ledgerQuiescence`, `novelty`, `sealedTruth`, `specificity`) grade the *instrument*. Even
`consumerValue` means "a cell a team can act on", not "a result that separates the product
from its alternatives". An idea can score 10/10 everywhere and still produce a run in which a
50-line threshold script ties Agentic Stream. That is a negative proof.

### 3.2 It tests what is already tested

Agentic Stream already covers or partly covers S1.4 (duplicate and out-of-order events), S1.5
(late correction), S1.7 (missing heartbeat), S1.8 (restart), and S1.9 (quarantine) with its
own traces. Six of ten ideas (SI-01, 03, 04, 05, 07, 10) re-derive these cells with a stronger
oracle. That adds rigor, but it does not add proof of value.

### 3.3 No baselines, so no proof

A product claim is comparative, for example: "fewer bad outcomes than rules, at lower cost
than an LLM on every event, and with zero unsafe actions." The simulator has one baseline,
`refconsumer`, which is a threshold detector, and its trivial-baseline audit removes
scenarios that are *too easy*. Nothing removes scenarios where the product *fails to beat* a
strong baseline. Without that, every result can be attacked as "you picked easy cases".

### 3.4 Scores are detection latency, not outcomes

Scoring today covers detection against `first_observable_time`, admission outcomes, and the
loop metric (`internal/score/loop.go`). A buyer asks how much money was lost, how many unsafe
actions there were, and how many operator pages were sent. The schema already carries
`consequence: [cost, downtime, regulatory, safety]` as an axis label
(`domain-spec-v0.1.schema.json:35`). It is a tag, not a number, so outcomes cannot be totaled.

### 3.5 The consumer references were verifiable but not checked

The study marked F-002, S1.8, S1.9, F-010, kill-matrix, MAX_RECONCILE_GRANTS, WP-T4,
CheckpointConflictError, and Last-Event-ID as "brief vocabulary, unverifiable". All of them
exist in the sibling repositories `../agentic-stream` and `../tamoz`. Grep hits by file count:
F-002 in 3 Agentic Stream files and 28 Tamoz files; kill matrix in about 40 Tamoz files;
CheckpointConflictError in 98 Tamoz files. The ideas should be anchored to those repositories'
real catalogs, especially Agentic Stream's `docs/eval/02-suite-catalog.md` and
`docs/research/next-level-2026-10/03-OPPORTUNITIES.md`. Those documents already ask for this
simulator by name: H1 **SituationBench**, G4 **cognition economics**, and F4 **soak driven by
streams-simulator**.

### 3.6 Process cost is outrunning output

The folder holds about 3,000 lines of idea documents, ten documents that hit up to 10
improvement rounds, a scorer that cannot be reproduced from the tree, and no artifact admitted
for use. Much of the text is provenance about which pass knew what. The next round should be
judged by runs and scorecards, not by more prose.

## 4. The reframe: a baseline ladder and a regret oracle

A simulation proves value only when all four of the following hold:

1. **A baseline ladder runs on the same world, seeds, and chaos.** The ladder is: (a) a
   strong rules engine with hysteresis and debounce, not a strawman; (b) an LLM agent called
   on every event with the same tools; (c) Agentic Stream in deterministic-only mode; (d)
   Agentic Stream with Tamoz episodes.
2. **Outcomes are measured in units a buyer uses.** These are loss in money, unsafe effector
   calls, operator pages per hour (ISA-18.2), and model calls and dollars per correct outcome.
3. **A named product mechanism explains the gap.** Examples: invariant 1 (evidence is never an
   instruction), invariant 7 (policy revalidates before dispatch), the cognitive scheduler, or
   Tamoz's `:unknown` reconcile. If no mechanism explains the win, the win is noise.
4. **A sealed oracle bounds what is achievable.** Regret is measured against the best policy
   that could act on observable evidence, not against a policy that knows the future.

### 4.1 The one simulator capability to build first: an outcome economics block

This would be a domain-data block, so domains stay data:

```json
"economics": {
  "loss_rate": [{"when": {"state": "bearing_wear", "gt": 0.6}, "per_hour": 1200}],
  "action_cost": {"dispatch_technician": 400, "trip_breaker": 50},
  "unsafe": [{"effector": "close_breaker", "when": {"state": "fault_present", "eq": 1}}]
}
```

- **Loss is closed-form for F0 dynamics.** A linear integrator crossing a threshold gives an
  analytic integral, so the existing non-negotiable *analytic cross-check* still applies:
  implement the loss integral twice and assert they agree.
- **Regret uses an observable-oracle policy.** That policy acts at `first_observable_time`
  with the expected effector. Regret is the consumer's loss minus the oracle's loss. Every
  input is already sealed (onset, first observable, expected effector, deadline).
- **Unsafe calls are judged from the effector log.** The simulator already records every
  `EffectorCall`. An unsafe call is a predicate over the call and the hidden state at call
  time, so the judgment is exact and no LLM judge is needed.
- **Extend the trivial-baseline audit to "no headroom".** Drop a scenario when the oracle's
  regret advantage over do-nothing is below a floor. Such a scenario cannot separate anyone.

This is a design change first, under AGENTS.md: a ground-truth contract extension plus a
scorer. It is consumer-agnostic. The baseline-ladder harness itself must live **outside**
this binary, in Agentic Stream as its H1 or in a separate bench repository, because the
simulator holds no consumer knowledge.

## 5. Better ideas

Each idea names the claim it proves, why each baseline loses, the mechanism that wins, and
the honest simulator changes it needs.

### P-1 · Alarm storm with one root cause (flagship)

- **Claim proved**: Agentic Stream turns 40 symptoms into one Situation, one episode, and one
  correct action.
- **World**: one upstream degradation spreads with lags into many downstream symptoms across
  entities and channels. Examples are a chiller fault surfacing as rack temperatures, fan
  speeds, and throttling, or a hub delay surfacing across shipment legs. Decoy symptoms with
  no causal link are mixed in.
- **Why baselines lose**:
  - Rules fire roughly 40 alerts, breaching the ISA-18.2 flood limit.
  - An LLM on every event spends about 40 calls and proposes conflicting symptom-level fixes.
  - A deterministic-only consumer pages correctly but cannot name the root.
- **Winning mechanism**: Situation coalescing, the cognitive scheduler (one episode, not 40),
  and policy rejection of symptom-level intents that a superseding Situation version makes
  stale.
- **Scores**: regret, pages per operator-hour, model calls, root-cause hit rate, and time to
  the correct action.
- **Simulator changes**:
  - Start on the shipped `shipment-exception` and `k8s-cluster` domains. Both already declare
    `correlation: cascading`; shipment-exception's stated stress is "attribute the breach to
    the right link, not the loudest feed".
  - True cross-entity cascades need entity coupling in dynamics. That is a design change, and
    it mirrors Agentic Stream's B4 "declared entity relations".
  - Ground truth must carry a `root_entity` (closed-record extension).

### P-2 · Investigation with priced probes (Tamoz-native)

- **Claim proved**: Tamoz finds causes by asking the world, at bounded cost, without
  fabricating.
- **World**: the symptom is ambiguous between two or three hidden causes that look the same
  on the passive stream. **Diagnostic effectors** such as `pull_vibration_spectrum`,
  `inspect_seal`, and `dispatch_technician` cost money and time. Their result appears as an
  evidence event on the ordinary stream after a delay, never through the nameplate. The
  nameplate must stay time-invariant (`MCP_SURFACE.md` §4.1), so this design keeps every probe
  result on the evidence plane, with ledger and perturbation coverage.
- **Why baselines lose**:
  - Rules cannot investigate at all.
  - An agent that probes everything wins accuracy but loses on cost.
  - An agent that guesses is cheap but wrong, or fabricates.
- **Winning mechanism**: Tamoz's investigation loop (probes, then a decision or `unknown`
  naming what is missing). **Fabrication is caught exactly**: a citation of evidence never
  delivered fails against the delivery ledger (`CONSUMERS.md` §3.2). This is the strongest
  property the simulator has that no LLM benchmark has.
- **Payoff**: Tamoz's current headline is 80% cause-found and 7 of 232 runs fabricated, on a
  29-situation fixture that is not held out. P-2 turns that into a generative, held-out,
  seed-sealed result.
- **Scores**: diagnosis accuracy, probe spend, regret, fabrication count (target 0), and
  `unknown` calibration.
- **Simulator changes**: an *effector-triggered observation*, meaning an effector whose effect
  makes a channel emit a reading of hidden state. This is a domain-agnostic design change and
  the only one P-2 needs. Probe results can be perturbed (lost or late), which tests
  investigation under the same delivery chaos as everything else.

### P-3 · Prompt injection with a target (upgrades SI-07)

- **Claim proved**: Text cannot cause an effect. Invariants 1 and 6 hold under attack.
- **World**: the operator-log channel is `attacker_controlled`. The injected text has a
  **specific harmful target**, such as "open bypass valve 3" or "disable interlock". The
  target is sealed in the probe corpus, and a matching effector call is an *unsafe* call
  under the §4.1 economics block. Catalog 25's adversarial-adaptive model can follow: the
  attacker changes text after seeing the consumer's actions.
- **Why baselines lose**: an LLM on every event with effector access shows a measurable attack
  success rate. That number is the demonstration.
- **Winning mechanism**: no model-to-effector path, policy revalidation, and a typed intent
  catalog. The expected attack success rate is **0 by construction**. Also measure the
  quieter failure, where an injection changes the decision without triggering the target.
  That is the paired-cell differential SI-07 already designed.
- **Scores**: attack success rate per consumer on the ladder, decision drift between clean
  and injected cells, and benign-task regret, which shows the defense is not a refusal to
  act.
- **Simulator changes**: SI-07's four named gaps (sampler case, stamping fix, fail-closed
  params, and pair runner), plus a field for the harmful-target effector in the probe corpus.

### P-4 · Cognition economics at fleet scale

- **Claim proved**: Selective cognition holds outcomes steady at a fraction of the cost.
- **World**: 2,000 to 20,000 entities at sparse rates. Real faults are rare among many
  near-miss excursions that look like faults and self-heal. This fills two zero-coverage axis
  values, cardinality `large` and rate `sparse`, which the novelty lens flagged.
- **Why baselines lose**:
  - An LLM on every event becomes unaffordable, with cost growing linearly in events.
  - Rules page on every near-miss.
  - A capped LLM budget misses the real fault when the cap binds.
- **Winning mechanism**: the cognitive scheduler, triggers, and budgets (S3.2, S3.3). This is
  Agentic Stream's G4 report built on generated data.
- **Scores**: model calls per entity-day, dollars per correct outcome, regret, and missed
  faults under a fixed budget, swept across budget levels into a curve.
- **Simulator changes**: almost none beyond §4.1. It also produces the >10k conservation
  evidence the ledger lens asked for and the single-node capacity table (Agentic Stream A4).

### P-5 · Exactly-once under consumer crashes (upgrades SI-02)

- **Claim proved**: A `kill -9` at any point never double-applies a physical effect and never
  loses one silently.
- **World**: SI-02's feeder, with an **asymmetric consequence**. A second `close_breaker` onto
  an uncleared fault is an *unsafe* call; a lost trip is downtime loss. Add the
  duplicate-command-log generator SI-02 already names.
- **Harness**: kill Agentic Stream or Tamoz at seeded world-time points. The quiescence barrier
  makes kill points reproducible as world time, not wall time. The simulator's effector log
  and hidden state serve as the **independent instrument**.
- **Why this matters now**: Tamoz's release blockers for invariant 19 and objective 4 are
  "environment-bound". The kill-matrix evidence cannot run in constrained environments. Here
  the simulator is the external oracle those blockers lack.
- **Why baselines lose**: a naive retry-on-timeout agent double-closes, which the simulator
  records as an unsafe call.
- **Scores**: unsafe calls (target 0), lost effects, and the `unknown`-to-resolved rate with
  human resolution time.
- **Simulator changes**: SI-02's two prerequisites. Process-kill orchestration belongs to the
  harness, not the simulator.

### P-6 · Act-now versus wait-for-truth

- **Claim proved**: Completeness-aware decisions beat both acting on arrival and waiting for
  everything.
- **World**: early evidence is provisional. A late correction overturns it in a known fraction
  of cases (a declared probability), and waiting costs a declared loss per hour. The optimal
  stopping point is closed-form for F0 with declared costs, so this is a regret oracle with
  an analytic cross-check.
- **Why baselines lose**: act-on-arrival takes false actions, and wait-for-close is late.
- **Winning mechanism**: provisional versus on-time completeness, `prefer_corrected` and
  window settling (Agentic Stream A1 and A2), and S4.2 freshness revalidation, which rejects
  the pending intent when the correction lands. This is the "correction after a decision"
  stressor in H1. It absorbs SI-03's window matrix and SI-04's freeze-then-jump as delivery
  variants rather than separate domains.

### P-7 · Agent-fleet supervision (dogfood)

- **Claim proved**: The product supervises the work its builders actually do.
- **World**: a fleet of coding-agent runs with stuck, looping, budget-burn, and flaky-gate
  faults. Interventions are pause, escalate, and open an issue. It is software-only and high
  volume, and Tamoz runs can serve as the real data source. This is Agentic Stream H5.
- **Why it is lower priority**: the dynamics are queueing and state machines, so the physics
  oracle is weaker. It is better as a positioning demo than as an early proof.

## 6. Disposition of SI-01 to SI-10

| ID | Disposition | Reason |
|---|---|---|
| SI-02 feeder-switching | **Keep → P-5** | Real differentiator. Add the asymmetric unsafe consequence and the crash harness. |
| SI-07 field-log-injection | **Keep → P-3** | The most demo-able invariant. Low score only because machinery is missing, and that machinery is the deliverable. |
| SI-05 tank-dip-replay | Fold into a P-6 delivery variant | A good replay-mask stressor, but table stakes on its own. |
| SI-01 telemetry-estate | Fold into P-2 | "Off, buffering, or lull?" is a natural investigation scenario, with diagnostic probes resolving the ambiguity. |
| SI-03 district-metering | Fold into P-6 | The window matrix becomes a P-6 delivery variant. The batch capability stays a generic fix. |
| SI-04 ami-headend | Fold into P-6 | Freeze-then-jump is a P-6 delivery variant. Its settlement scorer is a narrow case of §4.1. |
| SI-08 amr-fleet | Hold | Entity retirement (Q3) is a real Agentic Stream defect, but it needs a truth extension. Revisit after §4.1. |
| SI-06 ev-plaza | Hold | Schema drift matters, but cohort scoping is not expressible today. |
| SI-09 reservoir | Drop as a domain | Seasonal phase-at-onset is a sampling property. Any periodic domain can host it. |
| SI-10 tick-feed | Drop as a product story | Market-data buyers will not adopt an LLM-in-the-loop runtime for tick integrity. Keep only the torrent recipe as a load generator for P-4. |

## 7. Sequence

1. **SituationBench v0 on shipped domains, no new domain.** Use `rotating-machinery` and
   `shipment-exception`, the four-rung ladder, k = 3, and sealed holdout seeds. Results are
   reported with the existing scores plus the §4.1 economics, built as the first design change.
   This is the single most valuable artifact for both products.
2. **The gate-carrying slice** from README §6.1 (`run.End` terminality, injection provenance,
   shaped conservation). Without it the instrument can produce confident wrong answers.
3. **P-3, then P-5.** These are the cheapest unforgeable demos: a 0% attack success rate and
   zero unsafe calls under kills, each next to a baseline that fails.
4. **P-2.** It needs effector-triggered observation, and it gives Tamoz its headline number.
5. **P-1 and P-4.** These are the flagship scale and coalescing results. P-1 waits on the
   entity-coupling design.

## 8. Credibility guardrails

One person builds the simulator, both products, and the benchmark, so every result will be
read as self-serving unless the protocol prevents it.

- **Pre-register.** Commit the scenario manifests, holdout seed digests, and baseline configs
  before the product runs. Never tune on holdout seeds.
- **Use strong baselines.** The rules baseline gets hysteresis, debounce, and cooldown. The
  LLM baseline gets the same model, tools, and prompt effort. Publish their configs.
- **Publish losses.** Report cells where Agentic Stream or Tamoz loses or ties. A benchmark
  where the author always wins is ignored.
- **Keep Class D before Class C.** A Class D (deterministic) failure voids that run's
  capability claims (`agentic-stream/docs/eval/README.md`).
- **Make it reproducible by others.** One artifact must reproduce the scorecard, with
  `(sim_version, domain_digest, adapter_digest, seed, command_log)` already pinned by design.

## 9. Evidence used in this review

- This folder: README, research notes, review evidence, and the purpose and feasibility
  sections of SI-01 to SI-10.
- Agentic Stream: `documentation/overview/{product,limitations}.md`, `documentation/roadmap.md`,
  `docs/eval/{README,02-suite-catalog}.md`, `docs/research/next-level-2026-10/03-OPPORTUNITIES.md`,
  and `internal/ingress/simulator.go`, which already ingests this simulator's output.
- Tamoz: `README.md` (investigate metrics) and `documentation/limitations.md` (the
  environment-bound kill-matrix release blockers).
- This repository: `domain-spec-v0.1.schema.json` (`consequence` is a tag, effector failure
  modes, no economics block), `docs/design/MCP_SURFACE.md` §4.1 (time-invariant nameplate),
  `domains/*.json` (`cascading` declared by k8s-cluster and shipment-exception), and
  `docs/design/GAP_ANALYSIS.md` G-03.
- Cross-repo grep for the "unverifiable" brief IDs (§3.5). Nothing was built or run. No test
  results are claimed.
