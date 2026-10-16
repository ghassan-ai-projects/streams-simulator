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

## MUST: architecture integrity

These rules are binding for every agent and every subagent. A change that breaks one is not done, whatever the gates say.

1. **Logic lives in its module.** Every rule, validation, default, policy, derived value and orchestration step belongs to
   the domain layer (or app layer) of the module that owns it. `internal/cli` holds only flag parsing, one facade call per use
   case, the output document and the failure mapping. `internal/mcp` holds only the protocol. A facade only delegates.
2. **One use case, one implementation.** When two surfaces (CLI, director/MCP, plant wire, canvas bridge) need the same
   operation, the owning module exposes it once and both call it. A surface never re-parses, re-validates, re-computes or
   re-shapes a module's result beyond its own output document.
3. **Never loosen a gate to pass.** Do not raise a line cap, add an allowed import, lower a coverage floor, add a test hook,
   or remove or weaken any check in `test/architecture` or `scripts/` to make a change pass. Move the code instead. A gate
   change is allowed only when the owner approved it in the conversation, and the commit message names that approval.
   Registering a genuinely new package with the kind, layer and imports the standard prescribes is not loosening.
   `make gate-ratchet` (part of `ci-check`) enforces this against the branch point from main: a cap may fall, never rise;
   an allow-list may shrink, never grow; a coverage floor may rise, never fall; a test hook may be removed, never added.
   The owner's approval is a commit line `Owner-approved-gate-change: <reason>`.
4. **No new module, package or top-level folder without a design.** State its kind, layer, responsibility and callers
   ([MODULARITY_STANDARD](governance/MODULARITY_STANDARD.md)) before writing it. Go packages live under `cmd/` and
   `internal/` (plus the root `doc.go`); anything else needs owner approval (`TestGoPackagesLiveUnderCmdAndInternal`). Data a module
   embeds lives in that module's edge package (the built-in component library is in `internal/component/internal/files/builtin`).
