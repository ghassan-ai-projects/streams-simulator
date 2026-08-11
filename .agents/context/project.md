# Project Context

## What This Repo Is

Streams Simulator (`streamsim`) is a deterministic, closed-loop world simulator for testing stream processors. It is an independent product with no knowledge of its consumers: a consumer is a file (an output adapter) plus, optionally, an MCP client that invokes effectors and submits a verdict.

It is a test instrument first. Determinism, sealed ground truth, and analytic oracles constrain everything.

## Current State

- Design is complete and committed under `docs/` — research, design, contracts, adapters, examples. Status: design, ready to build.
- Module path `github.com/ghassan-ai-projects/streams-simulator` is set.
- There is no `cmd/` directory yet.
- There are no `internal/*` packages yet.
- The root [doc.go](../../doc.go) package exists so Go tooling has something to operate on.
- Implementation begins at stage S0 of [docs/design/IMPLEMENTATION_PLAN.md](../../docs/design/IMPLEMENTATION_PLAN.md) (three decisions, ~half a day), then S1-S2 (sixteen days) to close the loop against the reference consumer.

## What Agents Should Optimize For

- Build the documented design; do not invent architecture outside it.
- Keep the determinism boundary airtight. A run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`.
- Treat domains, adapters, and effectors as data. No per-domain code.
- Preserve parity between prose, `Makefile`, CI, and the design docs.
- Prefer durable repo files over long always-loaded guidance.

## Main Risks

- Adding consumer knowledge (names, schemas, behaviours) to the core — forbidden.
- Weakening determinism (timing, map iteration, non-canonical JSON, unseeded randomness).
- Building a broker/physics engine/plugin system the cut list removed.
- Drifting from the documented contracts (`docs/contracts/*.schema.json`).
- Documenting commands that do not match actual Makefile/CI behavior.
