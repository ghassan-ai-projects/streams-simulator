# P7h review record — coupling over typed links

- **Reviewed**: the P7h working tree against `3d59980` (2026-10-13)
- **Reviewer**: one independent Sonnet subagent, with experiments in a scratch copy (a panic on pruned-history reads, random
  forward read schedules, a fan-out benchmark).
- **Outcome**: 2 High, 3 Medium, 3 Low. The Highs and the fan-out cost are fixed by a redesign of history retention; the rest is
  fixed, documented or registered.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | History reach was one level deep: in a chain unit → header → plant with unaligned steps, forward reads hit pruned history and depended on read order | Redesign: a source keeps history back to the oldest time anything downstream can read (`readersFloor`, transitive over the reader graph); `TestAThreeLevelChainAgreesWithTheReference` (unit 1 s, header 70 s, plant 600 s), `TestRandomForwardReadsAgreeWithAWorldReadOnce` (120 seeded reads, bit-identical to a world read once), `TestARelayWithNoHistoryOfItsOwnKeepsItsSourcesHistoryForItsReaders` (mutation-checked: without transitivity it fails) |
| 2 | High | Past reads of a `follow` state beyond its sources' history are order-dependent | Documented in §21 ("Reading into the past"): a linked value is defined back to the start of its sources' history, like any state |
| 3 | Med | Pushing readers scanned every reader on every nested read: quadratic in fan-out | Pushing is now only a memory bound, guarded by the time the source last pushed its readers (nested reads are O(1)); fan-out 25 → 400 readers: 36 ms → 472 ms, linear |
| 4 | Med | A retired reader read after its retirement uses pruned history | Documented in §21: a retired instance's states are defined up to its retirement; a retired reader stops holding its sources' history once pushed through its retirement |
| 5 | Med | Truth oracle worlds replay no churn, so presence-sensitive aggregates differ from the run | Registered as R7h-01 (P7k) |
| 6 | Low | Spec and tests overclaimed "every read schedule" | §21 now names what is proven; the chain, relay and random tests added |
| 7 | Low | `LinkReads` listed a link type once per use | De-duplicated; `TestAKindListsEachLinkTypeItReadsOnce` |
| 8 | Low | A link read of an override-added source state | No change: validation requires the state on the source's base kind, which every instance has |

Checked and sound per the reviewer: the cycle ban (overrides, self links, `follow`), termination and re-entrancy, churn births
and deaths, aggregate semantics, sorted summation, the uncoupled hot path (about 3% on a read microbenchmark) and the repository
rules.