5. **Delegated work is reviewed at the boundary.** The agent that delegates reviews every diff its subagents make to
   `test/architecture`, `scripts/`, `AGENTS.md` and module facades before merging. Each PR lists every gate change and its
   reason: run `git diff main -- test/architecture scripts` and justify each line.

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
- [governance/MODULARITY_STANDARD.md](governance/MODULARITY_STANDARD.md) for package kinds, layer rules and the M1-M11 modularity bar; its [PLAN.md](governance/MODULARITY_PLAN.md) tracks the migration rounds and [DEFERRED.md](governance/DEFERRED.md) lists known defects that are intentionally not fixed by structural rounds.

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
- `internal/cli` - wiring only: facade (`Main`) over the commands (`app`: flag parsing, scenario from flags, command registry, JSON output and failure documents) and the process and file edges; a gate (`test/architecture/cli_wiring_test.go`) keeps logic out
- `internal/world` - seeded discrete-event world core; domain specs are data, loaded through one schema
- `internal/perturb` - perturbation layer between world and adapter (what the observer got, not what happened)
- `internal/adapter` - declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats
- `internal/sink` - inproc, file, http-push sinks
- `internal/director` - the director and operator logic behind the MCP tools: world registry, world commands, run lifecycle (seal and reveal truth, score, verify, audit), the operator view a capability token resolves to; typed inputs, no protocol knowledge
- `internal/mcp` - protocol only: one MCP server, two roles, `director` (catalog, world, clock, fault, perturb, truth) and `operator` (nameplate, effectors, invoke, verdict); tool definitions, schemas, argument decoding and transport (`mcp.ServeDirector`: stdio session plus the operator HTTP endpoint, in the `transport` edge); it calls `director` and holds no rule (`test/architecture/mcp_protocol_test.go`)
- `internal/run` - run orchestration, delivery ledger, artifacts, replay, and quiescence, scenario checking, running, replay and export from files (`run.RunScenario`, `run.ReplayFile`, `run.ExportScenarioFile`), and resolving the domain and adapter a run names (`run.LoadInputs`); there is no separate `internal/ledger` package
- `internal/journal` - the run journal (`run-journal-v0.1`): observer and director JSON Lines streams written while a run happens
- `internal/report` - the run report (`run-report-v0.1`): one deterministic JSON or Markdown document summarising a run folder (composes `run`, `journal` and `grading`)
- `internal/grading` - offline grading of a finished run: loads the published files, replays the artifact for its effector calls and scores with `score` (composes `run` and `score`, which do not know each other)
- `internal/component` - the component library (equipment, instruments, vocabulary as data) and `Compose`, which writes a composition as an ordinary domain (id templates, cycles and every other domain rule are the `domain` module's, checked on the composed domain); composes `domain` for validation and formatting
- `internal/lint` - the domain plausibility check (`streamsim domain lint`): a noiseless nominal day of a domain sampled on a fixed grid and judged by five rules; composes `world` and `domain`
- `internal/release` - the release manifest: file digests, canonical body, ed25519 signature, writing
- `internal/truth` - sealed ground truth and independent analytic solver
- `internal/score` - online/offline scoring policies
- `internal/audit`, `internal/suite`, `internal/refconsumer` - trivial-baseline audit, scenario generation and publishing (`suite.Publish`), shipped reference consumer (`refconsumer.Consume` runs it over a trace file, optionally over the operator endpoint)
- `internal/plantwire` - the plant wire node: a run spoken to a host over the host's owner-only socket (births from the domain through an allow-list, data, the `NSYNC`/`HSYNC` barrier, host records in sequence order); schemas embedded from `contracts/plant-wire`
- `internal/device`, `internal/deviceworld` - device emulation (rules, capability file, unix-socket serving until interrupted) and its world integration bridge (`deviceworld.OpenDevice` composes a device with a bound world)
- `internal/kernel` - the operating-system edge: reads, owner-only directories, atomic owner-only writes, append logs, owner-only unix sockets, interrupt handling; only module edge packages import it, and it knows no record or file name
- `internal/domain`, `internal/model`, `internal/jsonschema`, `internal/jsonvalue`, `internal/schemas`, `internal/canonical`, `internal/randutil`, `internal/wall` - data loading (directory loading of domains and adapters lives in the `domain` and `adapter` files edges), shared records and bounded foundations
- `contracts/plant-wire/` - the self-contained plant wire contract (spec, one schema per record, fixtures, pinned `DIGEST`); `test/contract` validates it, `make contract-sync` copies it from `PLANT_WIRE_CONTRACT_DIR`
- `contracts/` - every JSON Schema contract the binary embeds (copies in `internal/schemas`, kept equal by a test) and the plant wire contract
- `examples/` - example domains the tests load
- `governance/` - the registers the gates read: findings, phase plans, the modularity standard and deferred defects
- `test/pin/` - the behaviour pin the phase gate diffs against
- `docs/` - a working directory for design notes, research and evidence; nothing outside it may read from it (`test/architecture/docs_boundary_test.go`)
- `documentation/` - curated public product, usage, architecture, benchmark, operations, and governance documentation

Dependency direction:

- `cmd` -> `cli` -> the facades it wires (`mcp`, `run`, `suite`, `grading`, `report`, `release`, `refconsumer`, `device`, `deviceworld`, `domain`, `adapter`, `model`)
- `mcp` -> `director`; `director` composes `run`/`score`/`truth`/`audit`/`world`; `mcp` imports none of those directly
- `grading` -> `run`/`score`; `report` -> `grading`/`run`/`journal`; the CLI holds no business logic, only flags, documents and failures
- `run` composes `world`/`perturb`/`adapter`/`sink`; truth and scoring consume world/evidence. `plantwire` drives a `run` as a plant wire node. `deviceworld` bridges `device` and `world` without either core importing the bridge.
- Dependencies flow downward only.
- Domain specs, adapters, and effectors are **data** (JSON), never code. If any domain needs a code branch in the binary, the simulator is wrong and the domain found the bug.
- The simulator has no knowledge of its consumers. No consumer name, schema, field, or behaviour appears in the binary.

Determinism: a run is a pure function of `(sim_version, domain_digest, adapter_digest, seed, command_log, sink)`. One artifact reproduces any failure.

## Build, Test, and Lint

Primary commands:

- `make quick` (inner loop, about 30 s with the test cache: fmt, function length, build, vet, short tests including `test/architecture`)
- `make ci-check`
- `make phase-gate` (simulation-canvas goal: `ci-check` plus the Enola check)
- `make test-long` and `make test-race` (on demand only: long simulations and the race detector; not in any pipeline)

Use `make quick` while working. Run `make phase-gate` once per pull request, before pushing (CI runs `ci-check` again); not after every step.
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

Every module is a public facade (`internal/<m>`) over private layers
(`internal/<m>/internal/{domain,app,<edge>}`) with no internal leaks, shaped by
its kind (pure core, core with an edge, surface; foundations stay single
packages) as described in [STANDARD.md](governance/MODULARITY_STANDARD.md).
Domain layers import no `os`/`net`/`exec` and read no wall clock; file, socket
and clock access sits only in declared edge packages; a facade only delegates;
`cli` and `mcp` hold wiring and protocol only. Tests live at the layer they
prove (test bar T1-T10). Use [.agents/prompts/module-refactor.md](.agents/prompts/module-refactor.md)
to migrate a package. A structural round
changes no behaviour: run `scripts/behaviour-pin` and diff it against
`test/pin/BEHAVIOUR_PIN.txt`. Defects found during
a round are added to `DEFERRED.md`, not fixed in the same commit.

## Go Standards

### Clean-code and architecture bar

- Every Go source file, including tests, must be at most **300 total lines** (comments and blank lines count). Split files by a named responsibility within their owning package; do not split functions arbitrarily or compress code to meet the limit.
- Function names state domain intent. Each function performs one task at one abstraction level. Entry points read as a sequence of simulator operations; put parsing, serialization, record bookkeeping, and concrete mechanics in named steps below their callers.
- Every production Go function, method and anonymous function must have at most **15 physical body lines**, from the opening brace through the closing brace, including comments and blank lines. Test functions are excluded from this limit. Keep private helpers in stepdown reading order. Extract named responsibilities; do not compress statements to meet the limit.
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
- Comments: as few as possible. Each package has its one-line package doc comment (gated); otherwise state intent through names and structure, and add a single short line only where a non-obvious rule cannot be expressed in code. Do not add doc comments to exported symbols by default.
- Use table-driven tests with `t.Run()` and `t.Parallel()` where safe.
- Use `t.Context()` in tests when appropriate.
- RFC 8785 canonical JSON everywhere a digest is computed.

See [.agents/context/go-style.md](.agents/context/go-style.md) for the repo-specific style rules.

## Canvas (TypeScript)

- The canvas lives in `canvas/` with its own toolchain (Node 24, npm, exact pinned versions); read `canvas/README.md` and
  `governance/CANVAS_UI_ACCEPTANCE.md` first. Gates: `make canvas-check` (part of `make phase-gate`).
- For Canvas architecture or refactoring work, also read the
  [accepted UI architecture](docs/simulation-canvas/ui-architecture-20261016/README.md); it is the target structure,
  dependency contract, migration order and enforcement plan.
- Layers `model` ← `app` ← `adapters`; `ui` uses `app` and `model` only and never touches transport (dependency-cruiser).
- The canvas computes nothing the binary owns: validation, composition, simulation, truth, scoring and verification come
  from `streamsim` through the bridge. A missing capability goes into the simulator first.
- Strict TypeScript (no `any`, no non-null assertion, no ts-ignore); files ≤ 300 lines; functions and components ≤ 40
  lines; cyclomatic ≤ 10, cognitive ≤ 15; no duplication of 75 tokens; no unused export; zero lint warnings.
- Tests: unit (`model` ≥ 90 %, `app` ≥ 80 % lines), component tests with axe, Playwright end-to-end against the real binary.
  No mocks of the binary in end-to-end tests, no fixed sleeps.
- Desktop only for now (≥ 1280 px). Every UI round ends with a design review in Chrome against the checklist in
  `governance/CANVAS_UI_ACCEPTANCE.md` §3, recorded in the phase review.
- Go never imports or embeds anything under `canvas/` (`test/architecture/canvas_test.go`).

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
- Do not put code, contracts, fixtures, examples, pins or gate registers in `docs/`, and do not read from `docs/` in code, tests or scripts.
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
