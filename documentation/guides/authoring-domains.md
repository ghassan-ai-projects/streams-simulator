# Authoring a domain

> Status: Implemented data-authoring workflow. Authority: domain schema, loader, and validation tests. Verified by: domain validation smoke test. Last verified: 2026-11-03.

A domain is data, and there are two routes to one: **compose** a plant from the component library (the usual route for a plant, a fleet or a site), or **write** the JSON by hand (for a world the library cannot express). Both end in the same ordinary domain file that `domain validate`, `run`, the director and the canvas load, and neither needs a Go branch.

## Route 1: compose a plant from the component library

The library holds equipment (a chiller, a pump, an air-handling unit), instruments (an accelerometer, a flow meter) and vocabulary, all as data embedded in the binary (`internal/component/internal/files/builtin/`). A **composition** ([`composition-v0.1`](../../contracts/composition-v0.1.schema.json)) names the equipment you use (`streamsim/chiller@1` with a preset, parameters, instance `names` and an `area`), the instruments you add or drop, and the `connections` between ports; `domain compose` writes it as an ordinary domain, so a world never depends on the library at run time. Rules: [COMPONENTS_V0_1.md](../../docs/design/COMPONENTS_V0_1.md).

```bash
go run ./cmd/streamsim component list                  # equipment and instruments (--unit, --min, --max narrow it)
go run ./cmd/streamsim component show streamsim/chiller
go run ./cmd/streamsim component examples              # ready-made plants: chiller-pilot, factory-utilities, belt-conveyor-drives-a and -b
go run ./cmd/streamsim component new my-plant --title "My plant" --example chiller-pilot > my-plant.composition.json
go run ./cmd/streamsim domain compose my-plant.composition.json --out domains/my-plant.domain.json
go run ./cmd/streamsim domain lint domains/my-plant.domain.json
```

`component new` prints the composition as `{composition}` and writes nothing, so extract the `composition` member into your file; `--empty` copies only the frame (site, areas, axes, profiles). A composition that breaks a library rule fails `composition_invalid` with a pointer into it; a composed domain that breaks the domain contract fails `domain_invalid`. A customer library is a directory of `equipment/*.component.json` and `instruments/*.instrument.json` added with `--components-dir` (repeatable) or `STREAMSIM_COMPONENTS_DIRS`; the `streamsim/` namespace stays reserved for the built-in one, and `component validate` checks a library. Examples: [`examples/compositions/`](../../examples/compositions/) and their composed domains in [`examples/`](../../examples/). The canvas builds the same compositions visually and calls the same commands. An entry may also say which kind file its kind presents to a host (`implements`, the plant wire `KIND`) and, when the equipment's own channels and effectors are the raw values of a board, the board mapping that turns them into that kind file (`board_type`, the plant wire `BOARD` `<board_type>.board.json`; it needs `implements`). Both are copied onto the composed kind, and `domain compose`, `run`, `plantwire node` and `plantwire connect` then need `--kinds-dir` naming the directory that holds the kind files and board mappings: the belt conveyor plants keep theirs in [`examples/belt-conveyor/kinds/`](../../examples/belt-conveyor/kinds/) (each kind through its board mapping: `feeder-controller`, `belt-controller` and `drive-controller`; every start, stop and feed command has a writable metric of the same name to read it back). An example plant carries the kind files it implements: `component example` and `component new --example` list them in `kind_files`, and `component install-kinds <example> --kinds-dir <dir>` writes them into a kinds folder without overwriting a different file. An entry may name, per fault of its equipment, the right actions in this plant (`expected_actions`, keyed by the component's fault id, each `{kind, effector}` with the composed kind name and the component's effector name): compose writes them onto the composed fault with the effector qualified (`bulk_feeder__set_feed`), refuses a key the equipment declares no fault for (`/equipment/<i>/expected_actions/<fault>`), and the domain rules check the kind and effector. A label then expects them on the run's instances, so scoring credits, say, a feed cut on the feeder for a drive fault; the belt conveyor plants declare them for their drive faults.

