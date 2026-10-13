# Components v0.1: equipment, instruments and compositions

- **Status**: contracts of phase P7g of the simulation-canvas goal ([PLAN](../../governance/CANVAS_PLAN.md),
  [design and reviews](../simulation-canvas/component-library.md)). `internal/component` loads and composes them
  (P7j): `streamsim component list|show|validate` and `streamsim domain compose`.
- **Contracts**: [`component-v0.1`](../../contracts/component-v0.1.schema.json),
  [`instrument-v0.1`](../../contracts/instrument-v0.1.schema.json),
  [`composition-v0.1`](../../contracts/composition-v0.1.schema.json),
  [`component-vocabulary-v0.1`](../../contracts/component-vocabulary-v0.1.schema.json).
- **Data**: built-in files under [`components/`](../../components/); customer libraries in directories outside the repository.
- **Worked example**: [`components/compositions/chiller-pilot.composition.json`](../../components/compositions/chiller-pilot.composition.json)
  composes to [`examples/chiller-pilot.domain.json`](../../examples/chiller-pilot.domain.json) by the rules below.

A **component** is reusable equipment: the process it models (a domain-spec v0.2 kind without channels), the points where it can
be measured, the ports it connects through, and the sizes and settings a user may choose. An **instrument** is a reusable sensor
that turns a measurement point into a channel. A **composition** names equipment, sizes, instruments and connections; composing it
writes an ordinary domain (copy-in), so a world never depends on the library at run time.

## 1. Identity, versions, libraries

- An id is `namespace/name`: `streamsim/chiller`. `namespace` and `name` match `^[a-z][a-z0-9-]{1,31}$`. The `streamsim`
  namespace is reserved for the built-in library: compose refuses a `streamsim/` file found in any other directory.
- `version` is a major number. A change that can alter a composed domain is a new major; a reference is `streamsim/chiller@1`.
- A library is a directory holding `equipment/*.component.json`, `instruments/*.instrument.json` and optionally
  `vocabulary.json`. Compose reads the built-in library and then each `--components-dir` in command-line order. The pair
  (id, version) is the key: the same pair in two places is an error, whatever the content. The first directory is the built-in
  library; `--components-dir` is repeatable (default `components`) and `STREAMSIM_COMPONENTS_DIRS` is appended after the flags.
  Files are read in sorted name order; a missing `equipment/` or `instruments/` folder is an empty one, a missing directory is an error.

## 2. Equipment (`component-v0.1`)

| Member | Meaning |
|---|---|
| `id`, `version`, `title`, `description` | Identity; descriptions are library text and are never copied into a domain |
| `class` | Equipment class path, ISO 14224 where one exists: `static/heat_exchanger/chiller`; first segment open |
| `tags`, `icon` | Search and display hints |
| `kind` | A domain-spec v0.2 kind without `name` and `channels`: `state`, `dynamics`, `flows`, `system`, `faults`, `effectors`, `boundaries` |
| `measurement_points[]` | `{name, state, unit_family, range, tag, default_instrument, optional?}` |
| `ports[]` | `{name, type, direction: provides|consumes, state}` (section 5) |
| `presets`, `default_preset` | Complete named variants; `default_preset` is required when `presets` exists |
| `parameters[]` | `{name, title, unit?, type: number|integer|enum, min?, max?, values?, cases?, default, targets?}` |

