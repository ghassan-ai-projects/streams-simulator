# AGENTS.md - Streams Simulator

Canonical instructions for coding agents in this repository. Read this file first, then load only the context files needed for the task.

## Purpose

Streams Simulator (`streamsim`) is a standalone, deterministic, closed-loop world simulator for testing stream processors. It generates realistic event streams from data-defined domains, corrupts their delivery in specified ways, carries sealed ground truth, accepts commands back through declared effectors so the world actually changes, and scores whatever consumed it.

It is a **test instrument**. An instrument less trustworthy than the system it measures is worse than no instrument, because it produces confident wrong answers. Determinism guarantees, generative ground truth, and analytic oracles come before features.

The design and evidence archive is maintained in [docs/](docs/README.md); the curated public documentation is in [documentation/](documentation/README.md). The nine non-negotiables in the spec (determinism, analytic cross-check, reference consumer, delivery ledger, quiescence barrier, sealed oracle, injection probe, `silent_no_effect` test, trivial-baseline audit) are not negotiable.

## Engineering Priorities

Use these priorities in order:

1. Correctness and safety.
2. Simplicity of design and implementation.
3. Test evidence for changed behavior.
4. Consistency with existing repo rules and automation.
5. Speed of implementation.

If two options both work, choose the one that is easier to read, easier to test, and easier for the next agent to extend.

## Read Order

Before editing:

1. Read this file.
2. Read [README.md](README.md).
3. Read the relevant design docs in `docs/design/` and `docs/research/` before touching architecture or protocol surfaces.
4. Check the worktree with `git status --short`.
5. Read the smallest relevant context files under `.agents/context/`.
6. Make a short plan before editing.

Start with these context files:

- [.agents/context/project.md](.agents/context/project.md) for repository scope and current state.
- [.agents/context/architecture.md](.agents/context/architecture.md) for layout and dependency direction.
- [.agents/context/testing.md](.agents/context/testing.md) for commands and testing bar.
- [.agents/context/go-style.md](.agents/context/go-style.md) for coding conventions.
- [.agents/context/review-checklist.md](.agents/context/review-checklist.md) before handoff.
- [docs/refactoring/modularity-20261010/STANDARD.md](docs/refactoring/modularity-20261010/STANDARD.md) for package kinds, layer rules and the M1-M11 modularity bar; its [PLAN.md](docs/refactoring/modularity-20261010/PLAN.md) tracks the migration rounds and [DEFERRED.md](docs/refactoring/modularity-20261010/DEFERRED.md) lists known defects that are intentionally not fixed by structural rounds.

Use the prompt files under `.agents/prompts/` when the task matches them.

## Current Repository State

- Module path: `github.com/ghassan-ai-projects/streams-simulator` (set).
- Design, research, contracts, and implementation evidence are maintained under `docs/`; public usage documentation is under `documentation/`.
- The implementation includes `cmd/streamsim` and the `internal/` packages described below.
- Eight domains are committed in `domains/` and two adapters are committed in `adapters/`; the schema suite validates every shipped domain, so new inputs must be reconciled with the tests and release manifest in the same change.
- The root package in [doc.go](doc.go) remains so Go tooling has a stable module root.
- Historical stage planning is in [docs/design/IMPLEMENTATION_PLAN.md](docs/design/IMPLEMENTATION_PLAN.md); current status and limitations are in [documentation/limitations.md](documentation/limitations.md).

Do not invent architecture outside the documented design. The spec was written to be built as specified; deviations need a design change first.

## Architecture Overview

The documented shape (see [docs/design/TECHNICAL_DESIGN.md](docs/design/TECHNICAL_DESIGN.md)):

