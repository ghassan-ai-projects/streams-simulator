# Impact analysis — Agentic Stream

> **Status note (2026-10-12):** parts of this document are superseded by the owner decisions in [REVIEW-2026-10-12.md](REVIEW-2026-10-12.md) §1–§2 and tracked in [FINDINGS.md](FINDINGS.md). The kill-switch chain (§6.1) and authority revocation are dropped or deferred; the remaining items belong to the next goal.

- **Date**: 2026-10-12
- **Status**: analysis; **no code changed in Agentic Stream or here**
- **Scope**: what the simulation-canvas direction ([plant-wire-protocol.md](plant-wire-protocol.md),
  [reference-scenario-factory-utilities.md](reference-scenario-factory-utilities.md)) means for
  `/Users/ghassan/my-projects/agentic-stream`
- **Decided**: Tamoz + Agentic Stream are tested as **one system**; Agentic Stream's own policy,
  interlocks and approvals are part of what is scored.
- **Evidence**: `internal/ingress` (README, `simulator.go`, `simulator_data.json`), `internal/device`
  (`catalog.go`, `effectors.go`, UDS transport), `internal/spec`, `internal/policy`,
  `internal/interlock`, `internal/notify`, `internal/replay`, `examples/*/*.situation.yaml`,
  `internal/api/internal/transport/http.go`

## 1. Verdict

Agentic Stream carries most of the integration work. It becomes the **host** of the plant
wire: it discovers devices from births instead of hand-kept tables, keeps a device inventory,
and turns Tamoz's intents into plant-wire commands through its existing policy. Its
`SituationSpec` stays the per-scenario pack and stays hand-written. Three existing pieces are
removed, two are reshaped, and several design questions are specific to Agentic Stream (§6).

## 2. Integration today

| Area | Today | Tied to the simulator by |
|---|---|---|
| Observations in | `ingress.ReplaySimulator` reads `trace-record-v0.1` JSONL (`runtime_config` … `trace_end`) | `simulator_data.json`: channel → payload field (`vibration→rms_mm_s`, …) |
| Commands out | `device` gateway over a Unix socket to `streamsim device serve`; device contract v1 (command, receipt, result, state) | Capability catalog (routes, presets, bounds, safe stops) matched by digest on both sides; operations `set_led`, `set_pwm_lease`, `safe_stop` |
| Scenario knowledge | `SituationSpec` YAML: inputs, time, windows, operators, situation, cognition (prompt, tools, diagnosis catalog), intents | Input event types and payload fields (`data.rms_mm_s`) per channel |
| Safety | `policy` (approvals, principals), `interlock` (durable global switch that blocks the action plane) | — |
| Humans | `notify` (Channel-B notification outbox, HTTP delivery) | — |
| Health | `/health/live`, `/health/ready` (no dependency details by design) | — |
| Event schemas | `internal/spec/internal/domain/event_schema_data.json`, **embedded** in the binary: one entry per event type (`pond.ph.observed/1.0`, `pump.axial_vibration.observed/1.0`, …) | Every new world's event types — today a new world needs an Agentic Stream release |

## 3. Changes required

### 3.1 Plant-wire host connector (replaces the simulator trace path)

- Session handling: `node_birth`, `device_birth`, `device_data`, deaths, `host_state`,
  `rebirth_request`; sequence tracking with gap, duplicate and reorder detection.
- **Device inventory** keyed by (kind, id), built from births, with metric metadata and
  announced commands.
- **Quality codes** (`good`, `stale`, `bad`, `uncertain`) mapped onto the envelope's existing
  `quality` annotation.
- Store-and-forward on reconnect after `host_state` offline.
- Conformance tests against the shared plant-wire fixtures (the existing `contractsv1`
  conformance pattern).

### 3.2 Event mapping (removes `simulator_data.json`)

| Plant wire | Envelope |
|---|---|
| Device kind, id | `entity {type, id}` |
| Kind + metric | `type`: `<kind>.<metric>.observed` |
| Value, unit | `data {value, unit}` |
| Event time | `event_time`; `observed_at` set on arrival |
| Quality | `quality` |

Schema registry: one generic measurement schema per value type (number, integer, boolean,
enum, string), replacing the embedded per-event-type registry in
`internal/spec/internal/domain/event_schema_data.json`. This is what makes a new world
possible without an Agentic Stream release. `SituationSpec` operators
then read `data.value` rather than per-channel field names.

### 3.3 Command path (reshapes the capability catalog)

```
Tamoz intent ─▶ policy re-check (ADR-038) ─▶ route ─▶ plant-wire command(kind, id, name, args)
```

| Today (capability catalog) | After |
|---|---|
| Which operations exist | **From births** (announced commands) |
| Routes: intent → operation + target binding | **Stay**, in the `SituationSpec` / policy; targets resolved against the inventory and constrained to a kind |
| Presets and bounds | **Stay**, in policy; bounds may only narrow what the birth announces |
| Safe stops | Replaced by the declared privileged `emergency_shutdown` command (see §6.1) |
| Digest match with the simulator | Removed; replaced by the expected-inventory check (§3.4) |

Receipts and results become evidence records, so a result that claims success while the plant
did nothing is visible to the next Situation.

### 3.4 Policy additions

- **Default deny** for any announced command not named in policy.
- **Expected inventory** (commissioning list): a device whose birth claims an unexpected kind
  or id is quarantined and its commands denied. Kind-based policy is otherwise exploitable by
  a device that claims a more permissive kind.
