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
| D1 | Where the library lives | Go repository, as data files with their own schema; built-in components in `internal/component/internal/files/builtin/`, customer components in directories outside the repository | One source for CLI, MCP, plant wire and UI; Go validation and tests apply; customer data never enters GitHub |
| D2 | What a component is | Two catalogs: **equipment** (the process: states, dynamics, actions, faults, measurement points, ports) and **instruments** (the sensor: value type, unit family, range, noise, drift, cadence, availability, instrument faults) | Factories differ in instrumentation more than in equipment; a user picks equipment, then accepts or changes its default instruments |
| D3 | How a component enters a world | **Copy-in**: composing writes concrete `kinds`, `instances` and `links` into an ordinary domain file, with a `provenance` record (component id, version, digest, preset, parameters) | Domains stay self-contained; digests, replay and validation are unchanged; a domain without the library still loads |
| D4 | How components are sized | **Presets plus bounded parameters**, substituted at declared JSON pointers; no arithmetic | The forbidden list excludes an expression language; presets (`small`, `medium`, `large`) carry consistent derived values, parameters cover the few free choices (setpoint, range, count) |
| D5 | How equipment connects | **Ports** on equipment (`chilled_water_out`, `load_in`, `power_in`…) typed by a small port vocabulary; a link joins compatible ports and becomes a domain `link`; coupling through links needs aggregation over typed links (R6e-01), built here | Without coupling a factory is a set of independent machines; with it, AHU load drives chiller load drives cooling-tower duty |
| D6 | Naming and taxonomy | Equipment classes follow ISO 14224 (rotating, static, electrical, instrumentation); instrument tags follow ISA-5.1 letters (`TT`, `PT`, `FT`, `LT`, `VT`, `IT`, `JT`); faults follow ISO 14224 failure modes where one exists; units are SI with a declared unit family | A recognised taxonomy makes the catalog predictable and searchable, and gives each fault a defensible name |
| D7 | Versions | `namespace/name@major`: `streamsim/chiller@1`; a breaking change is a new major; a domain records which version it was composed from | Pre-release now, but customer domains will outlive a library release |
| D8 | Truth | Component descriptions, fault catalogs and hidden states are director-side data; nothing new reaches an operator (consistent with R7e-09) | The library must not become a new leak |
| D9 | Evidence | Every component ships with tests that run it: nominal stability, determinism, each action's effect, each fault observable by a default instrument, numerical stability (dt/τ) and a trivial-baseline audit | An instrument is only useful if its parts are trustworthy (AGENTS.md) |

## 3. Shapes

The contracts and their rules are [COMPONENTS_V0_1](../design/COMPONENTS_V0_1.md) (phase P7g); the worked example is
[`examples/compositions/chiller-pilot.composition.json`](../../examples/compositions/chiller-pilot.composition.json) and the
domain it composes to, [`examples/chiller-pilot.domain.json`](../../examples/chiller-pilot.domain.json).

## 4. Architecture

- **Module `internal/component`** (core with an edge): domain layer (load, validate, presets, parameter substitution,
  instrument attachment, port compatibility, composition to domain), files edge through `kernel`, facade. It imports
  `domain` (for validation of the result) and `model`; nothing imports it except `cli`, `director` (catalog tools) and the
  bridge later.
- **Schemas** `component-v0.1`, `instrument-v0.1`, `composition-v0.1` under `contracts`, embedded copies under
  `internal/schemas`.
- **CLI**: `streamsim component list|show|validate`, `streamsim domain compose <composition> [--components-dir …]`.
- **MCP** (director role): `sim.component.list`, `sim.component.describe`, `sim.domain.compose`, so an agent can build a world
  from the library.
- **Domain spec v0.2 additions**: `provenance` on instance sets (structure only); aggregation of dynamics inputs over typed
  links (`sum`, `mean`), closing R6e-01 — the only change to world semantics, with an analytic oracle.
- **Customer libraries**: the built-in library is embedded in the binary; `--components-dir` (repeatable) and `STREAMSIM_COMPONENTS_DIRS` add customer directories after it; ids carry a namespace, so a
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

## 6. Review outcome (2026-10-13)

Two independent reviews: one of the direction (verdict: right direction, needs changes before P7g), one of missing pieces.
They change the design as follows.

