# H0 review B (adversarial)

Repo: /Users/ghassan/my-projects/agentic-stream, branch `fan-h0`, `git diff 91c2079d..fan-h0`.
Focused `go test` of runtime, device, spec, architecture, replay passes on the branch. Probes were run with `go test -overlay` (nothing written to the repo); the probe sources are in the session scratchpad.

Verdict: no High. Seven Medium findings, the main ones being (M1) the shipped examples silently lost their baseline choices and the pin was rewritten to bless it, and (M2) the event-schema directory accepts definitions that `validate` passes and ingest later rejects.

## Medium

### M1. The shipped example specs lost their baseline choices; the pin was regenerated to accept it
- internal/replay/testdata/baseline/example-choices.json (changed in 5fc2f120), examples/*/*.situation.yaml (no `baseline:` anywhere), internal/replay/internal/domain/baseline.go:`fillBaselineEnums`.
- Before H0.3 the thermal examples chose `mode=bounded_cooling` and `state=alert` in `over_ceiling`. After it, none of `examples/**` declares `baseline.byPhase`, so every enum falls to `enum[0]`: `mode=hold`, and `set_indicator state=off` in `over_ceiling` (and `off` in `cooling`). Three example specs changed (thermal-chamber, real-world-sensor sim and bench). The only spec with `byPhase` is the test fixture `internal/replay/testdata/baseline/thermal-mapping.situation.yaml`.
- Effect: the rule baseline that shadow replay compares a candidate against now recommends "hold / off" while over the ceiling. The instrument is quietly less truthful. The round-0 pin existed to catch exactly this, and it was updated in the same commit instead of the examples being migrated.
- Related: nothing requires `byPhase` to cover every phase or every enum field of the intent. A phase or field left out silently takes `enum[0]`, with no compile warning (see the "warm" case in baseline_test.go: `speed` falls to `slow`).
- Fix: add `baseline.byPhase` to the three examples so the pin returns to the round-0 values (the pin diff should be empty). Add a compile-time rule or warning: an intent with a `baseline` must give a value for every enum parameter in every phase it names, or say explicitly which fields default to the first value.

### M2. Event-schema directory: field definitions are not validated, so `validate` passes and runtime ingest breaks
- internal/spec/internal/domain/event_schemas.go:103 (`registryEntry.validate` checks only event_type, schema_version and the ref name).
- Reproduced: a file `{"x.y/1":{"event_type":"x.y","schema_version":"1","fields":{"v":{"path":"v","type":"banana"}}}}` loads without error. `registerInputSchemas` then stores `{"type":"banana"}`; compiling that JSON Schema fails ("value must be one of array, boolean, ..."), so every event of that type fails validation at ingest, not at `validate` or at start. `path` (for example `../../etc`), `unit`, `enum` element types and an empty `type` are also unchecked. An empty `type` silently means number (`schemaProperty`).
- Fix: in `Merge`/`validate`, check each field's `type` is one of the JSON Schema types the registry supports, that `enum` values match `type`, that `path` is a plain field name, and compile the generated schema (`EventSchemaJSON` then `canonicaljson.CompileSchemaJSON`) at load time so a bad definition fails `--event-schemas` and `validate`.

### M3. The event-schema set is not bound to anything durable; live-versus-replay parity is not provable
- cmd/agentic-stream/event_schemas.go:36 (digest only goes to `slog.Info`); `validate` does not log it; the spec digest does not cover it; the run report and the DB do not record it.
- A live run and its replay can use different directories (or one with, one without). A missing schema fails loudly at compile, but two different directories may define the same new ref differently. The only guard is `checkRegisteredDigest` ("conflicting bytes") when the same DB is reused; a fresh replay DB has no such guard, so a replay can validate events against a different definition than the live run did, with no record of it.
- Related: `LoadDeployment` (internal/spec/internal/store/deployments.go:120) round-trips through JSON and drops the unexported `eventSchemas`, so a deployment-loaded spec resolves only built-ins. `explain` uses that path today. For a custom schema, `inputRequiresBootIdentity` (internal/operators/internal/domain/boot.go:18) would silently return false (`!ok`), switching off per-boot state separation. Latent, but silent.
- Fix: persist the registry digest with the deployment (store it beside the spec digest) and print it in the run report and in replay output; refuse a replay whose digest differs from the recorded one. Make `EventSchema` return an error, not `false`, for a ref present in the compiled inputs but missing from the registry.

### M4. Dispatch and the start check disagree about who owns a route, and the new check is the only guard in the emulator profile
- internal/runtime/internal/app/effect_routing.go:45-56 versus route_ownership.go:22-30 and the message at :45.
- `RequireOwners` says "with a device present the fallback owns no route", but `route()` still sends a route the catalog does not name to the fallback. Under `--effect-profile emulator` the fallback is the simulated effector (cmd/agentic-stream/effect_profile.go:180), which accepts anything. So any command whose route is not in the catalog is a simulated success, which is the `silent_no_effect` shape. Spec intents are covered by the start check; nothing else is. The physical profile is safe because the fallback is fail-closed.
- Regression in defence in depth: before, `set_indicator` and `select_thermal_mode` with no device configured were refused even when the fallback was simulated. Now they are simulated successes. Intended for the simulated profile, but any wiring slip that drops `GatewayEffector` under emulator now degrades silently (`TestWithoutADeviceEveryNonWatchRouteIsASimulatedEffect` pins this).
- Fix: make `route()` use the same rule as `owns()` (with a device present, a non-watch route the device does not name goes to the fallback only if the fallback declares it owns the route, and the emulator fallback should say so explicitly), or fail the dispatch when a device is present and the route is unnamed. Add a dispatch-time test that asserts it for the emulator profile.

### M5. Tests that would still pass with the code broken
- cmd/agentic-stream/event_schemas_test.go:74 covers `serve` and `run-live` only for a missing or malformed directory, which fails in `loadLoggedEventSchemas` before the schemas are used. Deleting `spec.WithEventSchemas(flags.eventSchemas)` in serve_source.go:27 or run_live.go:86 (so a live run ignores the directory) fails no test. Only `run` and `validate` have a happy path.
- internal/spec/internal/app/event_schemas_test.go has no case for a bad field type (M2), a symlinked file, a non-regular file, an empty or `null` document, a duplicate key, or an upper-case extension.
- internal/runtime/internal/composition/intent_owners_test.go never builds a pipeline with a device present (its three cases use the fail-closed or simulated fallback with no `WithDevice`). The with-device strictness is proved only by the cmd test, and only for one removed route.
- Fix: add a `serve`/`run-live` test with a valid directory and a spec that needs it; add the loader cases above; add a composition case with `WithDevice` and a catalog that omits one intent.

### M6. The domain-word gate has large holes in what it reads and what it matches
- internal/architecture/domain_knowledge_test.go:169 (`domainWordUses`), :234 (`specWords`).
- Only whole `*ast.BasicLit` strings are checked. `"set_" + "indicator"`, `"%s_indicator"`, a constant split across lines, runes and byte slices, identifiers such as `SetIndicator`, and `strings.Join` all pass.
- Phases, entity types, presets and enum values are matched exactly. `"... WHERE phase = 'over_ceiling'"`, `"state=alert"` or a log line holding a phase pass. Only intent types and event types are substring matches.
- The vocabulary leaves out the most domain-specific words: payload field names and units (`celsius`, `rpm`, `fan_tach`), operator, window, trigger, hypothesis and fact names, spec `metadata.name`, labels, and device catalog route and target names (`set_indicator` is caught only because it is also an intent). `grep -in "thermal|celsius|rpm"` over production Go is clean today, so this is a gap in the gate, not an existing leak.
- The source is only `examples/*/*.situation.yaml` (not `*.yml`, not deeper, not the testdata specs). Deleting an example shrinks the vocabulary silently; there is no floor.
- Generated-file skip (`file.generated`, repository_test.go:208) trusts any file whose first lines are `// Code generated ... DO NOT EDIT.`, so a hand-written file can opt out. `domainKnowledgeExemptPaths` is a prefix match on two paths (fine today, but a new file under them is exempt without review).
- The domain data embedded in the binary (`internal/spec/internal/domain/event_schema_data.json`: motor, pump, zone, bay event types and fields) is not Go and not scanned.
- Fix: also scan identifiers and concatenated constants (evaluate constant expressions with `go/types` or walk `BinaryExpr` of literals), match phases, entities, presets and enum values as whole words inside literals (split on non-identifier characters), add the field, unit, operator and name words, glob `examples/**`, require a minimum number of vocabulary words, and drop the generated skip for files in this module that are not in a short allow-list.

