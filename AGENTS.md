# AGENTS.md - Streams Simulator

Canonical instructions for coding agents in this repository. Read this file first, then load only the context files needed for the task.

## Purpose

Streams Simulator (`streamsim`) is a standalone, deterministic, closed-loop world simulator for testing stream processors. It generates realistic event streams from data-defined domains, corrupts their delivery in specified ways, carries sealed ground truth, accepts commands back through declared effectors so the world actually changes, and scores whatever consumed it.

It is a **test instrument**. An instrument less trustworthy than the system it measures is worse than no instrument, because it produces confident wrong answers. Determinism guarantees, generative ground truth, and analytic oracles come before features.

The design is fully specified in [docs/](docs/README.md). The nine non-negotiables in the spec (determinism, analytic cross-check, reference consumer, delivery ledger, quiescence barrier, sealed oracle, injection probe, `silent_no_effect` test, trivial-baseline audit) are not negotiable.

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

Use the prompt files under `.agents/prompts/` when the task matches them.

## Current Repository State

- Module path: `github.com/ghassan-ai-projects/streams-simulator` (set).
- Design is complete and committed under `docs/` (research, design, contracts, adapters, examples).
- Implementation is complete: the CLI lives in `cmd/streamsim/` and `internal/` holds
  the world, perturbation, adapter, sink, MCP, ledger, and truth packages. The shipped
  domains (`domains/*.domain.json`), adapters (`adapters/*.adapter.json`), and
  `bin/streams-simulator` are produced by `make build`.
- Domain specs, adapters, and effectors are **data** (JSON), never code.

Do not invent architecture outside the documented design. The spec was written to be built as specified; deviations need a design change first.

## Architecture Overview

The documented shape (see [docs/design/TECHNICAL_DESIGN.md](docs/design/TECHNICAL_DESIGN.md)):

- `cmd/streamsim/main.go` - entrypoint, flags, wiring, shutdown
- `internal/world` - seeded discrete-event world core; domain specs are data, loaded through one schema
- `internal/perturb` - perturbation layer between world and adapter (what the observer got, not what happened)
- `internal/adapter` - declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats
- `internal/sink` - inproc, file, http-push sinks
- `internal/mcp` - one MCP server, two roles: `director` (catalog, world, clock, fault, perturb, truth) and `operator` (nameplate, effectors, invoke, verdict)
- `internal/ledger` - delivery ledger; distinguishes transport misses from reasoning misses
- `internal/truth` - sealed ground truth and scoring
- `docs/` - the design, research, contracts, and examples (source of truth for behavior)

Dependency direction:

- `cmd` -> `mcp` -> `world`/`perturb`/`adapter`/`sink`/`ledger`/`truth`
- Dependencies flow downward only.
- Domain specs, adapters, and effectors are **data** (JSON), never code. If any domain needs a code branch in the binary, the simulator is wrong and the domain found the bug.
- The simulator has no knowledge of its consumers. No consumer name, schema, field, or behaviour appears in the binary.

Determinism: a run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`. One artifact reproduces any failure.

## Build, Test, and Lint

Primary commands:

- `make ci-check`
- `make build`
- `go vet ./...`
- `go test ./...`
- `make lint`
- `git diff --check`
- `pre-commit run --all-files` when `pre-commit` is installed

Important behavior:

- `make build` skips gracefully until `cmd/` exists.
- `make test` and `go test ./...` operate on the root scaffold package until real packages exist.
- `make lint` depends on `golangci-lint` and may fail if the environment cannot write to its cache.

See [.agents/context/testing.md](.agents/context/testing.md) for the testing and validation bar. The analytic cross-check (implement the integrator twice, assert agreement) is a non-negotiable correctness oracle, not a consistency check.

## Go Standards

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
