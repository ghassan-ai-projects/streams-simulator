# Release and publication procedure

> Status: Implemented release procedure; repository gates are green at the snapshot commit, and publication evidence is generated per release. Authority: CI, Makefile, manifest command, and release evidence. Verified by: full `make ci-check` run and manifest smoke command. Last verified: 2026-11-03.

The project must be able to prove what it released and what it measured. A release claim is bound to an exact commit and evidence bundle.

## Pre-release checks

Run the following from an intentional working tree:

```bash
make quick          # inner loop while working
make ci-check
make phase-gate     # ci-check + canvas-check + Enola architecture + behaviour pin
make soak
make perf
make manifest MANIFEST_AUTHOR="Name <author@example>" MANIFEST_REVIEWER="Reviewer <review@example>"
```

`ci-check` already runs `go vet`, `docs-check` (including `git diff --check`), `gate-ratchet` (no gate cap, allow-list or coverage floor looser than at the branch point from main), the short test tier, `deadcode`, `vulncheck` and bounded fuzzing. `make test-long` and `make test-race` run on demand only. Run `make phase-gate` once per pull request; it needs Node 24 and the Enola tool. Retain the outputs, tool versions, and generated manifest. If the repository has a known baseline failure, do not describe the release as green.

## Manifest rules

The manifest binds simulator identity, toolchain, domains, adapters, consumer, suite settings, author, and independent reviewer identity. Generate it for the exact commit whose artifacts are being published. A checked-in manifest from another commit is historical evidence only.

## Publication rules

- Do not publish benchmark scores without replay identity, fresh gate evidence, and a matching manifest.
- Publish per-scenario and aggregate results with transport/reasoning classification.
- State the exact domain and adapter inventory.
- State known limitations and unsupported paths.
- Announce contract or behavior changes in [`CHANGELOG.md`](../../CHANGELOG.md).

## Current status

The shipped inventory is eight committed domains and two adapters; the checked-in manifest records commit `daad36e` and placeholder identities. A working tree passing its gates is not a release claim; a release claim requires the pre-release checks above from a clean, intentional commit, with a manifest generated for that exact commit. See [benchmark evidence](../benchmark/evidence.md) and [limitations](../limitations.md).

## Next reads

- [Benchmark evidence](../benchmark/evidence.md)
- [Compatibility](../overview/compatibility.md)
- [Contributor quality guide](../governance/quality.md)
