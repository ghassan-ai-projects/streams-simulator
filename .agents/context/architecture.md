# Architecture Context

## Repository Architecture

The accepted architecture is summarized in [`documentation/architecture/`](../../documentation/architecture/), while the detailed design record remains in [docs/design/](../../docs/design/). Read both when a change affects a protocol, contract, trust boundary, or replay guarantee.

## Current Structure

- `AGENTS.md` is the canonical agent entrypoint.
- `README.md` introduces the project.
- `Makefile` defines the authoritative local commands.
- `.github/workflows/ci.yml` defines CI parity for core checks.
- `.golangci.yml` and `.pre-commit-config.yaml` enforce code quality and hygiene.
- `documentation/` holds curated public usage, architecture, benchmark, operations, and governance docs.
- `docs/` holds the engineering archive, canonical contract sources, design history, research, adapters/examples fixtures, and reviews.

## Current Package Ownership

- `cmd/streamsim/main.go`: version metadata and CLI entrypoint
- `internal/cli`: flags, application wiring, commands and shutdown
- `internal/world`: seeded discrete-event world core; domains are data loaded through one schema; facade over `internal/world/internal/domain`
- `internal/perturb`: perturbation layer between world and adapter (what the observer got, not what happened); facade over `internal/domain`
- `internal/adapter`: declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats; facade over `internal/adapter/internal/domain` with a `files` edge
- `internal/adapter/conformance`: `adapter verify` — renders the fixture, validates the declared output schema and byte-compares the golden
- `internal/sink`: inproc, file, http-push; facade over `internal/sink/internal/domain` with `files` and `httppush` edge packages
- `internal/mcp`: facade over `internal/mcp/internal/app`; one server, two roles — `director` (catalog · world · clock · fault · perturb · truth) and `operator` (nameplate · effectors · invoke · verdict)
- `internal/run`: run orchestration, delivery ledger, artifacts, replay, and quiescence; facade over `internal/run/internal/app` (orchestration), `domain` (pure rules) and the `durable`, `quiesce` and `clock` edges
- `internal/truth`: ground-truth generation and independent analytic solver; facade over `internal/truth/internal/domain`
- `internal/score`: instrument and consumer scoring, shared online/offline policies; facade over `internal/score/internal/domain`; the host packs a `score.Evidence`
- `internal/audit`: trivial-baseline evaluation over delivered evidence; facade over `internal/audit/internal/domain`
- `internal/suite`: scenario generation, composition and admission; facade over `internal/suite/internal/domain`
- `internal/refconsumer`: shipped reference detector and operator-client integration; facade over `internal/refconsumer/internal/domain` and the `mcpclient` edge
- `internal/device`: data-defined capability admission and vendored device transport; facade over `internal/device/internal/domain` (rules, codec) and the `uds` transport edge
- `internal/device/contract`: the vendored device wire-protocol schemas (embedded) and conformance fixtures
- `internal/deviceworld`: device-to-world effector binding; facade over `internal/deviceworld/internal/domain`
- `internal/domain`: load, validate, compile and digest domain data; facade over `internal/domain/internal/domain` with a `files` edge
- `internal/model`: shared simulator records and strict JSON decoding
- `internal/jsonschema`: core schema compilation and validation
- `internal/schemas`: embedded contract sources
- `internal/canonical`: canonical JSON encoding and digest operations
- `internal/randutil`: seeded random streams
- `internal/wall`: wall-clock seam and deterministic build replacement
- `test/`: integration and end-to-end suites

## Deliberate Placements

- **Perturbation sits between world and adapter** — the world produces what physically happened, the perturbation layer what the observer got. A scenario the consumer never saw is scored as a transport miss, not a reasoning miss.
- **The adapter sits last** — every consumer sees the same perturbation, so cross-consumer comparison means something.
- **MCP manages the simulator, it does not transmit the streams.** Evidence leaves on a sink; MCP carries control and actuation.

## Dependency Direction

- `cmd` -> `cli` -> `mcp`/`run`/`suite`/`score`/`refconsumer`
- `run` -> `world`/`perturb`/`adapter`/`sink`/`domain`/`model`/`canonical`
- `truth` -> `world`/`domain`/`model`; `score` -> `domain`/`world`/`model` (hosts pack a `score.Evidence`)
- `deviceworld` -> `device`/`world`/`model`; neither core imports the bridge
- Dependencies flow downward only.
- `test/architecture/dependencies_test.go` is the complete direct-import allowlist; this list is an ownership summary.
- Domain specs, adapters, and effectors are data (JSON), never code.
- The binary contains no consumer knowledge and no effector names.

Avoid:

- circular imports
- business logic in MCP handlers
- persistence concerns leaking into `world`
- transport concerns leaking into `model`
- per-domain code branches

## Modularity standard

Package kinds, layer shapes, purity/I-O rules (M1-M11) and the migration plan
are in [docs/refactoring/modularity-20261010/](../../docs/refactoring/modularity-20261010/README.md).
Where this file and `STANDARD.md` disagree about layering, `STANDARD.md` is
current and this file is updated in the same round that changes ownership.
