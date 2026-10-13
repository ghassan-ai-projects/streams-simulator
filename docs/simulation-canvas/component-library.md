# Component library — design and plan

- **Date**: 2026-10-13
- **Decision**: the owner wants a catalog of predefined equipment and sensors before the canvas UI, so a person (or an
  agent) can model a factory without knowing every sensor, range, noise model or fault.
- **Placement**: Go side, as data, served to the UI through the bridge (decided 2026-10-13). The frontend owns presentation
  only (palette, icons, forms, layout).

## 1. The problem

A domain today is written by hand: every kind declares its states, dynamics, channels (with units, ranges, noise, cadence,
drift, availability), effectors (with arguments, interlocks, leases, acknowledgement modes) and faults. That is precise but
expert work. To model a factory a user should pick "three chillers, two cooling towers, a compressed-air system", size them,
choose how they are instrumented, connect them, and get a valid, realistic, deterministic world.

## 2. Decisions

| # | Question | Decision | Why |
|---|---|---|---|
| D1 | Where the library lives | Go repository, as data files with their own schema; built-in components in `components/`, customer components in directories outside the repository | One source for CLI, MCP, plant wire and UI; Go validation and tests apply; customer data never enters GitHub |
| D2 | What a component is | Two catalogs: **equipment** (the process: states, dynamics, actions, faults, measurement points, ports) and **instruments** (the sensor: value type, unit family, range, noise, drift, cadence, availability, instrument faults) | Factories differ in instrumentation more than in equipment; a user picks equipment, then accepts or changes its default instruments |
| D3 | How a component enters a world | **Copy-in**: composing writes concrete `kinds`, `instances` and `links` into an ordinary domain file, with a `provenance` record (component id, version, digest, preset, parameters) | Domains stay self-contained; digests, replay and validation are unchanged; a domain without the library still loads |
| D4 | How components are sized | **Presets plus bounded parameters**, substituted at declared JSON pointers; no arithmetic | The forbidden list excludes an expression language; presets (`small`, `medium`, `large`) carry consistent derived values, parameters cover the few free choices (setpoint, range, count) |
| D5 | How equipment connects | **Ports** on equipment (`chilled_water_out`, `load_in`, `power_in`…) typed by a small port vocabulary; a link joins compatible ports and becomes a domain `link`; coupling through links needs aggregation over typed links (R6e-01), built here | Without coupling a factory is a set of independent machines; with it, AHU load drives chiller load drives cooling-tower duty |
| D6 | Naming and taxonomy | Equipment classes follow ISO 14224 (rotating, static, electrical, instrumentation); instrument tags follow ISA-5.1 letters (`TT`, `PT`, `FT`, `LT`, `VT`, `IT`, `JT`); faults follow ISO 14224 failure modes where one exists; units are SI with a declared unit family | A recognised taxonomy makes the catalog predictable and searchable, and gives each fault a defensible name |
| D7 | Versions | `namespace/name@major`: `streamsim/chiller@1`; a breaking change is a new major; a domain records which version it was composed from | Pre-release now, but customer domains will outlive a library release |
| D8 | Truth | Component descriptions, fault catalogs and hidden states are director-side data; nothing new reaches an operator (consistent with R7e-09) | The library must not become a new leak |
| D9 | Evidence | Every component ships with tests that run it: nominal stability, determinism, each action's effect, each fault observable by a default instrument, numerical stability (dt/τ) and a trivial-baseline audit | An instrument is only useful if its parts are trustworthy (AGENTS.md) |

## 3. Shapes

### 3.1 Equipment (`component-v0.1`, `components/equipment/*.component.json`)

```json
{
  "id": "streamsim/chiller", "version": 1, "class": "static/heat_exchanger/chiller",
  "title": "Water-cooled chiller", "tags": ["hvac", "utilities"], "icon": "chiller",
  "kind": { "name": "chiller", "state": [...], "dynamics": [...], "effectors": [...], "faults": [...] },
  "measurement_points": [
    {"name": "chw_supply_temp", "state": "chw_supply", "unit_family": "temperature", "default_instrument": "streamsim/rtd@1", "tag": "TT"}
  ],
  "ports": [
    {"name": "load_in", "type": "thermal_load", "direction": "in"},
    {"name": "condenser_heat_out", "type": "thermal_load", "direction": "out"},
    {"name": "power_in", "type": "electrical", "direction": "in"}
  ],
  "presets": {"small": {"/kind/state/0/max": 350, "...": 0}, "large": {"...": 0}},
  "parameters": [{"name": "setpoint_c", "pointer": "/kind/state/2/initial", "min": 4, "max": 12, "default": 7}]
}
```

- `kind` is an ordinary domain-spec v0.2 kind **without channels**: channels come from instruments attached at measurement
  points.
- `presets` and `parameters` only replace values at JSON pointers inside the component; the result must validate as a kind.

### 3.2 Instrument (`instrument-v0.1`, `components/instruments/*.instrument.json`)

```json
{
  "id": "streamsim/rtd", "version": 1, "tag": "TT", "unit_family": "temperature", "unit": "degC",
  "channel": {"value_type": "number", "resolution": 0.1, "cadence": {"mode": "periodic", "period_s": 10},
              "noise": {"model": "gaussian", "sigma": 0.05}, "drift": {"model": "linear", "rate_per_hour": 0.001}},
  "faults": ["stuck_at", "drift", "open_circuit", "noise_burst"]
}
```

