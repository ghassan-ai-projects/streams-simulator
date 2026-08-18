# Benchmark methodology

> Status: Implemented methodology with partial batch automation. Authority: scoring/audit contracts and design records. Verified by: audit and scoring tests. Last verified: 2026-08-17.

The benchmark design is built around four ways a stream benchmark can flatter a consumer: trivial scenarios, transport ambiguity, oracle leakage, and action acknowledgements that are mistaken for effects.

## Scenario validity

Before a scenario enters a graded suite:

1. The scenario declares the property it is intended to stress.
2. The actual delivered stream is used by the baseline, in order.
3. A trivial detector is fitted with hindsight against the scenario.
4. If that baseline solves the scenario for the wrong reason, the scenario is excluded or redesigned.
5. The accepted scenario is generated from a byte-stable seed and profile.

The audit is a quality control on the suite, not a score for the consumer.

## Evidence-aware scoring

The scorer must use the delivery ledger to classify transport misses separately from reasoning misses. A consumer cannot be penalized for an event it never received, and it cannot receive credit for a conclusion that depends on an event the ledger says was not delivered.

The action tuple includes the command identity, declared effector, target entity, acknowledgement, and resulting world evidence. `silent_no_effect` is deliberately scored against the resulting world state rather than the acknowledgement alone.

## Reproducibility

Every executed accepted scenario must retain its domain and adapter digests, simulator version, seed, profile, command log, sink choice, truth label, ledger, verdict, and score. Suite generation supplies the scenario definition and label; the execution harness supplies the selected adapter and run evidence. The exact artifact set must reproduce before a result is compared with another consumer.

## What to report

Report per-scenario results before aggregate summaries:

- detection and timing outcomes;
- transport versus reasoning classification;
- action outcomes and silent failures;
- incomplete or ambiguous evidence;
- simulator, domain, adapter, suite, and consumer digests;
- the manifest and validation commands used.

## Source record

The detailed rationale and prior-art analysis are in [`docs/design/GROUND_TRUTH_AND_SCORING.md`](../../docs/design/GROUND_TRUTH_AND_SCORING.md), [`docs/design/DOMAIN_CATALOG.md`](../../docs/design/DOMAIN_CATALOG.md), and [`docs/research/PRIOR_ART.md`](../../docs/research/PRIOR_ART.md). Those records contain design history and must be read with their status labels.

## Next reads

- [Benchmark overview](README.md)
- [Evaluation workflow](../guides/benchmark-workflow.md)
- [Truth and scoring](../architecture/truth-and-scoring.md)
