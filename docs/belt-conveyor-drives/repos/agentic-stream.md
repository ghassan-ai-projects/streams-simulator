# agentic-stream for the belt-conveyor-drives use case

Date: 2026-11-07. Method: reading code and docs only (no builds, tests or binaries). Repository:
`/Users/ghassan/my-projects/agentic-stream`, branch `belt-conveyor-drives`.

Conventions: "F" sections are verified facts with `path:line` (paths relative to the agentic-stream
root unless absolute). "I" sections are my inferences and are marked as such. Line numbers refer to the
files as read today; the docs under `documentation/` are descriptions, the code is the authority
(`AGENTS.md:11-14`). Things I did not read in full are flagged "not verified".

## 0. Summary

agentic-stream is the consumer, brain-scheduler and authority plane: it ingests the streamsim plant wire
as events, keeps one deterministic "Situation" per entity, wakes Tamoz (the reasoner) only when a CEL
trigger says so, and lets only policy-approved, catalog-bounded commands reach devices. For the
conveyor it can already: ingest per-drive current/temperature/vibration over plant wire, track a
single drive's degradation with windowed mean/slope/rms features and phases, act on one named device
through a closed capability catalog with leases, approval, interlock and read-back verification, and
promote verified answers into human-approved rules. It cannot today: correlate several entities
(drive 3 degrading AND drives 1/2 rising) inside the deterministic plane, compute time-to-threshold,
or run an ordered multi-step response. The biggest structural fact is that a Situation, its operators
and its intents are all scoped to one entity (section 5). The conveyor therefore has to be modelled as
one entity ("the line") whose events carry a per-drive discriminator, or the engine must learn
cross-entity features. Either is a bounded, but non-trivial, change.

## 1. What agentic-stream is and how it is built

### F1.1 Purpose and posture

- A streaming-native agent runtime: unbounded evidence becomes durable, versioned Situations; a bounded
  agent episode starts only when a deterministic scheduler decides reasoning is useful; agents return
  typed Decisions and Intents; a separate deterministic policy and action plane decides what executes
  (`AGENTS.md:6-9`, `README.md:3-7`).
- Unreleased development snapshot, single node, SQLite WAL (`README.md:9-11`, `README.md:86-97`,
  `documentation/overview/limitations.md:9-14`). No backward compatibility owed (`AGENTS.md:45-47`).
- Ten release-blocking invariants, among them: events are evidence not instructions; time/watermark/
  completeness explicit; Situation versions immutable; model cannot execute effects; policy revalidates
  every Intent right before dispatch; replay never performs effects (`AGENTS.md:21-35`,
  `documentation/architecture/invariants.md:7-21`).
- Forbidden in v1: graph engine, broker, LangChain, web UI, vector retrieval, multi-agent mesh, LLM
  call per event, per-domain code branches ("domains are data, through the SituationSpec")
  (`AGENTS.md:153-160`). This matters for the conveyor: a conveyor-specific code path is not allowed;
  anything needed must be generic or data.

### F1.2 Pipeline (ingestion to action)

The architecture rule states the flow `ingress -> eventlog -> engine/operators -> situations ->
cognition -> episodes -> evidence/decisions -> policy -> actions` (`AGENTS.md:102-104`). Plane table:
ingress/eventlog (normalize, dedupe, quarantine), stream (event-time state, windows, operators,
Situation versions), cognition (decide whether a version deserves an episode), episode/executor (read a
snapshot, propose), policy (revalidate, risk, approval, interlock, epoch), action (lease, dispatch,
idempotency, reconcile, outcome) (`documentation/architecture/overview.md:42-50`).

- Ingress: normalized JSONL files, simulator JSONL adapter, live normalized JSONL on a Unix socket, and
  the plant wire host (`documentation/overview/limitations.md:20-24`,
  `documentation/overview/status.md:12`, `status.md:27`). No Kafka/NATS/MQTT
  (`documentation/overview/status.md:34-36`).
- Stream plane: validate and dedupe, apply event-time rules, windowed features, reduce Situation state,
  publish a version (`documentation/design/stream-processing.md:10-21`). State changes are serialized per
  virtual partition (`stream-processing.md:23-24`).
- Time model: bounded out-of-orderness watermark per partition, per-source clocks, idle timeout, late
  policies `drop_with_audit | history_only | correct | correct_and_reconsider`; clock-skew events are
  refused (`stream-processing.md:39-64`). Completeness is the weakest of the inputs' completeness
  (`stream-processing.md:75-83`).
- Cognition: deterministic admission with trigger (CEL), score, threshold, material delta, debounce,
  cooldown, coalescing, capacity and freshness (`documentation/design/cognition.md:37-55`).
- Episode: binds one immutable snapshot, executor, allowed evidence/tools, allowed intent types, risk
  ceiling, dispatch policy, deadline and a finite budget (`cognition.md:59-66`).
