# Design index

This section gives the public reading order for the simulator design. The detailed design record remains under [`docs/design/`](../../docs/design/) and contains historical plans, reviews, and known contradictions; public summaries must not silently promote those records to implemented behavior. Current runtime behavior is owned by code/tests, contracts by `contracts/`, and accepted deviations by `docs/DECISIONS.md`.

## Current design spine

1. [Transport analysis](../../docs/research/TRANSPORT_ANALYSIS.md) — why MCP carries control while sinks carry evidence.
2. [Technical design](../../docs/design/TECHNICAL_DESIGN.md) — the original system model; read with [decisions](../../docs/DECISIONS.md) for accepted deviations.
3. [MCP surface](../../docs/design/MCP_SURFACE.md) — detailed director/operator contract; implemented tool names are summarized in [the MCP reference](../reference/mcp.md).
4. [Consumer contract](../../docs/design/CONSUMERS.md) — the reference consumer and integration boundary.
5. [Ground truth and scoring](../../docs/design/GROUND_TRUTH_AND_SCORING.md) — oracle, ledger, action, and benchmark rules.
6. [Domain catalog](../../docs/design/DOMAIN_CATALOG.md) — specified coverage and build order; specified is not shipped.

## Later design records

These records in `docs/design/` and `docs/` describe surfaces added after the original spine; each carries its own status note, so read it before relying on the text.

- [Domain spec v0.2](../../docs/design/DOMAIN_SPEC_V0_2.md), [components](../../docs/design/COMPONENTS_V0_1.md) and [scenario file](../../docs/design/SCENARIO_FILE.md) — worlds as data.
- [Run journal](../../docs/design/RUN_JOURNAL.md) and [run report](../../docs/design/RUN_REPORT.md) — what a run records and how it is summarised.
- [CLI surface](../../docs/design/CLI_SURFACE.md) and [canvas bridge](../../docs/design/CANVAS_BRIDGE.md) — the CLI and the canvas as clients of the same modules.
- [Rule promotion](../../docs/rule-promotion/PLAN.md) — campaigns that play a consumer with and without learning (work in progress).

## Physical devices

[Physical devices](../../docs/design/PHYSICAL_DEVICES.md) is the design record for kind files, twins, the relay, mixed sessions, the live control plane, node interrupts and the twin report, with an as-built note for each round. It is implemented; the public summary is [the guide](../guides/physical-devices.md), [the architecture page](../architecture/physical-devices.md) and [the limitations](../limitations.md#physical-devices).

## Status discipline

The archive contains documents written at different stages. In particular, older technical-design text may describe SQLite, Go versions, or commands that do not match the current implementation. The accepted implementation uses file-based JSON artifacts; verify current behavior against [`internal/`](../../internal/) and [the limitations page](../limitations.md).

## Historical review material

The implementation plan, gap analysis, critical review, readiness review, and MCP interaction review remain useful for provenance and adversarial context. They are not current product reference pages. Use them to understand why a guarantee exists, then verify the current code and tests.

## Next reads

- [Architecture overview](../architecture/overview.md)
- [Nine non-negotiables](../architecture/invariants.md)
- [Working archive](../../docs/README.md)
