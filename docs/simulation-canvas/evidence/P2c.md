# P2c evidence — CLI to wiring only

- **Date**: 2026-10-12
- **Change**: structural only. `internal/cli` held business logic and protocol serving. Each
  responsibility moved to the module that owns it; two modules were created because none owned
  the work (`grading`, `release`). A gate now fails when the CLI regrows logic.
- **Behaviour**: unchanged. `make behaviour-pin-check` matches the committed pin; the pin file is
  not edited. Error text is byte-identical because function bodies moved verbatim (the CLI keeps
  its own `streamsim: ` wraps where it had them).

## What moved where

| Code | From | To |
|---|---|---|
| Director session over stdio and the operator HTTP endpoint (`Director`, `directorOn`, `operatorEndpoint`, `serveOperator`, `runSession`) | `cli/internal/serve` | `mcp/internal/transport` (`BindOperator`, `Serve`, `Endpoint`) composed by `mcp/internal/protocol.ServeDirector`; facade `mcp.ServeDirector`, `mcp.ServeOptions` |
| `loadDirector` (catalog and adapters into the director) | `cli/internal/app/mcp.go` | `mcp/internal/app.LoadDirector` |
| Device socket served until interrupted (`Device`, `deviceUntil`, `listenFailure`, `shutDown`) | `cli/internal/serve` | `device/internal/uds/serve.go` (`Serve`); facade `device.Serve` |
| Capability file loading | `cli/internal/app/device_world.go` | `device/internal/files` (`LoadCapabilities`); facade `device.LoadCapabilitiesFile` |
| Device-world wiring (world creation, entity selection, binding, validation, plant, device clock, device construction) | `cli/internal/app/device_world.go`, `device.go` | `deviceworld/internal/app` (`OpenDevice`), pure steps in `deviceworld/internal/domain/composition.go` (`SelectEntity`, `ValidateComposition`), bindings file in `deviceworld/internal/files`; facade `deviceworld.OpenDevice`, `OpenOptions` |
| Score workflow (artifact, verdict, label, ledger, history, replay for evidence, grade) | `cli/internal/app/score.go` | new module `grading`: `grading/internal/app` (`Grade`), `grading/internal/domain` (`DecodeLedger`, `DecodeHistory`, `DecodeDocument`, `ArtifactEvidence`, `ReplayMismatchError`), `grading/internal/files`; facade `grading.Grade`. The CLI turns the typed `*grading.ReplayMismatchError` into the `replay_mismatch` outcome |
| Release manifest (file digests, canonical body, ed25519, writing) | `cli/internal/app/manifest.go`, `manifest_options.go` | new module `release`: `release/internal/domain` (`Body`, `Canonical`, `Document`, `Seed`, `Sign`, `FileNames`), `release/internal/app` (`Publish`), `release/internal/files`; facade `release.Publish`. Flag parsing stays in the CLI |
| `loadCatalog` | `cli/internal/app/catalog.go` | `domain/internal/files.LoadCatalog`; facade `domain.NewCatalogFromDir` |
| `loadAdapters`, `decodeAdapterDirectory`, `loadAdapterEntries`, `FileNames` | `cli/internal/app/catalog.go`, `cli/internal/files` | `adapter/internal/files.LoadDir`; facade `adapter.LoadDir` |
| `loadSimulatorInputs`, `describeSimulatorInputs` | `cli/internal/app/simulator_inputs.go` | `run/internal/app.LoadInputs`; facade `run.LoadInputs` (run already imports `domain` and `adapter`) |
| Reference consumer orchestration (read trace, connect over MCP, detect, write verdict) | `cli/internal/app/refconsumer.go` | `refconsumer/internal/app` (`Consume`), `refconsumer/internal/files`; facade `refconsumer.Consume`, `ConsumeOptions`, `ConsumeSummary`. Flags and admission stay in the CLI |
| Scenario export pipeline (`exportConfig`, `checkedExport`, `encodeScenario`, artifact and input loading) | `cli/internal/app/scenario.go` | `run/internal/app.ExportScenarioFile` returning `ScenarioExport{Scenario, Omitted, Refused, Raw}`; facade `run.ExportScenarioFile(ctx, artifactPath, dirs)`. The CLI writes the file |
| Scenario run (`checkedScenarioConfig`, `executeScenarioRun`, `publishRun`) | `cli/internal/app/run.go` | `run/internal/app.RunScenario`; facade `run.RunScenario(ctx, scenario, options)` with a typed `*run.ScenarioCheckError` the CLI restates for a scenario file or for flags |
| Replay (`executeReplay`, input loading) | `cli/internal/app/replay.go` | `run.ReplayFile(ctx, artifactPath, dirs)`; `cmdVerify` decides from the typed result |
| Suite workflow (catalog load and describe, profile admission, file name, marshal, write, `SampleNS`) | `cli/internal/app/suite.go` | `suite/internal/app.Publish`, `suite/internal/files`, `suite/internal/domain` (`FileName`, `DefaultSampleNS`, `UnknownProfileError`); facade `suite.Publish`. `suite` becomes a core with an edge |
| The `--start-time` bound (`world.MaxStartNS`) | `cli/internal/app/run_options.go` | `run.MaxStartNS` re-exports the world's bound; the check and its usage text are unchanged and the CLI no longer imports `world` |
| The rule for when `adapter verify` fails | `cli/internal/app/catalog.go` | `adapter.VerifyFailure` |
| Tests of whole flows (run → replay → verify, refconsumer → score, tampered verify, scripted effector ids) | `cli/internal/app/workflow_command_test.go`, `workflow_steps_test.go` | `test/acceptance/workflow_test.go`, through the built binary |
| Offline-equals-online score test, ledger decoding tests, manifest signing test, device-world tests, serve tests | `cli/internal/app`, `cli/internal/serve` | `grading/internal/app`, `grading/internal/app` and `domain`, `release/internal/domain`, `deviceworld/internal/app`, `device/internal/uds`, `mcp/internal/protocol`, `mcp/internal/transport` |
| Flag diagnostics and option-parsing tests | `workflow_*_test.go` | stay in the CLI as `diagnostics_command_test.go` and `run_script_test.go` |

