# Survey B: internal/world, internal/perturb, internal/device, internal/deviceworld

Read-only survey. Every production file was read. `go test -cover` on the four packages passes
(world 66%, perturb 86%, device 84%, deviceworld 83%). Confidence tags: [H] read + traced, [M] traced, not run.

## 0. Cross-cutting facts
- world and perturb contain NO I/O, clock reads, env, goroutines or locks (grep clean). Only `time.Parse/Format` (perturb/helpers.go:49-76).
- All I/O/goroutines/locks/global state sit in device: uds.go (net/os/goroutine/slog), contract.go (`embed`, global `compiled` + `schemaMu`), device.go:66 (`Device.mu`), safe_stop.go:111,115 + exchange.go:51,55 (global slog in rule code).
- Wall clock enters only at the edge: internal/cli/device_world.go:81 `time.Now()` (bypasses `internal/wall`, which has dead `Now`).
- deviceworld drives a separate `world.World` ("device-world", seed 1, EmitDisabled; cli/device_world.go:31) from device micros. World stays single-goroutine only because `Device.mu` serialises and the listener is sequential.

## 1. internal/world (21 files, 2272 LOC)
**Today**: seeded DES core: entities/channels, dynamics (F0/F1 integrators), faults, effectors + idempotency, native event emission.
**Should own**: exactly that, as ONE package (all of it shares private `World` state; splitting forces exporting internals). Keep. Improve internally, not by package split.

**Used outside (prod)**: New, Options, World, Advance, Clock, StateValue, Reading (truth only), InjectFault, ClearFault, ListFaults, ActiveFaultsCount, InvokeEffector, EffectorCalls, EmittedCount, EntityIDs, InitialEntityIDs, AddEntity, Retire, Entity(), NextEventNS, PendingKicks, SetEmitter, SetFailureMode (run/quiescence.go:55), RenderID (mcp/run.go), InvokeResult, EffectorCall, ErrInterlockRefused (deviceworld/plant_execution.go:37), `World.ID` (audit/scenario_setup.go:16), `World.Spec` (deviceworld/validation.go:36), `World.StartNS` (deviceworld/plant.go:101), Entity.Type/BornNS (mcp/world_config.go:66).
**Dead / over-exposed**:
- No users anywhere, tests included: `PendingEvents` (clock.go:21), `HiddenStateSnapshot` (effector_log.go:35; its only caller of `sortedStateNames`), `ErrEffectorRefused` (effectors.go:39).
- Used only by tests: `DynamicsFor` (clock.go:93, only world-internal), `FaultInfo.Severity` (never populated, faults.go:128), `ModeOK/Slow/AckLost/Reject/Partial/SilentNoEffect` (only world; `ModeConfirmedNoEffect` used by a deviceworld test).
- Exported fields never read outside the package: `World.Seed/ClockNS/Noiseless/EmitDisabled` (world.go:46-50), `Entity.States` (type `*stateValue` is unexported, unusable), `Entity.RetiredNS`.
- `AddEntity(id, at, params)`: `params` is ignored by `addEntity` (entities.go:31); the only caller passes nil (run/commands.go:87). `renderID`'s params path (entities.go:18,70) is therefore dead.
- `SetFailureMode` is documented "test-only" (effector_policy.go:42) but run/quiescence.go:55 calls it in prod.

**I/O / clock / state sites**: none. Lazy compile cache `argSchemas` per world (effector_policy.go:78). RNG only via `w.substream` (clock.go:50); every draw site is a named substream.

**Mixed responsibilities**
- clock.go mixes the event loop (Advance/processScheduled*) with entity registry queries (EntityIDs, AddEntity, Entity), RNG substreams, dynamics lookup and `quantize`.
- effector_log.go is three things: call ledger (recordCall/EffectorCalls/resultDigest), hidden-state snapshot, kick counter (PendingKicks).
- availability.go mixes availability renewal with churn births/retire (entity lifecycle) and `isConfirmationChannel`.
- `recordCall` takes 11 positional params incl. two bools (effector_log.go:8); the same string/args tuple (effector, entityID, commandID, args, atNS) is threaded through 6 functions in effectors.go. A small `invocation` struct would fix this.
- Dead parameters threaded everywhere: `atNS` in pickFailureMode/ackLatency/interlockHolds, `t` in cadenceEmits/linkDelay, `ent` in `initial`, `updateAvailability` returns an always-nil error (availability.go:13), `integration.go:20-22` comment mentions a "state record" that is not returned.