- Decision/Intent: a Decision carries up to 16 intents, 64 facts, 16 alternatives
  (`internal/contractsv1/internal/domain/schemas/v1/decision-v1.json:20-22`); an Intent has type, risk
  class R0-R4, parameters, evidence ids, `expires_at`, optional `compensates`
  (`.../schemas/v1/intent-v1.json:7-17`).
- Policy then outbox then effector then outcome; unknown outcomes stop automatic retry and wait for
  reconciliation (`documentation/design/decisions-and-actions.md:68-77`, `limitations.md:26-35`).

### F1.3 The SituationSpec is the unit of configuration

A SituationSpec (YAML/JSON, canonical digest) has sections `inputs, time, windows, operators,
situation, cognition, actions` (`documentation/contracts/situation-spec.md:16-29`). It is checked by an
embedded JSON schema plus semantic checks (`situation-spec.md:34-44`).

- Inputs: `eventType, schemaVersion, schema, partitionKey, entityType` (all required)
  (`internal/spec/internal/domain/schema.json:127-176`). Event schemas are data
  (`internal/spec/internal/domain/event_schema_data.json`); `AGENTS.md:153-158` forbids re-authoring
  domain data in Go. The built-in schemas include `motor.current.observed`, `motor.temperature.observed`,
  `motor.vibration.observed`, `motor.heartbeat.observed`, `pump.*` families
  (`event_schema_data.json:107-141`, `:232`); `--event-schemas` may add schemas
  (`AGENTS.md:155-156`).
- Operators: exactly three kinds: `aggregate`, `slope`, `missing_heartbeat` (`schema.json:262-267`);
  aggregates `mean, rms, slope, count, sum, min, max, latest` (`schema.json:293-303`); an optional
  per-operator `where` CEL filter (`schema.json:314-316`; compiled at `cel.go:56-58`); an optional
  `fields` list of 2-8 fields (`schema.json:279-286`, semantics not verified).
- Windows: only `tumbling` and `sliding`, emit `on_update | on_close | early_and_close`
  (`schema.json:214-224`). `stream-processing.md:87-90` also names count and decay windows; the schema
  enum does not list them (doc and schema disagree, schema wins).
- Situation: one `type`, one `entityKey`, occurrence open/close CEL, phases with severity, transitions
  with `minDuration` (hysteresis/debounce), reducers (`schema.json:393-430`;
  `examples/real-world-sensor/zone-thermal-sim.situation.yaml:129-203`).
- CEL environment exposes only `features, situation, delta, event_time, watermark` plus `ext.Bindings`
  (`internal/spec/internal/domain/cel.go:13-27`). No cross-entity variable, no user functions.
- Cognition: triggers with `when`, `score`, `threshold`, `lane (fast|deep)`, `debounce`, `cooldown`,
  `expiresAfter`, `materialDelta`; an executor block with name (`tamoz | fixture | native | rule`),
  `dispatchPolicy (active|shadow)`, `riskCeiling`, objective, prompt, tools, budget, diagnosis catalog
  (`zone-thermal-sim.situation.yaml:205-262`; `schema.json:539-545`, `:618-621`).
- Actions: the closed intent vocabulary: per intent `risk`, `parameterSchema`, `modelWritableFields`,
  `presets`, `policy (automatic|approval)`, `rateLimitPerHour`, optional `baseline.byPhase`
  (`zone-thermal-sim.situation.yaml:264-351`; `situation-spec.md:68-93`).

### F1.4 Identity: Situation = tenant x deployment x partition x type x entity

"The engine derives the Situation identity from a stable combination of deployment, tenant, partition,
Situation type and entity" (`documentation/learn/domain-model.md:39-41`; `limitations.md:69-78`). Every
operator state is keyed by entity (`internal/operators/internal/domain/types.go:32-38`, the `State`
comment "keyed by entity"). A Situation cannot read another entity's features: the CEL `features` map is
built from one Situation's own facts (`internal/engine/internal/domain/situation_features.go:8-17`,
`situations.CELFeatures(compiled, state.Facts, state.Evidence)`). Confidence is a constant 1.0 and
hypotheses are not Situation state (`domain-model.md:85-98`, `limitations.md:69-78`).

## 2. How it detects, predicts and acts

### F2.1 Detection (deterministic, no model)

Detection is windowed features plus CEL transitions with `minDuration`. Worked example: the thermal
spec opens on `temp_mean_5m > 30 && temp_slope_5m > 6 && ambient_count >= 2 && ambient_slope < 4`, and
the "ambient discount" is a plain CEL conjunction across two input streams of the same entity
(`zone-thermal-sim.situation.yaml:133-137`, `:156-180`). Heartbeat loss makes completeness uncertain and
gates the cognition trigger (`:209-211`; `stream-processing.md:66-67`). The motor example
(`examples/predictive-maintenance/predictive-maintenance.situation.yaml`) uses a 15 min vibration RMS
window, temperature 30 min, a 6 h trend window and phases candidate/watch/warning/incident/recovering/
resolved (operator and phase names from `grep` at lines 51-139; contents not read in full).

