# P2d evidence — director out of the MCP surface

- **Date**: 2026-10-12
- **Change**: structural only. `internal/mcp/internal/app` (about 1000 production lines) held the
  director, which is application logic, not protocol. It is now the module `internal/director`;
  `internal/mcp` keeps tool definitions, schemas, argument decoding, the wire form of a refusal and
  transport. A gate fails when `mcp` regrows logic.
- **Behaviour**: unchanged. `make behaviour-pin-check` matches the committed pin; the pin file is
  not edited. Function bodies moved verbatim; tool descriptions, schemas, refusal codes and refusal
  text are byte-identical (the existing "Active faults. app.Director only." description string is
  untouched; it is a known wire-text oddity, not part of this round).

## What moved where

| Code | From | To |
|---|---|---|
| World registry, identities, world creation (`Director`, `WorldRecord`, `CreateWorld`, `registerWorld`, `publishWorld`, nameplate building) | `mcp/internal/app/{director,world,world_config}.go` | `director/internal/app` (same files) |
| World commands (advance, clock state, fault, perturb, env, describe, destroy) | `mcp/internal/app/world_commands.go` | `director/internal/app/world_commands.go`; gains `AdvanceBy` (the relative-delta checks that sat in the protocol handler) and `RetireEntity` (the world lookup and retire that sat in the protocol handler) |
| Run lifecycle (begin, seal and reveal truth, end, score, verify, audit) | `mcp/internal/app/run.go` | `director/internal/app/run.go` |
| Catalog and adapter queries the handlers answered by reaching into `Director.Catalog` / `Director.Adapters` | `mcp/internal/protocol/director_catalog_handlers.go`, `director_resources.go` | `director/internal/app/catalog.go` (`CatalogList`, `CatalogDescribe`, `CatalogCoverage`, `DomainSpec`, `ListAdapters`) |
| Operator view: capability check, effector validation, acknowledgement, report, nameplate | `mcp/internal/app/operator.go` | `director/internal/domain/operator.go` (pure); the token resolution the handlers did moves to `director/internal/app/operator_ops.go` (`RequireOperator`, `OperatorNameplate`, `OperatorEffectors`, `OperatorInvoke`, `OperatorReport`) |
| Refusal taxonomy (`Code`, `ToolError`, `ToolErrorf`) | `mcp/internal/app/errors.go` | `director/internal/domain/errors.go`; facade `director.Code`, `director.Code…` constants, `director.ToolErrorf` |
| Capability-token entropy edge | `mcp/internal/capability` | `director/internal/capability` |
| `LoadDirector` | `mcp/internal/app/serve.go` (with `ServeOptions`) | `director/internal/app/load.go`, now `LoadDirector(ctx, domainsDir, adaptersDir, outDir)`; facade `director.NewFromDirs`. `ServeOptions` (with the operator address) stays in `mcp/internal/protocol` and `mcp.ServeOptions` |
| Argument readers `Str`, `Num`, `IntArg` | `mcp/internal/app/arguments.go` | `mcp/internal/protocol/arguments.go` as `stringArg`, `numberArg`, `intArg` |
| Decoding of `sim.world.create` arguments (`seed`, `start_time` presence, `entities`, defaults) | `mcp/internal/app/world_config.go` (map arguments) | `mcp/internal/protocol/world_request.go` decodes into the typed `director.WorldRequest`; the defaults (`native-jsonl`, in-process sink, stepped time, default start) stay in the director |
| Exactly-one-of `by_ns` / `to_ns` check | `mcp/internal/protocol` | stays in `mcp/internal/protocol` (argument decoding), same precedence: before any world lookup |
| Operator-server construction over a single view (`NewOperatorServer`, `SingleView`, `OperatorResolver`) | `mcp/internal/protocol`, `app` | removed: only tests used them; `NewOperatorServerResolver(*director.Director)` is the one operator server |
| `mcp.Director`, `mcp.NewDirector`, `mcp.SetOperatorEndpoint` | `mcp` facade | `director.Director`, `director.New`, `Director.SetOperatorEndpoint` |

Tests moved with the code at the layer they prove:

| Test | From | To |
|---|---|---|
| Report and operator-view tests (report refusal, run binding, acknowledgement allow-list, every mode reads like success) | `mcp/internal/app` | `director/internal/domain` (they exercise the view over a world, no director); plus new `operator_view_test.go` |
| Use-case, lock-order, capability, scoring-evidence, audit, destroy, concurrent-read, load tests | `mcp/internal/app` | `director/internal/app` (typed `WorldRequest` in place of argument maps) |
| Argument-reader test | `mcp/internal/app` | `mcp/internal/protocol/arguments_test.go` |
| Golden closed loop over the operator endpoint; prefix indistinguishability (two worlds, same delivered prefix, identical operator responses) | `mcp/internal/protocol` | `director/closed_loop_test.go`, `director/operator_surface_test.go` (external `director_test`, driving `mcp.NewOperatorServerResolver` and a test-only `RunOf` seam in `export_test.go`, because both observe and park the run beneath the tools) |
| MCP strictness, role separation, tool and resource tests | `mcp/internal/protocol` | stay; they read world identities from `sim.world.create` instead of reaching into the registry |
| New facade contract tests (lifecycle, catalog and audit, operator token) | — | `director/facade_test.go` |

