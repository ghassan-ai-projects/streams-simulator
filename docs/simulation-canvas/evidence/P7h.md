# P7h evidence — coupling over typed links

- **Date**: 2026-10-13
- **Change**: domain-spec v0.2 §21: linked inputs `{state, coef, link: {type, aggregate, default}}` and `follow` on dynamics,
  aggregates `sum`, `mean`, `min`, `max`, `count_running`; validation (declared link into the reading kind, source state,
  defaults, no stateful form follows, kind-level cycle ban over links); world (incoming-link index, presence from birth to
  retirement, history retained back to the oldest time anything downstream can read, readers pushed forward as a memory bound);
  COMPONENTS §5 updated.
- **Fixture**: [`coupled-plant`](../../../internal/testsupport/testdata/coupled-plant.domain.json): three units (1 s) feeding a
  header `hdr-a` (70 s mean-flow step), an unlinked header `hdr-b`, and a plant (600 s) reading both headers.

## Exit evidence

| Check | Result |
|---|---|
| A header with parallel units agrees with an independent recursion: sum and max/min by `follow`, an integrator over the followed sum, a lag over the mean, `count_running` across a trip | `TestAHeaderFollowsAndReadsTheAggregatesOfItsLinkedUnits` (1e-9) |
| An unlinked instance keeps its own dynamics; `mean` falls back to its default | `TestAnUnlinkedHeaderKeepsItsOwnDynamicsAndTheDefaults` |
| A retired unit leaves the aggregates after its retirement | `TestARetiredUnitLeavesTheAggregatesAfterItsRetirement` |
| Read-order independence: six read schedules give bit-identical header values that match the reference | `TestCouplingDoesNotDependOnTheOrderEntitiesAreRead` |
| A three-level chain with unaligned steps, and a relay with no history of its own | `TestAThreeLevelChainAgreesWithTheReference`, `TestARelayWithNoHistoryOfItsOwnKeepsItsSourcesHistoryForItsReaders` |
| 120 seeded random forward reads across all three levels, with a fault, equal a world read once | `TestRandomForwardReadsAgreeWithAWorldReadOnce` |
| The retention is necessary and transitive | Mutation: without the readers' floor, three tests fail; without transitivity, the relay test fails |
| Fan-out cost is linear | One source, 25 → 400 readers, 30 simulated minutes: 36 ms → 472 ms |
| Each aggregate, empty sets, presence boundaries | `TestEachAggregateOverLinkedValuesAndOverNone`, `TestALinkedInstanceCountsFromItsBirthThroughItsRetirement` |
| Validation | `TestLinkReadsNameADeclaredLinkIntoTheirKindAndAStateOfItsSource` (7 cases incl. two-kind and self cycles), `TestALinkInputIsNoDependencyInsideItsOwnKind` (mutation-checked), `TestAStatefulFormCannotFollowLinkedInstances` |
| `LinkReads` lists each link type once | `TestAKindListsEachLinkTypeItReadsOnce` |
| v0.1-equivalent worlds unchanged | Behaviour pin identical to `BEHAVIOUR_PIN.txt`; every existing world, run and replay test passes |

## Scope notes

- Load sharing by capacity needs a product of states and moves to P7i with `product` (a share is `capacity / sum(capacity)`).
- Feedback between kinds (a header pressure read back by its units) is refused by the cycle ban; it needs an explicit delay
  form, to be specified when a catalog component needs it.
- The header-store pattern is the integrator over a followed sum shown here; a `store` reads linked inputs through the same input
  path.
