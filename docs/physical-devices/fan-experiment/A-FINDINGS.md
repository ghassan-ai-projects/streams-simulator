# A0 findings: Agentic Stream's plant wire adapter (design and briefs A1 to A4)

Investigated 2026-10-20, read-only, in `/Users/ghassan/my-projects/agentic-stream` at commit `91c2079d` (HEAD of `fan-h0`; the H0
agent had not committed yet). Paths are relative to that repository; `file:line` is as of `91c2079d`. "UNVERIFIED" marks what I read
but did not run. Plan: `PLAN.md` D1, D3, D4, D5, D7, D9, D10, section 5 A0 to A4, section 8.1 item 4. H0 context: `H0-FINDINGS.md`.

## 0. Corrections to the plan that the code forces (read first)

| # | Plan says | Code says | Consequence |
|---|---|---|---|
| C1 | Fan deadline 120 s so an operator can approve at the node | The dispatcher bounds one provider call to 0.9 x its lease: `internal/actions/internal/app/dispatch.go:28-34`; the lease is `sources.DefaultLease` = 1 min (`internal/sources/sources.go:113`) and `composeDispatcher` passes no `LeaseFor` (`internal/runtime/internal/composition/planes.go:142`). A call past 54 s becomes `UnknownOutcomeError` (`dispatch.go:45-47`), and the follow-up verification runs on the same expired context (`dispatch.go:48`), so the command sits in `reconciling` until an operator resolves it by hand. | A `DCMD` held at the node for approval cannot be waited on past 54 s. Fix in A3a: a `--dispatch-lease` flag plumbed to `actions.Config.LeaseFor` (the field exists, `internal/actions/service.go:36`), and a start-time check that every route's `expires_after_ms` plus the verification window is below 0.9 x lease. Two commands cannot overlap in the dispatcher (`DispatchOnce` takes one), so an LED command waits behind a fan approval; accepted for the experiment (UNVERIFIED how `dispatchApprovedCommands` orders them). |
| C2 | idempotency key = sha256 of canonical `{device, command, args, expected_boot_id}` | Agentic Stream already gives every command a per-intent key: `sha256(tenant\|intent_id\|type\|target)` (`internal/policy/internal/domain/commands.go:13`), carried as `actionport.Command.IdempotencyKey`. The streamsim node keeps the key for the session and replays the stored receipt and result with no second effect when key and content repeat (`internal/plantwire/internal/app/commands.go:32-34,57-58`, `execution.go:12-14`); the brief says nodes do not check that the key is a content digest (`briefs/gateway-plant-wire.md:62`). | A content-derived key would make a second approved `set_cooling 45` in the same boot (within 24 h) replay the first answer and never move the fan. **Recommendation: forward `Command.IdempotencyKey` unchanged** (a retry of one intent keeps its key; a new intent gets a new key). SPEC section 9 wording ("is the digest of its content") should say "host-chosen, unique per intent" in 1.4.0. Owner decision Q1. |
| C3 | Agentic Stream's catalog names `stop_led` / `stop_fan` as safe stops (PLAN 3.2) | `validateSafeStops` forces `operation == "safe_stop"` (`internal/device/internal/domain/catalog.go:191-199`), and the catalog loader rejects unknown members (`catalog.go:73`). | The catalog keeps `safe_stop`. The plant wire adapter resolves the safe stop as the **kind-declared** `safe_stop: true` command of the device's commissioned kind (`stop_fan`, `stop_led`), never from the catalog. No catalog format change. |
| C4 | Heartbeat event "from node liveness" | `missing_heartbeat` uses event time for presence and the **processing clock** for absence (`internal/operators/internal/domain/heartbeat.go:15-29,98-160`), and counts only events with `data.quality` absent or `valid` and no envelope quality flags (`operators/.../samples.go:58-71`). | Emit the heartbeat only while the link and the device are alive; stop emitting when not, and the timer raises `link_missing_90s` by itself. Its event time must come from the data stream (latest sample or `NSYNC.world_time_ns`), never the host clock (section 4). |
| C5 | Origin check needs only the two consent flags and the list | Consent exists only in the `physical` profile (`internal/device/internal/domain/profile.go:50-84`); `ValidateEffectProfile` knows no board. | A3b adds `HardwareBoards` to `EffectProfileConfig`/`ProfileRequest` with the rule "boards only with `physical` + `--live-actuation` + `--owner-authorized`". Both rungs run profile `physical` (D3: no restart at the switch). |

## 1. The ports and what an adapter reuses

