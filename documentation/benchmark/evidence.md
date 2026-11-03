# Benchmark and release evidence

This page covers the evidence behind a graded result (one use of the simulator) and, since the release gates are shared, behind a release claim.

> Status: Implemented evidence requirements; repository gates are green at the snapshot commit, with per-release evidence generation. Authority: CI, manifest command, artifacts, and tests. Verified by: smoke validation, full `make ci-check` run, and working-tree review. Last verified: 2026-11-03.

Evidence is part of the benchmark output. A prose statement such as “all gates are green” is not a substitute for the exact command, commit, toolchain, inputs, and artifact that support it.

## Minimum evidence bundle

- exact simulator commit and version;
- Go toolchain version;
- domain, adapter, and suite digests;
- seed, profile, command log, and sink configuration;
- run artifact and trace/ledger outputs;
- sealed truth/label and consumer verdict;
- score output;
- results of replay/verification;
- release manifest with author and independent-review identity where required.

## Release gates

Before publishing benchmark numbers, run the project’s required checks and the relevant slow checks:

```bash
make ci-check
make soak
make perf
make canvas-check   # when the canvas or its contracts changed
make manifest MANIFEST_AUTHOR="Name <author@example>" MANIFEST_REVIEWER="Reviewer <review@example>"
```

These commands are meaningful only when their required tools are present and the output is retained. The `ci-check` target intentionally fails closed for missing release-gate tools. `soak`, `perf`, `test-long` and `test-race` are not part of `ci-check` or CI and run on demand; `phase-gate` adds `canvas-check`, the Enola `architecture` check and `behaviour-pin-check` to `ci-check`.

## Current repository status

The shipped inventory is eight committed domains and two adapters. The checked-in [`release-manifest.json`](../../release-manifest.json) records the commit it was generated at (`daad36e`) and placeholder author and reviewer identities, so it is not a release claim for any later commit; regenerate it for the exact commit whose evidence you publish; a manifest from another commit is historical evidence only.

## Evidence ownership

- Code and tests establish implemented behavior.
- `contracts/` establishes contract shape.
- The run artifact establishes the experiment inputs and command history.
- The ledger establishes delivery outcomes.
- The manifest binds a release claim to exact inputs and identities.

## Next reads

- [Release procedure](../operations/release.md)
- [Artifact reference](../reference/artifacts.md)
- [Limitations](../limitations.md)
