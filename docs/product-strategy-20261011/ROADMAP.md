# Roadmap to a world-class product

- **Date**: 2026-10-11
- **Status**: proposal, ordered by dependency. Sizes are relative (S ≈ days, M ≈ 1–2 weeks,
  L ≈ 3–6 weeks), not commitments.
- **Rule**: each phase ends in a runnable artifact or a published number, never a document
  alone. AGENTS.md's change policy applies: every item names its contract, invariant, test
  evidence and documentation page before implementation.

## Dependency order

```
0 Reconcile ─▶ 1 Suite runner ─▶ 2 Economics ─▶ 3 Baseline ladder ─▶ 5 Proof scenarios ─▶ 6 Release
                                       │                                     ▲
                                       └──────▶ 4 Capabilities ─────────────┘
```

## Phase 0 — Reconcile and verify (S)

Make the records match the code and close the two verification risks.

| Deliverable | Exit criterion |
| --- | --- |
| Status column on `DEFERRED.md`; refresh `limitations.md`, `roadmap.md`, `docs/README.md` banner | A reader sees no statement a test contradicts |
| Decide `docs/simulation-ideas/`: commit as evidence or archive | Worktree clean |
| Cross-architecture determinism CI job | `scripts/behaviour-pin` output identical on darwin/arm64, linux/amd64, linux/arm64, or divergence recorded as a design decision |
| Fresh `make ci-check`, soak and perf run at one clean commit | Manifest generated for that commit with author and reviewer identity |
| Resolve D-52 with the domain owner | `cold-chain-transit` fault is observable or marked as an intentional control |
| Domain linter in `domain validate` (GAPS §1.4) | Linter fails D-52 on the pre-fix domain |

## Phase 1 — Suite runner and scorecard aggregate (M)

Turn generation into execution. This is already on `documentation/roadmap.md`.

- `streamsim suite run` executes each scenario against a named adapter and a consumer reached
  through the operator endpoint (or the shipped reference consumer), persisting the run
  artifact, ledger, verdict and scorecard per scenario.
- `streamsim suite report` aggregates per-scenario scorecards into one document: per-label
  rates, negative-class false-alarm rate, deadline adherence, mechanism metrics.
- `streamsim score diff <a> <b>` compares two suite reports and lists regressions.
- No consumer knowledge enters the binary: the consumer is an endpoint or a command line.

**Exit**: one command produces a scorecard for the reference consumer on all eight domains,
reproducible from the same seed on a second machine.

## Phase 2 — Outcome economics (M–L)

Full proposal: [DESIGN-outcome-economics.md](DESIGN-outcome-economics.md).

- Domain-data `economics` block: loss rate, action cost, unsafe-call predicates.
- Truth record gains `oracle_loss`, `do_nothing_loss`, `headroom`.
- Scorecard bundle v0.3: loss incurred, action cost, unsafe calls, regret against the
  observable-oracle policy.
- Audit extension: scenarios with no headroom over doing nothing leave the graded suite.
- Analytic cross-check: loss integral implemented twice (numeric and closed form for F0/F1).

**Exit**: shipped domains `aquaculture-pond` and `rotating-machinery` carry economics; their
scorecards report regret; the cross-check test passes; replay of an old artifact still works
(economics absent ⇒ metrics absent, not zero).

## Phase 3 — Baseline ladder (M)

Full proposal: [DESIGN-baseline-ladder.md](DESIGN-baseline-ladder.md).

- In-repo reference policies of increasing strength (hysteresis and debounce rules,
  peer-residual, sustained-for) so "strong baseline" is a versioned, tested artifact.
- Paired-seed comparison, bootstrap confidence intervals, win rate.
- Held-out seed protocol: commit-reveal so a consumer cannot be tuned to the suite.
- Headroom audit now removes scenarios where the strongest reference policy matches the oracle.

**Exit**: a published table where every rung runs on identical worlds, seeds and chaos, with
intervals, and at least one scenario where the strongest rule baseline demonstrably loses.

## Phase 4 — Capabilities the proof scenarios need (L)

Only what a named scenario requires; each is a design change first.

| Capability | Unlocks | Constraint |
| --- | --- | --- |
| Declared entity relations and coupling in dynamics; `root_entity` in ground truth | Alarm storm with one root cause | Truth contract extension; analytic solver must cross-check propagation lags |
| Effector-triggered observation (a probe whose result arrives on the evidence plane after a delay, perturbable) | Investigation with priced probes; fabrication detected against the ledger | Nameplate stays time-invariant |
| Event-time domains (`time_reference` other than `wall`) | Lateness, watermark and skew scenarios that match the product's name | Perturbation and ledger semantics must define event time vs arrival time |
| Targeted adversarial probes tied to an effector | Prompt-injection with a harmful target, scored as an unsafe call | Depends on phase 2 |

**Exit**: each capability has a domain that uses it, an analytic oracle, and a replay test.

## Phase 5 — Proof scenarios and the first public scorecard (M)

Three flagship scenarios (detail in `../simulation-ideas/PRODUCT-PROOF-REVIEW.md` §5):

1. **Alarm storm, one root cause.** Forty symptoms, one cause; measures pages per
   operator-hour, model calls, root-cause hit rate, regret.
2. **Investigation with priced probes.** Ambiguous symptom, costly diagnostic effectors;
   measures diagnosis accuracy, probe spend, fabrication count (target 0), `unknown`
   calibration.
3. **Targeted injection.** Operator-log text with a sealed harmful target; measures unsafe
   calls (target 0).

**Exit**: one published scorecard on shipped plus these domains, with intervals, a named
mechanism for each gap, and the sensitivity analysis from the economics design. Include a
negative result: at least one scenario where the product does not beat the baseline.

## Phase 6 — Reach and release (M)

| Deliverable | Notes |
| --- | --- |
| Tagged `v0.x` releases, signed checksums, SBOM, container image | Reproducible build already a goal |
| Contract freeze policy and `v1.0` of the six schemas | Define what is a breaking change and the migration path for old artifacts |
| Adapters as data: OpenTelemetry JSON, CloudEvents, Prometheus exposition | No code branches; verified by golden files |
| `domain scaffold` and authoring guide with a worked example | Cuts time to first custom domain |
| CI recipe: run the suite, diff against the base branch, fail on regression | A GitHub Action wrapper is optional |
| Library mode documentation for in-process use | `internal/` is private, so decide the public Go surface first |

**Exit**: a team outside the repository runs the suite on their consumer in under a day
using only public documentation.

## What not to build

- More domains or idea documents before phases 2–3 produce a number.
- A broker, a plugin system or an expression language (forbidden by AGENTS.md).
- Per-consumer code in the binary; a consumer is a file or an endpoint.
- A hosted service before the offline product is credible.
- A leaderboard before the held-out protocol exists; a leaderboard on public seeds
  rewards overfitting.

## Risks

| Risk | Mitigation |
| --- | --- |
| Economics are assumptions and can be gamed or dismissed | Label every number as domain-declared; report rank stability under ±30 % parameter change; never present absolute money as measurement |
| Strong baseline is itself a straw man | Version it, publish its tuning budget equal to the consumer's, accept third-party baselines through the same interface |
| Digest moves break stored artifacts | Economics are additive; artifacts without them replay unchanged; bump scoring bundle, not world digest |
| Scope creep from capabilities | Capability only enters with a scenario that needs it and an oracle that covers it |
| Win-rate claims on thin samples | Minimum N from `ground_truth.min_scenarios`, paired seeds, intervals mandatory in reports |
