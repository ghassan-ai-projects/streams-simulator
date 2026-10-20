# H0 review A (correctness and architecture)

Repository `/Users/ghassan/my-projects/agentic-stream`, branch `fan-h0`, `91c2079d..6a8ba9cb` (7 commits). Read-only review.
Bar: that repository's AGENTS.md and `internal/architecture`, plan section 5.1 and H0-FINDINGS.md (decisions 1 to 5).

## Verdict

No High findings. Each H0 item does what 5.1 says. The decided behaviour changes, and only those, are present. Four Medium findings
(a fail-open default in the new start check, an unvalidated registry directory, an untested safety path of H0.5, a business rule in
composition) should be fixed before the host module (A3) builds on this. The rest is Low.

Verified by running (all green, no repo edits): `go test -short -count=1` over architecture, spec, replay, device, actionport, watch,
operators, canonicaljson, runtime, cmd, eventlog, ingress, engine, policy; `go test -count=1 -run Experiment ./cmd/agentic-stream`
(non-short: `TestExperimentInterlockStopsEffects`, `...ClosedLoopThroughServe`, `...ClosedLoopUnderAContinuousFeed`,
`...ShadowReplayComparesTheCandidate`, `...CommandsKeepTheirFlags` all pass, unchanged); `scripts/check-deadcode.sh` clean;
`golangci-lint run` over runtime, spec, replay, device, architecture, cmd: 0 issues; `make docs-check` ok. Not run: the full
`make ci-check` (race, coverage floors, vulncheck).

## Per-item check against 5.1

| Item | Result |
|---|---|
| H0.1 routing by catalog ownership | Done. `ClassifyRoute`/`RouteKind` deleted; `route()` is watch constant, then device if `OwnsRoute`, else fallback (`internal/runtime/internal/app/effect_routing.go:45-58`). Gateway owns exactly its catalog routes (`catalog.go NamesRoute`), simulated owns all, fail-closed none. Decision 1 (simulated profile makes device routes simulated effects) implemented and tested (`TestWithoutADeviceEveryNonWatchRouteIsASimulatedEffect`). AR-13 not done, as decided (5). |
| H0.2 strict start check | Done and strict per decision 2: with a device present the fallback owns nothing (`route_ownership.go:22-30`); only for `DispatchActive` (the compiled spec normalises an empty policy to shadow, `normalize.go:46`, so the comparison is sound). Tested at three layers (app table, composition with real specs, cmd: refuses before listening). |
| H0.3 baseline from spec data | Done. No field, phase or value name left in `replay`. Examples untouched (`git diff --stat -- examples` empty); no spec, policy or catalog digest pin changed (spec/policy contract tests and `e4f86620` parity pass). The re-pin of `example-choices.json` is exactly what the commit says (single-intent entries only, set_indicator over_ceiling alert to off, cooling watch to off, select_thermal_mode over_ceiling bounded_cooling to hold, three specs each); whole-catalog entries unchanged. The old mapping is reproduced byte for byte from data by `baseline_mapping_test.go`. |
| H0.4 guard | Done, non-vacuous (self-test, stale allow-list entries fail, exempt paths must exist). |
| H0.5 registry from a directory | Done: immutable `EventSchemas`, `spec.WithEventSchemas`, root persistent `--event-schemas`, `event_schemas_digest` logged at start of serve, run-live and run, `replay.Request` carries the loaded registry (`compileRequestSpec`), no process-global state left, embedded set stays the default, `LoadDeployment` falls back to builtin and the docs say so. The digest is not recorded in the run artifact, as decided (4). |

## Gate changes (`git diff 91c2079d..fan-h0 -- internal/architecture scripts`)

1. `internal/architecture/import_rules_test.go:197`: `internal/watch` gains the edge `internal/watch/internal/domain` (an allow-list grows). Judged acceptable (facade to its own domain for one constant, the same shape 20-plus facades already have, named in the commit), but see L1: it was avoidable.
2. `internal/architecture/domain_knowledge_test.go` (new, 268 lines) and its README rows: a new gate, tightens only. The built-in allow-list is explicit, with reasons.
3. `internal/architecture/README.md`: documents the new gate.
4. `scripts/slow-tests.txt`: two `internal/runtime` entries removed (tightens the budget). Not related to H0; see L2.