Attaching an instrument at a measurement point produces a channel that observes the point's state; the instrument's faults
become kind faults on that channel's state path (the existing observation offset and bias forms).

### 3.3 Composition (`composition-v0.1`)

The input of `streamsim domain compose` and of the builder:

```json
{
  "id": "factory-utilities", "title": "Factory utilities",
  "equipment": [
    {"use": "streamsim/chiller@1", "names": ["CH-1", "CH-2", "CH-3"], "preset": "large", "parameters": {"setpoint_c": 6.5},
     "instruments": {"remove": ["condenser_flow"], "replace": {"chw_supply_temp": "streamsim/thermistor@1"}}}
  ],
  "connections": [{"from": "AHU-1.chilled_water_in", "to": "CH-1.load_in"}],
  "profiles": [...]
}
```

`compose` is deterministic: the same composition and library give the same domain bytes (formatted with `domain format`),
with `provenance` on each instance set.

## 4. Architecture

- **Module `internal/component`** (core with an edge): domain layer (load, validate, presets, parameter substitution,
  instrument attachment, port compatibility, composition to domain), files edge through `kernel`, facade. It imports
  `domain` (for validation of the result) and `model`; nothing imports it except `cli`, `director` (catalog tools) and the
  bridge later.
- **Schemas** `component-v0.1`, `instrument-v0.1`, `composition-v0.1` under `docs/contracts`, embedded copies under
  `internal/schemas`.
- **CLI**: `streamsim component list|show|validate`, `streamsim domain compose <composition> [--components-dir …]`.
- **MCP** (director role): `sim.component.list`, `sim.component.describe`, `sim.domain.compose`, so an agent can build a world
  from the library.
- **Domain spec v0.2 additions**: `provenance` on instance sets (structure only); aggregation of dynamics inputs over typed
  links (`sum`, `mean`), closing R6e-01 — the only change to world semantics, with an analytic oracle.
- **Customer libraries**: `--components-dir` (repeatable) and `STREAMSIM_COMPONENTS_DIRS`; ids carry a namespace, so a
  customer's `acme/...` never collides with `streamsim/...`; nothing customer-specific is committed.

## 5. First catalog (factory utilities)

| Group | Equipment | Notes |
|---|---|---|
| Cooling | chiller, cooling tower, air-handling unit (AHU), chilled-water pump | The coupled loop: AHU load → chiller → cooling tower |
| Heat | boiler, heat exchanger | Steam or hot water; efficiency drift, scaling |
| Air | air compressor, receiver tank, dryer | Header pressure, leaks as a fault |
| Fluids | centrifugal pump, control valve (with position feedback), storage tank | Stores and flows (P6a) |
| Drives | VFD + induction motor, conveyor | Speed feedback, bearing wear, overload trip |
| Electrical | feeder with energy meter | Demand and peak (P6b) |

Instruments: RTD and thermistor (`TT`), pressure transmitter (`PT`), magnetic and vortex flow meter (`FT`), level radar
(`LT`), accelerometer (`VT`), current clamp (`IT`), power meter (`JT`), valve position (`ZT`), humidity (`MT`), run status
(digital), counter. Each with stuck, drift, dropout and noise-burst faults where physical.

## 6. Plan (before P8a)

| Phase | Name | Deliverable | Exit evidence |
|---|---|---|---|
| P7g | Component library: contracts | The three schemas, this design reviewed, the port and unit-family vocabularies, ISO 14224 / ISA-5.1 mapping table | Schemas validate their fixtures; a reviewed design record |
| P7h | Component module and compose | `internal/component`, CLI and MCP commands, provenance, customer directories | Compose is deterministic (byte-identical twice); composed domains validate; gates and coverage |
| P7i | Coupling over links | Aggregation of inputs over typed links (R6e-01) with an analytic oracle | A two-unit loop agrees with its closed form; v0.1-equivalent worlds unchanged |
| P7j | Instrument catalog | About 12 instruments with their faults | Each instrument attached to a test kind: noise, drift and each fault visible in its channel |
| P7k | Equipment catalog | About 15 equipment components with presets, measurement points and ports | Per component: 24 h nominal run stays in range and replays; each action moves its state; each fault observable by a default instrument; dt/τ stable; trivial-baseline audit passes |
| P7l | Factory utilities world | `factory-utilities` composed only from the library (the P12 factory, moved earlier) | Validates, runs, verifies, reports; replay and audit; the composition and the domain are both committed |

The UI phases then build on it: the builder (P11) edits compositions and domains with the catalog as its palette, and the
bridge (P8b) serves the library.

## 7. Open questions for the owner

1. Is copy-in (D3) right, or should a domain keep referencing its components so a library update can flow into existing
   worlds? Copy-in is simpler and keeps worlds stable; an "update from library" action can be added later from provenance.
2. Which plant families come first after factory utilities (process lines, packaging, food and beverage, data center)?
3. Should the library hold real vendor equipment models (named products) or stay generic classes? Generic first is
   recommended; vendor-specific presets can live in customer libraries.
