# Domains and adapters

> Status: Implemented extension model with a partial suite-harness limitation. Authority: domain/adapter contracts, loaders, and conformance tests. Verified by: catalog and adapter smoke tests. Last verified: 2026-08-17.

Streams Simulator extends through data files, not per-domain or per-consumer code.

## Domains

A domain specification declares the world as one or more **kinds** (the template: state variables, dynamics, channels, faults, effectors) and the **instance sets** that follow them (id template, count bounds, churn), plus profiles and property axes. Channel, fault and effector names are unique across the domain; every other reference resolves inside its kind. The validated document is the form the simulator core reads. The loader validates the JSON against the domain contract, computes a digest, and exposes it through the catalog.

The committed domain set is under [`domains/`](../../domains/): eight domains, each validated against the embedded schema by the shipped-domain test. The design catalog describes more domains than are currently committed. See [limitations](../limitations.md).

The world and domain loader are data-defined, but the suite harness is not fully generic yet: [`internal/suite/internal/domain/admission.go`](../../internal/suite/internal/domain/admission.go) still branches on the profile names `correlated_cascade` and `sensor_pathology`. Treat “data-defined” as a core runtime guarantee, not as a claim that every benchmark-generation convenience is domain-neutral today.

Use the common path to inspect and validate data:

```bash
go run ./cmd/streamsim catalog list
go run ./cmd/streamsim catalog describe rotating-machinery
go run ./cmd/streamsim domain validate domains/rotating-machinery.domain.json
go run ./cmd/streamsim domain format domains/rotating-machinery.domain.json
```

## Adapters

An adapter projects native `sim-event-v0.1` records into a consumer wire format. The adapter is data and has its own identity, version, schema, and golden fixture. Deleting an adapter must not change world behavior.

The shipped adapters are listed by:

```bash
go run ./cmd/streamsim adapter list
go run ./cmd/streamsim adapter verify adapters/native-jsonl.adapter.json
```

## Extension rules

- Add domain behavior to JSON before adding code.
- Keep consumer field names and interpretation in an adapter.
- Add a schema and golden fixture for a new adapter.
- Validate canonical digests and replay implications before changing a contract.
- Add conformance tests that prove the new data loads through the existing path.

## Source authority

The machine-readable domain and adapter contracts are in [`contracts/`](../../contracts/). The design catalog and coverage rationale are in [`docs/design/DOMAIN_CATALOG.md`](../../docs/design/DOMAIN_CATALOG.md). The loader tests are under [`internal/domain/`](../../internal/domain/) and [`internal/adapter/`](../../internal/adapter/); adapter conformance is `adapter.Verify` in [`internal/adapter/`](../../internal/adapter/), implemented in [`internal/adapter/internal/domain/verify.go`](../../internal/adapter/internal/domain/verify.go).

## Next reads

- [Authoring domains](../guides/authoring-domains.md)
- [Authoring adapters](../guides/authoring-adapters.md)
- [Contracts](../reference/contracts.md)
