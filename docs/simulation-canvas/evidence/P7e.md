# P7e evidence — pre-UI review and architecture cleanup

- **Date**: 2026-10-13
- **Scope**: architecture and modularity (owner decision); behaviour findings are registered for P7f
  ([reviews/P7e.md](../reviews/P7e.md)).

## Changes

| Change | Effect |
|---|---|
| `internal/kernel` (kind "os": facade over `files`, `sockets`, `process`) | One place touches the operating system; 16 per-module `os` call sites, three durable-append implementations, two socket stacks and two signal handlers are gone; written files are owner-only and atomic |
| `internal/jsonvalue` (foundation) | One number reader and one strict decoder instead of five and two copies |
| `model` names the run-folder files and run modes | Eight packages stop spelling `run.json`, `ledger.jsonl`, … themselves |
| `TestDomainLayersImportOnlyPureModules` | A domain layer imports only foundations, pure cores, the domain-spec module and reviewed value-type imports |
| `make deadcode` (`scripts/check-deadcode`, `scripts/test-hooks.txt`) | Fails on dead code and on production code that only tests reach, except reviewed test hooks; the old target printed findings without failing |
| Test-only production surface deleted | `run.ScenarioConfig`/`CheckScenario`/`ScenarioFromArtifact`, `suite.Generate`, `score.Offline`, `world.KnownFailureMode`, `adapter.FixtureEvents`, `plantwire.ContractDigest`, `kernel.Owned`, `mcp.NewDirectorServer`, `schemas.SimEvent`, `device.Listen`/`SetFaults`/`AcceptedCommandCount`, eleven director code aliases |
| Legacy device capability form removed | 192 lines |
| Test helpers shared (`testsupport.ArgsForSchema`), CLI input flags registered once | Fewer copies |
| Production comments removed (2,568 lines) | Package docs and gate markers only (owner rule; STANDARD updated) |
| Architecture documentation | `documentation/architecture/overview.md`, `contracts.md`, STANDARD §1, AGENTS.md, the package map |
| Two truth fixes taken before the scope decision | Operator calls act at the world clock; run and world ids carry no seed (pin re-recorded for id text only) |

## Gates

- `make phase-gate` passes; behaviour pin unchanged by the structural commits.
- `scripts/check-deadcode` passes with 30 declared test hooks.
