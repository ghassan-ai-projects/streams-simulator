# Design index

This section gives the public reading order for the simulator design. The detailed design record remains under [`docs/design/`](../../docs/design/) and contains historical plans, reviews, and known contradictions; public summaries must not silently promote those records to implemented behavior. Current runtime behavior is owned by code/tests, contracts by `docs/contracts/`, and accepted deviations by `docs/DECISIONS.md`.

## Current design spine

1. [Transport analysis](../../docs/research/TRANSPORT_ANALYSIS.md) — why MCP carries control while sinks carry evidence.
2. [Technical design](../../docs/design/TECHNICAL_DESIGN.md) — the original system model; read with [decisions](../../docs/DECISIONS.md) for accepted deviations.
3. [MCP surface](../../docs/design/MCP_SURFACE.md) — detailed director/operator contract; implemented tool names are summarized in [the MCP reference](../reference/mcp.md).
4. [Consumer contract](../../docs/design/CONSUMERS.md) — the reference consumer and integration boundary.
5. [Ground truth and scoring](../../docs/design/GROUND_TRUTH_AND_SCORING.md) — oracle, ledger, action, and benchmark rules.
6. [Domain catalog](../../docs/design/DOMAIN_CATALOG.md) — specified coverage and build order; specified is not shipped.

## Status discipline

The archive contains documents written at different stages. In particular, older technical-design text may describe SQLite, Go versions, or commands that do not match the current implementation. The accepted implementation uses file-based JSON artifacts; verify current behavior against [`internal/`](../../internal/) and [the limitations page](../limitations.md).

## Historical review material

The implementation plan, gap analysis, critical review, readiness review, and MCP interaction review remain useful for provenance and adversarial context. They are not current product reference pages. Use them to understand why a guarantee exists, then verify the current code and tests.

## Next reads

- [Architecture overview](../architecture/overview.md)
- [Nine non-negotiables](../architecture/invariants.md)
- [Working archive](../../docs/README.md)
