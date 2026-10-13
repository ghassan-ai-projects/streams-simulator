# Domain spec v0.2, semantics part 1: kinds and instances

- **Status**: implemented in phase P5a of the simulation-canvas goal ([PLAN](../../governance/CANVAS_PLAN.md)); amended in phase P5a2
  (several kinds and instance sets per domain, sections 2.3 and 3); continued in phase P5b (sections 8 to 12).
- **Contract**: [`contracts/domain-spec-v0.2.schema.json`](../../contracts/domain-spec-v0.2.schema.json), embedded in the binary
  (`internal/schemas`), parity-tested against the committed copy. v0.1 is removed: the project is pre-release and the eight shipped
  domains are rewritten in the same change.
- **Later parts**: `store` and `flow` are section 13 (P6a), `schedule`, `regime` and `peak` section 16 (P6b), protection section 17 (P6c), supervisory control section 18 (P6d), the event clock section 19 (P6e). Aggregation of dynamics inputs over links (`sum`, `mean`) is not built yet (finding R6e-01).

## 1. Why

v0.1 described one anonymous population: a top-level `entities` block next to `state`, `dynamics`, `channels`, `faults` and
`effectors` that all described "the entity". The canvas, the plant wire and the factory reference world need the two ideas
separated ([canvas-design §4](../simulation-canvas/canvas-design.md), [plant-wire §14](../simulation-canvas/plant-wire-protocol.md)):

- a **kind** is the template: what state an entity has, how it evolves, what it exposes (channels), how it can fail (faults)
  and what can be done to it (effectors);
- an **instance** is a named entity that follows a kind.

## 2. Document shape

```
id, version, title, description?, axes, stresses     unchanged from v0.1
kinds[]        one or more kinds
instances[]    one or more instance sets, each naming a kind; every kind is named by at least one set
profiles[], ground_truth                             unchanged from v0.1
```

### 2.1 Kind

`kinds[0]` = `{name, state?, dynamics?, channels, faults, effectors?}`. The five sections are the v0.1 definitions, member for
member. `name` matches `^[a-z][a-z0-9_-]{0,63}$` and becomes the `entity_type` of every instance (v0.1's optional
`entities.entity_type`, now mandatory and named once).

Every instance of a kind has every state, channel, fault and effector of the kind. Part 1 has no override, so every instance of a
kind is identical apart from its id.

### 2.2 Instances

`instances[i]` = `{kind, id_template, count{default,min?,max?}, churn?}`. These are v0.1's `entities` members, with the kind named
explicitly. `kind` must name a declared kind (an issue at `/instances/i/kind` otherwise) and every kind must be named by at least one
instance set (an issue at `/kinds/k/name` otherwise).

An instance set is a *templated population*: `id_template` renders ids (`{n}` is the 1-based index, any other `{name}` placeholder
renders `a`), `count.default` instances exist at the start, and `churn` births and retires instances during a run. Each set has its
own template, count and churn; a kind may have several sets, and a domain several kinds.

**Per-instance parameters.** v0.1 allowed no variation between entities. Variation is the add-or-narrow override of section 9,
and named instances (`M1`, `AC1`) are section 8 (both P5b).

### 2.3 Several kinds (P5a2)

A domain holds several kinds, each with its own state, dynamics, channels, faults and effectors. A world entity belongs to one kind
(the kind of the instance set whose template its id fits), and that kind selects its state, dynamics, channels, faults and effectors.

**Names.** Decision: names are **unique across the whole domain**, not qualified by kind. This keeps every name that crosses a boundary
(an event's `channel`, a fault id in a label, an effector name in a command, the keys of a profile's `fault_weights`, an adapter
mapping) a plain string that resolves to one declaration, so the adapter, perturbation, scoring and MCP surfaces are unchanged.

| Name | Scope | Violation |
|---|---|---|
| kind `name` | domain | duplicate issue at `/kinds/k/name` |
| channel `name`, fault `id`, effector `name` | domain (across all kinds) | duplicate issue at the second declaration |
| state `name` | its kind (two kinds may each declare `level`) | duplicate issue within a kind |
| profile `name`, profile `fault_weights` keys | domain; the keys are domain-wide fault ids | undeclared fault |

**References resolve inside the kind.** A channel `observes` and `observation_bias.state`, a dynamics `target` and inputs, a fault
`affects`, an effector `state_deltas` and interlock, a detector channel, a fault's `expected_effector` and an effector's
`confirmation_channels` name declarations of the **same kind**. A reference to another kind's declaration is an undeclared-name issue
at the same pointer as before. A fault or effector acts only on entities of its own kind: injecting a fault or invoking an effector
on an entity of another kind is refused by the world.

**Per-kind behaviour.** An entity emits only its kind's channels. A `peer_residual` detector pools the entity's peers: the live
entities **of the same kind**. A suite scenario draws its entity from the entities of the injected fault's kind, applies profile setup
calls only when the entity's kind declares the effector, and audits against the channels and the negative-class control fault of that
kind. A profile's `entity_count` replaces the default count of the **first** instance set only.

**Explicit entity ids** (section 5) are matched to an instance set by template; an id that fits no set is refused. The checks of
count bounds apply per set to the ids that fit it.

**RNG draw order.** The world draws only from named substreams (`<entity>/<channel>/noise`, `fault/<id>/<entity>/<onset>/<state>`,
`<entity>/<effector>/fault_shape`, `churn/...`), so the draws of one entity never depend on another entity or kind. The rules that
keep single-kind worlds bit-identical:

1. Initial entities are created in instance-set order, then in index order within a set; each entity schedules its kind's channels in
   declaration order. The event queue breaks ties by scheduling order, so a single-kind world schedules exactly as before.
2. Churn births of the first instance set draw from `churn/births`, as before; the set at index `i > 0` draws from `churn/births/i`.
   Retirement draws stay `churn/lifetimes/<entity id>`. Each set keeps its own next index, starting at the number of its initial
   entities.
3. Substream names carry the entity id, channel, fault id or effector name, all domain-unique, so adding a kind to a domain does not
   change a draw of the existing kinds (`TestEntitiesOfOneKindDrawTheSameStreamWhetherOrNotAnotherKindExists`).

## 3. The runtime form

