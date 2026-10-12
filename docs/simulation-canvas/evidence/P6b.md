# P6b evidence — `schedule`, `regime`, `peak`

- **Date**: 2026-10-12
- **Change**: the dynamics vocabulary gains a clock-aligned `schedule` on F0 states, a two-regime lag (`regime`) and the
  billing-period demand peak (`peak`, with `peak_mtd` seeding and a ratchet). Spec:
  [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) section 16. Closes review item S9.
- **Not in this phase**: calendar billing months and projected demand (carried as R6b-01 to P12), `trip`/`lockout` (P6c), the
  event clock (P6e).

## Decisions

| Question | Decision | Why |
|---|---|---|
| Schedule home | An `f0.schedule` added to the F0 value | Exogenous profiles are analytic in time; one place, no new tier |
| Schedule alignment | Periodic profiles align to the Unix epoch; one-off profiles to the world start | A tariff follows the clock whatever the world start |
| Regime | A lag per regime (`gain`, `time_constant_s`), switched by a selector state at step boundaries | Exact exponential within a regime; the crossing time is analytic and the switch lands on the next boundary (hybrid rule) |
| Peak | Input sampled at step starts; epoch-aligned blocks close when a step crosses a boundary; period roll-over keeps the previous peak; value = max(peak so far, ratchet × previous) | Block-aligned demand as utilities bill it; any world start works because closing is by crossing |
| Peak reads | A read between steps returns the committed value | The value changes only when a block closes |

## What was built

- `internal/model/vocabulary.go` (`Schedule`, `SchedulePoint`, `Regime`, form names); F0 `schedule`; F1 `selector`, `below`,
  `above`, `block_s`, `period_s`, `ratchet`, `previous_peak`. Contract and embedded schema, with a `regime` definition.
- Domain layer: `validation_vocabulary.go` (schedule order and period, peak input and alignment, regime selector);
  `f1Parameters` became a table of checks.
- World: `schedule.go`, `regime_peak.go`; F1 dispatch for the two forms; the peak's block state is part of the step snapshot.
- Fixture `internal/testsupport/testdata/utility-plant.domain.json`: tariff and load schedules, a demand peak with a ratchet,
  a transformer oil regime, a load-surge fault and a shed-load effector.

## Oracles

| Test | Independent derivation | Result |
|---|---|---|
| `TestAScheduleIsATableLookup` | Hand-computed table values at boundaries and interpolated points, periodic and one-off, across the period wrap | Within 1e-9 |
| `TestARegimeSwitchesAtTheAnalyticCrossing` | Piecewise closed form switching at `ceil(600 · ln 4)` s | Within 1e-6 |
| `TestADemandPeakIsTheLargestBlockMeanWithARatchet` | Direct recomputation of block means, period roll-over and ratchet over two days | Within 1e-9 |
| `TestScheduleRegimeAndPeakRulesAreLocated` | Each rule refused at its pointer | Pass |
| `TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest` (utility case) | Two runs byte-identical; replay digest matches | Pass |

End to end with the binary: a one-day `run` of the utility plant with `meter-1=load_surge@32400` emits 11520 records; the
tariff steps at 07:00 and 21:00, the demand peak never falls and is floored by the ratchet (0.6 × 90 kW), the oil switches to
its hotter regime above 65 °C; `replay` matches the digest.

## Gates

- `make phase-gate`: passed in 1:48; behaviour pin unchanged.
- No comment lines added in Go.
