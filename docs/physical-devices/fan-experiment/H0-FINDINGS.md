# H0 findings: remove hardcoded domain knowledge from Agentic Stream

Investigated 2026-10-20 in `/Users/ghassan/my-projects/agentic-stream` (branch `real-run`, HEAD `91c2079d`; the only dirty path is the
untracked `documentation/visual-guide/`, not ours). Read-only. Plan: `PLAN.md` section 5.1. Paths below are relative to the
Agentic Stream repository. "UNVERIFIED" marks what I read but did not run.

## 0. Repository facts that bind every round

- **Gates.** `make ci-check` = `proto-check` + `tidy build vet lint-ci coverage-check deadcode vulncheck docs-check`.
  `coverage-check` runs `go test -short -race -count=1 -shuffle=on -cover ./...` with a 35 s suite / 15 s package / 5 s test budget and a
  70 % per-package floor. `go test -short -count=1 ./...` (no race) took **17 s** here, all green. Full `ci-check` time: UNVERIFIED
  (lint and vulncheck not run). `make tidy` may touch `go.mod`/`go.sum`.
- **protoc.** `protoc --version` = `libprotoc 35.1`, the Makefile pin is `PROTOC_VERSION ?= 35.1`. `make proto-check` passes in 2 s. The
  36.0 problem is gone on this host.
- **Branches.** No naming rule is documented (AGENTS.md, `.agents/context/*`). Existing names are kebab-case topics (`review-fixes`,
  `code-improvements-4`, `e2e-round-6`, `docs/video-walkthrough`); work lands by pull request ("Merge pull request #61"). Suggest
  `h0-no-hardcoded-domain` (one branch, one commit per round).
- **Code rules** (AGENTS.md, `.agents/context/quality-bar.md`): files at most 300 lines, production functions at most 15 body lines,
  no comments inside a module (`internal/<m>/internal/**`, `cmd/`), doc comment on every exported facade symbol, wrap errors with `%w`,
  never loosen a gate, `deadcode` fails on production code reachable only from tests (every new function needs a real caller),
  comment/refactor with `gopls`/`go/ast` not `sed`. `docs/` is git-ignored and code may not read it
  (`TestNothingExecutableReadsFromTheDocsArchive`), so `docs/reviews/.../02-architecture.md` (AR-13) is a local file only.
- **Gate tables** (`internal/architecture/README.md`): `allowedImports` (`import_rules_test.go`), `forbiddenImports`, `packageLayers`,
  `facadeSpecs` (`facades_test.go`). An allow-list may shrink; a new edge is a reviewed table edit in the same change;
  `TestAllowedImportsHaveNoStaleEdges` fails if an approved edge is unused, so removing the last use of an import means removing the edge.
  Cross-module gates live only in `internal/architecture` (one file per subject, repo parsed once by `loadRepository`; model:
  `timetext_test.go`). `internal/actions` is forbidden to import `internal/device` and `internal/watch`.

## H0.1 A device route is a route the catalog names

**Current code.**
- `internal/runtime/internal/domain/lifecycle.go:36-44` `ClassifyRoute`: `install_watch_condition` -> `WatchRoute`;
  `set_indicator`, `select_thermal_mode` -> `DeviceRoute`; else `FallbackRoute`.
- `internal/runtime/internal/app/effect_routing.go`: `CompositeEffector{watch, serial, fallback}`; `route()` (lines 49-64),
  `Dispatch`, `DispatchAuthorized`, `VerifyDeviceCommand` (lines 107-120). Built only at `internal/runtime/internal/composition/planes.go:52-55`
  (`NewCompositeEffector(watch, cfg.Effector)` then `.WithSerial(gatewayEffector)` if `cfg.GatewayEffector != nil`).
- Callers of `ClassifyRoute`: `effect_routing.go:53,111` and `lifecycle_test.go:65-69`. Callers of the composite: `planes.go` and
  `effect_routing_test.go` (about 15 sites: lines 58, 75, 80, 94-97, 113, 120, 149, 163, 180-183).
