# Project Context

## What This Repo Is

Streams Simulator (`streamsim`) is a deterministic, closed-loop world simulator for testing stream processors. It is an independent product with no knowledge of its consumers: a consumer is a file (an output adapter) plus, optionally, an MCP client that invokes effectors and submits a verdict.

It is a test instrument first. Determinism, sealed ground truth, and analytic oracles constrain everything.

## Current State

- The implementation is present under `cmd/streamsim` and `internal/`; the original S0–S6 plan is historical execution context.
- The engineering archive is under [`docs/`](../../docs/README.md); the curated public documentation is under [`documentation/`](../../documentation/README.md).
- Module path `github.com/ghassan-ai-projects/streams-simulator` is set.
- Eight domain files and two adapters are committed; the schema suite validates every shipped domain, so new inputs must land together with test, inventory, and manifest updates.
- The root [doc.go](../../doc.go) package remains so Go tooling has a stable module root.
- Current implementation status and known limitations are in [`documentation/limitations.md`](../../documentation/limitations.md).

## What Agents Should Optimize For

- Build the documented design; do not invent architecture outside it.
- Keep the determinism boundary airtight. A run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`.
- Treat domains, adapters, and effectors as data. No per-domain code.
- Preserve parity between public documentation, `Makefile`, CI, contracts, and the implementation.
- Prefer durable repo files over long always-loaded guidance.

## Main Risks

- Adding consumer knowledge (names, schemas, behaviours) to the core — forbidden.
- Weakening determinism (timing, map iteration, non-canonical JSON, unseeded randomness).
- Building a broker/physics engine/plugin system the cut list removed.
- Drifting from the documented contracts (`docs/contracts/*.schema.json`).
- Documenting commands that do not match actual Makefile/CI behavior.
