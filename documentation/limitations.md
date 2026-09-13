# Limitations and current status

> Status: Current working-tree report. Authority: repository state, tests, and implementation review. Verified by: full `go test ./...` run, `make ci-check`, and inventory review. Last verified: 2026-09-13.

This page is the honest counterpart to the product overview. It separates implemented capability from release evidence and from design intent.

## Current working-tree status

The current checkout implements the documented pipeline and passes the repository gates at its commit. The remaining observations are about release evidence and environment boundaries, not about a red suite:

| Observation | Evidence | Consequence |
| --- | --- | --- |
| Eight domains are shipped and validated together. | [`domains/`](../domains/) contains eight committed domain files; the embedded-schema suite validates each one. | New domains must land together with test, inventory, and manifest updates in the same change. |
| Some integration tests need local loopback sockets. | MCP operator endpoint tests and HTTP-push equivalence tests bind local listeners. | A restricted execution environment may fail those tests with `operation not permitted`; rerun in a network-capable environment before interpreting the result. |
| The checked-in release manifest is historical. | [`release-manifest.json`](../release-manifest.json) pins the commit and toolchain it was generated at. | Generate a fresh manifest for the exact commit before publishing evidence. |
| Design and implementation records disagree in places. | Earlier design text describes SQLite, Go 1.26, and commands not present in the current CLI; decisions and code record file artifacts and Go 1.25.12. | Treat archive design text as status-labeled source material, not an unqualified promise. |
| The repository lacks a formal compatibility matrix and release channel. | No current support matrix or release policy establishes one. | Validate the exact commit and toolchain for every external result. |
| Documentation checking is deliberately lightweight. | `make docs-check` checks required pages and executable smoke commands; it does not fetch links or parse every Markdown construct. | Review links, commands, fences, and diagrams manually as part of contribution review. |

## Scope boundaries

- The simulator is not a production broker, controller, physics engine, or hosted service.
- HTTP-push is an integration sink, not a deterministic local benchmark transport.
- The operator MCP role is not a standalone CLI mode; it is served by a director process.
- The score command reads `verdict.json` and `ledger.jsonl` beside the run artifact and `label.json` beside it unless `--label` is provided. It does not accept a `--verdict` flag.
- The offline score command does not receive an effector-call log, so action-loop metrics are incomplete. Use the MCP director score path when action evidence is required.
- Suite generation is not a complete batch runner: it produces audited scenario definitions and labels, while a harness must choose an adapter, execute scenarios, and persist score output.
- The CLI replay and verify commands require current `domains/` and `adapters/` directories even when the run artifact contains embedded specifications; the replay library is less restrictive.
- The current code uses file-based JSON/JSONL run artifacts; older SQLite language in the design archive is historical or superseded by the accepted decision.
- No actual external production consumer is shipped; the reference consumer is a test instrument and integration example.

The `streamsim device serve` command is a deterministic emulator gateway link,
not a physical-HIL claim. Its capability catalog is a simulator-side device
contract and must be kept aligned with the paired Agentic Stream catalog and
allow-listed digest before a cross-repository handshake can be treated as
evidence. When `--world` is supplied, the CLI binds the emulator to a
simulated world plant through `internal/deviceworld`; without it, the device
uses its local catalog projection. World bindings are loaded from structured
configuration; the adapter does not contain per-domain argument callbacks.

## What is not a release claim

The presence of tests for all nine non-negotiables is valuable evidence, but it is not itself a green release gate. A release claim requires a clean run of the required checks, a matching manifest, reproducible artifacts, and an explicit review identity.

## How to clear the limitations

1. ~~Decide whether `cold-chain-transit` is shipped~~ Done: the domain is committed and validated with the rest of the inventory.
2. ~~Reconcile stale design, README, context, and decision text~~ Done for the public set and agent instructions; archive text under `docs/` remains status-labeled by decision.
3. ~~Add or document compatibility, support, conduct, changelog, and release policy~~ Done: see [compatibility](overview/compatibility.md), [SUPPORT](../../SUPPORT.md), [CODE_OF_CONDUCT](../../CODE_OF_CONDUCT.md), [CHANGELOG](../../CHANGELOG.md), and the [release procedure](operations/release.md).
4. Extend `docs-check` with local link/fragment and example checks, then keep it in CI; manual diagram and prose review will still be required.
5. Add a suite runner that records the selected adapter and persists per-scenario score evidence.
6. Generate fresh release evidence (gates, soak, perf, manifest) from a clean, intentional commit at publication time.

## Next reads

- [Roadmap](roadmap.md)
- [Release procedure](operations/release.md)
- [Benchmark evidence](benchmark/evidence.md)
- [Quality bar](QUALITY_BAR.md)
