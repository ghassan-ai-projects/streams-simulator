# Quickstart

> Status: Verified local deterministic path. Authority: current CLI and shipped domain/adapter files. Verified by: adapter validation, domain validation, run, and replay. Last verified: 2026-08-17.

This path exercises the local, deterministic workflow: inspect the catalog, validate an adapter, run a short scenario, and verify the resulting artifact.

## 1. Inspect installed data

```bash
go run ./cmd/streamsim catalog list
go run ./cmd/streamsim catalog coverage
go run ./cmd/streamsim adapter list
```

The commands print JSON so their output can be captured by scripts. Use `catalog describe` for one domain:

```bash
go run ./cmd/streamsim catalog describe rotating-machinery
```

## 2. Validate an adapter

The shipped native adapter emits newline-delimited JSON (JSONL) and has a golden fixture:

```bash
go run ./cmd/streamsim adapter verify adapters/native-jsonl.adapter.json
```

The command checks the adapter contract and compares its output with the committed golden data.

## 3. Run a short deterministic scenario

```bash
mkdir -p /tmp/streamsim-quickstart
go run ./cmd/streamsim run --domain rotating-machinery --adapter native-jsonl --seed 7 --sink file --out /tmp/streamsim-quickstart --duration 120
```

The command prints the run ID, trace digest, reproducibility flag, emitted count, and ledger count. The output directory contains the run artifact and evidence files produced by the run.

## 4. Verify the artifact

```bash
go run ./cmd/streamsim verify /tmp/streamsim-quickstart/run.json
```

Verification loads the artifact, loads the matching current files from `domains/` and `adapters/`, replays the command log against a fresh world, and compares the resulting trace digest. A mismatch is evidence to investigate, not a result to average away. The replay library can use embedded specifications, but the CLI currently requires those directories.

## 5. Explore the next layer

- Add a declared fault or perturbation with the syntax in [the CLI reference](../reference/cli.md).
- Run the [reference consumer](../guides/consumer-integration.md) against a native JSONL trace.
- Use [replay and debugging](../guides/replay-and-debugging.md) to inspect a failure.
- Use [MCP integration](../reference/mcp.md) when the consumer must control and actuate a live world.

## Next reads

- [CLI reference](../reference/cli.md)
- [Determinism and replay](../architecture/determinism.md)
- [Limitations](../limitations.md)