### F2.2 Prediction

There is no prognostic operator. The available deterministic building blocks are windowed
`slope` (units chosen by the spec, e.g. `celsius_per_hour`) and arithmetic in CEL
(`zone-thermal-sim.situation.yaml:91-98`; `cel.go:13-27`). A grep of `internal/spec`, `internal/operators`
and `examples` for remaining-life or extrapolation vocabulary finds only the unrelated playbook code
(no operator, no feature). Prediction in the product is therefore "trend features + Tamoz judgment",
not a calibrated RUL model. Confidence is not computed (`domain-model.md:85-91`).

### F2.3 Acting: intents, risk routes, approval

- Risk classes: R0 and R1 route `automatic`, R2 routes `approval`, R3 and R4 are `denied`
  (`internal/contractsv1/internal/domain/risk.go:26-31`). `policy: approval` can only raise R0/R1 to
  approval (`risk.go:76-82`); `automatic` never relaxes the route
  (`documentation/contracts/situation-spec.md:69-72`). A spec cannot make R3 executable at all
  (`limitations.md:61-67`).
- Approval: Ed25519-signed through a relay, approver authority by entity, risk class and role, bound
  to the intent and exact decision bytes (`documentation/design/rule-promotion.md:80`;
  `internal/policy/README.md:66-80`). Authority by role only covers R0-R2
  (`internal/policy/internal/domain/principals.go:149`).
- Policy re-checks right before dispatch: owner, epoch, situation freshness, completeness (incomplete
  source health denies consequential classes), expiry, risk, interlock, hourly rate
  (`internal/policy/README.md:46-55`; `internal/policy/internal/domain/routing.go:30-58`;
  `internal/contractsv1/internal/domain/risk.go:92-99`).
- Freshness: an Intent goes stale (`situation_version_stale`) only when a newer version satisfies the
  trigger's `materialDelta`, so benign provisional versions do not cancel it
  (`decisions-and-actions.md:53-67`).
- The interlock is a durable global switch that blocks command creation and effect delivery; it is
  read-only for ordinary paths, and "no production path calls [Set] yet"
  (`internal/interlock/UBIQUITOUS_LANGUAGE.md:5-12`, with the `Set` caveat as printed there;
  `limitations.md:85-93` says the trip is the exception that works while running).
- Model isolation: the worker gets scoped read-only evidence tools only, no effector handle
  (`documentation/architecture/worker-boundary.md:34-48`, `:55-64`).

### F2.4 Command path to devices (capability catalog)

- A closed capability catalog maps an intent type to a device route:
  `{operation, target, target_bindings, selector_field, expires_after_ms, presets, bounds, target_arg}`
  plus `safe_stops` (`internal/device/internal/domain/catalog.go:17-42`). The model writes only an
  enum selector (`modelWritableFields`); concrete values (duty, duration) come from deterministic
  presets within `bounds` (`zone-thermal-sim.situation.yaml:325-350`;
  `examples/real-world-sensor/testdata/plant-wire/thermal-bench.catalog.json:9-28`).
- Resolving a target: the situation's entity must be bound to a device in `target_bindings`, otherwise
  refusal (`internal/device/internal/domain/materialize.go:40-52`, exercised by
  `internal/device/internal/domain/fleet_route_test.go:19-37`). The model can never name the device
  (`internal/device/UBIQUITOUS_LANGUAGE.md:30`). An intent's `entity_id` or `target` parameter must
  equal the dispatched episode's entity (`internal/decisions/internal/domain/intent_parameters.go:28-56`).
- Plant wire effector: one `DCMD` per approved command; the idempotency key is forwarded unchanged;
  the deadline is in the node's time (world time for barrier-paced worlds) plus the route expiry
  (`documentation/design/plant-wire-adapter.md:134-151`). `authority_epoch` from `--plant-wire-epoch`;
  every command falls back `safe_off` on lease or host loss (`plant-wire-adapter.md:146-149`).
- Safe stop over plant wire is NOT implemented: "wait for a production need"
  (`plant-wire-adapter.md:287-288`). Device-wire (serial) has `SafeStop` with a priority lane and a
  latch (`internal/device/effectors.go:69-73`; `internal/device/internal/app/session_safety.go:18-40`);
  my grep finds no caller of it outside the device module, telemetry and plantwire inventory checks
  (`internal/plantwire/internal/app/inventory.go:55-67` only validates safe stops).
- Consent per board: a command to a hardware board needs the physical profile plus `--live-actuation`,
  `--owner-authorized` and the board in `--hardware-boards`; checked twice; simulated boards are
  commanded freely (`plant-wire-adapter.md:115-132`).
