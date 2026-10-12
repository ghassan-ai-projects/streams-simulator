# Domain spec v0.2, semantics part 1: kinds and instances

- **Status**: implemented in phase P5a of the simulation-canvas goal ([PLAN](../simulation-canvas/PLAN.md)); amended in phase P5a2
  (several kinds and instance sets per domain, sections 2.3 and 3).
- **Contract**: [`docs/contracts/domain-spec-v0.2.schema.json`](../contracts/domain-spec-v0.2.schema.json), embedded in the binary
  (`internal/schemas`), parity-tested against the committed copy. v0.1 is removed: the project is pre-release and the eight shipped
  domains are rewritten in the same change.
- **Later parts**: add-or-narrow overrides, typed links, system entities and the board are P5b; the new dynamics vocabulary
  (`store`, `flow`, `regime`, `trip`, ...) and the event clock are P6. Nothing in this part depends on them.

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

**Per-instance parameters.** v0.1 allowed no variation between entities, so part 1 has none to carry. Variation is the add-or-narrow
override of P5b; named instances (`M1`, `AC1`) arrive there with it.

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

**RNG draw order.** The world draws only from named substreams (`<entity>/<channel>/noise`, `fault/<id>/<state>`,
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
`Compiled.CheckEntityIDs` implements the rule. Ids that fit the template are all the template can render, so churned births are
unaffected.

## 6. Canonical formatting

`streamsim domain format <path>` rewrites a valid domain file in place and prints `{formatted, changed, digest}`. The format is a pure
function of the document:

- two-space indentation, newline at the end of the file, strings written as UTF-8 without HTML escaping, number text preserved;
- object members in a fixed order: the document's members in schema order (`id`, `version`, `title`, `description`, `axes`,
  `stresses`, `kinds`, `instances`, `profiles`, `ground_truth`); within any other object the identifying and structural members first
  (`id`, `name`, `kind`, `id_template`, `count`, `churn`, `target`, `tier`, `form`, `state`, `unit`, `initial`, `default`, `min`,
  `max`, `dynamics`, `channels`, `faults`, `effectors`), then the rest alphabetically, then `description`;
- array order is semantic and never changed;
- an object or array that fits in 110 columns is written on one line, otherwise one member per line.

The command refuses a document that does not validate and leaves the file untouched. It is idempotent, and the digest does not change.
The canvas and agents call it so a hand edit and a canvas edit of the same world produce the same bytes.

## 7. Deviations recorded

[`DECISIONS.md`](../DECISIONS.md) D-03, D-04 and D-14 carry deviation entries describing what v0.2 does and does not change in them.
Part 1 changes none of their runtime behaviour.
