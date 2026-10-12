# P5b evidence — overrides, links, system entities, board

- **Date**: 2026-10-12
- **Change**: domain-spec v0.2 gains named instances, add-or-narrow overrides on an instance set, typed links, system
  kinds and boards. Spec: [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) sections 8 to 12.
  Closes finding R5a-02 and the board item of the review (P6 of [REVIEW-2026-10-12.md](../REVIEW-2026-10-12.md)).
- **Not in this phase**: aggregation of dynamics inputs over links (P6), the new dynamics vocabulary (P6), issuing and
  rebooting `boot_id` (P7). Links and boards are validated structure only.

## Decisions

| Question | Decision | Why |
|---|---|---|
| Named instances | `instances[i].names`; a set needs names, a template or both; the population is the names in order, then `count.default` templated ids; `count` and `churn` belong to the template | A name has no index to count or churn from |
| Name resolution | `InstanceSetFor` tries names across all sets, then templates; a name is domain-unique and must not fit the template of any set | A name must never be swallowed by a template or render twice |
| Explicit ids (R5a-02) | A name is accepted; only templated ids are counted against `count.min` and `count.max` | Names are bounded by their list |
| Override scope | One instance set; a named instance is a set of one name | The simplest unit that gives per-instance variation |
| Add | States, dynamics (for added states only), channels, faults, effectors; names unique domain-wide, a state name unique within the effective kind | Redeclaring is retyping; the existing duplicate rules express it |
| Narrow | `min < max` for a state (inside the kind's bounds if it has any, initial inside), a channel range inside the kind's range, an effector numeric argument `minimum` or `maximum` only tightened | "Never widen" is checkable from the kind alone |
| Remove, retype | Not expressible: the schema has no member for it and refuses unknown members | A located schema issue, with no extra code |
| Runtime form | `Compiled.InstanceKind(set)`: the kind itself without an override, an effective copy with one; the world builds each entity from it | One lookup changes state, dynamics, channels, faults and effectors together; single-kind worlds keep the same pointers |
| Lookups | `Fault`, `Channel`, `Effector`, `FaultKind` also search effective kinds; `EntityChannel`, `StateNamesOf`, `EffectorOf` give the entity's view | Added names resolve domain-wide; narrowing must apply per entity |
| Peers | Still the entities of the same kind (by name), overridden or not | Like-for-like pooling is unchanged |
| System entity | `kinds[].system: true`: exactly one instance set with exactly one name, no template, no churn, no `peer_residual` fault | A channel then belongs to exactly one entity |
| Links | `link_types[{name, from, to}]` between kinds, `links[{type, from, to}]` between instances; endpoints must be identifiable instances of the declared kinds; no duplicates | Structure only; P6 aggregates over it |
| Board | `boards[{name, members}]`; unique names, identifiable members, each instance on at most one board; the board owns `boot_id` | A reboot affects every entity on the device; runtime reboot is P7 |
| Contract layout | State, dynamics, channel, fault and effector definitions moved to `$defs` | A kind and an override share one definition |

## What was built

| Area | Change |
|---|---|
| Contract | Both copies of `domain-spec-v0.2.schema.json`: `$defs`, `kinds[].system`, `instances[].names` and `override`, `link_types`, `links`, `boards`; `id_template` and `count` no longer schema-required (domain issues at the same pointers) |
| `internal/model` | `Kind.System`, `InstanceSet.Names` and `Override`, `Override`, `Narrowing` and the link and board records (`override.go`) |
| `internal/domain` | `overrides.go` (registration, effective kinds, domain-wide lookups), `narrowing.go`, `structure.go` and `structure_links.go` (system kinds, links, boards), names in `instances.go`; checks run on the effective kind with a separate subject list so pointers stay in the document |
| `internal/world` | Entities take their kind from `InstanceKind`; population lists names first; peers compare kind name; argument-schema cache keyed by effector declaration; `StateNamesOf`, `EffectorOf` |
| Consumers | Run state history, device binding validation, suite fault entity choice, perturb `out_of_range` and the operator nameplate read the entity's or the set's declarations |
| Format | `domain format` knows the new members |
| Fixture | `internal/testsupport/testdata/structured-plant.domain.json`: named tank with added state, dynamics, channel, fault and effector and a narrowed level bound; named pump with a narrowed channel range and effector argument; a system `header`; three links; two boards |

## Proof

| Proof | Result |
|---|---|
| Override rules at the domain layer | `TestOverridesThatAddAreCheckedAtTheirPointers` (nine violations), `TestOverridesCannotRemoveOrRetype`, `TestNarrowingMayOnlyTightenWhatTheKindDeclares` (thirteen violations), `TestNarrowingWithinTheDeclaredBoundsIsAccepted`, `TestAnOverrideBuildsTheEffectiveKindOfItsSetAndLeavesTheKindAlone`, `TestNarrowedBoundsApplyToTheEntityNotTheKind`, `TestOverrideDeclarationsResolveDomainWide` |
| Named instances | `TestNamedInstancesResolveBeforeTemplatesAndJoinTheirSet`, `TestExplicitIdsMayBeNamesAndAreCountedOnlyWhenTemplated`, `TestInstanceSetsOfNamesAreCheckedAtTheirPointers` |
| Link, system and board rules at the domain layer | `TestLinksAreTypedAndTheirEndsExist`, `TestASystemKindHasExactlyOneNamedInstance`, `TestBoardsHostInstancesOnAtMostOneBoard`, `TestTheStructuredDomainKeepsItsLinksAndBoards` |
| Overrides take effect | `TestAnAddedStateFollowsItsClosedFormOnlyOnTheOverriddenInstance`, `TestANarrowedStateBoundClampsOnlyTheOverriddenInstance`, `TestAnAddedChannelEmitsOnlyFromTheOverriddenInstance`, `TestAnAddedFaultAndEffectorReachOnlyTheOverriddenInstance`, `TestANarrowedArgumentBoundIsEnforcedOnlyForTheOverriddenInstance`, `TestOutOfRangeUsesTheRangeNarrowedForTheEntity`, `TestStateHistoryOfAnOverriddenEntityIncludesItsAddedStates` |
| Analytic cross-check with an override | The added state against the closed form of a first-order lag; `TestTheSolverFindsTheClosedFormCrossingOfAFaultAddedByAnOverride` (first-observable and unavoidable times of the added fault on the added channel against the closed-form crossing on the scan grid) |
| Replay identity with an override | `TestARunWithOverridesIsByteReproducibleAndReplaysToTheSameDigest` (two runs byte-identical, the added channel in the trace, both added and narrowed effectors applied, artifact replays to a matching digest) |
| Suite | `TestEveryGeneratedScenarioInjectsAFaultIntoAnEntityThatCarriesIt` |
| Shipped domains | Behaviour pin matches `BEHAVIOUR_PIN.txt` unchanged |

## Behaviour pin changes

None. The new members are optional and absent from every shipped domain; `catalog describe` of a shipped domain prints the
same document.

## Gates

| Gate | Result |
|---|---|
| `make function-length` | passes |
| `make behaviour-pin-check` | matches, no pin line changed |
| Device path | `internal/device`, `internal/deviceworld` and `test/acceptance` pass |
| `make phase-gate` | see the report of the phase; wall clock about 1:02 for the test stage |

## Known limits

- An override varies a whole instance set. Per-name variation inside a set with a template is a set of one name.
- The scenario pre-check (`CheckEffectorArgs`) and the operator nameplate use the kind's declaration; a narrowed
  argument is enforced when the entity is invoked.
- Link endpoints and board members must be identifiable instances (a name or a template fit); whether an index lies in
  the starting population is not checked.
- An added fault is only injected into the entities of its set; a `peer_residual` detector on an added channel pools peers
  of the kind, which may lack the channel.
- The profile `entity_count` still replaces the default count of the first instance set only.

The commit hash of a phase is recorded by the first commit of the next phase.
