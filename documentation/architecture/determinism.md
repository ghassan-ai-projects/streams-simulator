# Determinism and replay

> Status: Implemented for the deterministic local path. Authority: `internal/run` and the run-artifact schema. Verified by: quickstart replay and CLI subprocess tests. Last verified: 2026-08-17.

The determinism boundary is the set of inputs that must be stable for a run to reproduce:

```text
(simulator version, domain digest, adapter digest, seed, command log, sink)
```

The world also records its start time, entity selection, scenario profile, time mode, and other configuration in the run artifact. A changed input is a different experiment, not a replay failure to suppress.

## How a run becomes reproducible

1. Load and digest the domain and adapter.
2. Create a world with an explicit seed and simulation start time.
3. Record every world-mutating command in order.
4. Advance simulation time rather than relying on wall time.
5. Render the same delivered event sequence through the same adapter and sink mode.
6. End the run and write the artifact, trace digest, counts, and ledger evidence.
7. Replay the artifact’s command log against a fresh world and compare the trace digest.

The replay library can use embedded domain and adapter specifications when available. The current `streamsim replay` and `streamsim verify` commands still load the matching files from `domains/` and `adapters/`, so those directories must be present unless the caller uses the library directly. The commands verify identity/digest matches against those current files. Replay does not re-run an MCP conversation or trust a human to reproduce calls in the same order.

## What breaks reproducibility

- `wall` time mode makes delivery depend on wall-clock behavior.
- An external HTTP receiver can introduce timing, availability, and side effects outside the simulator.
- A changed domain or adapter file changes its digest and invalidates the original input set.
- A changed simulator version may legitimately produce a different trace.
- A command omitted from the command log is an unreproducible mutation.
- A run that times out at the quiescence barrier is incomplete; it is not a clean successful run.

## Verification output

`streamsim verify <run.json>` reports whether the replay matches, whether the simulator version matches, the expected and observed digests, and—when available—the first divergence index and detail. Preserve that output with the run artifact when diagnosing a failure.

## Test evidence

The implementation tests determinism and replay in [`internal/run/`](../../internal/run/), [`internal/world/`](../../internal/world/), and the CLI subprocess tests in [`internal/cli/subprocess_test.go`](../../internal/cli/subprocess_test.go). The underlying artifact contract is [`docs/contracts/run-artifact-v0.1.schema.json`](../../docs/contracts/run-artifact-v0.1.schema.json).

## Next reads

- [Artifact reference](../reference/artifacts.md)
- [Replay and debugging](../guides/replay-and-debugging.md)
- [Compatibility](../overview/compatibility.md)