### M7. `baseline.byPhase` is checked against the enum but not against the rest of the intent's authority
- internal/spec/internal/domain/baseline.go:`baselineChoiceProblem`.
- A choice passes if the value is in the parameter's `enum`. The compiler does not look at `modelWritableFields` or `presets.default`. The decision validator rejects a non-writable parameter whose value differs from the default preset (`verifyPresetEquality`, internal/decisions/internal/domain/intent_parameters.go:81), so a baseline choice for such a field is a decision the validator refuses at replay time, not at compile time. Not reproduced end to end; the argument is from the validator code.
- Also not checked: parameter constraints other than `properties.*.enum` (for example `allOf` or `oneOf` in `parameterSchema`).
- Fix: at compile, run the intent's own parameter validation (and the preset-equality rule) over the full parameter set a baseline phase would produce.

## Low

### L1. Directory loading: symlinks, special files, size and count are unbounded
- internal/spec/internal/app/event_schemas.go:29-46. `entry.IsDir()` is false for a symlink, so `link.json` pointing outside the directory (or at a FIFO or `/dev/zero`) is read with `os.ReadFile`: the symlink is followed (probe: target loaded), a FIFO blocks, and there is no size or file-count cap. `.JSON` and `.json.bak` are ignored silently (probe: `UPPER.JSON` not loaded). The directory comes from an operator flag, so this is hardening, not an attack path.
- Fix: `os.OpenRoot(dir)`, regular files only (`entry.Type().IsRegular()`), a size cap (for example 1 MiB per file, 64 files), and an error for a `.json`-like name with the wrong case.

