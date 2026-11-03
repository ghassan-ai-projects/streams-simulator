# Streams Simulator

Streams Simulator (`streamsim`) is a deterministic, closed-loop world simulator. It builds a plant, a fleet, a site or any other system as data, runs it on a seeded clock, delivers what it produces to whatever consumes it, takes commands back so the world really changes, and keeps the hidden truth of everything that happened.

What you do with that world is up to you:

- **Build and explore worlds.** Compose a plant from a library of equipment, instruments and vocabulary, or write a domain as JSON; run it, inject faults, perturb delivery, watch it in the visual canvas.
- **Feed a stream processor, agent or controller.** Project events into the consumer's own wire format, hand it effectors to act with, and let it close the loop.
- **Grade it.** Sealed ground truth, a delivery ledger and independent analytic oracles score what the consumer detected and did, offline or online, and compare runs.
- **Put a real board beside simulated ones.** The plant wire connects the simulator, a gateway and a host so physical and simulated devices share one session, with interrupts, approvals and a twin-versus-board report.
- **Run campaigns.** Generate many seeded one-fault runs, play them against a consumer with and without learning, and score the lot (the rule-promotion work uses this).
- **Reproduce and audit.** A run is a pure function of its inputs; one artifact replays any result byte for byte.

The simulator has no knowledge of its consumers. Domains, adapters and effectors are JSON data; a new consumer never needs a branch in the binary.

> The repository passes its gates at the current commit, but no release claim is made from a working tree: see [current limitations](documentation/limitations.md) before publishing results, and generate fresh evidence for the exact commit.

## Why you can trust what it says

A simulator that is less trustworthy than what it drives gives confident wrong answers, so determinism and truth come before features. Nine non-negotiables hold whatever the use: determinism, analytic cross-check, a reference consumer, a delivery ledger, a quiescence barrier, a sealed oracle, an injection probe, a `silent_no_effect` test and a trivial-baseline audit. See [invariants](documentation/architecture/invariants.md).

## The pipeline

```text
domain JSON → world → perturbation → adapter → sink → consumer
                  │                         │
                  ├── sealed truth          └── verdict
                  └── delivery ledger ───────────┘
                                      scorer
```

The Model Context Protocol (MCP) carries control, actuation and audit; the configured sink carries the event stream. The operator surface is capability-scoped and never exposes hidden truth. The CLI, the director MCP tools and the canvas are equal clients of the same modules.

## What is in the repository

| Area | Where | Documented in |
| --- | --- | --- |
| Binary, CLI and MCP server | `cmd/streamsim`, `internal/` | [CLI reference](documentation/reference/cli.md), [MCP reference](documentation/reference/mcp.md) |
| Worlds as data: 8 shipped domains, 2 adapters, the component library, example plants | `domains/`, `adapters/`, `examples/`, `internal/component` | [domains and adapters](documentation/architecture/domains-and-adapters.md), [authoring domains](documentation/guides/authoring-domains.md) |
| Visual canvas (TypeScript, a client of the binary) | `canvas/` | [canvas README](canvas/README.md) |
| Physical devices and the plant wire | `contracts/plant-wire/`, `internal/plantwire`, `internal/device*` | [physical devices guide](documentation/guides/physical-devices.md) |
| Runs, journals, reports, grading, suites and campaigns | `internal/run`, `journal`, `report`, `grading`, `suite` | [artifacts](documentation/reference/artifacts.md), [benchmark workflow](documentation/guides/benchmark-workflow.md) |
| JSON Schema contracts | `contracts/` | [contracts](documentation/reference/contracts.md) |

## Quick start

Requirements: Go `1.25.12` or a compatible toolchain; see [install](documentation/getting-started/install.md).

```bash
go run ./cmd/streamsim help
go run ./cmd/streamsim catalog list
go run ./cmd/streamsim adapter verify adapters/native-jsonl.adapter.json
go run ./cmd/streamsim run --domain rotating-machinery --adapter native-jsonl --seed 7 --sink file --out /tmp/streamsim-run --duration 120
go run ./cmd/streamsim verify /tmp/streamsim-run/run.json
```

The complete walkthrough is [the quickstart](documentation/getting-started/quickstart.md). The CLI prints JSON documents; run folders hold the trace, ledger, state history, journal and replay metadata. The canvas needs Node 24 and is started with `streamsim canvas`; see the [canvas README](canvas/README.md).

## Documentation

Start at [`documentation/README.md`](documentation/README.md); it is the current, public documentation set.

- [Product and concepts](documentation/overview/product.md)
- [Architecture and the nine non-negotiables](documentation/architecture/overview.md)
- [Consumer integration and MCP](documentation/guides/consumer-integration.md)
- [Physical devices](documentation/guides/physical-devices.md)
- [Benchmark methodology and evidence](documentation/benchmark/README.md)
- [Limitations and roadmap](documentation/limitations.md)
- [Contributor quality guide](documentation/governance/quality.md)

[`docs/`](docs/README.md) is a dated working archive of design notes, research, plans and reviews. It records how decisions were made; where it disagrees with `documentation/`, the code or `contracts/`, the archive is out of date.

## Build and test

```bash
make quick        # inner loop
make ci-check     # full pipeline
make phase-gate   # ci-check plus canvas gates and the architecture check
```

`make ci-check` includes release-gate tools that intentionally fail closed when unavailable. The slower `make soak`, `make perf` and `make manifest` checks are in the [release procedure](documentation/operations/release.md).

## Contributing and policies

[Contributing](CONTRIBUTING.md) · [Security](SECURITY.md) · [License](LICENSE) · [Code of conduct](CODE_OF_CONDUCT.md) · [Support](SUPPORT.md) · [Changelog](CHANGELOG.md). Coding agents start at [AGENTS.md](AGENTS.md).
