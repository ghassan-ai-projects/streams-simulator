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

## Target Structure (per the design)

- `cmd/streamsim/main.go`: entrypoint, flags, wiring, shutdown
- `internal/world`: seeded discrete-event world core; domains are data loaded through one schema
- `internal/perturb`: perturbation layer between world and adapter (what the observer got, not what happened)
- `internal/adapter`: declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats
- `internal/sink`: inproc, file, http-push
- `internal/mcp`: one server, two roles — `director` (catalog · world · clock · fault · perturb · truth) and `operator` (nameplate · effectors · invoke · verdict)
- `internal/run`: run orchestration, delivery ledger, artifacts, replay, and quiescence
- `internal/truth`: sealed ground truth and scoring
- `test/`: integration and end-to-end suites

## Deliberate Placements

- **Perturbation sits between world and adapter** — the world produces what physically happened, the perturbation layer what the observer got. A scenario the consumer never saw is scored as a transport miss, not a reasoning miss.
- **The adapter sits last** — every consumer sees the same perturbation, so cross-consumer comparison means something.
- **MCP manages the simulator, it does not transmit the streams.** Evidence leaves on a sink; MCP carries control and actuation.

## Dependency Direction

- `cmd` -> `cli` -> `mcp`/`run`/`suite`/`score`/`refconsumer`
- `run` -> `world`/`perturb`/`adapter`/`sink`/`truth`
- Dependencies flow downward only.
- Domain specs, adapters, and effectors are data (JSON), never code.
- The binary contains no consumer knowledge and no effector names.

Avoid:

- circular imports
- business logic in MCP handlers
- persistence concerns leaking into `world`
- transport concerns leaking into `models`
- per-domain code branches