### L2. A duplicate key inside one schema file silently takes the last definition
- internal/spec/internal/domain/event_schemas.go:78 (`decodeEventSchemaDocument` decodes into a Go map). Probe: two `"x.y/1"` entries in one file, no error, the second wins. Merging with the built-ins still catches a differing built-in ref, so the "never change a built-in" invariant holds; only new-ref ambiguity is affected. Also, `null` and `{}` files load without error (a directory holding only such files passes the "holds no *.json" check).
- Fix: decode with a token-level duplicate-key check; refuse an empty document.

### L3. `fallbackOwns` is fail-open for effectors that do not implement `RouteOwner`
- internal/runtime/internal/app/route_ownership.go:32-38. A wrapper around the fail-closed effector (telemetry, logging) that does not forward `OwnsRoute` makes the start check treat the fallback as owning every route. Today nothing wraps it. A typed-nil fallback also passes `fallback == nil`.
- Fix: make `RouteOwner` part of the `Effector` contract used by the composite, or treat a fallback without `OwnsRoute` as owning none and say so.

### L4. A catalog route named `install_watch_condition`
- effect_routing.go:46 and route_ownership.go:23 always send it to the watch effector, whatever the catalog says. Safe (the device can never receive it, as the new test shows), but nothing tells the catalog author. The device catalog loader could reject that route name.

### L5. `--event-schemas` is a persistent root flag that the other commands ignore
- cmd/agentic-stream/main.go (`registerEventSchemasFlag(root)`): `explain`, `intent`, `quarantine`, `operator` and the rest accept and ignore it. Either reject it on commands that do not use it or document it.

### L6. The public `device.GatewayEffector.OwnsRoute` panics on a nil receiver
- internal/device/effectors.go:51-55 (`e.effector` dereferences `e`). The app-level method is nil-safe and the composition guards the pointer, but the exported wrapper differs from the one under it.

## Checked and found sound
- Route names are exact string matches taken from `Intent.Type` on both sides; no normalisation gap between the spec and the catalog (`policy/internal/domain/commands.go:16` sets `effector_route` to the intent type).
- The watch route cannot be claimed by a device, in dispatch or in verification.
- Policy: the start check uses the same `!= active` test as the episode runner (`runner_decision.go:77`); shadow never writes intents or commands; the spec is fixed for the life of a pipeline, so an intent cannot be added after the start. Compensation intents must be declared in the same catalog (`checkIntentAuthority`), so they are covered.
- The built-in registry cannot be overridden or shadowed: a ref in both must be identical (probe: differing built-in refused). Ingest validates from the DB table, and a second differing definition for the same ref in one DB is refused ("conflicting bytes").
- `byPhase` values must be strings that appear in the declared enum, phases must be declared, and the intent catalog sent to the executor does not carry the baseline.