`domain lint` runs a noiseless nominal day of the domain (`--hours`, `--seed`) and judges it with five plausibility rules: a non-finite value, a reading outside its range, a state pinned at its bound, a state that never settles, a conservation balance that does not close. An `error` finding exits 1; warnings exit 0. Run it on every composed plant: a domain can validate and still be implausible. It is a plausibility check, not a calibration.

Two more read-only commands help review: `domain assumptions <domain>` states what a domain assumes and does not model, and `domain diff <a> <b>` shows what changed between two domains or run folders and flags a changed model under an unchanged version. `domain lineage --domains-dir <dir>` shows how each plant is produced. See the [CLI reference](../reference/cli.md).

## Route 2: write a domain by hand

### Start from the contract

Read [`contracts/domain-spec-v0.2.schema.json`](../../contracts/domain-spec-v0.2.schema.json), the [kinds and instances semantics](../../docs/design/DOMAIN_SPEC_V0_2.md) and a committed file under [`domains/`](../../domains/). Identify the axes and runtime behavior the domain is meant to stress before adding fields.

A domain declares one or more **kinds** (the template: state, dynamics, channels, faults, effectors) and one or more **instance sets** (the id template, the count bounds and any churn), each following a kind; every kind needs at least one set. The kind name becomes the `entity_type` of every instance. Channel, fault and effector names must be unique across the whole domain, and a state, channel, fault or effector refers only to declarations of its own kind. Explicit entity ids given to a run (a scenario's `world.entity_ids`, the director's world creation) must fit the id template of an instance set and its count bounds.

The design catalog is a planning document, not proof that a domain is shipped. A domain becomes shipped only when it is intentionally committed, loaded by the catalog, covered by schema/conformance tests, and included in the release evidence.

### Instances, overrides, links and boards

An instance set may list `names` (`M1`) next to or instead of an `id_template`. `override` on a set adds states, channels, faults and effectors or narrows a state bound, a channel range or an effector argument bound; it never removes or retypes. A plant-wide measurement belongs to a kind with `system: true` and one named instance. `link_types` and `links` give the typed topology between instances; a kind reads the instances linked into it through a linked input or `follow` (`sum`, `mean`, `min`, `max`, `count_running`; section 21, exercised by `internal/testsupport/testdata/coupled-plant.domain.json`), and `boards` group the instances hosted on one device (the board owns the `boot_id`). Every rule and its pointer is in [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) sections 8 to 12; `internal/testsupport/testdata/structured-plant.domain.json` exercises all of them.

### Stores and flows

A channel that reports a code may declare `states`, a list of `{value, name, attention}` (`attention` is `alarm`, `warn` or `null`; `value` is unique in the channel), so the domain says what each code means and a consumer decodes it without a table of its own. The world ignores it. An instrument declares the same list and `domain compose` copies it onto the channel the instrument produces (the built-in annunciator declares the five alarm states). Rules: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 30.

A level that accumulates (a tank, a pressure header) is a state with `f1.form: "store"` and a `capacity`. `flows` on the kind move `conductance × Π scale × (from − to)` per second between two stores, or drain one store toward zero; `scale` names actuator and health states such as a pump speed or a pipe health a fault lowers. Model faults on a flow through a health state in `scale`. Every store of a kind shares one `dt_ms`. Rules and pointers: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 13; `internal/testsupport/testdata/transfer-plant.domain.json` is a worked example.

### Schedules, regimes and demand peaks

An exogenous profile (a tariff, a daily load shape) is an `f0.schedule` of points, repeating with `period_s` and aligned to the clock. A behaviour that changes past a threshold (transformer oil, tower capacity against wet bulb) is `f1.form: "regime"` with a `selector`, a `threshold` and a `below` and `above` lag. A demand charge is `f1.form: "peak"` over one input with `block_s`, `period_s` and an optional `ratchet`. Rules: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 16; `internal/testsupport/testdata/utility-plant.domain.json` uses all three.

### Trips, lockouts and preconditions

