# P7k catalog spec (working brief)

Authoring brief for the built-in library under `components/`. Rules: [COMPONENTS_V0_1](../design/COMPONENTS_V0_1.md), world
forms: domain-spec v0.2 (§13 stores, §16 schedules and site clock, §18 supervisory, §21 links and `follow`, §23 `curve` /
`product` / `sum`, §24 `pid`, §25 observation faults, §26 `alarm`). Fidelity goal: decisions matter correctly, not engineering
accuracy. Units SI-ish as in `components/vocabulary.json` (extend the vocabulary when a family is missing).

## Every component

- `id` `streamsim/<name>`, version 1, ISO 14224 `class`, tags, presets `small|medium|large` (complete variants: every size-
  dependent coefficient, range and fault delta), `default_preset: medium`, 1–3 parameters (setpoints, enum options).
- States with honest initial values at nominal steady state for each preset (a nominal day must not start with a big transient).
- Dynamics only from the vocabulary; sample steps ≥ 1 s and within the stability bounds (validator checks them).
- `running` (0..1) where the equipment starts and stops; power, flow and heat scale with it through `product` (R7g-02).
- A KPI state (algebraic) where the trade has one: specific power (kW per unit output), COP, approach temperature, efficiency.
- Measurement points for everything a real plant meters, with ISA-5.1 tags and default instruments; optional points for the
  rarer ones.
- Ports for what it exchanges (`provides` a state / `consumes` into a state, typed by the vocabulary).
- Faults (3–5): real failure modes, ISO 14224 names, detectors on its own points, honest SNRs; at least one negative-class event
  where a real one exists (a heat wave, a rush order).
- Effectors (1–4): start/stop, setpoint, mode; acknowledge/shelve for alarms where it has one; leases for timed actions; no
  descriptions copied into domains (library text only).

## Equipment (16)

| Name | Group | Core states and forms | Ports | Faults | Actions |
|---|---|---|---|---|---|
| `chiller` (exists) | Cooling | add `running`; power and heat via `product`; COP via `curve` of load fraction | provides chw_supply; consumes load, condenser water temp | condenser fouling, refrigerant undercharge, supply sensor drift (instrument) | start, stop, setpoint |
| `cooling-tower` | Cooling | basin temp lag toward wet bulb + approach (curve of fan speed and load) | consumes heat_rejected, wet bulb; provides basin temp | fan failure, fill fouling | fan speed setpoint |
| `pump` | Fluids | speed, flow = product(speed, curve), head, power ∝ speed³ (curve), current | provides flow | impeller wear, strainer blockage, seal leak | start, stop, speed |
| `ahu` (exists) | Cooling | cooling load from occupancy schedule and outdoor temp | provides load; consumes chw supply temp | coil fouling, damper stuck | supply air setpoint |
| `heat-exchanger` | Cooling | outlet temps via effectiveness curve; approach KPI | consumes hot/cold inlet; provides outlets | fouling (ramp over days) | — |
| `air-compressor` | Air | `running`, load state (sequencer or trip), air delivered = product(running, loaded, capacity); power | provides air flow | airend wear (same kW, less air), fails to load, overheating trip | start, stop, load/unload |
| `air-receiver` | Air | header pressure = store fed by linked compressor flow minus demand | consumes air flow; provides pressure | drain valve stuck open (constant loss), leak | drain |
| `air-dryer` | Air | dew point lag; pressure drop | consumes air flow; provides dew point | desiccant exhaustion (ramp) | regenerate |
| `storage-tank` | Fluids | level store with inflow/outflow links | consumes inflow; provides level | leak, level sensor stuck | — |
| `control-valve` | Fluids | position lag to command; flow = product(position curve, pressure) | provides flow | sticking (hysteresis), positioner drift | position setpoint |
| `motor-drive` | Drives | VFD speed lag, current, winding temp regime, overload trip | consumes load torque; provides power | bearing wear (vibration), overload trip, winding overheat | start, stop, speed |
| `conveyor` | Drives | belt speed, load, motor current | consumes power | belt slip, jam (trip) | start, stop |
| `boiler` | Heat | steam pressure store, firing rate pid, efficiency curve | provides steam flow | scaling (efficiency ramp), flame failure trip | start, stop, pressure setpoint |
| `feeder-meter` | Electrical | power = follow sum of linked loads; 15-min demand via `peak`; energy integrator | consumes power | meter stale (freeze), CT ratio error (scale) | — |
| `transformer` | Electrical | winding temp lag of load²; derating regime above threshold | consumes power | cooling fan failure, overload | — |
| `generator` | Electrical | standby start (anti-recycle), fuel store, output | provides power | fails to start, fuel low | start, stop |

## Process loads and environment (4)

| Name | Core | Ports |
|---|---|---|
| `production-line` | rate (setpoint, ramp-limited), heat, air and power ∝ rate (`product`); cycle time and scrap curves of air pressure and coolant temp; shift calendar from a linked site calendar | consumes air pressure, coolant temp, shift; provides heat load, air demand, power |
| `packaging-line` | like production-line, air-heavy, heat-tolerant | same |
| `weather` (system) | dry bulb daily sine + seasonal; wet bulb = curve(dry bulb, humidity); heat-wave negative-class event | provides dry bulb, wet bulb |
| `site-calendar` (system) | weekly shift schedule with `utc_offset_s` parameter; tariff price schedule | provides shift (0..1), tariff |

## Instruments (13)

`rtd` (exists), `thermistor` (TT), `pressure-transmitter` (PT), `mag-flow-meter` (exists), `vortex-flow-meter` (FT),
`level-radar` (LT), `accelerometer` (VT), `current-clamp` (IT), `power-meter` (exists), `valve-position` (ZT), `humidity`
(MT), `run-status` (digital, boolean), `pulse-counter` (counter). Each: realistic resolution, noise (span-relative where the
datasheet is), cadence, availability; faults where physical: `stuck` (with `stuck_at` and a quality where the device knows),
`drift`, `dropout`, `noise` burst, `scale` (CT ratio, calibration).

## Evidence per component (harness, one table-driven test over the whole library)

For every component × preset: a one-entry composition composes, validates and lints with no errors; a 2 h run replays
byte-identically; every effector moves the states it names; every fault is observable above its first SNR by the truth
solver; the trivial-baseline audit runs. Every instrument attaches to a matching point of some component.