- Verification is by the device's own readings: a succeeded `DRESULT` counts only when a good,
  non-historical reading of the commanded writable metric matches within resolution inside 6 s; a
  receipt is never counted as done; unknown outcomes open a reconciliation and block the device until
  readings clear it (`plant-wire-adapter.md:200-248`).
- Dispatch lease: `--dispatch-lease` (default 1m) bounds one command; a call open after 0.9 of it is an
  unknown outcome; `serve` refuses a catalog route whose `expires_after_ms` does not fit
  (`documentation/reference/cli.md:138`; `plant-wire-adapter.md:262-271`).

### F2.5 Playbooks and rule promotion (human approves every rule)

- Purpose: a recurring, verified case (same diagnosis code, device kind digest, intent, preset) becomes
  a deterministic rule that answers without a model call (`documentation/design/playbook.md:8-16`).
- Owner decisions: a human approves every new rule, nothing promotes itself (decision 1); rule has CEL
  `when`, `envelope`, `recommend` (intent+preset only, no free values), `mode`, provenance (decision 2);
  pinned to the kind digest (3); per-rule circuit with causes `audit_disagreement, rate_cap,
  kind_changed, no_effect, out_of_cap` (5); draft after 3 verified cases (6); compiler refuses any rule
  recommending an R3+ intent, one marked `emergency` or marked `safetyFunction` (17); a case is verified
  only if the action took effect, the situation stayed closed for the diagnosis code's settle window,
  and (decision 19) the reading returned to its pre-fault level (`documentation/design/rule-promotion.md:35-51`,
  `playbook.md:46-57`, `:122-157`).
- Flow: admission asks `playbook.Route` in the same transaction: rule, audit, or model
  (`playbook.md:92-109`). Rules start in `shadow`, move to `active` after 10 agreeing and 0 disagreeing
  shadow matches (`playbook.md:49-50`, `:195-198`); 1 in 3 active fires are re-audited by a shadow Tamoz
  episode, a disagreement opens the circuit (`playbook.md:127-135`, `:150-154`). Tamoz drafts through a
  separate `DraftPlaybookRule` call; the runtime checks, replays and stores a pending proposal; a human
  approves on the signed endpoint (`playbook.md:170-196`). The playbook is a customer file, never in a
  repo, read at start (`playbook.md:35-44`).
- Implemented per roadmap (`documentation/roadmap.md:21-22`; `status.md:28`).
- Limits relevant here: the playbook is a per-entity pattern. A fleet cap per rule (decision 18) is
  marked "after the first slice" (`rule-promotion.md:49`).

### F2.6 Relationship to Tamoz

- "Tamoz is the only production reasoner. Agentic Stream owns time and authority; Tamoz, in its own
  repository, owns judgment." `fixture` is a demo executor, `native` an in-process loop for tests and
  baselines, `rule` the playbook executor (`AGENTS.md:57-61`). Tamoz is a separate-process Go-protocol
  gRPC `EpisodeWorker` over a Unix socket (optional mTLS) that can call back `EvidenceTools` for scoped
  read-only evidence; the contract is `proto/agenticstream/runtime/v1/runtime-v1.proto`
  (`documentation/architecture/worker-boundary.md:5-30`, `:66-72`).
- Shared contract copies: the decision schema is vendored on the Tamoz side
  (`documentation/design/rule-promotion.md:69-72`, `:135-136`).
- Tools the episode may use in the experiment spec: `features_query, evidence_get,
  history_prior_incidents, knowledge_search` (`zone-thermal-sim.situation.yaml:244-248`). Evidence queries
  are "tenant/entity/time/row-scoped" (`internal/evidence/README.md:11` table row for Transport,
  `README.md` Public operations "EventLogQuery").
- I did not read Tamoz itself; its side is covered in the tamoz repo note.

## 3. The thermal LED/fan experiment: end-to-end path

Reference documents: `examples/real-world-sensor/README.md`, the spec
`examples/real-world-sensor/zone-thermal-sim.situation.yaml`, the plant wire adapter note, the
experiment tests `cmd/agentic-stream/experiment_*_test.go`, the research-lab handoff
`/Users/ghassan/my-projects/streams-simulator/docs/physical-devices/fan-experiment/HANDOFF.md` and
`/Users/ghassan/my-projects/agent-research-lab/real-world-sensor/assessment/CURRENT-STATUS.md`
(dated 2026-09-09, partly superseded by the handoff).

### F3.1 Data path (world to stream to Situation)

1. streamsim emulates a thermal board (temperature, humidity, ambient, fan tach, heartbeat) as a plant
   wire node and dials the host's owner-only Unix socket (`serve --plant-wire-socket`,
   `plant-wire-adapter.md:54-60`; `documentation/reference/cli.md:151`).
2. The host admits `NBIRTH` (protocol, environment, contract version+digest, inventory), commissions kinds
   by kind digest, then reads `DDATA`; every `NSYNC` is answered with `HSYNC` only when the events before
   it have left the ingest queue and the consumer has no reasoning in flight, so a barrier-paced world
   cannot outrun reasoning (`plant-wire-adapter.md:60-93`).