**Names and lengths.** Component, point, port, fault and effector names match `^[a-z][a-z0-9]*(_[a-z0-9]+)*$` (no double
underscore, which compose uses to qualify). Limits that keep every qualified name inside the domain's 64 characters: a kind name
(the entry's `as` or the component name) ≤ 20, a point ≤ 20, an instrument fault id ≤ 16, an equipment fault or effector ≤ 40, a
port type ≤ 20.

**Targets** address one field: `<section>/<name>/<path>`. `section` is `state`, `dynamics` (named by its `target` state),
`flows` (named by its `name`), `effectors`, `faults` or `measurement_points`. `path` walks the entry's members; inside an array
whose items have a key, a segment selects the item by that key (`inputs/<state>`, `affects/<state>`, `state_deltas/<state>`,
`failure_modes/<mode>`); only an array of unkeyed items (a schedule's `points`, `seasonality`) is addressed by position.
Examples: `state/setpoint/initial`, `dynamics/chw_supply/f1/inputs/load/coef`, `faults/load_swing/affects/load/delta`,
`measurement_points/power/range/max`.

**Presets and parameters.** A composition entry without `preset` uses `default_preset`. The preset's values are applied first,
then each parameter (the entry's value or the parameter's default): a `number`/`integer` parameter writes its value to each of
its `targets`; an `enum` parameter writes, for the chosen value, the target values listed in `cases[value]`. A target named both by
the preset and by a parameter is a load error. Every preset of a component must yield a valid kind; there is no arithmetic, so a
preset writes every size-dependent value (capacities, coefficients, ranges, fault deltas) itself.

**Load-time checks** (P7j refuses the library): a target that resolves to nothing; a preset or parameter naming an unknown
target; a parameter default outside `min`/`max`/`values`; `cases` keys that differ from `values`; a measurement point or port whose
state is undeclared; a point `unit_family`, a port `type` or a `tag` absent from the merged vocabulary; a `default_instrument`
that resolves to no instrument or to one of another family or tag; a fault detector, observation or confirmation channel that
names no measurement point, or an `expected_effector` that names no effector; a name that breaks the length or pattern rules.
Also refused: a state whose unit is not in the point's (or the port type's) unit family, a tag that does not measure the point's
family, an inverted range, a name declared twice, a target value of another JSON type than the member it replaces, a target
that is an object or a list, a kind section holding something other than objects. That every preset yields a valid kind is not
checked at load: compose validates the composed domain.

## 3. Instruments (`instrument-v0.1`)

| Member | Meaning |
|---|---|
| `id`, `version`, `title`, `description` | Identity; descriptions are never copied into a domain |
| `tag` | ISA-5.1 letters; must equal the tag of the point it is attached to |
| `unit_family`, `unit` | What it measures and the unit it reports; the unit belongs to the family |
| `range` | The instrument's own range, in its unit (optional) |
| `scaling` | `absolute` (default) or `span`: with `span`, `channel.noise.sigma` and fault magnitudes are fractions of the point's span |
| `channel` | `value_type`, `resolution`, `fidelity`, `absence`, `cadence`, `noise`, `drift`, `availability`, `link_delay` (domain-spec v0.2 channel members; they may name no state) |
| `faults[]` | `{id, form, magnitude, duration_s?, is_negative_class?}` |

An instrument's `class` in a listing is its unit family and its `tags` the one tag letter.

**Attaching** an instrument at a measurement point writes one channel:

| Channel member | Value |
|---|---|
| `name` | `<kind>.<point>` |
| `observes` | the point's `state` |
| `unit` | the instrument's unit |
| `observation_gain`, `observation_offset` | the conversion from the state's unit to the instrument's unit, both in the point's family: with `base = s·x + o` for the state unit and `base = s'·y + o'` for the instrument unit, gain `s/s'` and offset `(o − o')/s'`; omitted when the units are equal |
| `range` | the point's range (in the state's unit) converted to the instrument's unit, intersected with the instrument's range |
| other members | the instrument's `channel` members; `noise.sigma` multiplied by the converted span when `scaling` is `span` |

**Instrument faults** (forms the world produces today, through a hidden bias state):

| Form | Domain fault |
|---|---|
| all | a state `<point>__bias` (instrument unit, initial 0, no dynamics) is added once per attached point whose instrument has an `offset` or `drift` fault; the channel gets `observation_bias {state: <point>__bias, coef: 1, noise_scale_at_full: 1}`; the fault id is `<kind>__<point>__<id>`; its detector is `single_channel_snr` on the channel with `first_observable_snr` 3 and `unavoidable_snr` 10; `is_negative_class` is copied |
| `offset` | `affects [{state: bias, delta: magnitude}]`, onset `step`, magnitude 1 |
| `drift` | `magnitude` is per hour and `duration_s` (required) is the time to full drift: `delta = magnitude · duration_s / 3600`, onset `ramp` with `rate_per_hour = 3600 / duration_s`, magnitude 1 |
| `stuck`, `freeze`, `dropout`, `scale`, `noise` | no bias state: the domain fault carries `observation: {channel: <kind>.<point>, form, value?, factor?, quality?}` (domain-spec v0.2 §25), onset `step`, magnitude 1; `stuck` sticks at the `low` or `high` end of the channel's converted range (`stuck_at`); `scale` and `noise` copy `factor`; `quality` is copied |

Offset and drift magnitudes are in the instrument's unit (fractions of the span with `scaling: span`).

## 4. Compositions (`composition-v0.1`)

| Member | Meaning |
|---|---|
| `id`, `version`, `title`, `description`, `stresses`, `axes`, `profiles`, `ground_truth` | Copied into the domain; profile fault weights use composed (qualified) fault ids |
| `site` | `{name, timezone}`, copied to the domain's `site` (structure only until the site calendar of P7k) |
| `areas[]` | `{name, parent?}`: the ISA-95 hierarchy, copied to `site.areas`; a parent must be declared earlier |
| `equipment[]` | `{use, as?, preset?, parameters?, names[] and/or (id_template, count), area?, instruments?}`; `area` must be declared in `areas` |
| `connections[]` | `{from: "<instance id>.<port>", to: "<instance id>.<port>"}`; an endpoint splits at its last dot |

`names`, `id_template` and `count` follow domain-spec v0.2 instance sets (§2.2, §8). `instruments` changes an entry's default
instrumentation: `replace` (point to reference), `remove` (points) and `add` (optional points). Removing or replacing an absent
point (an optional point not added is absent), adding a non-optional one, or leaving unattached a point a fault's detector or
observation or an effector's confirmation needs, is a compose error. A replacement must keep the point's tag and unit family.
One entry is one kind: instances of an entry are identical apart from their ids; a variation needs another entry.

## 5. Ports and connections

A port **provides** a state (read by what it connects to) or **consumes** into a state (which receives what connected ports
provide); its `type` fixes the unit family. A connection joins a `provides` port to a `consumes` port of the same type; anything
else is a compose error. Compose writes one `link_type` per port type and kind pair, named `<type>__<from kind>__<to kind>`, and one
`link` per connection. An endpoint names an instance of the composition (a name, or an id that fits an id template) and splits
at its last dot.

Links carry values through domain-spec v0.2 §21 (P7h): compose writes a consuming port's state as `follow` (aggregate `sum`)
over the port's link type, so an instance with incoming links takes the sum of the providing states, and an instance without
links keeps its own dynamics; connected and unconnected instances of one entry stay one kind. One consuming state of a kind takes
one link type and one providing state: connections that feed it otherwise are a compose error, as is a consuming state with no
dynamics entry to carry the `follow`. A component reads other
aggregates (`mean`, `min`, `max`, `count_running`) through linked inputs in its own dynamics. Compose rejects a composition
whose connections make kinds read each other in a cycle.

## 6. Vocabulary (`vocabulary.json`)

- **Unit families** with a base unit and conversions `base = scale × value + offset`; the base unit is in `units` with scale 1
  and offset 0.
- **Port types**, each with a unit family that exists. **Tag letters** (ISA-5.1), each with the unit families it accepts.
- **Merging**: compose merges the built-in vocabulary with every library's, entry by entry: a library may add a unit to a family,
  a family, a port type, a tag letter or a family to a tag's `measures`. Defining an existing unit, port type or base differently
  is an error. After each merge a family's base must be a unit with scale 1 and no offset, and every port type and tag must name
  a family that exists, so a library may add a unit to a built-in family without repeating the other units.

## 7. Names and metadata in a composed domain

- The kind is the entry's `as`, or the component's name with hyphens as underscores; two entries with one kind name are an error.
- A channel is `<kind>.<point>`; an equipment fault `<kind>__<id>`; an effector `<kind>__<name>`; an instrument fault
  `<kind>__<point>__<id>`; a bias state `<point>__bias` (states are scoped to their kind).
- Every reference compose copies is qualified: fault detectors (`channel`, `channel_a`, `channel_b`), `expected_effector`,
  `confirmation_channels`, and boundary and dynamics members that name channels.
- A channel carries its point's ISA-5.1 letters in `tag`; an instance set carries `area` and `provenance` (domain-spec v0.2 §20):
  component, version, the component file's digest, the preset used (the default included), every parameter value (defaults
  included), the instrument attached at each point (sorted by point) and each instrument file's digest. Digests are RFC 8785
  canonical JSON hashed with SHA-256.
- Points are attached in the component's declaration order; compose writes kinds in composition order and formats the domain with
  `domain format`.

## 8. Custom components without code

Every part of the library is data, so a user adds equipment, instruments and vocabulary without changing the binary:

- a custom equipment or instrument is a file that follows these schemas, in the user's own namespace (`acme/kiln`), in a
  directory outside this repository;
- a custom library may carry its own `vocabulary.json` that adds units, unit families, port types and tag letters (section 6);
- a custom equipment uses any world form the engine has. Only a new kind of physics (a new dynamics form) needs code, and that is
  a design change of the domain spec, not of the library.

## 9. Truth and determinism

- Library descriptions are never copied into a domain's state, channel, fault or effector descriptions: they name failure modes a
  consumer must not read. A composed effector carries no description.
- The same composition with the same library files gives the same domain bytes: no directory listing order, environment variable
  or clock enters compose. A gate keeps every id under `components/` in the `streamsim` namespace.
