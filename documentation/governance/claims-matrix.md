# Documentation claim-to-source matrix

Validation snapshot: verified 2026-09-13 against commit `523fb77`. The repository declares Go `1.25.12` as its module minimum; CI uses the patched Go `1.25.13` toolchain. This local review ran with a compatible toolchain. The shipped inventory is eight committed domains and two adapters.

This matrix is the handoff evidence for public claims. It prevents the curated documentation from becoming a second, drifting implementation record.

| Claim area | Public home | Authority | Verification/evidence | Status |
| --- | --- | --- | --- | --- |
| Product purpose and boundaries | [`overview/product.md`](../overview/product.md) | `README.md`, design decisions, implementation boundaries | Repository review and package map | Verified summary |
| Toolchain | [`overview/compatibility.md`](../overview/compatibility.md) | [`go.mod`](../../go.mod), CI workflow | `go version`, `.github/workflows/ci.yml` | Verified: module minimum Go 1.25.12; CI Go 1.25.13 |
| CLI commands and flags | [`reference/cli.md`](../reference/cli.md) | [`internal/cli/cli.go`](../../internal/cli/cli.go) | `go run ./cmd/streamsim help`; targeted command smoke tests | Verified |
| Domain inventory | [`architecture/domains-and-adapters.md`](../architecture/domains-and-adapters.md) | `domains/`, release manifest, shipped-domain tests | `catalog list`; `go test ./internal/schemas` | Verified: eight shipped domains, each validated by the schema test |
| Adapter inventory | [`architecture/domains-and-adapters.md`](../architecture/domains-and-adapters.md) | `adapters/`, adapter tests | `adapter list`; `adapter verify` | Verified |
| Event pipeline | [`architecture/pipeline.md`](../architecture/pipeline.md) | `internal/world`, `perturb`, `adapter`, `sink`, `run` | run ledger and golden tests | Verified summary |
| Determinism/replay | [`architecture/determinism.md`](../architecture/determinism.md) | `internal/run`, run-artifact schema | quickstart `verify`; replay tests | Verified for deterministic local path |
| MCP director surface | [`reference/mcp.md`](../reference/mcp.md) | `internal/mcp/server.go`, `schemas.go` | MCP strictness tests | Verified |
| MCP operator surface | [`reference/mcp.md`](../reference/mcp.md) | `internal/mcp/operator.go`, operator endpoint | operator E2E tests | Implemented; environment may block loopback tests |
| Contract schemas | [`reference/contracts.md`](../reference/contracts.md) | `docs/contracts/`; embedded copies in `internal/schemas/` | byte-equality and schema tests | Verified |
| Run output files | [`reference/artifacts.md`](../reference/artifacts.md) | `internal/run/run.go` | quickstart output inventory | Verified |
| Truth and scoring | [`architecture/truth-and-scoring.md`](../architecture/truth-and-scoring.md) | `internal/truth`, `internal/score`, verdict/truth schemas | truth and score tests | Substantial evidence; release status conditional |
| Nine non-negotiables | [`architecture/invariants.md`](../architecture/invariants.md) | design archive plus named tests | per-gate test files | Named evidence present; full suite green at snapshot commit |
| Release posture | [`operations/release.md`](../operations/release.md) | CI, manifest, exact commit | `make ci-check`, `make manifest`, replay | Gates green at snapshot commit; publication still requires fresh evidence for the exact commit |
| Open-source policies | root policy files and [governance](quality.md) | `CONTRIBUTING.md`, `SECURITY.md`, `LICENSE`, `CODE_OF_CONDUCT.md`, `SUPPORT.md`, `CHANGELOG.md` | file presence and review | Present |

## Current validation snapshot

- `go run ./cmd/streamsim adapter verify adapters/native-jsonl.adapter.json` — passed.
- `go run ./cmd/streamsim domain validate domains/rotating-machinery.domain.json` — passed.
- Documented deterministic run and `verify` — passed with matching digest.
- `make docs-check` — passed.
- Link-target scan over `README.md` and `documentation/**/*.md` — zero broken local targets.
- `go test ./...` — full suite passes, including the shipped-domain schema validation over all eight domains and the loopback-bound operator and HTTP-push tests. A restricted sandbox may block loopback listeners; see [limitations](../limitations.md).

## Maintenance rule

Update this matrix whenever a public claim changes, a command/contract changes, a domain or adapter is added, or release status changes. A claim without an authority and verification path is not ready for the public hub.

## Next reads

- [Quality bar](../QUALITY_BAR.md)
- [Limitations](../limitations.md)
- [Contributor quality](quality.md)