| Piece (evidence) | Reuse | Re-implement |
|---|---|---|
| `actionport.Effector`, `AuthorizedEffector`, `DeviceStateVerifier`, `VerifiedEffector`, `UnknownOutcomeError`, `Command{CommandID, IntentID, TenantID, EffectorRoute, NormalizedTarget, IdempotencyKey, PolicyDigest, NotBeforeMonoUS, Payload}` (`internal/actionport/internal/domain/effect.go:9-86`) | all of it as is; H0 adds `RouteOwner` (`H0-FINDINGS.md` H0.1) | nothing |
| Dispatcher use: `DispatchAuthorized(callCtx, cmd, auth)` then, on success **or** unknown outcome only, one `VerifyDeviceCommand` (`actions/internal/app/dispatch.go:38-74`); `VerificationPending` -> status `manual_review` until the verifier returns a final status (`actions/internal/domain/dispatch.go:40-58,127-147`); a verifier that returns a status with no error also closes an unknown outcome (`DeviceCheck.ReconcilesUnknown`, `dispatch.go:121-125`) | the contract | the **semantics**: accepted + `succeeded` -> `Effect{VerificationPending:true}`, then verify; terminal refusal -> known failure (plain error, status `failed`); loss after the command may have reached the node -> `UnknownOutcomeError` |
| Device-wire effector `GatewayEffector` (`device/internal/app/gateway_effector.go:44-175`): materialize, one serial exchange under `Session.mu`, `failedDeviceEffect` (`sent` => unknown), `acceptedDeviceEffect` | pattern only | the exchange: plant wire is multiplexed by `command_id`, so a per-command mailbox replaces the lock (the dispatcher is still serial) |
| `Session` handshake and claims (`device/internal/app/session.go:74-223`, `session_exchange.go:58-125`): `authority.Claim`, `BindCommand`, `AssertClaim` before the send and after the result, `RecordDeviceState`, `OpenReconciliation`, `ReleaseClaim` at close; errors `ErrReconciliationRequired` | **the `authority.Service` facade** (`authority/operations.go:20-105`), given by `cmd` exactly as `effect_profile.go:155` does; `Owner{Epoch, Instance}`, `DeviceBoot{DeviceID, BootID}`, `SealReconciliationEvidence` (`device/.../session_evidence.go:26`) | the call order (claim -> bind -> assert -> send -> receipt -> assert -> result -> assert), written for plant wire |
| Receipt cache by idempotency key (`session.go:63`, `session_exchange.go:141-151`) | - | **not needed**: the plant wire node replays by key for 24 h (SPEC section 9 "Idempotency"); keep only an in-flight map by `command_id` |
| Catalog `CapabilityCatalog.Routes` (`Operation, Target, TargetBindings, SelectorField, ExpiresAfterMs, Presets, Bounds`; `device/.../catalog.go:24-48`), `Materialize` (`materialize.go:28-157`: target check, selector -> preset, bounds, `expires_after_ms`) | the catalog file and loader (`device.LoadCapabilityCatalog`, alias `device.CapabilityCatalog`) | **do not copy** `Materialize`: A3a splits out of it a transport-neutral `BoundedCommand` use case in `internal/device` (target, operation, bounded parameters, expiry; device-wire `Materialize` calls it) so both adapters share one implementation (AGENTS "Duplication Scan" 2/3, `dupl` threshold 75 `.golangci.yml:76`). Presets must carry physical units only (D7): the route code is unchanged, `{"level_percent":45}` is just a parameter map |
| Verification by state query (`gateway_effector.go:142-175`, `verification.go:19-46`, "single numeric output parameter, energized iff positive") | - | plant wire verification: section 5 |
| Safe-stop lane `SafeStop` (`gateway_effector.go:107-137`, `session_safety.go`) | - | **No production caller exists** (grep of `cmd` and `internal` outside the device package finds none at `91c2079d`; whether `make deadcode` excuses it is UNVERIFIED). Defer a plant wire `SafeStop` to a later round with a real trigger; do not add an unreachable export (`scripts/check-deadcode.sh`) |
| Ingress live path: `ingress.Service.ServeLive` listens, reads framed lines, `processLine` = admit (`domain.AdmitEnvelope` + `eventlog.ValidateEnvelope`, `ingress/internal/app/admission.go:14-23`) -> quarantine raw/envelope or `sink` (`ingress/internal/app/live.go:33-72`); the sink is `Pipeline.ingestLiveEvent` = lock, assert owner, `log.Append`, `runAfterIngest` (`runtime/internal/app/pipeline_run.go:45-87`) | admission, quarantine and the sink | the producer: an in-process feed, not a socket. A2 adds to `ingress` one use case `ServeFeed(ctx, feed, sink)` (admit an envelope, same quarantine) and to `runtime` `Pipeline.RunFeed`, both thin; `RunLiveSocket` stays |
| Checkpointing | file replay only (`ingress/internal/store`, `ReplayJSONL`); the live socket has none | none: a plant wire session has no resume point; store-and-forward arrives as `historical:true` DDATA |
| Owner-only unix listener | two copies exist (`ingress/internal/transport/listener.go:13-65`, `worker/internal/transport/uds.go:19-60`); the worker one also enforces a 0700 parent, which SPEC section 2 requires | `cmd` opens `worker.ListenEvidenceSocket(path)` (`worker/worker.go:36`) and hands the `net.Listener` to `plantwire.New`. A third copy would trip `dupl`. Name is odd; extracting a shared foundation is a separate refactor (Q8) |

## 2. Startup composition and flags

