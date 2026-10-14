# Design proposal: baseline ladder and comparative proof

- **Status**: proposal; depends on [outcome economics](DESIGN-outcome-economics.md)
- **Date**: 2026-10-11
- **Invariants preserved**: the binary has no consumer knowledge; the reference consumer
  remains a test instrument; sealed truth is never exposed before verdict.

## 1. Problem

A product claim is comparative: *fewer bad outcomes than rules, at lower cost than an LLM on
every event, with zero unsafe actions.* The simulator today has one baseline, a threshold
detector, and an audit that removes scenarios that are **too easy**. Nothing removes
scenarios where a good rules engine **ties** the product, so any result can be dismissed as
"you picked easy cases."

## 2. The ladder

Every rung runs on the same world, seeds, perturbations, effector set and wall budget.

| Rung | Policy | What a win over it proves |
| --- | --- | --- |
| 0 | Do nothing | The scenario has stakes (headroom) |
| 1 | Fixed threshold, hindsight-tuned | Not trivially solvable (existing audit) |
| 2 | Strong rules: per-channel hysteresis, `sustained_for`, peer residual, debounce, dedup | Beats what a careful engineer ships in an afternoon |
| 3 | Rules plus the effector playbook (expected effector per label) | Beats an expert-written runbook |
| 4 | Model-on-every-event with the same tools | The product is cheaper and safer, not only more accurate |
| 5 | Product, deterministic-only mode | Reasoning adds value beyond the deterministic core |
| 6 | Product, full | The headline claim |
| ∞ | Observable oracle | The achievable bound; never beaten except by luck |

Rungs 0–3 are deterministic, versioned policies shipped with the repository as reference
policies. Rungs 4–6 are external consumers reached through the operator endpoint and live in
a bench harness outside the binary.

## 3. Where each piece lives

| Piece | Location | Reason |
| --- | --- | --- |
| Rungs 0–3 | `internal/refconsumer` as selectable policies | The reference consumer is already a non-negotiable; stronger versions are the same instrument |
| Paired comparison, intervals, held-out protocol | `internal/score` / `internal/suite` | Consumer-agnostic statistics over scorecards |
| Rungs 4–6 and cost accounting of model calls | Bench repository or the consumer's own eval suite | The simulator cannot observe a consumer's internal calls and must not know its name |
| Third-party baselines | Same operator-endpoint interface | Anyone can submit a rung 2 competitor, which defuses the straw-man objection |

## 4. Fairness rules

1. **Identical inputs**: same seeds, same perturbation stream, same quiescence barrier.
2. **Identical tools**: same effector list, same nameplate; no rung sees truth.
3. **Equal tuning budget**: a policy tuned on training seeds gets the same number of tuning
   scenarios as the consumer; tuning on evaluation seeds is disqualifying.
4. **Cost is reported, not hidden**: model calls, tokens and dollars per correct outcome sit
   beside accuracy.
5. **No scenario selection after the fact**: the suite is generated and sealed before any
   rung runs; exclusions come only from the audits, with reasons recorded.

## 5. Statistical protocol

| Concern | Rule |
| --- | --- |
| Sample size | At least `ground_truth.min_scenarios` (100 in shipped domains), with `min_per_label` |
| Pairing | Compare consumers on identical seeds; report paired differences |
| Uncertainty | Percentile bootstrap, 10,000 resamples, 95 % interval, seeded for reproducibility |
| Headline metrics | Regret, unsafe calls, false-action cost, deadline adherence, cost per correct outcome |
| Multiple comparisons | Pre-register the headline metrics; report others as exploratory |
| Negative class | Always reported; a consumer that wins by acting on everything is penalised by false-action cost |

Reports show intervals; a point estimate alone is rejected by the report generator.

## 6. Held-out seeds

Tuning on a public suite is the dominant failure mode of benchmarks (see
`../research/PRIOR_ART.md` §5). Protocol:

1. Publisher commits to a hash of a secret seed list per suite version.
2. Consumers tune on a published training suite.
3. After submission the seed list is revealed, the held-out suite is generated, and the hash
   is verified.
4. Each suite version is used once for public claims and then becomes training data for the
   next.

The determinism guarantee makes this auditable: the reveal plus the domain and adapter
digests reproduce every scenario exactly.

## 7. Scorecard-of-scorecards

A suite report is a single JSON document, schema-versioned:

```json
{
  "suite_id": "...", "suite_digest": "sha256:...",
  "simulator_version": "0.x", "bundle": "scorecard-bundle-v0.3",
  "seeds": {"mode": "held_out", "commitment": "sha256:..."},
  "rungs": [
    {"id": "rules-strong@1.2", "kind": "reference",
     "metrics": {"regret": {"mean": 0, "ci95": [0, 0]}, "unsafe_calls": 0}}
  ],
  "paired": [{"a": "product-full", "b": "rules-strong@1.2",
              "metric": "regret", "diff": {"mean": 0, "ci95": [0, 0]}, "wins": 0, "n": 0}],
  "excluded": [{"scenario": "...", "reason": "no_headroom"}],
  "sensitivity": {"economics_scale": [0.7, 1.3], "ranking_stable": true}
}
```

## 8. What a good outcome looks like

- At least one scenario family where rung 2 loses decisively, with the named mechanism that
  wins.
- At least one family where the product does **not** beat rung 2, published beside the wins.
  A suite with no losses is a suite no one will believe.
- Cost per correct outcome showing where model spend is not justified.

## 9. Test plan

1. Reference policies: deterministic output for a fixed seed; version string in every
   scorecard.
2. Paired bootstrap: known synthetic distributions recover known differences and coverage.
3. Held-out: commitment verifies; a tampered seed list fails verification.
4. Fairness: a policy that reads sealed truth through any operator tool fails the isolation
   test (existing oracle-isolation coverage extends to the new policies).
5. Report generator rejects metrics without intervals.

## 10. Open questions

1. Is rung 3 (expert runbook) authored per domain as data, or derived from
   `expected_effector` per label? The latter is free but optimistic.
2. How is tuning budget measured for a model-based consumer: scenarios, calls, dollars?
3. Who hosts the commitment and reveal so the protocol is credible without a service?
4. Should the third-party baseline interface be a documented operator-endpoint profile, or a
   manifest file the suite runner consumes?