No cap raised, no coverage floor lowered, no test hook added, nothing weakened.

## Findings

### Medium

**M1. The start check fails open for any fallback that does not declare routes, and the facade hides the declaration.**
`internal/runtime/internal/app/route_ownership.go:32-38` (`fallbackOwns`): a fallback that does not implement `actionport.RouteOwner` "owns every
route". `device.NewSimulatedEffector()` and `NewFailClosedEffector()` return `actionport.AuthorizedEffector`, which does not include `RouteOwner`
(`internal/device/effectors.go:82,89`), so ownership is found by a runtime type assertion. Anyone who wraps the fallback (telemetry, logging,
a decorator in `cmd/effect_profile.go:180 fallbackEffector`) silently turns `physical` fail-closed ownership into "owns everything": H0.2 stops
biting and nothing fails. Dispatch itself would still refuse, so no safety loss at dispatch, but the instrument reports a clean start it
should not. Fix: make the fallback type in `PipelineConfig.Effector` / the facade return types include `RouteOwner`
(`interface{ AuthorizedEffector; actionport.RouteOwner }`), and treat "does not implement it" as "owns nothing" (test effectors implement
it). Add a test with a wrapped fail-closed effector.

**M2. Directory schemas are only checked for name consistency, not for validity.**
`internal/spec/internal/domain/event_schemas.go:103-111` (`registryEntry.validate`) checks `event_type`, `schema_version` and the ref only.
Probe (go test with `-overlay`, repo untouched): a field with `"type": "bogus"` and an empty `path` merges without error and
`EventSchemaJSON` emits `{"a":{"type":"bogus"}}`; that string is what `registerInputSchemas` writes to `event_schemas` and what event
admission later compiles, so the failure shows up at admission time (or as quarantines), not at `serve` start, which the docs promise
("stops the command before it opens a database"). Fix: in `validate`, require `type` in {"", number, string, boolean, integer}, a non-empty
`path`, `enum` only on string fields, and compile the emitted JSON schema once at merge time. Add table cases to
`TestMergingRefusesAMalformedOrConflictingSchema`.

**M3. The safety-relevant consumer of the registry (boot identity) is not proven to use the directory registry.**
`internal/operators/internal/domain/boot.go:18` now reads `r.spec.EventSchema(...)`, but no operators test constructs a spec with a directory
registry (grep: no `EventSchemas` in `internal/operators`). If someone changes it back to a builtin lookup, nothing fails, and a custom device
schema with `quality` and `boot_id` silently loses the boot-identity requirement (`!ok` returns false). Also no test drives a custom
schema through `serve`/`run-live` end to end (the cmd tests cover only the failure path and `run`); the spec test covers compile and deployment
only. Fix: add `boot_test.go` cases with a spec compiled against a merged registry declaring `quality` and `boot_id` (expect required), and the same
spec against builtin only (expect not required); add one `run-live` with `--event-schemas` that admits an event of the custom type.

**M4. A business rule lives in the composition root.**
`internal/runtime/internal/composition/planes.go:64-80`: `requireIntentEffectors` decides "only active dispatch is checked" and `intentTypes`
projects the spec's intents. Per AGENTS (logic lives in the owning module, composition wires) the decision belongs in `runtime/internal/app`
(next to `RequireOwners`) or in `spec` (e.g. `CompiledSpec.DispatchedIntentTypes()` returning nil for shadow). Composition would then be
`routes.RequireOwners(compiled.DispatchedIntentTypes())`. Not a gate failure today; it is the first place the next "check X at start" will be added.

### Low

