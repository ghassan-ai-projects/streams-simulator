# streamsim and the multi-drive belt conveyor: what it can and cannot express today

Date: 2026-11-07. Branch `belt-conveyor-drives` (cut from `origin/main` at `fed4e6e8`). Read-only study: code, contracts, docs and
tests were read; no build, no test, no binary run. Nothing was changed except this file.

How to read it: **[V]** is a verified fact with a `file:line` (or function) reference. **[I]** is an inference from those facts
(including "this should work": nothing here was executed, so every [I] "yes" is a design claim to be proven by a test before it is
relied on). Paths are relative to `/Users/ghassan/my-projects/streams-simulator`.

Short answers first.

| Question | Answer |
|---|---|
| Can three motor-drive instances share one belt force, divided by the number of running drives, with each drive's share coming back to it? | **Not with today's library, and not by composition alone. Yes with data-only additions** (a new belt component, ports on motor-drive, a `curve` for 1/n, one `sample_s` on a loop leg). The simulator core needs no change for the equal-share case. Unequal sharing (torque bias) needs a division, which no F1 form provides. |
| Is cross-instance aggregation (sum, count running) possible? | Yes in the core (`sum`, `mean`, `min`, `max`, `count_running`), but a **composition** can only ever emit `sum`. `sum` of 0/1 `enable` states is a running count, so it still works through composition. |
| Does the simulator grade this use case? | **No.** Scoring is per faulted entity, one expected effector, detection latency against a noise-floor onset. Lead time to trip, life consumed on survivors, stranding, throughput given up and false alarms as a plan-level cost do not exist as truths or metrics. |
| Top gaps | (1) No truth for "when would drive 3 trip / what would have been consumed" (counterfactual outcome truth). (2) Outcome scoring across entities (survivor life, stranding, throughput). (3) No division form (unequal share, torque bias). (4) No feed effector, no belt component with a load model, no belt-force/restart model in the library. (5) F2 reference models (Arrhenius) are specified but refused at validation. (6) Compositions cannot express `count_running`/`mean` or per-instance parameters. |

---

## 0. Repository rules that shape every proposal (all [V], from `AGENTS.md`)

- Domains, adapters, effectors are **data**; no per-domain code branch (`AGENTS.md` "Architecture Overview" and "Forbidden Changes").
- "Deliberately not an expression language" (`docs/design/TECHNICAL_DESIGN.md:311-330`, repeated in the schema description at
  `contracts/domain-spec-v0.2.schema.json:~458`). New numeric behaviour is a **named closed form with an equation in the design doc
  and an independent oracle test**, not a formula string.
- Files at most 300 lines, production functions at most 15 lines, no gate loosened, no new package without a stated design
  (`AGENTS.md` MUST 3 and 4). Pre-release: schemas, domains, digests may break freely (memory note "No backward compatibility").
- Nine non-negotiables: determinism, analytic cross-check, reference consumer, delivery ledger, quiescence barrier, sealed oracle,
  injection probe, `silent_no_effect` test, trivial-baseline audit. Any new truth or metric must keep the sealed-oracle boundary.
- The simulator holds no consumer knowledge (no agentic-stream or tamoz names in the binary). Control logic (cut feed, run empty,
  stop drive 3 at a time) belongs to the consumer; the simulator provides effectors and a world that reacts.

## 1. Components

### 1.1 What the library contains [V]

Equipment in `internal/component/internal/files/builtin/equipment/` (20 files) and instruments in `.../instruments/` (17 files). The
ones that matter here:

| Component | File | Role for this use case |
|---|---|---|
| `streamsim/conveyor@1` | `equipment/conveyor.component.json` (690 lines) | A belt **with its own single motor**; not a belt shared by several drives. |
| `streamsim/motor-drive@1` | `equipment/motor-drive.component.json` (800 lines) | Induction motor on a VFD with winding temperature and trips. The only drive model. |
| `streamsim/feeder-meter@1` | `equipment/feeder-meter.component.json` (328 lines) | An **electrical** feeder meter (power sum, billing demand, energy). Not a material feeder. |
| Instruments | `instruments/{accelerometer,rtd,thermistor,current-clamp,power-meter,tachometer,run-status,...}.instrument.json` | Sensors with fault forms (section 3.3). |