3. A site file (outside the repository, `--plant-wire-site`) maps device to entity, `(kind, metric)` to
   event type and field, heartbeat rules and wire-quality to event quality; the module contains no
   metric, event type or entity-type word (`plant-wire-adapter.md:95-104`;
   `examples/real-world-sensor/testdata/plant-wire/thermal-bench.site.json:1-30`;
   `internal/plantwire/internal/domain/site.go:19-52`).
4. Readings become CloudEvents-style envelopes with `PartitionKey = Entity.ID` and data
   `{quality, boot_id, seq, schema_version, <field>}`
   (`internal/plantwire/internal/domain/events.go:44-52`). Note: the envelope does NOT carry the device
   name in `data` (see gap G1).
5. Ingress admits or quarantines like any source; engine operators compute windowed features per entity;
   reducers write facts; the Situation phase machine advances (`zone-thermal-sim.situation.yaml:62-203`).
6. Cognition trigger `overtemp_needs_decision` admits a deep-lane episode on phases
   watch/cooling/over_ceiling with debounce 5 s and cooldown 20 s (`:205-224`).
7. Tamoz is the executor with `dispatchPolicy: active`, `riskCeiling: R2`, budget 30 s wall time and
   5 model calls, and an objective that explicitly requires an independent tach confirmation and forbids
   treating a commanded fan as a cooled zone (`:225-262`).

### F3.2 Control path (decision to device and back)

1. Tamoz returns a Decision with typed intents from the declared vocabulary: `install_watch_condition`
   (R0, automatic), `set_indicator` (R1, automatic, enum `off|watch|alert`), `select_thermal_mode` (R2,
   approval, enum `hold|bounded_cooling`) (`:264-351`). The model writes only the enum.
2. Validation binds the Decision to episode/attempt/fence/snapshot; policy then routes by risk. The
   experiment catalog turns `select_thermal_mode: bounded_cooling` into `set_cooling level_percent 45`
   bounded 0-60, target `fan-01` bound for entity `zone-01`, route expiry 20 s; `safe_stops` for the LED
   and fan (`thermal-bench.catalog.json:2-28`).
3. After approval (human for R2, or a scripted approver in the test) the action plane dispatches one
   `DCMD`; the node answers `queued/accepted/refused`, then `DRESULT`; the outcome ledger records pending
   verification; verification succeeds only if the twin's `level_percent` reading becomes 45 inside the
   6 s window (the fan lease is 5 s in the twin, so 0 after) (`plant-wire-adapter.md:158-166`, `:200-216`).
4. The fan experiment test (simulated rung) runs the committed `fan-experiment/simulate.plan.json` against
   the real streamsim node, waits for succeeded commands on routes `set_indicator` and
   `select_thermal_mode`, asserts twin read-back (`state=alert`, `level_percent=45`) and zero quarantined
   events (`cmd/agentic-stream/experiment_plantwire_fan_streamsim_test.go:19-37`).
5. Status per the handoff (2026-10-21): the simulated rung passes all seven criteria of the plan; the
   real-board rungs P5/P6 are waiting for the owner at the bench; open item W15: a refused command
   leaves its verification `awaiting`, making `verify-run` say `fail` on a clean run
   (`/Users/ghassan/my-projects/streams-simulator/docs/physical-devices/fan-experiment/HANDOFF.md:1-60`,
   lines in section 1 and 2). Earlier physical evidence: operator-observed fan rotation and safe-stop via
   direct serial, no tachometer (`.../assessment/CURRENT-STATUS.md`, executive verdict paragraph).
6. The experiment is declared acceptance test #1: `AGENTS.md:51-56` lists what must not break (CLI flags,
   spec grammar, compiled digests, thermal catalog, `runtime-v1.proto`, notification goldens, device-wire
   records, run-export format); guards are `cmd/agentic-stream/experiment_*_test.go` and
   `internal/notify/experiment_contract_test.go`. Any change for the conveyor must not move those
   (inference: add new files, do not edit the thermal spec or catalog).

### F3.3 Pump station: the multi-entity precedent