**Vocabulary / types**
- `Advance` returns `(int, int, error)` (clock.go:25): (emitted, effectsApplied) unnamed.
- Failure mode is a raw `string` (`Mode`, `ForceFailureMode`, `SetFailureMode`); a `FailureMode` string type is JSON-shape neutral.
- `resultDigest` (effector_log.go:15) is a JSON string, not a digest. `Reading` (public, noise-free) vs `reading` (private, cadence) vs `observe`. "kick/driver/shadow/effect" are four words for one concept; `Mode` means failure mode here but cadence mode in model.
- Domain literals as bare strings in switches: cadence modes (cadence.go:11,22), ValueType (reading.go:12), Tier "F0/F1/F2" (state.go:118), F1 forms (integration.go:82, integration_forms.go:9), onset shapes (dynamics.go:89), interlock operators (effector_policy.go:65). These belong as typed enums in `model` (cross-package, out of scope here).
- `Entity`/`channelRunState` use zero values as sentinels: `walkLastNS==0` (reading.go:38), `clearedNS>0` (state.go:49), `next>0` (observe.go:98).

## 2. internal/perturb (8 files, 898 LOC)
**Today**: applies named delivery perturbations to a native event stream; holds reorder/flap/backfill buffers. **Should own**: same. Single package is right; it is pure and 85% covered.
**Used outside**: New, Layer.{Apply,Clear,Process,Flush}, Delivered, Names (suite/commands.go). Unused outside: `Active` type (never returned; unexport), `ActiveIDs` (internal only), all 19 name consts (only Names/validation; suite uses `Names`).
**I/O/state**: none; global read-only `paramRules` (parameters.go:63). `Layer` has no lock; callers must serialise.
**Mixed**:
- transforms.go is a 3-level switch ladder (applyOne -> apply{Multiplicity,Delivery,Timing,Payload,PayloadContent}) over the same name string; a `map[string]transformFunc` would replace ~80 lines.
- Every transform takes unused `atNS` and often an unused receiver `l`; `Process` has a dangling "Post-pass" comment (perturb.go:111).
- Time rewrites round-trip `string -> time.Parse -> string` (helpers.go:48-77) and silently return the input on parse error; reason strings are `model.Delivery*` untyped consts; `Delivered.Delivered` stutters.
**Vocabulary**: `Apply` activates and `applyOne` transforms; `DelayTail{mean_s, sigma_s}` actually sums two exponentials (transform_timing.go:22) and there is no sigma. Defaults are scattered literals (rate 0.02 for duplicate, 0.01 elsewhere; `paramFloat(..., def)` at each use).

## 3. internal/device (20 files, 2049 LOC + contract/ fixtures)
**Today**: serial-device emulator: wire codec + contract schemas, capability catalog, admission/state machine, protocol fault schedule, NDJSON session, UDS listener, wire-frame fault gate. Five responsibilities in one flat package.
**Used outside (prod)**: Listen, ParseFaultSpec, ValidateFaultSchedule, FaultInjection, Config, New, Device, Plant, Capabilities, LoadCapabilities, `Capabilities.SafeStopNames` (cli); PlantCommand, PlantEffect, ErrPlantUnavailable/Interlocked (deviceworld).
**Dead/test-only exported**: `TargetNames` (no users), `Device.Advance` + `manualMono` default clock (no prod/test callers; doc says "no-op with custom clock" but it is not), `SetFaultSchedule` (no users), `SafeStopper`, `TargetCapability` (returned only by unexported `target()`), `ErrInjectedDisconnect`, `QueryStateControl`, `ProtocolVersion`, `ServeConn*`, `EncodeRecord/DecodeRecord`, `HandleCommand`, `BootID`, `Reboot`, `SetFaults`, `AcceptedCommandCount`, `WireFaults`, `Faults`, `Outcome`, 7 `Fault*` consts: used only by tests or within the package. `ApplyCommand` is used by cli/deviceworld tests and as HandleCommand's inner step.
**I/O/state sites**: uds.go:92-117 Listen/Lstat/Remove/ListenUnix + `go acceptDeviceConnections` (no join, no way to wait; active conn not closed by listener.Close); uds.go:157-164 slog; contract.go:26-96 embed + global schema cache; device.go:65 mutex (plant/world called under lock: command.go:7, state.go:70).
**Mixed**:
- uds.go: session loop (ServeConn), protocol records (malformedReceipt/Result, QueryStateControl, isQueryState) and OS socket lifecycle in one file.
- state.go: HandleCommand (wire framing + encode) lives on the domain type; `device` core cannot be used without the jsonschema/embed codec.
- capability_*.go: file-format decode + digest + domain validation + admission data in one flow; catalog has two generations (routes vs "legacy" targets) with two digest rules (capability_load.go:116-126).
- command.go/admission.go/command_params.go re-assert `command["target"].(string)` etc. from `map[string]any` in 4 places (admission.go:14-17,52-58; command.go:53-55; command.go:9-10).
- Strict-decode + reject-trailing-JSON is copy-pasted 3x (codec.go:94, capability_load.go:93, deviceworld/bindings_compile.go:23).
**Vocabulary / types**: three "fault" concepts: `Faults{AckLost,Stuck}` (physical), `FaultInjection` (schedule by accepted ordinal), `WireFaults` (frame index); `Outcome` vs wire `result` vs `PlantEffect`; "target/route/capability/binding". Reject codes ("wrong_boot", "wrong_target", "out_of_range", "not_ready", "expired", "unknown_operation", "malformed", "duplicate", "interlocked"), statuses and message_types are bare strings repeated in 6 files; Fault names are untyped consts. Receipt/Result/State are `map[string]any` with `float64(...)` casts (outcome.go, state.go:112, uds.go:66-86). Micros (device) vs ns (world) convert in deviceworld only.