The validated v0.2 document, decoded into `model.DomainSpec`, **is** the runtime form: the world, perturbation layer, adapter engine,
truth, scoring and run code read `Kinds` and `Instances` directly. (P5a lowered the document to the flat v0.1 shape; that
intermediate form and `lowering.go` are removed in P5a2.) The cross-checks run per kind on the document, so their pointers are
document pointers with no mapping.

| Concept | Runtime (`model.DomainSpec`) |
|---|---|
| `id`, `version`, `title`, `description`, `axes`, `stresses`, `profiles`, `ground_truth` | same member |
| `kinds[k]` | `Kinds[k]`: `Name`, `State`, `Dynamics`, `Channels`, `Faults`, `Effectors` |
| `instances[i]` | `Instances[i]`: `Kind`, `IDTemplate`, `Count`, `Churn` |
| an entity | `Entity.Type` is the kind name; the world keeps its `*model.Kind` and instance-set index |

Domain-wide lookups (`Compiled.Fault`, `Channel`, `Effector`, `FaultKind`, `Kind`) resolve a unique name to its declaration; per-kind
lookups (`StateNames(kind)`, `KindChannelNames(kind)`) list one kind's names.

**Digest.** `domain_digest` is the RFC 8785 digest of the validated **v0.2 document**, as v0.1's was of its document. Formatting
never changes it. Every shipped domain's digest therefore changes once, with its bytes, and `world_digest` (which includes the domain
digest and `sim_version`) with it. `Compiled.Raw`, the document embedded in run artifacts as `domain_spec`, is the v0.2 document;
replay parses it through the same path.

**`sim_version`** is `0.3.0`: the digest inputs of every run changed.

**Equivalence.** The runtime change is a re-expression, so a single-kind domain must behave exactly as before P5a2:

1. `TestEveryShippedDomainCompilesToItsV01RuntimeWorld` projects each shipped domain's single kind and instance set back into the
   v0.1 flat shape and compares its RFC 8785 digest with the digest recorded before the P5a rewrite (the same recorded values);
