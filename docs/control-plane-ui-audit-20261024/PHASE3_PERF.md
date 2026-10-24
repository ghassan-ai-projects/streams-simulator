# Phase 3 performance: live values on the 23-instance factory

Round 3A. Budget from the plan: "a performance budget for live trends on the 23-instance factory, measured before the phase starts and met at its end."

## What is measured

The factory is `examples/factory-utilities.domain.json`: 23 instances, 82 channels (series). The load is 30 simulated minutes at one reading per series every 10 s: 14,760 observations plus one `clock.advanced` record a minute, the same order as the real recording of `w-1` (14,490 observations in 30 minutes).

Two things, both through the real stack except the network (fake bridge, real decoder, real `ReviewStore`, real React tree):

1. **Ingest.** The whole load pushed through the bridge follower, the journal decoder and the store, then one 250 ms flush.
2. **Plant update.** The plant screen mounted with the factory (graph, outline, inspector), then 180 flushes of 10 simulated seconds each. Every flush carries a reading for all 82 series, which is the worst case of a Watch-paced run (120 s per s: each 250 ms flush covers 30 simulated seconds). Counts are InstanceNode renders (calls to `describeNode`, which each node render makes once) and React commits (Profiler).

Counts are deterministic and asserted in the default test run (`LivePlantRenders.test.tsx`). Wall times are not asserted anywhere; they are printed by:

```
cd canvas
VITE_PERF=1 npx vitest run src/ui/review/internal/LivePlantLoad.test.tsx --reporter=verbose
```

(The suite is skipped without `VITE_PERF=1`.) Code: `src/test/factoryFixture.ts`, `src/test/livePlantHarness.tsx`, `src/ui/review/internal/LivePlantLoad.test.tsx`, `src/ui/review/internal/LivePlantRenders.test.tsx`.

Machine: one Apple Silicon Mac (Darwin 27), Node 26.11, jsdom 30, Vitest 5.0.5, other work running. Absolute times are for comparison on this machine only; the counts are the evidence.

## Baseline (before round 3A)

Measured on the code at the start of the round. At that commit only a connected session could show live values on the plant, so the baseline drives the same `ConnectedPlant` and `PlantScreen` tree with `openLive`. Three runs.

| Measure | Baseline |
|---|---|
| Ingest of 14,760 observations (one batch, decode and store) | 58 to 61 ms |
| InstanceNode renders over 180 flushes | 8,280 (46 per flush: 23 instances, drawn twice) |
| React commits over 180 flushes | 370 (2 per flush: an update and a nested update) |
| React render time over 180 flushes (Profiler actual duration) | 1.53 to 1.58 s |
| Wall time of one flush including the feed, median / p95 | 18.2 to 19.4 ms / 25.5 to 27.9 ms |

## After round 3A

Three runs, `watch` mode (the new open mode), same load.

| Measure | After |
|---|---|
| Ingest of 14,760 observations | 62 to 63 ms (the decoder now also reads quality and event time; the index keeps two more arrays) |
| InstanceNode renders over 180 flushes | 4,102 (23 per flush: every instance, once) |
| React commits over 180 flushes | 370 (unchanged; the second commit is the graph library's own and no longer redraws the nodes) |
| React render time over 180 flushes | 0.76 to 0.80 s |
| Wall time of one flush including the feed, median / p95 | 5.7 to 6.0 ms / 9.0 to 9.6 ms |

What changed, in two parts:

- Node components are memoised (`nodeTypes`), so the graph library's second pass over the nodes no longer redraws them (46 to 23 per flush).
- The flow nodes are kept between flushes when their content is equal (`NodeMemory`), and handles are re-measured only for nodes whose shape (row count, hidden channels, frame count) changed, not on every flush.

Quiet flushes, same plant, counted in the default test run:

| Flush carries | Node renders before | Node renders after |
|---|---|---|
| The clock only | 23 | 0 |
| One reading for one instance (CT-1) | 46 | 1 |
| A reading for every instance | 46 | 23 |

"Before" in this table is the code with the node memory, the memoised node types and the selective re-measure switched off, in the same tree.

## Memory

`PlaybackIndex` keeps one time series per entity and channel. A review of a finished run keeps all of it (unchanged). A watched run (`watch` mode) keeps at least the newest 2,048 readings of each series and never more than twice that: when a series reaches 4,096 readings its oldest are dropped back to 2,048. At the 10 s cadence of the load that is at least 5 h 40 min of every series (at the fastest series in the factory, 2 s, 68 minutes), and the longest trend window is 30 minutes.

| Load | Observations ingested | Readings held | Ingest |
|---|---|---|---|
| 30 minutes | 14,760 | 14,760 | 62 ms |
| 6 hours | 177,120 | 177,120 (below the cap) | 0.6 s |
| 24 hours | 708,480 | 204,672 (bounded) | 2.4 s |

Heap retained after a forced collection: 24 hours watched 11.8 MB against 26.4 MB unbounded; 30 minutes under 1 MB either way. The bound matters for runs of days; the dominant cost of a long run is the 250 ms redraw, not the index.

Limits of the policy: a trend or a past value older than the retained window is not available while the run is watched (the plant view only shows the newest reading and the last 30 minutes). A review of an unfinished run opened from the Review screen is not bounded (existing behaviour, not changed in this round).

## Not measured

- A real browser (Chrome): jsdom has no layout or paint; the graph is drawn with stubbed sizes. Render counts and React time are meaningful, absolute times are not a frame budget.
- Several hours of wall time.
- The inspector's own cost (one entity, at most a dozen channels, a trend over at most 4,096 readings per channel). Its work is a few hundred operations per flush and is not in the table above.
