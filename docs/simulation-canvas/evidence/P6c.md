# P6c evidence — trips, lockouts, anti-recycle, boundaries, preconditions

- **Date**: 2026-10-12
- **Change**: a shared closed `condition`, the `trip` and `anti_recycle` forms, kind `boundaries` and effector `requires`.
  Spec: [`docs/design/DOMAIN_SPEC_V0_2.md`](../../design/DOMAIN_SPEC_V0_2.md) section 17. Closes review item S14 (the
  `responder` grammar is P6d). The `precondition` vocabulary item planned for P6d landed here as `requires`.
- **Not in this phase**: `responder`, `fallback_on_lease_loss`, the plant sequencer (P6d); scoring boundary violations (P12).

## Decisions

| Question | Decision | Why |
|---|---|---|
| `lockout` naming (S14) | Latched trip = lockout; timer = `anti_recycle`; lockout-tagout = an isolation state a precondition checks | One word per mechanism |
| Comparators | `lt`, `lte`, `gt`, `gte`, shared with the interlock | One closed set; no expressions |
| Trip timing | Condition sampled at step boundaries; tripped from the next boundary after `for_s` of true samples | Same hybrid rule as every other form; an analytic crossing gives the exact step |
| Reset | A latched trip clears when its `reset_on` state changes at a boundary | A reset command is an ordinary effector (`delta: 1` on a counter); no special command type |
| Boundaries | A boundary names a trip state and a severity | A boundary and a trip share one evaluator |
| Preconditions | `requires` conditions checked at the call time, before any random draw; refusal is generic to the operator, `precondition_refused` in the director log | MCP_SURFACE §4.5: a refusal must not hand the diagnosis to the consumer |
| Look-ahead | Interlocks and preconditions of a call dated after the clock evaluate on a copy of the entity's state | Found while building: a future-dated call integrated the entity ahead and earlier emissions read the later value (also true of interlocks before) |

## What was built

- Model: `Condition`, `Boundary`, trip and anti-recycle members on `F1Dyn`, `Kind.Boundaries`, `Effector.Requires`; contract with
  a `condition` definition.
- Domain layer: `validation_protection.go` (trip, anti-recycle, boundaries, requires).
- World: `protection.go` (shared condition, trip and anti-recycle steps, precondition refusal, look-ahead,
  `ViolatedBoundaries`); trip and anti-recycle values are discrete, so a read between steps returns the committed value.
- Fixture: a chiller kind in `utility-plant.domain.json` (latched low-flow trip, oil boundary, anti-recycle, LOTO and hand-mode
  preconditions, reset, isolate and release effectors).

## Oracles

| Test | Independent derivation | Result |
|---|---|---|
| `TestATripFiresOnTheStepAfterItsConditionHeldForItsDuration` | Analytic crossing of an F0 ramp, `floor(t_c) + 1 + for_s + 1` | Exact step; boundary reported from that step only |
| `TestALatchedTripHoldsUntilResetAndAnUnlatchedOneClears` | Expected boundary steps for trip, restore and reset, latched and unlatched | Pass |
| `TestPreconditionsRefuseALockedOutRecyclingOrIsolatedDevice` | Timer left 599 s one step after the edge; refusals during the timer, under lockout and isolation; generic operator reason | Pass |
| `TestACallDatedAheadOfTheClockDoesNotChangeEarlierReadings` | Same readings with and without a call dated 900 s; fails without the look-ahead | Pass |
| `TestTripAntiRecycleBoundaryAndPreconditionRulesAreLocated` | Each rule refused at its pointer | Pass |
| `TestVocabularyFixtureRunsAreByteReproducibleAndReplayToTheSameDigest` (chiller case) | Two runs byte-identical; replay matches | Pass |

End to end with the binary: `CH1=flow_loss@600`, `stop_chiller` at 1200 s and `start_chiller` at 1500 s: the lockout channel
reads 0 until the fault and 1 from 612 s, the start is refused (`precondition_refused` in the director journal), replay matches.

## Gates

- `make phase-gate`: passed; behaviour pin unchanged. Five long single-goroutine run tests now skip the `-race` pass (they run in the coverage pass), taking the run package from 36 s to 17 s under race.
- No comment lines added in Go beyond the `determinism-safe` marker.