`cli/internal/files` stays: it is the CLI's own file edge (the scenario file it reads, the scenario
and suite files it writes). `cli/internal/serve` is deleted; `cli/internal/process` stays.

Size: the CLI command layer went from 2405 to 1727 production lines (2521 with the serve edge); no
business function is left in it.

## New modules

| Module | Kind | Layers | Package layer |
|---|---|---|---|
| `internal/grading` | core with an edge | facade · `domain` · `app` · `files` | 6 (composes `run` and `score`, both 5) |
| `internal/release` | core with an edge | facade · `domain` · `app` · `files` | 1 (imports `canonical`) |

New layers in existing modules: `mcp/transport` (edge), `deviceworld/app` and `deviceworld/files`,
`device/files`, `refconsumer/app` and `refconsumer/files`, `suite/app` and `suite/files`.
`deviceworld`, `refconsumer` and `suite` are reclassified from core to core with an edge.

## Dependency allow-list (`test/architecture/dependencies_test.go`)

| Package | Before | After |
|---|---|---|
| `internal/cli` | adapter canonical device deviceworld domain mcp model refconsumer run score suite world | adapter device deviceworld domain grading mcp model refconsumer release run suite |
| `internal/mcp` | audit domain model run schemas score truth world | adds `adapter` |
| `internal/deviceworld` | device model world | adds `domain` |
| `internal/grading` | — | domain model run score world |
| `internal/release` | — | canonical |

The MCP SDK is no longer allowed in `internal/cli`. The CLI's `canonical`, `score` and `world`
edges are gone; the gate reports no stale edge.

## Audit of production files the branch touched

| Files | Verdict |
|---|---|
| `cli/internal/app/*` (arguments, catalog, cli, device, device_options, failure_*, failures, mcp, manifest_options, refconsumer, replay, run, run_flag_issues, run_options, run_script, scenario, score, suite) | Misplaced logic found and moved (table above). What remains is flags, scenario from flags, registry, output and failure documents, and one-to-three facade calls per command. `adapterVerdict` decides when `adapter verify` fails from `adapter.VerifyResult`; kept as output shaping, noted as the one borderline residue |
| `cli/internal/serve/serve.go` | Misplaced (protocol and socket serving); moved to `mcp/transport` and `device/uds` |
| `model/{effector,model,scenario,validation}.go` | Right place: shared records (`EffectorAcknowledgement`, `Scenario`, `ValidationError`, `Issue`, `Contract`) and the contract validators for documents that cross the run boundary |
| `run/api.go`, `operations.go`, `internal/app/{advance,command_executor,initialize,replay,replay_commands,replay_execute,run,scenario,scenario_export}.go`, `internal/domain/{rules,scenario,scenario_walk,span}.go` | Right place and layer: orchestration in `app`, pure rules in `domain`. `scenario_export.go` gained `ExportScenario`, `inputs.go` is new |
| `world/{api,operations,service}.go`, `internal/domain/{effector_policy,effectors,faults,invocation_refusal,registry,world}.go` | Right place: refusal and admission rules belong to the world |
| `perturb/{operations}.go`, `internal/domain/{parameters,perturb}.go` | Right place: parameter and window checks belong to the perturbation layer |
| `adapter/{api,operations}.go`, `internal/domain/{adapter,lookup,validation}.go` | Right place; `LoadDir` added to the `files` edge |
| `domain/api.go`, `internal/domain/{catalog,compilation,dynamics_graph,loading,validation,validation_records}.go` | Right place; `LoadCatalog` added to the `files` edge |
| `mcp/internal/app/operator.go` | Right place (acknowledgement allow-list from P2b) |
| `refconsumer/{operations}.go`, `internal/domain/refconsumer.go`, `internal/mcpclient/mcpclient.go` | Right place; orchestration added in the new `app` layer |
| `schemas/schemas.go`, `testsupport/testsupport.go` | Right place: embedded contract sources; shared test fixtures |