Today: `serve` -> `serveRuntime` -> `core.openEffects(...)` dials the gateway synchronously and fails startup on a bad handshake
(`serve.go:135-149`, `live.go:121-130`, `effect_profile.go:105-166`) -> `startContinuousPipeline` -> `openContinuousPipeline` compiles the
spec, logs `spec loaded` with the spec digest (`serve_source.go:26-42`) -> `core.startPipeline` -> `runtime.NewPipeline` -> `composeEffectors`
builds `CompositeEffector(watch, fallback).WithSerial(gateway)` (`planes.go:43-57`; `PipelineConfig.GatewayEffector *device.GatewayEffector`,
`composition/pipeline.go:44-46`) -> sources start in `startContinuousSource` (`serve_source.go:52-64`). `validateServeSources` requires
`--spec` with exactly one of `--trace`, `--live-socket` (`serve.go:96-111`). Profiles: `simulated` (no gateway), `emulator` (gateway, simulated
fallback), `physical` (gateway + `--live-actuation` + `--owner-authorized`, fail-closed fallback) (`effect_profile.go:180-185`, `profile.go:33-84`).

Where plant wire enters (all additive, `TestExperimentCommandsKeepTheirFlags` stays green):

| Seam | Change |
|---|---|
| Flags on `serve` only (`registerServeFlags`; `run-live` is a trace runner) | `--plant-wire-socket PATH` (host listens), `--plant-wire-environment simulated\|hardware\|mixed` (default `simulated`; the fan experiment sets `mixed`), `--plant-wire-kinds DIR` (kind files), `--plant-wire-site FILE` (section 4), `--hardware-boards a,b` (D3), `--plant-wire-epoch N` (default Unix seconds at start, Q2), `--dispatch-lease DUR` (C1). `--device-catalog` is reused for the routes |
| `effectProfileOptions` | `PlantWireSocket != ""` counts as a gateway link (`profileConfig` sets `GatewayOptions`); mutually exclusive with `--device-socket` (one effect link per process in the A rounds; both together needs two catalogs: Q7); needs no `--device-firmware-digest`; `HardwareBoards` rule from C5; host environment `hardware`/`mixed` requires `physical` + both consent flags |
| `validateServeSources` | `--plant-wire-socket` is a third continuous source with `--spec`, exclusive with `--trace`; allowed together with `--live-socket` |
| `effects` struct (`live.go:45-48`) and `PipelineConfig` | `effects.plant *plantwire.Host`; `PipelineConfig.DeviceEffectors []DeviceEffector` (interface = `actionport.VerifiedEffector` + H0's `RouteOwner`) replacing the single concrete `GatewayEffector`; the composite asks each owner; two owners of one route stop `serve` at start (extends H0.2). Composition imports no `plantwire` (it takes the interface) |
| Source start | `startContinuousSource` adds `run("plant wire", func(ctx) error { return pipeline.RunFeed(ctx, host.Run) })` and `waitForSocket` (the live-socket readiness poll generalised) |
| Start-time logs (A4) | `policy.DigestForVersion(compiled.Digest)` (the value `validate` prints, `main.go:96-101`) beside `spec loaded`, plus the commissioned kind digests, the site file digest and the epoch |
| Dispatch timeout | C1 |

## 3. Module placement and gate edits

New module `internal/plantwire` (kind: adapter module with an edge, beside `device`; responsibility: host one plant wire session and serve
two existing ports from it; callers: `cmd/agentic-stream` only). Layers (one-way, `AGENTS.md` shape):

| Package | Holds | Imports (new edges in bold) | Layer (reviewed; must be strictly above its imports) |
|---|---|---|---|
| `internal/plantwire/internal/domain` | pure: record views, sequencing, admission order, environment rules, kind-digest check, commissioning, origin, entity directory seam, site model and mapper (builds `contractsv1.Envelope`), quality table, heartbeat state machine, DCMD plan, receipt classification, expected-value rule | `canonicaljson`, `contractsv1`, `kernel`, `actionport` | 13 |
| `internal/plantwire/internal/wire` | `go:embed` of the vendored contract, schema validation, 64 KiB / 1 MiB limits, exact 64-bit integers, digest and version check | `domain`, `canonicaljson`, the existing `jsonschema/v6` dependency | 14 |
| `internal/plantwire/internal/transport` | framed connection over an injected `net.Listener`, lane-aware writer with the 2 s write deadline, accept loop | `domain`, `wire` | 15 |
| `internal/plantwire/internal/app` | host supervisor, session, barrier, ingest, effector, authority calls | `domain`, `wire`, `transport`, `actionport`, `authority`, `contractsv1`, **`device`** (catalog, `BoundedCommand`), `sources`, `telemetry`, `canonicaljson` | 25 |
| `internal/plantwire` | facade: `Config`, `Host`, `New`, `Run`, `Dispatch`, `DispatchAuthorized`, `VerifyDeviceCommand`, `OwnsRoute`, `Close` | `app`, `domain`, `actionport`, `contractsv1` | 26 |

Table edits in `internal/architecture` (a new module is allowed, `AGENTS` rule; none of them loosens a gate): `allowedImports` (5 packages + `cmd/agentic-stream`
gains `internal/plantwire` and `internal/worker` if not present), `packageLayers` (5 entries), `facadeSpecs` (the facade's reviewed operations,
each one `return h.app.X(...)`), `TestReasoningAndReplayCannotReachEffectImplementations` device-adapter list (add `plantwire`: reaches no
`actions`, `policy`, `episodes`), `language_test` (`UBIQUITOUS_LANGUAGE.md`), `documentation_test` (package comments),
`documentation/architecture/repository-map.md` (`TestModuleMapListsEveryPackage`), `internal/plantwire/README.md`. No `durableOwners`,
`opaqueStoreModules` or `applicationLayerSpecs` entry (no tables, no store). `TestDigestTextHasOneOwner` (`README.md:119`): no `sha256:`
literal in Go; use `kernel.EncodeDigest/DecodeDigest`. `TestApplicationLayersDoNotTouchInfrastructure` and the clock gate: `app` and `domain`
import no `net`, `os`, `time.Now`; time comes from `sources.Clock`, sockets from `transport`. `TestDomainPackagesArePure`: the embed lives in `wire`, not `domain`.
Merge note: H0 edits the same tables (`import_rules_test.go`, `package_layers_test.go`, README). Branch the A work from the `fan-h0` tip (owner rule: branch per milestone
from the previous tip) to avoid conflicts.

## 4. Mapping data and entity resolution

Evidence: `zone.temp.observed/1.0` fields `celsius, quality, raw_value, calibration_id, firmware_id, schema_version, boot_id, seq, device_mono_us`;
`zone.humidity.observed/1.0` uses `percent`; `zone.heartbeat.observed/1.0` has no value field (`internal/spec/internal/domain/event_schema_data.json`).
`eventlog` rejects a payload key the schema does not declare and a wrong enum (`eventlog/internal/domain/schema.go:30-59`). The event `quality`
enum is `valid, warming, invalid, disconnected, rail_high, rail_low`, not the wire's eight codes (SPEC section 10). The golden shape is
`examples/thermal-chamber/testdata/trace-invalid-quality.jsonl` (non-valid: `data.quality` plus an envelope `quality` flag, no value).
Operators fence on `boot_id` (`operators/.../boot.go:13-48`); `seq` and `device_mono_us` are carried but nothing reads them.

**Data, not code** (AGENTS "Do not re-author domain data in Go"; H0's principle: no thermal word in the binary). One site file given at start
(`--plant-wire-site`), outside the repository, schema and a test copy inside `internal/plantwire/internal/domain/testdata`; the embedded
`simulator_data.json` is the precedent for the shape, not for the location (Q3):

```json
{"version":1,"source":"plantwire",
 "devices":{"zone-01":{"kind":"thermal_sensor","entity":{"type":"thermal_zone","id":"zone-01"}},
            "led-01":{"kind":"indicator"},"fan-01":{"kind":"fan"}},
 "readings":[{"kind":"thermal_sensor","metric":"temperature","event_type":"zone.temp.observed","schema_version":"1.0","field":"celsius"},
             {"kind":"thermal_sensor","metric":"humidity","event_type":"zone.humidity.observed","schema_version":"1.0","field":"percent"}],
 "heartbeats":[{"kind":"thermal_sensor","event_type":"zone.heartbeat.observed","schema_version":"1.0","every_ms":10000}],
 "quality":{"good":"valid","warming_up":"warming","suspect":"invalid","out_of_range":"invalid","stale":"invalid",
            "disconnected":"disconnected","manual_override":"invalid","out_of_service":"disconnected"}}
```

Rules (all in `domain`, none names a metric): one envelope per metric of a non-historical `DDATA` whose (kind, metric) has a `readings` rule;
`partition_key` and `entity.id` = the device's entity id; `event_time` = `event_time_ns` in UTC; `ingested_at` = host clock; `data` = `{field: value
(only when wire quality is good), quality, boot_id, seq (the DDATA record's seq), device_mono_us (when sent), schema_version}`; non-valid quality adds
the envelope flag; event id = `pw-` + first 32 hex of sha256(device, boot_id, metric, event_time_ns), so a re-sent sample dedupes in the log and a
new boot never collides. Historical DDATA is not ingested live (SPEC section 7: never updates live state); counted and logged (Q5). Metrics with
no rule (LED, fan) feed commands and verification only. The quality table is the owner's to tune; no entry may map a non-`good` code to `valid`.

**Heartbeat (required by the bench spec, `zone-thermal-bench.situation.yaml` operator `link_missing`, expected every 30 s, missing after 90 s).** A
state machine in `domain`, one per `heartbeats` rule and entity: it tracks `T` = the largest of the session's `NSYNC.world_time_ns` and every
`event_time_ns` seen on that node; when the device is born, not dead and its last good sample is not older than three of its scan periods
(default 6 s), and `T` has advanced `every_ms` since the last heartbeat, it emits one event at event time `T` with `quality: valid`, `boot_id`,
`seq`. No wall-clock read, so replay and a paced world agree, and a silent link stops the events so the operator raises `link_missing_90s`
(C4). 10 s is well inside the spec's 30 s. UNVERIFIED: that `T` keeps moving on a hardware-only session (DDATA every 2 s) and on an unpaced
simulated one (NSYNC each slice): both follow from the SPEC, neither was run.

**Entity resolution, the single seam** (PLAN 8.1 item 4): `domain.Directory` with `EntityOf(device) (type, id)` and `DeviceFor(route, entity)`.
Implementation now: devices from the site file, routes from the catalog (`Route.Target`, inverse of `TargetBindings`, as `Materialize.checkTarget`
does, `materialize.go:61-72`). Later (areas and links in births, a contract change) a second implementation, no change to callers.

## 5. Commands, receipts, results, verification

`Dispatch(ctx, cmd)`, in order (each step returns before the next on failure; "known" = status `failed`, "unknown" = `UnknownOutcomeError`):

1. `OwnsRoute(cmd.EffectorRoute)` (H0's `RouteOwner`): the catalog's routes whose device is commissioned. No session or device not alive: known failure `not_ready`, nothing sent.
2. Resolve with the shared `BoundedCommand` (A3a): target device, operation (`set_cooling`, `show_state`), bounded args, `expires_after_ms`. A kind-schema check of the args against the commissioned
   `args_schema` runs here too (so a catalog preset that drifts from the kind fails at start, not in the field).
3. **Origin check** (D3): device -> board -> board `environment` from the birth. `hardware` board not in `--hardware-boards`: known failure with provider result
   `{refused:"origin_not_authorized", board}`, nothing sent. `simulated`: allowed. Process-level consent is already enforced by the profile (C5).
4. `authority`: `Claim` -> `BindCommand` -> `AssertClaim` (reconciliation barrier => `ErrReconciliationRequired`, as the device path).
5. Send `DCMD`: `command_id` = `cmd.CommandID`; `idempotency_key` = `cmd.IdempotencyKey` (C2); `device`, `command`, `args` from step 2; `expected_boot_id` = the device's current boot;
   `deadline_ns` = `clock.Now().UTC()` + `expires_after_ms` (nanoseconds, UTC; a paced streamsim session lives on that clock, D2); `not_before_ns` = 0 (`NotBeforeMonoUS` is 0 today and has no wire meaning);
   `authority_epoch` = the host epoch; `policy_digest` = `cmd.PolicyDigest`; `fallback_on_lease_loss` = `safe_off` for kind commands carrying `lease_ms`, else `hold_last`
   (reference host uses `safe_off` for all, `host/commands.go:34`; Q6). Strict `DCMD` (unknown member is `malformed`): build from the schema, never from a map copy.
6. Wait on the per-command mailbox (fed by the session reader) until the context ends:

| Observed | Dispatch returns |
|---|---|
| `DRECEIPT queued` | keep waiting (record it in the provider result) |
| `DRECEIPT refused`, any reason (`local_control`, `expired`, `policy_not_allowed`, `stale_epoch`, `wrong_boot`, `duplicate`, `not_ready`, `malformed`, `out_of_range`, `unknown_*`, `interlocked`, ...) | known failure; the reason is the error text and `provider_result.receipt.reason`. `local_control` is terminal, never retried by the adapter |
| `DRECEIPT accepted` then `DRESULT succeeded` | `Effect{VerificationPending:true}` (then step 7) |
| `DRESULT failed` | known failure |
| `DRESULT unknown`, or session lost / context ended after `accepted`, or after the send with no receipt at all | `UnknownOutcomeError`; open the reconciliation barrier for the device boot (`authority.OpenReconciliation`), as `session_exchange.go:206-211` |
| session lost after `queued` and before a terminal receipt | known failure `session_ending` (SPEC section 9, `briefs/agentic-stream-host.md` "Handling refusals"; the node never keeps it for a new host, D-83) |
| receipt/result for another `command_id`, boot or session | dropped and counted; if it is the awaited command's boot mismatch => unknown |

7. `VerifyDeviceCommand(ctx, cmd)`: re-derive the expectation from the same pure rule (no stored state needed, so it also serves the reconcile-after-unknown call): **each argument
   whose name is a writable metric of the device's kind is expected to be read back on that metric** (`level_percent`, `state`); other arguments (`lease_ms`) are not. Tolerance =
   the metric's `resolution` (0 for an enum). Take the first `good`, non-historical sample of that metric on the same boot with `event_time` after the result, within a window
   shorter than the lease (default 4 s; the fan's lease is 5 s, 2 s gaps in hardware DDATA): equal within tolerance => `succeeded`, else `failed`; no sample in the window => no
   status and an error (the dispatcher then records unknown, `domain/dispatch.go:136-141`). A `DSTATE` (`outputs`/`values`) is accepted as a second source; **do not depend on
   `DQUERY`**: a relay sends no `DSTATE` for a real device (`briefs/agentic-stream-host.md`, "a relay sends none"), so DDATA is the primary evidence (UNVERIFIED that the gateway answers `DQUERY`).
   Evidence = `authority.SealReconciliationEvidence("plantwire.sample", device, document{device_id, boot_id, ...}, feedback)`, shaped as `ParseReconciliation` requires
   (`actions/internal/domain/reconciliation.go:51-107`). Also queue a `command_evidence` entry for the next `HSYNC` (`observed` / `not_observed`; only a barrier carries it, D-112).
8. Close of the session: release claims (`session_close.go:39-50`), fail in-flight mailboxes as in step 6, mark devices offline.

Reboot (`DDEATH reboot`, new `boot_id`): `RecordDeviceState` with a document `{device_id, boot_id, ...}` opens the reconciliation (`authority/.../state.go:57-68`) exactly as for device-wire.
How a barrier clears in production today is UNVERIFIED: `Session.ResolveReconciliation` has no caller outside tests (grep). The adapter mirrors the device path and adds no clearing rule (Q4).

## 6. Barrier, liveness, sessions, origin

| Rule | Design |
|---|---|
| `NSYNC` | The reader answers at once with `HSYNC{through_seq: NSYNC.seq}` on the control lane (SPEC section 8: the barrier covers the NSYNC itself), plus up to 128 queued `command_evidence`. The reader never calls the sink: data goes through a bounded queue to a feeder goroutine, so a slow pipeline (`ingestLiveEvent` runs the whole batch, `pipeline_run.go:63-79`) can never delay a barrier or a receipt (three missed barriers end the world, SPEC 8). Dispatch runs in the maintenance loop when `serve` has one (`handDispatchToMaintenance`, `pipeline_run.go:175-188`), so it does not wait on the same goroutine as ingestion |
| Heartbeat | `STATE{online:true, heartbeat:n, heartbeat_interval_ms:1000, authority_epoch, protocol_min/max}` every second after the first; the relay arms its watchdog only after a heartbeat above zero (D-69). Ticker injected (tests never sleep, T5) |
| Handshake | `NBIRTH` -> refusal order as the reference host (`streamsim domain/relay_admission.go`): `protocol_invalid`, `environment_mismatch` (host setting and node-vs-boards, SPEC 2), `contract_unsupported` (**exact** `contract_version` and `contract_digest`, D5), `version_unsupported`, `inventory_mismatch` (a kind digest that does not recompute, a duplicated kind, a required device missing). Then `STATE`. A device whose kind is not commissioned, or whose digest differs from the commissioned kind file, is quarantined: no events, no commands, logged and counted; the session stays |
| Commissioning | Kind files in `--plant-wire-kinds` (byte-identical to streamsim's `examples/devices/thermal-board/*.kind.json`); `kind_digest` = sha256 of RFC 8785 of the entry without `kind_digest`. A test pins them against `streamsim kind validate` output (UNVERIFIED: that `canonicaljson.Marshal` and streamsim's `canonical` agree on every number form; the contract's own fixtures give vectors) |
| New session without restart (D3) | `Run` loops: accept one connection, serve it, on EOF or `NDEATH` end the session (section 5 step 8), accept the next. A second connection while one is live is closed. `bd_seq`, `prior_last_seq`, `session_id` come from the node. Heartbeat counter and both sequences restart per session |
| Gaps | A `seq` gap sends one `NCMD rebirth` until the answering `NBIRTH`; a repeated `seq` is dropped (SPEC 7, stream is ordered) |
| Rebirth | A second `NBIRTH` in a live session replaces the registry; `DBIRTH`s re-register devices; a changed `metadata_version` or command set => quarantine until restart |
| Origin | Board `environment` is claimed by the node; D3 argues the dangerous direction is impossible in streamsim. The adapter never infers origin from names (SPEC 2) |

## 7. Vendoring the contract

Today: `internal/contractsv1/internal/domain/schemas/v1/*.json` embedded with `go:embed` and compiled by `canonicaljson.CompileSchemaJSON`
(`contractsv1/internal/domain/schemas.go:32-89`); there is no sync tooling and no folder-level digest. For plant wire (D5, byte-identical, pinned `DIGEST`, exact match):

- Copy the **whole** `contracts/plant-wire/` folder (1.4.0 after S4) verbatim to `internal/plantwire/internal/wire/plantwire-contract/` (`SPEC.md`, `DIGEST`, `schemas/`, `fixtures/`). Embed `DIGEST` and `schemas/*.json`; tests read `fixtures/` from disk.
- `make plant-wire-sync` (rsync from `PLANT_WIRE_CONTRACT_DIR`, as streamsim's `contract-sync`, `Makefile:268-274`); no sync in release builds. Agentic Stream's Makefile, CI and docs must agree (Definition of Done).
- Guard tests in `wire`: (a) recompute the folder digest (file name, newline, RFC 8785 of each schema, newline, in name order, as `domain/digest.go:17-27` does) and equal `DIGEST`; (b) the version constant equals the SPEC header. The folder has no version file: ask streamsim to add a one-line `VERSION` in 1.4.0 (Q9) instead of parsing prose; (c) the safety-field floor (SPEC 11) over the embedded schemas.
- The fixtures are a free **conformance suite**: `fixtures/exchanges/*.jsonl` (25 scripted host/node exchanges: session start, version, environments, gap-rebirth, expired, stale epoch, policy-not-allowed, queued-across-reboot, lying results, host loss) drive the host with a scripted node. They are the A1/A3 acceptance tests.
- `jsonschema/v6` is already a dependency; no new top-level dependency.

## 8. Tests and test peers

Agentic Stream's end-to-end tests are in-process: `startExperiment` runs `serve` with stand-ins that speak only public contracts (a Tamoz-shaped worker over gRPC and a device-wire stand-in
over a unix socket, `cmd/agentic-stream/experiment_standins_test.go`, `experiment_e2e_test.go:70-95`). The only subprocess is a re-exec of the test binary (`executor/remote/unix_socket_worker_test.go:19-27`). No test runs an external binary.

| Level | Peer | Where |
|---|---|---|
| unit | the vendored fixture exchanges, a scripted node | `internal/plantwire` layers |
| module | an in-process scripted plant wire node (a Go test helper speaking the schemas; `internal/testsupport/plantwirefake`, skipped by `deadcode`) | `internal/plantwire` |
| `serve` e2e | the same fake node, in `startPlantWireExperiment` beside `startExperiment`: births the three thermal kinds, streams temp/humidity, answers a command; assert zero rows in `event_quarantine`, an `outcomes`/`verifications` row, the policy digest in the DCMD | `cmd/agentic-stream/experiment_plantwire_*_test.go` |
| real peer | `bin/streams-simulator plantwire host` is a **host**, so not a peer for us. Our peers are `plantwire node --domain thermal-board --kinds-dir ... --host-socket S` (simulated) and `plantwire connect --plan simulate.plan.json` (S8). `bin/standin-gateway` is a `go test -c` binary (needs streamsim's config, not a CLI): UNVERIFIED how it is driven. Gate on an env var (`STREAMSIM_BIN`), skip when unset (no network, no external binary in the default suite), run by the main session in P2/P4 | `cmd/agentic-stream/..._streamsim_test.go` |

Note the streamsim node needs a paced session (S1) for deadlines in UTC; against an unpaced one (world time 2026-01-01) every `deadline_ns` is `expired`: UNVERIFIED which streamsim flag selects pacing until S1 lands.

## 9. Proposed A0 design summary

- **Module** `internal/plantwire` (section 3): facade + `domain` (13) + `wire` (14) + `transport` (15) + `app` (25). One `Host` value: `Run(ctx, emit)` serves sessions and feeds ingestion; `Dispatch*`/`VerifyDeviceCommand`/`OwnsRoute` is the effect port.
- **Types** (domain): `Record` views per type, `Sequencer`, `Admission`, `Commissioned{kind->digest}`, `Origin`, `Directory`, `Site`, `Mapper`, `HeartbeatClock`, `CommandPlan`, `ReceiptOutcome`, `Expectation`. (app): `Host`, `Session`, `Mailbox`, `Effector`.
- **Data**: site file, kind files directory, catalog (existing), vendored contract folder. Nothing thermal in Go.
- **Flags**: section 2. Both rungs: `--effect-profile physical --live-actuation --owner-authorized --plant-wire-environment mixed --hardware-boards mega-01 --plant-wire-socket S --plant-wire-kinds K --plant-wire-site F --device-catalog C --dispatch-lease 3m`.
- **ADR**: `documentation/design/plant-wire-adapter.md` (tracked; `docs/` is git-ignored and unreadable by code) states kind, layer, responsibility, callers (AGENTS rule 4 before any code).

## 10. Rounds (one commit each, branch `fan-a` from the `fan-h0` tip; sizes S/M/L as in the plan)

Gates every round (in order): `gofmt -l .`; `go build ./...`; `go vet ./...`; focused `go test` of touched packages; `go test ./internal/architecture/...`; `make lint` (dupl,
funlen 15/15, files 300); `make deadcode`; plus where noted. Before the pull request once: `make ci-check` (proto-check, race, coverage floor 70 %/package, budgets). No `//nolint`, no test
sleeps, no comments inside modules, `%w` wraps.

| Round | Content | Size | Extra gates | Exit |
|---|---|---|---|---|
| **A0** | ADR page; stubs of `README.md`, `UBIQUITOUS_LANGUAGE.md`; agree Q1 to Q9 with the owner | S | `make docs-check` | owner accepts the ADR |
| **A1a** | vendored contract + `plant-wire-sync` target + guard tests; `domain` (sequencing, admission, environment, kind digest, commissioning, heartbeat state machine skeleton); `wire` codec and limits; arch tables for the 3 packages | M | `internal/plantwire/...` coverage; fixture `invalid/`, `must-ignore/`, `valid/`, `injection/` pass | every `valid/` record accepted, every `invalid/` refused at its member; digest equals `DIGEST` |
| **A1b** | `transport` + `app` Host/Session: listener injection, handshake, `STATE` heartbeats, `NSYNC`->`HSYNC` at once, gap rebirth, devices registry and quarantine, second session after the first, facade, `cmd` flags and wiring (`--plant-wire-socket/environment/kinds`), profile rules | M | `go test ./cmd/agentic-stream -run 'Experiment\|Profile'` | all `fixtures/exchanges/*.jsonl` host-side expectations pass; two sessions in a row on one socket; `environment_mismatch` etc. |
| **A2** | site file loader + schema; mapper; heartbeat emission; `ingress.ServeFeed` + `runtime.Pipeline.RunFeed` (thin); `validateServeSources`; historical handling; golden test against `trace-invalid-quality.jsonl` shape | M | `go test ./internal/ingress/... ./internal/runtime/...`; `make docs-check` | scripted node's DDATA + liveness -> zero `event_quarantine` rows; bench spec opens a situation in a test; `link_missing` fires after the node goes silent |
| **A3a** | (needs H0 landed) device `BoundedCommand` extraction + `Materialize` on top of it; `DeviceEffectors` slice and start-time ownership check; `--dispatch-lease` and its start check | S-M | `go test ./internal/device/... ./internal/runtime/... ./internal/actions/...`; device-wire behaviour pin unchanged (existing e2e) | device-wire tests untouched and green |
| **A3b** | effector: route ownership, D7 materialization (`show_state`, `set_cooling`), DCMD plan, mailbox, receipt/result table of section 5, authority calls, session-end semantics | L | coverage of `app`; fixtures `lying-result-*`, `expired`, `stale-epoch`, `policy-not-allowed`, `duplicate-*`, `queued-across-reboot`, `host-loss`, `reboot-reconcile` | every row of the section 5 table has a test |
| **A3c** | origin check + `HardwareBoards` profile rule + flag; verification from DDATA (+DSTATE), evidence, `command_evidence` in `HSYNC` | M | `go test ./internal/device/... ./cmd/agentic-stream` | hardware board refused without consent, accepted with it; `verified-result` fixture; second identical `set_cooling` really runs (C2) |
| **A4** | policy digest and epoch logged at start; `startPlantWireExperiment` e2e with the fake node; env-gated real streamsim test; docs (`documentation/operations`, repository map, flags); P2 exit run by the main session | S-M | `make phase-gate` is streamsim's; here `make ci-check` once | P2 exit of the plan (section 6) |

Order: A0 -> A1a -> A1b -> A2 -> A3a -> A3b -> A3c -> A4. A1a/A1b/A2 do not depend on H0; A3a does (H0.1 `RouteOwner`, H0.2 start check). A3 may start with H0 merged, otherwise the host adds a third route name (the thing H0 removes).

## 11. Risks

1. C1: a long approval wait is not representable without raising the dispatch lease; the lease also bounds crash recovery of a `dispatching` row (an abandoned lease becomes an unknown outcome, `actions/.../lease.go:58-64`), so a 3 min lease means a 3 min wait to recover after a crash mid-dispatch.
2. C2: key semantics differ between the plan text, SPEC section 9 and what the node does; get it settled with the other two repositories (the Python gateway) before A3b.
3. Duplicated rules across repositories: admission order, environment rules, sequencing and digest are re-implemented in Go from the SPEC (the reference host is in another repository). The vendored `fixtures/exchanges` are the only shared oracle; keep them in the test suite from A1b.
4. `canonicaljson` vs streamsim `canonical` number formatting (kind digest and folder digest): one disagreement makes every session `inventory_mismatch`. A1a pins it with the fixture vectors first.
5. Reader/sink coupling: if the reader ever calls the sink inline, a slow batch stalls barriers and receipts. The bounded queue's overflow policy (block the reader vs drop historical) must be decided; recommend: never drop live data, drop `historical` first, and refuse the session (`connection_lost`) if the queue stays full for 2 barrier intervals.
6. One effect link per process in the A rounds (Q7); D4 keeps device-wire alive but not beside plant wire in the same process.
7. A hardware-only session sends no `NSYNC` (D-112): evidence never reaches the node; the effect shows `unknown` there. Accepted in the plan.
8. Reconciliation barrier after a board reboot: commands to that device are refused until an operator path exists (Q4). A mid-run reboot at the bench ends the run.
9. The fan lease (5 s) against a 2 s DDATA cadence leaves one or two samples to verify; a late or missing sample fails the verification and marks a working command `failed`/unknown. Tolerable for P4/P6 criteria (command chain, not temperature) but visible.
10. `T` for heartbeat on an unpaced simulated node is world time; mixing it with hardware time in one `mixed` session gives two time bases in one event stream (the situation key is per entity, so only a shared entity would suffer; the plan binds the zone to one side per rung).

## 12. Open questions (owner or other repositories)

- Q1. Idempotency key: forward AS's per-intent key (recommended) or content digest per SPEC text. Needs a SPEC 1.4.0 sentence and agreement from the gateway (Python) and streamsim.
- Q2. `authority_epoch` (integer, never decreasing across restarts, gateway persists the floor): a flag defaulting to Unix seconds at start (recommended), or a durable counter (a migration and a store layer, much more code). Today's owner epoch is an opaque random string (`control` ubiquitous language).
- Q3. Site file location: external (recommended, H0 principle; digest logged) or embedded like `simulator_data.json`. Same for the quality table defaults.
- Q4. Reboot reconciliation clearing path (UNVERIFIED how production clears it): mirror device-wire and leave it open, or specify an operator command in a later round.
- Q5. Historical DDATA: drop from live ingest (recommended) or ingest as late events (`latePolicy: correct_and_reconsider` could re-open situations).
- Q6. `fallback_on_lease_loss` per command: `safe_off` for lease commands (recommended), or a route/catalog member (format change, rejected by the strict loader) or the kind.
- Q7. Two effect links (device-wire and plant wire) in one process: not in A rounds.
- Q8. Extract the owner-only unix listener into one foundation (three copies would exist) as a separate refactor, not in A1.
- Q9. Ask streamsim for a `VERSION` file in the contract folder at 1.4.0 and a `contract-sync`-style note for the other repositories.
- Q10. `SafeStop` for plant wire: needs a production trigger (deadcode) and is not required by the experiment; leave to a round after A4.