- The route is the intent type: `internal/policy/internal/domain/commands.go:16` sets `"effector_route": p.Row.IntentType`.
- Where the effectors come from (`cmd/agentic-stream/effect_profile.go`): `simulated` profile -> `device.NewSimulatedEffector()`, no gateway;
  `emulator`/`physical` -> `device.NewGatewayEffector` over the UDS gateway link plus a fallback from `fallbackEffector`
  (`physical` -> `FailClosedEffector`, `emulator` -> `SimulatedEffector`). Other effectors: the watch effector (`*watch.Service`),
  nothing else in production.
- The catalog is already queryable: `device.CapabilityCatalog = domain.CapabilityCatalog` (`internal/device/catalog.go:7`) with exported
  `Routes map[string]Route` (`internal/device/internal/domain/catalog.go`), held by `app.GatewayEffector.catalog`
  (`internal/device/internal/app/gateway_effector.go:16-20`). The facade `device.GatewayEffector` (`internal/device/effectors.go:30`) does not
  expose it. The route name is also what `Materialize` looks up, so "catalog names the route" is exactly the existing closed-set rule.
- `install_watch_condition` is `watch.InstallRoute` in `internal/watch/internal/domain/condition.go:14`, **not** exported from the
  `internal/watch` facade; `internal/decisions/internal/domain/binding.go:91` has a second literal.

**Constraints.** `runtime/internal/app` may import `actionport`, `watch`, not `device` (`import_rules_test.go:95`); `composition` may import
`device` and `watch`. `actionport` imports nothing (pure contract; `TestContractPackagesExcludePersistenceAndTransport`), so a generic router
may live in `actionport/internal/domain` but cannot name `watch` or `device`.

**Proposed change.**
1. `actionport`: add `RouteOwner interface{ OwnsRoute(route string) bool }` (domain + alias in `actionport.go`).
2. `internal/device`: `GatewayEffector.OwnsRoute(route)` (one-line delegation to `app.GatewayEffector.OwnsRoute`, which does
   `_, ok := e.catalog.Routes[route]`). `SimulatedEffector.OwnsRoute` returns true; `FailClosedEffector.OwnsRoute` returns false (used by H0.2).
3. `internal/watch`: export `const InstallRoute = domain.InstallRoute` in the facade (`watch.go`; also lets `decisions` stop duplicating it only
   if its import rule allows `watch`, which I did not check: leave `binding.go:91` as an allow-listed built-in in H0.4).
4. Replace `ClassifyRoute`/`RouteKind`: `CompositeEffector` keeps its constructor and `WithSerial` (rename `WithDevice`), and `route()` becomes
   `route == watch.InstallRoute -> watch` (still "watch effector is not configured" when nil), `device != nil && device.OwnsRoute(route) -> device`,
   else fallback. `VerifyDeviceCommand` uses the same decision. Delete `ClassifyRoute`, `RouteKind` and its test.
5. Optional AR-13 round: move the router to `actionport/internal/domain/router.go` as a generic `Router{fallback; owners}` with
   `Own(effector, routes...)`, forwarding `Dispatch`, `DispatchAuthorized`, `DeviceStateVerifier`; composition registers
   `watch.InstallRoute -> watch` and the catalog's route names -> gateway. Then drop the `runtime/internal/app -> internal/watch` edge only if
   nothing else uses it (`Pipeline` does, so it stays).

**Behaviour change to decide (flag to owner).** Today in the `simulated` profile (no gateway) `set_indicator`/`select_thermal_mode` are
**refused** ("serial effector is not configured", `effect_routing_test.go:95-96,107-122`). After H0.1 there is no catalog to name them, so they go to
the simulated fallback and are accepted as simulated effects. In `physical` they are refused by the fail-closed fallback, in `emulator` they reach
the gateway if the catalog names them (unchanged). I recommend accepting this (a simulated profile touches no device) and recording it; the
alternative (keep a "known device route" list) is the hardcoding being removed.