- `cmd/streamsim/main.go` - version metadata and CLI entrypoint
- `internal/cli` - flags, application wiring, commands, shutdown
- `internal/world` - seeded discrete-event world core; domain specs are data, loaded through one schema
- `internal/perturb` - perturbation layer between world and adapter (what the observer got, not what happened)
- `internal/adapter` - declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats
- `internal/sink` - inproc, file, http-push sinks
- `internal/mcp` - one MCP server, two roles: `director` (catalog, world, clock, fault, perturb, truth) and `operator` (nameplate, effectors, invoke, verdict)
- `internal/run` - run orchestration, delivery ledger, artifacts, replay, and quiescence; there is no separate `internal/ledger` package
- `internal/truth` - sealed ground truth and independent analytic solver
- `internal/score` - online/offline scoring policies
- `internal/audit`, `internal/suite`, `internal/refconsumer` - trivial-baseline audit, scenario generation, shipped reference consumer
- `internal/device`, `internal/deviceworld` - device emulation and its world integration bridge
- `internal/domain`, `internal/model`, `internal/jsonschema`, `internal/schemas`, `internal/canonical`, `internal/randutil`, `internal/wall` - data loading, shared records and bounded foundations
- `docs/` - the engineering design, research, contracts, fixtures, and evidence archive
- `documentation/` - curated public product, usage, architecture, benchmark, operations, and governance documentation

Dependency direction:

- `cmd` -> `cli` -> application packages (`mcp`, `run`, `suite`, `score`, `refconsumer`)
- `run` composes `world`/`perturb`/`adapter`/`sink`; truth and scoring consume world/evidence. `deviceworld` bridges `device` and `world` without either core importing the bridge.
- Dependencies flow downward only.
- Domain specs, adapters, and effectors are **data** (JSON), never code. If any domain needs a code branch in the binary, the simulator is wrong and the domain found the bug.
- The simulator has no knowledge of its consumers. No consumer name, schema, field, or behaviour appears in the binary.

