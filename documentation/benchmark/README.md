# Grading a consumer

> Status: Partially implemented grading and benchmark workflow. Authority: suite, audit, score, grading and manifest packages. Verified by: package tests and the limitations recorded below. Last verified: 2026-11-03.

Grading a consumer is one use of Streams Simulator, not the whole product: worlds can also be built, explored, driven from the canvas, joined by physical devices through the plant wire, and run as campaigns (see the [product overview](../overview/product.md)). This section covers the grading use. It is useful for benchmark work only when the benchmark measures consumer reasoning rather than transport accidents, leaked labels, or a simulator bug.

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

The repository has suite generation (`streamsim suite`), campaign generation of seeded one-fault runs (`streamsim suite campaign`), trivial-baseline auditing, online and offline scoring (`streamsim score`), run reports, release-manifest generation, and tests for the nine non-negotiables. Suite and campaign generation are not a complete batch runner: they emit scenario definitions, session plans and labels, while a harness must choose the adapter, execute each scenario, collect run evidence, and invoke scoring. No benchmark result should be published without fresh validation and a manifest for the exact commit and inputs.

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
