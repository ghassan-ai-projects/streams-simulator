# The fan experiment: simulate first, then the real fan (plan, 2026-10-20)

Status: plan for review. Nothing in it has started. It supersedes sections 1 to 3 of `../PLAN-NEXT.md` for this experiment;
section 4 of that file (small simulator items) is folded in below.

## 1. Goal and the owner's three rules

The thermal LED and fan experiment (acceptance test #1) runs as:

```
Simulate:  Tamoz ⇄ Agentic Stream ⇄ plant wire ⇄ streamsim (zone, LED, fan twin)
Real:      Tamoz ⇄ Agentic Stream ⇄ plant wire ⇄ streamsim ⇄ plant wire ⇄ gateway ⇄ serial ⇄ Arduino (DHT11, LED, fan)
```

| Rule | Holds? | What it takes |
|---|---|---|
| R1. Going from simulated to real is a change to the simulator's configuration only | **Yes** (D2, D3) | The session plan gains `gateway_socket` and a `real` list. Agentic Stream's spec, catalog and flags stay the same and it keeps running; its hardware consent names the real board once, at start (D3) |
| R2. No work in Tamoz | **Yes** (verified, section 3.2) | Configuration only: worker profile, database path, model key, pinned Ruby |
| R3. Agentic Stream: configuration and specs only, kept outside the repository | **No, one time.** | Agentic Stream has no plant wire code. It needs one plant wire host module (code, with a design first). It is generic, written once, and after it every experiment is configuration only. Everything else for Agentic Stream is configuration |

R3 is the only rule this plan cannot meet as stated. No configuration of today's Agentic Stream reaches streamsim over the plant
wire. The alternative that avoids Agentic Stream code (streamsim speaking device-wire-v1 and Agentic Stream's JSONL to Agentic
Stream) puts consumer knowledge into the binary, which this repository forbids, and still needs new design. This plan
therefore asks Agentic Stream for one module and nothing else.

## 2. What was checked (so nothing below is a guess)

Every row was read in code or run on 2026-10-20, on `streams-simulator` `real-run` (= `main` 8070602e plus the old plan),
`agentic-stream` `real-run` at 91c2079d, `tamoz` `real-run` at 9170538c, `agent-research-lab` `main` at faf457b.

| Fact | Evidence |
|---|---|
| Agentic Stream has no plant wire code | No match for `NBIRTH` or "plant wire" in its code; its own review says so: `agentic-stream/docs/product-strategy-review-2026-10-18/PHYSICAL_DEVICES_IMPACT.md` F3 |
| Agentic Stream already has the ports a plant wire host plugs into | `internal/actionport` (`Effector`, `AuthorizedEffector`, `DeviceStateVerifier`, `VerifiedEffector`); `internal/runtime/internal/app/effect_routing.go` routes commands and the post-dispatch state check to one device effector; `internal/ingress` appends normalized events, with the simulator channel mapping as data (`internal/ingress/internal/domain/simulator_data.json`) |
| Tamoz has no thermal knowledge at run time | The allowed intents, the intent catalog (digest-checked, `IntentCatalog.verify_wire`), the prompt and the objective arrive from Agentic Stream on each request (`gems/tamoz-stream/lib/tamoz/stream/situation_request.rb`, `decision_builder.rb`). `request_bounded_cooling` exists only in Tamoz test fixtures |
| The fan route is `select_thermal_mode {mode: hold \| bounded_cooling}` (R2, signed human approval) | `agentic-stream/examples/real-world-sensor/zone-thermal-*.situation.yaml`, `thermal-capability-catalog.json` |
| A simulated world runs on a virtual clock that is not paced to the wall clock | `governance/DEFERRED.md` D-71; `connect_run.go` and `relay_sim.go` use `TimeStepped` and `model.DefaultStartTimeNS` |
| A connect session cannot script disturbances | `contracts/session-plan-v0.1.schema.json` has no scenario or fault member; `hot-afternoon.scenario.json` is an offline run script |
| The current binary is `bin/streams-simulator` | `bin/streamsim` is a stale build from 2026-10-13 (it lacks `relay`, `connect`, `twin-report`) |

## 3. Surprises found, and what this plan does about each

### 3.1 Time (the largest one)

1. **The simulated world runs as fast as the host answers.** A session advances one slice per barrier (`NSYNC`/`HSYNC`). If
   Agentic Stream answers at once, a 20-minute warm spell passes in seconds of wall time, long before Tamoz decides or a human
   approves the R2 fan command. If it does not answer, each barrier waits up to 30 s, and after three missed barriers the world
   ends (SPEC section 8).
2. **World time starts on 2026-01-01** (`DefaultStartTimeNS`), not now. A simulated device checks `deadline_ns` against world
   time; a real one against UTC. Agentic Stream would need two clocks in one session.

**Plan:** streamsim gains a **paced session** (S1). The world's start time is the wall time at session start, recorded in
`run.json` as an input (so replay stays deterministic), and slices are released on the wall clock. The barrier becomes a cap,
not the pace. Agentic Stream then answers every `NSYNC` at once and uses UTC for every device, simulated or real. The owner
agreed on 2026-10-20 that this belongs in the simulator.

### 3.2 The device model does not match Agentic Stream's catalog in two places

| Item | Agentic Stream catalog and firmware | streamsim kind files | Plan |
|---|---|---|---|
| LED `pattern` | strings `off`, `slow_blink`, `solid` | integer 0 to 2 | Superseded by D7 (section 4.2): Agentic Stream stops sending board values; the LED becomes a state |
| Safe stops | operation `safe_stop` per target | commands `stop_led`, `stop_fan` | Agentic Stream configuration: its catalog names `stop_led` and `stop_fan`; the gateway maps them to the firmware's `safe_stop` (already in its brief) |

The fan matches today (`set_pwm_lease {duty_permille 0..600, lease_ms 5000..10000}` on both sides), but D7 changes it too.

### 3.3 The fan's effect will be small

The `bounded_cooling` preset is duty 450 with a 5 s lease, and Agentic Stream does not renew a device lease (searched: its
"renew" is the authority claim, not the command). The twin's fan cools with gain −0.02 °C per permille and a 180 s time
constant, so one approved command lowers the zone by about 0.25 °C before the lease lapses. The DHT11 resolves 1 °C. **The
first runs will show the command, receipt, result and the fan spinning for 5 s, not a visible temperature drop.** The old
"hot afternoon" script renewed the lease every 5 s for exactly this reason.

**Plan:** the first pass criteria (section 6) are the command lifecycle and the fan's own report, not a temperature drop.
Raising the preset lease to 10 s is configuration (still about 0.5 °C). A closed loop that visibly cools needs lease renewal
under one approval in Agentic Stream: a product feature, out of scope here, listed as D6.

### 3.4 A simulated zone with a real fan is meaningless

A real fan cools the real DHT11, not the twin's zone. The real rung therefore binds `zone-01` **and** `fan-01` to the board.
`led-01` may stay simulated or be real.

### 3.5 A session with every instance real is probably refused

When every instance is bound, the world's board has no members and is not announced, but the relay still writes
`environment: mixed` on the merged birth (`relay_mixed.go` `mergedBirth`). SPEC section 2 says `mixed` needs at least one board
of each environment, so a strict host refuses `environment_mismatch`. Not yet confirmed by a test. **Plan:** S2 writes the
test first, then derives the environment from the boards actually announced.

### 3.6 Smaller items

| Item | Plan |
|---|---|
| `ambient_warm_spell` raises the zone by 16 °C, not 4 °C (fault delta 4 × onset magnitude 4) | S3: magnitude 1 |
| Disturbances cannot be scripted in a live session | S5: the session plan names faults at world offsets (disturbances only; streamsim still never sends an ordinary command) |
| `plantwire connect --out` reuses an existing run folder (D-114); the reference host exits without its report when stopped (D-115) | S6, S7 |
| The relay drops a gateway silent for 30 s; the DHT11 kind declares no `scan_class`, so readings go stale after 30 s | The gateway sends every 2 s; noted in its brief already |
| `HSYNC` evidence is not carried for hardware devices (D-112); an instrument on another device cannot confirm an effect live (D-111) | Accepted for this experiment: the effect check is the device's own report; independent confirmation stays an offline bench check |
| Agentic Stream's two specs differ: the sim spec reads ambient and fan tach, the bench spec only temperature, humidity and heartbeat | Use the **bench spec for both rungs**. The twin's sensor has exactly temperature and humidity, so no spec changes at the switch (R1) |
| The bench spec needs `zone.heartbeat.observed`, and **without it no decision is ever made**: its `missing_heartbeat` operator sets `link_missing_90s` after 90 s without a heartbeat, and the trigger requires the link not missing | Required, not optional: the Agentic Stream host emits a heartbeat event per zone from the node's liveness (mapping data, A2), at least every 30 s |
| The bench spec's thresholds: a situation opens when the 5-minute mean is above 30 °C **and** the slope above 6 °C per hour; it goes to `cooling` above 9 °C per hour, `over_ceiling` above 38 °C | Simulated: the twin starts at 34 °C, so only the warm spell's rise opens it (S3, S5). Real: the DHT11 must read above 30 °C on average for minutes and still be rising; a hand on the sensor may not be enough (O5: a warm air stream at a safe distance). Expect the first decision a few minutes after warming starts, not seconds |
| Humidity is ingested but no operator or reducer uses it, so it is not in the situation's facts | Nothing to do; noted so nobody expects Tamoz to reason about humidity from the snapshot alone |
| Agentic Stream's `make ci-check` stopped at the pinned `protoc` 35.1 on this host (36.0) in September | Check in P0; install the pinned version or record the gap |
| The old demo canvas (PID 27234) still runs from a previous session's `/private/tmp` scratchpad | P0: copy `hot-afternoon.scenario.json` if wanted, then stop it |
| The board was disconnected after the September probes | P5 preflight starts from wiring, not from "it worked before" |
| **Agentic Stream hardcodes thermal knowledge in two places** (owner's top priority, 2026-10-20). See section 5.1 | Fixed first, before the host module, as H0 (section 5.1) |

## 4. Decisions for the owner (recommendations first)

| # | Decision | Recommendation |
|---|---|---|
| D1 | Accept one Agentic Stream module (the plant wire host) as the only code outside streamsim and the gateway | **Decided 2026-10-20 (owner): yes, as an adapter.** The plant wire host is one adapter behind Agentic Stream's existing ports (the effect port `actionport` for commands, the ingress append path for readings), beside the device-wire adapter (D4); nothing above the ports knows which adapter is in use, and the adapter holds no device or experiment knowledge |
| D2 | Pace sessions to the wall clock in streamsim (S1) | **Decided 2026-10-20 (owner): yes.** Live sessions are paced (world clock starts at the session's wall time, recorded in `run.json`; one world second per wall second; the barrier is a cap). The unpaced mode stays for offline tests and benchmarks |
| D3 | Agentic Stream's hardware consent | **Decided 2026-10-20 (owner): consent per named hardware board, given once at start; no Agentic Stream restart at the switch.** Agentic Stream starts once with its existing consent (`--live-actuation --owner-authorized`) plus the boards it may command as hardware (for example `--hardware-boards mega-01`). The plant wire adapter reads each device's origin from the birth (the board's `environment`, SPEC section 2): a `simulated` device is commanded freely; a `hardware` device only when its board is on the list; any other hardware board is refused, visibly. The origin is a claim, but streamsim makes the dangerous direction impossible (the relay refuses a gateway that claims `simulated` or `mixed`; a simulated world can only announce `simulated`). The switch is then only the session plan: stop the streamsim session, start it with `real.plan.json`; Agentic Stream and Tamoz keep running (the host accepts a new session after the old one ends, A1). R1 holds completely |
| D4 | Agentic Stream's device-wire-v1 path | **Decided 2026-10-20 (owner): keep it, as one adapter beside the plant wire host, behind the same effect port (`actionport`); not retired and not planned here.** Retiring it is a later, separate decision |
| D5 | D-113 (contract version check) | **Decided 2026-10-20 (owner): keep the exact match for now.** Gateway, streamsim and Agentic Stream's host carry the identical contract folder and refuse any other version or digest; a contract change is synced to all three together. The minimum-version rule comes with the first field deployment |
| D6 | Lease renewal under one approval (visible cooling) | **Decided 2026-10-20 (owner): option A, bursts for now.** Each fan command runs one 5 s lease (preset unchanged); the pass criteria are the command chain, not a temperature drop. The lease is 5 to 10 s in four places (Agentic Stream catalog, fan kind file, gateway catalog, firmware constants `MIN_FAN_LEASE_MS`/`MAX_FAN_LEASE_MS`); beyond 10 s needs a reflash and weakens the last line of defence. Visible cooling (Agentic Stream renewing the lease while a cooling mode holds, under one approval) is the first follow-up after P6 |
| D7 | Where device values live | **Decided 2026-10-20 (owner): no board-specific value in Agentic Stream; the fan takes a level** (`set_cooling {level_percent}`, reported as `level_percent`). See section 4.2. A temperature setpoint with a local control loop is the industrial target for later |
| D8 | Run the old device-wire baseline (`../PLAN-NEXT.md` section 1) first | A **reduced** one (P0.5): repeat `.e2e-run/round-007-simulator-pass2` on today's checkouts (LED only, R1, no approval), with Tamoz on the D11 model. It checks the parts this plan builds on (Agentic Stream to Tamoz, policy, dispatch, reconciliation) and is the "before" evidence for H0. **When something is broken** (section 4.3) |
| D8a | What to do with a failure in P0.5 | Triage before fixing, section 4.3 |
| D11 | Tamoz's model | **Decided 2026-10-20 (owner): the model already configured in Tamoz, `zai/glm-5.3-flash`** (`~/.tamoz/config.yaml`), set as the stream worker profile's `model_roles` entry, key from `ZAI_API_KEY`; not Ollama. Note: the September pass used a scripted stand-in model (an OpenAI-style endpoint on a random local port, one logged request), so it proved the plumbing, not Tamoz's judgement |
| D9 | Who approves the fan | **Decided 2026-10-20 (owner): the node, for now.** See section 4.1 |
| D10 | Scope of streamsim as the control plane | **Decided 2026-10-20 (owner): make it work on the bench first.** No authentication, several users or TCP until a first customer needs them (D6 of `PHYSICAL_DEVICES.md` stands). The canvas shows what crosses the plant wire (readings, commands, receipts, results, effect) and the node's controls. Showing the host's progress and notifications (situation, decision, approval requested) and answering approvals from the canvas is a later plant wire 1.4.0 (host events, approval answers). It is additive: Agentic Stream's host module can emit them later from its existing event outbox (Channel B), so A0 does not design for them now, only keeps the session able to carry more record types |

### 4.1 D9: approval at the node now, Tamoz and Telegram later

What exists, checked in code: Agentic Stream classifies R2 as "needs a human approval" in its contract
(`contractsv1/internal/domain/risk.go`: R0, R1 automatic; R2 approval; R3, R4 denied), publishes `ApprovalRequested`,
`ApprovalWithdrawn` and `ApprovalResolved` on its durable event feed (Channel B, SSE at `/v1/events`), and accepts a signed answer
at `POST /v1/approvals/{id}` (Channel C). Tamoz's `tamoz-stream-subscriber` reads that feed and takes `--approval-relay FILE`;
its `ApprovalRelay` has injected ports (delivery channel, Channel C submission, nonce store, signer) with **no production
implementation** (only test fakes). So today nobody can answer an R2 approval, and an R2 fan command waits until it expires.

Decision for this experiment:

- **The fan runs as R1.** In the Agentic Stream bench spec copy (outside its repository): `select_thermal_mode` `risk: R1`,
  `policy: automatic`. Agentic Stream sends the command without its own approval. The policy digest changes; the session plan's
  `policies` and the gateway's allow-list carry the new one.
- **The human gate is the node's "approve each command"** (built in M5): the operator approves or denies each fan command in the
  canvas, connected with `plantwire connect --control-socket`. A denial is `refused local_control`; no answer before the deadline
  is `expired`. The decision is recorded in streamsim's journal (`interrupts.jsonl`), not in Agentic Stream's approval ledger.
- **The deadline is the route's `expires_after_ms`** in Agentic Stream's capability catalog, per route (today: LED 60 s, fan 20 s,
  safe stops 1 s). The catalog copy sets the fan to 120000 (120 s, as the reference host uses), so the operator has time to
  approve. A long deadline does not lengthen the fan's run: that is the 5 s lease.
- **What this run does not prove:** Agentic Stream's own human-approval path. Its record says "automatic"; R1 understates a
  motor's risk, and safety rests on the node approval, the gateway allow-list and the firmware lease.

Why the node is a gate and not Agentic Stream's approver: in the production shape Agentic Stream talks to the gateway directly
and streamsim is absent, so an approval that lives only in streamsim would vanish there; and streamsim signing Agentic Stream's
approvals would put Agentic Stream's operator interface and the approval key in the simulator, which `PHYSICAL_DEVICES.md`
section 10 rules out. The node approval only holds or releases a command Agentic Stream already sent.

### 4.3 D8: when the baseline run fails

Each failure is classified first, with its exact evidence (command, log lines, the three commits), and nothing is patched
around:

| Class | Example | Action |
|---|---|---|
| Our setup | a flag renamed since September, a missing key, a stale socket, a wrong path | Fix the configuration and re-run. No code |
| A streamsim defect | `device serve` refuses something it should accept | Fix here, one commit with a test, as usual |
| An Agentic Stream or Tamoz regression | worked in September, fails now, same inputs | Record it in this folder (`P0.5-FINDINGS.md`) and stop that line. The owner decides per finding; a fix is its repository's work, in its own session and branch, with a test. Never weaken a check to get a pass |
| A model outcome | the model chose "watch" instead of the LED | Not a failure of the plumbing: read the decision record, adjust the telemetry feed, re-run |

The rules R2 and R3 (no work in Tamoz or Agentic Stream for the experiment) do not forbid fixing a regression in those
repositories; they forbid building the experiment there. Each such fix is the owner's call.

### 4.2 D7: semantic commands in the kind file, raw values in one board mapping

The test for any value: **would it mean the same on another board?**

| Value | Example | Same on another board? | Where it lives |
|---|---|---|---|
| A state or mode word | `watch`, `bounded_cooling` | yes | the kind file (allowed words); Agentic Stream or Tamoz chooses |
| A number in physical units | 45 % cooling, 22 °C setpoint, 120 s | yes | the kind file (unit and range); Agentic Stream or Tamoz chooses, Agentic Stream sees and checks it |
| A raw board value | PWM 450 ‰, pattern code 1, a pin, a start-up threshold | no | **one board mapping file**, read by the twin and by the gateway |

Consequences:

- **Agentic Stream's catalog** keeps the route (intent to device and command), the selector field, and presets only in
  physical units (for example `bounded_cooling` = 45 %); no PWM, pattern or other board value. Tamoz is unaffected (it already
  writes only `state` or `mode`). The device-wire adapter kept by D4 keeps its own presets with values.
- **The kind files** change: the indicator offers `show_state {state: off | watch | alert}`; the fan offers `set_cooling
  {level_percent: 0..60}` (recommended; a mode word `select_mode {mode: hold | bounded_cooling}` is the alternative). The safe
  stops stay. Kind digests change.
- **Metrics follow the same rule.** Agentic Stream verifies a command by comparing the device's report with what it sent, so a
  device reports in the command's units: the fan reports `level_percent`, the LED reports `state`. The board mapping works in
  both directions (command to raw value, raw reading to unit).
- **The lease is declared by the board, not chosen by Agentic Stream:** the kind declares how long a command's effect lasts
  (5 to 10 s for this fan); Agentic Stream reads it (it needs it for lease renewal later, D6).
- **One mapping file per board**, so the twin and the real board cannot drift apart. Its schema sits beside `KIND` in the plant
  wire contract folder, which changes the contract `DIGEST`: **contract 1.4.0**, synced to all three repositories together (D5).
  Nothing on the Agentic Stream or gateway side is built yet, so this costs no rework.
- **Safety.** The model still never chooses a raw value; the firmware keeps its limits (duty at most 600 ‰, lease 5 to 10 s).
  Agentic Stream no longer sees the raw value it causes. Later, if wanted, the board announces its mapping's digest in its birth
  and Agentic Stream records it with each command (a further wire change, not in this plan).

Later (a separate goal, not this plan): **Tamoz and Telegram become Agentic Stream's R2 approver** by implementing the relay's
ports in Tamoz (a Telegram delivery channel, a Channel C client, a signer whose key Tamoz never sees). The fan goes back to R2;
the node approval stays as a second, local veto. Agentic Stream needs no change for it: the event feed and the answer endpoint
exist. No webhook is added to Agentic Stream; a tool that wants one gets a small subscriber of `/v1/events`.

## 5. Work by place

Sizes: S (a round), M (two or three rounds), L (a milestone).

### streamsim (this repository), branch per milestone from `real-run`

| # | Work | Size | Exit |
|---|---|---|---|
| S1 | Paced session: start time = wall time at start, recorded in `run.json`; slices released on the wall clock; barrier as a cap; replay reproduces it | M | A connect session of 10 minutes takes 10 minutes; `replay` and `verify` match; D-71 closed |
| S2 | Environment of a merged birth from the boards announced; test for an all-real plan first | S | All-real plan announces `hardware`; mixed stays `mixed` |
| S3 | Warm spell magnitude 1 and its description. **Add a positive heat-load fault** (`zone_heat_load`: a heat source inside the zone raises the zone, not the room; the zone model already has a `heat_load` input, so this is a fault entry on it, data only). Today the twin's only cause of overheating is `ambient_warm_spell`, which the domain labels negative class ("nothing on the board is wrong"), so the only disturbance we could script is one where the right answer may be *not* to cool. Warming the real DHT11 with warm air is physically a local heat load, so the twin should offer the same disturbance. The warm spell stays as the negative test case | S | Zone rises 4 °C in the lint day; the heat load raises the zone and not the ambient |
| S4 | D7: semantic kind commands and metrics (`show_state`, `set_cooling`/`level_percent`, declared lease); the board mapping file and its schema in the contract folder (contract 1.4.0, `DIGEST` reviewed); the twin implements a command through the board mapping (today its responses only copy an argument into a metric: `direct`, `first_order`, `none`) | M | `kind validate` prints the new digests; the twin's fan at 45 % behaves as duty 450 ‰ did; contract tests green |
| S5 | Scripted disturbances in the session plan (`session-plan-v0.2`: faults at world offsets) | M | The warm spell and the stuck sensor fire at their times in a live session and on replay |
| S6 | D-114: `connect --out` refuses a used run folder | S | Test |
| S7 | D-115: reference host exits cleanly with its report | S | Test |
| S8 | Experiment data: `examples/devices/thermal-board/fan-experiment/` with `simulate.plan.json` and `real.plan.json` (differ only in `gateway_socket` and `real`) and a README | S | `plantwire connect --check` passes both |
| S9 | Dress rehearsal without Agentic Stream: `connect` with the simulate plan against `plantwire host`, then the real plan against the stand-in gateway | S | Two retained run folders; `verify` matches |

### 5.1 Agentic Stream H0: remove hardcoded domain knowledge (first, owner's top priority)

**Finding.** A sweep of Agentic Stream's production Go code for every intent type, event type and enum value its example specs
declare (2026-10-20) found exactly two places with domain knowledge in code:

| Place | What it hardcodes | Effect |
|---|---|---|
| `internal/runtime/internal/domain/lifecycle.go:36-44` `ClassifyRoute`, used by `internal/runtime/internal/app/effect_routing.go` (`CompositeEffector`) | The route names `set_indicator` and `select_thermal_mode` are device routes; anything else goes to the fallback, which in a live profile refuses (fail-closed) | **A new device command needs a code change.** A second board, a vent, a pump: each would be refused until someone edits this list |
| `internal/replay/internal/domain/baseline.go:214-224` `baselineEnumValue` | The field names `state` and `mode`, the thermal phases `over_ceiling`/`cooling`, and the values `alert`/`watch`/`bounded_cooling`/`hold` | The replay's rule baseline only understands the thermal spec; any other domain gets the first enum value. Off the live path |

Not domain knowledge, and left as they are: `install_watch_condition` (Agentic Stream's own built-in watch intent:
`internal/watch/internal/domain/condition.go`, `internal/decisions/internal/domain/binding.go:91`) and the test-support fixtures
(`internal/testsupport`, `internal/executor/fixture`). No event type is hardcoded.

Agentic Stream's own review saw the first one (`docs/reviews/2026-10-09-project-review/02-architecture.md`, AR-13, Low) and
recommended moving the list to `actionport`. That moves the list; it does not remove it. Its own vocabulary already says
the capability catalog is "the closed set of routes and safe stops a device accepts. Configuration, never model output.
Identified by its digest" (`internal/device/UBIQUITOUS_LANGUAGE.md`), so the name list repeats in code what the catalog already
decides as data.

**Fix.**

| # | Change | Size |
|---|---|---|
| H0.1 | **A device route is a route the device catalog names.** The device effector reports the routes it owns (from its catalog); the composite router sends a route to the watch effector if it is the built-in watch route (the constant from `watch`, not a literal), to the device effector if that effector owns it, else to the fallback (fail-closed in a live profile, as today). The closed set stays closed, now by a digest-pinned file instead of a code list. With AR-13: the router moves out of the `runtime` composition root into the action plane | S |
| H0.2 | **Fail at start, not at dispatch.** When `serve` starts, every intent in the spec that can produce a command must be owned by an effector (watch, device catalog, or an explicit non-device effector); otherwise `serve` refuses to start and names the intent. Today a misnamed route is discovered only when a command is refused | S |
| H0.3 | **The replay baseline reads its mapping from the spec.** Each intent may declare, per situation phase, the preset or value the rule baseline chooses; the code keeps no field, phase or value name. This is a spec schema change, so the spec digests pinned in Agentic Stream's tests change in the same commit | S to M |
| H0.5 | **Event schemas from outside the binary.** The event-schema registry (`internal/spec/internal/domain/event_schema_data.json`, 52 schemas for bays, motors, ponds, pumps, reefers, shipments, zones) is embedded with `go:embed` and is the only source the spec compiler (`references.go`), the operators (`boot.go`) and deployments (`store/deployments.go`) consult. It is data, not logic, but a new experiment's reading type needs an edit in Agentic Stream's repository and a rebuild. Fix: the registry is also read from a directory given at start (outside the repository, like the spec); the embedded one stays as the default. Later, with the plant wire adapter, a commissioned kind file's metrics (name, unit, range) can supply the schema directly, so a device is described once (F7, D7) | S to M |
| H0.4 | **A guard so it cannot return.** An architecture test in Agentic Stream fails when production Go code contains an intent type, event type, phase or enum value that a spec under `examples/` declares (Agentic Stream's own built-ins, such as the watch intent, listed explicitly) | S |

Tests: a route with a new name that the catalog names reaches the device; a route the catalog does not name is refused; a spec
intent no effector owns stops `serve` at start; the existing experiment end-to-end tests (`cmd/agentic-stream/experiment_e2e_test.go`,
`experiment_live_feed_test.go`) pass unchanged; the baseline produces the same choices for the thermal spec as today.

H0 needs no plant wire and no decision: it runs on today's device-wire path, so it can start at once, and the host module (A3)
then inherits catalog-owned routing instead of adding a third route name. After H0, a new device command is configuration and
data only (kind file, board mapping, spec intent, catalog route).

### Agentic Stream (one module, design first)

Its `AGENTS.md` requires a design before code. The ports exist, so the design is "a new adapter behind existing ports":

| # | Work | Size |
|---|---|---|
| A0 | ADR: a `plantwire` host module that owns one plant wire session and serves two existing ports from it: an ingress source (readings to normalized events, through the ingress append path) and an effector (`actionport.AuthorizedEffector` plus `DeviceStateVerifier`). Kind files are the source of the device vocabulary; event mapping is data. Entity resolution (device to entity, entity and intent to device) is one seam, read from the catalog now (section 8.1, item 4) | S |
| A1 | Host socket (owner-only), `NBIRTH`/`STATE` handshake, contract version and digest, `mixed`/`hardware` acceptance, commissioning by `kind_digest`, heartbeats, `HSYNC` answered at once; **a new session after the old one ends is accepted without a restart** (the switch, D3) | M |
| A2 | `DDATA` with quality to `zone.temp.observed`, `zone.humidity.observed`, `zone.heartbeat.observed` (mapping data, like `simulator_data.json`); `DDEATH`, rebirth, `boot_id` change; zero quarantines | M |
| A3 | Commands: the governed command to `DCMD` (`expected_boot_id`, `deadline_ns` in UTC, `authority_epoch`, `policy_digest`, `idempotency_key`); `DRECEIPT`/`DRESULT` into the existing reconciliation; state check from `DSTATE`/`DDATA`; refusals `local_control`, `queued`, `expired`, `session_ending` as ordinary refusals; the plant wire effector owns the routes its catalog names (H0); **origin check (D3)**: a command to a device on a `hardware` board is sent only with the consent flags and the board on `--hardware-boards`, otherwise refused visibly | M |
| A4 | Log the policy digest at start | S |

Test peers already exist in streamsim: `plantwire node`, `plantwire connect`, the stand-in gateway, and `plantwire host` as the
behaviour to copy. Agentic Stream configuration, kept outside its repository: the bench spec copy, `principals.yaml`, the
catalog with `stop_led`/`stop_fan` safe stops, routes to `show_state` and `set_cooling` with presets only in physical units
(D7), and the fan's `expires_after_ms` at 120000, the fan intent as R1 automatic (D9), the event mapping, and the `serve` flags.
The host module speaks contract 1.4.0 (D5, D7).

### Gateway (`agent-research-lab`, Python)

| # | Work | Size |
|---|---|---|
| G1 | The plant wire node side, exactly as `../briefs/gateway-plant-wire.md` (dial the relay, births with the kind files and their digests, readings every 2 s with quality, commands with epoch, policy allow-list and idempotency, safe stops mapped to the firmware's `safe_stop`, rebirth, host loss). Take the `kind_digest`s from `streamsim kind validate` rather than writing RFC 8785 in Python. Translate the kind's semantic commands and metrics through the board mapping file (D7), the same file the twin reads; contract 1.4.0 | M |
| G2 | Test against streamsim's relay with a serial PTY in place of the board | S |

The firmware does not change. The gateway no longer talks to Agentic Stream directly.

### Tamoz

Nothing. Configuration: worker profile, database path, model key, pinned Ruby (rbenv).

### Outside the repositories (owner)

| # | Work |
|---|---|
| O1 | Re-attach the board: DHT11 on D2, L293D `1A→D6`, `2A→D7`, enable D5, separate 3 to 6 V motor supply; fill `ARDUINO-BENCH-MANIFEST` with measured values |
| O2 | A physical e-stop (cut the motor supply) and a current limit on the motor supply |
| O3 | The Tamoz model key; Ruby via rbenv; the pinned `protoc` for Agentic Stream's gate. No approver key is needed while the fan runs as R1 (D9) |
| O4 | Optional: an independent instrument (`../briefs/bench-instrument.md`). Without it the fan's effect is the board's own report plus your eyes |
| O5 | The disturbance on the real rung is physical: warm the DHT11 (a hand, a warm air stream at a safe distance) |

## 6. Phases, order and exit criteria

| Phase | Who | Depends on | Exit |
|---|---|---|---|
| P0 Housekeeping and decisions | owner + main session | none | D1 to D8 answered; `make phase-gate` green on `real-run`; demo canvas stopped; three commits recorded |
| P0.5 Reduced baseline (device-wire, as in September) | Haiku runs it | P0 | Agentic Stream + Tamoz + `device serve` emulator: one decision, one command, reconciled, `verify-run` pass; or a precise list of what drifted. Also the "before" evidence for H0 |
| P0.7 **H0: remove hardcoded domain knowledge** (section 5.1) | an Agentic Stream session | P0.5 | H0.1 to H0.4 merged; P0.5 re-run passes unchanged |
| P1 Simulator readiness S1 to S9 | streamsim sessions | P0 | All S exits; dress rehearsal folders retained |
| P2 Agentic Stream host A0 to A4 | an Agentic Stream session | D1; may start with P1 | Accepts streamsim's simulated node with zero quarantines; one governed command refused, one accepted with result; a hardware board from the stand-in gateway is refused without consent and accepted with it; a second session is accepted without a restart |
| P3 Gateway G1, G2 | a gateway session | may run with P2 | Brief's acceptance checks 1, 5, 6, 7 against the relay with a PTY |
| P4 **Simulated rung**: `connect` with `simulate.plan.json` ⇄ Agentic Stream ⇄ Tamoz | main session | P1, P2 | Section 7 criteria, all simulated; retained run folder plus Agentic Stream `export-run`/`verify-run` |
| P5 Hardware preflight | owner | O1 to O3, P3 | Runbook stage 0 (local safety without Agentic Stream): lease lapse stops the fan; safe stop works; e-stop works |
| P6 **Real rung**: the same, with `real.plan.json`; Agentic Stream and Tamoz not restarted (D3) | owner at the bench, main session | P4, P5 | Section 7 criteria with `zone-01` and `fan-01` real |

## 7. Pass criteria (both rungs)

1. The host accepts the birth; zero quarantines; readings arrive every 2 s or faster.
2. The disturbance raises the zone (simulated: the scripted `zone_heat_load`; real: warming the sensor with warm air).
3. Tamoz proposes `set_indicator`; Agentic Stream executes it (R1, automatic); `DRECEIPT accepted`, `DRESULT succeeded`.
4. Tamoz proposes `select_thermal_mode bounded_cooling`; Agentic Stream sends it (R1, D9); the command waits at the node until
   the operator approves it in the canvas (`queued`, then `accepted`); `DCMD set_cooling {level_percent: 45}` reaches `fan-01`,
   the result follows, and `level_percent` reads 45 until the lease lapses, then 0 (on the real board the gateway has sent
   `set_pwm_lease {450, 5000}` to the firmware). A second proposal, denied in the canvas, is `refused
   local_control`; a third, left unanswered, is `expired` after 120 s.
5. The policy digest is accepted by the session plan's list (simulated) or the gateway's list (real); a command with another
   digest is refused `policy_not_allowed`.
6. A hold from the canvas gives `refused local_control`, and Agentic Stream does not retry in a loop.
7. Simulated rung only: `streamsim verify` on the run folder matches; real rung: the field trace replays to a host
   (`plantwire replay-trace`) and is announced `simulated`.

Not claimed by these runs: a measured temperature drop (3.3), independent confirmation of the fan (O4), any fault matrix.

## 8. What does not change

The wire records of the plant wire contract (1.4.0 only adds the board mapping schema, D7; the session plan of S5 is not part
of the wire); Tamoz; the
firmware; streamsim's rule that it never sends an ordinary command; the nine non-negotiables; no gate is loosened.

## 8.1 Later goals (recorded, not planned here)

Discussed with the owner on 2026-10-20; none of them blocks the fan experiment. **Order (owner, 2026-10-20): item 4 is the
first follow-up after P6**; then item 1, which item 3 depends on; item 2 when item 1 shows where rules suffice.

1. **Does Tamoz earn its cost (A/B/C).** After P6, the same scenarios three ways: rules only (the replay `BaselineExecutor`),
   Agentic Stream with rules, Agentic Stream with Tamoz. Scenarios: `zone_heat_load` (cool), `ambient_warm_spell` (hold back),
   the stuck DHT11 (distrust the reading), `fan_stall` (the command ran, the zone did not cool). Scored against the twin's sealed
   truth and the trivial baseline.
2. **Rules first, the model only when hard.** Today the spec has one executor and every action decision goes to it; triggers
   decide only *when*. A trigger chooses who decides (`decide: rules | tamoz`), the rule side being the baseline executor made
   data-driven by H0.3. Agentic Stream product work, design first.
3. **Tamoz proposes rule changes from what it learned.** Tamoz already learns from outcomes (`tamoz-stream-subscriber` reads
   `outcome.recorded` and `outcome.reconciled` into its memory). New: Tamoz drafts a spec change (a threshold, a case that is now
   clear moved to a rule trigger, a known false alarm written into the objective), and the change goes through the same gate as
   any spec change: Agentic Stream compiles it, it is replayed against every recorded run and twin scenario old versus new and
   scored, a human approves it, and it is deployed as a new spec version with a new digest. Tamoz proposes, never applies; it
   cannot widen risk classes or actions. Over time, cases Tamoz keeps solving the same way become rules, so model calls fall.
4. **Several rooms, and the layout described once (the first follow-up after P6).** Findings (2026-10-20):
   - The domain format already models several rooms as data: `instances` take several `names` per kind, `links` connect
     named instances (`fan-01` `cools` `zone-01`), and the zone's temperature sums the fan duty over its `cools` links
     (`"link": {"aggregate": "sum", "type": "cools"}`). Three rooms are three names per kind, three links and a board per room.
     Only the twin scaffold (`kind twin`) is one instance per kind (`twin_resolve.go:25`); the domain is not.
   - The layout (which fan cools which zone, which area a device is in) lives in the domain and does not leave streamsim: the
     births carry boards and devices, not links or areas. Agentic Stream learns it from its catalog's `target_bindings` (config),
     and a fan's readings carry the fan's id, not its zone's.
   - The session plan binds every real device to a domain instance, so the domain is the site model for simulated and real
     devices alike: streamsim can announce the layout for both, and the gateway need not know it. In the direct production
     shape (no streamsim), Agentic Stream keeps the configured layout.
   - Agentic Stream also counts `rateLimitPerHour` per tenant and intent type, not per entity
     (`internal/policy/internal/store/intent_reads.go:30`): six fan commands an hour are shared by all rooms. And nothing correlates
     rooms (one cause heating all three); that is a spec and design question in Agentic Stream.
   - Work then: a three-room domain (two simulated rooms beside the real one, a mixed session); the layout in the births
     (a contract change, synced to all three repositories); Agentic Stream's adapter reading it; the rate limit raised (config) or
     made per entity (Agentic Stream code).
   - **For this plan only:** A0 designs the adapter's entity resolution (device to entity, entity and intent to device) as one
     seam, implemented from the catalog now, so the layout from the births is an addition later and not a rewrite. No contract
     member is added before a consumer needs it.

## 9. How the work runs

Per the standing rules: Sonnet writes code and its tests, Haiku runs gates and long scripts in the background, the main session
briefs, reviews every diff and commits one change per item; a branch per milestone from `real-run`, merged at the end;
`make phase-gate` before each push; browser checks in the owner's Chrome. Agentic Stream and gateway work runs in sessions in
their own repositories, each given the relevant rows of section 5 and its brief.
