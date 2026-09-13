# Release and publication procedure

> Status: Implemented release procedure; repository gates are green at the snapshot commit, and publication evidence is generated per release. Authority: CI, Makefile, manifest command, and release evidence. Verified by: full `make ci-check` run and manifest smoke command. Last verified: 2026-09-13.

The project must be able to prove what it released and what it measured. A release claim is bound to an exact commit and evidence bundle.

## Pre-release checks

Run the following from an intentional working tree:

```bash
go test ./...
go vet ./...
git diff --check
make ci-check
make soak
make perf
make manifest MANIFEST_AUTHOR="Name <author@example>" MANIFEST_REVIEWER="Reviewer <review@example>"
```

Retain the outputs, tool versions, and generated manifest. If the repository has a known baseline failure, do not describe the release as green.

## Manifest rules

The manifest binds simulator identity, toolchain, domains, adapters, consumer, suite settings, author, and independent reviewer identity. Generate it for the exact commit whose artifacts are being published. A checked-in manifest from another commit is historical evidence only.

## Publication rules

- Do not publish benchmark scores without replay identity, fresh gate evidence, and a matching manifest.
- Publish per-scenario and aggregate results with transport/reasoning classification.
- State the exact domain and adapter inventory.
- State known limitations and unsupported paths.
- Announce contract or behavior changes in the changelog when a release process exists.

## Current status

The repository gates are green at the snapshot commit and the shipped inventory is eight committed domains. A release claim still requires the pre-release checks above from a clean, intentional commit, with a manifest generated for that exact commit. See [benchmark evidence](../benchmark/evidence.md) and [limitations](../limitations.md).

## Next reads

- [Benchmark evidence](../benchmark/evidence.md)
- [Compatibility](../overview/compatibility.md)
- [Contributor quality guide](../governance/quality.md)
