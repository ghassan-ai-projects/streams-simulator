# Streams Simulator documentation

This is the curated documentation set for Streams Simulator (`streamsim`), a deterministic, closed-loop world simulator for testing stream processors.

The repository has two documentation layers:

- **`documentation/`** is the public set: product explanation, getting started, architecture, guides, operations, benchmark methodology, and reference material.
- **`docs/`** is the engineering archive: detailed design records, research notes, plans, reviews, audits, fixtures, and machine-readable contract files. It is authoritative for design intent and historical evidence, not automatically for current behavior.

Read [limitations](limitations.md) before treating the current checkout as a release. Read [the quality bar](QUALITY_BAR.md) before changing the documentation set.

## Choose a path

### I want to understand the product

- [Product](overview/product.md) — what the simulator is, is not, and who it serves
- [Concepts](overview/concepts.md) — worlds, events, perturbations, truth, and scoring
- [Architecture overview](architecture/overview.md) — the components and trust boundaries
- [Pipeline](architecture/pipeline.md) — how one event moves through the system

### I want to run it

- [Install and build](getting-started/install.md)
- [Quickstart](getting-started/quickstart.md)
- [CLI reference](reference/cli.md)
- [Replay and debugging](guides/replay-and-debugging.md)

### I want to integrate a consumer

- [Consumer integration](guides/consumer-integration.md)
- [MCP reference](reference/mcp.md)
- [Truth and scoring](architecture/truth-and-scoring.md)
- [Artifact reference](reference/artifacts.md)

### I want to extend the simulator

- [Domains and adapters](architecture/domains-and-adapters.md)
- [Authoring domains](guides/authoring-domains.md)
- [Authoring adapters](guides/authoring-adapters.md)
- [Contracts](reference/contracts.md)
- [Design index](design/README.md)

### I want to evaluate or publish results

- [Benchmark overview](benchmark/README.md)
- [Methodology](benchmark/methodology.md)
- [Evidence and release posture](benchmark/evidence.md)
- [Benchmark workflow](guides/benchmark-workflow.md)
- [Limitations](limitations.md)

### I want to contribute

- [Contributor quality guide](governance/quality.md)
- [Claim-to-source matrix](governance/claims-matrix.md)
- [Architecture decisions](adr/README.md)
- [Repository contribution rules](../CONTRIBUTING.md)
- [Security policy](../SECURITY.md)
- [Roadmap](roadmap.md)

## Reader journeys

- **Evaluator:** [product](overview/product.md) → [quickstart](getting-started/quickstart.md) → [methodology](benchmark/methodology.md) → [limitations](limitations.md) → [evidence](benchmark/evidence.md)
- **Consumer integrator:** [quickstart](getting-started/quickstart.md) → [consumer integration](guides/consumer-integration.md) → [MCP](reference/mcp.md) → [contracts](reference/contracts.md) → [artifacts](reference/artifacts.md)
- **Domain author:** [concepts](overview/concepts.md) → [domain authoring](guides/authoring-domains.md) → [domain contract](reference/contracts.md) → [validation](architecture/domains-and-adapters.md)
- **Adapter author:** [concepts](overview/concepts.md) → [adapter authoring](guides/authoring-adapters.md) → [adapter contract](reference/contracts.md) → [golden verification](guides/authoring-adapters.md#workflow)
- **Failure investigator:** [operations](operations/operations.md) → [troubleshooting](operations/troubleshooting.md) → [replay](guides/replay-and-debugging.md) → [artifacts](reference/artifacts.md)
- **Security reviewer:** [security model](architecture/security-model.md) → [invariants](architecture/invariants.md) → [security policy](../SECURITY.md) → [injection research](research/benchmark-validity.md)
- **Maintainer/release owner:** [contributing](../CONTRIBUTING.md) → [quality](governance/quality.md) → [release](operations/release.md) → [manifest evidence](benchmark/evidence.md)
- **Researcher:** [research index](research/README.md) → [prior art](research/prior-art.md) → [benchmark validity](research/benchmark-validity.md)

## Documentation authority

Public pages explain the current product in reader-friendly terms. They do not replace the machine-readable contracts or the design record:

| Question | Authoritative source |
| --- | --- |
| What JSON is accepted or emitted? | Canonical schemas under [`docs/contracts/`](../docs/contracts/); embedded copies under [`internal/schemas/`](../internal/schemas/) are verified against them. |
| What is implemented? | Go packages under [`internal/`](../internal/) plus their tests. |
| Why does the architecture work this way? | Accepted deviations in [`docs/DECISIONS.md`](../docs/DECISIONS.md); design history under [`docs/design/`](../docs/design/) is status-labeled. |
| What evidence supports a release claim? | Exact-commit CI output and manifest, summarized in [benchmark evidence](benchmark/evidence.md). |

If a public page and a contract disagree, trust the contract and open a documentation issue. If a public page and runtime behavior disagree, trust code/tests and update the page. If a public page describes a behavior that is only designed, it must say so.

## Design and archive map

- [Public design index](design/README.md)
- [Working archive index](../docs/README.md)
- [Documentation organization plan](../docs/DOCUMENTATION_ORGANIZATION_PLAN.md)
- [Documentation quality bar](QUALITY_BAR.md)

## Project policies

- [License](../LICENSE)
- [Contributing](../CONTRIBUTING.md)
- [Security](../SECURITY.md)
- [Code of conduct](../CODE_OF_CONDUCT.md)
- [Support](../SUPPORT.md)
- [Changelog](../CHANGELOG.md)

## Next reads

- New reader: [product](overview/product.md) → [concepts](overview/concepts.md) → [quickstart](getting-started/quickstart.md)
- Consumer author: [consumer integration](guides/consumer-integration.md) → [MCP reference](reference/mcp.md) → [artifacts](reference/artifacts.md)
- Maintainer: [architecture](architecture/overview.md) → [invariants](architecture/invariants.md) → [quality guide](governance/quality.md)