## The CLI gate

`test/architecture/cli_wiring_test.go` (`TestCLIHoldsOnlyWiring`), with its proofs in
`cli_wiring_rules_test.go`:

| Rule | Proof it fails |
|---|---|
| Only the facade, `app`, `files` and `process` exist under `internal/cli`; any other package, nested ones included, fails | Synthetic: `app/logic/logic.go`, `internal/serve/serve.go`. Real tree: an injected `app/logic/logic.go` failed with "the CLI has no package internal/cli/internal/app/logic" |
| Every `app` file is on `cliAppFiles` with one wiring role | Synthetic: a new `ledger.go` |
| Each `app` file declares the module facades it may import (only `score.go` imports `grading`, only `suite.go` imports `suite`, only `manifest.go` imports `release`, and so on); a module's private layer or an unwired module never passes; stdlib is a short list; `encoding/json` only in `cli.go` | Synthetic: JSON decoding, `crypto/ed25519`, `os`, a private layer, `grading` in `catalog.go`, `truth`. Real tree: an added `run` import in `score.go` failed |
| Each `app` file has its own line cap (current size plus headroom) | Synthetic: 80 lines in `score.go`. Real tree: 40 added lines in `score.go` failed |
| No `go`, `select`, `defer` or channel; only `Build`, `Main`, `Run` exported | Synthetic: a goroutine in `mcp.go`, an exported `GradeRun` |
| The CLI's own edges are wiring: `files` imports `fmt` and `os` only (40 lines), `process` `io`, `os` and `time` only (35 lines); no module | Synthetic: a module import, `net/http`, growth past the cap. Real tree: an injected `net/http` import in `process.go` failed |

All injections into the real tree were reverted; the gate passes on the committed shape.

## Gates

| Gate | Result |
|---|---|
| `make phase-gate` (function-length, fmt-check, docs-check, tidy, build, vet, `lint-ci`, `test-short`, `test-simdet`, `coverage-check`, deadcode, vulncheck, fuzz-soak, Enola check, behaviour pin, `git diff --check`) | Passed |
| `go test ./...` | Passed |
| `make function-length` | Passed |
| `make behaviour-pin-check` | Matches the committed pin |
| `make coverage-check` | Every package at or above 70 %; the new `files` edges have their own tests |
| Enola check (`--fail-on=cycles,layers`, min confidence 0.8) | Passed with a partial verdict: the pinned baseline lacks the TypeScript producer, so only `go`, `manifests` and `mdintent` facts were graded. New advisory findings below the failure policy: dependency depth of `grading` (13), `grading/internal/app` (12), `grading/internal/domain` (11), `deviceworld` (11); resolved: `cli/internal/serve` depth 14 and others |
| Architecture suite | Passed: package classification, layer order, allow-list with no stale edge, module shapes, facade rules, I/O edge inventory, file size, function length |

Review: [reviews/P2c.md](../reviews/P2c.md). The gate table above was re-run after the review fixes
(`make phase-gate` exit 0, `go test ./...` passing, behaviour pin matching).

## Not done

- The Enola baseline was not re-pinned (the instruction is to pin once before the round and not to
  hide findings); the partial verdict is inherited from the existing baseline.
- The depth advisories for `grading` follow from its composing `run` and `score`; no change was made.
- R2c-01 (duplicated `files` wrappers) and R2c-02 (modules emitting the `streamsim: ` prefix) are
  recorded in FINDINGS.md and not fixed here.
