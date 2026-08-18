# Authoring a domain

> Status: Implemented data-authoring workflow. Authority: domain schema, loader, and validation tests. Verified by: domain validation smoke test. Last verified: 2026-08-17.

A domain is data. The extension workflow should not require a new Go branch.

## Start from the contract

Read [`docs/contracts/domain-spec-v0.1.schema.json`](../../docs/contracts/domain-spec-v0.1.schema.json) and a committed file under [`domains/`](../../domains/). Identify the axes and runtime behavior the domain is meant to stress before adding fields.

The design catalog is a planning document, not proof that a domain is shipped. A domain becomes shipped only when it is intentionally committed, loaded by the catalog, covered by schema/conformance tests, and included in the release evidence.

## Validate locally

```bash
go run ./cmd/streamsim domain validate domains/my-domain.domain.json
go run ./cmd/streamsim catalog describe my-domain
go test ./internal/domain ./internal/schemas
```

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
