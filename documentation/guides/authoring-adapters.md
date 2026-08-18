# Authoring an adapter

> Status: Implemented declarative adapter workflow. Authority: adapter schema, loader, and golden tests. Verified by: adapter verification smoke test. Last verified: 2026-08-17.

An adapter is a declarative projection from native `sim-event-v0.1` to a consumer wire format.

## Workflow

1. Read [`docs/contracts/output-adapter-v0.1.schema.json`](../../docs/contracts/output-adapter-v0.1.schema.json).
2. Copy the shape of a committed adapter under [`adapters/`](../../adapters/).
3. Define the output encoding, metadata, field mappings, and any declared omission/mangling behavior in JSON.
4. Add the consumer-facing schema or vendor fixture in the adapter’s own data area.
5. Add a golden input/output fixture.
6. Run adapter verification and conformance tests.

```bash
go run ./cmd/streamsim adapter list
go run ./cmd/streamsim adapter verify adapters/my-adapter.adapter.json
go test ./internal/adapter
```

## Rules

- The adapter must not alter world truth or command semantics.
- It must be deterministic for the same native event and configuration.
- It must fail clearly on unsupported required fields; do not silently invent defaults that change meaning.
- Consumer names and semantics belong in adapter data or the external consumer, not in the simulator core.
- A golden mismatch is a contract change requiring review, not a fixture update to make a test green.

## Source authority

The authoritative public adapter contracts are in [`docs/contracts/`](../../docs/contracts/). The runtime loader and golden/conformance tests are under [`internal/adapter/`](../../internal/adapter/).

## Next reads

- [Domains and adapters](../architecture/domains-and-adapters.md)
- [Contracts](../reference/contracts.md)
- [Consumer integration](consumer-integration.md)