## 4. internal/deviceworld (5 files, 562 LOC)
**Today**: (a) strict binding-catalog decoder, (b) composition validation against a world spec, (c) `device.Plant`/`SafeStopper` adapter that advances and invokes the world. Correct as one package: it IS the adapter; keep.
**Used outside**: LoadBindings, ValidateBindings, Binding, New (cli/device_world.go only). `deviceworld.Plant` type name collides conceptually with `device.Plant` (port); fine, but note.
**I/O**: none. No lock (relies on Device.mu). `stateValue()` (plant.go:91) duplicates the inline read at plant_execution.go:53-55. `worldInstant` (plant.go:98-107) guesses boot-relative vs epoch micros by magnitude ("historical test form"): ambiguous heuristic. `compileBindings`/`validateBindingCatalog` iterate maps, so which of several errors surfaces is nondeterministic (message only).

## 5. Probable bugs / hazards, ranked
1. [H~80%] deviceworld/plant_execution.go:75: world command_id is constant `"safe-stop/"+target`; world replays any known command_id inside `idempotency_window_s` (900-3600s in shipped domains, default 3600; world/effectors.go:85-93,116-122). A second lease expiry/reboot for the same target inside the window gets the cached result (EffectApplied=true) and applies NO kick, while `completeSafeStop` (plant_execution.go:82-92) returns Energized=false. The device then reports safe_state while the world still runs: the instrument lies. Fix: unique id per safe stop (device boot id + ordinal or atMicros), or bypass idempotency. This changes the effector-call log, so it needs a regression test and a recorded behaviour change.
2. [H~85%] device/command_execution.go:62-66: `rememberExecution` clears only `AckLost`. The stored outcome keeps `Duplicate`, `Disconnect`, `Fault`, `AcceptedCommand`. A retry with the same idempotency_key returns it (command_execution.go:10-15), so exchange.go:71-79,91-93 re-injects the disconnect on EVERY retry (retry livelock) and repeats duplicate/log lines. The disconnect fault has no test at all (grep). Documented as one-shot ("consumed when ordinal reached", state.go:23).
3. [M~75%] world reads are not pure. `stateAt` -> `naturalValue` -> `integrateF1` (integration.go:54-67) takes a partial RK4 step ending exactly at the read time, and `stochasticEnvelope` (dynamics.go:122-133) draws one Norm per read-bounded step. Run worlds read all states at every emission (run/emission.go:44-48) while solver worlds (EmitDisabled, truth/solver.go:158) do not, so step partitions and draw counts differ between "the world" and the oracle's copy. Also a fault's walk is shared across its states but each state has its own substream (state.go:127), so multi-state stochastic faults depend on which state is read first. Reads at t earlier than the last integration silently return the later value (docs claim "value at t"). Not fixable without changing numbers: document, pin with a test, decide separately.
4. [M] `World.Reading` (observe.go:10) is documented noise-free but on a non-Noiseless world draws the noise substream (reading_values.go:68) and mutates drift state (reading.go:37-48). Callers (truth) use Noiseless worlds, but nothing enforces it.
5. [M] perturb Reorder decides swaps by `atNS%2==0` (helpers.go:83), not by the layer's RNG: events at whole-second timestamps always swap pairs. Deterministic but not a seeded distribution as the parameter doc implies.
6. [L] Idempotency key is `command_id` alone (world/effectors.go:86,121): same id with a different effector/entity/args replays the first result. Interlock refusals are not cached (by design?).
7. [L] Device dedup replay returns the ORIGINAL command_id inside the receipt/result for a retry with a new command_id (digest excludes command_id, params.go:39-51). Check against contract; maybe intended.
8. [L] Capability loading: `EnergizeField` for route catalogs = alphabetically first bound field (capability_targets.go:35-45); a route with `lease_ms` + another numeric field makes the no-plant energized rule depend on lease_ms. Error text depends on map order (capability_bounds.go:10-33, capability_routes.go:67-76).
9. [L] `Listen` accept loop has no join/close of the active connection, and ServeConn skips `gate.flush()` on injected disconnect (a held swapped frame is lost). `ClearFault`/`InjectFault` accept past times (retroactive reads). `Perturb.order` and unflushed open-ended flap buffers grow unbounded. Cyclic F1 `inputs` would recurse (domain validation presumably prevents; unverified).

