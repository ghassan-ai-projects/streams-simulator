# Authoring a domain

> Status: Implemented data-authoring workflow. Authority: domain schema, loader, and validation tests. Verified by: domain validation smoke test. Last verified: 2026-08-17.

A domain is data. The extension workflow should not require a new Go branch.

## Start from the contract

Read [`docs/contracts/domain-spec-v0.2.schema.json`](../../docs/contracts/domain-spec-v0.2.schema.json), the [kinds and instances semantics](../../docs/design/DOMAIN_SPEC_V0_2.md) and a committed file under [`domains/`](../../domains/). Identify the axes and runtime behavior the domain is meant to stress before adding fields.

A domain declares one **kind** (the template: state, dynamics, channels, faults, effectors) and one **instance set** (the id template, the count bounds and any churn) that follows it. The kind name becomes the `entity_type` of every instance. Explicit entity ids given to a run (a scenario's `world.entity_ids`, the director's world creation) must fit the instance set's id template and count bounds.

The design catalog is a planning document, not proof that a domain is shipped. A domain becomes shipped only when it is intentionally committed, loaded by the catalog, covered by schema/conformance tests, and included in the release evidence.

## Validate locally

```bash
go run ./cmd/streamsim domain validate domains/my-domain.domain.json
go run ./cmd/streamsim domain format domains/my-domain.domain.json
go run ./cmd/streamsim catalog describe my-domain
go test ./internal/domain/... ./internal/schemas
```

`domain format` rewrites the file in the canonical layout (stable member order, line-fitting objects), so a hand edit and a tool edit of the same world produce the same bytes; run it before committing. It never changes the digest. A validation failure reports every issue with a JSON pointer into the document.

Use the exact output digest in review notes. If the domain changes after a run, that run must be treated as a different input set.

## Design checks

- Every entity and channel has a reason to exist in the scenario.
- State dynamics have a boundary case and an analytic or metamorphic test where applicable.
- Faults and effectors are declared, parameterized, and scoped to valid entities.
- Absence semantics are explicit: signal, normal, ambiguous, or heartbeat.
- A scenario is not trivially solvable by a one-line detector.
- Free text cannot become a hidden control channel.
- The domain does not encode knowledge of a particular consumer.

## Completion evidence

Add or update fixtures and tests under the existing package boundaries. Update the catalog/coverage evidence, release inventory, and public limitations if the domain is partial. Do not call a new domain shipped solely because the CLI can load it from a dirty working tree.

## Next reads

- [Domains and adapters](../architecture/domains-and-adapters.md)
- [Benchmark methodology](../benchmark/methodology.md)
- [Contracts](../reference/contracts.md)