| Finding | Change |
|---|---|
| Domain-wide name uniqueness breaks copy-in of one component twice (two presets of a chiller) | P7g defines a deterministic qualification rule: kind `chiller_large`, members `chiller_large.chw_supply_temp`; profiles' fault weights and adapters checked against it |
| Instrument faults such as stuck, open circuit and dropout are not expressible today (only additive bias through a hidden state) | P7i adds observation faults (stuck, freeze, dropout, scale error) as world semantics with a design change; until then an instrument lists only faults the engine can produce |
| Link coupling is a cross-entity read path (integration order, history, draw order, oracle), and `sum`/`mean` are too small | Coupling becomes its own design-first phase (P7h) before the module; it adds `sum`, `mean`, `min`, `max`, `count_running` and load sharing by capacity, a cycle ban over links, and a system-kind header store pattern |
| JSON pointers into arrays are brittle; presets cannot keep derived values consistent | Parameters address by name (`state:chw_supply.initial`); presets are full named overlays, each validated and run; a bound moved by a preset must keep its dependants inside it |
| Overrides are add-or-narrow, so per-instance variation forks kinds | One composition entry is one kind; variation inside it only through add or narrow |
| Ports did not say what a connection means | A port declares `provides` or `consumes` a state with a unit family; compose emits the link and the aggregation |
| No nonlinear relations: efficiency curves, pump and fan laws, valve Cv, heat = flow × ΔT | P7i adds a data-defined piecewise-linear `curve` (1-D and 2-D tables) and a bounded `product` of states — tables, not an expression language |
| No control layer | P7i adds a `controller` form (PID with output limits, anti-windup, auto/manual, cascade by chaining) |
| No alarm object | P7i adds an ISA-18.2 alarm (limit, on-delay, deadband, priority, acknowledge, shelve) and its KPIs |
| Units are prose | P7g adds a unit vocabulary with families and conversions; compose rejects a family mismatch |
| No demand side, weather or calendar | P7k adds process loads, a weather source (dry and wet bulb) and a site calendar (time zone, weekdays, shifts) |
| No plausibility checking for composed plants | P7j adds `domain lint`: a nominal day with energy and mass balance residuals and efficiency bands |
| Tag naming, hierarchy and quality codes | P7g adds ISA-95 site/area/unit and ISA-5.1 tags as instance metadata; quality codes on events (R7d-01) land with the instrument faults in P7i |
| Scale | P7l adds a 2,000-tag fixture and a speed budget |
| Prerequisites | R6a-01 and R7e-04 (stability validated by the validator), R7e-09 (descriptions to the operator) land before the catalog |
| Provenance and determinism | Provenance records component id, version and library digest; compose depends on no directory order or environment variable; duplicate ids across directories are an error; a gate keeps customer namespaces out of `internal/component/internal/files/builtin/` |
| Pilot first | The module phase composes one chiller with three instruments end to end before the catalog |

## 7. Plan (before P8a)

| Phase | Name | Deliverable | Exit evidence |
|---|---|---|---|
| P7g | Contracts and vocabularies | Equipment, instrument and composition schemas; naming rule; name-addressed parameters and presets; port, unit-family and unit vocabularies; tag and hierarchy metadata | One worked example (a chiller in two sizes, two instrument sets, one AHU link) composed by hand and accepted by the existing domain validator |
| P7h | Coupling over links | Cross-entity inputs over typed links: `sum`, `mean`, `min`, `max`, `count_running`, load sharing; cycle ban; header-store pattern | Analytic oracle for a header with parallel units; read-order independence; v0.1-equivalent worlds unchanged |
| P7i | World forms for plants | `curve` tables, bounded `product`, `controller` (PID), ISA-18.2 `alarm`, observation faults and quality codes | Oracles per form; prerequisites R6a-01, R7e-04, R7e-09 done |
| P7j | Component module, compose and lint | `internal/component`; `component list/show/validate`, `domain compose`, `domain lint`; MCP catalog tools; provenance; customer directories; one chiller pilot | Deterministic compose; the pilot composes, validates, lints and runs |
| P7k | Catalogs | About 12 instruments; about 15 utility components and process loads, a weather source, a site calendar; an incident set and KPI truth per component | Per component: nominal day in range and replayed, actions move their states, each fault observable above its SNR, stable by validation, trivial-baseline audit |
| P7l | Factory utilities world | `factory-utilities` composed only from the library; a 2,000-tag fixture and speed budget | Validates, lints, runs, verifies, reports; replay and audit; speed within budget |

## 8. Later (after the UI)

Reliability arrivals (MTBF or Weibull drawn into the command log, wear by run hours and starts, planned maintenance), work
orders and spare parts, SIS voting and bypass, electrical detail (reactive power, power factor, imbalance, transfer),
historian behaviour (compression, interpolation, backfill), protocol semantics (Modbus, OPC UA, BACnet), multi-site runs,
and calibration against real plant data.

## 9. Open questions for the owner

1. Copy-in (D3) confirmed by the direction review; an "update from library" action can be added later from provenance.
2. Which plant families come first after factory utilities (process lines, packaging, food and beverage, data center)?
3. Generic classes first (both reviews agree); vendor presets can live in customer libraries.
4. Scope before the UI: **all of P7g–P7l** (owner decision, 2026-10-13: building the UI first would double the work).
