# Compatibility and support

> Status: Current compatibility boundary; no formal OS support matrix is promised. Authority: `go.mod`, CI, and runtime packages. Verified by: toolchain and workflow review. Last verified: 2026-08-17.

This page records the compatibility boundary of the current repository. It is intentionally conservative: a toolchain version in `go.mod` is a build requirement, not a promise that every future version is supported.

## Current build surface

| Area | Current requirement or behavior |
| --- | --- |
| Language | Go `1.25.12`, as declared by [`go.mod`](../../go.mod). |
| Binary | `cmd/streamsim`; build with `make build` or run with `go run ./cmd/streamsim`. |
| Runtime dependencies | Standard library plus the official MCP Go SDK and declared Go modules in [`go.mod`](../../go.mod). |
| Data format | JSON domain and adapter specifications; JSONL traces for the shipped JSONL adapters. |
| Sinks | `inproc`, `file`, and `http-push`; HTTP push is an integration boundary, not a deterministic local sink. |
| MCP transports | Director over stdio; operator endpoint over streamable HTTP, created by a director process. |
| Canonicalization | RFC 8785-style canonical JSON is used where digests are computed. |

## Platform expectations

The project is developed and tested as a Go command-line application. The local build and test path should work on supported Go platforms with filesystem access. HTTP-push scenarios additionally require a reachable receiver and are not suitable for offline unit tests.

The repository does not currently publish a formal OS matrix, compatibility promise, or signed release channel. Treat the current checkout as source-level software and validate the exact commit, toolchain, domain, adapter, and consumer inputs for any result.

## Compatibility rules

- Contract version identifiers such as `sim-event-v0.1` and schema versions are part of the interchange surface.
- A domain or adapter digest mismatch is a replay error, not a warning to ignore.
- A simulator version mismatch may explain a different trace; it must be visible in replay output and release evidence.
- Do not assume a file trace produced by one adapter can be consumed as another adapter’s format.
- Do not add consumer-specific fields to the native event model; add or change an adapter contract instead.

## Next reads

- [Install and build](../getting-started/install.md)
- [Contracts](../reference/contracts.md)
- [Release evidence](../benchmark/evidence.md)
