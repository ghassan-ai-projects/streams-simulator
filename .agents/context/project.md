# Project Context

## What This Repo Is

Streams Simulator (`streamsim`) is a deterministic, closed-loop world simulator: worlds as data, a seeded clock, delivery faults, effectors that change the world, sealed ground truth and grading. Around that core sit a visual canvas (`canvas/`), physical-device sessions over the plant wire, run journals and reports, and generated suites and campaigns. It is an independent product with no knowledge of its consumers: a consumer is a file (an output adapter) plus, optionally, an MCP client that invokes effectors and submits a verdict.

Grading a consumer is one use of it, not its whole scope. Determinism, sealed ground truth, and analytic oracles constrain everything.

## Current State

- The implementation is present under `cmd/streamsim` and `internal/`; the original S0–S6 plan is historical execution context.
- The curated, current documentation is under [`documentation/`](../../documentation/README.md); [`docs/`](../../docs/README.md) is a dated working archive that may be out of date.
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
- Drifting from the documented contracts (`contracts/*.schema.json`).
- Documenting commands that do not match actual Makefile/CI behavior.
