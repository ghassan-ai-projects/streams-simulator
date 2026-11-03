# Compatibility and support

> Status: Current compatibility boundary; no formal OS support matrix is promised. Authority: `go.mod`, CI, and runtime packages. Verified by: toolchain and workflow review. Last verified: 2026-11-03.

This page records the compatibility boundary of the current repository. It is intentionally conservative: a toolchain version in `go.mod` is a build requirement, not a promise that every future version is supported.

## Current build surface

| Area | Current requirement or behavior |
| --- | --- |
| Language | Go `1.25.12`, as declared by [`go.mod`](../../go.mod). |
| Binary | `cmd/streamsim`; build with `make build` or run with `go run ./cmd/streamsim`. CI builds with Go `1.26.10` on Linux and cross-compiles `linux/amd64`. |
| Canvas | TypeScript app in [`canvas/`](../../canvas/) (Node `>=24`, npm); not embedded in the binary. Build it, then serve it with `streamsim canvas --ui-dir canvas/dist`. Desktop browsers, 1280 px or wider. |
| Runtime dependencies | Standard library plus the official MCP Go SDK and declared Go modules in [`go.mod`](../../go.mod). |
| Data format | JSON specifications (domains, adapters, component library files, compositions, kind files, scenarios, campaigns) validated against the schemas in [`contracts/`](../../contracts/); JSONL traces, ledgers and journals for the shipped JSONL adapters and run folders. |
| Plant wire | Contract `v1.5.0` in [`contracts/plant-wire/`](../../contracts/plant-wire/SPEC.md); a gateway must match its version and digest exactly. |
| Sinks | `inproc`, `file`, and `http-push`; HTTP push is an integration boundary, not a deterministic local sink. |
| MCP transports | Director over stdio; operator endpoint over streamable HTTP, created by a director process. |
| Canonicalization | RFC 8785-style canonical JSON is used where digests are computed. |

## Platform expectations

The project is developed and tested as a Go command-line application, with a browser canvas served by the same binary. The local build and test path should work on supported Go platforms with filesystem access; the owner-only unix sockets used by `device serve` and `plantwire` need a unix-like system. HTTP-push scenarios additionally require a reachable receiver and are not suitable for offline unit tests.

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