There is **no material feeder**, no belt-force model, no pulley, no gearbox, no VFD torque sharing component. `B-9` of
`docs/simulation-canvas/ui-polish-sim-authoring-20261015/simulator-backlog.md:17` records that a pump cannot be driven by
a drive (that row predates the conveyor's `torque_out` port, or is stale: the conveyor does provide `load_torque`, see 1.2).

### 1.2 `streamsim/conveyor@1` [V]

All references are to `equipment/conveyor.component.json`.

States (lines 11-115): `running`, `reset_count`, `grip` (0.3..1, initial 1), `idler_drag` (>=0.5), `material_load` (percent, 0..100),
`overload_ok`, `enable`, `speed_target`, `pulley_speed`, `belt_speed`, `load_demand`, `current_target`, `motor_current`, `current_ratio`,
`power`, `drive_torque`, `vibration`, `throughput`.

Dynamics:

- `material_load` is **F0**: baseline 62 plus two sinusoids (periods 10800 s and 86400 s), line 119. It is an exogenous profile,
  **not** the integral of a feed and not fed from any port. There is no input port for feed rate.
- `overload_ok` is a **latched `trip`** on `current_ratio > 1.35` for 10 s, `reset_on: reset_count` (line 139-155).
- `pulley_speed` is a first-order lag (8 s) on `enable x 85` (line 186); `belt_speed = pulley_speed x grip` (line 202), so slip is
  `belt_speed < pulley_speed`.
- `load_demand = enable x material_load x idler_drag` (line 219); `current_target = 19.425 x enable + 0.333 x load_demand`;
  `motor_current` lags it (3 s) (lines 240-257); `drive_torque = 3.8196 x motor_current` (line 299).
- `throughput = 3 x material_load x 0.0117647 x belt_speed` (line 329), a usable "throughput given up" signal.

Faults (lines 348-440): `belt_slip` (ramp, `grip` -0.15 at 0.25/h), `jam` (step, `idler_drag` x(1+3)), `idler_bearing_wear` (ramp,
vibration +0.4 and drag x1.3), `surge_load` (**negative class**, 900 s pulse, `material_load` x1.4).

Effectors (lines 446-): `set_running` (0/1), `reset_trip`. **No feed, no belt-speed, no torque-bias effector.**

Ports (lines 599-612): `power_out` (type `electrical`, provides `power`) and `torque_out` (type `load_torque`, provides `drive_torque`).
The conveyor **consumes nothing**.

Presets small/medium/large (7.5 kW/50 t/h, 30 kW/300 t/h, 160 kW/1500 t/h); parameter `overload_trip_ratio` (1.1..1.8).

### 1.3 `streamsim/motor-drive@1` [V]

References are to `equipment/motor-drive.component.json`.

States (lines 6-115): `running`, `speed_setpoint` (100..1480 rpm), `reset_count`, `load_torque`, `ambient`, `overload_ok`, `overtemp_ok`,
`enable`, `speed_target`, `speed`, `shaft_power`, `power_target`, `power`, `motor_current`, `current_ratio`, `thermal_load`,
`winding_temp`, `vibration`, `efficiency`.

Dynamics:

- `load_torque` is **F0** (baseline 167.125 Nm + diurnal sine), line 121; `ambient` F0.
- `overload_ok`: latched trip on `current_ratio > 1.35` for 10 s (line 151). `overtemp_ok`: latched trip on `winding_temp > 140` for 10 s
  (line 168). `enable = running x overload_ok x overtemp_ok` (line 185).
- `speed` lags `speed_setpoint x enable` (tau 4 s, line 223). `shaft_power = load_torque x speed x 0.000104712` (product).
  `power` -> `motor_current` (`1.7 x power + 20.535 x enable`) -> `current_ratio` (lines ~300-306).
- `thermal_load = 65 x current_ratio x current_ratio` (line 319): **I-squared heating** is modelled.
- `winding_temp` is a `regime` lag with a **thermal runaway** above 130 C (gain 1.6, tau 600 s vs 900 s), line 336. Inputs: `ambient` and
  `thermal_load`.
- `vibration = 0.5 x enable + 0.0012 x speed` (line 364); `efficiency` is a 1-D `curve` over `shaft_power`.
- `kpis`: `efficiency`.

Faults (lines 419-): `bearing_wear` (ramp 0.1/h, vibration +5), `excess_load` (ramp 1.0/h, `load_torque` x2.5), `winding_overheat`
(ramp 0.5/h, `thermal_load` x2.5), `load_step` (**negative class**, 1200 s pulse, `load_torque` x1.25).

Effectors: `set_running` (line 513), `set_speed` (line 558, 100..1480 rpm), `reset_trip` (line 603). All take a `drive_id` string argument; the
effect is applied to the entity named in the call. **No torque-bias effector.**

Ports (lines 711-726): `load_in` (type `load_torque`, **consumes** `load_torque`) and `power_out` (type `electrical`, provides `power`).
Nothing exposes `enable`, `winding_temp`, a drag or an age state through a port.

Not modelled [V by absence, `motor-drive` dynamics list]: locked-rotor/stall current (current is power-based, so at `speed = 0` the
current is about zero regardless of `load_torque`), breakaway torque, torque limit, slip of the VFD, any per-drive share.

### 1.4 Ports and `connections` [V]

- Contract: `contracts/composition-v0.1.schema.json` (259 lines). `equipment[]` entries have `use`, `as`, `preset`, `parameters`, `names`
  or `id_template`+`count`, `area`, `instruments`; `connections[]` are `{from: "<instance>.<port>", to: "<instance>.<port>"}`
  (schema `$defs.endpoint`). The block that is copied through unchanged is `axes`, `profiles`, `ground_truth`.
- Port record (`contracts/component-v0.1.schema.json`, `ports`; Go type `internal/component/internal/domain/types.go:23-29`):
  `name`, `type`, `direction` (`provides`|`consumes`), `state`, optional `sample_s`. Port **types** are data in
  `internal/component/internal/files/builtin/vocabulary.json` (`port_types`, 19 types incl. `load_torque`, `electrical`); a custom
  library may add its own.
- Compose (`internal/component/internal/domain/compose_connect.go`):
  - `connectOne` (lines 37-51) requires `from` to be a **provides** port, `to` a **consumes** port, and equal port types.
  - `linkSet.add` (lines 84-92) creates one **link type per (consuming kind, port, providing kind)** named
    `<to.kind>__<port>__<from.kind>` and one `links[]` entry per connection.
  - `record`/`followConflict` (lines 94-117) refuse: one link type carrying two different provided states into one consuming state;
    one consuming state sampled at two different `sample_s`.
  - `compose_follow.go:43-53` (`followDocument`): the composed dynamics gets
    `follow: {sources:[{link, state}...], aggregate: "sum"}`, plus `sample_s` from the consuming port. **The aggregate is hard-coded
    `sum`.** A composition cannot ask for `count_running`, `mean`, `min` or `max`, and cannot add an `inputs[].link` read.
  - `applyFollow` (lines 27-41) requires that the consuming kind already has a `dynamics` entry whose `target` is the port's state
    (otherwise: "kind ... has no dynamics for state ... which a connection feeds").
- World semantics of `follow` (`internal/world/internal/domain/coupling_values.go:27-36` `followValue`, `:49-56` `gather`,
  `:67-77` `add`, `:79-90` `value`): while at least one linked, present source instance exists, the consuming state **is** the aggregate
  of the sources' states; otherwise its own dynamics apply. `sum` adds, `count_running` counts sources whose state is `>= 0.5`
  (`runningThreshold`, line 5). Faults are applied **on top** of the followed value (`internal/world/internal/domain/state.go:18-27`
  calls `naturalOf` first, which includes `followValue`, then applies faults).
- Follow restrictions: a state with form `store`, `peak`, `trip`, `anti_recycle` or `sequencer` cannot follow
  (`internal/domain/internal/domain/coupling.go:11`, `dynamicFollow` at 41-50). `lag`, `product`, `sum`, `curve`, `integrator` can.
- The domain layer also supports a **pull** form that composition does not use: an F1 input can read `link: {type, aggregate, default,
  sample_s}` (`contracts/domain-spec-v0.2.schema.json:495-505`; world `inputState` at `coupling_values.go:16-25`). The only test domain
  using `count_running` is `internal/testsupport/testdata/coupled-plant.domain.json:108`.
- **Cycles.** Kinds that read each other form a graph; a cycle without a sampled read is refused
  (`internal/domain/internal/domain/coupling.go:73-80`: "kinds read each other over links in a cycle ... needs a sampled read
  (sample_s)"). A consuming port's `sample_s` makes the read use the state "as it was at the previous sample instant of that grid"
  (`coupling_values.go:58-65` `sampleInstant`), which is what closes a loop (test:
  `internal/world/internal/domain/coupling_loop_test.go`). `B-20` of the backlog (`simulator-backlog.md:28`) records that a composition
  cannot say where the loop is sampled **unless the component's port declares `sample_s`**.

### 1.5 Wiring three motor-drives to one belt: what is possible

Today, as shipped [V + I]:

1. Three `motor-drive` instances, `names: ["D-1","D-2","D-3"]`, in **one** equipment entry share one kind, one preset and one set of
   parameters. [V: `names` array, schema above; I: per-instance preset needs separate entries with distinct `as` names, each
   becoming its own kind, `as` at most 20 characters.]
2. `conveyor.torque_out -> D-n.load_in` is type-legal. Each drive's `load_torque` would then **equal the conveyor's whole
   `drive_torque`** (`sum` over one source), not a share of it. Nothing divides by the number of running drives. [V: `followDocument`
   always `sum`; I: consequence.]
3. The belt (conveyor) never sees the drives: it consumes nothing, so the drives' `enable`, `power`, `winding_temp` do not come back.
   Its own motor is a fourth, independent motor. So **the shared-load scenario is not expressible with the shipped library.**

With data-only changes (library additions under the engine's existing rules) [I, to be proven by a test]:

- A new **`belt`** equipment file (new component, `streamsim/belt-conveyor@1`) with states
  - `force_demand` (algebraic: breakaway/running force from material mass, see 2.4),
  - `drives_running` (F0 baseline, fed by a consuming port `drives_in`; with `sum` over each drive's 0/1 `enable` this **is** the running
    count),
  - `share` (algebraic `product` of `force_demand` and a `curve` over `drives_running` with table x `[0,1,2,3,4]`, values
    `[0,1,0.5,0.3333,0.25]`, which is exact at integer counts, `algebraic.go:39-48` `curveValue` is piecewise linear and clamps at the ends),
  - and a provides port `share_out` (type `load_torque`) on `share`.
- `motor-drive` gains a provides port `enable_out` on `enable` (new port type, e.g. `run_state`, in the vocabulary), and its existing
  `load_in` takes `belt.share_out`. Each drive then follows the same `share` value; a drive with `enable = 0` simply stops reading torque
  because `shaft_power` is `load_torque x speed` and `speed` is 0.
- The loop (drive `enable` -> belt `drives_running` -> `share` -> drive `load_torque` -> trip -> `enable`) needs `sample_s` on the
  belt's consuming port `drives_in` (the composer copies it, `compose_follow.go:49-51`). The one-sample lag in the running count is
  physically a VFD reaction time; pick a sample of a few seconds.
- Because faults apply on top of followed states (`state.go:18-27`), a **per-drive** fault on `load_torque` (a seized gearbox adding drag,
  `excess_load`) works on the faulted drive; for it to also reach the survivors the belt must read the sum of drives' drags through a
  second port (another `sum` follow) so that the survivors' share rises. [I]
- `names` plus the three unmerged commits on `canvas-plant-view-instances` (`b6b3a946`, `7153d0cc`, `c75285c1`, all dated 2026-11-05,
  checked with `git log origin/main..canvas-plant-view-instances`): they move instance-set expansion into the `domain` module,
  add `streamsim domain population`, and teach the **canvas** to draw a plant declared by `id_template` (previously it showed 0
  instances). **They do not matter for the simulator side** of this use case if the drives are declared with explicit `names`
  (`D-1`,`D-2`,`D-3`); they matter only if the canvas must draw a templated population. Note `compose_connect.go:77`
  (`specs.InstanceRenders`) is already on `origin/main`.

Cross-instance aggregation summary [V]: core supports `sum|mean|min|max|count_running` over linked present instances, in `follow` and in
`inputs[].link`; composition exposes `sum` only.

## 2. Dynamics forms: what exists and is implemented

### 2.1 Inventory [V]

Enum at `contracts/domain-spec-v0.2.schema.json:463`: `first_order_lag, rc_network, dead_time, saturation, hysteresis, integrator,
threshold_integrator, store, regime, peak, trip, anti_recycle, sequencer, curve, product, sum, pid, alarm`. Equations are fixed in
`docs/design/TECHNICAL_DESIGN.md:311-355` (first seven forms) and the schema descriptions for the later forms.
Implemented in the world (not just specified):

| Form | Implementation | Note |
|---|---|---|
| `first_order_lag`, `integrator`, `threshold_integrator` | `internal/world/internal/domain/integration.go:88-98` (`f1Derivative`) with RK4 in `integration_forms.go:119-127` | `integrator`: `dx/dt = gain x Sum coef x input`; signed coefficients allowed (`inputValue`, integration.go:100-110) |
| `rc_network` | `integration_forms.go:96-117` | two cascaded lags, RK4 |
| `dead_time` | `integration_forms.go:83-94` and a ring buffer sized `ceil(delta/dt)` (`state.go:~95-125`, `delayCapacity`) | **the schema requires `time_constant_s` for `dead_time`** (`domain-spec-v0.2.schema.json:570-573`), so a pure delay is a delay plus a (small) lag |
| `saturation`, `hysteresis` | `integration_forms.go:61-81` | clamp bounds are **constants** in the spec (`clamp.min/max`), not states |
| `curve` (1-D, 2-D bilinear), `product`, `sum` | `algebraic.go:9-58` | algebraic: no `dt_ms`; `clamp` applies; `product` multiplies `coef x state` terms, there is **no division** |
| `trip` (latched/unlatched, `for_s`, `reset_on`), `anti_recycle`, `sequencer` | `protection.go`, `sequencer.go` | |
| `regime`, `peak`, `store` | `regime_peak.go`, `stores.go` | |
| `pid`, `alarm` | `controller.go`, `alarm.go` | |
| **F2** reference models (`arrhenius_spoilage`, ...) | **Not implemented.** `internal/domain/internal/domain/validation_records.go:79-81` refuses `tier: F2`: "specified-but-unbuilt seam; not_implemented". F3 likewise (line 76-78) | |

Stability guard: an F1 lag/rc/dead_time/store step must satisfy `dt/tau < 2.78` for RK4 (`validation_stability.go:8-9, 25-40`). Each
dynamics entry has its own `dt_ms`, so hours-long ageing states can step at a coarse `dt_ms` while electrical states step at 1 s [V].

### 2.2 Can we express each needed piece?

| Need | Verdict | How, or why not |
|---|---|---|
| `share = total / n_running` | **Yes for equal sharing, [I]** | `curve` of the integer running count to `1/n` (table x `[0,1,2,3,...]`), then `product` with total. Exact at integer counts because the count from `sum` of 0/1 states is exact (`curveValue`, `algebraic.go:39-48`). Not expressible: a general `a/b` of two live states. |
| Unequal sharing (torque bias, droop, drive-specific weight) | **No** | needs `w_i / Sum w_j` = division of two states. A 2-D `curve` grid over `(F, Sum w)` is bilinear and inexact between nodes (`algebraic.go:43-47`). Needs a new named form (see section 8). |
| Arrhenius ageing rate | **Approximately** | a 1-D `curve` over `winding_temp` giving relative ageing rate (e.g. 2^((T-Tref)/10)); linear interpolation between table nodes, **clamped flat outside the table** (`algebraic.go:50-54` `segment`), so no true exponential tail. Table density sets the error. The F2 `arrhenius_spoilage` shortcut is not available (2.1). |
| Integrated life consumed | **Yes** | `integrator` of the ageing-rate state (`integration.go:92-93`), or `threshold_integrator` with `direction: above` to count only the overtemperature part. Its hidden state is the truth; also a candidate KPI (section 5.4). |
| Transport delay of material along the belt | **Partly** | `dead_time` with a fixed `dead_time_s` (30 s for 60 m at 2 m/s) plus a small lag; material on belt = `integrator` over `feed (+1)` and `delayed_feed (-1)`. A **speed-dependent** delay (L/v with v changing) is not expressible: `dead_time_s` is a constant in the spec (`delayCapacity`, `state.go:~120-130`). Hold a constant belt speed or accept the approximation. |
| Breakaway vs running force | **Yes (algebraic)** | a 2-D `curve` over `(belt_speed, loading)` returning a force factor that is higher at `speed ~ 0`; or `hysteresis`/`regime` on `belt_speed`. Must be added as belt-component data. Missing today: `motor-drive` has no stall current, so the drive-side "cannot restart" has to be a separate algebraic margin state, see 2.5. |
| Pulley slip limit | **Partly** | an unlatched `trip` on `share` (or force per drive) `> limit` sets `grip` to a reduced value (`conveyor.grip` already feeds `belt_speed`, conveyor.component.json:202). The limit is a constant or an offset from another state (`condition.threshold_state`, schema `$defs.condition`). No capstan (e^(mu x theta)) equation, but the limit can be a constant per preset. |
| Seized drive adds drag | **Yes** | a step fault with `affects` on a `drag` state of the drive; the belt reads the sum of drives' drags through another port, so the survivors' share rises. Needs new ports (section 1.5). |
| A drift crossing a hidden threshold then trips | **Yes** | a `ramp` fault on `current_ratio`, `load_torque` or `thermal_load` raises the hidden value while the visible channels move slowly; the existing latched `trip` (`for_s`) fires when the threshold is crossed. Faults apply to the value every reader sees (`state.go:18-27`), so the trip sees the faulted value. Example already shipped: `motor-drive.winding_overheat` ramp + `overtemp_ok` trip (motor-drive.component.json:168, 464). |
| Cumulative damage independent of a trip | **Yes** | the ageing chain above; no trip involved. |
| A latent state that only shows in unobserved channels (survivor life) | **Yes** | hidden state, no channel (state is only observed through channels, `domain-spec-v0.2.schema.json` `kinds[].state` description). |

### 2.3 Fault envelopes [V]

`internal/world/internal/domain/dynamics.go:44-146`: `step`, `ramp` (`min(rate x hours, 1)`), `exponential` (`1 - e^(-rate x t)`),
`intermittent` (duty/period), `stochastic` (a random walk using a substream seeded by fault id, entity, onset and state:
`state.go:~110` `faultSubstream`), `pulse` (requires `duration_s`). Contribution is `add` or `mult` (`value x (1+mult) + add`,
`state_drivers.go:25-39`). A fault takes `severity` as its only parameter (`faults.go:55-70, 118-125`: "has no parameter ... (declared:
severity)"). The ramp saturates at 1, so a slow degradation has a bounded size (set by `delta`).

### 2.4 A candidate belt force model, in existing forms [I]

`mass_on_belt = integrator(feed - dead_time(feed, 30 s))`; `friction_force = k1 x (belt_mass + mass_on_belt)`; breakaway/running
factor via 2-D `curve(belt_speed, mass_on_belt)`; `share = force x curve_1/n(drives_running)`. All states hidden. Feed rate is an F0
setpoint state moved by an effector (section 4).

### 2.5 "Stranded belt" as a state [I]

`restart_capacity = Sum over healthy drives of capability` (a `sum` over a port), `restart_margin = curve(restart_capacity, mass_on_belt)`
(2-D grid: required vs available), `stranded = trip(when restart_margin < 1 and belt_speed < eps, latched)`. All closed forms. The
"cannot restart" outcome is then a hidden state that can be sealed as truth (section 5.4).

## 3. Faults and perturbations

### 3.1 World (degradation) faults [V]

- Declared per kind: `contracts/domain-spec-v0.2.schema.json` `$defs.fault` (lines ~862-1000): `id`, `onset.shape`
  (`step|ramp|exponential|intermittent|stochastic|pulse`), `rate_per_hour`, `magnitude`, `duty_cycle`, `duration_s`; `affects[]`
  (`state`, `delta`, `multiplicative`) or `observation` (sensor fault); `observability.detector`; `is_negative_class`; optional
  `expected_effector`, `deadline_s`, `counterfactual` text.
- Fault ids are unique across the whole domain (schema description of `kinds`); compose qualifies library fault ids per kind
  (`internal/component/internal/domain/compose_qualify.go`).
- Injection: `World.InjectFault(entity, fault, onsetNS, params)` (`faults.go:28-45`); a fault must be declared by the entity's kind
  (`faults.go:35-37`). Onset can be in the future and a fault can be cleared (`fault.clear`).
- **Scheduling by seed.** Two routes: (a) a scenario/command log lists `fault.inject` ops with explicit `onset_ns`
  (`contracts/scenario-v0.1.schema.json`, `$defs.fault_inject`); (b) the suite generator draws entity, fault (by profile `fault_weights`)
  and onset from the seed: `internal/suite/internal/domain/scenario_draw.go:20-31` (`drawOnset`: pre-degraded with probability
  `pre_degraded_fraction`, onset before start; else uniformly in the first half of the duration) and `:44-60` (`weightedFault`).
  Campaigns draw several faults per run (`campaign_spec.go:23-37`). The stochastic envelope draws from a fault-specific substream, so
  the same seed reproduces it.
- Jobs: `sim.job.start` takes `faults: [{entity_id, fault, at_s, until_s}]` (`documentation/reference/mcp.md` "sim.job.start details"). In
  an Evaluate job the director hides the plan until the job ends (same section).

### 3.2 What this gives the scenarios [I]

| Scenario (README) | Expressible fault shape |
|---|---|
| 1. Slow bearing failure on drive 3 | `bearing_wear` already exists on `motor-drive` (ramp 0.1/h on vibration); add a coupled ramp on `thermal_load`/drag and the ageing chain. |
| 2. Sudden trip under full load | a `step` fault on `running` or on `current_ratio`; the latched `trip` and `reset_trip` already exist. |
| 3. Seized gearbox | `step` fault on a drive `drag` state (new state) so the shared belt sees extra resistance. |
| 4. Feed surge | the negative-class pulse `surge_load` (conveyor) / `load_step` (motor-drive) already model "all currents up together, nothing wrong". |
| 5. Sensor faults hiding degradation | instrument faults (3.3) combined with scenario 1. |

### 3.3 Sensor faults and the perturb layer [V]

- Sensor (instrument) faults are **domain faults** with `observation`: `stuck`, `freeze`, `dropout`, `scale`, `noise`, with optional
  `quality` (`good|suspect|out_of_range|stale|disconnected`) (`domain-spec-v0.2.schema.json:871-882`). Instrument files add `offset` and
  `drift`, which act through a hidden `<point>__bias` state (`instrument-v0.1.schema.json:124-135`; `compose_faults.go:14-49`
  `biasEffect`: `offset` is a step on the bias; `drift` is a ramp whose `rate_per_hour = 3600 / duration_s`).
- Shipped examples: accelerometer `cable_break` (stuck low, `disconnected`), `dropout`, `mount_noise` (noise x15, `suspect`),
  `loose_mount` (scale 0.7); rtd `offset 2.0`, `drift 0.05 over 86400 s`, `open_circuit`, `dropout`, `noise_burst`; current-clamp
  `offset`, `drift`, `clamp_open`, `dropout`, `harmonics`, `ct_ratio` (scale 1.25). (Loaded from the instrument JSON files in
  `.../instruments/`.)
- `observation_bias` on a channel makes a fault on a hidden state change what a channel reads (a fouled probe reads high and stable,
  `domain-spec-v0.2.schema.json` `channel.observation_bias`, noise scaling to `noise_scale_at_full`).
- **Delivery perturbations** are a separate layer (`internal/perturb/internal/domain/perturb.go:19-38`): `duplicate_burst`, `id_reuse`,
  `reorder`, `delay_tail`, `gross_backfill`, `drop`, `clock_skew`, `non_monotonic`, `out_of_enum`, `out_of_range`, `unit_mismatch`,
  `oversize`, `malformed`, `nan_inf`, `storm`, `producer_flap`, `time_encoding`, `precision_edge`, `injection_probe`. They change what the
  observer receives, never the process (MCP doc, `sim.catalog.perturbations`). They are applied by `sim.perturb.apply` with a window and
  recorded in the ledger (non-negotiable 4), so "a stuck vibration channel" (a sensor fault, truth-bearing) and "a dropped vibration
  stream" (a perturbation, delivery-only) are distinguishable in scoring.
- Known behaviour gap recorded by the authoring probes: `dropout` freezes the value with no staleness flag in the canvas
  (`simulator-backlog.md:25`, `B-15`) and a stuck sensor is "to probe" (`B-7`). Whether `quality` reaches the adapter output was not checked.

## 4. Effectors and commands

### 4.1 Effector contract [V]

`domain-spec-v0.2.schema.json` `$defs.effector` (lines ~1000-1100): `name`, `args_schema`, `risk_class`, `safe_stop`, `ack`
(`latency_ms`, `failure_modes`: `ok|slow|ack_lost|reject|partial|confirmed_no_effect|silent_no_effect`), `interlock` (refuse, plus
autonomous action), `responder` (a simulated crew after `delay_s`), `lease` (hold for N seconds, fallback `safe_off|hold_last|revert`),
`requires` (preconditions on states), `confirmation_channels`, `effect.state_deltas` (`delta`, `from_arg`, `assign_from_arg`) with
`time_constant_s` and `dead_time_s`, `idempotency_window_s`. Effects are "kicks" that add to or replace the value of any state as read
(`dynamics.go:12-42`, `state_drivers.go:41-94`), so an effector can assign an F0 setpoint or `running`.

### 4.2 Mapping the actions the README wants

| Action | Today | Needed |
|---|---|---|
| Stop or start one drive | `motor-drive.set_running` per entity (motor-drive.component.json:513) | none |
| Reset a tripped drive | `motor-drive.reset_trip` (line 603) | none |
| Change belt speed | `motor-drive.set_speed` per drive (line 558); the conveyor has none | belt-level speed: each drive's setpoint should follow one belt value, or one effector on the belt kind |
| Cut feed / run belt empty | **none** (conveyor `material_load` is an F0 profile with no setpoint, conveyor.component.json:119) | a feed component (F0 `feed_rate` setpoint + an effector `set_feed`); the belt integrates feed through the delay |
| Torque bias between drives | **none**, and needs a division to be useful (2.2) | weight state per drive + a division form |
| "Stop drive 3 at a chosen time" | The consumer calls `set_running` at that clock; the simulator **cannot date an operator call** (`documentation/reference/mcp.md`, `sim.effector.invoke` row: "an operator cannot date a call") | none in the simulator: the consumer schedules its own call; for a pre-planned call use a job's `calls: [{entity_id, effector, args, at_s}]` (director side) |
| Alarm only | no effector needed; the verdict (`sim.consumer.report`) carries detections | none |

### 4.3 Director vs operator surfaces [V]

`documentation/reference/mcp.md`:

- Director (stdio): catalog, world create/describe/destroy, `sim.clock.advance`, `sim.fault.inject|clear|list`, `sim.perturb.apply|clear`,
  `sim.run.begin|end`, `sim.truth.seal|reveal|seal_status`, `sim.score`, `sim.scenario.audit`, `sim.job.*`, `sim.domain.compose`,
  `sim.world.invoke` (director invokes an effector).
- Operator (HTTP, capability token): `sim.nameplate.read`, `sim.effector.list`, `sim.effector.invoke` (needs `command_id`; acts at the
  current clock), `sim.consumer.report` (quiescence + verdict, never a score). The operator view holds no truth store, fault registry,
  perturbation log or hidden state (mcp.md, "Operator tools").
- Time: stepped clock; `sim.clock.advance` with `await_consumer` waits for the consumer's quiescence report (non-negotiable 5).

### 4.4 Plant wire [V]

`contracts/plant-wire/SPEC.md`: the simulator is a **node** that speaks to a host over an owner-only unix socket: births from the domain
through an allow-list, data, `NSYNC`/`HSYNC` barrier (lines 70-80, 153-157), host `DCMD` commands with `deadline_ns`, `not_before_ns`
(a **scheduled** command exists here, line 78 and 188), `authority_epoch`, `policy_digest`, `fallback_on_lease_loss`; commands the
node declares `safe_stop` skip the freshness and policy checks and are never queued (lines 205-206). A domain kind can `implements` a kind
file so the simulated device and a real board are two implementations of one wire kind
(`documentation/architecture/physical-devices.md`). For this use case the plant wire is the route by which agentic-stream would drive a
belt exactly as it drives the fan; no change to the wire is needed for simulated drives. [I]

## 5. Truth, oracle and grading

### 5.1 How ground truth is produced [V]

- A scenario is one entity + one fault + onset (`suite/internal/domain/scenario_draw.go`). `truth.BuildRecord`
  (`internal/truth/internal/domain/record.go:31-44`) runs the **observability solver**: two **noiseless** worlds (clean and faulted)
  advanced in `sampleNS` steps (default 60 s, horizon 72 h, `solver.go:27-35`), comparing the detector quantity against
  `first_observable_snr x sigma` and `unavoidable_snr x sigma` (`solver_scan.go:55-78`). Output: `first_observable_time_ns`,
  `unavoidable_time_ns`, `effective_sigma`, channels (`ground-truth-v0.1.schema.json`).
- The solver is **numeric** (`Method: model.SolveNumeric`, `solver_scan.go`); `SolveAnalytic` exists as a constant
  (`internal/model/truth.go:47-48`) and is not produced by the solver [V by grep]. The "analytic cross-check" non-negotiable is a **test**
  of the production integrator against an independent recurrence (`documentation/architecture/invariants.md:10`; examples:
  `coupling_loop_test.go` `newCoolingReference`, `stores_oracle_test.go`, `coupling_oracle_test.go`, `algebraic_oracle_test.go`).
- The record also has `deadline_ns = first_observable + fault.deadline_s` when the fault declares `deadline_s`
  (`record.go:91-93`), `expected_effector`, `is_negative_class`, `pre_degraded`, `perturbations`, `counterfactual` **text**
  (200 chars), and `trivial_baseline_verdict` (`ground-truth-v0.1.schema.json`, `additionalProperties: false`).
- Lifecycle: install truth, seal, run, consumer verdict, close run, score while sealed, then reveal/unblind
  (`documentation/architecture/truth-and-scoring.md`).

### 5.2 Scoring [V]

`internal/score/internal/domain/score.go:20-79` Scorecard: `instrument` (ledger complete, counts), `consumer` (duplicate handling, lateness,
evidence grounding, action fidelity, interlock handling), `loop` (`resolved`, `time_to_resolution_ns`, `deadline_adhered`,
`unnecessary_action`, `action_appropriate`, `false_success`, `silent_no_effect` counts), `judgment` (`detected`, `label_correct`,
`detection_latency_ns` against `first_observable_time`, `suspicious` if before it, `false_positive` on a negative class).
Specifics:

- `loop.go`/`loop_metrics.go`: a call is "right" only if `call.Effector == gt.ExpectedEffector`, `call.EntityID == gt.EntityID` and the
  effect applied (`rightCall`); `resolved` is recovery of **the first state the fault affects** (`fault.Affects[0].State`) halfway back
  to baseline (`recoveryTarget`, `recoveryLevels`); `deadline_adhered` needs a right call before `deadline_ns`.
  `unnecessary_action` = any effector call on a negative-class scenario.
- So scoring is **single entity, single fault, single expected effector, state recovery**. The consumer's real lever here (feed, torque
  on other drives) is neither "right" nor recovered.
- Trivial-baseline audit (`internal/audit`, `ground-truth-v0.1.schema.json` `trivial_baseline_verdict`): fixed threshold, z-score, first
  difference, moving-median residual and channel silence, each fitted with hindsight on the scenario's channels. It audits
  **detection** scenarios; it does not audit policies (e.g. "always cut feed").

### 5.3 KPI truth, the closest thing to outcome truth [V]

Kinds may declare `kpis` (state names); the run seals each entity's true `mean`, `min`, `max`, `samples` over the run
(`internal/model/kpi.go:10-22`, `world/internal/domain/kpi_truth.go:13-26`, `kpi_track.go:26-41`) in the director journal, plus ISA-18.2 alarm
KPIs. For a monotone state (life consumed, energy) `max` is its final value. This is **a per-entity number in the director journal**,
not a score: nothing compares a consumer to it.

### 5.4 Safety boundaries are declared, not scored [V]

`kinds[].boundaries[]` name a trip state whose tripped value means a boundary violation, with severity
(`domain-spec-v0.2.schema.json`, `kinds.boundaries`). `World.ViolatedBoundaries` (`protection.go:126-140`, facade `world/operations.go:89`)
computes them, but **no non-test code outside the world calls it** (repo-wide grep). So a `stranded` boundary could be declared today
and nothing would read it.

### 5.5 What the use case's truths need

| Wanted truth/metric | Exists? | What it would take |
|---|---|---|
| "The time drive 3 would trip if nobody acts" (lead-time reference) | No. The record has `first_observable`/`unavoidable`/`deadline_ns` from a fixed `deadline_s`, not a computed trip time | Extend the oracle with a third noiseless "no-action" world; scan until a named trip state reaches `tripped_value`; add a field to `ground-truth` (the schema is `additionalProperties: false`, bump it; pre-release, allowed). Truth module, domain layer (`internal/truth/internal/domain/solver_scan.go` sibling), plus `model`, contract, fixtures. Medium. |
| Lead time of a consumer's warning | Partly: `detection_latency_ns` is against the noise-floor onset, not the trip | scoring metric = `trip_time - detection_time`; score module. Small once the truth above exists. |
| Life consumed on survivors, with vs without the consumer's actions | KPI mean/min/max of a state, one run only | needs a baseline (counterfactual) run for comparison and a cross-entity cost metric in `score`. Medium. |
| Never strand a loaded belt | Declarable as `boundaries` + `trip`, never scored | a score policy that reads violated boundaries over the run (`ViolatedBoundaries` already exists). Small to medium. |
| Throughput given up | `conveyor.throughput` state exists | KPI compare to a counterfactual baseline; score metric. Small to medium. |
| False alarms | `is_negative_class` faults + `judgment.false_positive` + `loop.unnecessary_action` exist | works for a **single entity**; needs per-run false-alarm counting across entities/time (`score`). Small to medium. |
| Expected-action grading for non-faulted entities | No (`rightCall` is entity-bound) | allow the expected effector to name a different entity or a set; truth + score. Medium. |

## 6. Closest existing domains and the joined-run pattern

### 6.1 Domains [V]

- `domains/rotating-machinery.domain.json` (actuation `advisory`): one kind `pump` with F0/first-order-lag states; faults
  `bearing_wear` (ramp), `misalignment`, `overload` (step), `lubrication_due`, `sensor_drift`, `sensor_failure`, `transient_none`;
  effectors `schedule_maintenance`, `reduce_load`, `stop_motor`. Instances via `id_template` `site-{site}/pump-{n}`; **no links**, so
  nothing is shared between machines.
- `domains/discrete-line-oee.domain.json` (actuation `interlocked`): one kind `assembly_line` (`curve`, `integrator`, lag); faults `jam`,
  `speed_drop`, `downtime_cascade`; effectors `clear_jam`, `adjust_speed`, `reset_line`. No links.
- Coupled example worlds: `internal/testsupport/testdata/coupled-plant.domain.json` (links, `sum/mean/min/max/count_running`),
  `examples/factory-utilities.domain.json` and `examples/compositions/{factory-utilities,chiller-pilot}.composition.json`
  (the only many-to-one coupling by composition: `P-1..P-3.flow_out -> CH-1.flow_in`, `M-1..M-4.heat_out -> CH-1.load_in`, summed).
  `canvas/e2e/fixtures/boiler-steam-loop.composition.json` instantiates `motor-drive` twice but does not connect its `load_in`.
- Nothing in the repo composes `streamsim/conveyor` with `streamsim/motor-drive` (grep).

### 6.2 The fan experiment as the pattern [V]

`docs/physical-devices/fan-experiment/` (GOAL.md, PLAN.md, HANDOFF.md): the chain is
`Tamoz <-> Agentic Stream <-> plant wire <-> streamsim (zone, LED, fan twin)`; "real" differs by a session plan only
(`examples/devices/thermal-board/fan-experiment/{simulate,real}.plan.json`). Artifacts: kind files
(`examples/devices/thermal-board/{fan,indicator,thermal_sensor}.kind.json`), a board mapping (`arduino-thermal.board.json`), a domain
implementing them (`thermal-board.domain.json`, `arduino-thermal.bench.domain.json`). Running rules the owner set: no consumer knowledge in
the binary; agentic-stream needed one generic plant-wire host module (PLAN.md section 1, R3). The simulated rung passed all seven criteria
(HANDOFF.md section 1).

Reuse for the belt: a `belt.domain.json` (composed) whose kinds `implement` kind files for drive/belt/feeder metrics and commands, a
session plan, and the same host. The simulator side needs no new wire feature (4.4). The consumer side (agentic-stream's host module is
generic) is the other repos' note. [I]

## 7. Canvas

[V] The canvas (`canvas/`, desktop only, `documentation/guides/canvas.md`) opens a domain and draws **instances as nodes grouped by
area and typed links as edges**; you can browse the library, place and name instances, set parameters, connect ports, see type and
cycle errors pinned to elements, run in Explore/Evaluate, inject faults and perturbations, and review runs. It computes nothing the
binary owns (`AGENTS.md`, "Canvas"). A plant of 3 drives + belt + feeder + sensors would draw and run through this path if the domain exists;
the canvas needs no feature for it. Two caveats [I]: (a) a plant declared with `id_template` needs the unmerged canvas commit `c75285c1` to
list its instances (use `names` to avoid it); (b) the belt-length/transport picture (material on the belt) would be states and charts, not
a belt-shaped view; no belt widget exists.

## 8. Gaps this use case exposes

Sizes: S under a day of agent work with tests, M a few days, L a design round plus implementation. "Blocked" lists the architecture rule
that constrains the shape, not a reason to refuse.

| # | Gap | Where it lives (module/layer) | Size | Constraint |
|---|---|---|---|---|
| G1 | **Belt component with drive coupling**: `streamsim/belt-conveyor` equipment (force, mass on belt via `dead_time`, `drives_running`, `share`, slip, restart margin), new port types, ports on `motor-drive` (`enable_out`, `load_in` unchanged), `sample_s` on the belt's `drives_in` | `internal/component/internal/files/builtin` (data) and `vocabulary.json` | M | Data only; no Go change expected for the equal-share case. A new library file is allowed (it is data in the module's edge package, MUST 4). Each form used already has an oracle; add a world-level oracle test for the shared-load loop (non-negotiable 2). |
| G2 | **Feed component and effectors**: `feeder` with `feed_rate` setpoint, `set_feed` effector, optional hopper store; belt integrates it | component data | S | Data only |
| G3 | **Per-drive effectors**: torque bias / share weight state + `set_bias`; belt speed as one shared state followed by the drives (speed coupling) | component data, plus G4 for bias | S (speed), blocked by G4 (bias) | |
| G4 | **Division** (`a / b` of two states) for unequal torque sharing, droop, `w_i / Sum w`; today only `curve` of an integer count approximates it | new F1 form in `internal/model` (enum), `internal/domain` (validation), `internal/world/internal/domain/algebraic.go` (evaluator), contract schema, `TECHNICAL_DESIGN` §6.2 equation, an independent oracle test | M | Closed named form, not an expression language; precedent: `curve` and `product` were added the same way. Needs a design note and divide-by-zero rule (value of a safe default or clamp). 15-line functions, 300-line files apply. |
| G5 | **Compositions can only `sum`**: no `count_running`/`mean`/`min`/`max`, no per-instance parameters, no pull `inputs[].link` | `internal/component/internal/domain/compose_connect.go`, `compose_follow.go`, `contracts/composition-v0.1.schema.json` | S-M | Workaround exists (sum of 0/1 enable). Add an optional aggregate on the consuming **port** in component data (`aggregate` next to `sample_s`). Not required for the first version. |
| G6 | **Pure delay** (`dead_time` requires `time_constant_s` in the schema; delay is constant) and no speed-dependent transport delay | schema, `integration_forms.go`; variable-delay would be a new form | S (schema relax) / M (variable delay) | Constant delay is enough for the first version. |
| G7 | **F2 reference models are refused** (`arrhenius_spoilage` etc.) so real ageing exponentials are not available; `curve` approximates | `internal/world` + `internal/domain` | M-L | `curve` table is the data-only route; an F2 implementation needs a published-vector unit test (`TECHNICAL_DESIGN` §6.2) |
| G8 | **Locked-rotor / stall current and breakaway torque in `motor-drive`**: current is power-based, so a stalled belt draws no current | component data (motor-drive revision: add `stall_current` via a `curve` on `speed` and `load_torque`) | S-M | Data only; changes motor-drive digests and presets (pre-release, allowed) |
| G9 | **Truth: counterfactual outcome truths** ("time a named trip state would fire if nobody acts", "state at end with and without action", "stranded") | `internal/truth/internal/domain` (solver), `internal/model/truth.go`, `contracts/ground-truth-v0.1.schema.json`, `internal/director` (seal/reveal), fixtures | M-L | Must stay sealed; operator view must not gain a field (`test/architecture`, mcp_protocol). Computed numerically, deterministically, from noiseless worlds (existing technique). |
| G10 | **Scoring beyond one entity/state**: lead time to trip, cross-entity cost (survivor life, throughput), stranding via `boundaries`, false alarms across the run, expected action on another entity | `internal/score/internal/domain` (`loop.go`, `loop_metrics.go`, `judgment`), `internal/grading`, `contracts/consumer-verdict-v0.1.schema.json` possibly, `internal/report` | L | Online/offline parity of the scorecard must hold (`truth-and-scoring.md`); scorecard bundle version bump (`scoreBundleVersion` in `score.go`). |
| G11 | **Boundaries unused**: `ViolatedBoundaries` has no consumer; a `stranded` boundary is never scored | `internal/run` or `internal/score` reads it each advance; sealed in the journal | S-M | Part of G10 |
| G12 | **Multi-fault / correlated scenarios in the generator**: the suite draws one fault per scenario; this use case needs "drive 3 degrades, then trips, survivors age", and a feed surge to be confused with it | `internal/suite/internal/domain` (`generation.go`, campaigns allow several faults per run: `campaign_spec.go`) | S-M | Campaigns already draw multiple faults; the single-label ground truth (G9/G10) is the real block. |
| G13 | **Observability detector for load-sharing** (`peer_residual` over the three drives' currents exists) | no change needed: `observability.detector.form = peer_residual` is already defined and implemented | S | Use it for "drive 3 differs from its peers" and `conservation_residual` for "sum of drive torques vs belt demand". |
| G14 | **Silent/late actuation realism** (`silent_no_effect`, `lease`, `requires`) is available; a "stop drive 3" that is acked but not applied is expressible | no change | none | Use as a scenario knob. |
| G15 | **Canvas**: nothing required; templated populations need `c75285c1` | `canvas/` | none (S to merge) | |

Gaps blocked by architecture rules, in one list:

- No expression language -> G4 must be a closed named form with an equation and an oracle, not a formula string.
- Domains are data -> every belt/feed/drive behaviour above has to be component or domain JSON; the only Go changes are G4, G5 (small),
  G7 (optional), G9-G11.
- 15-line functions and 300-line files -> the solver extension (G9) and the score policies (G10) must be split into named steps and files
  in their modules; no logic in `internal/cli` or `internal/mcp` (MUST 1 and 2).
- The simulator has no consumer knowledge -> the controller (predict, cut feed, shift torque, stop at a time) lives in agentic-stream or
  tamoz; the simulator only exposes effectors and truth.
- Never loosen a gate -> new score fields and truth fields need contract/fixture updates, not allow-list edits.

## 9. Suggested build order for the simulator side (a proposal, not a decision)

1. **Data first (G1, G2, G8-light, G13):** belt, feeder, `motor-drive` ports, ageing chain with a `curve`, a hand-written or composed
   `belt-conveyor-drives` domain with 3 drives; prove the shared-load loop against an independent recurrence (world oracle test); then run
   `domain lint` and the canvas. No Go changes except tests.
2. **Truth (G9):** counterfactual "no action" world in the solver for a named trip state; seal `trip_time_ns` (and life consumed via KPI
   max); extend the ground-truth contract.
3. **Scoring (G10, G11):** lead time, survivors' life vs baseline, stranding via boundaries, throughput given up, false alarms.
4. **Optional (G4, G5, G7):** division form for torque bias; port-level aggregate; Arrhenius F2.

Open questions that change the plan (from README "Open questions for the customer"): whether the VFDs share by torque or speed droop
(droop makes G4 mandatory), the real drive ratings and trip thresholds, and which effectors the customer allows (feed is the strongest lever and
needs G2).