2. the behaviour pin shows `trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, every `verify` and `replay` result and the
   generated suites byte-identical for all eight domains and both adapters. The one deliberate change is `catalog describe`, which now
   prints the v0.2 document (`kinds`, `instances`) instead of the flat runtime shape.
3. A two-kind test domain (`internal/testsupport/testdata/two-kind-plant.domain.json`) passes the analytic cross-check for each kind,
   the truth solver's closed-form crossing times, the suite and audit rules above, and run replay identity.

## 4. Validation issues carry pointers (finding R1-01)

Every issue a domain produces is *located*: an RFC 6901 pointer into the **v0.2 document**.

- Schema violations: a "missing required property" issue points at the missing member (`/kinds/0/channels/2/resolution`), not at
  its parent; "additional property not allowed" points at the unexpected member. (Both come from the shared validator, so every
  contract benefits.)
- Cross-reference issues (an undeclared state, fault or effector, a duplicate name, a dynamics cycle, an unimplemented option) point
  at the member that is wrong: `/kinds/0/channels/1/observes`, `/kinds/0/faults/2/affects/0/state`,
  `/profiles/0/fault_weights/<fault>`. The checks run on the document itself, so the pointer is built where the check is made.
- A document that is not JSON has no location and keeps one unlocated issue.

## 5. Explicit entity ids (finding R2-02)

A scenario's `world.entity_ids` and the director's `sim.world.create` entity list name instances explicitly. **Decision:** an explicit
id list is checked against the domain's instance set, by the run module when it creates the world, so every surface (CLI flags,
scenario file, director, replay) is covered once:

1. each id matches the `id_template` of some instance set (the first that fits): `{n}` matches digits, any other placeholder matches
   `[A-Za-z0-9_.-]+`, literal text matches exactly;
2. for each set that has ids, their number is within `count.min` (at least 1) and `count.max`;
3. every instance set's own `count.default` must lie within its `count.min` and `count.max` (a domain issue at
   `/instances/i/count/default`).

An empty list selects the default population: each set's `count.default` ids, sets in declaration order. A violation is a refusal at world creation (a scenario violation at `/world`).
`Compiled.CheckEntityIDs` implements the rule; P5b extends it to names (section 8). Ids that fit the template are all the template can render, so churned births are
unaffected.

## 6. Canonical formatting

`streamsim domain format <path>` rewrites a valid domain file in place and prints `{formatted, changed, digest}`. The format is a pure
function of the document:

- two-space indentation, newline at the end of the file, strings written as UTF-8 without HTML escaping, number text preserved;
- object members in a fixed order: the document's members in schema order (`id`, `version`, `title`, `description`, `axes`,
  `stresses`, `kinds`, `instances`, `link_types`, `links`, `boards`, `profiles`, `ground_truth`); within any other object the
  identifying and structural members first (`id`, `name`, `kind`, `system`, `names`, `id_template`, `count`, `churn`, `target`, `tier`, `form`, `state`, `unit`, `initial`, `default`, `min`,
  `max`, `dynamics`, `channels`, `faults`, `effectors`), then the rest alphabetically, then `description`;
- array order is semantic and never changed;
- an object or array that fits in 110 columns is written on one line, otherwise one member per line.

The command refuses a document that does not validate and leaves the file untouched. It is idempotent, and the digest does not change.
The canvas and agents call it so a hand edit and a canvas edit of the same world produce the same bytes.

## 7. Deviations recorded

[`DECISIONS.md`](../DECISIONS.md) D-03, D-04 and D-14 carry deviation entries describing what v0.2 does and does not change in them.
Part 1 changes none of their runtime behaviour.

## 8. Named instances (finding R5a-02)

`instances[i]` may carry `names`, a list of explicit entity ids (`M1`, `AC1`). A set needs `names`, an `id_template`, or both.
Sections 8 to 12 are phase P5b; every issue is located at a pointer into the document.

- The initial population of a set is its names in order, then `count.default` templated ids. `count` belongs to the template: it
  is required with an `id_template` and refused without one, and so is `churn` (births render ids from the template).
- A name is unique across the domain, matches the entity-id alphabet of `sim-event`, and must not fit the `id_template` of **any**
  set (a template could render it too, and churn or `entity.add` could then collide with it).
- `Compiled.InstanceSetFor(id)` tries names across all sets first, then templates in set order, so a name wins over a template
  of an earlier set. `CheckEntityIDs`, the world (`AddEntity`, initial population) and the links and boards below all use it. In
  the explicit-id rule (section 5) a name is accepted and not counted against `count.min` or `count.max`.

## 9. Overrides: add or narrow, never remove or retype

`instances[i].override` varies the instances of **one set** (a named instance is a set of one name). The effective kind of the
set is the kind plus the override; a set without an override shares the kind itself.

| Member | Meaning | Rule |
|---|---|---|
| `state`, `channels`, `faults`, `effectors` | Added declarations | Same definitions as in a kind. Channel, fault and effector names stay unique across the domain, so redeclaring a kind's is the duplicate issue; a state name must not repeat one of the kind (or of the override) |
| `dynamics` | Added dynamics | May drive only states the override adds (changing how an existing state evolves would retype it); inputs may read any state of the effective kind; cycles are refused |
| `narrow.state[]` | `{name, min, max}` | Names a state of the kind; `min < max`; inside the kind's bounds when it has any (a state with neither bound is unbounded); the initial value lies inside |
| `narrow.channels[]` | `{name, range{min,max}}` | Names a channel of the kind; `min < max`; inside the kind's range when it has one |
| `narrow.effectors[]` | `{name, args{<arg>{minimum?, maximum?}}}` | Names an effector of the kind and a numeric property of its `args_schema`; a bound may only tighten the kind's `minimum` or `maximum`; `minimum <= maximum` |

Nothing else is expressible: the schema has no member for removal and refuses unknown members, so removing or retyping (a unit, a
value type) is a located schema issue. Added items are validated by the same checks as a kind's, against the **effective** kind, at
`/instances/i/override/...`.

**Runtime.** The world gives each entity the effective kind of its instance set, so an added state integrates, an added channel
emits, an added fault or effector is accepted only on entities of that set (others are refused as for any other kind), and a
narrowed state bound clamps that entity's reads. A narrowed effector argument is enforced when that entity is invoked; a narrowed
channel range is the range `out_of_range` perturbation exceeds for that entity. Domain-wide lookups (`Compiled.Fault`, `Channel`,
`Effector`, `FaultKind`) also find added declarations; `EntityChannel(id, name)` and `InstanceKind(set)` give the entity's view.
Peers (section 2.3) remain the entities of the same kind, overridden or not. Draw order: substream names carry the entity id, so
an override changes only the draws of its own entities.

**Limits.** The scenario checker and the operator nameplate list effectors and channels domain-wide, so they show a kind's
declaration; a narrowed argument is enforced at invocation, not by the scenario pre-check. A suite scenario draws its entity from
those that carry the fault, so a fault added by an override is only injected into the entities of that set.

## 10. System entities

`kinds[k].system: true` marks a kind of plant-wide measurements (`air_header`, `power_bus`). Decision: a system kind is named by
**exactly one** instance set with **exactly one name and no `id_template`** (so no churn). A channel therefore belongs to exactly
one entity; for any other kind the channel belongs to each entity of the kind. Issues: `/kinds/k/system` (not one set),
`/instances/i/names` (not one name, or a template), and a `peer_residual` fault of a system kind (one entity has no peers) at its
detector form.

## 11. Typed links

`link_types[]` = `{name, from, to, description?}` declares a kind of link between two kinds (`feeds`: `tank` to `pump`).
`links[]` = `{type, from, to}` joins two instances. Issues: duplicate link type name (`/link_types/i/name`), undeclared `from` or
`to` kind (`/link_types/i/from`), undeclared type (`/links/i/type`), an endpoint that is no name and fits no template
(`/links/i/from`), an endpoint of the wrong kind, and a link stated twice (`/links/i`). Endpoints must be identifiable instances;
whether an index lies within the starting population is not checked. Links are structure only: the aggregation of dynamics inputs
over links (`sum`, `mean`) is P6.

## 12. Boards

`boards[]` = `{name, members[], description?}` groups the instances hosted on one device. Decision: the **board owns the `boot_id`**
(plant wire [section 14](../simulation-canvas/plant-wire-protocol.md) and finding P6 of the review): a reboot affects every member
together. Rules: board names are unique (`/boards/i/name`); each member is an identifiable instance (`/boards/i/members/j`); an
instance is on at most one board, including twice on the same board. System entities may be members. Only the grouping is modelled
here. Issuing `boot_id`, rebirth after a reboot and the lease and safe state of a board belong to the plant-wire phases (P7).

## 13. Stores and flows (phase P6a)

`f1.form: "store"` with `capacity` (> 0) is an accumulator: its level changes by `(Σ coef × input + inflows − outflows) /
capacity` per second, kept within the state's `min`/`max` and the form's `clamp`. `kinds[k].flows[]` = `{name, from, to?,
conductance, scale[]}` moves `conductance × Π scale × (from − to)` per second from one store to another; without `to` it drains
toward zero. `scale` names actuator and health states (a pump speed, a pipe health a fault lowers).

Decisions:

- **Integration order.** Every integrated state of an entity advances together, earliest step first, and every store of an
  entity steps on one grid with one RK4 over the vector of levels. Each state keeps snapshots of its last steps back to the
  entity's longest step, so a state reading another at a step boundary gets its value at that exact time. Without this the
  result depended on which state was read first (finding 1 of the P6a review, which also held for a lag fed by a lag before
  P6a), so an emitting world and a silent oracle world could differ. Every store of a kind, override additions included,
  shares one `dt_ms` (`/kinds/k/dynamics/i/dt_ms`). A level is kept within the form's `clamp` and the state's `min`/`max`.
- **Hybrid determinism.** Inputs and scale states are read at the step boundary and held for the step, so the system is linear
  between steps, as the vocabulary's rule requires. A fault or effect between boundaries acts on a flow from the next boundary.
- **Faults act through scale states.** A flow reads the integrated level of each store, not a fault or effect laid over it; a
  blockage or a leak is modelled by a health state in `scale`.
- **Scope.** Flows join stores of one entity. Flows between instances (along links) are the link aggregation of a later phase.
  An override may add a store state but not a flow.
- **Issues:** a duplicate flow name (`/kinds/k/flows/i/name`), a `from` that is no store, a `to` that is no other store, an
  undeclared scale state (`/kinds/k/flows/i/scale/j`), a store without `capacity`.

The analytic cross-check is independent of the RK4: two stores against the closed form (conservation of `C_a·a + C_b·b` and
exponential decay of `a − b`), a drained store with inflow against its exponential, the transfer plant fixture against the 2×2
matrix exponential, and the observability solver's crossing of a pipe blockage against crossing times derived from that matrix
exponential. Tolerance 1e-6 absolute; the crossing times agree exactly on the 60 s grid.

## 14. Closed grammar (finding R0-11)

The schema holds no expression. A test walks the embedded contract and classifies every unconstrained string member as prose
(`description`, `title`, `unit`, ...), an identifier (`name`, `id_template`, ...) or a reference the cross-check resolves
(`state`, `channel`, `from`, `scale`, ...); a new free string fails it until classified. Objects of unchecked members are limited
to effector argument schemas and arguments, reference-model parameters and profile weights, and no member is named like an
expression (`expr`, `formula`, `condition`, `script`, ...).

## 15. Smaller rules settled in P6a

- **Narrowing** names each state, channel and effector once, and an argument's narrowed bounds must leave a value: a minimum
  above the declared maximum is refused (R5b-02).
- **A `peer_residual` fault added by an override** may not watch a channel only that override adds: its set has no peers that
  carry the channel (R5b-01).
- **A profile's entity count** sizes the first instance set that has an `id_template` (R5b-03).
- **Idempotency** compares argument numbers by value: `3` and `3.0` are the same request (R2-05). Numbers are compared as
  64-bit floats, so integers above 2^53 that differ only beyond that precision are one request.

## 16. Schedule, regime and peak (phase P6b)

- **`f0.schedule`** = `{period_s?, interpolation: step|linear, points[{at_s, value}]}` adds an exogenous piecewise profile to the
  F0 value (a tariff, an order book, a load shape). With `period_s` the profile repeats and `at_s` is the offset within a period
  aligned to the Unix epoch, so a daily profile follows the clock whatever the world start; without it `at_s` counts from the
  world start. `step` holds the last point reached (before the first point: the last point of the previous period, or the first
  point when not periodic); `linear` interpolates between neighbours, across the period boundary when periodic. Points are in
  strictly increasing `at_s`, inside the period (`/…/f0/schedule/points/i/at_s`).
- **`f1.form: "regime"`** = `{selector, threshold, below: {gain, time_constant_s}, above: {…}, inputs}`: a first-order lag toward
  `gain × Σ coef × input` whose gain and time constant switch when the selector state is at or above the threshold. The selector
  (the target itself for a runaway) is read at each step boundary, so the switch happens at the first step boundary at or after
  the analytic crossing time; between switches each regime has its exact exponential. An undeclared selector is refused at
  `/…/f1/selector`.
- **`f1.form: "peak"`** = `{inputs: [one], block_s, period_s, ratchet?, previous_peak?}`: the demand peak. The input is sampled at
  each step start; a block (aligned to the Unix epoch) closes when a step crosses its boundary, and its mean raises the period's
  peak; a billing period (aligned the same way) rolls the peak into the previous peak and starts again from zero. The value is
  `max(period peak so far, ratchet × previous peak)`; the state's `initial` seeds the period peak so far and `previous_peak` the
  previous period's peak (S9: seeded `peak_mtd` and ratchet). The value changes only when a block closes, so a read between
  steps returns the committed value. When the world starts (or an entity is born) inside a block, its first block is partial:
  it closes at the clock boundary with the samples it has. `block_s` is a whole number of steps and `period_s` a whole number of blocks
  (`/…/f1/block_s`, `/…/f1/period_s`). A calendar month is not a fixed `period_s`; billing periods of unequal length, and the
  projected demand of the open block, are left to the reference scenario (P12).

Cross-checks: the schedule against direct table values at boundary and interpolated times; the regime against the piecewise
closed form switching at `ceil(τ_below · ln((g·u − x0)/(g·u − threshold)))` (1e-6); the peak against an independent recomputation
of block means, period roll-over and ratchet over two days (1e-9).

**Site clock (P7k).** A periodic schedule may declare `utc_offset_s` (it needs `period_s`: a non-zero offset without a period is refused at `/…/f0/schedule/utc_offset_s`): its `at_s` is then local time within the period, so a
daily profile or a weekly shift calendar follows the site's clock. In an epoch-aligned weekly period, Monday 00:00 is
`at_s` 345600 (1970-01-01 was a Thursday).

## 17. Protection: conditions, trips, anti-recycle, boundaries, preconditions (phase P6c)

Review item S14 separates three things the word `lockout` covered: a **latched trip** (`lockout`), a **timer** (`anti_recycle`)
and **lockout-tagout isolation** (a state that a precondition checks).

- **`condition`** = `{state, operator: lt|lte|gt|gte, threshold}`: the one closed comparator set, shared with the interlock.
- **`f1.form: "trip"`** = `{when: condition, for_s, tripped_value, latch?, reset_on?}`. The condition is sampled at each step
  boundary; once it has been sampled true for `for_s` the target takes `tripped_value` from the next boundary, otherwise it holds
  the state's `initial`. An unlatched trip clears when the condition stops holding; a latched trip (a lockout) stays tripped until
  the `reset_on` state changes at a boundary (a reset command adds 1 to a counter), and trips again at once if the condition
  still holds. A latched trip needs `reset_on` (`/…/f1/reset_on`).
- **`f1.form: "anti_recycle"`** = `{watch, edge: fall|rise, hold_s}`: an edge of the watched state at a boundary starts a timer;
  the target is the time left in seconds, 0 once elapsed.
- **`kinds[k].boundaries[]`** = `{name, state, severity: critical|major|minor}`: a safety boundary is a trip state, so a
  boundary and a trip share one evaluator; the world reports the boundaries violated at a time (`world.ViolatedBoundaries`) for
  scoring (P12). The state must be a trip (`/kinds/k/boundaries/i/state`).
- **`effectors[e].requires[]`**: preconditions the device checks at the call time. Not locked out, anti-recycle elapsed, not
  isolated, not in hand mode are each a condition on a state. A failed precondition refuses the call with no effect and no
  random draw; the operator reads the generic `effector_refused` (MCP_SURFACE §4.5) and the director's call log records
  `precondition_refused`.
- **Effects start strictly after their time**, as before: a command at `t` is seen by a condition sampled after `t`.

**Look-ahead.** A command may be dated after the world clock (a scenario file lists its calls before the clock advances). Its
interlock and preconditions are evaluated on a throwaway copy of the entity's integration state, so evaluating them never
commits integration past the clock and never changes a reading at an earlier time. Before this, a future-dated call integrated
the entity ahead and an earlier emission read the later value; that affected interlocks on integrated states too.

**Notes.** A refused precondition, like an interlock refusal, is not cached for idempotency: a retry with the same
`command_id` is evaluated again and may succeed once the condition holds. `anti_recycle` and `reset_on` react to any change of
their state, so they should watch discrete states (a run flag, a reset counter), not a continuously moving one. The
conditions of `trip`, `anti_recycle` and `regime` are read at each step's start, so a cycle through them is a one-step delay,
not an algebraic loop; the dependency graph does not list them. A boundary is read on the look-ahead copy when asked about a
time after the clock. `tripped_value` must differ from the state's initial value (`/…/f1/tripped_value`).

Cross-checks: the trip fires on the step computed from the analytic crossing of an F0 ramp (`floor(t_c) + 1 + for_s + 1`);
latch, clear and reset at the expected boundaries; the anti-recycle timer reads 599 s one step after its edge; preconditions
refuse during the timer, while locked out and while isolated; a call dated ahead of the clock leaves earlier readings unchanged.

## 18. Supervisory control: setpoints, sequencer, responder, lease (phase P6d)

Review items D3 and S2: the factory is commanded at supervisory level (setpoints, bands, demand limits) and the plant's own
sequencers act on them, so actions do not fight contactor-level logic.

- **Setpoints in conditions.** A `condition` may name `threshold_state`; its threshold is then that state's value plus
  `threshold` (an offset or deadband). A supervisory effector assigns the setpoint state (`assign_from_arg`), and every trip,
  precondition and sequencer reading it follows.
- **`f1.form: "sequencer"`** = `{stage_up: condition, stage_down: condition, stages, hold_s}`: the target counts running stages
  (0 to `stages`). At each step boundary, once `hold_s` has passed since the last change, one stage starts while `stage_up`
  holds, or one stops while `stage_down` holds (start wins). Other dynamics use the count as an input (compressor output).
- **`effectors[e].responder`** = `{match_arg, delay_s, on_match, on_miss, clear_fault}`: a simulated person or crew. When the
  call is acknowledged, the responder arrives `delay_s` later; if the string argument `match_arg` names a fault active on the
  entity at the arrival time (as known when the call is made), `on_match` applies at arrival and `clear_fault` clears the
  fault, otherwise `on_miss` applies. The outcome shows only through states; the acknowledgement is the same either way.
  Responder availability, reports and approvals are the reference scenario's (S6, P12).
- **`effectors[e].lease`** = `{arg, fallback: safe_off|hold_last|revert, safe_deltas}`: an acknowledged call holds its effect for
  the seconds in its numeric argument `arg`. Unless the same effector is acknowledged again on the entity before or at the instant the lease lapses
  (which withdraws the pending fallback and starts a new lease), the fallback applies at the lapse: `safe_off` sets each
  `safe_deltas` state to its `delta`, `revert` undoes the call's additive deltas (an assigning effect cannot be reverted,
  `/…/lease/fallback`), `hold_last` keeps the effect. A refused or unacknowledged call does not renew. Two assignments to one state at
  the same instant resolve to the later-issued one.
- **Lease timing.** A lease counts from the effect's start (the call time plus the effect's dead time). A `revert` fallback
  undoes the total additive contribution of every call in the lease chain (a renewal carries the withdrawn revert into the
  new one), with the effect's own time constant, so the state returns to where it started. A call without the lease
  argument neither renews nor ends the lease. Only acknowledged calls (`ok`, `slow`, `ack_lost`) arm a lease or dispatch a
  responder; a `partial`, `confirmed_no_effect` or `silent_no_effect` call does neither.
- **Responder timing.** The match is decided when the call is acknowledged, from the faults injected so far; a fault injected
  later, or cleared later by the director, does not change an already scheduled outcome.
- **Sequencer start-up.** The first stage change waits `hold_s` from the entity's first step (world start or birth).
- **Issues:** an undeclared `threshold_state`, a responder `match_arg` that is no string argument, a lease `arg` that is no
  numeric argument, a `safe_off` lease without `safe_deltas` or with a delta read from an argument, a responder delta reading
  an argument that is no numeric argument, a delta on an undeclared state.

Cross-checks: the sequencer stages at the exact boundaries implied by the band crossings and `hold_s`, up and down, capped at
`stages` and 0; a responder told the active fault clears it at arrival and applies `on_match`, a wrong cause applies `on_miss`
and leaves the fault; a lease falls back at its lapse (safe value and revert) unless renewed, and renewal moves the lapse.

## 19. Event clock (phase P6e)

Review item S11: steps of a second or a minute cannot see a 100–500 ms voltage sag. Dynamics keep their step grid; the event
clock adds exact instants between steps.

- **`onset.shape: "pulse"`** with `duration_s`: the fault acts on `[onset, onset + duration_s)` and then clears by itself (a
  director clear may end it earlier). Its start and end are **event instants**. `duration_s` on any other shape is refused
  (`/…/onset/duration_s`).
- **Trips on the event clock.** A trip samples its condition at its step start and at every event instant of its entity inside
  the step. Between samples the condition is taken as constant, and the trip fires when the condition has held from a sample up
  to the next sample (or the step end) for at least `for_s`. A 0.3 s sag trips a 0.2 s ride-through trip; a 0.1 s sag rides
  through. The tripped value appears from the next step boundary, as for every trip; an unlatched trip that fired and cleared
  inside one step leaves no mark at a boundary (latch it, or use a boundary over a longer `for_s`). This also makes a trip
  visible at the boundary where `for_s` has elapsed since the first true sample (one step earlier than in P6c).
- **Event-driven channels on the event clock.** When a pulse fault is injected, each event-driven channel of the entity that
  observes (or is triggered by) a state the pulse affects is checked at its start and end, besides its regular checks, so a
  power-quality channel records the sag and the recovery at their exact times. A channel that does not see the pulse is never
  checked at its instants, so its records cannot reveal an injection time. A check emits only when the reading changed by the
  channel's resolution, does not move the regular checks, and is never scheduled before the world clock.
- **Clearing.** A director or responder clear can only shorten a fault, never extend one that already ended; an early clear of
  a pulse is itself an event instant and is checked.
- **Trip timing.** A trip fires when its condition has held from a sample up to the next sample for `for_s` **and** still holds
  just before that next sample, so a hold that ends inside the last step does not trip.
- **Unchanged worlds.** Only pulse faults create event instants, so a world without them integrates, samples and draws exactly
  as before (behaviour pin unchanged).

**RNG draw order (finding R0-12).** Every random draw comes from a named substream, seeded from the run seed and the full name
`<world_id>/<name>`; substreams are independent, and within one the draws happen in the order of the events that consume them
(world events in time order, ties in scheduling order). The event clock adds no substream: its only draws are those of the
readings it emits, from that channel's own substreams.

| Substream (`<world_id>/` omitted) | Consumed by | When |
|---|---|---|
| `<entity>/<channel>/noise` | Measurement noise | Each reading (regular or event-clock check) with sigma > 0, unless the model is `quantization` or the world is noiseless |
| `<entity>/<channel>/drift` | Sensor drift | Each reading after the first, for a `random_walk` drift |
| `<entity>/<channel>/enum` | String channel values | A change of an event-driven string channel |
| `<entity>/<channel>/jitter` | Periodic cadence jitter | Scheduling the next regular emission when `jitter_s` > 0 and the world is not noiseless |
| `<entity>/<channel>/human` | Human-driven cadence | Scheduling the next human emission |
| `<entity>/<channel>/availability` | Outages | Advancing the channel's availability when uptime < 1 (an event-clock check advances it too) |
| `link/<entity>/<channel>/delay` | Transport delay | Each emitted record, unless the model is `constant` |
| `<entity>/<effector>/fault_shape` | Effector failure mode | Each executed call (after interlock and preconditions) of an effector that declares failure modes |
| `<entity>/<effector>/delay` | Acknowledgement latency | Each executed call |
| `fault/<fault>/<entity>/<onset>/<state>` | Stochastic fault envelope | One walk per injection and state on a fixed grid (the state's step, 60 s for an input), drawn in grid order the first time any step is read, so any read order, look-ahead or re-injection gives the same values |
| `churn/births`, `churn/births/<set>` | Churn births | Each birth of the instance set |
| `churn/lifetimes/<entity>` | Churn lifetimes | Each churned entity's birth |

Integration (every F1 form, stores, trips, sequencers), schedules, responders, leases, preconditions and the event clock's
trip samples draw nothing.

Cross-checks: a 0.3 s sag between step boundaries trips a 0.2 s ride-through trip and a 0.1 s sag does not, while the voltage at
both neighbouring boundaries is nominal (so a boundary-sampled trip would miss both); an event-driven channel records the sag at
600.4 s and the recovery at 600.7 s; the periodic channel's records are identical with and without the sag.

## 20. Composed instance sets: `area` and `provenance` (phase P7g)

An instance set may carry `area` (an ISA-95 area or unit name, for display and grouping) and `provenance` (where a composed set
came from: component id, major version, the component file's digest, the preset, parameter values, the instrument attached at
each measurement point and each instrument file's digest; [COMPONENTS_V0_1](COMPONENTS_V0_1.md) §7). A domain may carry `site`
(`name`, `timezone`, `areas` with parents) and a channel `tag` (the ISA-5.1 letters of its measurement point). All are structure
only: the world never reads them, and they enter the domain digest like any other member.

## 21. Coupling over links (phase P7h)

A kind reads the states of the instances linked **into** it, over a declared link type whose `to` is that kind. Two forms:

- **A linked input**: an F1 input `{state, coef, link: {type, aggregate, default?}}` contributes `coef × aggregate`, where the
  aggregate is taken over the `state` of every instance linked into this one by `type`. It works in every form that reads inputs
  (lags, integrators, stores, regimes).
- **`follow`** on a dynamics entry, `{sources: [{link, state}], aggregate}`: while at least one instance linked in over any of
  the sources' link types is present, the target state is the aggregate over all of them, each source's instances read at its
  `state` (sources in their order, each in sorted instance-id order), so one header sums the air demand of several kinds of line
  (R7k-02); otherwise the entry's own `f0` or `f1` applies, and it keeps integrating in the background. A link type appears in
  one source of a follow at most. A store, peak, trip, anti-recycle or sequencer state cannot follow.

| Aggregate | Value | With no linked instance present |
|---|---|---|
| `sum` | the sum | 0 |
| `mean`, `min`, `max` | the mean, the least, the greatest | the input's `default` (required); `follow` falls back to its own dynamics |
| `count_running` | how many linked states are at least 0.5 | 0 |

Rules:

- A linked instance is present at `t` when it was born at or before `t` and, if retired, retired at or after `t`; an instance a link
  names that does not exist contributes nothing. Values are read through the source's full state path (dynamics, faults, effector
  kicks, clamp), and the sum runs in sorted instance-id order, so the value does not depend on which instance was read first.
- Validation: the link type is declared, its `to` is the reading kind, its `from` kind declares the state, and `mean`, `min` and
  `max` inputs carry a `default`. Kinds may not read each other in a cycle over links, a kind reading itself included; a
  feedback loop needs an explicit delay form, not yet specified. A linked input is no dependency inside the reading kind.
- Determinism: a source keeps its history back to the oldest time any instance downstream of it (its readers, their readers,
  and so on) can still read, so a read never falls on pruned history. Before a source steps past `t` the world also steps its
  readers to `t` (a retired reader only to its retirement), which bounds that history. A value read forward in time therefore does
  not depend on which instances were read first or how often; the world tests prove it against an independent recursion for a
  three-level chain with unaligned steps, a relay with no history of its own, and seeded random forward read schedules.
- Reading into the past: like any state, a linked value is defined back to the start of the history its sources keep; a
  `follow` state, which keeps no history of its own, is no exception. A retired instance's states are defined up to its
  retirement while its sources still keep that history.
- A domain without linked inputs or `follow` behaves exactly as before: links remain structure only.

## 22. Validation and world corrections (phase P7i, prerequisites)

- **State bounds** are independent: `min` alone bounds below, `max` alone bounds above, and a state declares neither to be
  unbounded (R7e-04; before, `min` alone was ignored and a positive `min` alone clamped to 0).
- **Assignments with a time constant** approach their target from the value they replace (the state's value with every earlier
  driver applied): `target − (target − underlying)·e^(−(t − start)/τ)`. A `silent_no_effect` shadow assignment follows the same
  rule (R7e-04; before, the ramp started from 0).
- **Stability** is validated (R7e-04, R6a-01): a lag, an RC network stage or a regime must step below `dt/τ = 2.78` (the RK4
  stability bound), a dead time with a lag below `dt/τ = 2` (explicit Euler). A store must step below `rate·dt = 2.78`, where
  `rate` is the Gershgorin bound of its exchange: the sum over its flows of `conductance × scale bound / capacity`, doubled for a
  flow to another store; a scale state's declared `max` bounds its factor, and a scale state with no `max` counts as 1.
- **Flow scales** are not negative: a scale state below 0 stops the flow instead of reversing it (R6a-01).
- **Id templates** of two instance sets may not render a common id (exact check over the templates' languages; R7e-04).
- **Leases**: the lease argument's schema declares `minimum` or `exclusiveMinimum` of at least 0, so a negative lease is refused
  at invocation (R7e-04).
- **Observation bias**: an absent `noise_scale_at_full` is 1, as the schema documents (R7g-01; before, it silenced the noise).
- **The operator view** of an effector (nameplate and `sim.effector.list`) is its name, risk class and argument schema with the
  annotation keywords (`description`, `title`, `$comment`, `examples`) removed at every depth; effector descriptions are
  director-side text and may name the faults an effector remedies (R7e-09).

## 23. Algebraic forms: `curve`, `product`, `sum` (phase P7i)

Three F1 forms are algebraic: the state is a function of other states at the same instant, evaluated when it is read, with no
step, no history and no `dt_ms`. Each input is weighted (`coef × state`, or `coef × aggregate` for a linked input, §21); the
result passes the form's `clamp` and then the state's bounds, faults and effector kicks like any state.

| Form | Value | Shape |
|---|---|---|
| `sum` | `Σ coef × input` | one or more inputs |
| `product` | `Π coef × input` | 2 to 4 inputs |
| `curve` | a piecewise-linear table of the weighted inputs, held at its end values outside the table | one input: `table: {x, values}`; two inputs: `table: {x, y, grid}`, bilinear, `grid[i][j]` at `(x[i], y[j])` |

Axes increase strictly and hold at least two points; `values` has one entry per `x`, `grid` one row per `x` of one entry per `y`.
Algebraic states take part in the dynamics dependency check, so an algebraic loop is refused. These are tables and bounded
products, not an expression language: an efficiency curve, a pump or fan law, a heat duty `flow × ΔT` (with `ΔT` a `sum`), or a
power that scales with `running` (R7g-02) is written with them.

## 24. Controller: `pid` (phase P7i)

A `pid` dynamics is a sampled controller: at each step boundary `t` it reads `pv` and `sp` and computes its output, which the
target holds from the end of the step (one scan of latency, like a control system's output). With `s = +1` for `action:
reverse` (the output rises while `pv` is below `sp`) and `s = −1` for `direct`, step `h` seconds and the previous step's `pv`:

- error `e = s·(sp − pv)`; proportional `kp·e`; derivative on the measurement `−s·kd·(pv − pv_prev)/h`;
- integral `I + ki·e·h`, kept only when the unclamped output does not push further past a limit (conditional integration: no
  windup); the output is `clamp(P + I + D)` with a required `clamp` of `min` below `max`;
- the integral starts at the target's initial value, so the first output is near it;
- while the `manual` state is at least 0.5 the output follows `manual_output` (clamped) and the integral is set to
  `output − P − D`, so returning to automatic does not bump.

Gains are non-negative and `kp` or `ki` is non-zero. A cascade chains controllers: an outer controller's target is an inner
controller's `sp`. Closed-loop stability is the domain author's: the validator cannot know the plant.

## 25. Observation faults and quality (phase P7i)

A fault may act on an instrument rather than the process: `observation: {channel, form, value?, factor?, quality?}` in place of
(or beside) `affects`. While the fault's envelope (onset shape × severity) is above zero, the channel of that entity reports:

| Form | Report |
|---|---|
| `stuck` | `value`, quantized |
| `freeze` | its noise-free reading at the fault's onset, captured when the world reaches the onset |
| `dropout` | nothing: no event, as when the producer is down |
| `scale` | the reading multiplied by `1 + (factor − 1)·envelope` (a gain error; a ramp onset makes it grow) |
| `noise` | the noise sigma multiplied by `1 + (factor − 1)·envelope` (a noise burst) |

The process is untouched: states, other channels and confirmations read the true values. Several faults may act at once; the
latest stuck or freeze wins, and scale and noise factors multiply.

**Quality.** An event carries `quality` (`suspect`, `out_of_range`, `stale`, `disconnected`, the plant-wire vocabulary) while an
acting observation fault declares one; the worst wins. Absent means good, and `quality: good` (the default) is the silent case
where the instrument vouches for a wrong value. The native adapter passes `quality` through, and the plant wire sends it in
`DDATA` instead of `good`.

**Validation and truth.** The channel is a number or counter channel of the fault's kind (any channel for `dropout`), a `noise`
fault needs a noisy channel, and the detector is `single_channel_snr` on that channel. The truth solver measures an
observation fault by what it does to the channel: the change in the noise-free reading plus the change in the noise sigma, and
an immediate detection for a dropout.

## 26. Alarm: ISA-18.2 (phase P7i)

An `alarm` dynamics watches a state against a limit and holds an ISA-18.2 state code as its target: `0` normal, `1`
unacknowledged, `2` acknowledged, `3` returned to normal unacknowledged, `4` shelved. It is sampled at each step boundary and
its code holds from the end of the step, like a control system's alarm scan.

- **Condition**: `direction: high` is `watch > limit`, `low` is `watch < limit`; `limit` is a number or `limit_state` a state
  (exactly one). Once active, the alarm stays active until the value is `deadband` back inside the limit.
- **Delays**: the condition must hold for `on_delay_s` at successive samples before the alarm activates, and the clear condition
  for `off_delay_s` before it returns.
- **Acknowledgement**: a change of the `acknowledge` state (an acknowledge command's counter) acknowledges: an active alarm
  becomes acknowledged, a returned one becomes normal. Every new activation needs a new acknowledgement.
- **Shelving**: while the `shelve` state is at least 0.5 the code is `4`; the alarm keeps its state underneath and shows it again
  when unshelved. A lease on a shelve command (§18) gives a timed shelve.
- **Priority** (`low`, `medium`, `high`, `critical`) is declared for the alarm-system KPIs (rate, standing and chattering
  alarms), which belong to the KPI truth of P7k.

A channel that observes the alarm state shows annunciation to a consumer; acknowledge and shelve are ordinary effectors.

## 27. Domain lint (phase P7j)

`streamsim domain lint <domain.json> [--hours H] [--seed S]` judges a domain that already validates by simulating it: a nominal
day, so a composed plant that is valid but behaves absurdly is caught before use. It reads the domain and changes nothing.

**The nominal day.** A noiseless world (`Noiseless`, `EmitDisabled`) with the default population, seed `S` (default 1), from
the default world start for `H` simulated hours (default 24, at most 720), with no fault and no effector call. Every entity's
every state (`StateValue`) and every number or counter channel (the noise-free `Reading`) is sampled every 60 s from the start,
plus the end of the run. Samples are read in time order across all entities.

**Findings.** Each carries `severity` (`error` or `warning`), `code`, `entity`, the `channel`, `state` or `fault` it concerns,
`at_s` (simulated seconds since the start) and a `detail` with the numbers. A series is reported once per code, at the first
sample that triggers it.

| Code | Severity | Fires when |
|---|---|---|
| `non_finite` | error | a state or reading is NaN or infinite; the other rules skip that series |
| `out_of_range` | error | a channel that declares a `range` reads outside it; `detail` gives the worst reading and how many samples are outside |
| `pinned_at_bound` | warning | an integrated state (tier F1, not algebraic, not a discrete form: `trip`, `anti_recycle`, `sequencer`, `pid`, `alarm`, and not the target of a `follow`) sits at its declared `min` or `max` in more than 50 % of the samples, whatever its span: an integrated state that never leaves its bound (span 0) is reported too. A tier F0, scheduled, binary-flag, discrete-form, algebraic or followed state is never reported |
| `not_settled` | warning | an integrated state (as above, without the follow exclusion) changes between the start of the last 10 % of the run and the end by more than 5 % of its observed span per hour; a state driven by a seasonal or scheduled F0 source cycles daily and is not reported. Drivers are followed through its F1 inputs (a linked input reads the providing kind), through the states the conditions of its form read (a sequencer's stage conditions, a trip's condition, a regime's selector, an anti-recycle's watch, so a compressor that loads and unloads on a cycling header pressure is cyclic), through its `follow` to the providing kind's state and that state's dynamics, and, for a store, through the flows it takes part in to the other store and the scale states |
| `conservation_residual` | warning | for a fault whose detector is `conservation_residual`, the balance of its channels (inputs minus outputs, as the truth solver sums them) exceeds 3 times the combined noise sigma (the root of the summed squared sigmas of all its channels) at any sample; `detail` gives the worst residual and the limit |

**Output.** One `cli-output-v0.1` document `{domain, digest, seed, hours, samples, findings, errors, warnings}`, findings
sorted by entity, then channel, state or fault, then code, then time. The same domain, seed and hours give byte-identical
output. Any `error` finding makes the command fail with code `domain_lint` (exit 1), with the same report as the failure
`result`; warnings alone exit 0.

**Limits.** The lint judges the nominal day only: it exercises no effector and injects no fault, so a state that moves only
under command is idle here. Its thresholds are fixed; a warning is a prompt to look, not a defect.

## 28. Feedback loops: sampled link reads (phase P7k)

A linked input or a `follow` may declare `sample_s`: it then reads the linked states as they were at the latest instant of a
grid of that period, counted from the world start, strictly before the read (the world start itself for the first period).
A sampled read closes a feedback loop between kinds, such as a chiller and its cooling tower or a header and the consumers that
draw on it.

- **Validation**: kinds may read each other in a cycle only when every cycle contains a sampled read; `sample_s` is a whole
  number of milliseconds and a multiple of every integrated step of the source kind, so a sampled read always falls on the
  source's step boundaries; a kind read by sample does not churn.
- **Determinism**: a sampled read looks strictly into the past, so a loop never asks a state for a value it is still computing;
  the world tests prove a tank and cooler loop against an independent recurrence and against random read schedules.
- **History**: a source keeps history for a sampled reader back to the reader's progress minus the sample period and four of its
  steps, so a loop keeps a bounded history.

## 29. KPI truth (phase P7l)

`kinds[k].kpis` names the states that are the kind's key performance indicators (specific power, COP, approach, loading);
each is a declared state, at most once (`/kinds/k/kpis/i`). A component declares them in its `kind` and compose copies them.
A run's world tracks the KPI truth of every entity whose kind has KPIs or an `alarm` state, and the run seals it into its
director journal as one `kpi.truth` record at the close; the report shows it with the rest of the truth.

- **Sampling**: from the entity's birth, on a grid of its kind's shortest alarm step (60 s without an alarm), until its
  retirement; samples run on the world clock in time order with the world's events, after the events of the same instant,
  and never in a look-ahead copy. A KPI is the mean, minimum and maximum of its samples.
- **Alarms** (ISA-18.2 / EEMUA 191): an alarm is active while its code is `1` or `2`; an activation is a sample where it turns
  active after the first sample (unshelving an active alarm annunciates it again). Per alarm: `activations`, `chattering`
  (three activations within 60 s), `active_s`, `longest_active_s` and `active_at_end`. For the alarm system:
  `activations`, `peak_per_10_min` (the most activations in any 600 s window), `flood_intervals` (epoch-aligned 10-minute
  intervals with more than 10 activations), `chattering`, `stale` (alarms active for 24 h or more) and `standing_at_end`.
- **Determinism**: the samples read the states the world would read anyway, so tracking changes no trace, ledger or
  state history; the record is in the director journal's digest and a replay reproduces it. Worlds outside a run (truth
  oracles, the lint, the audit) do not track.

Cross-checks: a scripted alarm's activations, active time and state at the end against an independent scan of its code on
the alarm grid (two horizons); a KPI's mean, minimum and maximum against the same samples read from a world that does not
track; chattering, peak rate and flood intervals on constructed activation times.
