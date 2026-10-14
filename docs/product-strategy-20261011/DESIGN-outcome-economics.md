# Design proposal: outcome economics

- **Status**: proposal; a design change per AGENTS.md ("deviations need a design change first")
- **Date**: 2026-10-11
- **Contracts touched**: `domain-spec`, `ground-truth`, scorecard bundle. Not touched: the
  event envelope, the adapter contract, the world digest of existing domains.
- **Invariants preserved**: determinism, sealed oracle, analytic cross-check, domains as
  data, consumer-agnostic binary.

## 1. Problem

Scoring answers *was the fault detected, and in time?* A buyer asks *what did it cost, and
how much of that was avoidable?* Today consequence is an axis tag and
`counterfactual.if_no_action` is a label of at most 200 characters (for example
`stock_loss`). Neither can be summed, compared across consumers, or bounded.

## 2. Principle

Every number is **declared by the domain author and labelled as such**. The simulator does
not claim a pond's oxygen crash costs a specific amount of money; it claims *given these
declared stakes, this consumer lost X and the best observable policy lost Y*. The value is
the comparison and the bound, not the absolute figure.

## 3. Domain-data block

An optional `economics` object per domain, with units named by the author:

```json
"economics": {
  "currency": "unit",
  "loss_rate": [
    {"id": "stock_at_risk", "entity_scope": "each",
     "when": {"state": "dissolved_oxygen_true", "lt": 3.0}, "per_hour": 1200}
  ],
  "action_cost": {"start_aerator": 15, "dispatch_technician": 400},
  "unsafe": [
    {"effector": "close_breaker", "when": {"state": "fault_present", "gt": 0.5},
     "severity": "safety"}
  ]
}
```

- `loss_rate` terms accrue while a predicate over **true** state holds. Predicates are
  closed-form comparisons (`lt`, `gt`, `between`), never an expression language.
- `action_cost` is charged per accepted effector call, whether or not it took effect.
- `unsafe` predicates are evaluated over the effector call and hidden state at call time.
- Absence of the block means "no economics": every economic metric is **absent**, not zero.

## 4. Counterfactual worlds

Determinism makes counterfactuals cheap: a run is a pure function of its command log, so
three worlds differ only in commands.

| World | Command log | Purpose |
| --- | --- | --- |
| Actual | the consumer's recorded effector calls | what happened |
| Do-nothing | none | the cost of inaction |
| Oracle | the label's `expected_effector` at `first_observable_time` on the faulted entity | the best an observer of the evidence could do |

Regret is `loss(actual) + action_cost(actual) − loss(oracle) − action_cost(oracle)`.
Headroom is `loss(do-nothing) − loss(oracle)`. The oracle uses only information the
instrument already seals, so it is bounded by what was observable and never by hindsight.

## 5. Ground-truth extension

Added to the sealed record, additive and `omitempty`:

```json
"economics": {
  "do_nothing_loss": 5400.0,
  "oracle_loss": 310.0,
  "oracle_action_cost": 15.0,
  "headroom": 5090.0,
  "currency": "unit"
}
```

Computed by the truth solver at seal time, so a consumer with no reveal cannot read it.
Old artifacts and labels without the block replay and score exactly as before.

## 6. Scorer additions (bundle `scorecard-bundle-v0.3`)

| Metric | Definition |
| --- | --- |
| `loss_incurred` | integral of declared loss terms over the actual world |
| `action_cost` | sum over accepted effector calls |
| `unsafe_calls` | count of calls matching an `unsafe` predicate at call time |
| `regret` | as §4 |
| `regret_ratio` | `regret / headroom`, defined only when headroom exceeds the floor |
| `false_action_cost` | action cost on negative-class scenarios |

Online and offline scoring must produce identical values for every shared metric; the
existing parity test (`internal/score/internal/domain/parity_test.go`) extends to cover them.

## 7. Analytic cross-check

Loss over a state trajectory is an integral of an indicator. For F0 and F1
first-order-lag dynamics under a step input, the threshold-crossing time is closed form.

1. **Numeric path**: step the world at the domain's `dt_ms` and accumulate.
2. **Analytic path**: compute crossing times from the lag form and integrate piecewise.
3. **Assertion**: agree within a stated tolerance for every fault of every shipped domain
   that declares economics.

This is the same non-negotiable as the existing solver cross-check, applied to a new number.

## 8. Audit extension: no headroom

The trivial-baseline audit removes scenarios that are too easy. Add the converse, a scenario
that cannot separate anyone:

- **No-headroom**: `headroom` below a declared floor ⇒ excluded from graded evidence.
- **Saturated**: the strongest reference policy ([baseline ladder](DESIGN-baseline-ladder.md))
  reaches the oracle's loss ⇒ kept as a regression fixture, never counted as evidence.

## 9. Sensitivity reporting

Economics are assumptions, so every suite report includes:

- results under each declared parameter scaled by 0.7 and 1.3;
- whether the ranking of consumers changes;
- a marker on any comparison that flips.

A claim that survives ±30 % is a claim about behaviour; one that does not is a claim about
the author's numbers.

## 10. Failure modes and their controls

| Failure | Control |
| --- | --- |
| Author tunes numbers to favour a consumer | Economics live in the domain digest and are part of the published artifact; sensitivity report |
| Oracle policy is weaker than a clever consumer ⇒ negative regret | Allowed and reported; flags a mis-specified expected effector |
| Loss predicate reads unobservable state, leaking via scoring | Scoring is post-run and sealed; consumer never receives it before verdict submission |
| Per-entity scope multiplies cost by fleet size | `entity_scope` explicit; report per-entity and total |
| Digest drift | Existing domains unchanged; new block changes only domains that adopt it |

## 11. Test plan

1. Schema: valid, invalid, and absent economics; `additionalProperties: false`.
2. Solver: do-nothing and oracle losses against a hand-computed trajectory.
3. Cross-check: numeric vs closed form across all faults of the two pilot domains.
4. Replay: artifact recorded before and after the change replays; economics absent ⇒ metrics
   absent.
5. Parity: online equals offline for every new metric.
6. Audit: a synthetic zero-headroom scenario is excluded and reported with its reason.
7. Property: regret is non-negative when the consumer is the oracle policy.

## 12. Rollout

1. Contract change and schema, behind an absent-means-absent default.
2. Pilot on `aquaculture-pond` (deadline and counterfactual already declared) and
   `rotating-machinery`.
3. Scorecard bundle bump and the suite report fields.
4. Extend remaining domains only as a proof scenario needs them.

## 13. Open questions

1. Should `loss_rate` support state-weighted terms (loss proportional to deviation) in
   addition to thresholds? Proportional loss is more realistic and still closed form for F1.
2. Is the oracle policy always a single call at `first_observable_time`, or may a domain
   declare a short sequence for multi-step remediation?
3. Do negative-class scenarios carry an economics block of their own for false-action cost,
   or is the global `action_cost` sufficient?
4. Where is the cost of model calls recorded: here, or only in the bench harness, because the
   simulator cannot observe a consumer's internal calls?
