# Streams Simulator

Streams Simulator (`streamsim`) is a deterministic, closed-loop world simulator for testing stream processors.

It generates event streams from data-defined domains, perturbs delivery in declared ways, carries sealed ground truth, accepts commands through declared effectors, and scores a consumer’s verdict against what actually happened.

> The repository is tested green at its current commit, but no release claim is made from a working tree: see [current limitations](documentation/limitations.md) before publishing benchmark results, and generate fresh evidence for the exact commit.

## Why it exists

This is a test instrument. An instrument less trustworthy than the system it measures is worse than no instrument. Streams Simulator therefore treats determinism, generative truth, delivery classification, closed-loop actuation, and adversarial validity controls as product requirements.

The simulator has no knowledge of its consumers. Domains and adapters are JSON data; a new consumer should not require a consumer-specific branch in the binary.

## The pipeline

```text
domain JSON → world → perturbation → adapter → sink → consumer
                  │                         │
                  ├── sealed truth          └── verdict
                  └── delivery ledger ───────────┘
                                      scorer
```

The Model Context Protocol (MCP) carries control, actuation, and audit. The configured sink carries event evidence. The operator surface is capability-scoped and does not expose hidden truth.

## Quick start

Requirements: Go `1.25.12` or a compatible toolchain; see [install](documentation/getting-started/install.md).

```bash
go run ./cmd/streamsim help
go run ./cmd/streamsim catalog list
go run ./cmd/streamsim adapter verify adapters/native-jsonl.adapter.json
go run ./cmd/streamsim run --domain rotating-machinery --adapter native-jsonl --seed 7 --sink file --out /tmp/streamsim-run --duration 120
go run ./cmd/streamsim verify /tmp/streamsim-run/run.json
```

The complete walkthrough is [the quickstart](documentation/getting-started/quickstart.md). The CLI prints JSON results; run artifacts include the trace, ledger, state history, and replay metadata.

## Documentation

The curated public documentation is under [`documentation/`](documentation/README.md):

- [Product and concepts](documentation/overview/product.md)
- [Architecture and nine non-negotiables](documentation/architecture/overview.md)
- [Consumer integration and MCP](documentation/guides/consumer-integration.md)
- [Domains, adapters, and contracts](documentation/architecture/domains-and-adapters.md)
- [Benchmark methodology and evidence](documentation/benchmark/README.md)
- [Limitations and roadmap](documentation/limitations.md)
- [Contributor quality guide](documentation/governance/quality.md)

The existing [`docs/`](docs/README.md) tree is the engineering archive: detailed design records, research, plans, reviews, audits, fixtures, and canonical contract files.

## Build and test

```bash
make build
go test ./...
go vet ./...
make ci-check
```

`make ci-check` includes release-gate tools that intentionally fail closed when unavailable. The slower `make soak`, `make perf`, and `make manifest` checks are documented in [release procedure](documentation/operations/release.md).

## Open-source policies

- [Contributing](CONTRIBUTING.md)
- [Security](SECURITY.md)
- [License](LICENSE)
- [Code of conduct](CODE_OF_CONDUCT.md)
- [Support](SUPPORT.md)
- [Changelog](CHANGELOG.md)

## Current inventory note

Eight domain files are committed in the release inventory alongside two adapters. The schema suite validates every shipped domain from the `domains/` directory, so new domain files must be added together with their test and inventory updates. See [limitations](documentation/limitations.md) for the current status boundary.