A protection trip is `f1.form: "trip"` with a `when` condition (`lt`, `lte`, `gt`, `gte` against a threshold) held for `for_s`; `latch: true` makes it a lockout that a change of `reset_on` clears. `anti_recycle` is a timer started by an edge of a watched state. `boundaries` mark trip states as safety boundaries with a severity. An effector's `requires` lists the conditions the device checks before acting (not locked out, timer elapsed, not isolated, not in hand mode). Rules: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 17; the chiller in `internal/testsupport/testdata/utility-plant.domain.json`.

### Supervisory control

Command the plant at supervisory level: an effector assigns a setpoint state, and conditions compare against it with `threshold_state`. A `sequencer` form stages equipment up and down within a band. An effector `responder` models a person or crew whose delayed effect depends on whether the call named the active fault; an effector `lease` falls back to a safe state when not renewed. Rules: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 18; the air system in `internal/testsupport/testdata/utility-plant.domain.json`.

### Short events

A disturbance shorter than a step (a voltage sag) is a fault with `onset.shape: "pulse"` and `duration_s`. Trips see it on the event clock, and an `event_driven` channel records its start and end. Rules: [DOMAIN_SPEC_V0_2.md](../../docs/design/DOMAIN_SPEC_V0_2.md) section 19, which also lists every random substream and when it draws.

### Outcome cost and baseline policies

`ground_truth.outcome_weights` prices what happened in a run, in the domain's own unit: `boundaries` (`{kind, name, per_violation, per_hour}`), `trips` (`{kind, state, per_trip}` for a trip state), `kpis` (`{kind, state, per_unit}` for one of the kind's `kpis`; a negative weight rewards more) and `calls` (`{effector, per_call}`, an unneeded call). `ground_truth.baseline_policies` are fixed policies a run is compared with: `{name, when: {kind, state, operator, threshold}, for_s, then: [{effector, target, args, delay_s}]}`, where `target` is `matched` (the instance whose condition fired) or a declared instance name, and `args` leaves out the call's entity argument (a required string argument named `*_id`), which is the target. `outcome_weights.tie` (default 0) is the cost difference within which two runs tie, and the optional `outcome_weights.unit` names the unit J is counted in (the conveyor plants: `equivalent production hours`); a policy's optional `label` says what it does in a reader's words (`vibration alarm: feed 0, then stop the drive`). The report shows both; neither changes a cost or a verdict. Every kind, boundary, state, KPI, effector, instance and argument named is checked when the domain loads; `streamsim outcome` scores a run with them. The conveyor plants `examples/belt-conveyor-drives-{a,b}` carry both.

## Validate locally

```bash
go run ./cmd/streamsim domain validate domains/my-domain.domain.json
go run ./cmd/streamsim domain format domains/my-domain.domain.json
go run ./cmd/streamsim domain lint domains/my-domain.domain.json
go run ./cmd/streamsim catalog describe my-domain
go test ./internal/domain/... ./internal/schemas
```

`domain format` rewrites the file in the canonical layout (stable member order, line-fitting objects), so a hand edit and a tool edit of the same world produce the same bytes; run it before committing. It never changes the digest. `domain format` prints `{formatted, changed, digest}`. A validation failure reports every issue with a JSON pointer into the document.

Use the exact output digest in review notes. If the domain changes after a run, that run must be treated as a different input set.

## Design checks

- Every entity and channel has a reason to exist in the scenario.
- State dynamics have a boundary case and an analytic or metamorphic test where applicable.
- Faults and effectors are declared, parameterized, and scoped to valid entities.
- Absence semantics are explicit: signal, normal, ambiguous, or heartbeat.
- A scenario is not trivially solvable by a one-line detector.
- Free text cannot become a hidden control channel.
- The domain does not encode knowledge of a particular consumer.

## Completion evidence

Add or update fixtures and tests under the existing package boundaries. Update the catalog/coverage evidence, release inventory, and public limitations if the domain is partial. Do not call a new domain shipped solely because the CLI can load it from a dirty working tree.

## Next reads

- [Domains and adapters](../architecture/domains-and-adapters.md)
- [Benchmark methodology](../benchmark/methodology.md)
- [Contracts](../reference/contracts.md)
