# Domain spec v0.2, semantics part 1: kinds and instances

- **Status**: implemented in phase P5a of the simulation-canvas goal ([PLAN](../simulation-canvas/PLAN.md)).
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
kinds[]        exactly one kind in part 1 (maxItems 1)
instances[]    exactly one instance set in part 1, naming that kind
profiles[], ground_truth                             unchanged from v0.1
```

### 2.1 Kind

`kinds[0]` = `{name, state?, dynamics?, channels, faults, effectors?}`. The five sections are the v0.1 definitions, member for
member. `name` matches `^[a-z][a-z0-9_-]{0,63}$` and becomes the `entity_type` of every instance (v0.1's optional
`entities.entity_type`, now mandatory and named once).

Every instance of a kind has every state, channel, fault and effector of the kind. Part 1 has no override, so every instance of a
kind is identical apart from its id.

### 2.2 Instances

`instances[0]` = `{kind, id_template, count{default,min?,max?}, churn?}`. These are v0.1's `entities` members, with the kind named
explicitly. `kind` must equal `kinds[0].name`; a mismatch is an issue at `/instances/0/kind`.

An instance set is a *templated population*: `id_template` renders ids (`{n}` is the 1-based index, any other `{name}` placeholder
renders `a`), `count.default` instances exist at the start, and `churn` births and retires instances during a run.

**Per-instance parameters.** v0.1 allowed no variation between entities, so part 1 has none to carry. Variation is the add-or-narrow
override of P5b; named instances (`M1`, `AC1`) arrive there with it.

### 2.3 Why one kind

The runtime world holds one state set, one channel set and one fault set per domain. Several kinds need that to become per-kind,
which is a world change, and "no change to world, perturb, adapter, truth, score and run" is this part's contract. `maxItems: 1`
makes the limit explicit and cheap to lift: the schema, the lowering and the pointer mapping are the only places that name it.

## 3. Compilation to the runtime world

A v0.2 document compiles to the **runtime form**, `model.DomainSpec`, which is exactly the v0.1 document shape and is the only form
the world, perturbation layer, adapter engine, truth, scoring and run code read.

| v0.2 | Runtime (`model.DomainSpec`) |
|---|---|
| `id`, `version`, `title`, `description`, `axes`, `stresses`, `profiles`, `ground_truth` | same member |
| `kinds[0].state`, `.dynamics`, `.channels`, `.faults`, `.effectors` | `state`, `dynamics`, `channels`, `faults`, `effectors` |
| `kinds[0].name` | `entities.entity_type` |
| `instances[0].id_template`, `.count`, `.churn` | `entities.id_template`, `.count`, `.churn` |

`internal/domain/internal/domain/lowering.go` is the whole mapping. After it the v0.1 cross-checks run unchanged on the runtime form.

**Digest.** `domain_digest` is the RFC 8785 digest of the validated **v0.2 document**, as v0.1's was of its document. Formatting
never changes it. Every shipped domain's digest therefore changes once, with its bytes, and `world_digest` (which includes the domain
digest and `sim_version`) with it. `Compiled.Raw`, the document embedded in run artifacts as `domain_spec`, is the v0.2 document;
replay parses it through the same path.

**`sim_version`** is `0.3.0`: the digest inputs of every run changed.

**Equivalence.** The change is a re-expression, so a rewritten domain must compile to the runtime world its v0.1 form produced:

1. `TestEveryShippedDomainCompilesToItsV01RuntimeWorld` compares the RFC 8785 digest of each domain's compiled runtime form with
   the digest of its v0.1 compiled form, recorded before the rewrite;
2. the behaviour pin shows `trace.jsonl`, `ledger.jsonl`, `world_state_history.jsonl`, every `verify` and `replay` result and the
   generated suites byte-identical for all eight domains and both adapters.

## 4. Validation issues carry pointers (finding R1-01)

Every issue a domain produces is *located*: an RFC 6901 pointer into the **v0.2 document**.

- Schema violations: a "missing required property" issue points at the missing member (`/kinds/0/channels/2/resolution`), not at
  its parent; "additional property not allowed" points at the unexpected member. (Both come from the shared validator, so every
  contract benefits.)
- Cross-reference issues (an undeclared state, fault or effector, a duplicate name, a dynamics cycle, an unimplemented option) point
  at the member that is wrong: `/kinds/0/channels/1/observes`, `/kinds/0/faults/2/affects/0/state`,
  `/profiles/0/fault_weights/<fault>`. The checks work in the runtime layout (`/channels/1/observes`); one function,
  `documentPointer`, places a runtime pointer into the v0.2 document.
- A document that is not JSON has no location and keeps one unlocated issue.

## 5. Explicit entity ids (finding R2-02)

A scenario's `world.entity_ids` and the director's `sim.world.create` entity list name instances explicitly. **Decision:** an explicit
id list is checked against the domain's instance set, by the run module when it creates the world, so every surface (CLI flags,
scenario file, director, replay) is covered once:

1. the number of ids is within `count.min` (at least 1) and `count.max`;
2. each id matches the `id_template`: `{n}` matches digits, any other placeholder matches `[A-Za-z0-9_.-]+`, literal text matches
   exactly;
3. the instance set's own `count.default` must lie within `count.min` and `count.max` (a domain issue at
   `/instances/0/count/default`).

An empty list selects the default population. A violation is a refusal at world creation (a scenario violation at `/world`).
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