`examples/pump-station/` has eight pumps, one `pump_condition` Situation per pump, per-pump relative
features (15 min mean against the pump's own 24 h mean), fleet routes where each intent commands the
pump bound to the situation's entity, `schedule_maintenance` (R1, one preset per work order),
`reduce_load` (R2, approval) and `stop_motor` (R3, `emergency`, never a rule target)
(`examples/pump-station/README.md:5-29`; `documentation/design/rule-promotion.md:143-160`). The test
`cmd/agentic-stream/experiment_pump_station_streamsim_test.go` drives it against streamsim (107 lines,
not read). This is the closest existing shape to three drives, but note it is "N independent copies",
not a coupled system: no feature of pump A ever depends on pump B.

## 4. What it can already do for the conveyor use case

| Need | Status | Basis |
|---|---|---|
| Ingest 3 drives x (current, torque, winding temp, vibration, speed) + belt speed + feed rate over plant wire | Works, with site/kind files per drive | F3.1; section 5 for entity caveat |
| Per-drive degradation features (RMS vibration, temp slope, current mean) with hysteresis and minimum durations | Works | F1.3, F2.1 |
| Sensor-fault awareness: bad quality maps to event quality, heartbeat loss makes completeness uncertain and gates reasoning | Works (stuck value not detected) | `site.go:19-27`, `plant-wire-adapter.md:95-100`, `stream-processing.md:66-83`; no stuck/flat-line operator in the three operator kinds (`schema.json:262-267`) |
| Cross-entity correlation (drive 3 degrading AND drives 1/2 rising) | NOT supported in the deterministic plane | F1.4; only a model episode could compare, and its evidence tools are entity-scoped |
| Feed surge vs one-drive failure | Only if all three drives live in one Situation (see I5.1). Otherwise a model would have to infer it from one drive's data | F1.4 |
| Trend/RUL | Slope features only; no time-to-threshold, no calibrated confidence | F2.2 |
| Coordinated steps (cut feed, then stop drive 3 at a chosen time) | Partially: up to 16 intents per Decision, each independent | decision-v1.json:22; no ordering/dependency fields in the intent schema (grep for `depends/after/sequence` in `schemas/v1/*.json` returned nothing) |
| Timed action ("stop at a chosen time") | Not as a primitive; a route has only `expires_after_ms` and a preset, no scheduled-at; a `watch` (R0) can install an expiring condition that fires on evidence | catalog.go:22-31; `internal/watch/README.md:3-5` |
| Safety: lease on commands, expiry on intents, interlock, owner epoch, approval for R2, R3+ denied, hardware consent, read-back verification, unknown-outcome reconciliation | Works | F2.3, F2.4 |
| Device-side safe stop over plant wire | Not implemented | `plant-wire-adapter.md:287-288` |
| Human-approved rules for recurring cases | Works per entity | F2.5 |
| Determinism, replay without effects, run artifacts and verdict | Works | `AGENTS.md:30-35`; `status.md:23`, `:29` |

## 5. Gaps for the conveyor, with where they live and rough size

Sizes: S = under a day, M = a few days, L = a week or more, in this repository's gate-heavy workflow
(15-line functions, 300-line files, architecture tests).

### G1. Several devices cannot share one entity without losing the device identity (M, plantwire + site file)

Fact: a site device carries at most one `entity`; readings are keyed by `(kind, metric)` and emit one
event type per key (`internal/plantwire/internal/domain/site.go:19-27`, `:53-73`); the event data holds
no device name (`events.go:50-52`). Consequence: if drives 1, 2, 3 are three devices of the same kind
mapped to entity `conveyor-1`, their `current` events share one event type and one operator window and
are merged; a spec cannot tell drive 3 from drive 1.
Options (inference): (a) one kind and one event type per drive position (`drive_1`, `drive_2`,
`drive_3` kind files, event types `drive1.current.observed` and so on). Pure data, no code change, but
triples kind files, event schemas and operators. (b) add the device name to the observation `data`
(`events.go:50-52`) and let operators partition by it. This changes the engine's partition model and
the event schemas, i.e. touches the thermal pins; large. Option (a) is the cheapest; its ugliness is
data duplication only.

### G2. No cross-entity features: the heart of "the failing motor notifies the others" (M to L, engine/operators/spec)

Fact: operators, Situations, CEL `features` and intents are all one-entity (F1.4, `cel.go:13-27`).
Two routes (inference):
- Model the line as one entity (G1 option a). Then "drive 3 degrading AND drives 1/2 rising" is an
  ordinary CEL conjunction in `transitions.when`, exactly like the thermal ambient discount
  (`zone-thermal-sim.situation.yaml:156-161`). Size M, data only. Downside: the Situation is per line,
  so the per-drive phase is encoded in facts, not in independent per-drive Situations.
- Generic cross-entity operator (e.g. an operator that aggregates "the other entities of this group"):
  new operator kind, group concept in the spec, engine/state/restore/timer changes, replay and
  playbook consequences, new invariants proof. Size L. It also conflicts with `AGENTS.md:21-31`
  (invariant 4: serial per partition) unless the group is the partition, which is route one again.
Recommendation (inference): route one for the first scenarios.

### G3. No time-to-threshold / RUL feature (S to M, spec CEL or operators)

Fact: only windowed slope and CEL arithmetic (F2.2). Inference: a ramp-extrapolation is
expressible today as a CEL expression over two features (`(limit - level)/slope`) inside transitions and
trigger scores, and can be reduced into a fact only if it is an operator output (reducers take operator
inputs, `zone-thermal-sim.situation.yaml:181-204`). A published "time to limit" feature therefore needs
either a small generic operator (e.g. `extrapolate`) or the Tamoz episode doing the arithmetic from
`features_query`. A new operator kind needs schema, compiler, operator runtime, state, tests and docs;
size M. Calibrated confidence is explicitly absent (`domain-model.md:85-91`), so any RUL carries no
uncertainty from this repo.

### G4. Pull torque/current vs belt-speed normalization (S, spec)

Inference: drive current rises with feed; to tell "feed surge" from "drive 3 failing" one needs ratios
of features, which CEL can do (`ambient discount` precedent) provided the inputs share an entity (G1
route a). No code beyond data.

### G5. Ordered multi-step response with a delay (M to L, decisions + actions + policy)

Fact: intents in one Decision are independent; no ordering field; each passes policy and rate limits
separately (`intent-v1.json`, `decisions-and-actions.md:68-77`). A freshness rule can cancel a later
step: an Intent goes stale if a newer material version arrives (`decisions-and-actions.md:53-67`), which
is helpful (don't stop drive 3 if the situation resolved) but also means a "cut feed, then 30 s later
stop drive 3" plan can lose its second step as soon as the cut feed changes the Situation materially.
Options (inference): (a) two separate episodes: step 2 is proposed by a new episode after step 1 is
verified (no code, but needs the trigger to be written so the second wake happens; M data/spec work);
(b) a generic "after intent X verified" or "not_before" field in the intent and command (`not_before_ns`
exists on the wire and is currently unset, `plant-wire-adapter.md:143-145`): new contract field,
schema, policy and action changes; L.

### G6. "Stop at a chosen time" and the restart trap (M, catalog + spec)

Fact: the catalog's `presets` are static; `expires_after_ms` is a lease, not a schedule
(`catalog.go:22-31`). A drive stop is an actuation that most plants would class above R2; R3/R4 are
denied outright (`risk.go:26-31`) and the playbook compiler also refuses `emergency` and
`safetyFunction` intents (`rule-promotion.md:49`). So a safe automatic drive stop must be declared
R1/R2 (a software-level mode, e.g. "controlled ramp-down request" to the VFD), with approval for R2,
which is a domain decision for the customer (inference). The simulator-side restart trap must be
expressed in the world, not in agentic-stream.

### G7. Safe stop over plant wire (M, plantwire)

Fact: not implemented, deferred "for a production need" (`plant-wire-adapter.md:287-288`). The
catalog already supports `safe_stops` and the fan example declares them
(`thermal-bench.catalog.json:3-6`), but the plant wire effector only falls back `safe_off` through
leases. For a conveyor where "host lost" must mean a defined belt state, this is a real gap. Wire
contract change would be needed in streamsim too (the contract is vendored byte-identically,
`plant-wire-adapter.md:48-52`).

### G8. Fleet caps and group-level rule promotion (L, playbook)

Fact: decision 18 (rule fires on at most N entities in a window, otherwise one fleet-level case) is
"after the first slice" (`rule-promotion.md:49`). With the line as one entity (G2 route one) this is
mostly moot, but a plant with several conveyors needs it.

### G9. Operator tooling is CLI over a stopped runtime (S to M, outside the core)

Fact: approvals have a signed HTTP endpoint; quarantine redrive, circuit close, reconcile, inspection are
CLI commands that need the runtime owner lease, i.e. the runtime stopped (the interlock trip is the
exception); no HTTP inspection API (`limitations.md:85-93`; `status.md:44-45`). For a live conveyor
demo an operator view of "why did it stop drive 3" is `explain`/`situation`/`episode`/`intent` CLI
only.

### G10. Single node, local ingress only (not a gap for the sim, flagged for the customer)

Fact: no broker ingress, loopback server only, no distributed state (`limitations.md:18-24`, `:37-43`;
`roadmap.md:36-45`). Plant wire is the integration path for real devices.

## 6. Contract and protocol expectations from streamsim

### F6.1 What streamsim must speak

- Plant wire v1 session protocol (node dials the host; births, data, `NSYNC/HSYNC`, `DCMD`,
  receipts and results). The contract folder is vendored byte-identically from streamsim and pinned by
  version and digest; tests fail on any drift; `make plant-wire-sync` copies from
  `PLANT_WIRE_CONTRACT_DIR` (`plant-wire-adapter.md:48-52`; `AGENTS.md:78`). Agentic-stream never edits
  the copy.
- Device kinds: each birth announces a `kind_digest`; the host commissions kinds from kind files
  (`--plant-wire-kinds`, directory of `*.kind.json`); a kind whose digest differs is quarantined (no
  events, no commands) (`plant-wire-adapter.md:69-71`; `cli.md:153`). A directory holds one definition
  per kind name (`examples/pump-station/README.md:44-45`).
- Kind file shape: `{kind, metrics[{name, data_type, unit, eu_range, resolution, writable}], commands[{name,
  args_schema, risk_class, safe_stop}]}`
  (`examples/real-world-sensor/testdata/plant-wire/kinds/fan.kind.json:1-24`).
- Site file: devices to entity, readings to event types, heartbeats, quality map
  (`thermal-bench.site.json`; `site.go:19-52`).
- Capability catalog: routes and safe stops (F2.4). `serve` requires the catalog's targets and safe stops
  to be devices of the site file, and the site's devices to be announced by the node
  (`plant-wire-adapter.md:279-283`).
- Policy digest: `serve` logs it; the node's session plan must list it (`policies`)
  (`plant-wire-adapter.md:262-264`; `experiment_plantwire_fan_streamsim_test.go:55-57`, `:69-82`).
- Two-clock handling: commands carry node-time deadlines; barrier-paced worlds hold time between
  `NSYNC`s, so route `expires_after_ms` must exceed pacing steps (pump station uses 3 min and
  `--dispatch-lease 5m`, `examples/pump-station/README.md:14`).
- Alternative ingest paths: normalized JSONL, simulator JSONL adapter (`cli.md`, `limitations.md:20-24`;
  `internal/ingress/internal/domain/simulator_event.go:106` maps simulator events with
  `PartitionKey = ident.entityID`).

### F6.2 Does a multi-drive conveyor domain fit without changes?

Fits without agentic-stream code changes (inference from F3, F6.1):
- Kind files per drive type or per drive position, a site file, an event-schema file via
  `--event-schemas` (`AGENTS.md:155-156`), a SituationSpec, a capability catalog, and a principals file
  for approvals. All of these are data.
- The line as a single entity `conveyor-1` with distinct event types per drive position (G1 route a), CEL
  cross-drive conditions in phases (G2 route one), intents `cut_feed`, `reduce_feed`, `shift_torque_bias`,
  `request_drive_stop`, `set_indicator` as catalog routes with single targets (feeder, three drives) all
  bound to entity `conveyor-1` in `target_bindings`. `Bound` requires the situation's entity in the route's
  bindings and allows a single `target` route (`materialize.go:44-50`), so one entity can reach several
  devices through several routes (inference from that code).

Does not fit without changes: independent per-drive Situations that look at each other (G2 second
option); a published time-to-threshold feature (G3); ordered/timed multi-step actions (G5, G6); safe stop
on the plant wire (G7).

## 7. Inferences and open questions (all unverified)

- I7.1 The thermal example is the right template for the first conveyor spec: ambient discount maps to
  "feed discount" (all drives up with feed), tach confirmation maps to "belt speed / pulley speed
  confirms a shed load", and the deep Tamoz lane maps to the failure-mode diagnosis catalog
  (`zone-thermal-sim.situation.yaml:225-262`).
- I7.2 With one entity per line, `entity_id` binding (`intent_parameters.go:28-56`) is trivially
  satisfied for all routes, which keeps the model unable to name a device (`UBIQUITOUS_LANGUAGE.md:30`
  of `internal/device`).
- I7.3 `rateLimitPerHour` is per intent type in event time; if feed cutting is a hot loop, the limit must
  be set high enough yet bounded (`internal/policy/README.md:56-58` says the rate is counted in event
  time). Not verified for what happens on a barrier-paced world that jumps minutes.
- I7.4 The restart trap and "stop at a chosen time" depend on what the customer allows; the repository's
  hard rule is R3+ is never executable (`risk.go:26-31`). A drive stop that the plant treats as a safety
  action will need either approval (R2) with a human in the loop, or to be handled by the plant's own
  protection with agentic-stream only advising.
- I7.5 Not read: `internal/cognition` scoring internals, the episode scheduler's concurrency rule (the
  rule-promotion note says one live episode per situation,
  `documentation/design/rule-promotion.md:72`, which also serializes steps G5(a)), `internal/engine`
  timers, Tamoz's tool implementations, the full predictive-maintenance spec, and the pump-station
  experiment test.
- I7.6 If the conveyor should be one Situation per drive (cleaner alerts per drive), G2's generic
  cross-entity route is the structural fix. That is a design change under `AGENTS.md:13`
  ("a deviation needs a design change first") and `:106-108` architecture rules; plan it as a design
  note before code.

## 8. Practical sequence for the use case (inference, not a commitment)

1. Data-only slice: three drive position kinds, a feeder kind, a belt kind; site file with
   `conveyor-1`; event schemas; spec with phases `normal, drive_stress, drive_degrading, shed_load,
   stopped`; catalog with feeder/drive routes; scripted approver; run against a streamsim conveyor
   domain with sealed truth. Proves G1a, G2 route one, G4.
2. Add a generic `extrapolate`/time-to-threshold operator only if scenario 1 (slow bearing failure)
   shows CEL arithmetic and Tamoz arithmetic are not enough (G3).
3. Decide with the customer which actions are allowed and at what risk; design G5/G6 only after that.
4. Keep every change additive: do not alter the thermal spec, catalog or goldens (`AGENTS.md:51-56`).