**Tests to update/add.** Rewrite `effect_routing_test.go` cases: "no device effector for set_indicator" and `TestDeviceRoutesNeverReachTheFallback`
become "a route the catalog names reaches the device, a route it does not name goes to the fallback"; add a new route name (e.g. `vent_open`) owned
by a fake `RouteOwner` to prove no code edit is needed; fail-closed physical fallback refuses an unnamed route (exists at line 80). Add
`OwnsRoute` tests in `internal/device` (catalog with two routes; `device_peer_test.go`/`fixtures_test.go` already load the thermal catalog).
End-to-end: `cmd/agentic-stream/experiment_e2e_test.go`, `experiment_live_feed_test.go` must pass unchanged (they use `emulator` + the thermal
catalog `internal/contractsv1/internal/domain/conformance/v1/thermal-capability-catalog.json`).
**Digests:** none. **Risk:** low; semantic change above only.

## H0.2 `serve` refuses to start when a spec intent has no owner

**Where it fits.** `serve` -> `serveRuntime` -> `core.openEffects` (gateway handshake, `live.go:113`) -> `startContinuousPipeline` ->
`openContinuousPipeline` (`serve_source.go:26`, compiles the spec) -> `core.startPipeline` (`live.go:143`) -> `runtime.NewPipeline` ->
`composition.NewPipeline` (`composition/pipeline.go:53`) -> `composeEffectors` (`planes.go:41`). `run-live` follows the same path. `composition` is the
only place that holds the spec (`cfg.Spec`), the watch service, the gateway effector and the fallback together, and `cmd` must stay wiring. Put
the check in `composeEffectors`'s caller (`NewPipeline`) right after the composite is built; an error there surfaces from `serve` before the HTTP
listener starts (cleanups close the gateway claims).

**Proposed change.** `CompositeEffector.Unowned(routes []string) []string` (a route is owned if it is the watch route, the gateway owns it, or the
fallback owns it; a fallback that does not implement `RouteOwner` owns everything, so custom/test effectors and the simulated one are unchanged).
`composition` calls it with `cfg.Spec.Actions.Intents[].Type` and returns
`fmt.Errorf("spec intent %q has no effector: not the watch route, not named by the device catalog, and the %s fallback refuses it", ...)`.
**Only when `spec.EffectiveDispatchPolicy(cfg.Spec.Cognition.Executor.DispatchPolicy) == spec.DispatchActive`**: a shadow-dispatch spec never
enters action governance (`internal/spec/spec.go:98-101`), so its intents cannot produce commands.

