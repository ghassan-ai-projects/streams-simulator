# Benchmark and release evidence

> Status: Implemented evidence requirements; current checkout is not release-green. Authority: CI, manifest command, artifacts, and tests. Verified by: smoke validation and current working-tree review. Last verified: 2026-08-17.

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
make manifest MANIFEST_AUTHOR="Name <author@example>" MANIFEST_REVIEWER="Reviewer <review@example>"
```

These commands are meaningful only when their required tools are present and the output is retained. The `ci-check` target intentionally fails closed for missing release-gate tools.

## Current repository status

The checked-in [`release-manifest.json`](../../release-manifest.json) describes an older commit and toolchain, and the current working tree contains an untracked seventh domain that makes the full test suite fail its six-domain assertion. Treat the manifest as historical evidence until a fresh manifest is generated after reconciliation.

## Evidence ownership

- Code and tests establish implemented behavior.
- `docs/contracts/` establishes contract shape.
- The run artifact establishes the experiment inputs and command history.
- The ledger establishes delivery outcomes.
- The manifest binds a release claim to exact inputs and identities.

## Next reads

- [Release procedure](../operations/release.md)
- [Artifact reference](../reference/artifacts.md)
- [Limitations](../limitations.md)
