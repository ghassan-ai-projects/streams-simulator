# Architecture Context

## Repository Architecture

The public summary of the architecture is [`documentation/architecture/overview.md`](../../documentation/architecture/overview.md) (module groups, system shape and import direction); this file adds the per-module layers and ownership notes agents need. The complete import allow-list is `test/architecture/dependencies_test.go`; the package kinds and layer rules are in [`governance/MODULARITY_STANDARD.md`](../../governance/MODULARITY_STANDARD.md). The design record in [docs/design/](../../docs/design/) is a dated archive: read it for history, and verify it against the code and `contracts/`. Read the relevant `documentation/architecture/` page when a change affects a protocol, contract, trust boundary, or replay guarantee.

## Current Structure

- `AGENTS.md` is the canonical agent entrypoint.
- `README.md` introduces the project.
- `Makefile` defines the authoritative local commands.
- `.github/workflows/ci.yml` defines CI parity for core checks.
- `.golangci.yml` and `.pre-commit-config.yaml` enforce code quality and hygiene.
- `documentation/` holds curated public usage, architecture, benchmark, operations, and governance docs.
- `contracts/` holds every JSON Schema contract and the self-contained plant wire contract (`contracts/plant-wire/`); `domains/`, `adapters/` and `examples/` hold shipped data.
- `canvas/` is the TypeScript canvas client (its own toolchain; `canvas/ARCHITECTURE.md` is generated from `canvas/architecture.json`).
- `governance/` holds the registers the gates read; `test/architecture` holds the executable gates.
- `docs/` is a dated working archive of design notes, research, plans and reviews; nothing outside it reads from it.

## Current Package Ownership