**Effectiveness caveat.** Because the simulated fallback owns everything, the check only bites in `physical` (fail-closed fallback) and when a
custom fallback declines routes. In `emulator` a misnamed route would still be simulated silently, as today. Stricter variant for the owner to
choose: in `emulator`/`physical` make the fallback own nothing, so every intent must be watch or catalog (needs `Owns`-false for the
emulator's simulated fallback, i.e. a `fallbackEffector` change in `effect_profile.go`; not needed for the fan run, which is `physical`).

**Tests.** `composition/pipeline_test.go`: thermal-sim spec + `device.NewFailClosedEffector(physical)` + no gateway -> error names `set_indicator`
(first in declaration order) ; shadow-dispatch spec -> no error; simulated fallback -> no error. `cmd` level (stand-ins in
`experiment_standins_test.go`): `serve --effect-profile emulator` with a catalog copy lacking `select_thermal_mode` still starts (documents the
caveat), and, if the strict variant is chosen, refuses. The `cmd` contract test `TestExperimentCommandsKeepTheirFlags` is unaffected (no flag change).
**Digests:** none. **Gate edits:** none expected (composition already imports `spec`, `device`, `watch`).

## H0.3 Replay baseline reads its mapping from the spec

**Current code.** `internal/replay/internal/domain/baseline.go:205-224` (`baselineEnumValue`, `enumValue`) hardcode `state`/`mode`,
`over_ceiling`/`cooling`, `alert`/`watch`/`bounded_cooling`/`hold`. Only caller: `fillBaselineEnums` <- `baselineParameters` <- `selectIntent`.
The spec reaches it via `internal/replay/internal/app/baseline.go` (`baselineIntents` projects `spec.Intent` into `domain.Intent{Type, Risk, Schema}`)
from `NewDeterministicBaseline` (`app/shadow_worker.go:25`, used by `run --worker-socket`). Tests: `baseline_test.go:71-107`
(`TestBaselinePolicyMapsPhaseToDeclaredStateAndModeEnums`, which must be rewritten), `cmd/.../experiment_shadow_test.go` (asserts only that
comparisons exist and no effect rows).

**Surprise: the mapping is dead on every shipped spec.** `selectIntent` returns the **first** declared intent whose parameters can be satisfied.
In all three thermal specs the first intent is `install_watch_condition` (entity_id only, no required fields, no enum), so the baseline always
picks it and never reaches `set_indicator`/`select_thermal_mode`. No existing test pins the baseline decision bytes for any example spec
(UNVERIFIED by execution; read from code). So "same choices as today" holds trivially if the examples are not changed.

**Spec schema/Go shape.** `internal/spec/internal/domain/schema.json` `$defs/intent` has `additionalProperties: false`; `Intent` is in
`spec.go:214-224`. Proposed optional field on an intent, smallest and local to the intent that owns the parameter:
```yaml
- type: set_indicator
  baseline:
    byPhase:
      over_ceiling: {state: alert}
      cooling:      {state: watch}
```
Go: `Intent.Baseline *BaselineChoice \`json:"baseline,omitempty" yaml:"baseline,omitempty"\``, `BaselineChoice{ByPhase map[string]map[string]any}`.
Compiler checks (in `references.go`, next to the other intent checks): each phase is declared in `situation.phases`; each field is in
`parameterSchema.properties` and has an enum containing the value. Replay: `domain.Intent` gains `ByPhase`; `fillBaselineEnums` takes
`enum[0]` and overrides with `ByPhase[phase][field]` when present and in the enum; `baselineEnumValue`/`enumValue` shrink to that.
Not in the intent catalog: `compileIntentEntry`/`intentCatalogEntry` (`internal/episodes/internal/domain/intent_catalog.go:136`) copy fields one by
one, so the cross-repo catalog digest `e4f86620...` (`intent_catalog_parity_test.go`) is **unchanged**.

**Pinned digests.**
| Change | Effect |
|---|---|
| Schema gains optional field, Go field is `omitempty`, example specs untouched | **No pin changes** (spec digest is `canonicaljson.Seal(DomainSpec, spec)` of the compiled struct, `compiler.go:53`) |
| Example thermal specs gain `baseline:` blocks | Changes 3 spec digests in `internal/spec/experiment_contract_test.go:16-18` (`d5907b52...`, `a6153efe...`, `b1b60be7...`) and 3 policy digests in `internal/policy/experiment_contract_test.go:20-22` (`e7e08b9f...`, `7a7f4506...`, `d973fc8a...`). Those policy digests are the bench gateway `--device-policy-digest` allow-lists, so every deployed gateway must be updated. A search of streams-simulator, agentic-stream, agent-research-lab and tamoz found these digests only in those two test files |
Recommendation: do **not** add blocks to the existing thermal examples (breaks "P0.5 re-run passes unchanged" for no behaviour gain); prove the
mapping with a test spec under `internal/replay/testdata` and add the block to the new fan-experiment spec, whose digest is pinned fresh.
Also `cmd/agentic-stream/experiment_contract_test.go` is flags only (no digests).

**Tests.** Add first (characterization, before any change): baseline decision digest for each example spec at a fixed snapshot. Then:
rewrite `baseline_test.go:71` to supply the mapping as data; compiler tests for the three rejection cases (unknown phase, unknown field, value not in
enum); a spec round-trip test that an intent without `baseline` compiles to identical canonical JSON. Docs: `documentation/contracts/situation-spec.md`.
`make docs-check` (UNVERIFIED whether it lists schema fields). **Risk:** S; the schema file is the contract, no other copy.

## H0.4 Guard against hardcoded domain words

**Home and shape.** New `internal/architecture/domain_knowledge_test.go` (about 120 lines, parallel, uses `loadRepository`/`repo.production` and
`gopkg.in/yaml.v3`, already a direct dependency). It builds a vocabulary from every `examples/**/*.situation.yaml` and fails on any production
Go string literal that matches. Add a row in `internal/architecture/README.md`. Test-only change: no module, no import-table edit
(`internal/architecture` allows `{}`; test files are not covered by `allowedImports`).

**Vocabulary (from `actions.intents[].type`, `inputs[].eventType`/`entityType`, `situation.type`, `situation.phases[].name`, intent preset names,
every enum value under `parameterSchema.properties`).** Match intent and event types with substring, phases/enums/presets with exact equality (they
are common words). I ran this scan (python, read-only); production hits today:

| Literal | Kind | Where | Verdict |
|---|---|---|---|
| `set_indicator`, `select_thermal_mode`, `alert`, `watch`, `bounded_cooling`, `hold`, `over_ceiling`, `cooling` | domain | `runtime/internal/domain/lifecycle.go:40`, `replay/internal/domain/baseline.go:218,221` | removed by H0.1 and H0.3 |
| `install_watch_condition` | Agentic Stream built-in | `watch/internal/domain/condition.go:14`, `decisions/internal/domain/binding.go:91`, (`lifecycle.go:38` goes in H0.1) | allow-list, name the two files |
| `resolved` | built-in situation phase | `situations/internal/domain/evaluate.go:60` (`PhaseResolved`), `policy/internal/domain/routing.go:72`, `policy/internal/app/approval.go:36` (approval disposition, unrelated word) | allow-list with reasons |
| `default` | built-in preset name | `decisions/internal/domain/intent_parameters.go:74`, `contractsv1/internal/domain/declarations.go:4` | allow-list |
| `create_maintenance_ticket` | domain | `internal/executor/fixture/executor.go:80,133`, `internal/testsupport/executorconformance/conformance.go:43` | exempt the two fixture/test-support paths, as the plan already says (note `composition/planes.go` imports `executor/fixture` for demo mode) |
| `native`, `tamoz` | executor names | `effect_profile.go:219`, `composition/worker.go:26`, `run_command.go:38` (flag default), `replay/internal/domain/shadow_comparison.go:82` | exclude executor names from the vocabulary; see section 7 |

Embedded JSON data (`event_schema_data.json`, `simulator_data.json`) is data and not scanned. The test lists built-ins as an explicit
`map[literal][]file` with a reason; a stale allow-list entry fails (mirror `TestAllowedImportsHaveNoStaleEdges`). Add a self-test that a synthetic
file containing `"set_indicator"` is caught. Land it **after** H0.1 and H0.3 so it is green with only built-in allow-list entries.

## H0.5 Event schemas from a directory

**Current use of the registry** (`internal/spec/internal/domain/event_schema.go`: `//go:embed event_schema_data.json`, `builtins =
sync.OnceValues(loadRegistry)`, global `LookupEventSchema`, `EventSchemaJSON`; facade `spec.LookupEventSchema`, `spec.go:43`):
- compile: `references.go:110` (`checkInputSchema`) and `:172` (`checkOperatorField`);
- operators at runtime: `internal/operators/internal/domain/boot.go:19` (device `quality`/`boot_id` fields);
- deployment: `internal/spec/internal/store/deployments.go:91` (`registerInputSchemas`) writes the schema JSON to table `event_schemas`
  (`RegisterEventSchema`: immutable per (event_type, version), a different byte string is refused with "conflicting bytes").
- **Ingress/eventlog admission does not read the registry**: it loads `schema_json` from the DB (`internal/eventlog/internal/store/store.go:28`),
  filled by the deployment step above. So a directory registry only has to reach compile, operators boot and `SaveDeployment`.
- Commands that compile a spec: `validate` (`main.go:82`), `serve` (`serve_source.go:27`), `run-live` (`run_live.go:81`), and `run`/replay
  (`replay/internal/app/session.go:33`, `recorded_source.go:14`, `shadow_worker.go:26`, via `replay.Request.SpecPath`). `explain` uses
  `spec.LoadDeployment` (stored `compiled_ir`), which never looks schemas up.
- Test callers of the global lookup: 9 sites (`quarantine_command_test.go:69`, `ingress/.../helpers_test.go:109`, `spec_test.go:79`,
  `event_schema_test.go:78`, `spectest_test.go:16`, `event_schema_store_test.go:21,99`, `eventlog/schema_validation_test.go:17`).
- Data pin: `internal/spec/experiment_contract_test.go` `zoneSchemasDigest` (digest of the `zone.*` entries of the embedded file).
- Second embedded domain registry, not in the plan: `internal/ingress/internal/domain/simulator_data.json` maps simulator channels to field names
  (vibration, ph, co2, ...; **no zone channels**). It is used only by `--trace-format simulator`; `serve --live-socket` requires `normalized`
  (`serve.go:77`), so the fan run is unaffected. Same fix pattern if ever needed.

**Proposed change (no process-global state).**
1. `spec.EventSchemas` value type with `spec.BuiltinEventSchemas()` (the embedded set) and `spec.LoadEventSchemas(dir)` (reads `*.json` files in the
   existing format, `map[ref]entry`; result = embedded set merged with the directory; a ref present in both must be identical or loading fails, so a
   directory can add schemas but never silently change a built-in one). Exposes `Digest()` (canonical JSON, `DomainEventSchemas`-style or reuse an
   existing domain tag: UNVERIFIED which tag is appropriate in `canonicaljson`).
2. `spec.CompileFile(ctx, path, opts ...CompileOption)` with `spec.WithEventSchemas(reg)` (variadic, so the roughly 30 test callers compile
   unchanged). `CompiledSpec` carries an unexported `schemas *EventSchemas` (skipped by `json`, so the digest is unchanged) and a method
   `(*CompiledSpec).EventSchema(ref)`; `operators/boot.go:19` and `store.registerInputSchemas` use it; the global `LookupEventSchema` stays as the
   embedded default for the 9 test sites. A `CompiledSpec` rebuilt by `LoadDeployment` has no registry and falls back to the embedded one.
   UNVERIFIED: any test that `reflect.DeepEqual`s a compiled and a loaded spec would see the new pointer; grep before coding.
3. CLI: one persistent root flag `--event-schemas <dir>` (`newRootCommand`, `main.go:34`) so `validate`, `serve`, `run-live`, `run` all take it;
   `replay.Request` gains `EventSchemaDir` and `replay/internal/app` passes `spec.WithEventSchemas`. `serve`, `run-live` and `run` log
   `event_schemas_digest` at start (slog). New flag is additive for `TestExperimentCommandsKeepTheirFlags`.
4. Determinism: the registry is an input to compile and, via the DB, to admission; the DB already refuses a different byte string for the same
   (event_type, version), which protects one database. A **fresh** replay database has no such memory, so `run` against a different directory
   could silently validate differently. The run artifact (`internal/runartifact/internal/domain/manifest.go`, `schema_version` 1) records
   `spec_digest`, `policy_digest`, firmware and capability digests but **not** the schema registry, and `event_schemas` rows are not exported
   (grep of `internal/runartifact`). Adding a manifest field is a versioned contract change: do not do it in H0.5; log the digest now and decide
   separately (owner). The spec digest does not cover it because it hashes the compiled spec, not the registry.
**Gate edits:** none for imports (`operators` and `store` already import `spec`/`spec/internal/domain`). `deadcode` needs the flag wired in the
same commit. **Digests:** none (embedded file untouched, so `zoneSchemasDigest` and spec digests are unchanged).
**Tests:** directory with a new `fan.speed.observed/1.0` compiles and registers; same ref with different bytes is refused; a missing or
malformed directory fails `serve` before listening; `validate --event-schemas` prints the same digests as the embedded path for existing specs.

## 7. Sweep: other domain knowledge the plan missed

- `tamoz` (a consumer name) in production Go: `cmd/agentic-stream/run_command.go:38` default `--worker-name tamoz`;
  `internal/replay/internal/domain/shadow_comparison.go:82` the key `"tamoz"` in the **sealed comparison document** (digest-bound; renaming it
  changes comparison digests, so not an H0 change); `TamozDecisionJSON` fields in replay; 39 case-insensitive mentions in total. Not a spec
  vocabulary item, so H0.4 excludes it, but it contradicts "no consumer name in the binary" spirit: record as backlog.
- `internal/ingress/internal/domain/simulator_event.go:105` builds `<entityType>.<channel>.observed` event types: a naming convention in code,
  not a domain word. `simulator_data.json`: see H0.5.
- No `zone-01`, `fan-01`, `.observed` event-type literal, or entity-name literal in non-test Go outside the above (regex sweep of non-test `.go`).
- `internal/runtime/internal/composition/pipeline.go:44` and `effect_routing.go:26` comments say "thermal"; remove with H0.1 (comments inside a module
  are forbidden anyway).

## Suggested rounds (one commit each, on one branch)

| # | Commit | Gate list (run in this order) |
|---|---|---|
| 0 | `test: pin the baseline decision of every example spec` (characterization, no production change) | `gofmt -l .`; `go test ./internal/replay/... ./internal/spec/...`; `go test ./internal/architecture/...` |
| 1 | H0.1: `RouteOwner`, `OwnsRoute` on gateway/simulated/fail-closed, `watch.InstallRoute` export, router uses ownership, delete `ClassifyRoute` | `gofmt -l .`; `go build ./...`; `go vet ./...`; `go test ./internal/runtime/... ./internal/device/... ./internal/actionport/... ./internal/watch/...`; `go test ./cmd/agentic-stream -run 'Experiment'`; `go test ./internal/architecture/...`; `make lint`; `make deadcode` |
| 1b | (optional) AR-13: generic `actionport` router | same as 1, plus `go test ./internal/architecture/...` for import tables |
| 2 | H0.2: start-time ownership check | as 1, plus `go test ./internal/runtime/internal/composition/...` |
| 3 | H0.3: `baseline` intent field, schema, compiler checks, replay mapping | `gofmt -l .`; `go build ./...`; `go vet ./...`; `go test ./internal/spec/... ./internal/replay/... ./internal/policy ./internal/episodes/...` (pins: catalog digest `e4f86620...` and the 6 experiment pins must stay green); `go test ./cmd/agentic-stream -run 'Experiment'`; `make lint`; `make docs-check` |
| 4 | H0.4: architecture guard (after 1 and 3) | `go test ./internal/architecture/...`; `make lint` |
| 5 | H0.5: `EventSchemas`, `CompileOption`, `--event-schemas` | `gofmt -l .`; `go build ./...`; `go vet ./...`; `go test ./internal/spec/... ./internal/operators/... ./internal/engine/... ./internal/replay/... ./cmd/agentic-stream/...`; `go test ./internal/architecture/...`; `make lint`; `make deadcode`; `make docs-check` |
| end | Before the pull request | `make ci-check` (once; includes `proto-check`, race, coverage floor, budgets) |

Sizing: rounds 0-2 and 4 are small; round 3 is small-to-medium (schema, compiler, replay, docs); round 5 is the largest (facade, compiled-spec
field, flag, replay request) and may be split into 5a (type, option, operators/store use, no flag) and 5b (flag, replay request, docs).

## Decisions for the owner

1. Simulated profile now accepts device-named routes as simulated effects (H0.1 behaviour change).
2. H0.2 strictness: only `physical` bites by default; make `emulator` strict too?
3. H0.3: leave existing thermal example specs untouched (recommended, no pin or gateway allow-list change) and use the mapping only in the new
   fan spec and a replay test spec.
4. H0.5: record the schema registry digest in the run-artifact manifest (contract version bump) or log only.
5. Do the AR-13 move now (round 1b) or later.
