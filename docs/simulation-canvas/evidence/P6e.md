# P6e evidence — event clock

- **Date**: 2026-10-12
- **Change**: a `pulse` fault shape whose start and end are event instants; trips sample their condition at event instants
  inside a step; event-driven channels are checked at them; the RNG draw-order table. Spec:
  [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) section 19. Closes review item S11 and finding R0-12.
- **Found**: aggregation of dynamics inputs over links belongs to no phase; carried as R6e-01 (P12).

## Decisions

| Question | Decision | Why |
|---|---|---|
| What creates instants | Only pulse faults (start and end) | Existing worlds stay bit-identical; a sag is the case S11 names |
| Trip sampling | Step start plus the entity's event instants; fire when the condition held from a sample to the next sample (or step end) for `for_s` | Ride-through semantics (IEC 61000-4-34): a short sag passes, a long one trips |
| Trip timing | Visible at the boundary where `for_s` has elapsed since the first true sample | Follows from the hold-until-next-sample rule; one step earlier than P6c, and closer to the analytic time |
| Channels | Event-driven channels get one-off checks at a pulse's start and end; regular checks unchanged | The record carries the exact time; no cadence chain moves |
| Draws | The event clock draws only through the readings it emits, from that channel's own substreams | Proven by the periodic channel's identical records with and without the sag |

## What was built

- Model `FaultOnset.DurationS`, `ShapePulse`; contract `pulse` with required `duration_s`.
- World: `event_clock.go` (pulse scheduling, event instants from the injection order, one-off channel checks),
  `sampleTrip` in `protection.go`; a `kindEventCheck` queue event.
- Fixture: the chiller gains a supply voltage, a latched 0.2 s sag trip, a power-quality channel and two pulse faults
  (0.3 s sag, 0.1 s blip as the negative class); `start_chiller` requires the sag trip clear.

## Oracles

| Test | Independent derivation | Result |
|---|---|---|
| `TestASagTripsOnTheEventClockAndAShortOneRidesThrough` | Voltage nominal at both neighbouring boundaries (a sampled trip misses it); 0.3 s ≥ 0.2 s trips, 0.1 s does not | Pass |
| `TestAnEventDrivenChannelRecordsASagAndOtherStreamsDrawAsBefore` | Records at 600.4 s and 600.7 s; periodic channel identical with and without the sag | Pass |
| `TestATripFiresOnTheStepAfterItsConditionHeldForItsDuration` (updated) | Visible at `floor(t_c) + 1 + for_s` | Exact |
| `TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest` (sag case) | Two runs byte-identical; replay matches | Pass |

End to end with the binary: `CH1=voltage_sag@600.4`: the power-quality channel records 80 % at 00:10:00.4 and 100 % at
00:10:00.7; replay matches.

## Gates

- `make phase-gate`: passed in 1:36; behaviour pin unchanged. The flaky operator-endpoint transport test now checks the closed server instead of dialing a port another test may reuse.
- No comment lines added in Go.