- `cmd/streamsim/main.go`: version metadata and CLI entrypoint
- `internal/cli`: wiring only — facade (`Main`) over the commands in `internal/app` (flag parsing, scenario from flags, command registry, JSON output, failure documents) and the `process` and `files` edges; `test/architecture/cli_wiring_test.go` fails when logic creeps in
- `internal/world`: seeded discrete-event world core; domains are data loaded through one schema; facade over `internal/world/internal/domain`
- `internal/perturb`: perturbation layer between world and adapter (what the observer got, not what happened); facade over `internal/domain`
- `internal/adapter`: declarative output adapters projecting native `sim-event-v0.1` into consumer wire formats; facade over `internal/adapter/internal/domain` with a `files` edge
- `internal/adapter` also loads an adapter directory (`adapter.LoadDir`) and owns `adapter verify` (`adapter.Verify`): renders the fixture, validates the declared output schema and byte-compares the golden
- `internal/kernel`: the operating-system edge every module edge uses: reads, owner-only directories, atomic owner-only writes, append logs, owner-only unix sockets, interrupt handling; facade over `files`, `sockets` and `process` edges; only edge layers import it
- `internal/sink`: inproc, file, http-push; facade over `internal/sink/internal/domain` with `files` and `httppush` edge packages
- `internal/director`: the director and operator logic behind the MCP tools — world registry and identities, world commands (advance, fault, perturb, env, retire), run lifecycle (begin, end, seal and reveal truth, score, verify, audit), the operator view a capability token resolves to, the refusal taxonomy; typed inputs, no protocol knowledge; facade over `internal/director/internal/domain` (refusal codes, operator view, job plan), `app` (registry, jobs and use cases, `LoadDirector`), the `capability` edge (token entropy) and the `runner` edge (the job's wall pacer and goroutine); composes `run`, `score`, `truth`, `audit`, `world`
- `internal/canvasbridge`: the canvas bridge behind `streamsim canvas` — facade over `domain` (pairing and sessions, request admission, workspace path classes, the truth gate on run files, the action registry mapping each canvas action to a director MCP tool or a streamsim command), `app` (use cases), and the `transport` (loopback HTTP, Server-Sent Events, the built UI), `files` (the workspace `os.Root`), `process` (the director MCP child, streamsim commands, the poll ticker) and `entropy` edges; no simulation logic (docs/design/CANVAS_BRIDGE.md)
- `internal/mcp`: protocol only — facade over `internal/mcp/internal/protocol` (tool definitions, closed schemas, argument decoding into typed director values, SDK registration, `ServeDirector`) and the `transport` edge (stdio session and operator HTTP endpoint); one server, two roles — `director` (catalog · world · clock · fault · perturb · truth) and `operator` (nameplate · effectors · invoke · verdict); reaches the simulator only through `director` and foundations, held to it by `test/architecture/mcp_protocol_test.go`
- `internal/run`: run orchestration, delivery ledger, artifacts, replay, and quiescence, scenario checking and export (`ExportScenarioFile`), `ReplayFile`, `RunScenario`, and `LoadInputs` (the domain and adapter a run names); facade over `internal/run/internal/app` (orchestration), `domain` (pure rules) and the `durable`, `quiesce` and `clock` edges
- `internal/journal`: the run journal (`run-journal-v0.1`): two append-only JSON Lines streams written while a run happens, the observer stream (`journal.jsonl`, what a consumer could know) and the director stream (`sealed/journal.jsonl`, truth-bearing), each with a stream digest and a reader that tolerates a torn last line; facade over `internal/journal/internal/domain` (records, encoding, digest, reader) and a `files` edge (the two files)
- `internal/truth`: ground-truth generation and independent analytic solver; facade over `internal/truth/internal/domain`
- `internal/report`: the run report (`run-report-v0.1`) and the run comparison (`run-comparison-v0.1`): deterministic JSON or Markdown documents summarising a run folder (identity, digests, file-versus-recorded-digest verification, optional replay, delivery counts, faults, perturbations and effector outcomes from the command log, consumer, scoring, open questions) or comparing two ended runs (inputs, plan difference, comparability limits, whole-run results); facade over `internal/report/internal/domain` (assembly and renderings), `app` (read, grade, assemble) and a `files` edge; composes `run`, `journal` and `grading`
- `internal/grading`: offline grading of a finished run from its published files (replay for effector calls, then `score`); facade over `internal/grading/internal/domain` (decoding, evidence rules), `app` (workflow) and a `files` edge; composes `run` and `score`
- `internal/release`: the release manifest (file digests, canonical body, ed25519 signature, writing); facade over `internal/release/internal/domain`, `app` and a `files` edge
- `internal/score`: instrument and consumer scoring, shared online/offline policies; facade over `internal/score/internal/domain`; the host packs a `score.Evidence`
- `internal/audit`: trivial-baseline evaluation over delivered evidence; facade over `internal/audit/internal/domain`
- `internal/lint`: the domain plausibility check (`domain lint`): a noiseless nominal day of a domain sampled on a fixed grid and judged by five rules (non-finite, out-of-range, pinned-at-bound, not-settled, conservation residual); facade over `internal/lint/internal/domain`; composes `world` and `domain`
- `internal/suite`: scenario generation, composition and admission, and `Publish` (generate and write a suite file); facade over `internal/suite/internal/domain`, `app` and a `files` edge
- `internal/refconsumer`: shipped reference detector and operator-client integration; facade over `internal/refconsumer/internal/domain`, `app` (`Consume`: trace file in, verdict file out, optionally over the operator endpoint), the `mcpclient` edge and a `files` edge
- `internal/device`: data-defined capability admission and vendored device transport; facade over `internal/device/internal/domain` (rules, codec), the `uds` transport edge (including `Serve`, the socket served until interrupted) and a `files` edge (capability catalog)
- `internal/device/contract`: the vendored device wire-protocol schemas (embedded) and conformance fixtures
- `internal/plantwire`: the plant wire node: a run spoken to a host as births, data, the `NSYNC`/`HSYNC` barrier and host records over the host's owner-only socket; facade over `internal/plantwire/internal/domain` (records, digests, schema checks, sequencing, kind capture from a birth, relay, interrupt rules, board mapping, twin report), `app` (the session, relay loop, control port), the edges `uds` (sockets, ticks, control server), `host` (the reference host) and `files`, and `standin`, a scripted node for tests that production never imports
- `internal/plantwire/contract`: the embedded plant wire schemas, byte-identical with `contracts/plant-wire/schemas`
- `internal/deviceworld`: device-to-world effector binding and `OpenDevice` (capabilities, fault schedule, bindings, world and plant composed before any listener opens); facade over `internal/deviceworld/internal/domain`, `app` and a `files` edge
- `internal/domain`: load, validate, compile and digest domain data, including a directory into a catalog (`NewCatalogFromDir`); facade over `internal/domain/internal/domain` with a `files` edge
- `internal/component`: the component library and `Compose`: load the embedded built-in library, then customer directories (`NewLibrary`), of equipment, instruments and vocabulary, check them, list and describe them, and compose a composition into an ordinary domain; the `Assumptions` of a domain or twin (what it declares about each reading and command and what it does not model, `domain-assumptions-v0.1`) the `Lineage` of the plants of a workspace (which composition or twin produces each, kind files, runs, edit route, `domain-lineage-v0.1`) and the `Diff` of two domains, files or run folders (`domain-diff-v0.1`: per kind readings, dynamics, commands, faults, states, affected instances and relations, the silent-edit flag); facade over `internal/component/internal/domain` (library rules, targets, presets, instruments, ports, links, assumptions, lineage) and a `files` edge (library directories, the workspace folders and the run folders' `run.json`); composes `domain` for validation and formatting
- `internal/model`: shared simulator records and strict JSON decoding
- `internal/jsonschema`: core schema compilation and validation
- the built-in component library (vocabulary, equipment, instruments) is data in `internal/component/internal/files/builtin`, embedded by that files edge; example compositions are in `examples/compositions`
- `internal/schemas`: embedded contract sources
- `internal/canonical`: canonical JSON encoding and digest operations
- `internal/randutil`: seeded random streams
- `internal/jsonvalue`: one reading of decoded JSON values: numbers of any decoded form, documents that refuse undeclared fields
- `internal/wall`: wall-clock seam and deterministic build replacement
- `canvas/` (not Go): the browser client; Go never imports it and it reaches the binary only through `canvasbridge`
- `internal/testsupport`: shipped-fixture paths and short socket paths for tests; imported only by `_test.go` files
- `test/`: integration and end-to-end suites

## Deliberate Placements

- **Perturbation sits between world and adapter** — the world produces what physically happened, the perturbation layer what the observer got. A scenario the consumer never saw is scored as a transport miss, not a reasoning miss.
- **The adapter sits last** — every consumer sees the same perturbation, so cross-consumer comparison means something.
- **MCP manages the simulator, it does not transmit the streams.** Evidence leaves on a sink; MCP carries control and actuation.

## Dependency Direction

- `mcp` -> `director` + foundations; `director` -> `run`/`score`/`truth`/`audit`/`world`/`adapter`/`component`/`domain`/`model`
- `component` -> `domain` + foundations; `canvasbridge` -> `kernel` only (it drives the binary as child processes)
- `cmd` -> `cli` -> the facades it wires: `mcp`/`canvasbridge`/`run`/`suite`/`grading`/`report`/`release`/`refconsumer`/`device`/`deviceworld`/`plantwire`/`component`/`domain`/`adapter`/`lint`/`model`/`kernel`
- `grading` -> `run`/`score`/`domain`/`world`/`model`; `report` -> `grading`/`run`/`journal`/`model`; `release` -> `canonical`
- `run` -> `world`/`perturb`/`adapter`/`sink`/`journal`/`domain`/`model`/`canonical`; `journal` -> `model`
- `lint` -> `world`/`domain`/`model`
- `truth` -> `world`/`domain`/`model`; `score` -> `domain`/`world`/`model` (hosts pack a `score.Evidence`)
- `suite` -> `audit`/`truth`/`perturb`/`world`/`domain`/`model`/`randutil`; `audit` -> `domain`/`perturb`/`world`/`model`
- `plantwire` -> `run`/`truth`/`world`/`domain`/`model`; `deviceworld` -> `device`/`domain`/`world`/`model`; neither core imports the bridge
- Only module edge packages import `kernel`
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

Package kinds, layer shapes and purity/I-O rules (M1-M11) are in
[governance/MODULARITY_STANDARD.md](../../governance/MODULARITY_STANDARD.md); the migration rounds are in
[governance/MODULARITY_PLAN.md](../../governance/MODULARITY_PLAN.md). Where this file and the standard
disagree about layering, the standard and `test/architecture` are current and this file is updated in the
same round that changes ownership.