## 6. Target shape
- **world**: one package. Behaviour-neutral tidy: delete dead API, unexport fields/types, `FailureMode` type, `invocation` struct, named `AdvanceResult`, split clock.go/effector_log.go/availability.go by named concern (registry, lifecycle, call ledger). Keep stateAt purity issues (item 3) out of the refactor.
- **perturb**: one package; unexport `Active`; name->transform table replaces the switch ladder; drop unused `atNS`/receiver params.
- **device** (the one that needs structure; every new package has a current caller):
  - `internal/device` = domain + use cases: Device state machine, admission, Capabilities (+ loader), Plant/SafeStopper ports, fault schedule. Imports only `canonical`; no jsonschema/embed/net/os/slog. Typed `commandView` parsed once; `RejectCode` consts.
  - `internal/device/contract` (dir already exists with schemas/conformance) becomes Go package `contract`: embed + `EncodeRecord/DecodeRecord` + `ProtocolVersion` + `maxFrameBytes`; this is the only place with jsonschema.
  - `internal/devicewire` (or `device/uds`): `ServeConn*`, frame gate (`WireFaults`), malformed/query_state control, `Listen`/stale-socket handling; owns `HandleCommand` framing and slog. Imports device + contract. cli then imports devicewire for `Listen`.
  - test/architecture/dependencies_test.go allowlist updated in the same change. Do NOT introduce a generic "app/service" layer.
- **deviceworld**: keep; share one strict-JSON decode helper (a small function in `model` or `canonical`) with device; delete `stateValue` duplicate.

## 7. Ordered rounds (small, behaviour-neutral unless marked)
1. Dead-surface prune (world: PendingEvents, HiddenStateSnapshot+sortedStateNames, ErrEffectorRefused, AddEntity params; device: TargetNames, Advance, SetFaultSchedule; unexport perturb.Active, World.Seed/ClockNS/Noiseless/EmitDisabled, Entity.States/RetiredNS). Guard: `go build ./...`, `make function-length`, tests using these in own package. Proof: existing suites + `deadcode`.
2. Typed vocabulary, JSON-neutral: `world.FailureMode`, `device.RejectCode/Status/MessageType`, `perturb.Name`. Guard: string values and JSON shapes byte-identical. Proof: conformance golden frames (device), artifact digest tests (run), `determinism_test.go`.
3. world internal tidy: `invocation` struct, drop dead params, `AdvanceResult`, file moves. Guard: RNG draw order (substream names and call order unchanged), idempotency/expiry, interlock before replay order. Proof: `effector_order_test.go`, `determinism_test.go`, `oracle_test.go`, replay-digest tests in run.
4. perturb transform table. Guard: per-perturbation RNG draw order (short-circuit `!Delivered || rng...`), `l.nextID` allocation order, `ActiveIDs` application order. Proof: `perturb_test.go` plus a new fixed-seed golden over all 19 names across a mixed stream.
5. device: parse `commandView` once; extract `contract` package; extract wire/session/listener package. Guard: admission order boot -> freshness -> target -> operation -> bounds (transport_order_test.go), dedup-before-admit, fault-ordinal accounting incl. stale/expired, frame bytes (canonical, sorted keys, trailing newline), lock scope (plant called under mu). Proof: `conformance_test.go`, `uds_test.go`, `wire_test.go`, `fault_schedule_test.go`; add a test with an injected disconnect/duplicate first, so item 2 is pinned.
6. Corrections, each separate with its own regression test and an explicit behaviour-change note: (a) device one-shot flags in dedup (hazard 2); (b) unique safe-stop command_id (hazard 1). Both are fault-injection/evidence changes: update docs/refactoring evidence.
7. Last / optional, needs design sign-off: typed Receipt/Result/State structs in device (canonical encoding must stay byte-identical); resolve world read-purity (hazard 3), which changes numbers.