Determinism: a run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`. One artifact reproduces any failure.

## Build, Test, and Lint

Primary commands:

- `make ci-check`
- `make function-length`
- `make build`
- `go vet ./...`
- `go test ./...`
- `make lint`
- `git diff --check`
- `pre-commit run --all-files` when `pre-commit` is installed

Important behavior:

- `make build` builds `cmd/streamsim` into `bin/`.
- `make test` and `go test ./...` exercise the implemented packages; inspect current limitations before treating a working-tree result as release evidence.
- `make lint` depends on `golangci-lint` and may fail if the environment cannot write to its cache.

See [.agents/context/testing.md](.agents/context/testing.md) for the testing and validation bar. The analytic cross-check (implement the integrator twice, assert agreement) is a non-negotiable correctness oracle, not a consistency check.

## Modularity bar

Packages are one of four kinds (foundation, pure core, core with an I/O edge,
surface), documented in [STANDARD.md](docs/refactoring/modularity-20261010/STANDARD.md).
Pure packages and every `internal/domain` layer import no `os`/`net`/`exec` and
read no wall clock; file, socket and clock access sits only in declared edge
packages; `cli` and `mcp` hold wiring and protocol only. A structural round
changes no behaviour: run `scripts/behaviour-pin` and diff it against
`docs/refactoring/modularity-20261010/BEHAVIOUR_PIN.txt`. Defects found during
a round are added to `DEFERRED.md`, not fixed in the same commit.

## Go Standards

### Clean-code and architecture bar

- Every Go source file, including tests, must be at most **300 total lines** (comments and blank lines count). Split files by a named responsibility within their owning package; do not split functions arbitrarily or compress code to meet the limit.
- Function names state domain intent. Each function performs one task at one abstraction level. Entry points read as a sequence of simulator operations; put parsing, serialization, record bookkeeping, and concrete mechanics in named steps below their callers.
- Every production Go function, method and anonymous function must have at most **15 physical body lines**, from the opening brace through the closing brace, including comments and blank lines. Test functions are excluded from this limit. Keep private helpers in stepdown reading order. Extract named responsibilities; do not compress statements or remove useful comments to meet the limit.
- Apply function-length refactoring to production source only; leave test files unchanged and use the existing tests to validate behavior.
- Packages own simulator responsibilities, not generic controller/service/store layers. Preserve the world → perturbation → adapter → sink pipeline and director/operator truth boundary. Create a package only for a distinct responsibility with a concrete caller and a downward dependency direction.
- Preserve exported signatures, JSON shapes, errors, command/delivery order, RNG draws, digest inputs, locks, cancellation, and effects during refactoring. Record intentional corrections separately and prove them with regression tests.
- Add meaningful boundary tests in each modified production package. Run focused tests and review the diff before each round's commit; run the full repository gate before handoff.
- The executable file-size, 15-line function, package-dependency, package kind/layer, I/O-edge inventory and package-documentation gates live in `test/architecture`. The strict 15-line AST check runs through `make function-length`, `make ci-check` and the local pre-commit hook. The review criteria and round evidence are in [docs/refactoring/clean-code-20261002/](docs/refactoring/clean-code-20261002/BAR.md).

- Use `context.Context` as the first parameter for cancellable or I/O work.
- Use `log/slog` for logging.
- Wrap errors with `%w`.
- Keep handlers thin; business logic lives in `world`/`perturb`/`adapter`/`truth`, not in MCP handlers.
- Prefer standard library helpers such as `cmp`, `maps`, and `slices`.
- Prefer existing package boundaries and local helpers over new abstractions.
- Document exported symbols.
- Use table-driven tests with `t.Run()` and `t.Parallel()` where safe.
- Use `t.Context()` in tests when appropriate.
- RFC 8785 canonical JSON everywhere a digest is computed.

See [.agents/context/go-style.md](.agents/context/go-style.md) for the repo-specific style rules.

## Enola architecture review

- Use Enola for this repository, with [mcp-arch.yaml](mcp-arch.yaml). Generated snapshots and baselines live in ignored `.enola/`; commit review evidence, not generated state.
- Before structural edits, generate a fresh snapshot for this repository and pin it with `set_baseline`, or run `make architecture-baseline`. Pin once before the round; do not re-pin to hide findings.
- Use `query_insights` for cycles/layers and `impact_analysis` before changing a shared symbol. After editing, regenerate, verify receipt comparability, and inspect `diff_snapshot` for new coupling, findings and scope spillover.
- Run `make architecture` to enforce new cycle/layer findings at confidence 0.8, matching the Tamoz workflow. A report-only exit zero is not an enforced pass. Missing baselines, unavailable tools and incomparable snapshots are blockers for this check, not clean results.
- Confirm heuristic hotspots, complexity and performance candidates against code and tests. Go import checks and `test/architecture` remain authoritative for compilation and the detailed business package graph; Enola's inferred command/internal layers do not cover every internal ownership boundary.
- After an Enola extractor/configuration upgrade, check coverage and comparability before pinning a replacement baseline. These local targets complement `make ci-check`; they require a pre-change local baseline.

## Forbidden Changes

- Do not add secrets, credentials, or machine-specific private data.
- Do not add network calls to unit tests.
- Do not embed consumer knowledge (names, schemas, fields, behaviours) in the binary.
- Do not add a broker, physics engine, ORM, expression language, or plugin system. The cut list in the implementation plan may remove anything else; it may never remove the nine non-negotiables.
- Do not introduce per-domain code branches; domains are data.
- Do not add top-level dependencies without clear justification. The core has a zero-external-dependency requirement (the MCP Go SDK is the documented exception).
- Do not add abstraction layers "for future flexibility" without a current concrete need.
- Do not create duplicate canonical agent files such as `CODEX.md`.
- Do not broaden repository instructions with vague policy that cannot be enforced or reviewed.
- Do not do unrelated refactors while touching implementation files.

## Definition Of Done

A task is done when:

- the requested scope is complete
- the change is specific to this repository and useful to future agents
- the change is the simplest correct one that fits the documented design
- production-code changes include meaningful tests, and modified packages do not show 0% coverage
- behavior changes are proven against the design's oracles (analytic cross-check, replay determinism) where applicable
- `make ci-check` passes, unless the change is documentation-only and a narrower check is clearly sufficient
- documentation is updated when behavior, commands, or expectations change
- secrets are not added, security-sensitive changes are called out, and dependency or workflow permission changes receive extra review
- Makefile targets, CI, hooks, and documented commands agree