## Facade of `internal/director`

`New(ctx, *domain.Catalog, adapters, outDir) *Director`, `NewFromDirs(ctx, domainsDir, adaptersDir, outDir)`;
types `WorldRequest`, `Nameplate`, `EffectorInfo`, `Code` and the fifteen `Code…` constants,
`ToolErrorf`; `Director` methods: `CreateWorld(WorldRequest)`, `DescribeWorld`, `DestroyWorld`,
`Advance`, `AdvanceBy`, `ClockState`, `InjectFault`, `ClearFault`, `ListFaults`, `ApplyPerturb`,
`ClearPerturb`, `EnvInject`, `RetireEntity`, `BeginRun`, `EndRun`, `SealTruth`, `RevealTruth`,
`SealStatus`, `Score`, `VerifyRun`, `AuditScenario`, `CatalogList`, `CatalogDescribe`,
`CatalogCoverage`, `DomainSpec`, `ListAdapters`, `SetOperatorEndpoint`, `RequireOperator`,
`OperatorNameplate`, `OperatorEffectors`, `OperatorInvoke`, `OperatorReport`. The facade holds the
internal director behind a private field; no registry record, run or truth store leaves it.

Facade of `internal/mcp` after the round: `ServeDirector`, `ServeOptions`, `NewDirectorServer(*director.Director)`,
`NewOperatorServerResolver(*director.Director)`.

## New module

| Module | Kind | Layers | Package layer |
|---|---|---|---|
| `internal/director` | core with an edge | facade · `domain` (refusal taxonomy, operator view) · `app` (registry and use cases) · `capability` (entropy edge) | 6 (composes `run`, `score`, `truth`, `audit`, `world`) |

`internal/mcp` stays a surface at layer 7 with `protocol` and `transport` only (its `app` and
`capability` layers are gone).

## Dependency allow-list (`test/architecture/dependencies_test.go`)

| Package | Before | After |
|---|---|---|
| `internal/mcp` | adapter audit domain model run schemas score truth world | director model schemas |
| `internal/director` | — | adapter audit domain model run score truth world |

The I/O edge inventory entry for the entropy file moved to
`internal/director/internal/capability/capability.go`. `moduleShapes`: `mcp` is `protocol`,
`transport`; `director` is `domain`, `app`, `capability`. The surface-shape check now accepts a
surface that has an `app` or a `protocol` layer. `.golangci.yml` adds the director facade to the
`wrapcheck` ignore list: its refusals are returned to the SDK unwrapped because their text is the
wire text.

## The role gate

`test/architecture/mcp_protocol_test.go`, `TestMCPHoldsOnlyProtocol`: `internal/mcp` may contain
only the facade and the `protocol` and `transport` packages, and a file in it may import only
`internal/director`, its own packages and foundations. `TestMCPRulesRejectBusinessLogic` proves
seven rejections on injected snippets (a `run`, `score`, `truth`, `world` or `audit` import, an
`app` layer, an entropy edge); `TestMCPRulesAcceptProtocol` proves the accepted shape.

Failure proof on the real tree: a file `internal/mcp/internal/protocol/injected_violation.go`
importing `internal/score` made the gate fail with
`imports internal/score; mcp decodes and frames, the director decides, so reach it through internal/director`
(and `TestPackageDependencies` failed too); a file in a new `internal/mcp/internal/app` package
made it fail with `mcp has no package internal/mcp/internal/app`. Both files were removed and the
gate passes.

## Gate results

| Gate | Result |
|---|---|
| `make phase-gate` | exit 0 (function-length, fmt-check, docs-check, tidy, build, vet, golangci-lint 0 issues, `test-short`, `test-simdet`, `coverage-check`, deadcode, vulncheck, fuzz-soak, Enola `architecture` check, `behaviour-pin-check`, `git diff --check HEAD`) |
| `go test ./...` | passes |
| `make behaviour-pin-check` | "Behaviour pin matches" |
| Coverage | `director` 100 %, `director/internal/app` 73 %, `director/internal/domain` 94 %, `director/internal/capability` 75 % (floor 70 %) |
| Enola `make architecture` | passes against the existing baseline (no new cycle or layer finding); the baseline predates P2c, so its "new coupling" listing is large and reported only |

## Notes

- `TestServeRunsTheDirectorUntilTheSessionEndsThenStopsTheOperatorEndpoint` (transport) failed once
  in one run of `make test-simdet` and passed on rerun: it dials the released ephemeral port, so a
  parallel test that binds `127.0.0.1:0` can take the port. The test is unchanged by this round.
- Not done: no change to wire text, schemas or error codes; the `app.Director only.` description
  string noted above is left as it is.