**L1. The `watch` allow-list edge was avoidable.** `internal/watch/service.go:17` imports `watch/internal/domain` for one constant, which grew
`allowedImports`. `watch/internal/app` already imports the domain; an alias there (`app.InstallRoute`) and `watch.InstallRoute = app.InstallRoute`
needs no table edit. Under "move the code instead" this is the cleaner route. Also `decisions/internal/domain/binding.go:91` keeps a second
literal `"install_watch_condition"`; nothing ties it to `watch.InstallRoute`. A one-line test in `internal/architecture` asserting
equality of the literal would close the drift.

**L2. Unrelated gate edit in a feature commit.** `scripts/slow-tests.txt` drops `TestPredictiveMaintenanceAcceptance` and
`...InShadowMode...` in the H0.5b commit. They run 2.5 s each under `-race` here against a 5 s budget: a slower CI host can fail the budget.
Revert or move to its own commit with measured evidence. Violates "no unrelated changes".

**L3. Two decision functions can drift.** `CompositeEffector.route()` (`effect_routing.go:45`) still sends an unowned route to the fallback
(simulated in `emulator`), while `owns()` says the fallback owns nothing when a device is present. Safe today because the start check
prevents it for active specs and the validator refuses undeclared types, but "with a device present the fallback owns no route" is not
enforced at dispatch. Fix: derive both from one `ownerOf(route) (watch|device|fallback|none)`.

**L4. `GatewayEffector.OwnsRoute` on a nil facade value panics** (`internal/device/effectors.go:53-54`, `e.effector` dereference), while the app type
guards `e != nil`. Guard in the facade or document non-nil.

**L5. The root flag is accepted and ignored by every other command** (`explain`, approvals, ...). `registerEventSchemasFlag` is persistent on
the root, so `--help` of unrelated commands lists it. Acceptable for the plan ("root flag"); consider noting "ignored by other commands" in the usage string.
`validate` does not print the digest; `run`, `run-live`, `serve` do.

**L6. Guard coverage limits (H0.4).** Vocabulary comes only from `examples/*/*.situation.yaml` (`domain_knowledge_test.go:27`). A fan spec placed deeper
or outside `examples/` is not seen; the self-test depends on `set_indicator` staying in an example. Parameter property names are not vocabulary,
so `entity_id` (baseline.go), `quality` and `boot_id` (operators `boot.go`) stay in code: those are protocol conventions, not domain words, but
say so in the allow-list reason or README. Literal concatenation can dodge a string-literal scan; acceptable for a review gate.

**L7. Baseline mapping limits.** `baseline.byPhase` chooses string enum values only (schema, compiler and docs agree). The plan text says "preset or
value"; numeric, boolean and preset parameters (a fan duty) cannot be chosen. Also the baseline lives inside the spec digest, and the live
`policy_digest` derives from the spec digest (`policy/internal/domain/definition.go:40`), so editing replay-only baseline data changes the
digest a bench gateway allow-lists. Both are consequences of decision 3; record them in the fan spec notes.

**L8. Documentation gaps.** `internal/spec/UBIQUITOUS_LANGUAGE.md:20` still names `LookupEventSchema` as the code term (now `CompiledSpec.EventSchema`
and `EventSchemas.Lookup`); no terms added for route owner, event-schema registry or baseline choice; AGENTS.md "Forbidden Changes" still says
event schemas are changed by editing `event_schema_data.json` only (add: or a directory given with `--event-schemas`); AGENTS.md architecture list for
`internal/spec` does not mention the registry or `--event-schemas`.

## Where a new device command still needs a code change

Routing: none. A device route is a catalog entry plus a spec intent (H0.1, H0.2 tell you at start if they disagree).
Remaining places, all data or convention, none code-per-device:
- The reading's event schema: now a file in the `--event-schemas` directory (after M2 is fixed, validated at start).
- Replay baseline: the spec's `baseline.byPhase`, enum parameters only (L7).
- The bench gateway's `--device-policy-digest` allow-list changes with every spec edit (including baseline data).
- Fixed conventions in code: parameter `entity_id`, device fields `quality` and `boot_id` (operators), the watch route in `decisions`.
- `explain` and any `LoadDeployment` caller see only builtin schemas (documented).