- The policy file and `SituationSpec` digest are recorded with each run as **test inputs**.

### 3.5 Status for the canvas

`/health/*` deliberately exposes no dependency details. The canvas needs a separate,
read-only status view: plant-wire session state, last data time, Tamoz worker handshake and
last episode (see [impact-tamoz.md](impact-tamoz.md) §6), interlock state, pending approvals.

### 3.6 Removals

- `ReplaySimulator` and the `trace-record-v0.1` simulator format.
- `internal/ingress/internal/domain/simulator_data.json`.
- Device contract v1 device-specific operations and the digest-matched capability catalog
  on the simulator side.

No compatibility path is kept (pre-release).

## 4. The factory scenario pack

The pack is a hand-written `SituationSpec` (plus policy). It is **not** generated from the
simulator domain: it decides what the agent may propose and what that does to equipment, and
its diagnosis catalog must be independent of the simulator's fault list.

What it contains for the factory:

| Section | Content |
|---|---|
| `inputs` | Plant metrics by kind (`pump.motor_current.observed`, `air_header.pressure.observed`, …) |
| `operators` | Specific power, HX approach, sub-meter sum vs main meter, sensor disagreement, idle-time air flow |
| `situation` | See §6.2 — likely several situation types |
| `cognition` | Objective, prompt, tools, budget, **diagnosis catalog** (independent, with decoys and `unknown`) |
| `actions.intents` | e.g. `reduce_line_rate`, `stage_air_supply`, `transfer_cooling_duty`, `adjust_chilled_water`, `start_backup_power`, `recommend_shutdown`, `raise_alert` — with risk class and policy |

Automated cross-checks (not generation): every route targets a kind present in births; every
intent has a route; every input names a metric some kind announces.

## 5. What is scored on the Agentic Stream side

Because the system under test is Tamoz + Agentic Stream, the report attributes each outcome:

| Agentic Stream behaviour | Scored as |
|---|---|
| Policy denied an unsafe intent | Credit (prevented harm) |
| Policy denied a needed intent | Debit (blocked recovery) |
| Approval delay pushed an action past the deadline | Debit, attributed to approval |
| Session handling missed a gap / acted on stale data | Debit, attributed to host |
| Interlock tripped correctly on authority revocation | Credit |

`internal/replay` already runs shadow comparisons between a baseline and Tamoz. It is a
natural home for the baseline ladder's B1/B2 policies as alternative executors, so every
baseline runs through the same host.

## 6. Open design questions (Agentic Stream-specific)

### 6.1 Kill switch and the interlock — decided: option (a)

Under Tamoz ADR-039 the agent may only recommend shutdown. Options for executing it:

- **(a)** Agentic Stream routes the R4 recommendation to a human (simulated responder), who executes.
- **(b)** Agentic Stream's deterministic policy may itself issue `emergency_shutdown` on hard
  conditions, without the model.

Option (b) is not adopted: the automatic layer is the plant's own hardwired trips, simulated
in the world. Required change: an R4 route that delivers the recommendation to a person and
records who executed it and when.

Separately, **authority revocation** maps naturally onto `interlock`: a plant metric such as
`control_room.remote_enabled = false` trips the interlock and blocks the action plane.

### 6.2 Plant-scoped situations

Verified in `internal/spec/internal/domain/schema.json`: a spec has exactly **one** top-level
`situation` (keyed by `entityKey`), and operator kinds are limited to `aggregate`, `slope` and
`missing_heartbeat`; derived quantities can be written in CEL conditions over features. The
factory's confusable faults span entities and subsystems (AC1, AC2, the air header and M2
together). Possibly expressible today by partitioning on a plant id and filtering operators
with `where`; **to verify with a prototype spec**. Otherwise needed: plant-scoped situations
and several situation types per deployment — a one-time Agentic Stream change.

### 6.3 Alerts — decided: intents routed through `notify`

Tamoz proposes `raise_alert` as an intent (policy, rate limits, evidence). Agentic Stream
routes it to `notify`, extended to carry agent alerts as its single path to people. Only the
last hop differs: real recipients in production, the simulator's `people` device in
simulation. Required change: a `notify` delivery adapter that sends to the plant-wire session.
See [plant-wire-protocol.md](plant-wire-protocol.md) §15.

### 6.4 Transport — decided: one session per node

One full-duplex session carries data and commands, with control records ahead of data; a
single Agentic Stream session component owns it, handing data to ingress and accepting
commands only from the actions module. Replaces the live ingress socket plus device socket.
See [plant-wire-protocol.md](plant-wire-protocol.md) §13.

## 7. Summary of changes

| Change | Size | MVP |
|---|---|---|
| Plant-wire host connector with inventory and sequence handling | Large | Yes |
| Generic measurement schemas; envelope mapping | Medium | Yes |
| Command path reshaped: routes and bounds stay, operations from births | Medium | Yes |
| Default deny, expected inventory, policy recorded per run | Small | Yes |
| Read-only status view for the canvas | Small | Yes |
| Removals (§3.6) | Small | Yes |
| Factory `SituationSpec` pack | Medium | Yes |
| Plant-scoped / multiple situations (§6.2) | Unknown until verified | Likely |
| Baseline ladder through `replay` shadow executors | Medium | Before the first benchmark |
