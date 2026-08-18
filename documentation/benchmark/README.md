# Benchmarking Streams Simulator consumers

> Status: Partially implemented benchmark workflow. Authority: suite, audit, score, and manifest packages. Verified by: package tests and the limitations recorded below. Last verified: 2026-08-17.

Streams Simulator is useful for benchmark work only when the benchmark measures consumer reasoning rather than transport accidents, leaked labels, or a simulator bug.

## What is being measured

A benchmark run combines:

- a data-defined domain and scenario profile;
- a seeded world and command log;
- a declared delivery path and perturbation set;
- a consumer-visible adapter and sink;
- sealed ground truth and delivery ledger evidence;
- a consumer verdict and, where applicable, effector actions;
- a score produced from the complete evidence bundle.

The simulator is not the consumer. The same run can be used to compare consumers only when the domain, adapter, perturbations, timing, and evidence rules are held constant.

## Benchmark readiness

The repository has suite generation, trivial-baseline auditing, scoring, release-manifest generation, and tests for the nine non-negotiables. Suite generation is not yet a complete batch runner: it emits scenario definitions and labels, while a harness must choose the adapter, execute each scenario, collect run evidence, and invoke scoring. The current working tree is not a clean release baseline, so no benchmark result should be published from it without fresh validation and a manifest for the exact commit and inputs.

## Recommended reading order

1. [Methodology](methodology.md)
2. [Evaluation workflow](../guides/benchmark-workflow.md)
3. [Truth and scoring](../architecture/truth-and-scoring.md)
4. [Release evidence](evidence.md)
5. [Limitations](../limitations.md)

## Non-goals

- Optimizing a consumer against a single hand-fitted scenario.
- Treating a missing event as a consumer error without ledger evidence.
- Publishing a single aggregate number without per-scenario evidence and reproducible inputs.

## Next reads

- [Nine non-negotiables](../architecture/invariants.md)
- [Artifact reference](../reference/artifacts.md)
