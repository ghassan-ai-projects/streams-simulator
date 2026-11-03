# Authoring an adapter

> Status: Implemented declarative adapter workflow. Authority: adapter schema, loader, and golden tests. Verified by: adapter verification smoke test. Last verified: 2026-11-03.

An adapter is a declarative projection from native `sim-event-v0.1` to a consumer wire format. It is JSON data: a new consumer needs an adapter file, never a branch in the binary.

## Workflow

1. Read [`contracts/output-adapter-v0.1.schema.json`](../../contracts/output-adapter-v0.1.schema.json) and copy the shape of a committed adapter under [`adapters/`](../../adapters/) (`native-jsonl` is the smallest, `agentic-stream` shows a preamble, an id rewrite and nested objects).
2. Name it `adapters/<id>.adapter.json`. The `id` matches `^[a-z][a-z0-9-]{2,63}$`; `encoding`, `version` and a `record` template are required. Add a `preamble`, `postamble` or `entity_id_rewrite` only when the consumer needs them.
3. Map each output field with `from` operations (`const`, `source`, `counter`, `object` and the others in the schema). The `record` must carry a stable event identity derived from the native `seq`, because the duplicate probes and the verdict's `seq:N` citations depend on it.
4. Vendor the consumer's published record schema under `adapters/vendor/<consumer>/` and name it in `conformance.output_schema`.
5. Add a golden file under `adapters/golden/` and name it in `conformance.golden`. The golden is the expected rendering of the simulator's fixed 12-event conformance fixture; the fixture is not yours to change. Review the rendered bytes by hand before committing them.
6. Verify, then run a domain through the adapter.

```bash
go run ./cmd/streamsim adapter list
go run ./cmd/streamsim adapter verify adapters/my-adapter.adapter.json
go run ./cmd/streamsim run --domain rotating-machinery --adapter my-adapter --out /tmp/my-adapter-run
go test ./internal/adapter/...
```

`adapter verify` prints `{adapter, schema_checked, schema_ok, golden_checked, golden_match, records}`. Conformance paths resolve against `--adapters-dir` (default `adapters`), not the adapter file's folder; point `--adapters-dir` at your directory when it is elsewhere. It fails when a declared check fails, when a declared file is missing, and when the adapter declares neither check.

## Rules

- The adapter must not alter world truth or command semantics.
- It must be deterministic for the same native event and configuration.
- It must fail clearly on unsupported required fields; do not silently invent defaults that change meaning.
- Consumer names and semantics belong in adapter data or the external consumer, not in the simulator core.
- A golden mismatch is a contract change requiring review, not a fixture update to make a test green.
- The adapter's digest is a run input: changing it makes a run a different run (see [replay and debugging](replay-and-debugging.md)).

## Source authority

The authoritative public adapter contracts are in [`contracts/`](../../contracts/). The runtime loader and golden/conformance tests are under [`internal/adapter/`](../../internal/adapter/).

## Next reads

- [Domains and adapters](../architecture/domains-and-adapters.md)
- [Contracts](../reference/contracts.md)
- [Consumer integration](consumer-integration.md)
